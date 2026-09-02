package bootstrap

import (
	"os"
	"strings"
	"testing"
)

func bootstrapCardValue(cards []WorkbenchStatCardDTO, key string) int {
	for _, card := range cards {
		if card.Key == key {
			return card.Value
		}
	}
	return -1
}

func TestBootstrapResponseIsLightweight(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	for _, snippet := range []string{
		"type BootstrapResponse struct",
		"UserDTO",
		"[]AppMenuDTO",
		"`json:\"permissionVersion\"`",
		"ButtonPermissionKeys",
		"`json:\"buttonPermissionKeys\"`",
		"`json:\"buttonPermissionReady\"`",
		"snapshot, err := dingTalkH5PermissionSnapshotForUserDB(ctx, db, user)",
		"Menus:                 dingTalkH5MenusByKeysWithLabelsAndIcons(snapshot.menuKeys, snapshot.labels, snapshot.icons)",
		"ButtonPermissionKeys:  snapshot.buttonKeys",
		"APIPermissionKeys:     snapshot.apiKeys",
		"if user.ID == 0 && user.RoleID == 0",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("bootstrap response must keep lightweight identity/menu snippet %q", snippet)
		}
	}
	for _, snippet := range []string{
		"Users    []UserDTO",
		"Reviews  []ReviewDTO",
		"Template TemplateDTO",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("bootstrap response must not embed bulk payload field %q", snippet)
		}
	}
	bootstrapBody := functionBody(text, "func BootstrapContext")
	for _, snippet := range []string{
		"ListUsersContext(ctx, user)",
		"ListReviewsContext(ctx, user",
		"LoadTemplateContext(ctx)",
		"EnsureSeedContext(ctx)",
	} {
		if strings.Contains(bootstrapBody, snippet) {
			t.Fatalf("BootstrapContext must not load bulk data with %q", snippet)
		}
	}
	for _, snippet := range []string{
		"if user.RoleID == 0 {\n\t\treturn nil",
		"if db == nil || user.RoleID == 0",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("bootstrap menu loading must still honor direct user grants when roleID is 0, found %q", snippet)
		}
	}
}

func TestBootstrapUsesBatchedDingTalkH5PermissionSnapshot(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	bootstrapBody := functionBody(text, "func bootstrapForUserDB")
	for _, snippet := range []string{
		"type dingTalkH5PermissionSnapshot struct",
		"func dingTalkH5PermissionSnapshotForUserDB(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser)",
		"Select(\"`grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_effect`, `grant_edit_time`\")",
		"dingTalkH5PermissionGrantLikeClause()",
		"appapiperm.DingTalkH5APIDeclarations()",
		"appmenuperm.DingTalkH5MenuDeclarations()",
		"appmenuperm.DingTalkH5ButtonDeclarations()",
		"permissionVersionFallback(user)",
		"permission_edit_time",
		"permission_icon",
		"permissionsupport.TablesReady(db)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission snapshot should batch bootstrap permissions with %q", snippet)
		}
	}
	for _, forbidden := range []string{
		"permissionVersionForUserContext(ctx, db, user)",
		"dingTalkH5MenusForUserDB(ctx, db, user)",
		"dingTalkH5ButtonPermissionKeysForUserDB(ctx, db, user)",
		"dingTalkH5APIPermissionKeysForUserDB(ctx, db, user)",
	} {
		if strings.Contains(bootstrapBody, forbidden) {
			t.Fatalf("bootstrap must use one batched permission snapshot, found %q", forbidden)
		}
	}
}

func TestDingTalkH5MenusByKeysBuildsTree(t *testing.T) {
	menus := dingTalkH5MenusByKeysWithLabelsAndIcons([]string{
		"dingtalk_h5:menu:dashboard",
		"dingtalk_h5:menu:performance:mine",
		"dingtalk_h5:menu:performance:hrbp",
	}, nil, nil)
	if len(menus) != 2 {
		t.Fatalf("top-level menus = %d, want 2: %#v", len(menus), menus)
	}
	if menus[0].Key != "dashboard" {
		t.Fatalf("first top-level menu = %q, want dashboard", menus[0].Key)
	}
	performance := menus[1]
	if performance.Key != "performance" {
		t.Fatalf("second top-level menu = %q, want performance", performance.Key)
	}
	if len(performance.Children) != 2 {
		t.Fatalf("performance children = %d, want 2: %#v", len(performance.Children), performance.Children)
	}
	for _, menu := range menus {
		if menu.Key == "performance:mine" || menu.Key == "performance:hrbp" {
			t.Fatalf("child menu %q must not be returned as top-level item", menu.Key)
		}
	}
}

func TestDingTalkH5MenusByKeysUsesPermissionLabels(t *testing.T) {
	menus := dingTalkH5MenusByKeysWithLabelsAndIcons([]string{
		"dingtalk_h5:menu:dashboard",
		"dingtalk_h5:menu:performance:mine",
		"dingtalk_h5:menu:performance:template",
	}, map[string]string{
		"dingtalk_h5:menu:dashboard":            "移动工作台",
		"dingtalk_h5:menu:performance":          "绩效中心",
		"dingtalk_h5:menu:performance:mine":     "我的月度绩效",
		"dingtalk_h5:menu:performance:template": "模板配置",
	}, nil)

	if got := menus[0].Label; got != "移动工作台" {
		t.Fatalf("dashboard label = %q, want %q", got, "移动工作台")
	}
	performance := menus[1]
	if got := performance.Label; got != "绩效中心" {
		t.Fatalf("performance label = %q, want %q", got, "绩效中心")
	}
	if got := performance.Children[0].Label; got != "我的月度绩效" {
		t.Fatalf("first performance child label = %q, want %q", got, "我的月度绩效")
	}
	if got := performance.Children[1].Label; got != "模板配置" {
		t.Fatalf("second performance child label = %q, want %q", got, "模板配置")
	}
}

func TestWorkbenchStatsFromCountsReturnsCardsOnly(t *testing.T) {
	stats := workbenchStatsFromCounts(map[string]int{
		reviewStatusDraft:         1,
		reviewStatusManagerReview: 1,
		reviewStatusCompleted:     1,
	}, 3, 2)

	if len(stats.Cards) != 5 {
		t.Fatalf("workbench cards = %d, want 5", len(stats.Cards))
	}
	if bootstrapCardValue(stats.Cards, "queue") != 2 {
		t.Fatalf("queue card = %d, want 2", bootstrapCardValue(stats.Cards, "queue"))
	}
	if bootstrapCardValue(stats.Cards, "all") != 3 {
		t.Fatalf("all card = %d, want 3", bootstrapCardValue(stats.Cards, "all"))
	}
	if bootstrapCardValue(stats.Cards, "completed") != 1 {
		t.Fatalf("completed card = %d, want 1", bootstrapCardValue(stats.Cards, "completed"))
	}
}

func functionBody(src, signature string) string {
	start := strings.Index(src, signature)
	if start < 0 {
		return ""
	}
	body := src[start:]
	depth := 0
	for index, r := range body {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return body[:index+1]
			}
		}
	}
	return body
}
