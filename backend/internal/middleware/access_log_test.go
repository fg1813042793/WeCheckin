package middleware

import (
	"strings"
	"testing"
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
