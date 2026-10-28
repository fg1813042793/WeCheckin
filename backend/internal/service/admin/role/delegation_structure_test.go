package role

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRolePermissionTreesAndDeletesCarryCurrentAdmin(t *testing.T) {
	handler, err := os.ReadFile(filepath.Join("..", "..", "..", "handler", "admin", "role", "handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(handler)
	for _, snippet := range []string{
		"ApplicationPermissionTreeForAdminContext(ctx, admin.ID)",
		"DeleteForAdminContext(ctx, admin.ID",
		"BatchDeleteForAdminContext(ctx, admin.ID",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role handler delegation guard missing %q", snippet)
		}
	}
}

