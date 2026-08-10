package main

import (
	"os"
	"strings"
	"testing"
)

func TestV2ClientRoutesUseClientAPIPermissionMiddleware(t *testing.T) {
	src, err := os.ReadFile("../internal/routes/v2/client/routes.go")
	if err != nil {
		t.Fatalf("read client v2 routes: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"clientmw.ClientAuth(), clientmw.ClientPerm()",
		`client.GET("/me", pp.GetMyDetail)`,
		`client.POST("/events/:id/scores", routeparam.WithFormParam("event_id", "id", ev.SaveEventScore))`,
		`client.PUT("/exam-records/:id/answers", routeparam.WithFormParam("recordId", "id", cExam.SaveAnswer))`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("v2 client routes must use API permission middleware with %s", snippet)
		}
	}
}

func TestV2DingTalkH5RoutesUseAPIPermissionMiddleware(t *testing.T) {
	src, err := os.ReadFile("../internal/routes/v2/dingtalkh5/routes.go")
	if err != nil {
		t.Fatalf("read dingtalk h5 v2 routes: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"dingtalkh5mw.DingTalkH5Auth(), dingtalkh5mw.DingTalkH5Perm()",
		`auth.GET("/workbench", handler.Bootstrap.Workbench)`,
		`auth.POST("/reviews/:id/finalize", routeparam.WithBodyOrFormParam("id", "id", handler.Review.Finalize))`,
		`auth.PUT("/users/:id", routeparam.WithBodyOrFormParam("id", "id", handler.User.UpdateUser))`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("v2 dingtalk h5 routes must use API permission middleware with %s", snippet)
		}
	}
}
