package dingtalkh5

import (
	"errors"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestDingTalkH5RoutePermissionPrefersSpecificExportRoute(t *testing.T) {
	key, ok := dingTalkH5RoutePermission("GET", "/api/v2/dingtalk/h5/reviews/export")
	if !ok || key != "dingtalk_h5:api:review:export" {
		t.Fatalf("export route mapped to %q ok=%v", key, ok)
	}
}

func TestDingTalkH5RoutePermissionMapsFlowActions(t *testing.T) {
	tests := []struct {
		path string
		key  string
	}{
		{"/api/v2/dingtalk/h5/reviews/R202607/submit-self", "dingtalk_h5:api:review:self_submit"},
		{"/api/v2/dingtalk/h5/reviews/R202607/submit-hrbp", "dingtalk_h5:api:review:hrbp_submit"},
		{"/api/v2/dingtalk/h5/reviews/R202607/finalize", "dingtalk_h5:api:review:finalize"},
	}
	for _, tt := range tests {
		key, ok := dingTalkH5RoutePermission("POST", tt.path)
		if !ok || key != tt.key {
			t.Fatalf("%s mapped to %q ok=%v, want %q", tt.path, key, ok, tt.key)
		}
	}
}

func TestDingTalkH5RoutePermissionMapsTemplateSave(t *testing.T) {
	key, ok := dingTalkH5RoutePermission("PUT", "/api/v2/dingtalk/h5/template")
	if !ok || key != "dingtalk_h5:api:template:save" {
		t.Fatalf("template save route mapped to %q ok=%v", key, ok)
	}
}

func TestDingTalkH5RoutePermissionPrefersFeedbackOverviewOverDetail(t *testing.T) {
	tests := []struct {
		path string
		key  string
	}{
		{"/api/v2/dingtalk/h5/user-feedbacks/overview", "dingtalk_h5:api:feedback:list"},
		{"/api/v2/dingtalk/h5/user-feedbacks/123", "dingtalk_h5:api:feedback:detail"},
	}
	for _, test := range tests {
		key, ok := dingTalkH5RoutePermission("GET", test.path)
		if !ok || key != test.key {
			t.Fatalf("%s mapped to %q ok=%v, want %q", test.path, key, ok, test.key)
		}
	}
}

func TestDingTalkH5WorkflowTaskActionPermission(t *testing.T) {
	tests := map[string]string{
		"approve": "dingtalk_h5:api:workflow:approve",
		"reject":  "dingtalk_h5:api:workflow:reject",
		"return":  "dingtalk_h5:api:workflow:return",
		"submit":  "dingtalk_h5:api:workflow:submit",
	}
	for action, want := range tests {
		if got, ok := dingTalkH5WorkflowTaskActionPermission(action); !ok || got != want {
			t.Fatalf("action %s permission=(%q,%v), want %q", action, got, ok, want)
		}
	}
	for _, action := range []string{"", "cancel", "APPROVE"} {
		if _, ok := dingTalkH5WorkflowTaskActionPermission(action); ok {
			t.Fatalf("invalid action %q must not map to a permission", action)
		}
	}
}

func TestDingTalkH5RequestPermissionMapsWorkflowTaskActions(t *testing.T) {
	for action, want := range map[string]string{
		"approve": "dingtalk_h5:api:workflow:approve",
		"reject":  "dingtalk_h5:api:workflow:reject",
		"return":  "dingtalk_h5:api:workflow:return",
		"submit":  "dingtalk_h5:api:workflow:submit",
	} {
		c := app.NewContext(1)
		c.Request.Header.SetMethod("POST")
		c.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/tasks/task-1/complete")
		c.Request.SetBodyString(`{"action":"` + action + `"}`)
		got, err := dingTalkH5RequestPermission(c)
		if err != nil || got != want {
			t.Fatalf("action %s permission=%q err=%v, want %q", action, got, err, want)
		}
	}
}

func TestDingTalkH5RequestPermissionRejectsInvalidWorkflowTaskAction(t *testing.T) {
	tests := []struct {
		name string
		body string
		err  error
	}{
		{name: "malformed", body: `{`, err: errDingTalkH5RequestBodyInvalid},
		{name: "unsupported", body: `{"action":"cancel"}`, err: errDingTalkH5WorkflowActionInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := app.NewContext(1)
			c.Request.Header.SetMethod("POST")
			c.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/tasks/task-1/complete")
			c.Request.SetBodyString(test.body)
			if _, err := dingTalkH5RequestPermission(c); !errors.Is(err, test.err) {
				t.Fatalf("error=%v, want %v", err, test.err)
			}
		})
	}
}

func TestDingTalkH5RequestPermissionKeepsStaticRoutes(t *testing.T) {
	c := app.NewContext(1)
	c.Request.Header.SetMethod("GET")
	c.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/instances")
	got, err := dingTalkH5RequestPermission(c)
	if err != nil || got != "dingtalk_h5:api:workflow:view" {
		t.Fatalf("permission=%q err=%v", got, err)
	}
}
