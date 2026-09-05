package admin

import (
	"os"
	"strings"
	"testing"

	"wecheckin/backend/internal/support/adminmenuperm"
	"wecheckin/backend/internal/support/adminrouteperm"
)

func TestAdminUserFeedbackRoutesUseProtectedGroupAndExplicitDI(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		"registerUserFeedbackRoutes(admin)",
		"userfeedbackinfra.NewGormStore(database.GetDB())",
		"userfeedbackinfra.NewObjectStorage()",
		"userfeedbackapp.NewService(store, objectStorage)",
		"userfeedbackhttp.NewHandler(service)",
		`admin.GET("/user-feedbacks/overview", handler.Overview)`,
		`admin.GET("/user-feedbacks", handler.List)`,
		`admin.GET("/user-feedbacks/:id", handler.Detail)`,
		`admin.PATCH("/user-feedbacks/:id/status", handler.UpdateStatus)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("Admin routes missing user feedback registration %q", snippet)
		}
	}
	if strings.Contains(text, "AutoMigrate") {
		t.Fatal("route registration must not run AutoMigrate")
	}
}

func TestAdminUserFeedbackMenuMatchesMigration(t *testing.T) {
	want := map[string]adminmenuperm.Declaration{
		"admin:menu:user-feedback":        {Name: "用户反馈", Path: "/user-feedbacks", Perms: "user-feedback:list", Icon: "ChatDotRound", Sort: 20, Type: adminmenuperm.TypeMenu},
		"admin:menu:user-feedback:list":   {Name: "用户反馈查看", Perms: "user-feedback:list", Sort: 1, Type: adminmenuperm.TypeButton, ParentKey: "admin:menu:user-feedback"},
		"admin:menu:user-feedback:handle": {Name: "用户反馈处理", Perms: "user-feedback:handle", Sort: 2, Type: adminmenuperm.TypeButton, ParentKey: "admin:menu:user-feedback"},
	}
	for _, item := range adminmenuperm.Declarations(true) {
		if expected, ok := want[item.Key]; ok {
			expected.Key = item.Key
			if item != expected {
				t.Fatalf("menu %s=%#v want=%#v", item.Key, item, expected)
			}
			delete(want, item.Key)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing admin feedback menus: %#v", want)
	}
}

func TestAdminUserFeedbackPermissionCatalogAndMiddlewareMatchMigration(t *testing.T) {
	categories := adminrouteperm.Categories()
	foundCategory := false
	for _, category := range categories {
		if category.Key == "admin:api-category:user-feedback" && category.Name == "用户反馈" && category.Sort == 90 {
			foundCategory = true
		}
	}
	if !foundCategory {
		t.Fatal("missing admin user feedback API category")
	}
	wantDeclarations := map[string]struct {
		method string
		path   string
	}{
		"user-feedback:list":   {method: "GET", path: "/api/v2/admin/user-feedbacks"},
		"user-feedback:handle": {method: "PATCH", path: "/api/v2/admin/user-feedbacks/:id/status"},
	}
	for _, declaration := range adminrouteperm.Declarations() {
		if expected, ok := wantDeclarations[declaration.Perms]; ok {
			if declaration.CategoryKey != "admin:api-category:user-feedback" || declaration.Method != expected.method || declaration.Path != expected.path {
				t.Fatalf("declaration=%#v", declaration)
			}
			delete(wantDeclarations, declaration.Perms)
		}
	}
	if len(wantDeclarations) != 0 {
		t.Fatalf("missing admin feedback declarations: %#v", wantDeclarations)
	}

	middleware, err := os.ReadFile("../../../middleware/admin/route_permissions.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, snippet := range []string{
		`"GET /api/v2/admin/user-feedbacks/overview": "user-feedback:list"`,
		`"GET /api/v2/admin/user-feedbacks":          "user-feedback:list"`,
		`{method: "GET", path: "/api/v2/admin/user-feedbacks/:id", perm: "user-feedback:list"}`,
		`{method: "PATCH", path: "/api/v2/admin/user-feedbacks/:id/status", perm: "user-feedback:handle"}`,
	} {
		if !normalizedContains(string(middleware), snippet) {
			t.Fatalf("admin middleware missing %q", snippet)
		}
	}

	migration, err := os.ReadFile("../../../../migrations/20260905101000_add_user_feedback_permissions.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"'admin:api-category:user-feedback'",
		"'admin:api:user-feedback:list'",
		"'/api/v2/admin/user-feedbacks'",
		"'user-feedback:list'",
		"'admin:api:user-feedback:handle'",
		"'/api/v2/admin/user-feedbacks/:id/status'",
		"'user-feedback:handle'",
	} {
		if !strings.Contains(string(migration), fragment) {
			t.Fatalf("feedback migration missing %q", fragment)
		}
	}
}

func normalizedContains(text, snippet string) bool {
	return strings.Contains(strings.Join(strings.Fields(text), " "), strings.Join(strings.Fields(snippet), " "))
}
