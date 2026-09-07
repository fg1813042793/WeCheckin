package common

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	_ "wecheckin/backend/docs/swagger"
)

func TestSwaggerIndexEnablesNativeSearch(t *testing.T) {
	h := server.New()
	RegisterSwagger(h)

	response := ut.PerformRequest(h.Engine, http.MethodGet, "/swagger/index.html", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /swagger/index.html status = %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	for _, expected := range []string{
		`<title>WeCheckin API</title>`,
		`url: "/swagger/doc.json"`,
		`filter: true`,
		`SwaggerUIBundle.presets.apis`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("Swagger index missing %q", expected)
		}
	}
	for _, placeholder := range []string{"{{.Title}}", "{{.Version}}", "{{.Host}}", "{{escape .Description}}"} {
		if strings.Contains(body, placeholder) {
			t.Errorf("Swagger index contains unresolved placeholder %q", placeholder)
		}
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Swagger index content type = %q, want text/html; charset=utf-8", contentType)
	}
}

func TestSwaggerDocumentRouteRemainsAvailable(t *testing.T) {
	h := server.New()
	RegisterSwagger(h)

	response := ut.PerformRequest(h.Engine, http.MethodGet, "/swagger/doc.json", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /swagger/doc.json status = %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	if !strings.Contains(body, `"title": "WeCheckin API"`) {
		t.Error("Swagger document is missing the rendered API title")
	}
	for _, placeholder := range []string{"{%escape .Description%}", "{%.Title%}", "{%.Version%}", "{%.Host%}", "{%.BasePath%}"} {
		if strings.Contains(body, placeholder) {
			t.Errorf("Swagger document contains unresolved template placeholder %q", placeholder)
		}
	}
	if !strings.Contains(body, `"example": "{{content}}"`) {
		t.Error("Swagger document must preserve workflow notification template examples")
	}
}
