package dingtalkh5

import (
	"os"
	"strings"
	"testing"

	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
)

func TestDingTalkH5UserFeedbackRoutesUseProtectedGroupAndExplicitDI(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		"userfeedbackinfra.NewGormStore(db)",
		"userfeedbackinfra.NewObjectStorage()",
		"userfeedbackapp.NewService(feedbackStore, feedbackStorage)",
		"userfeedbackhttp.NewHandler(feedbackService)",
		`auth.GET("/user-feedbacks/overview", feedbackHandler.Overview)`,
		`auth.GET("/user-feedbacks", feedbackHandler.List)`,
		`auth.POST("/user-feedbacks", feedbackHandler.Create)`,
		`auth.GET("/user-feedbacks/:id", feedbackHandler.Detail)`,
		`auth.POST("/user-feedbacks/:id/supplements", feedbackHandler.Supplement)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk H5 routes missing user feedback registration %q", snippet)
		}
	}
	if strings.Contains(text, "AutoMigrate") {
		t.Fatal("route registration must not run AutoMigrate")
	}
}

func TestDingTalkH5UserFeedbackMenuMatchesMigration(t *testing.T) {
	var found *appmenuperm.Declaration
	for i := range appmenuperm.DingTalkH5MenuDeclarations() {
		item := appmenuperm.DingTalkH5MenuDeclarations()[i]
		if item.Key == "dingtalk_h5:menu:feedback" {
			found = &item
			break
		}
	}
	if found == nil || found.Name != "我的反馈" || found.Path != "feedback" || found.Icon != "chat" || found.Sort != 110 {
		t.Fatalf("feedback menu = %#v", found)
	}
	migration := readFeedbackPermissionMigration(t, "20260905101000_add_user_feedback_permissions.sql")
	for _, fragment := range []string{"'dingtalk_h5:menu:feedback'", "'feedback'", "'mine'", "110"} {
		if !strings.Contains(migration, fragment) {
			t.Fatalf("initial feedback migration missing %q", fragment)
		}
	}
	updateMigration := readFeedbackPermissionMigration(t, "20260905102000_update_user_feedback_h5_menu.sql")
	for _, fragment := range []string{"'dingtalk_h5:menu:feedback'", "'我的反馈'", "'chat'"} {
		if !strings.Contains(updateMigration, fragment) {
			t.Fatalf("feedback menu update migration missing %q", fragment)
		}
	}
}

func TestDingTalkH5UserFeedbackRoutesReuseMigratedPermissions(t *testing.T) {
	want := map[string]string{
		"GET /api/v2/dingtalk/h5/user-feedbacks/overview":         "dingtalk_h5:api:feedback:list",
		"GET /api/v2/dingtalk/h5/user-feedbacks":                  "dingtalk_h5:api:feedback:list",
		"POST /api/v2/dingtalk/h5/user-feedbacks":                 "dingtalk_h5:api:feedback:create",
		"GET /api/v2/dingtalk/h5/user-feedbacks/:id":              "dingtalk_h5:api:feedback:detail",
		"POST /api/v2/dingtalk/h5/user-feedbacks/:id/supplements": "dingtalk_h5:api:feedback:supplement",
	}
	for _, declaration := range appapiperm.DingTalkH5RouteDeclarations() {
		key := declaration.Method + " " + declaration.Path
		if expected, ok := want[key]; ok {
			if declaration.PermissionKey != expected {
				t.Fatalf("route %s permission=%q want=%q", key, declaration.PermissionKey, expected)
			}
			delete(want, key)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing user feedback route declarations: %#v", want)
	}
	for _, declaration := range appapiperm.DingTalkH5APIDeclarations() {
		if strings.Contains(declaration.Key, "overview") {
			t.Fatalf("overview must reuse list permission, got separate declaration %#v", declaration)
		}
	}
}

func readFeedbackPermissionMigration(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("../../../../migrations/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
