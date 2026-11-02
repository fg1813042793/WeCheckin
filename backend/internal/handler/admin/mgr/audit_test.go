package mgr

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

func TestLoginCaptchaAuditKeepsRequestContextAndRedactsInput(t *testing.T) {
	for _, tc := range []struct {
		name, id, reason, message string
		configErr                 error
	}{
		{"invalid", "", "captcha_invalid", "验证码错误或已过期", nil},
		{"unavailable", strings.Repeat("a", 32), "captcha_unavailable", "验证码服务暂不可用，请稍后重试", nil},
		{"configuration failure", "", "system_error", "", errors.New("database password-secret")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, recorder := auditlog.WithRecorder(context.Background())
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
				if tx.Statement.Context.Value(testContextKey{}) != "request" {
					t.Fatal("request context lost before captcha validation")
				}
				if tc.configErr != nil {
					tx.AddError(tc.configErr)
					return
				}
				tx.Statement.Dest.(*model.Setup).Value = "1"
			}); err != nil {
				t.Fatal(err)
			}
			previous := database.DB
			database.DB = db
			t.Cleanup(func() { database.DB = previous })
			c := &app.RequestContext{}
			c.Request.Header.SetMethod("POST")
			c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			c.Request.SetBodyString("name=attempted-account&password=password-secret&captchaCode=code-secret&captchaId=" + tc.id)
			NewAdminMgrHandler().AdminLogin(context.WithValue(ctx, testContextKey{}, "request"), c)
			details := recorder.Snapshot()
			if details.ReasonCode != tc.reason || details.ActorID != 0 || details.ActorName != "" {
				t.Fatalf("audit = %+v", details)
			}
			var response struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(c.Response.Body(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Code == 0 || (tc.message != "" && response.Msg != tc.message) {
				t.Fatalf("response = %+v", response)
			}
			raw, _ := json.Marshal(details)
			for _, secret := range []string{"password-secret", "code-secret", "attempted-account"} {
				if strings.Contains(string(raw), secret) {
					t.Fatal("audit contains credentials or an unverified account")
				}
			}
		})
	}
}

type testContextKey struct{}
