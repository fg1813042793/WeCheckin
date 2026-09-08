package appapiperm

import (
	"strings"
	"testing"
)

func TestDingTalkH5APIDeclarationsAreCategorized(t *testing.T) {
	categories := DingTalkH5APICategories()
	if len(categories) == 0 {
		t.Fatalf("dingtalk h5 API categories must not be empty")
	}
	categoryKeys := map[string]bool{}
	for _, category := range categories {
		if !strings.HasPrefix(category.Key, "dingtalk_h5:api-category:") {
			t.Fatalf("unexpected dingtalk h5 API category key %q", category.Key)
		}
		categoryKeys[category.Key] = true
	}

	declarations := DingTalkH5APIDeclarations()
	if len(declarations) == 0 {
		t.Fatalf("dingtalk h5 API declarations must not be empty")
	}
	required := map[string]bool{
		"dingtalk_h5:api:workbench:view":       false,
		"dingtalk_h5:api:review:self_submit":   false,
		"dingtalk_h5:api:review:hrbp_submit":   false,
		"dingtalk_h5:api:review:finalize":      false,
		"dingtalk_h5:api:user:edit":            false,
		"dingtalk_h5:api:template:view":        false,
		"dingtalk_h5:api:template:save":        false,
		"dingtalk_h5:api:workflow:view":        false,
		"dingtalk_h5:api:workflow:start":       false,
		"dingtalk_h5:api:workflow:handle":      false,
		"dingtalk_h5:api:workflow:attachment":  false,
		"dingtalk_h5:api:workflow:withdraw":    false,
		"dingtalk_h5:api:workflow:delete":      false,
		"dingtalk_h5:api:workflow:comment":     false,
		"dingtalk_h5:api:workflow:remind":      false,
		"dingtalk_h5:api:workflow:form-revise": false,
		"dingtalk_h5:api:workflow:summary":     false,
		"dingtalk_h5:api:workflow:export":      false,
	}
	for _, declaration := range declarations {
		if !strings.HasPrefix(declaration.Key, "dingtalk_h5:api:") {
			t.Fatalf("unexpected dingtalk h5 API permission key %q", declaration.Key)
		}
		if !categoryKeys[declaration.CategoryKey] {
			t.Fatalf("declaration %q points to unknown category %q", declaration.Key, declaration.CategoryKey)
		}
		if _, ok := required[declaration.Key]; ok {
			required[declaration.Key] = true
		}
	}
	for key, ok := range required {
		if !ok {
			t.Fatalf("missing dingtalk h5 API permission %s", key)
		}
	}

	workflowRoutes := map[string]string{
		"GET /api/v2/dingtalk/h5/workflows/overview":                  "dingtalk_h5:api:workflow:view",
		"GET /api/v2/dingtalk/h5/workflows/categories":                "dingtalk_h5:api:workflow:view",
		"GET /api/v2/dingtalk/h5/workflows/definitions":               "dingtalk_h5:api:workflow:view",
		"GET /api/v2/dingtalk/h5/workflows/definitions/:id":           "dingtalk_h5:api:workflow:view",
		"POST /api/v2/dingtalk/h5/workflows/attachments":              "dingtalk_h5:api:workflow:attachment",
		"GET /api/v2/dingtalk/h5/workflows/drafts/:definitionId":      "dingtalk_h5:api:workflow:start",
		"PUT /api/v2/dingtalk/h5/workflows/drafts/:definitionId":      "dingtalk_h5:api:workflow:start",
		"DELETE /api/v2/dingtalk/h5/workflows/drafts/:definitionId":   "dingtalk_h5:api:workflow:start",
		"POST /api/v2/dingtalk/h5/workflows/instances":                "dingtalk_h5:api:workflow:start",
		"GET /api/v2/dingtalk/h5/workflows/instances":                 "dingtalk_h5:api:workflow:view",
		"GET /api/v2/dingtalk/h5/workflows/instances/:id":             "dingtalk_h5:api:workflow:view",
		"DELETE /api/v2/dingtalk/h5/workflows/instances/:id":          "dingtalk_h5:api:workflow:delete",
		"POST /api/v2/dingtalk/h5/workflows/instances/:id/withdraw":   "dingtalk_h5:api:workflow:withdraw",
		"POST /api/v2/dingtalk/h5/workflows/instances/:id/comments":   "dingtalk_h5:api:workflow:comment",
		"POST /api/v2/dingtalk/h5/workflows/instances/:id/reminders":  "dingtalk_h5:api:workflow:remind",
		"PATCH /api/v2/dingtalk/h5/workflows/instances/:id/form-data": "dingtalk_h5:api:workflow:form-revise",
		"GET /api/v2/dingtalk/h5/workflows/tasks":                     "dingtalk_h5:api:workflow:view",
		"POST /api/v2/dingtalk/h5/workflows/tasks/:id/complete":       "dingtalk_h5:api:workflow:handle",
		"GET /api/v2/dingtalk/h5/workflows/summary/definitions":       "dingtalk_h5:api:workflow:summary",
		"GET /api/v2/dingtalk/h5/workflows/summary/definitions/:id":   "dingtalk_h5:api:workflow:summary",
		"GET /api/v2/dingtalk/h5/workflows/summary/instances":         "dingtalk_h5:api:workflow:summary",
		"GET /api/v2/dingtalk/h5/workflows/summary/instances/:id":     "dingtalk_h5:api:workflow:summary",
		"GET /api/v2/dingtalk/h5/workflows/summary/export":            "dingtalk_h5:api:workflow:export",
	}
	for _, route := range DingTalkH5RouteDeclarations() {
		key := route.Method + " " + route.Path
		if expected, ok := workflowRoutes[key]; ok {
			if route.PermissionKey != expected {
				t.Fatalf("DingTalk H5 workflow route %s permission = %q, want %q", key, route.PermissionKey, expected)
			}
			delete(workflowRoutes, key)
		}
	}
	if len(workflowRoutes) != 0 {
		t.Fatalf("missing protected DingTalk H5 workflow routes: %+v", workflowRoutes)
	}
}

func TestDingTalkH5UserFeedbackCatalogMatchesPermissionMigration(t *testing.T) {
	wantDeclarations := map[string]struct {
		perms  string
		method string
		path   string
		sort   int
	}{
		"dingtalk_h5:api:feedback:list":       {perms: "feedback:list", method: "GET", path: "/api/v2/dingtalk/h5/user-feedbacks", sort: 10},
		"dingtalk_h5:api:feedback:create":     {perms: "feedback:create", method: "POST", path: "/api/v2/dingtalk/h5/user-feedbacks", sort: 20},
		"dingtalk_h5:api:feedback:detail":     {perms: "feedback:detail", method: "GET", path: "/api/v2/dingtalk/h5/user-feedbacks/:id", sort: 30},
		"dingtalk_h5:api:feedback:supplement": {perms: "feedback:supplement", method: "POST", path: "/api/v2/dingtalk/h5/user-feedbacks/:id/supplements", sort: 40},
	}
	for _, declaration := range DingTalkH5APIDeclarations() {
		if expected, ok := wantDeclarations[declaration.Key]; ok {
			if declaration.CategoryKey != "dingtalk_h5:api-category:feedback" || declaration.Perms != expected.perms || declaration.Method != expected.method || declaration.Path != expected.path || declaration.Sort != expected.sort {
				t.Fatalf("feedback declaration=%#v", declaration)
			}
			delete(wantDeclarations, declaration.Key)
		}
	}
	if len(wantDeclarations) != 0 {
		t.Fatalf("missing feedback API declarations: %#v", wantDeclarations)
	}

	foundCategory := false
	for _, category := range DingTalkH5APICategories() {
		if category.Key == "dingtalk_h5:api-category:feedback" && category.Name == "用户反馈" && category.Platform == "dingtalk_h5" && category.Sort == 80 {
			foundCategory = true
		}
	}
	if !foundCategory {
		t.Fatal("missing DingTalk H5 feedback API category")
	}
}

func TestDingTalkH5CompletedRevisionAPIsAndRoutesAreDeclared(t *testing.T) {
	wantAPIs := map[string]struct {
		perms  string
		method string
		path   string
	}{
		"dingtalk_h5:api:workflow:form-revision-create": {
			perms: "workflow:form-revision-create", method: "POST",
			path: "/api/v2/dingtalk/h5/workflows/instances/:id/form-revisions",
		},
		"dingtalk_h5:api:workflow:form-revision-handle": {
			perms: "workflow:form-revision-handle", method: "POST",
			path: "/api/v2/dingtalk/h5/workflows/form-revision-tasks/:id/complete",
		},
	}
	for _, declaration := range DingTalkH5APIDeclarations() {
		expected, ok := wantAPIs[declaration.Key]
		if !ok {
			continue
		}
		if declaration.CategoryKey != "dingtalk_h5:api-category:workflow" || declaration.Perms != expected.perms ||
			declaration.Method != expected.method || declaration.Path != expected.path {
			t.Fatalf("revision API declaration=%#v", declaration)
		}
		delete(wantAPIs, declaration.Key)
	}
	if len(wantAPIs) != 0 {
		t.Fatalf("missing revision API declarations: %#v", wantAPIs)
	}

	wantRoutes := map[string]string{
		"GET /api/v2/dingtalk/h5/workflows/instances/:id/form-revisions":          "dingtalk_h5:api:workflow:view",
		"POST /api/v2/dingtalk/h5/workflows/instances/:id/form-revisions/preview": "dingtalk_h5:api:workflow:form-revision-create",
		"POST /api/v2/dingtalk/h5/workflows/instances/:id/form-revisions":         "dingtalk_h5:api:workflow:form-revision-create",
		"GET /api/v2/dingtalk/h5/workflows/form-revisions/:id":                    "dingtalk_h5:api:workflow:view",
		"POST /api/v2/dingtalk/h5/workflows/form-revisions/:id/cancel":            "dingtalk_h5:api:workflow:form-revision-create",
		"POST /api/v2/dingtalk/h5/workflows/form-revision-tasks/:id/complete":     "dingtalk_h5:api:workflow:form-revision-handle",
	}
	for _, route := range DingTalkH5RouteDeclarations() {
		key := route.Method + " " + route.Path
		if expected, ok := wantRoutes[key]; ok {
			if route.PermissionKey != expected {
				t.Fatalf("revision route %s permission=%q, want %q", key, route.PermissionKey, expected)
			}
			delete(wantRoutes, key)
		}
	}
	if len(wantRoutes) != 0 {
		t.Fatalf("missing revision routes: %#v", wantRoutes)
	}
}

func TestClientAPIDeclarationsAreCategorized(t *testing.T) {
	categories := ClientAPICategories()
	if len(categories) == 0 {
		t.Fatalf("client API categories must not be empty")
	}
	categoryKeys := map[string]bool{}
	for _, category := range categories {
		if !strings.HasPrefix(category.Key, "client:api-category:") {
			t.Fatalf("unexpected client API category key %q", category.Key)
		}
		categoryKeys[category.Key] = true
	}

	declarations := ClientAPIDeclarations()
	if len(declarations) == 0 {
		t.Fatalf("client API declarations must not be empty")
	}
	required := map[string]bool{
		"client:api:bootstrap:view":    false,
		"client:api:user:view":         false,
		"client:api:upload:create":     false,
		"client:api:news:view":         false,
		"client:api:enroll:submit":     false,
		"client:api:event:score":       false,
		"client:api:survey:response":   false,
		"client:api:exam:answer":       false,
		"client:api:workflow:view":     false,
		"client:api:workflow:start":    false,
		"client:api:workflow:handle":   false,
		"client:api:workflow:withdraw": false,
	}
	for _, declaration := range declarations {
		if !strings.HasPrefix(declaration.Key, "client:api:") {
			t.Fatalf("unexpected client API permission key %q", declaration.Key)
		}
		if !categoryKeys[declaration.CategoryKey] {
			t.Fatalf("declaration %q points to unknown category %q", declaration.Key, declaration.CategoryKey)
		}
		if _, ok := required[declaration.Key]; ok {
			required[declaration.Key] = true
		}
	}
	for key, ok := range required {
		if !ok {
			t.Fatalf("missing client API permission %s", key)
		}
	}

	foundBootstrapRoute := false
	for _, route := range ClientRouteDeclarations() {
		if route.Method == "GET" && route.Path == "/api/v2/me/bootstrap" && route.PermissionKey == "client:api:bootstrap:view" {
			foundBootstrapRoute = true
			break
		}
	}
	if !foundBootstrapRoute {
		t.Fatalf("client bootstrap route must be protected by client:api:bootstrap:view")
	}

	workflowRoutes := map[string]string{
		"GET /api/v2/workflows/definitions":             "client:api:workflow:view",
		"GET /api/v2/workflows/drafts/:definitionId":    "client:api:workflow:start",
		"PUT /api/v2/workflows/drafts/:definitionId":    "client:api:workflow:start",
		"DELETE /api/v2/workflows/drafts/:definitionId": "client:api:workflow:start",
		"POST /api/v2/workflows/instances":              "client:api:workflow:start",
		"GET /api/v2/workflows/instances":               "client:api:workflow:view",
		"POST /api/v2/workflows/instances/:id/withdraw": "client:api:workflow:withdraw",
		"GET /api/v2/workflows/tasks":                   "client:api:workflow:view",
		"POST /api/v2/workflows/tasks/:id/complete":     "client:api:workflow:handle",
	}
	for _, route := range ClientRouteDeclarations() {
		delete(workflowRoutes, route.Method+" "+route.Path)
	}
	if len(workflowRoutes) != 0 {
		t.Fatalf("missing protected workflow routes: %+v", workflowRoutes)
	}
}
