package response

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestFailInternalHidesPrivateErrorAndLogsDiagnosticContext(t *testing.T) {
	oldLogf := internalErrorLogf
	var logged string
	internalErrorLogf = func(format string, args ...interface{}) {
		logged = fmt.Sprintf(format, args...)
	}
	t.Cleanup(func() { internalErrorLogf = oldLogf })
	c := app.NewContext(1)
	c.Request.Header.Set("X-Request-ID", "req-123")
	privateErr := fmt.Errorf("load workflow: %w", errors.New("Error 1064: syntax error near password='secret' at /Users/dev/service.go:42"))

	FailInternal(context.Background(), c, "workflow.start", "流程操作失败，请稍后重试", privateErr)

	body := string(c.Response.Body())
	if !strings.Contains(body, "流程操作失败，请稍后重试") {
		t.Fatalf("response body = %q, want public message", body)
	}
	for _, forbidden := range []string{"Error 1064", "secret", "/Users/dev", "load workflow"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response body leaked %q: %s", forbidden, body)
		}
	}
	for _, want := range []string{"operation=workflow.start", "request_id=req-123", "load workflow", "Error 1064", "secret"} {
		if !strings.Contains(logged, want) {
			t.Fatalf("internal log = %q, want %q", logged, want)
		}
	}
}

func TestFailInternalUsesStableDefaultsAndTraceIDFallback(t *testing.T) {
	oldLogf := internalErrorLogf
	var logged string
	internalErrorLogf = func(format string, args ...interface{}) {
		logged = fmt.Sprintf(format, args...)
	}
	t.Cleanup(func() { internalErrorLogf = oldLogf })
	c := app.NewContext(1)
	c.Request.Header.Set("X-Trace-ID", "trace-456")

	FailInternal(context.Background(), c, "", "", errors.New("private"))

	if body := string(c.Response.Body()); !strings.Contains(body, "服务异常，请稍后重试") || strings.Contains(body, "private") {
		t.Fatalf("response body = %q", body)
	}
	if !strings.Contains(logged, "operation=unknown") || !strings.Contains(logged, "request_id=trace-456") {
		t.Fatalf("internal log = %q", logged)
	}
}
