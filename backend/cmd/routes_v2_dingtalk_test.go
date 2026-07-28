package main

import (
	"os"
	"strings"
	"testing"
)

func TestV2DingTalkH5RoutesAreIsolated(t *testing.T) {
	src, err := os.ReadFile("routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes_v2_dingtalk.go: %v", err)
	}
	text := string(src)
	required := []string{
		`h.Group("/api/v2/dingtalk/h5")`,
		`group.POST("/login", handler.Login)`,
		`auth.GET("/bootstrap", handler.Bootstrap)`,
		`auth.GET("/reviews/export", handler.ExportReviews)`,
		`auth.POST("/reviews/:id/submit-self", withBodyOrFormParam("id", "id", handler.SubmitSelf))`,
		`auth.POST("/reviews/:id/finalize", withBodyOrFormParam("id", "id", handler.Finalize))`,
		`auth.GET("/template", handler.Template)`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("routes_v2_dingtalk.go missing %s", want)
		}
	}
	if strings.Contains(text, "AdminAuth") || strings.Contains(text, "ClientAuth") {
		t.Fatalf("dingtalk h5 routes must use the module token guard, not existing admin/client middleware")
	}
}

func TestV2RouteSuiteRegistersDingTalkH5Routes(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	if !strings.Contains(string(src), "registerV2DingTalkH5Routes(h)") {
		t.Fatalf("registerV2Routes must register dingtalk h5 route suite")
	}
}
