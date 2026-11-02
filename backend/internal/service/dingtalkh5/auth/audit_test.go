package auth

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
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	rd "wecheckin/backend/pkg/redis"
)

func TestAuthenticationAuditFailureReasonsWithoutSecrets(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		run          func(context.Context) error
	}{
		{"missing password", "invalid_credentials", func(ctx context.Context) error {
			_, err := LoginContext(ctx, "attempted-account", "", "", "")
			return err
		}},
		{"missing code", "auth_code_invalid", func(ctx context.Context) error {
			_, err := loginByAuthCodeContext(ctx, fakeDingTalkIdentityClient{}, "corp", "", "", "")
			return err
		}},
		{"exchange failure", "auth_code_exchange_failed", func(ctx context.Context) error {
			_, err := loginByAuthCodeContext(ctx, fakeDingTalkIdentityClient{err: errors.New("upstream code-secret token-secret")}, "corp", "code-secret", "", "")
			return err
		}},
		{"exchange timeout", "system_error", func(ctx context.Context) error {
			_, err := loginByAuthCodeContext(ctx, fakeDingTalkIdentityClient{err: context.DeadlineExceeded}, "corp", "code-secret", "", "")
			return err
		}},
		{"invalid identity", "auth_identity_invalid", func(ctx context.Context) error {
			_, err := loginByAuthCodeContext(ctx, fakeDingTalkIdentityClient{}, "corp", "code-secret", "", "")
			return err
		}},
		{"missing bind ticket", "bind_ticket_invalid", func(ctx context.Context) error {
			_, err := BindSelfContext(ctx, "", "attempted-account", "password-secret", "", "")
			return err
		}},
		{"ticket store unavailable", "system_error", func(ctx context.Context) error {
			_, err := readDingTalkH5SelfBindTicketContext(ctx, "ticket-secret")
			return err
		}},
		{"invalid issuer", "system_error", func(ctx context.Context) error {
			_, err := issueDingTalkH5LoginResponseContext(ctx, nil, nil, "", "")
			return err
		}},
		{"session store unavailable", "system_error", func(ctx context.Context) error {
			_, err := issueDingTalkH5LoginResponseContext(ctx, nil, &model.DingTalkH5PerfUser{ID: 7, Account: "verified-account", Name: "verified-name"}, "", "")
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, recorder := auditlog.WithRecorder(context.Background())
			if err := tc.run(ctx); err == nil {
				t.Fatal("expected failure")
			}
			details := recorder.Snapshot()
			if details.ReasonCode != tc.reason || details.Reason == "" {
				t.Fatalf("reason = %+v", details)
			}
			if details.ActorID != 0 || details.ActorName != "" || details.ActorType != "" {
				t.Fatalf("unverified actor = %+v", details)
			}
			raw, _ := json.Marshal(details)
			for _, secret := range []string{"code-secret", "token-secret", "ticket-secret", "password-secret", "attempted-account"} {
				if strings.Contains(string(raw), secret) {
					t.Fatal("audit contains credentials or an unverified account")
				}
			}
		})
	}
}

// Intercept every command so the tests never dial a Redis server.
type auditRedisHook struct{ process func(goredis.Cmder) error }

func (h auditRedisHook) DialHook(next goredis.DialHook) goredis.DialHook { return next }
func (h auditRedisHook) ProcessHook(_ goredis.ProcessHook) goredis.ProcessHook {
	return func(_ context.Context, cmd goredis.Cmder) error { return h.process(cmd) }
}
func (h auditRedisHook) ProcessPipelineHook(_ goredis.ProcessPipelineHook) goredis.ProcessPipelineHook {
	return func(_ context.Context, cmds []goredis.Cmder) error {
		for _, cmd := range cmds {
			if err := h.process(cmd); err != nil {
				return err
			}
		}
		return nil
	}
}

func TestLoginIssuerAuditsActorOnlyAfterSessionAndLoginRecordSucceed(t *testing.T) {
	for _, failRecord := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "record failure"}[failRecord], func(t *testing.T) {
			ctx, recorder := auditlog.WithRecorder(context.Background())
			client := goredis.NewClient(&goredis.Options{})
			client.AddHook(auditRedisHook{process: func(cmd goredis.Cmder) error {
				if recorder.Snapshot().ActorID != 0 {
					t.Fatal("actor recorded before session was accepted")
				}
				return nil
			}})
			previous := rd.RDB
			rd.RDB = client
			t.Cleanup(func() { rd.RDB = previous; _ = client.Close() })
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Update().Before("gorm:update").Register("audit:record", func(tx *gorm.DB) {
				if recorder.Snapshot().ActorID != 0 {
					t.Fatal("actor recorded before login update succeeded")
				}
				if failRecord {
					tx.AddError(errors.New("database token-secret"))
				}
			}); err != nil {
				t.Fatal(err)
			}
			result, err := issueDingTalkH5LoginResponseContext(ctx, db, &model.DingTalkH5PerfUser{ID: 7, Account: "account", Name: "verified-name"}, "", "")
			details := recorder.Snapshot()
			if failRecord {
				if err == nil || result != nil || details.ActorID != 0 || details.ReasonCode != "system_error" {
					t.Fatalf("failed issuance audit = %+v, err = %v", details, err)
				}
				return
			}
			if err != nil || result == nil {
				t.Fatalf("issuance failed: %v", err)
			}
			if details.ActorID != 7 || details.ActorName != "verified-name" || details.ActorType != "user" || details.ReasonCode != "" {
				t.Fatalf("actor = %+v", details)
			}
			raw, _ := json.Marshal(details)
			if strings.Contains(string(raw), result.Token) {
				t.Fatal("issued token leaked into audit")
			}
		})
	}
}

func TestBindTicketAuditDistinguishesInvalidAndUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name, payload, reason string
		err                   error
	}{
		{"missing", "", "bind_ticket_invalid", goredis.Nil},
		{"expired", `{"corpId":"corp","dingTalkUserId":"identity-secret","expiresAt":1}`, "bind_ticket_invalid", nil},
		{"malformed", "ticket-secret", "bind_ticket_invalid", nil},
		{"unavailable", "", "system_error", errors.New("redis token-secret")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := goredis.NewClient(&goredis.Options{})
			client.AddHook(auditRedisHook{process: func(cmd goredis.Cmder) error {
				cmd.(*goredis.StringCmd).SetVal(tc.payload)
				cmd.SetErr(tc.err)
				return tc.err
			}})
			previous := rd.RDB
			rd.RDB = client
			t.Cleanup(func() { rd.RDB = previous; _ = client.Close() })
			ctx, recorder := auditlog.WithRecorder(context.Background())
			_, err := readDingTalkH5SelfBindTicketContext(ctx, "ticket-secret")
			if err == nil || recorder.Snapshot().ReasonCode != tc.reason {
				t.Fatalf("ticket audit = %+v, err = %v", recorder.Snapshot(), err)
			}
			raw, _ := json.Marshal(recorder.Snapshot())
			for _, secret := range []string{"ticket-secret", "token-secret", "identity-secret"} {
				if strings.Contains(string(raw), secret) {
					t.Fatal("ticket credentials leaked into audit")
				}
			}
		})
	}
}
