package performance

import (
	"os"
	"strings"
	"testing"
)

func TestBootstrapResponseIsLightweight(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	typesText := string(typesSrc)
	reviewsText := string(reviewsSrc)
	bootstrapBody := reviewsText
	if start := strings.Index(reviewsText, "func BootstrapContext"); start >= 0 {
		bootstrapBody = reviewsText[start:]
		if end := strings.Index(bootstrapBody, "\n}\n\nfunc "); end >= 0 {
			bootstrapBody = bootstrapBody[:end+3]
		}
	}

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
		if !strings.Contains(typesText+reviewsText, snippet) {
			t.Fatalf("bootstrap response must keep lightweight identity/menu snippet %q", snippet)
		}
	}
	for _, snippet := range []string{
		"Users    []UserDTO",
		"Reviews  []ReviewDTO",
		"Template TemplateDTO",
	} {
		if strings.Contains(typesText, snippet) {
			t.Fatalf("bootstrap response must not embed bulk payload field %q", snippet)
		}
	}
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
		if strings.Contains(reviewsText, snippet) {
			t.Fatalf("bootstrap menu loading must still honor direct user grants when roleID is 0, found %q", snippet)
		}
	}
}

func TestBootstrapUsesBatchedDingTalkH5PermissionSnapshot(t *testing.T) {
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	snapshotSrc, err := os.ReadFile("permission_snapshot.go")
	if err != nil {
		t.Fatalf("read permission_snapshot.go: %v", err)
	}
	reviewsText := string(reviewsSrc)
	snapshotText := string(snapshotSrc)
	bootstrapBody := reviewsText
	if start := strings.Index(reviewsText, "func bootstrapForUserDB"); start >= 0 {
		bootstrapBody = reviewsText[start:]
		if end := strings.Index(bootstrapBody, "\n}\n\nfunc "); end >= 0 {
			bootstrapBody = bootstrapBody[:end+3]
		}
	}
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
		if !strings.Contains(snapshotText, snippet) {
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

func TestDingTalkH5RefreshReloadsBootstrapMenuLabels(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk h5 index page: %v", err)
	}
	page := string(pageSrc)
	for _, snippet := range []string{
		"async function refreshSessionAndDataSafely()",
		"const bootstrapped = await loadBootstrapSafely()",
		"if (!bootstrapped) return false",
		"return refreshDataSafely({ forceReference: true, contentLoading: true })",
		"async function refreshWithUserFeedback()",
		"const refreshed = await refreshSessionAndDataSafely()",
		"refreshData: refreshWithUserFeedback",
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("dingtalk h5 refresh must reload lightweight bootstrap menu labels with %q", snippet)
		}
	}
	refreshBody := page
	if start := strings.Index(refreshBody, "async function refreshData("); start >= 0 {
		refreshBody = refreshBody[start:]
		if end := strings.Index(refreshBody, "\n}\n\nfunction "); end >= 0 {
			refreshBody = refreshBody[:end+3]
		}
	}
	if strings.Contains(refreshBody, "loadBootstrapSafely()") {
		t.Fatalf("refreshData must stay scoped to current view; bootstrap reload belongs to refreshSessionAndDataSafely")
	}
}

func TestDingTalkH5BootstrapAndLoginExposeConfiguredBranding(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	authSrc, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
	}
	combined := string(typesSrc) + string(reviewsSrc) + string(authSrc)
	for _, snippet := range []string{
		"type DingTalkH5AppConfigDTO struct",
		"`json:\"appTitle\"`",
		"`json:\"appName\"`",
		"`json:\"logoText\"`",
		"`json:\"logoUrl\"`",
		"dingTalkH5AppConfigContext(ctx)",
		"AppConfig:             bootstrap.AppConfig",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("bootstrap/login should expose configured branding with %q", snippet)
		}
	}
}

func TestDingTalkH5FrontendRendersConfiguredBranding(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk h5 index page: %v", err)
	}
	appShellSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/AppShell.vue")
	if err != nil {
		t.Fatalf("read app shell: %v", err)
	}
	loginSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/LoginView.vue")
	if err != nil {
		t.Fatalf("read login view: %v", err)
	}
	combined := string(pageSrc) + string(appShellSrc) + string(loginSrc)
	for _, snippet := range []string{
		":app-config=\"appConfig\"",
		":app-config=\"appConfig\"",
		"state.appConfig",
		"payload.appConfig",
		"appLogoText",
		"appLogoUrl",
		`v-if="appLogoUrl"`,
		`<image class="brand-logo-img"`,
		"{{ appLogoText }}",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("dingtalk h5 frontend should render configured branding with %q", snippet)
		}
	}
}

func TestDingTalkH5FrontendSyncsConfiguredAppTitleToDingTalkTab(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk h5 index page: %v", err)
	}
	htmlSrc, err := os.ReadFile("../../../../../dingtalk-h5/index.html")
	if err != nil {
		t.Fatalf("read dingtalk h5 index html: %v", err)
	}
	page := string(pageSrc)
	for _, snippet := range []string{
		`import { isDingTalkRuntime, requestAuthCode, setNavigationTitle, waitForDingTalkJSAPI } from '../../utils/dingtalk'`,
		`function syncDingTalkPageTitle(title) {`,
		`document.title = text`,
		`setNavigationTitle(text)`,
		`watch(appTitle, (title) => {`,
		`syncDingTalkPageTitle(title)`,
		`{ immediate: true }`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("dingtalk h5 page should sync configured app title with %q", snippet)
		}
	}
	if strings.Contains(string(htmlSrc), "<title>OA管理</title>") {
		t.Fatalf("dingtalk h5 static html title should not keep the old OA管理 fallback")
	}
	for _, forbidden := range []string{
		"state.appTitle = 'OA管理'",
		"appTitle: 'OA管理'",
		"appName: 'OA管理'",
		"'OA管理'))",
	} {
		if strings.Contains(page, forbidden) {
			t.Fatalf("dingtalk h5 page should not eagerly write old OA管理 title fallback: %q", forbidden)
		}
	}
}
