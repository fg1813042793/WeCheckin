package main

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

type swaggerParameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
}

type swaggerOperation struct {
	Parameters []swaggerParameter `json:"parameters"`
}

func loadSwaggerOperations(t *testing.T) map[string]map[string]swaggerOperation {
	t.Helper()

	source, err := os.ReadFile("../docs/swagger/swagger.json")
	if err != nil {
		t.Fatalf("read generated Swagger JSON: %v", err)
	}
	var document struct {
		Paths map[string]map[string]swaggerOperation `json:"paths"`
	}
	if err := json.Unmarshal(source, &document); err != nil {
		t.Fatalf("parse generated Swagger JSON: %v", err)
	}
	return document.Paths
}

func TestSwaggerPathPlaceholdersAreDocumented(t *testing.T) {
	paths := loadSwaggerOperations(t)
	placeholderPattern := regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`)

	for path, methods := range paths {
		for method, operation := range methods {
			for _, match := range placeholderPattern.FindAllStringSubmatch(path, -1) {
				if !hasSwaggerParameter(operation.Parameters, match[1], "path", true) {
					t.Errorf("Swagger operation %s %s must document required path parameter %q", method, path, match[1])
				}
			}
		}
	}
}

func TestSwaggerDocumentsRepresentativeRequestParameters(t *testing.T) {
	paths := loadSwaggerOperations(t)
	tests := []struct {
		method string
		path   string
		name   string
		in     string
	}{
		{method: "post", path: "/api/v2/admin/auth/login", name: "name", in: "formData"},
		{method: "post", path: "/api/v2/admin/auth/login", name: "password", in: "formData"},
		{method: "get", path: "/api/v2/admin/users", name: "page", in: "query"},
		{method: "get", path: "/api/v2/admin/users", name: "size", in: "query"},
		{method: "get", path: "/api/v2/admin/users", name: "keyword", in: "query"},
		{method: "post", path: "/api/v2/admin/users", name: "name", in: "formData"},
		{method: "post", path: "/api/v2/surveys/{id}/responses", name: "body", in: "body"},
		{method: "post", path: "/api/v2/enrollments/{id}/joins", name: "day", in: "formData"},
		{method: "post", path: "/api/v2/enrollments/{id}/joins", name: "forms", in: "formData"},
		{method: "post", path: "/api/v2/admin/permissions", name: "key", in: "formData"},
		{method: "post", path: "/api/v2/admin/workflow-instances", name: "body", in: "body"},
		{method: "get", path: "/api/v2/dingtalk/h5/workflows/instances", name: "definitionName", in: "query"},
		{method: "get", path: "/api/v2/dingtalk/h5/workflows/tasks", name: "definitionName", in: "query"},
		{method: "post", path: "/api/v2/admin/scheduled-tasks/cron-preview", name: "body", in: "body"},
		{method: "post", path: "/api/v2/admin/workflow-definitions", name: "logo", in: "formData"},
		{method: "get", path: "/api/v2/admin/settings/content", name: "key", in: "query"},
		{method: "post", path: "/api/v2/admin/uploads", name: "file", in: "formData"},
	}

	for _, tt := range tests {
		operation, ok := paths[tt.path][tt.method]
		if !ok {
			t.Errorf("Swagger operation %s %s is missing", tt.method, tt.path)
			continue
		}
		if !hasSwaggerParameter(operation.Parameters, tt.name, tt.in, false) {
			t.Errorf("Swagger operation %s %s must document %s parameter %q", tt.method, tt.path, tt.in, tt.name)
		}
	}
}

func TestSwaggerDocumentsWriteRequestPayloads(t *testing.T) {
	paths := loadSwaggerOperations(t)
	noPayloadActions := map[string]bool{
		"post /api/v2/admin/auth/logout":                        true,
		"post /api/v2/admin/enrollments/{id}/clear":             true,
		"post /api/v2/admin/enrollments/{id}/export":            true,
		"patch /api/v2/admin/in-app-notifications/{id}/read":    true,
		"patch /api/v2/admin/in-app-notifications/read-all":     true,
		"post /api/v2/admin/scheduled-task-runs/{id}/cancel":    true,
		"post /api/v2/admin/scheduled-task-runs/{id}/retry":     true,
		"post /api/v2/admin/scheduled-tasks/{id}/run":           true,
		"post /api/v2/admin/surveys/{id}/copy":                  true,
		"patch /api/v2/admin/users/{id}/password":               true,
		"post /api/v2/admin/workflow-definitions/{id}/validate": true,
		"post /api/v2/admin/workflow-instances/{id}/resume":     true,
		"post /api/v2/admin/workflow-notifications/{id}/retry":  true,
		"post /api/v2/dingtalk/h5/logout":                       true,
		"patch /api/v2/dingtalk/h5/notifications/read-all":      true,
		"patch /api/v2/dingtalk/h5/notifications/{id}/read":     true,
		"post /api/v2/exams/{id}/start":                         true,
		"post /api/v2/me/logout":                                true,
	}

	for path, methods := range paths {
		for method, operation := range methods {
			if method != "post" && method != "put" && method != "patch" {
				continue
			}
			key := method + " " + path
			if hasSwaggerRequestPayload(operation.Parameters) || noPayloadActions[key] {
				continue
			}
			t.Errorf("Swagger write operation %s must document a body or formData payload, or be an explicit no-payload action", key)
		}
	}
}

func hasSwaggerParameter(parameters []swaggerParameter, name, location string, mustBeRequired bool) bool {
	for _, parameter := range parameters {
		if parameter.Name == name && parameter.In == location && (!mustBeRequired || parameter.Required) {
			return true
		}
	}
	return false
}

func hasSwaggerRequestPayload(parameters []swaggerParameter) bool {
	for _, parameter := range parameters {
		if parameter.In == "body" || parameter.In == "formData" {
			return true
		}
	}
	return false
}
