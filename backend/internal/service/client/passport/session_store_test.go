package passport

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

type sessionTestDialector struct{ gorm.Dialector }
type sessionTestMigrator struct{ gorm.Migrator }

func (d sessionTestDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return sessionTestMigrator{d.Dialector.Migrator(db)}
}
func (sessionTestMigrator) HasTable(interface{}) bool { return false }

type sessionTestRedisHook struct{ process func(goredis.Cmder) error }

func (h sessionTestRedisHook) DialHook(next goredis.DialHook) goredis.DialHook { return next }
func (h sessionTestRedisHook) ProcessHook(_ goredis.ProcessHook) goredis.ProcessHook {
	return func(_ context.Context, cmd goredis.Cmder) error { return h.process(cmd) }
}
func (h sessionTestRedisHook) ProcessPipelineHook(_ goredis.ProcessPipelineHook) goredis.ProcessPipelineHook {
	return func(_ context.Context, cmds []goredis.Cmder) error {
		for _, cmd := range cmds {
			if err := h.process(cmd); err != nil {
				return err
			}
		}
		return nil
	}
}

func TestLoginPropagatesSessionPersistenceFailures(t *testing.T) {
	hash, err := passwordutil.Hash("password-secret")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, failCommand, wantReason                              string
		nilDB, nilRedis, failAuthWrite, missingExpiry, singleLogin bool
	}{
		{name: "success"},
		{name: "nil database", nilDB: true, wantReason: "system_error"},
		{name: "nil redis", nilRedis: true, wantReason: "session_store_failed"},

		{name: "auth key write", failCommand: "set", wantReason: "session_store_failed"},
		{name: "session index write", failCommand: "sadd", wantReason: "session_store_failed"},
		{name: "session expiry write", failCommand: "expire", wantReason: "session_store_failed"},
		{name: "session index missing", missingExpiry: true, wantReason: "session_store_failed"},
		{name: "old sessions read", singleLogin: true, failCommand: "smembers", wantReason: "session_store_failed"},
		{name: "old session revoke", singleLogin: true, failCommand: "del", wantReason: "session_store_failed"},
	} {
		for _, login := range []struct {
			name string
			run  func(context.Context) (*LoginResponse, error)
		}{
			{"password", func(ctx context.Context) (*LoginResponse, error) {
				return LoginByPwdContext(ctx, "attempted-account", "password-secret", "", "")
			}},
			{"openid", func(ctx context.Context) (*LoginResponse, error) {
				return LoginUserContext(ctx, "openid-secret", "", "")
			}},
		} {
			t.Run(tc.name+"/"+login.name, func(t *testing.T) {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Fatalf("login panicked instead of returning a storage error: %v", recovered)
					}
				}()
				oldDB, oldRedis := database.DB, rd.RDB
				tokenutil.InvalidateSetupCache()
				t.Cleanup(func() { database.DB, rd.RDB = oldDB, oldRedis; tokenutil.InvalidateSetupCache() })
				ctx, recorder := auditlog.WithRecorder(context.Background())
				database.DB, rd.RDB = nil, nil
				if !tc.nilDB {
					db, err := gorm.Open(sessionTestDialector{mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true})}, &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
					if err != nil {
						t.Fatal(err)
					}
					if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
						switch dest := tx.Statement.Dest.(type) {
						case *model.User:
							*dest = model.User{ID: 7, Name: "verified-name", Password: hash, Status: 1, RoleID: 990077}
						case *[]model.Role:
							*dest = []model.Role{{ID: 990077, Name: "超级管理员", Status: 1}}
						case *model.Setup:
							if where, ok := tx.Statement.Clauses["WHERE"].Expression.(clause.Where); ok && len(where.Exprs) > 0 {
								if expr, ok := where.Exprs[0].(clause.Expr); ok && len(expr.Vars) > 0 && expr.Vars[0] == "USER_SINGLE_LOGIN" && tc.singleLogin {
									dest.Value = "1"
								}
							}
						}
					}); err != nil {
						t.Fatal(err)
					}
					if err := db.Callback().Update().Before("gorm:update").Register("test:auth_write", func(tx *gorm.DB) {
						if recorder.Snapshot().ActorID != 0 {
							t.Fatal("actor set before authentication persistence")
						}
						if values, ok := tx.Statement.Dest.(map[string]interface{}); ok && values["user_admin_token"] != nil && tc.failAuthWrite {
							tx.AddError(errors.New("private storage diagnostic"))
						}
					}); err != nil {
						t.Fatal(err)
					}
					database.DB = db
				}
				var commands []string
				if !tc.nilRedis {
					client := goredis.NewClient(&goredis.Options{})
					t.Cleanup(func() { _ = client.Close() })
					client.AddHook(sessionTestRedisHook{process: func(cmd goredis.Cmder) error {
						commands = append(commands, cmd.Name())
						if recorder.Snapshot().ActorID != 0 {
							t.Fatal("actor set before session persistence")
						}
						if cmd.Name() == tc.failCommand {
							err := errors.New("private storage diagnostic")
							cmd.SetErr(err)
							return err
						}
						switch cmd := cmd.(type) {
						case *goredis.StatusCmd:
							cmd.SetVal("OK")
						case *goredis.BoolCmd:
							cmd.SetVal(!tc.missingExpiry)
						case *goredis.IntCmd:
							cmd.SetVal(1)
						case *goredis.StringSliceCmd:
							cmd.SetVal([]string{"prior-session"})
						}
						return nil
					}})
					rd.RDB = client
				}
				result, err := login.run(ctx)
				details := recorder.Snapshot()
				if tc.wantReason != "" {
					if err == nil || result != nil {
						t.Fatal("storage failure returned login success")
					}
					if details.ReasonCode != tc.wantReason || details.ActorID != 0 || details.ActorName != "" {
						t.Fatalf("audit = %+v", details)
					}
					if strings.Contains(err.Error(), "private storage diagnostic") {
						t.Fatal("internal storage error exposed")
					}
					if (tc.failCommand == "sadd" || tc.failCommand == "expire" || tc.missingExpiry) && !strings.Contains(strings.Join(commands, ","), "del") {
						t.Fatal("partially created auth key was not removed")
					}
				} else {
					if err != nil || result == nil {
						t.Fatalf("successful storage returned error: %v", err)
					}
					if details.ActorID != 7 || details.ActorName != "verified-name" || details.ReasonCode != "" {
						t.Fatalf("audit = %+v", details)
					}
				}
				raw, _ := json.Marshal(details)
				for _, secret := range []string{"password-secret", hash, "private storage diagnostic"} {
					if strings.Contains(string(raw), secret) {
						t.Fatal("audit contains private storage or credential data")
					}
				}
			})
		}
	}
}
