package adminauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

func TestLoginAuditRejectsUnverifiedActorAndRedactsCredentials(t *testing.T) {
	hash, err := passwordutil.Hash("test-password-secret")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, password, reason, message string
		queryErr                        error
		status                          int
	}{
		{"missing", "test-password-secret", "invalid_credentials", "账号或密码错误", gorm.ErrRecordNotFound, 1},
		{"wrong password", "wrong-password-secret", "invalid_credentials", "账号或密码错误", nil, 1},
		{"database failure", "test-password-secret", "system_error", "账号或密码错误", errors.New("db token-secret"), 1},
		{"disabled", "test-password-secret", "account_disabled", "账号已禁用", nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
				if tc.queryErr != nil {
					tx.AddError(tc.queryErr)
					return
				}
				*tx.Statement.Dest.(*model.Admin) = model.Admin{ID: 7, Name: "verified-name", Password: hash, Status: tc.status}
			}); err != nil {
				t.Fatal(err)
			}
			previous := database.DB
			database.DB = db
			t.Cleanup(func() { database.DB = previous })
			ctx, recorder := auditlog.WithRecorder(context.Background())
			result, err := LoginContext(ctx, "attempted-account", tc.password, "", "")
			if result != nil || err == nil || err.Error() != tc.message {
				t.Fatalf("result/error = %v/%v", result, err)
			}
			details := recorder.Snapshot()
			if details.ReasonCode != tc.reason || details.Reason == "" {
				t.Fatalf("reason = %+v", details)
			}
			if details.ActorID != 0 || details.ActorName != "" || details.ActorType != "" {
				t.Fatalf("unverified actor = %+v", details)
			}
			raw, _ := json.Marshal(details)
			for _, secret := range []string{tc.password, hash, "token-secret", "attempted-account"} {
				if strings.Contains(string(raw), secret) {
					t.Fatal("audit contains credentials or an unverified account")
				}
			}
		})
	}
}

func TestAdminAccessAuditDistinguishesDenialAndSystemFailure(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if canceled {
			cancel()
		}
		ctx, recorder := auditlog.WithRecorder(ctx)
		_, _, _, err := adminLoginAccess(ctx, nil, &model.Admin{})
		cancel()
		want := "admin_access_denied"
		if canceled {
			want = "system_error"
		}
		if err == nil || recorder.Snapshot().ReasonCode != want {
			t.Fatalf("error/reason = %v/%+v", err, recorder.Snapshot())
		}
	}
}
