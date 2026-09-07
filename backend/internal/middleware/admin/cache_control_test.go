package admin

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestNoStorePreventsCachingSuccessfulResponses(t *testing.T) {
	h := server.New()
	h.Use(NoStore())
	h.GET("/api/v2/admin/example", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "ok")
	})

	resp := ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/example", nil).Result()
	if got := string(resp.Header.Peek("Cache-Control")); got != "private, no-store" {
		t.Fatalf("Cache-Control = %q, want %q", got, "private, no-store")
	}
}

func TestNoStoreHeaderSurvivesAbortedResponses(t *testing.T) {
	h := server.New()
	h.Use(NoStore())
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{"code": 1, "msg": "未登录"})
		c.Abort()
	})
	h.GET("/api/v2/admin/example", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "unexpected")
	})

	resp := ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/example", nil).Result()
	if got := string(resp.Header.Peek("Cache-Control")); got != "private, no-store" {
		t.Fatalf("Cache-Control = %q, want %q", got, "private, no-store")
	}
}
