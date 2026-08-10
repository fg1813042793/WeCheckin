package permission

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
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
		"AdminPermissionKeysContext",
		"DataScopeContext",
		"DataScopeWithRoleIDsContext",
		"DataScopeExtrasContext",
		"DataScopeExtrasWithRoleIDsContext",
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

func TestPermissionServiceCachesRuntimeTableReadiness(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"permissionTablesReadyCache",
		"permissionTablesReadyNegativeCacheTTL",
		"ResetPermissionTablesReadyCache",
		"markPermissionTablesReady",
		"permissionTablesReadyCache.RLock()",
		"permissionTablesReadyCache.Lock()",
		"time.Since(permissionTablesReadyCache.checkedAt)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must cache runtime table readiness with %s", snippet)
		}
	}
	if strings.Contains(text, "return db != nil && db.Migrator().HasTable(&model.Permission{}) && db.Migrator().HasTable(&model.PermissionGrant{})") {
		t.Fatalf("TablesReady must not query database schema on every runtime permission check")
	}
}

func TestEnsurePermissionSchemaContextUsesReadyCacheBeforeMigrator(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := testFunctionBody(t, text, "EnsurePermissionSchemaContext")
	cacheIndex := strings.Index(body, "permissionTablesReadyCached()")
	migratorIndex := strings.Index(body, "db.Migrator().HasTable")
	if cacheIndex < 0 {
		t.Fatalf("EnsurePermissionSchemaContext must check permissionTablesReadyCached before schema probing")
	}
	if migratorIndex < 0 {
		t.Fatalf("EnsurePermissionSchemaContext must still probe schema when readiness is unknown")
	}
	if cacheIndex > migratorIndex {
		t.Fatalf("EnsurePermissionSchemaContext must use cached readiness before calling db.Migrator()")
	}
}

func TestPermissionServiceCachesSubjectPermissionSets(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"subjectPermissionSetCacheTTL",
		"type subjectPermissionSetCacheEntry struct",
		"getSubjectPermissionSetCache(userID, roleIDs, prefixes)",
		"setSubjectPermissionSetCache(userID, roleIDs, prefixes, allowed, denied)",
		"invalidateSubjectPermissionSetCache",
		"copyBoolMap",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must cache subject permission sets with %s", snippet)
		}
	}

	body := testFunctionBody(t, text, "subjectPermissionSetsByRoleIDsAndPrefixes")
	for _, snippet := range []string{
		"if cachedAllowed, cachedDenied, ok := getSubjectPermissionSetCache(userID, roleIDs, prefixes); ok",
		"setSubjectPermissionSetCache(userID, roleIDs, prefixes, allowed, denied)",
		"Select(permissionGrantKeySelectColumns)",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("subjectPermissionSetsByRoleIDsAndPrefixes must use cached prefix snapshots with %s", snippet)
		}
	}
}

func TestPermissionServiceAggregatesMultipleUserRoles(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func ActiveRoleIDsForUserContext(ctx context.Context, db *gorm.DB, userID, primaryRoleID uint) ([]uint, error)",
		"func SubjectHasPermissionWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, key string) (bool, error)",
		"func SubjectAPIPermissionReadyWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, platform string) (bool, error)",
		"func SubjectMenuPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, platform string) ([]string, bool, error)",
		"func DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) ([]string, bool, error)",
		"normalizeRoleIDs(roleIDs...)",
		"model.UserRole",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must aggregate multiple user roles with %s", snippet)
		}
	}

	roleBody := testFunctionBody(t, text, "ActiveRoleIDsForUserContext")
	for _, snippet := range []string{
		"Table(\"user_roles AS ur\")",
		"JOIN roles r ON r.id = ur.user_role_role_id AND r.role_status = 1",
		"Select(\"ur.user_role_role_id AS user_role_role_id\")",
	} {
		if !strings.Contains(roleBody, snippet) {
			t.Fatalf("ActiveRoleIDsForUserContext must only aggregate enabled role bindings with %s", snippet)
		}
	}

	body := testFunctionBody(t, text, "subjectPermissionSetsByRoleIDsAndPrefixes")
	for _, snippet := range []string{
		"Where(\"`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_status` = 1\", SubjectRole, roleIDs)",
		"getSubjectPermissionSetCache(userID, roleIDs, prefixes)",
		"setSubjectPermissionSetCache(userID, roleIDs, prefixes, allowed, denied)",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("multi-role permission snapshots must use %s", snippet)
		}
	}
}

func TestPermissionServiceAvoidsRepeatedAPIPermissionQueries(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	hasBody := testFunctionBody(t, text, "SubjectHasPermissionWithRoleIDsContext")
	for _, snippet := range []string{
		"permissionLookupPrefixesForKey(key)",
		"subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, prefixes)",
	} {
		if !strings.Contains(hasBody, snippet) {
			t.Fatalf("SubjectHasPermissionWithRoleIDsContext must reuse cached prefix snapshots with %s", snippet)
		}
	}

	effectBody := testFunctionBody(t, text, "SubjectPermissionEffectContext")
	for _, snippet := range []string{
		"permissionLookupPrefixesForKey(key)",
		"subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, prefixes)",
	} {
		if !strings.Contains(effectBody, snippet) {
			t.Fatalf("SubjectPermissionEffectContext must reuse cached prefix snapshots with %s", snippet)
		}
	}

	readyBody := testFunctionBody(t, text, "SubjectAPIPermissionReadyWithRoleIDsContext")
	for _, snippet := range []string{
		"subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, []string{prefix})",
		"len(allowed) > 0 || len(denied) > 0",
	} {
		if !strings.Contains(readyBody, snippet) {
			t.Fatalf("SubjectAPIPermissionReadyWithRoleIDsContext must reuse cached API permission snapshots with %s", snippet)
		}
	}
	if strings.Contains(readyBody, ".Count(&count)") {
		t.Fatalf("SubjectAPIPermissionReadyContext must not run an extra COUNT query per request")
	}
}

func TestPermissionGrantMutationsInvalidateSubjectPermissionSetCache(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, name := range []string{
		"SetRoleApplicationMenuPermissionsTx",
		"SetRoleApplicationAPIPermissionsTx",
		"setRoleAdminPermissionKeysTx",
		"SetUserAdminPermissionOverridesTx",
		"SetUserApplicationMenuPermissionOverridesTx",
		"SetUserDataScopeExtrasTx",
	} {
		body := testFunctionBody(t, text, name)
		if !strings.Contains(body, "invalidateSubjectPermissionSetCache()") {
			t.Fatalf("%s must invalidate cached subject permission sets after grant changes", name)
		}
	}
}

func TestDataScopeContextFetchesUserAndRoleGrantsTogether(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"type permissionSubjectRef struct",
		"func grantsBySubjectsAndKeys",
		"func DataScopeWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint)",
		"grantsBySubjectsAndKeys(ctx, db, keys",
		"subjectType: SubjectUser",
		"subjectType: SubjectRole",
		"roleIDSet := uintSet(roleIDs)",
		"roleIDSet[grant.SubjectID]",
		"mergedDataScope",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("data scope should batch user and multi-role grant lookups with %s", snippet)
		}
	}

	body := text
	if start := strings.Index(body, "func DataScopeWithRoleIDsContext"); start >= 0 {
		body = body[start:]
		if end := strings.Index(body, "\n}\n\nfunc "); end >= 0 {
			body = body[:end+3]
		}
	}
	if strings.Contains(body, "grantsBySubjectAndKeys(") {
		t.Fatalf("DataScopeWithRoleIDsContext should not query user and role data scope grants separately")
	}

	compatBody := testFunctionBody(t, text, "DataScopeContext")
	for _, snippet := range []string{
		"roleIDs, err := ActiveRoleIDsForUserContext(ctx, db, userID, roleID)",
		"DataScopeWithRoleIDsContext(ctx, db, userID, roleIDs)",
	} {
		if !strings.Contains(compatBody, snippet) {
			t.Fatalf("DataScopeContext compatibility wrapper must resolve active roles with %s", snippet)
		}
	}
}

func TestDataScopeExtrasContextMergesUserAndRoleExtras(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func DataScopeExtrasContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (DataScopeExtras, error)",
		"func DataScopeExtrasWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) (DataScopeExtras, error)",
		"grantsBySubjectsAndKeys(ctx, db, []string{DataExtraPermissionKey}",
		"subjectType: SubjectUser",
		"subjectType: SubjectRole",
		"mergeDataScopeExtras(grants)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("data scope extras should batch user and multi-role grant lookups with %s", snippet)
		}
	}

	compatBody := testFunctionBody(t, text, "DataScopeExtrasContext")
	for _, snippet := range []string{
		"roleIDs, err := ActiveRoleIDsForUserContext(ctx, db, userID, roleID)",
		"DataScopeExtrasWithRoleIDsContext(ctx, db, userID, roleIDs)",
	} {
		if !strings.Contains(compatBody, snippet) {
			t.Fatalf("DataScopeExtrasContext compatibility wrapper must resolve active roles with %s", snippet)
		}
	}
}

func TestDataScopeBundleContextFetchesBaseAndExtraGrantsTogether(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func DataScopeBundleContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (DataScope, DataScopeExtras, error)",
		"func DataScopeBundleWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) (DataScope, DataScopeExtras, error)",
		"keys := []string{DataAllPermissionKey, DataDeptPermissionKey, DataSelfPermissionKey, DataCustomPermissionKey, DataExtraPermissionKey}",
		"grantsBySubjectsAndKeys(ctx, db, keys",
		"baseGrants := make([]model.PermissionGrant, 0, len(grants))",
		"extraGrants := make([]model.PermissionGrant, 0)",
		"mergeDataScopeExtras(extraGrants)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("data scope bundle should batch base and extra grant lookups with %s", snippet)
		}
	}

	body := testFunctionBody(t, text, "DataScopeBundleWithRoleIDsContext")
	if strings.Count(body, "grantsBySubjectsAndKeys(ctx, db") != 1 {
		t.Fatalf("DataScopeBundleWithRoleIDsContext should query permission grants once, body=%s", body)
	}
}

func TestMergeDataScopeExtrasUnionsAllowGrants(t *testing.T) {
	grants := []model.PermissionGrant{
		{
			PermissionKey: DataExtraPermissionKey,
			Effect:        EffectAllow,
			ScopeValue:    `{"deptIds":[2],"userIds":[10]}`,
		},
		{
			PermissionKey: DataExtraPermissionKey,
			Effect:        EffectAllow,
			ScopeValue:    `{"deptIds":[3,2],"userIds":[11]}`,
		},
		{
			PermissionKey: DataExtraPermissionKey,
			Effect:        EffectDeny,
			ScopeValue:    `{"deptIds":[99],"userIds":[99]}`,
		},
	}
	extras := mergeDataScopeExtras(grants)
	if !extras.Ready {
		t.Fatalf("merged extras should be ready when allow grants exist")
	}
	if got, want := extras.DeptIDs, []uint{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("merged extra dept ids = %#v, want %#v", got, want)
	}
	if got, want := extras.UserIDs, []uint{10, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("merged extra user ids = %#v, want %#v", got, want)
	}
}

func TestMergedDataScopeKeepsWidestRoleScope(t *testing.T) {
	grants := []model.PermissionGrant{
		{PermissionKey: DataSelfPermissionKey, Effect: EffectAllow},
		{PermissionKey: DataCustomPermissionKey, Effect: EffectAllow, ScopeValue: `{"deptIds":[3,2]}`},
		{PermissionKey: DataCustomPermissionKey, Effect: EffectAllow, ScopeValue: `{"deptIds":[5,3]}`},
	}
	scope, ok := mergedDataScope(grants)
	if !ok {
		t.Fatalf("merged data scope should be ready")
	}
	if scope.Mode != 4 {
		t.Fatalf("merged data scope mode = %d, want custom", scope.Mode)
	}
	got := map[uint]bool{}
	for _, id := range scope.DeptIDs {
		got[id] = true
	}
	for _, id := range []uint{2, 3, 5} {
		if !got[id] {
			t.Fatalf("merged custom data scope must include dept %d, got %#v", id, scope.DeptIDs)
		}
	}

	scope, ok = mergedDataScope(append(grants, model.PermissionGrant{PermissionKey: DataDeptPermissionKey, Effect: EffectAllow}))
	if !ok || scope.Mode != 2 {
		t.Fatalf("data:dept should be wider than data:custom/self, got scope=%#v ready=%v", scope, ok)
	}

	scope, ok = mergedDataScope(append(grants, model.PermissionGrant{PermissionKey: DataAllPermissionKey, Effect: EffectAllow}))
	if !ok || scope.Mode != 1 {
		t.Fatalf("data:all should win when any role grants all data, got scope=%#v ready=%v", scope, ok)
	}
}

func TestPermissionServiceLimitsRoleAssignmentGrantColumns(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"permissionGrantRoleAssignmentSelectColumns",
		"Select(permissionGrantRoleAssignmentSelectColumns)",
		"`grant_subject_id`, `grant_permission_key`, `grant_scope_value`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role assignment grant query must avoid SELECT * with %s", snippet)
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
	if !strings.Contains(text, "setRoleAdminPermissionKeysTx(tx, roleID, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs, false)") {
		t.Fatalf("permission service must keep unified role grant writer")
	}
}

func TestPermissionServiceDoesNotSyncCatalogDuringRoleGrantSave(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	adminBody := testFunctionBody(t, text, "SetRoleAdminPermissionKeysTx")
	for _, snippet := range []string{
		"setRoleAdminPermissionKeysTx(tx, roleID, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs, false)",
	} {
		if !strings.Contains(adminBody, snippet) {
			t.Fatalf("role admin grant save must skip request-path catalog sync with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"ensureBuiltinPermissions(tx)",
		"syncAdminAPIPermissions(tx)",
	} {
		if strings.Contains(adminBody, forbidden) {
			t.Fatalf("role admin grant save must not run slow catalog sync snippet %s", forbidden)
		}
	}
}

func TestEnsureApplicationCatalogBackfillsAdminAPIWithoutFullSync(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := testFunctionBody(t, text, "EnsureApplicationPermissionCatalogContext")
	if !strings.Contains(body, "ensureMissingAdminAPIPermissionsContext(ctx, db)") {
		t.Fatalf("admin API catalog read path must only backfill missing permissions")
	}
	if strings.Contains(body, "EnsurePermissionSchemaContext(ctx, db)") {
		t.Fatalf("application permission catalog read path must not run schema probing during permission tree listing")
	}
	if strings.Contains(body, "syncAdminAPIPermissions(db)") {
		t.Fatalf("admin API catalog read path must not run full sync/upsert during permission tree listing")
	}
	helperBody := testFunctionBody(t, text, "ensureMissingAdminAPIPermissionsContext")
	for _, snippet := range []string{
		"adminrouteperm.Categories()",
		"adminrouteperm.Declarations()",
		"TypeAPICategory",
		"TypeAPI",
		"createMissingPermissionsContext(ctx, db, items)",
	} {
		if !strings.Contains(helperBody, snippet) {
			t.Fatalf("admin API missing-permission helper must batch backfill with %s", snippet)
		}
	}
	if strings.Contains(helperBody, "upsertPermission") {
		t.Fatalf("admin API missing-permission helper must not upsert permissions one by one")
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
		"ParentKey:    declaration.CategoryKey",
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
		"EnsureApplicationPermissionCatalogContext",
		"syncClientMenuPermissions",
		"syncDingTalkH5MenuPermissions",
		"syncDingTalkH5ButtonPermissions",
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

func TestRoleApplicationPermissionAssignmentSelfHealsCatalog(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := testFunctionBody(t, text, "SetRoleApplicationMenuPermissionsTx")
	for _, snippet := range []string{
		"ensureApplicationPermissionCatalogForKeysTx(tx, keys)",
		"replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationMenuPermissionPrefixes(), keys, EffectAllow, nil, \"form\")",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("role application menu assignment must self-heal permission catalog with %s", snippet)
		}
	}
}

func TestPermissionServiceBatchesUserApplicationPermissionOverrides(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := testFunctionBody(t, text, "SetUserApplicationMenuPermissionOverridesTx")
	for _, snippet := range []string{
		"replaceSubjectGrantsByEffectsTx",
		"normalizedAllowKeys, normalizedDenyKeys := normalizeUserApplicationPermissionKeySets(allowKeys, denyKeys)",
		"replaceSubjectGrantsByEffectsTx(tx, SubjectUser, userID, prefixes, normalizedAllowKeys, normalizedDenyKeys, \"form\")",
		"buildSubjectGrantsByEffectsTx",
		"CreateInBatches(grants, 100)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user application permission overrides must batch grants with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"EnsurePermissionSchemaContext",
		"syncClientMenuPermissions",
		"syncDingTalkH5MenuPermissions",
		"syncClientAPIPermissions",
		"syncDingTalkH5APIPermissions",
		"createGrantTx",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("user application permission override save must not run slow request-path snippet %s", forbidden)
		}
	}
	batchBody := testFunctionBody(t, text, "replaceSubjectGrantsByEffectsTx")
	if strings.Count(batchBody, "buildSubjectGrantsTx(") > 0 {
		t.Fatalf("replaceSubjectGrantsByEffectsTx must not query permission metadata separately for allow and deny grants")
	}
	grantBody := testFunctionBody(t, text, "buildSubjectGrantsByEffectsTx")
	for _, snippet := range []string{
		"keys := normalizePermissionKeys(append(append([]string{}, allowKeys...), denyKeys...))",
		"Select(\"`id`, `permission_key`\")",
		"EffectAllow",
		"EffectDeny",
	} {
		if !strings.Contains(grantBody, snippet) {
			t.Fatalf("buildSubjectGrantsByEffectsTx must build allow/deny grants from one permission lookup with %s", snippet)
		}
	}
}

func TestPermissionServiceReplacesRoleAdminLoginGrant(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	adminPrefixesBody := testFunctionBody(t, text, "AdminPermissionPrefixes")
	if !strings.Contains(adminPrefixesBody, "AdminLoginPermissionKey") {
		t.Fatalf("role admin permission replacement must delete old admin:login grants before inserting new ones")
	}
}

func TestPermissionServiceUpsertsReplacementGrantBatches(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`"gorm.io/gorm/clause"`,
		"createGrantBatchTx",
		"clause.OnConflict",
		"`grant_subject_type`",
		"`grant_subject_id`",
		"`grant_permission_key`",
		"DoUpdates: clause.AssignmentColumns",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("replacement grant batches must use upsert to avoid duplicate-key failures with %s", snippet)
		}
	}
	for _, name := range []string{"replaceSubjectGrantsTx", "replaceSubjectGrantsByEffectsTx"} {
		body := testFunctionBody(t, text, name)
		if !strings.Contains(body, "createGrantBatchTx(tx, grants)") {
			t.Fatalf("%s must write grants through duplicate-safe batch helper", name)
		}
		if strings.Contains(body, "CreateInBatches(grants, 100)") {
			t.Fatalf("%s must not insert raw grant batches that can hit duplicate keys", name)
		}
	}
}

func testFunctionBody(t *testing.T, text, name string) string {
	t.Helper()
	start := strings.Index(text, "func "+name)
	if start < 0 {
		t.Fatalf("missing function %s", name)
	}
	next := strings.Index(text[start+1:], "\nfunc ")
	if next < 0 {
		return text[start:]
	}
	return text[start : start+1+next]
}
