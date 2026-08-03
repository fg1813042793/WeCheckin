package middleware

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"wecheckin/backend/pkg/logger"
)

func TestSafeParamMasksSensitiveFormFields(t *testing.T) {
	got := safeParam([]byte("name=alice&password=secret&token=abc123&mobile=13800138000&content=hello"))
	for _, leaked := range []string{"secret", "abc123", "13800138000", "hello"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("safeParam leaked sensitive value %q in %q", leaked, got)
		}
	}
	for _, want := range []string{"name=alice", "password=***", "token=***", "mobile=***", "content=***"} {
		if !strings.Contains(got, want) {
			t.Fatalf("safeParam() = %q, want to contain %q", got, want)
		}
	}
}

func TestSafeParamMasksSensitiveJSONFields(t *testing.T) {
	got := safeParam([]byte(`{"name":"alice","pwd":"secret","authorization":"Bearer abc","phone":"13800138000"}`))
	for _, leaked := range []string{"secret", "Bearer abc", "13800138000"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("safeParam leaked sensitive value %q in %q", leaked, got)
		}
	}
	for _, want := range []string{`"name":"alice"`, `"pwd":"***"`, `"authorization":"***"`, `"phone":"***"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("safeParam() = %q, want to contain %q", got, want)
		}
	}
}

func TestSafeParamMasksRichTextAndAnswersInJSON(t *testing.T) {
	got := safeParam([]byte(`{"title":"问卷","answers":{"q1":"非常隐私"},"schema":"<p>富文本</p>","description":"说明"}`))
	for _, leaked := range []string{"非常隐私", "<p>富文本</p>", "说明"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("safeParam leaked content value %q in %q", leaked, got)
		}
	}
	for _, want := range []string{`"title":"问卷"`, `"answers":"***"`, `"schema":"***"`, `"description":"***"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("safeParam() = %q, want to contain %q", got, want)
		}
	}
}

func TestAccessLogDetailSkipsMultipartAndLargeBody(t *testing.T) {
	multipart := accessLogDetail("POST", "", []byte("file-content"), "multipart/form-data; boundary=abc", len("file-content"))
	if !strings.Contains(multipart, "body:<skipped:multipart>") {
		t.Fatalf("multipart body must be skipped, got %q", multipart)
	}
	if strings.Contains(multipart, "file-content") {
		t.Fatalf("multipart body leaked in detail %q", multipart)
	}

	large := accessLogDetail("POST", "", []byte("large-private-body"), "application/json", maxAccessLogBodyBytes+1)
	if !strings.Contains(large, "body:<skipped:large>") {
		t.Fatalf("large body must be skipped, got %q", large)
	}
	if strings.Contains(large, "large-private-body") {
		t.Fatalf("large body leaked in detail %q", large)
	}
}

func TestAccessLogDetailHandlesUppercaseMethods(t *testing.T) {
	got := accessLogDetail("GET", "token=abc&name=alice", nil, "", 0)
	if !strings.Contains(got, "?name=alice&token=***") {
		t.Fatalf("GET query should be logged in masked normalized form, got %q", got)
	}
}

func TestSlowRequestThresholdUsesDefaultAndEnvOverride(t *testing.T) {
	t.Setenv("WECHECKIN_SLOW_REQUEST_MS", "")
	if got := slowRequestThreshold(); got != 800*time.Millisecond {
		t.Fatalf("default slow request threshold = %v, want 800ms", got)
	}

	t.Setenv("WECHECKIN_SLOW_REQUEST_MS", "1200")
	if got := slowRequestThreshold(); got != 1200*time.Millisecond {
		t.Fatalf("env slow request threshold = %v, want 1200ms", got)
	}

	t.Setenv("WECHECKIN_SLOW_REQUEST_MS", "bad")
	if got := slowRequestThreshold(); got != 800*time.Millisecond {
		t.Fatalf("invalid slow request threshold = %v, want fallback 800ms", got)
	}
}

func TestShouldLogSlowRequestHonorsThreshold(t *testing.T) {
	threshold := 800 * time.Millisecond
	if shouldLogSlowRequest(799*time.Millisecond, threshold) {
		t.Fatalf("request below threshold should not be logged as slow")
	}
	if !shouldLogSlowRequest(800*time.Millisecond, threshold) {
		t.Fatalf("request at threshold should be logged as slow")
	}
}

func TestSlowRequestLogLineContainsRouteMetadataWithoutBody(t *testing.T) {
	line := slowRequestLogLine(
		time.Date(2026, 7, 28, 12, 0, 0, 0, time.Local),
		"POST",
		"/api/v2/admin/users",
		200,
		901*time.Millisecond,
		"req-123",
	)

	for _, want := range []string{
		"[SLOW_REQUEST]",
		"200",
		"901ms",
		"requestId=req-123",
		"POST /api/v2/admin/users",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("slow request line = %q, want to contain %q", line, want)
		}
	}
	for _, forbidden := range []string{"body:", "password", "token"} {
		if strings.Contains(line, forbidden) {
			t.Fatalf("slow request line must not include sensitive body detail, got %q", line)
		}
	}
}

func TestAccessLogMiddlewareDoesNotEmitSlowLogBelowThreshold(t *testing.T) {
	t.Setenv("WECHECKIN_SLOW_REQUEST_MS", "60000")
	logs := captureAccessLogs(t, func() {
		h := server.New()
		h.Use(AccessLog())
		h.GET("/fast", func(ctx context.Context, c *app.RequestContext) {
			c.String(consts.StatusOK, "ok")
		})

		ut.PerformRequest(h.Engine, "GET", "/fast", nil).Result()
	})

	if !strings.Contains(logs, "[ACCESS]") {
		t.Fatalf("access log should be emitted, got %q", logs)
	}
	if strings.Contains(logs, "[SLOW_REQUEST]") {
		t.Fatalf("fast request should not emit slow log, got %q", logs)
	}
}

func TestAccessLogMiddlewareEmitsSlowLogAboveThreshold(t *testing.T) {
	t.Setenv("WECHECKIN_SLOW_REQUEST_MS", "1")
	logs := captureAccessLogs(t, func() {
		h := server.New()
		h.Use(AccessLog())
		h.GET("/slow", func(ctx context.Context, c *app.RequestContext) {
			time.Sleep(3 * time.Millisecond)
			c.String(consts.StatusOK, "ok")
		})

		ut.PerformRequest(h.Engine, "GET", "/slow", nil, ut.Header{Key: "X-Request-ID", Value: "req-abc"}).Result()
	})

	for _, want := range []string{"[SLOW_REQUEST]", "GET /slow", "requestId=req-abc"} {
		if !strings.Contains(logs, want) {
			t.Fatalf("slow access logs = %q, want to contain %q", logs, want)
		}
	}
	if strings.Contains(logs, "body:") {
		t.Fatalf("slow log should not include body detail, got %q", logs)
	}
}

func TestAccessLogMiddlewareSkipsUploadBody(t *testing.T) {
	t.Setenv("WECHECKIN_SLOW_REQUEST_MS", "60000")
	logs := captureAccessLogs(t, func() {
		h := server.New()
		h.Use(AccessLog())
		h.POST("/upload", func(ctx context.Context, c *app.RequestContext) {
			c.String(consts.StatusOK, "ok")
		})

		body := &ut.Body{Body: strings.NewReader("secret-file-content"), Len: len("secret-file-content")}
		ut.PerformRequest(h.Engine, "POST", "/upload", body, ut.Header{Key: "Content-Type", Value: "multipart/form-data; boundary=abc"}).Result()
	})

	if !strings.Contains(logs, "body:<skipped:multipart>") {
		t.Fatalf("upload body should be skipped, got %q", logs)
	}
	if strings.Contains(logs, "secret-file-content") {
		t.Fatalf("upload body leaked in logs %q", logs)
	}
}

func captureAccessLogs(t *testing.T, run func()) string {
	t.Helper()

	var buf bytes.Buffer
	previous := logger.Logger
	logger.Logger = log.New(&buf, "", 0)
	defer func() {
		logger.Logger = previous
	}()

	run()
	return buf.String()
}
