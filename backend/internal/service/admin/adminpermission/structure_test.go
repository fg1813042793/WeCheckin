package adminpermission

import (
	"os"
	"strings"
	"testing"
)

func TestAdminPermissionServiceManagesUnifiedPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"type PermissionNode struct",
		"type SaveRequest struct",
		"TreeContext",
		"ListContext",
		"normalizePermissionTypes",
		"`permission_type` IN ?",
		"permissionsupport.EnsurePermissionSchemaContext(ctx, db)",
		"AddContext",
		"EditContext",
		"DeleteContext",
		"model.Permission",
		"permission_icon",
		"permissionsupport.PlatformAdmin",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin permission service must expose unified permission management snippet %s", snippet)
		}
	}
}

func TestAdminPermissionTreeBackfillsApplicationCatalog(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"shouldEnsureApplicationPermissionCatalog(platform, filterTypes)",
		"permissionsupport.EnsureApplicationPermissionCatalogContext(ctx, db, platform, filterTypes)",
		"permissionsupport.PlatformDingTalkH5",
		"permissionsupport.TypeButton",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin permission tree must backfill application catalog before listing with %s", snippet)
		}
	}
}

func TestAdminPermissionTreeUsesCacheBeforeCatalogBackfill(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := testAdminPermissionFunctionBody(t, text, "TreeContext")
	if strings.Contains(body, "if !shouldEnsureApplicationPermissionCatalog") {
		t.Fatalf("permission tree cache must also cover catalog-backed trees")
	}
	cacheIndex := strings.Index(body, "getPermissionTreeCache(platform, filterTypes, now)")
	listIndex := strings.Index(body, "ListContext(ctx, platform, filterTypes...)")
	if cacheIndex < 0 || listIndex < 0 || cacheIndex > listIndex {
		t.Fatalf("permission tree must check cache before entering list/catalog backfill path")
	}
}

func TestAdminPermissionServiceDeletesChildrenAndGrants(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"collectDescendantKeys",
		"`permission_parent_key` IN ?",
		"Delete(&model.PermissionGrant{})",
		"Delete(&model.Permission{})",
		"permissionsupport.InvalidateRuntimePermissionCaches()",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin permission delete must clean child permissions and grants with %s", snippet)
		}
	}
}

func testAdminPermissionFunctionBody(t *testing.T, text, name string) string {
	t.Helper()
	start := strings.Index(text, "func "+name)
	if start < 0 {
		t.Fatalf("function %s not found", name)
	}
	brace := strings.Index(text[start:], "{")
	if brace < 0 {
		t.Fatalf("function %s has no body", name)
	}
	index := start + brace
	depth := 0
	for ; index < len(text); index++ {
		switch text[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start:index]
			}
		}
	}
	t.Fatalf("function %s body not closed", name)
	return ""
}

func TestAdminPermissionServiceCanRenamePermissionKey(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"req.Key = key",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission edit must not overwrite edited key with old key: %s", forbidden)
		}
	}
	for _, snippet := range []string{
		"oldKey := strings.TrimSpace(key)",
		"newKey := item.Key",
		`"permission_key":`,
		"`permission_key` = ? AND `permission_key` <> ?",
		"`grant_permission_key` = ?",
		"Update(\"grant_permission_key\", newKey)",
		"`permission_parent_key` = ?",
		"Update(\"permission_parent_key\", newKey)",
		"permissionsupport.InvalidateRuntimePermissionCaches()",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission rename must keep related records consistent with %s", snippet)
		}
	}
}
