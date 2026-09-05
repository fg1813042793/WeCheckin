package swagger

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type userFeedbackSwaggerDocument struct {
	Paths       map[string]map[string]userFeedbackSwaggerOperation `json:"paths" yaml:"paths"`
	Definitions map[string]userFeedbackSwaggerSchema               `json:"definitions" yaml:"definitions"`
}

type userFeedbackSwaggerOperation struct {
	Consumes    []string                               `json:"consumes" yaml:"consumes"`
	Description string                                 `json:"description" yaml:"description"`
	Parameters  []userFeedbackSwaggerParameter         `json:"parameters" yaml:"parameters"`
	Responses   map[string]userFeedbackSwaggerResponse `json:"responses" yaml:"responses"`
	Security    []map[string][]string                  `json:"security" yaml:"security"`
}

type userFeedbackSwaggerParameter struct {
	Name        string                     `json:"name" yaml:"name"`
	In          string                     `json:"in" yaml:"in"`
	Type        string                     `json:"type" yaml:"type"`
	Format      string                     `json:"format" yaml:"format"`
	Required    bool                       `json:"required" yaml:"required"`
	Description string                     `json:"description" yaml:"description"`
	Schema      *userFeedbackSwaggerSchema `json:"schema" yaml:"schema"`
}

type userFeedbackSwaggerSchema struct {
	Ref        string                               `json:"$ref" yaml:"$ref"`
	Type       string                               `json:"type" yaml:"type"`
	Format     string                               `json:"format" yaml:"format"`
	Enum       []string                             `json:"enum" yaml:"enum"`
	Default    *bool                                `json:"default" yaml:"default"`
	Properties map[string]userFeedbackSwaggerSchema `json:"properties" yaml:"properties"`
	Required   []string                             `json:"required" yaml:"required"`
	AllOf      []userFeedbackSwaggerSchema          `json:"allOf" yaml:"allOf"`
}

type userFeedbackSwaggerResponse struct {
	Schema userFeedbackSwaggerSchema `json:"schema" yaml:"schema"`
}

type userFeedbackSwaggerParameterExpectation struct {
	In       string
	Type     string
	Required bool
}

func TestUserFeedbackSwaggerDocumentsAllRoutes(t *testing.T) {
	doc := loadUserFeedbackSwagger(t)
	want := map[string]bool{
		"get /api/v2/dingtalk/h5/user-feedbacks/overview":          true,
		"get /api/v2/dingtalk/h5/user-feedbacks":                   true,
		"post /api/v2/dingtalk/h5/user-feedbacks":                  true,
		"get /api/v2/dingtalk/h5/user-feedbacks/{id}":              true,
		"post /api/v2/dingtalk/h5/user-feedbacks/{id}/supplements": true,
		"get /api/v2/admin/user-feedbacks/overview":                true,
		"get /api/v2/admin/user-feedbacks":                         true,
		"get /api/v2/admin/user-feedbacks/{id}":                    true,
		"patch /api/v2/admin/user-feedbacks/{id}/status":           true,
	}

	got := make(map[string]bool)
	for path, methods := range doc.Paths {
		if !isUserFeedbackSwaggerPath(path) {
			continue
		}
		for method := range methods {
			if isSwaggerHTTPMethod(method) {
				got[method+" "+path] = true
			}
		}
	}
	if len(got) != 9 {
		t.Errorf("Swagger user feedback operation count = %d, want 9: %v", len(got), sortedTrueKeys(got))
	}
	for operation := range want {
		if !got[operation] {
			t.Errorf("Swagger missing %s", strings.ToUpper(operation))
		}
	}
	for operation := range got {
		if !want[operation] {
			t.Errorf("Swagger has unexpected user feedback operation %s", strings.ToUpper(operation))
		}
	}
}

func TestUserFeedbackSwaggerDocumentsMultipartContracts(t *testing.T) {
	for document, doc := range userFeedbackSwaggerDocuments(t) {
		t.Run(document, func(t *testing.T) {
			create := requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks", "post")
			requireExactSwaggerParameters(t, create, map[string]userFeedbackSwaggerParameterExpectation{
				"formData:content":   {In: "formData", Type: "string", Required: true},
				"formData:images":    {In: "formData", Type: "file", Required: false},
				"formData:requestId": {In: "formData", Type: "string", Required: true},
			})

			supplement := requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks/{id}/supplements", "post")
			requireExactSwaggerParameters(t, supplement, map[string]userFeedbackSwaggerParameterExpectation{
				"path:id":            {In: "path", Type: "integer", Required: true},
				"formData:content":   {In: "formData", Type: "string", Required: false},
				"formData:images":    {In: "formData", Type: "file", Required: false},
				"formData:requestId": {In: "formData", Type: "string", Required: true},
				"formData:version":   {In: "formData", Type: "integer", Required: true},
			})
		})
	}
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
		property := schema.Properties[name]
		if property.Type != wantType {
			t.Errorf("status body property %q type = %q, want %q", name, property.Type, wantType)
		}
	}
	assertExactStrings(t, "status enum", schema.Properties["status"].Enum, []string{"pending", "processing", "resolved", "closed"})
	notifyUserDefault := schema.Properties["notifyUser"].Default
	if notifyUserDefault == nil || !*notifyUserDefault {
		t.Errorf("notifyUser default = %v, want true", notifyUserDefault)
	}
	wantRequired := []string{"requestId", "status", "version"}
	assertExactRequiredFields(t, "swagger.json", schema.Required, wantRequired)

	yamlDoc := loadUserFeedbackSwaggerYAML(t)
	yamlSchema, ok := yamlDoc.Definitions[strings.TrimPrefix(wantRef, "#/definitions/")]
	if !ok {
		t.Fatalf("Swagger YAML missing definition %q", wantRef)
	}
	assertExactRequiredFields(t, "swagger.yaml", yamlSchema.Required, wantRequired)
	assertExactStrings(t, "swagger.yaml status enum", yamlSchema.Properties["status"].Enum, []string{"pending", "processing", "resolved", "closed"})
	yamlNotifyUserDefault := yamlSchema.Properties["notifyUser"].Default
	if yamlNotifyUserDefault == nil || !*yamlNotifyUserDefault {
		t.Errorf("swagger.yaml notifyUser default = %v, want true", yamlNotifyUserDefault)
	}

	assertExactStrings(t, "PATCH consumes", operation.Consumes, []string{"application/json"})
	yamlOperation := requireSwaggerOperation(t, yamlDoc, "/api/v2/admin/user-feedbacks/{id}/status", "patch")
	assertExactStrings(t, "swagger.yaml PATCH consumes", yamlOperation.Consumes, []string{"application/json"})
}

func TestUserFeedbackSwaggerDocumentsAdminTimeFormats(t *testing.T) {
	for document, doc := range userFeedbackSwaggerDocuments(t) {
		t.Run(document, func(t *testing.T) {
			for _, path := range []string{
				"/api/v2/admin/user-feedbacks/overview",
				"/api/v2/admin/user-feedbacks",
			} {
				operation := requireSwaggerOperation(t, doc, path, "get")
				for _, name := range []string{"submittedFrom", "submittedTo"} {
					parameter := requireSwaggerParameter(t, operation, "query", name)
					if parameter.Type != "integer" || parameter.Format != "int64" {
						t.Errorf("%s GET %s query %s = type %q format %q, want integer/int64", document, path, name, parameter.Type, parameter.Format)
					}
				}
			}
		})
	}
}

func TestUserFeedbackSwaggerDocumentsSecurityAndResponseModels(t *testing.T) {
	tests := []struct {
		path       string
		method     string
		security   string
		dataSchema string
	}{
		{"/api/v2/dingtalk/h5/user-feedbacks/overview", "get", "H5AppToken", "#/definitions/application.Overview"},
		{"/api/v2/dingtalk/h5/user-feedbacks", "get", "H5AppToken", "#/definitions/application.FeedbackList"},
		{"/api/v2/dingtalk/h5/user-feedbacks", "post", "H5AppToken", "#/definitions/application.FeedbackDetail"},
		{"/api/v2/dingtalk/h5/user-feedbacks/{id}", "get", "H5AppToken", "#/definitions/application.FeedbackDetail"},
		{"/api/v2/dingtalk/h5/user-feedbacks/{id}/supplements", "post", "H5AppToken", "#/definitions/application.FeedbackDetail"},
		{"/api/v2/admin/user-feedbacks/overview", "get", "AdminToken", "#/definitions/application.Overview"},
		{"/api/v2/admin/user-feedbacks", "get", "AdminToken", "#/definitions/application.FeedbackList"},
		{"/api/v2/admin/user-feedbacks/{id}", "get", "AdminToken", "#/definitions/application.FeedbackDetail"},
		{"/api/v2/admin/user-feedbacks/{id}/status", "patch", "AdminToken", "#/definitions/application.FeedbackDetail"},
	}

	for document, doc := range userFeedbackSwaggerDocuments(t) {
		for _, test := range tests {
			name := document + "_" + test.method + "_" + test.path
			t.Run(name, func(t *testing.T) {
				operation := requireSwaggerOperation(t, doc, test.path, test.method)
				assertSwaggerSecurity(t, operation, test.security)
				assertSwaggerSuccessDataSchema(t, operation, test.dataSchema)
			})
		}
	}
}

func TestUserFeedbackSwaggerDocumentsSafeResponseIDsAndFirstImage(t *testing.T) {
	for document, doc := range userFeedbackSwaggerDocuments(t) {
		t.Run(document, func(t *testing.T) {
			for _, definition := range []string{
				"application.FeedbackSummary",
				"application.FeedbackDetail",
				"application.Message",
				"application.Attachment",
			} {
				schema, ok := doc.Definitions[definition]
				if !ok {
					t.Fatalf("%s missing definition %q", document, definition)
				}
				if schema.Properties["id"].Type != "string" {
					t.Errorf("%s %s.id type = %q, want string", document, definition, schema.Properties["id"].Type)
				}
			}
			for _, definition := range []string{"application.FeedbackSummary", "application.FeedbackDetail"} {
				property := doc.Definitions[definition].Properties["firstImageUrl"]
				if property.Type != "string" {
					t.Errorf("%s %s.firstImageUrl type = %q, want string", document, definition, property.Type)
				}
			}
		})
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
		"requestId", "幂等", "最多 6 张", "单张 10 MB", "JPG、PNG、WebP", "multipart 请求体总大小最多 64 MiB",
	)
	assertDescriptionContains(t,
		requireSwaggerOperation(t, doc, "/api/v2/dingtalk/h5/user-feedbacks/{id}/supplements", "post").Description,
		"pending", "processing", "累计最多 30 张", "version", "乐观锁", "multipart 请求体总大小最多 64 MiB",
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

func loadUserFeedbackSwaggerYAML(t *testing.T) userFeedbackSwaggerDocument {
	t.Helper()
	content, err := os.ReadFile("../../../../docs/swagger/swagger.yaml")
	if err != nil {
		t.Fatalf("read generated Swagger YAML: %v", err)
	}
	var doc userFeedbackSwaggerDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		t.Fatalf("decode generated Swagger YAML: %v", err)
	}
	return doc
}

func userFeedbackSwaggerDocuments(t *testing.T) map[string]userFeedbackSwaggerDocument {
	t.Helper()
	return map[string]userFeedbackSwaggerDocument{
		"swagger.json": loadUserFeedbackSwagger(t),
		"swagger.yaml": loadUserFeedbackSwaggerYAML(t),
	}
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

func requireExactSwaggerParameters(t *testing.T, operation userFeedbackSwaggerOperation, want map[string]userFeedbackSwaggerParameterExpectation) {
	t.Helper()
	if !containsString(operation.Consumes, "multipart/form-data") {
		t.Fatalf("operation consumes = %v, want multipart/form-data", operation.Consumes)
	}
	got := make(map[string]userFeedbackSwaggerParameter)
	for _, parameter := range operation.Parameters {
		got[parameter.In+":"+parameter.Name] = parameter
	}
	if strings.Join(sortedMapKeys(got), ",") != strings.Join(sortedMapKeys(want), ",") {
		t.Fatalf("multipart parameters = %v, want exactly %v", sortedMapKeys(got), sortedMapKeys(want))
	}
	for key, expected := range want {
		parameter := got[key]
		if parameter.In != expected.In || parameter.Type != expected.Type || parameter.Required != expected.Required {
			t.Errorf("multipart parameter %q = in %q type %q required %t, want in %q type %q required %t", key, parameter.In, parameter.Type, parameter.Required, expected.In, expected.Type, expected.Required)
		}
	}
}

func requireSwaggerParameter(t *testing.T, operation userFeedbackSwaggerOperation, in, name string) userFeedbackSwaggerParameter {
	t.Helper()
	for _, parameter := range operation.Parameters {
		if parameter.In == in && parameter.Name == name {
			return parameter
		}
	}
	t.Fatalf("Swagger operation missing %s parameter %q", in, name)
	return userFeedbackSwaggerParameter{}
}

func assertSwaggerSecurity(t *testing.T, operation userFeedbackSwaggerOperation, want string) {
	t.Helper()
	if len(operation.Security) != 1 || len(operation.Security[0]) != 1 {
		t.Fatalf("security = %v, want exactly %s", operation.Security, want)
	}
	scopes, ok := operation.Security[0][want]
	if !ok || len(scopes) != 0 {
		t.Errorf("security = %v, want exactly %s with no scopes", operation.Security, want)
	}
}

func assertSwaggerSuccessDataSchema(t *testing.T, operation userFeedbackSwaggerOperation, want string) {
	t.Helper()
	response, ok := operation.Responses["200"]
	if !ok {
		t.Fatal("Swagger operation missing 200 response")
	}
	baseFound := false
	dataRef := ""
	for _, schema := range response.Schema.AllOf {
		if schema.Ref == "#/definitions/response.Resp" {
			baseFound = true
		}
		if data, ok := schema.Properties["data"]; ok {
			dataRef = data.Ref
		}
	}
	if !baseFound {
		t.Error("200 response does not include response.Resp")
	}
	if dataRef != want {
		t.Errorf("200 response data schema = %q, want %q", dataRef, want)
	}
}

func isUserFeedbackSwaggerPath(path string) bool {
	return path == "/api/v2/admin/user-feedbacks" ||
		strings.HasPrefix(path, "/api/v2/admin/user-feedbacks/") ||
		path == "/api/v2/dingtalk/h5/user-feedbacks" ||
		strings.HasPrefix(path, "/api/v2/dingtalk/h5/user-feedbacks/")
}

func isSwaggerHTTPMethod(method string) bool {
	switch method {
	case "get", "put", "post", "delete", "patch", "head", "options":
		return true
	default:
		return false
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

func sortedPropertyNames(properties map[string]userFeedbackSwaggerSchema) []string {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func assertExactRequiredFields(t *testing.T, document string, got, want []string) {
	t.Helper()
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s status body required fields = %v, want exactly %v", document, got, want)
	}
}

func assertExactStrings(t *testing.T, name string, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s = %v, want exactly %v", name, got, want)
	}
}

func sortedTrueKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if value {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
