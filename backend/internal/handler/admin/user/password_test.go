package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
)

// Only persistence is replaced; requests exercise the handler, data-scope lookup
// and real password policy/hash service without connecting to a database.
type resetPasswordConn struct{ gorm.ConnPool }

func (c *resetPasswordConn) BeginTx(context.Context, *sql.TxOptions) (gorm.ConnPool, error) {
	return c, nil
}
func (*resetPasswordConn) Commit() error   { return nil }
func (*resetPasswordConn) Rollback() error { return nil }

func TestResetPasswordAcceptsSelectedPasswordAndReportsFailure(t *testing.T) {
	for _, tc := range []struct {
		name, current, password, wantPassword, wantMessage string
		omitPassword                                       bool
		queryError                                         error
	}{
		{name: "choose a new password instead of reused default", current: "123456", password: "Fresh123!", wantPassword: "Fresh123!"},
		{name: "legacy caller", current: "Old12345", omitPassword: true, wantPassword: "123456"},
		{name: "reused password", current: "123456", password: "123456", wantMessage: "新密码不能与当前或最近使用的密码相同"},
		{name: "empty password", current: "Old12345", password: "", wantMessage: "请输入新密码"},
		{name: "blank password", current: "Old12345", password: "      ", wantMessage: "请输入新密码"},
		{name: "oversized password", current: "Old12345", password: strings.Repeat("a", 73), wantMessage: "密码不能超过 72 字节"},
		{name: "short password", current: "Old12345", password: "abc", wantMessage: "密码至少需要 6 位"},
		{name: "user unavailable", current: "Old12345", password: "Fresh123!", queryError: gorm.ErrRecordNotFound, wantMessage: "用户不存在或不在当前管理员可管理范围内"},
		{name: "internal error", current: "Old12345", password: "Fresh123!", queryError: errors.New("private database unavailable"), wantMessage: "密码重置失败，请稍后重试"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := passwordutil.Hash(tc.current)
			if err != nil {
				t.Fatal(err)
			}
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: &resetPasswordConn{}, SkipInitializeWithVersion: true}), &gorm.Config{
				DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent),
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
				switch dest := tx.Statement.Dest.(type) {
				case *model.Admin:
					*dest = model.Admin{ID: 7, RoleID: 987654}
				case *model.Role:
					*dest = model.Role{ID: 987654, Name: "超级管理员", Status: 1}
				case *model.User:
					if tc.queryError != nil {
						tx.AddError(tc.queryError)
						return
					}
					*dest = model.User{ID: 9, Password: hash}
				case *model.Setup:
					tx.AddError(gorm.ErrRecordNotFound)
				case *[]uint:
					*dest = nil
				default:
					t.Fatalf("unexpected query destination %T", dest)
				}
			}); err != nil {
				t.Fatal(err)
			}
			var savedHash string
			if err := db.Callback().Update().Replace("gorm:update", func(tx *gorm.DB) {
				savedHash = tx.Statement.Dest.(map[string]interface{})["user_password"].(string)
				tx.RowsAffected = 1
			}); err != nil {
				t.Fatal(err)
			}
			previous := database.DB
			database.DB = db
			t.Cleanup(func() { database.DB = previous })
			form := url.Values{"id": {"9"}}
			if !tc.omitPassword {
				form.Set("password", tc.password)
			}
			c := app.NewContext(0)
			c.Set("admin", &model.Admin{ID: 7})
			c.Request.SetMethod("PATCH")
			c.Request.Header.SetContentTypeBytes([]byte("application/x-www-form-urlencoded"))
			c.Request.SetBodyString(form.Encode())
			NewAdminUserHandler().ResetPassword(context.Background(), c)
			if tc.wantMessage != "" {
				assertUserWriteErrorResponse(t, c, tc.wantMessage)
				if savedHash != "" {
					t.Fatal("failed reset modified password")
				}
				return
			}
			var body struct{ Code int }
			if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != 0 || !passwordutil.Verify(savedHash, tc.wantPassword) {
				t.Fatalf("selected password was not saved: response=%s", c.Response.Body())
			}
		})
	}
}
