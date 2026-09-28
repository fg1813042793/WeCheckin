package common

import (
	"encoding/json"
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
		`"/swagger/doc.json?ts=" + Date.now()`,
		`url: swaggerDocumentURL`,
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
	if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "no-store" {
		t.Errorf("Swagger index Cache-Control = %q, want no-store", cacheControl)
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
	var definition struct {
		Swagger string `json:"swagger"`
		OpenAPI string `json:"openapi"`
	}
	if err := json.Unmarshal([]byte(body), &definition); err != nil {
		t.Fatalf("Swagger document is not valid JSON: %v", err)
	}
	if definition.Swagger != "2.0" && !strings.HasPrefix(definition.OpenAPI, "3.0.") {
		t.Fatalf("Swagger document version = swagger:%q openapi:%q", definition.Swagger, definition.OpenAPI)
	}
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
	if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "no-store" {
		t.Errorf("Swagger document Cache-Control = %q, want no-store", cacheControl)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Swagger document Content-Type = %q, want application/json", contentType)
	}
}

func TestValidSwaggerDefinitionRequiresSupportedVersion(t *testing.T) {
	tests := []struct {
		name      string
		document  string
		wantValid bool
	}{
		{name: "Swagger 2", document: `{"swagger":"2.0"}`, wantValid: true},
		{name: "OpenAPI 3", document: `{"openapi":"3.0.3"}`, wantValid: true},
		{name: "missing version", document: `{"code":0,"data":{}}`, wantValid: false},
		{name: "unsupported OpenAPI", document: `{"openapi":"3.1.0"}`, wantValid: false},
		{name: "invalid JSON", document: `<html></html>`, wantValid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validSwaggerDefinition(test.document); got != test.wantValid {
				t.Fatalf("validSwaggerDefinition() = %t, want %t", got, test.wantValid)
			}
		})
	}
}
