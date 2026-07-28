package bootstrap

import (
	"os"
	"strings"
	"testing"
)

func TestMenuSeedWritesUnifiedPermissionsOnly(t *testing.T) {
	src, err := os.ReadFile("seed_menu.go")
	if err != nil {
		t.Fatalf("read seed_menu.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"permissionsupport.SyncAdminMenuPermissionsContext(context.Background(), db, enableExam)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("menu seed must write unified permissions with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"adminmenuperm.Declarations",
		"model.Menu",
		"model.Permission",
		"`menu_path`",
		"`menu_perms`",
		"`menu_parent_id`",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("menu seed must not use legacy menus table snippet %s", forbidden)
		}
	}
}
