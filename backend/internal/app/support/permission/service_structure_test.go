package permission

import (
	"os"
	"strings"
	"testing"
)

func TestPermissionServiceExposesUnifiedAccessFunctions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"AdminLoginPermissionKey",
		"SubjectRole",
		"SubjectUser",
		"EffectAllow",
		"EffectDeny",
		"EnsureUnifiedPermissionsContext",
		"SubjectHasPermissionContext",
		"AdminMenuPermissionsContext",
		"AdminPermCodesContext",
		"DataScopeContext",
		"EnsurePermissionSchemaContext",
		"SyncAdminMenuPermissionsContext",
		"SetRoleAdminPermissionKeysTx",
		"SetUserAdminPermissionOverridesTx",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must expose %s", snippet)
		}
	}
}

func TestPermissionServiceBackfillsOptionalPermissionColumns(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func EnsurePermissionSchemaContext(ctx context.Context, db *gorm.DB) error",
		"HasColumn(&model.Permission{}, \"Icon\")",
		"AddColumn(&model.Permission{}, \"Icon\")",
		"EnsurePermissionSchemaContext(context.Background(), db)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must backfill optional permission columns with %s", snippet)
		}
	}
}

func TestPermissionServiceSkipsStaleLegacyAdminMenuGrants(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`strings.HasPrefix(key, "admin:menu:")`,
		"continue",
		"gorm.ErrRecordNotFound",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must skip stale legacy admin menu grants with %s", snippet)
		}
	}
}

func TestPermissionServiceDoesNotReadLegacyRoleGrantTables(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"syncLegacyRoleGrants",
		"legacyAdminAPIPermissionKeysByMenuKeys",
		"role_menus",
		"role_depts",
		"MenuPermissionKey(",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission service must not read old role authorization table snippet %s", forbidden)
		}
	}
	if !strings.Contains(text, "setRoleAdminPermissionKeysTx(tx, roleID, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs, true)") {
		t.Fatalf("permission service must keep unified role grant writer")
	}
}

func TestPermissionServiceProvidesRuntimeRoleGrantLookupsWithoutLegacyTables(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"RoleAdminPermissionKeyMapContext",
		"RoleAdminAPIPermissionKeyMapContext",
		"RoleCustomDeptIDMapContext",
		"RoleCustomDeptIDsContext",
		"RoleIDsByCustomDeptIDsContext",
		"decodeDeptScope",
		"model.PermissionGrant",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must provide unified runtime role lookup %s", snippet)
		}
	}
	if strings.Contains(text, "func legacyRoleAllowsAdminAccess") {
		t.Fatalf("runtime admin access must not fallback to role_menus after old role grant tables are cleaned")
	}
}

func TestPermissionServiceSyncsAdminAPIPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"TypeAPI",
		"TypeAPICategory",
		"AdminAPIPermissionKey",
		"syncAdminAPIPermissions",
		"adminrouteperm.Categories",
		"adminrouteperm.Declarations",
		"ParentKey: declaration.CategoryKey",
		"Type:     TypeAPICategory",
		"PlatformAdmin",
		"Type:     TypeAPI",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must sync admin API permission snippet %s", snippet)
		}
	}
}

func TestPermissionServiceSyncsAdminMenusFromDeclarationsOnly(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"adminmenuperm.Declarations",
		"adminmenuperm.Declarations(enableExam)",
		"upsertAdminMenuPermission",
		"permission_icon",
		"TypeDirectory",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must sync admin menus from declarations with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"syncLegacyMenuTablePermissions",
		"legacyAdminMenuRow",
		"legacyMenuType",
		"var menus []model.Menu",
		"db.Find(&menus)",
		"MenuPermissionKey(menu.ID)",
		`HasTable("menus")`,
		`db.Table("menus")`,
		"adminmenuperm.Declarations(true)",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission service must not read legacy menus table snippet %s", forbidden)
		}
	}
}

func TestPermissionServiceSupportsUserLevelAPIAllowDeny(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`"admin:api:%"`,
		"subjectAdminPermissionSets",
		"AdminPermissionPrefixes",
		"EffectDeny",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must support user-level API allow/deny snippet %s", snippet)
		}
	}
}

func TestPermissionServiceSyncsClientAndDingTalkH5Menus(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"PlatformClient",
		"PlatformDingTalkH5",
		"syncClientMenuPermissions",
		"syncDingTalkH5MenuPermissions",
		"appmenuperm.ClientMenuDeclarations",
		"appmenuperm.DingTalkH5MenuDeclarations",
		"ClientMenuPermissionKeysContext",
		"DingTalkH5MenuPermissionKeysContext",
		"ApplicationMenuPermissionPrefixes",
		"UserPermissionPrefixes",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must sync app menu permission snippet %s", snippet)
		}
	}
}
