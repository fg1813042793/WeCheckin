package main

import "testing"

func TestSwaggerDocumentsWorkflowSearchFilters(t *testing.T) {
	paths := loadSwaggerOperations(t)
	if _, ok := paths["/api/v2/dingtalk/h5/workflows/categories"]["get"]; !ok {
		t.Error("Swagger must document GET /api/v2/dingtalk/h5/workflows/categories")
	}
	tests := []struct {
		path string
		name string
	}{
		{path: "/api/v2/dingtalk/h5/workflows/instances", name: "starterName"},
		{path: "/api/v2/dingtalk/h5/workflows/tasks", name: "definitionCategory"},
		{path: "/api/v2/dingtalk/h5/workflows/tasks", name: "starterName"},
		{path: "/api/v2/dingtalk/h5/workflows/tasks", name: "startTimeFrom"},
		{path: "/api/v2/dingtalk/h5/workflows/tasks", name: "startTimeTo"},
	}
	for _, test := range tests {
		operation, ok := paths[test.path]["get"]
		if !ok || !hasSwaggerParameter(operation.Parameters, test.name, "query", false) {
			t.Errorf("Swagger GET %s must document query parameter %q", test.path, test.name)
		}
	}
}
