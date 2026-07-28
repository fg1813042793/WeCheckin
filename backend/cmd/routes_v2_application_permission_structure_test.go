package main

import (
	"os"
	"strings"
	"testing"
)

func TestV2ClientRoutesUseClientAPIPermissionMiddleware(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"middleware.ClientAuth(), middleware.ClientPerm()",
		`client.GET("/me", pp.GetMyDetail)`,
		`client.POST("/events/:id/scores", withFormParam("event_id", "id", ev.SaveEventScore))`,
		`client.PUT("/exam-records/:id/answers", withFormParam("recordId", "id", cExam.SaveAnswer))`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("v2 client routes must use API permission middleware with %s", snippet)
		}
	}
}

func TestV2DingTalkH5RoutesUseAPIPermissionMiddleware(t *testing.T) {
	src, err := os.ReadFile("routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes_v2_dingtalk.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"handler.Auth(), handler.ApiPerm()",
		`auth.GET("/workbench", handler.Workbench)`,
		`auth.POST("/reviews/:id/finalize", withBodyOrFormParam("id", "id", handler.Finalize))`,
		`auth.PUT("/users/:id", withBodyOrFormParam("id", "id", handler.UpdateUser))`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("v2 dingtalk h5 routes must use API permission middleware with %s", snippet)
		}
	}
}
