package passport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientBootstrapUsesUnifiedPermissionSnapshot(t *testing.T) {
	src, err := os.ReadFile("bootstrap.go")
	if err != nil {
		t.Fatalf("read bootstrap.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"type ClientMenuDTO struct",
		"func BootstrapContext(ctx context.Context, userID string) (*BootstrapResponse, error)",
		"func BootstrapByIDContext(ctx context.Context, id uint) (*BootstrapResponse, error)",
		"fillUserRoleIDsContext(ctx, db, user)",
		"permissionsupport.SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, user.ID, user.RoleIDs, permissionsupport.PlatformClient)",
		"permissionsupport.SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, user.ID, user.RoleIDs, permissionsupport.PlatformClient)",
		"clientMenusByKeysWithLabels(menuKeys",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("client bootstrap must build menu/API snapshot from unified permissions with %q", snippet)
		}
	}
}

func TestFrontendClientMenusUseBootstrapPermissions(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..")
	required := map[string][]string{
		filepath.Join(root, "frontend", "api", "index.js"): {
			"bootstrap()",
			"return get(`${API_V2}/me/bootstrap`)",
		},
		filepath.Join(root, "frontend", "utils", "clientPermission.js"): {
			"CLIENT_MENU_KEY",
			"setClientPermissionSnapshot",
			"ensureClientPermissionSnapshot",
			"hasClientMenuPermission",
			"filterClientMenus",
		},
		filepath.Join(root, "frontend", "pages", "login", "login.vue"): {
			"ensureClientPermissionSnapshot",
			"await ensureClientPermissionSnapshot()",
		},
		filepath.Join(root, "frontend", "pages", "login", "login_pwd.vue"): {
			"ensureClientPermissionSnapshot",
			"await ensureClientPermissionSnapshot()",
		},
		filepath.Join(root, "frontend", "pages", "index", "index.vue"): {
			"visibleHomeMenus",
			"ensureClientPermissionSnapshot",
			"@click=\"openMenu(menu)\"",
		},
		filepath.Join(root, "frontend", "pages", "my", "my_index.vue"): {
			"visiblePrimaryMenus",
			"loadClientPermissions",
			"hasClientMenuPermission",
		},
	}
	for path, snippets := range required {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(src)
		for _, snippet := range snippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must include client menu permission snippet %q", path, snippet)
			}
		}
	}
}
