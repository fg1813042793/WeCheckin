package middleware

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/response"
)

type auditMemoryWriter struct {
	events []model.Log
	err    error
}

func (w *auditMemoryWriter) Write(_ context.Context, event *model.Log) error {
	w.events = append(w.events, *event)
	return w.err
}

func TestAuditCollectorRecordsBusinessFailuresDespiteHTTP200(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) { t.Fatal("unexpected fallback") }))
	h.POST("/api/v2/admin/auth/login", func(ctx context.Context, c *app.RequestContext) {
		auditlog.SetReason(ctx, "invalid_credentials", "账号或密码错误")
		response.Fail(c, "登录失败")
	})
	body := `{"name":"attempted-user","password":"never-store-this","captchaCode":"123456"}`
	result := ut.PerformRequest(h.Engine, "POST", "/api/v2/admin/auth/login", &ut.Body{Body: bytes.NewBufferString(body), Len: len(body)}, ut.Header{Key: "Content-Type", Value: "application/json"})
	if result.Code != 200 || len(w.events) != 2 {
		t.Fatalf("status=%d, events=%d", result.Code, len(w.events))
	}
	got := w.events[1]
	if got.Outcome != "failure" || got.ReasonCode != "invalid_credentials" || got.AttemptedAccount != "attempted-user" || got.AdminID != 0 || got.ActorType != "anonymous" {
		t.Fatalf("event=%#v", got)
	}
	if w.events[0].Outcome != "unknown" || got.EventID != w.events[0].EventID {
		t.Fatal("attempt/finalization must share a single event ID")
	}
	if strings.Contains(got.Content+got.Reason+string(got.Changes), "never-store-this") || strings.Contains(got.Content+got.Reason+string(got.Changes), "123456") {
		t.Fatal("credentials leaked")
	}
}

func TestAuditCollectorCapturesPermissionDenialsBeforeHandlers(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Set("admin", &model.Admin{ID: 7, Name: "auditor"})
		response.Fail(c, "无权限访问")
		c.Abort()
	})
	h.DELETE("/api/v2/admin/users/:id", func(context.Context, *app.RequestContext) { t.Fatal("must not reach handler") })
	ut.PerformRequest(h.Engine, "DELETE", "/api/v2/admin/users/12", nil)
	got := w.events[len(w.events)-1]
	if got.Outcome != "blocked" || got.AdminID != 7 || got.TargetID != "12" {
		t.Fatalf("event=%#v", got)
	}
}

func TestAuditCollectorRecognizesDetailedAdminPermissionFailures(t *testing.T) {
	for _, message := range []string{"缺少权限-admin:api:menu:list（菜单查看接口）", "接口未配置权限规则，请联系管理员"} {
		w := &auditMemoryWriter{}
		h := server.New()
		h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
		h.Use(func(_ context.Context, c *app.RequestContext) {
			c.Set("admin", &model.Admin{ID: 7})
			response.Fail(c, message)
			c.Abort()
		})
		h.GET("/api/v2/admin/permissions/tree", func(context.Context, *app.RequestContext) { t.Fatal("must not reach handler") })
		ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/permissions/tree", nil)
		if len(w.events) != 1 || w.events[0].Outcome != "blocked" || w.events[0].ReasonCode != "permission_denied" {
			t.Fatalf("permission error %q lost blocked audit: %#v", message, w.events)
		}
	}
}

func TestAuditCollectorPreservesAsyncAndPartialOutcomes(t *testing.T) {
	for _, test := range []struct {
		path    string
		data    interface{}
		outcome string
	}{
		{"/api/v2/admin/scheduled-tasks/1/run", map[string]interface{}{"dispatchPending": true}, "accepted"},
		{"/api/v2/admin/dingtalk-notifications", map[string]interface{}{"sentCount": 1, "failedCount": 1}, "partial"},
		{"/api/v2/admin/dingtalk-notifications", map[string]interface{}{"sentCount": 0, "failedCount": 2}, "failure"},
	} {
		t.Run(test.path+test.outcome, func(t *testing.T) {
			w := &auditMemoryWriter{}
			h := server.New()
			h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
			h.POST(test.path, func(ctx context.Context, c *app.RequestContext) { response.JSON(c, test.data) })
			ut.PerformRequest(h.Engine, "POST", test.path, nil)
			if got := w.events[len(w.events)-1].Outcome; got != test.outcome {
				t.Fatalf("outcome=%s", got)
			}
		})
	}
}

func TestAuditCollectorOnlyKeepsChangesForCommittedSuccess(t *testing.T) {
	for _, success := range []bool{true, false} {
		w := &auditMemoryWriter{}
		h := server.New()
		h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
		h.PUT("/api/v2/admin/roles/:id", func(ctx context.Context, c *app.RequestContext) {
			auditlog.SetChanges(ctx, []auditlog.Change{{Field: "dataScope", Before: 1, After: 2}})
			if success {
				response.JSON(c, nil)
			} else {
				response.Fail(c, "修改失败")
			}
		})
		ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/roles/1", nil)
		got := w.events[len(w.events)-1]
		if success != strings.Contains(string(got.Changes), "dataScope") {
			t.Fatalf("success=%v changes=%s", success, got.Changes)
		}
	}
}

func TestAuditWriteFailureFallsBackWithoutSilentlyLosingEvent(t *testing.T) {
	w := &auditMemoryWriter{err: errors.New("db down")}
	fallbacks := 0
	h := server.New()
	h.Use(auditLogWithWriter(w, func(e *model.Log, stage string) {
		fallbacks++
		if e.EventID == "" {
			t.Fatal("missing event ID")
		}
	}))
	h.POST("/api/v2/admin/auth/login", func(ctx context.Context, c *app.RequestContext) { response.Fail(c, "登录失败") })
	ut.PerformRequest(h.Engine, "POST", "/api/v2/admin/auth/login", nil)
	if fallbacks != 2 {
		t.Fatalf("fallbacks=%d", fallbacks)
	}
}

func TestAuditCollectorSkipsNormalReadsButAuditsDeniedReads(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.GET("/api/v2/admin/users", func(ctx context.Context, c *app.RequestContext) {
		if c.Query("deny") == "1" {
			c.JSON(consts.StatusForbidden, map[string]interface{}{"code": 1})
			return
		}
		response.JSON(c, []string{})
	})
	ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/users", nil)
	if len(w.events) != 0 {
		t.Fatal("normal reads should not produce audit noise")
	}
	ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/users?deny=1", nil)
	if len(w.events) != 1 || w.events[0].Outcome != "blocked" {
		t.Fatalf("events=%#v", w.events)
	}
}

func TestAccessLogMasksCaptchaAndBindingSecrets(t *testing.T) {
	for _, body := range []string{`{"captchaCode":"solved-answer","bindTicket":"secret-ticket","appSecret":"secret-config"}`, `captchaCode=solved-answer&bindTicket=secret-ticket&appSecret=secret-config`} {
		got := maskSensitiveParam(body)
		for _, secret := range []string{"solved-answer", "secret-ticket", "secret-config"} {
			if strings.Contains(got, secret) {
				t.Fatalf("sensitive access log: %s", got)
			}
		}
	}
}

func TestMalformedAndNestedAccessLogBodiesDoNotLeakSecrets(t *testing.T) {
	for _, body := range []string{`{"password":"private-value"`, `[[{"captchaCode":"private-value"}]]`, `token=private-value&bad=%ZZ`} {
		if got := maskSensitiveParam(body); strings.Contains(got, "private-value") {
			t.Fatalf("leaked malformed body: %s", got)
		}
	}
}

func TestExplicitExportRecordsReturnedRowsAndRejectsOversize(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.GET("/api/v2/admin/users/export", AuditExportPage(), func(ctx context.Context, c *app.RequestContext) {
		response.JSON(c, map[string]interface{}{"list": []map[string]int{{"id": 1}, {"id": 2}}, "total": 99})
	})
	ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/users/export?pageSize=20", nil)
	if last := w.events[len(w.events)-1]; last.AffectedCount != 2 || last.Action != "export" || last.Outcome != "success" || string(last.TargetIDs) != `["1","2"]` {
		t.Fatalf("export event=%#v", last)
	}
	ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/users/export?pageSize=999999", nil)
	if last := w.events[len(w.events)-1]; last.Outcome != "failure" {
		t.Fatalf("oversized export result=%s", last.Outcome)
	}
}

func TestAuditRequestIDCannotBeForged(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.POST("/api/v2/admin/users", func(ctx context.Context, c *app.RequestContext) { response.JSON(c, nil) })
	r := ut.PerformRequest(h.Engine, "POST", "/api/v2/admin/users", nil, ut.Header{Key: "X-Request-ID", Value: "forged"})
	got := w.events[len(w.events)-1].RequestID
	if got == "forged" || !strings.HasPrefix(got, "req-") || string(r.Header().Peek("X-Request-ID")) != got {
		t.Fatalf("requestId=%s", got)
	}
}

func TestBatchTargetIDsAreSeparateFromAffectedCount(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.DELETE("/api/v2/admin/survey-responses", func(ctx context.Context, c *app.RequestContext) { response.JSON(c, map[string]int{"deleted": 1}) })
	body := `{"ids":["11","12"]}`
	ut.PerformRequest(h.Engine, "DELETE", "/api/v2/admin/survey-responses", &ut.Body{Body: bytes.NewBufferString(body), Len: len(body)}, ut.Header{Key: "Content-Type", Value: "application/json"})
	got := w.events[len(w.events)-1]
	if got.RequestedCount != 2 || got.AffectedCount != 1 || string(got.TargetIDs) != `["11","12"]` {
		t.Fatalf("batch=%#v", got)
	}
}

func TestNotificationDiagnosisFailureIsNotAuditSuccess(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.POST("/api/v2/admin/dingtalk/settings/notification-test", func(ctx context.Context, c *app.RequestContext) {
		response.JSON(c, map[string]interface{}{"success": false, "message": "provider rejected test"})
	})
	ut.PerformRequest(h.Engine, "POST", "/api/v2/admin/dingtalk/settings/notification-test", nil)
	if got := w.events[len(w.events)-1]; got.Outcome != "failure" || got.ReasonCode != "notification_test_failed" {
		t.Fatalf("diagnosis was marked successful: %#v", got)
	}
}

func TestAnonymousBlockedMutationIsVisibleInSecurityCategory(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.DELETE("/api/v2/admin/users/:id", func(ctx context.Context, c *app.RequestContext) { response.Fail(c, "未登录") })
	ut.PerformRequest(h.Engine, "DELETE", "/api/v2/admin/users/5", nil)
	got := w.events[len(w.events)-1]
	if got.Category != "login" || got.Action != "access_denied" || got.AdminID != 0 || got.Outcome != "blocked" {
		t.Fatalf("anonymous failure hidden by operation scope: %#v", got)
	}
}

func TestAccessLogSkipsOpaqueSecretConfigurationValues(t *testing.T) {
	logs := captureAccessLogs(t, func() {
		h := server.New()
		h.Use(AccessLog())
		h.PUT("/api/v2/admin/settings/content", func(ctx context.Context, c *app.RequestContext) { response.JSON(c, nil) })
		body := `key=integration_config&value=opaque-secret-value`
		ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/settings/content", &ut.Body{Body: bytes.NewBufferString(body), Len: len(body)}, ut.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"})
	})
	if strings.Contains(logs, "opaque-secret-value") || !strings.Contains(logs, "redacted:sensitive-operation") {
		t.Fatalf("configuration leaked to access log: %s", logs)
	}
}

func TestLoginAuditIncludesDeviceMetadata(t *testing.T) {
	w := &auditMemoryWriter{}
	h := server.New()
	h.Use(auditLogWithWriter(w, func(*model.Log, string) {}))
	h.POST("/api/v2/dingtalk/h5/login", func(ctx context.Context, c *app.RequestContext) { response.Fail(c, "登录失败") })
	ut.PerformRequest(h.Engine, "POST", "/api/v2/dingtalk/h5/login", nil, ut.Header{Key: "User-Agent", Value: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 DingTalk/7.6.0"})
	got := w.events[len(w.events)-1]
	if got.Browser != "钉钉" || got.DeviceType != "移动端" || got.PlatformType != "iOS" {
		t.Fatalf("device metadata=%#v", got)
	}
}

func TestAccessLogOmitsNotificationTemplateBody(t *testing.T) {
	logs := captureAccessLogs(t, func() {
		h := server.New()
		h.Use(AccessLog())
		h.POST("/api/v2/admin/dingtalk-notification-templates", func(ctx context.Context, c *app.RequestContext) { response.JSON(c, nil) })
		body := `{"name":"提醒模板","content":"private-message-body","target":{"kind":"url","url":"https://example.test/?invite=private-link-value"}}`
		ut.PerformRequest(h.Engine, "POST", "/api/v2/admin/dingtalk-notification-templates", &ut.Body{Body: bytes.NewBufferString(body), Len: len(body)}, ut.Header{Key: "Content-Type", Value: "application/json"})
	})
	if strings.Contains(logs, "private-message-body") || strings.Contains(logs, "private-link-value") || !strings.Contains(logs, "redacted:sensitive-operation") {
		t.Fatalf("template body leaked to access log: %s", logs)
	}
}

func TestOversizeAuditChangesAreExplicitlyMarked(t *testing.T) {
	event := &model.Log{Outcome: "success"}
	applyAuditDetails(event, auditlog.Details{Changes: []auditlog.Change{{Field: "role.grants", After: make([]string, 1025)}}})
	if event.Outcome != "success" || event.SnapshotStatus != auditlog.SnapshotUnavailable || event.SnapshotReason != "encoding_limit" || event.ReasonCode != "" || string(event.Changes) != "[]" {
		t.Fatalf("oversize changes silently lost: %#v", event)
	}
}
