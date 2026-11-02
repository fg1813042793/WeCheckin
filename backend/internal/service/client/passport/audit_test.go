package passport

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
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

func TestPasswordLoginAuditDistinguishesMissingAccountAndDatabaseFailure(t *testing.T) {
	for _, tc := range []struct {
		err    error
		reason string
	}{
		{gorm.ErrRecordNotFound, "invalid_credentials"},
		{errors.New("database token-secret"), "system_error"},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) { tx.AddError(tc.err) }); err != nil {
				t.Fatal(err)
			}
			previous := database.DB
			database.DB = db
			t.Cleanup(func() { database.DB = previous })
			ctx, recorder := auditlog.WithRecorder(context.Background())
			result, err := LoginByPwdContext(ctx, "attempted-account", "password-secret", "", "")
			if result != nil || err == nil || err.Error() != "账号或密码错误" {
				t.Fatalf("result/error = %v/%v", result, err)
			}
			details := recorder.Snapshot()
			if details.ReasonCode != tc.reason || details.ActorID != 0 || details.ActorName != "" {
				t.Fatalf("audit = %+v", details)
			}
			raw, _ := json.Marshal(details)
			for _, secret := range []string{"token-secret", "password-secret", "attempted-account"} {
				if strings.Contains(string(raw), secret) {
					t.Fatal("audit contains credentials or an unverified account")
				}
			}
		})
	}
}

func TestDingTalkLoginAuditRejectsMissingCode(t *testing.T) {
	ctx, recorder := auditlog.WithRecorder(context.Background())
	_, err := loginByDingTalkOAuthCodeContext(ctx, nil, "corp", "", "", "")
	if err == nil || recorder.Snapshot().ReasonCode != "auth_code_invalid" {
		t.Fatalf("error/audit = %v/%+v", err, recorder.Snapshot())
	}
}

func TestLoginUserDatabaseFailureDoesNotReturnSuccessfulEmptyToken(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) { tx.AddError(errors.New("database unavailable")) }); err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previous })
	ctx, recorder := auditlog.WithRecorder(context.Background())
	result, err := LoginUserContext(ctx, "openid-secret", "", "")
	if err == nil || result != nil {
		t.Fatal("database failure returned login success")
	}
	if details := recorder.Snapshot(); details.ReasonCode != "system_error" || details.ActorID != 0 {
		t.Fatalf("audit = %+v", details)
	}
}
