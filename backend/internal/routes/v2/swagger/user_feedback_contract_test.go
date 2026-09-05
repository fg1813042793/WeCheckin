package swagger

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

type userFeedbackSwaggerDocument struct {
	Paths       map[string]map[string]userFeedbackSwaggerOperation `json:"paths"`
	Definitions map[string]userFeedbackSwaggerSchema               `json:"definitions"`
}

type userFeedbackSwaggerOperation struct {
	Consumes    []string                       `json:"consumes"`
	Description string                         `json:"description"`
	Parameters  []userFeedbackSwaggerParameter `json:"parameters"`
}

type userFeedbackSwaggerParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Schema      *struct {
		Ref string `json:"$ref"`
	} `json:"schema"`
}

type userFeedbackSwaggerSchema struct {
	Properties map[string]json.RawMessage `json:"properties"`
}

type userFeedbackSwaggerProperty struct {
	Type string `json:"type"`
}

func TestUserFeedbackSwaggerDocumentsAllRoutes(t *testing.T) {
	doc := loadUserFeedbackSwagger(t)
	want := map[string][]string{
		"/api/v2/dingtalk/h5/user-feedbacks/overview":         {"get"},
		"/api/v2/dingtalk/h5/user-feedbacks":                  {"get", "post"},
		"/api/v2/dingtalk/h5/user-feedbacks/{id}":             {"get"},
		"/api/v2/dingtalk/h5/user-feedbacks/{id}/supplements": {"post"},
		"/api/v2/admin/user-feedbacks/overview":               {"get"},
		"/api/v2/admin/user-feedbacks":                        {"get"},
		"/api/v2/admin/user-feedbacks/{id}":                   {"get"},
		"/api/v2/admin/user-feedbacks/{id}/status":            {"patch"},
	}

	operationCount := 0
	for path, methods := range want {
		for _, method := range methods {
			operationCount++
			if _, ok := swaggerOperation(doc, path, method); !ok {
				t.Errorf("Swagger missing %s %s", strings.ToUpper(method), path)
			}
		}
	}
	if operationCount != 9 {
		t.Fatalf("test setup covers %d operations, want 9", operationCount)
	}
}

func TestUserFeedbackSwaggerDocumentsMultipartContracts(t *testing.T) {
	doc := loadUserFeedbackSwagger(t)

	create := requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks", "post")
	requireMultipartParameters(t, create, map[string]string{
		"content":   "string",
		"images":    "file",
		"requestId": "string",
	})

	supplement := requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks/{id}/supplements", "post")
	requireMultipartParameters(t, supplement, map[string]string{
		"content":   "string",
		"images":    "file",
		"requestId": "string",
		"version":   "integer",
	})
}

func TestUserFeedbackSwaggerDocumentsExactAdminStatusBody(t *testing.T) {
	doc := loadUserFeedbackSwagger(t)
	operation := requireSwaggerOperation(t, doc, "/api/v2/admin/user-feedbacks/{id}/status", "patch")

	var bodyRef string
	for _, parameter := range operation.Parameters {
		if parameter.In == "body" && parameter.Schema != nil {
			bodyRef = parameter.Schema.Ref
			break
		}
	}
	const wantRef = "#/definitions/swagger.UserFeedbackStatusRequest"
	if bodyRef != wantRef {
		t.Fatalf("status body schema = %q, want %q", bodyRef, wantRef)
	}

	schema, ok := doc.Definitions[strings.TrimPrefix(wantRef, "#/definitions/")]
	if !ok {
		t.Fatalf("Swagger missing definition %q", wantRef)
	}
	got := sortedPropertyNames(schema.Properties)
	want := []string{"note", "notifyUser", "requestId", "status", "version"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("status body properties = %v, want exactly %v", got, want)
	}
	wantTypes := map[string]string{
		"note":       "string",
		"notifyUser": "boolean",
		"requestId":  "string",
		"status":     "string",
		"version":    "integer",
	}
	for name, wantType := range wantTypes {
		var property userFeedbackSwaggerProperty
		if err := json.Unmarshal(schema.Properties[name], &property); err != nil {
			t.Fatalf("decode status body property %q: %v", name, err)
		}
		if property.Type != wantType {
			t.Errorf("status body property %q type = %q, want %q", name, property.Type, wantType)
		}
	}
}

func TestUserFeedbackSwaggerDocumentsBehavioralBoundaries(t *testing.T) {
	doc := loadUserFeedbackSwagger(t)

	assertDescriptionContains(t,
		requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks/{id}", "get").Description,
		"当前登录用户", "归属", "不存在",
	)
	assertDescriptionContains(t,
		requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks", "post").Description,
		"requestId", "幂等", "最多 6 张", "单张 10 MB", "JPG、PNG、WebP",
	)
	assertDescriptionContains(t,
		requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks/{id}/supplements", "post").Description,
		"pending", "processing", "累计最多 30 张", "version", "乐观锁",
	)
	assertDescriptionContains(t,
		requireSwaggerOperation(t, doc, "/api/v2/admin/user-feedbacks/{id}/status", "patch").Description,
		"pending -> processing", "pending -> closed", "processing -> resolved", "resolved -> closed", "resolved -> processing", "closed -> processing", "requestId", "幂等", "version", "乐观锁",
	)
}

func TestBackendGuidelinesDocumentImplementedUserFeedbackModule(t *testing.T) {
	content, err := os.ReadFile("../../../../docs/development-guidelines.md")
	if err != nil {
		t.Fatalf("read backend guidelines: %v", err)
	}
	text := string(content)
	for _, fragment := range []string{
		"internal/modules/userfeedback",
		"domain`、`application`、`infrastructure`、`transport/httpadmin`、`transport/httpdingtalkh5",
		"notification_outbox",
		"同一数据库事务",
		"补偿删除",
		"internal/modules/userfeedback/...",
	} {
		if !strings.Contains(text, fragment) {
			t.Errorf("backend guidelines missing implemented user feedback contract %q", fragment)
		}
	}
}

func loadUserFeedbackSwagger(t *testing.T) userFeedbackSwaggerDocument {
	t.Helper()
	content, err := os.ReadFile("../../../../docs/swagger/swagger.json")
	if err != nil {
		t.Fatalf("read generated Swagger JSON: %v", err)
	}
	var doc userFeedbackSwaggerDocument
	if err := json.Unmarshal(content, &doc); err != nil {
		t.Fatalf("decode generated Swagger JSON: %v", err)
	}
	return doc
}

func swaggerOperation(doc userFeedbackSwaggerDocument, path, method string) (userFeedbackSwaggerOperation, bool) {
	methods, ok := doc.Paths[path]
	if !ok {
		return userFeedbackSwaggerOperation{}, false
	}
	operation, ok := methods[method]
	return operation, ok
}

func requireSwaggerOperation(t *testing.T, doc userFeedbackSwaggerDocument, path, method string) userFeedbackSwaggerOperation {
	t.Helper()
	operation, ok := swaggerOperation(doc, path, method)
	if !ok {
		t.Fatalf("Swagger missing %s %s", strings.ToUpper(method), path)
	}
	return operation
}

func requireMultipartParameters(t *testing.T, operation userFeedbackSwaggerOperation, want map[string]string) {
	t.Helper()
	if !containsString(operation.Consumes, "multipart/form-data") {
		t.Fatalf("operation consumes = %v, want multipart/form-data", operation.Consumes)
	}
	got := make(map[string]string)
	for _, parameter := range operation.Parameters {
		if parameter.In == "formData" {
			got[parameter.Name] = parameter.Type
		}
	}
	if strings.Join(sortedMapKeys(got), ",") != strings.Join(sortedMapKeys(want), ",") {
		t.Fatalf("multipart parameters = %v, want exactly %v", sortedMapKeys(got), sortedMapKeys(want))
	}
	for name, wantType := range want {
		if got[name] != wantType {
			t.Errorf("multipart parameter %q type = %q, want %q", name, got[name], wantType)
		}
	}
}

func sortedMapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertDescriptionContains(t *testing.T, description string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(description, fragment) {
			t.Errorf("description %q missing %q", description, fragment)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sortedPropertyNames(properties map[string]json.RawMessage) []string {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
