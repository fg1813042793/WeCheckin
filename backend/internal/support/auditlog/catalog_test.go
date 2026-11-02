package auditlog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestAuditedOperationsCoverPrioritizedModules(t *testing.T) {
	tests := []struct{ method, path, source, module, action string }{
		{"POST", "/api/v2/admin/auth/login", "admin", "auth", "password_login"},
		{"POST", "/api/v2/dingtalk/h5/login", "h5app", "auth", "password_login"},
		{"POST", "/api/v2/dingtalk/h5/bind-self", "h5app", "auth", "bind"},
		{"POST", "/api/v2/dingtalk/h5/sso-login", "h5app", "auth", "sso_login"},
		{"POST", "/api/v2/auth/password-login", "client", "auth", "password_login"},
		{"POST", "/passport/login_pwd", "client", "auth", "password_login"},
		{"PUT", "/api/v2/admin/users/12", "admin", "user", "update"},
		{"PUT", "/api/v2/admin/roles/2", "admin", "role", "update"},
		{"PUT", "/api/v2/admin/permissions/admin%3Amenu%3Auser", "admin", "permission", "update"},
		{"PUT", "/api/v2/admin/workflow-org-approver-assignments", "admin", "permission", "update"},
		{"POST", "/api/v2/admin/user-sessions/batch-force-offline", "admin", "auth", "force_offline"},
		{"GET", "/api/v2/admin/surveys/1/responses/export", "admin", "survey", "export"},
		{"DELETE", "/api/v2/admin/survey-responses", "admin", "survey", "delete"},
		{"PATCH", "/api/v2/admin/exams/2/status", "admin", "exam", "status"},
		{"POST", "/api/v2/admin/workflow-definitions/1/publish", "admin", "workflow", "publish"},
		{"POST", "/api/v2/admin/scheduled-tasks/1/run", "admin", "scheduler", "run"},
		{"DELETE", "/api/v2/admin/scheduled-task-runs", "admin", "scheduler", "delete"},
		{"PUT", "/api/v2/admin/dingtalk/settings", "admin", "settings", "update"},
		{"POST", "/api/v2/admin/in-app-notifications", "admin", "notification", "send"},
		{"POST", "/api/v2/admin/dingtalk-notification-templates", "admin", "notification", "create"},
		{"PUT", "/api/v2/admin/dingtalk-notification-templates/2", "admin", "notification", "update"},
		{"DELETE", "/api/v2/admin/dingtalk-notification-templates/2", "admin", "notification", "delete"},
		{"PATCH", "/api/v2/admin/user-feedbacks/1/status", "admin", "feedback", "status"},
		{"POST", "/api/v2/admin/user-feedbacks/1/notify", "admin", "feedback", "send"},
		{"POST", "/api/v2/admin/departments", "admin", "organization", "create"},
		{"POST", "/api/v2/admin/positions", "admin", "organization", "create"},
		{"PUT", "/api/v2/admin/dict/types/event_type", "admin", "dictionary", "update"},
		{"POST", "/api/v2/dingtalk/h5/workflows/tasks/task-1/complete", "h5app", "workflow", "complete"},
		{"POST", "/api/v2/dingtalk/h5/reviews/1/submit-self", "h5app", "performance", "submit-self"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			op, ok := Classify(tt.method, tt.path)
			if !ok || op.Source != tt.source || op.Module != tt.module || op.Action != tt.action {
				t.Fatalf("operation = %#v, tracked=%v", op, ok)
			}
		})
	}
}

func TestOrdinaryReadsAndComputationsAreNotAudited(t *testing.T) {
	for _, test := range [][2]string{{"GET", "/api/v2/admin/users"}, {"GET", "/api/v2/admin/logs"}, {"GET", "/api/v2/admin/auth/captcha"}, {"POST", "/api/v2/survey/validate"}, {"POST", "/api/v2/admin/scheduled-tasks/cron-preview"}, {"GET", "/health"}, {"OPTIONS", "/api/v2/admin/users"}} {
		if _, ok := Classify(test[0], test[1]); ok {
			t.Fatalf("unexpected audit: %v", test)
		}
	}
}

func TestAuditChangesExcludeSecretsAndBoundNestedData(t *testing.T) {
	data := SafeChanges([]Change{
		{Field: "password", Before: "old-secret", After: "new-secret"},
		{Field: "settings", Before: map[string]interface{}{"appSecret": "hidden-secret", "captchaCode": "1234"}, After: map[string]interface{}{"name": "ok"}},
		{Field: "roleIds", Before: []interface{}{float64(1)}, After: []interface{}{float64(2)}},
	})
	for _, secret := range []string{"old-secret", "new-secret", "hidden-secret", "1234"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("secret leaked: %s", data)
		}
	}
	if !json.Valid(data) || !strings.Contains(string(data), `"roleIds"`) {
		t.Fatalf("invalid changes: %s", data)
	}
	if got := CleanText("one\nline\rtest", 7); got != "one lin" {
		t.Fatalf("clean = %q", got)
	}
}

func TestAuditContextIsOptionalAndKeepsVerifiedActors(t *testing.T) {
	SetActor(context.Background(), 1, "ignored", "admin")
	ctx, recorder := WithRecorder(context.Background())
	SetActor(ctx, 7, "operator", "user")
	SetReason(ctx, "invalid_credentials", "账号或密码错误")
	got := recorder.Snapshot()
	if got.ActorID != 7 || got.ActorType != "user" || got.ReasonCode != "invalid_credentials" {
		t.Fatalf("details = %#v", got)
	}
}

func TestResourceNamesCannotSuppressMutationAudit(t *testing.T) {
	for _, method := range []string{"PUT", "DELETE"} {
		for _, path := range []string{"/api/v2/admin/dict/types/preview", "/api/v2/admin/permissions/read", "/api/v2/admin/dict/types/validate"} {
			if _, tracked := Classify(method, path); !tracked {
				t.Fatalf("mutation was skipped: %s %s", method, path)
			}
		}
	}
}

func TestSafeChangesPreservesTypedGrantAndScopeLists(t *testing.T) {
	before := []map[string]interface{}{{"permissionKey": "admin:api:user:list", "effect": "allow", "scope": []uint64{1, 2}}}
	after := []map[string]interface{}{{"permissionKey": "admin:api:user:list", "effect": "deny", "scope": []uint64{3}}}
	data := SafeChanges([]Change{{Field: "role.grants", Before: before, After: after}})
	for _, want := range []string{`"allow"`, `"deny"`, `[1,2]`, `[3]`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("grant detail missing %s: %s", want, data)
		}
	}
}

func TestGrantChangesAfterHundredthItemArePreserved(t *testing.T) {
	before, after := make([]map[string]interface{}, 101), make([]map[string]interface{}, 101)
	for i := range before {
		key := fmt.Sprintf("permission-%03d", i)
		before[i] = map[string]interface{}{"key": key, "effect": "allow"}
		after[i] = map[string]interface{}{"key": key, "effect": "allow"}
	}
	after[100]["effect"] = "deny"
	data, err := EncodeChanges([]Change{{Field: "role.grants", Before: before, After: after}})
	if err != nil || !strings.Contains(string(data), `"deny"`) {
		t.Fatalf("last permission silently lost: %s, %v", data, err)
	}
	if _, err := EncodeChanges([]Change{{Field: "role.grants", After: make([]string, 1025)}}); err == nil {
		t.Fatal("oversize relation must report an error")
	}
}
