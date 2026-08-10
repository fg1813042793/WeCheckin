package permission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminmenuperm"
	"wecheckin/backend/internal/support/adminrouteperm"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
	"wecheckin/backend/pkg/database"
)

const (
	SubjectRole = "role"
	SubjectUser = "user"

	EffectAllow = "allow"
	EffectDeny  = "deny"

	PlatformAdmin      = "admin"
	PlatformClient     = "client"
	PlatformDingTalkH5 = "dingtalk_h5"
	PlatformData       = "data"

	TypeLogin       = "login"
	TypeDirectory   = "directory"
	TypeMenu        = "menu"
	TypeButton      = "button"
	TypeAPI         = "api"
	TypeAPICategory = "api_category"
	TypeData        = "data"

	AdminLoginPermissionKey = "admin:login"
	DataAllPermissionKey    = "data:all"
	DataDeptPermissionKey   = "data:dept"
	DataSelfPermissionKey   = "data:self"
	DataCustomPermissionKey = "data:custom"
	DataExtraPermissionKey  = "data:extra"
)

const (
	dingtalkH5MenuPermissionCacheTTL           = 30 * time.Second
	subjectPermissionSetCacheTTL               = 30 * time.Second
	permissionTablesReadyNegativeCacheTTL      = 5 * time.Second
	userRolesTableReadyNegativeCacheTTL        = 5 * time.Second
	permissionGrantRoleAssignmentSelectColumns = "`grant_subject_id`, `grant_permission_key`, `grant_scope_value`"
	permissionGrantKeySelectColumns            = "`grant_subject_id`, `grant_permission_key`, `grant_effect`"
	permissionGrantScopeSelectColumns          = "`grant_subject_id`, `grant_scope_value`"
)

type dingtalkH5MenuPermissionCacheEntry struct {
	keys      []string
	ready     bool
	expiresAt time.Time
}

var dingtalkH5MenuPermissionCache = struct {
	sync.RWMutex
	items map[string]dingtalkH5MenuPermissionCacheEntry
}{
	items: map[string]dingtalkH5MenuPermissionCacheEntry{},
}

type subjectPermissionSetCacheEntry struct {
	allowed   map[string]bool
	denied    map[string]bool
	expiresAt time.Time
}

var subjectPermissionSetCache = struct {
	sync.RWMutex
	items map[string]subjectPermissionSetCacheEntry
}{
	items: map[string]subjectPermissionSetCacheEntry{},
}

var permissionTablesReadyCache = struct {
	sync.RWMutex
	checked     bool
	ready       bool
	schemaReady bool
	checkedAt   time.Time
}{}

var userRolesTableReadyCache = struct {
	sync.RWMutex
	checked   bool
	ready     bool
	checkedAt time.Time
}{}

type DataScope struct {
	Mode    int
	DeptIDs []uint
	Ready   bool
}

type DataScopeExtras struct {
	DeptIDs []uint
	UserIDs []uint
	Ready   bool
}

type permissionSubjectRef struct {
	subjectType string
	subjectID   uint
}

type RoleAssignmentMaps struct {
	AdminPermissionKeys         map[uint][]string
	AdminAPIPermissionKeys      map[uint][]string
	DeptIDs                     map[uint][]uint
	ClientMenuKeys              map[uint][]string
	DingTalkH5MenuKeys          map[uint][]string
	ClientAPIPermissionKeys     map[uint][]string
	DingTalkH5APIPermissionKeys map[uint][]string
}

func EnsureUnifiedPermissionsContext(ctx context.Context, db *gorm.DB, enableExam ...bool) error {
	if err := EnsurePermissionSchemaContext(ctx, db); err != nil {
		return err
	}
	if err := ensureBuiltinPermissions(db); err != nil {
		return err
	}
	if err := SyncAdminMenuPermissionsContext(ctx, db, firstBool(enableExam, true)); err != nil {
		return err
	}
	if err := syncAdminAPIPermissions(db); err != nil {
		return err
	}
	if err := syncClientMenuPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5MenuPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5ButtonPermissions(db); err != nil {
		return err
	}
	if err := syncClientAPIPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5APIPermissions(db); err != nil {
		return err
	}
	return ctxErr(ctx)
}

func EnsureApplicationPermissionCatalogContext(ctx context.Context, db *gorm.DB, platform string, types []string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	platform = strings.TrimSpace(platform)
	typeSet := permissionTypeSet(types)
	shouldInclude := func(values ...string) bool {
		if len(typeSet) == 0 {
			return true
		}
		for _, value := range values {
			if typeSet[value] {
				return true
			}
		}
		return false
	}
	if platform == "" || platform == PlatformAdmin {
		if shouldInclude(TypeAPICategory, TypeAPI) {
			if err := ensureMissingAdminAPIPermissionsContext(ctx, db); err != nil {
				return err
			}
		}
	}
	if platform == "" || platform == PlatformClient {
		if shouldInclude(TypeDirectory, TypeMenu, TypeButton) {
			if err := ensureMissingApplicationMenuPermissionsContext(ctx, db, appmenuperm.ClientMenuDeclarations()); err != nil {
				return err
			}
		}
		if shouldInclude(TypeAPICategory, TypeAPI) {
			if err := ensureMissingApplicationAPIPermissionsContext(ctx, db, appapiperm.ClientAPICategories(), appapiperm.ClientAPIDeclarations()); err != nil {
				return err
			}
		}
	}
	if platform == "" || platform == PlatformDingTalkH5 {
		if shouldInclude(TypeDirectory, TypeMenu) {
			if err := ensureMissingApplicationMenuPermissionsContext(ctx, db, appmenuperm.DingTalkH5MenuDeclarations()); err != nil {
				return err
			}
		}
		if shouldInclude(TypeButton) {
			if err := ensureMissingApplicationButtonPermissionsContext(ctx, db, appmenuperm.DingTalkH5ButtonDeclarations()); err != nil {
				return err
			}
		}
		if shouldInclude(TypeAPICategory, TypeAPI) {
			if err := ensureMissingApplicationAPIPermissionsContext(ctx, db, appapiperm.DingTalkH5APICategories(), appapiperm.DingTalkH5APIDeclarations()); err != nil {
				return err
			}
		}
	}
	return ctxErr(ctx)
}

func permissionTypeSet(types []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range types {
		for _, part := range strings.Split(value, ",") {
			item := strings.TrimSpace(part)
			if item != "" {
				result[item] = true
			}
		}
	}
	return result
}

func EnsurePermissionSchemaContext(ctx context.Context, db *gorm.DB) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if permissionTablesReadyCached() {
		return nil
	}
	if !db.Migrator().HasTable(&model.Permission{}) {
		markPermissionTablesReady(false)
		return ctxErr(ctx)
	}
	if !db.Migrator().HasColumn(&model.Permission{}, "Icon") {
		if err := db.Migrator().AddColumn(&model.Permission{}, "Icon"); err != nil {
			markPermissionTablesReady(false)
			return err
		}
	}
	markPermissionSchemaReady(db.Migrator().HasTable(&model.PermissionGrant{}))
	return ctxErr(ctx)
}

func SubjectHasPermissionContext(ctx context.Context, db *gorm.DB, userID, roleID uint, key string) (bool, error) {
	return SubjectHasPermissionWithRoleIDsContext(ctx, db, userID, []uint{roleID}, key)
}

func SubjectHasPermissionWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, key string) (bool, error) {
	if err := ctxErr(ctx); err != nil {
		return false, err
	}
	if db == nil {
		return false, fmt.Errorf("数据库连接异常")
	}
	roleIDs = normalizeRoleIDs(roleIDs...)
	if !TablesReady(db) {
		if key == AdminLoginPermissionKey && len(roleIDs) > 0 {
			return false, nil
		}
		return false, nil
	}
	if prefixes := permissionLookupPrefixesForKey(key); len(prefixes) > 0 {
		allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, prefixes)
		if err != nil {
			return false, err
		}
		if denied[key] {
			return false, nil
		}
		return allowed[key], nil
	}
	if userID > 0 {
		effect, ok, err := grantEffect(ctx, db, SubjectUser, userID, key)
		if err != nil {
			return false, err
		}
		if ok {
			return effect == EffectAllow, nil
		}
	}
	if len(roleIDs) == 0 {
		return false, nil
	}
	var grants []model.PermissionGrant
	if err := db.Select(permissionGrantKeySelectColumns).
		Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_permission_key` = ? AND `grant_status` = 1", SubjectRole, roleIDs, key).
		Find(&grants).Error; err != nil {
		return false, err
	}
	allowed := false
	for _, grant := range grants {
		if grant.Effect == EffectDeny {
			return false, nil
		}
		if grant.Effect == EffectAllow {
			allowed = true
		}
	}
	return allowed, nil
}

func RoleHasPermissionContext(ctx context.Context, db *gorm.DB, roleID uint, key string) (bool, error) {
	return SubjectHasPermissionContext(ctx, db, 0, roleID, key)
}

func SubjectPermissionEffectContext(ctx context.Context, db *gorm.DB, subjectType string, subjectID uint, key string) (string, bool, error) {
	if prefixes := permissionLookupPrefixesForKey(key); len(prefixes) > 0 && db != nil && TablesReady(db) {
		var userID uint
		var roleIDs []uint
		switch subjectType {
		case SubjectUser:
			userID = subjectID
		case SubjectRole:
			roleIDs = []uint{subjectID}
		}
		if userID > 0 || len(roleIDs) > 0 {
			allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, prefixes)
			if err != nil {
				return "", false, err
			}
			if denied[key] {
				return EffectDeny, true, nil
			}
			if allowed[key] {
				return EffectAllow, true, nil
			}
			return "", false, nil
		}
	}
	return grantEffect(ctx, db, subjectType, subjectID, key)
}

func permissionLookupPrefixesForKey(key string) []string {
	key = strings.TrimSpace(key)
	switch {
	case key == AdminLoginPermissionKey:
		return []string{AdminLoginPermissionKey}
	case strings.HasPrefix(key, "admin:menu:"):
		return []string{"admin:menu:%"}
	case strings.HasPrefix(key, "admin:api:"):
		return []string{"admin:api:%"}
	case strings.HasPrefix(key, "client:menu:"):
		return []string{"client:menu:%"}
	case strings.HasPrefix(key, "client:api:"):
		return []string{"client:api:%"}
	case strings.HasPrefix(key, "dingtalk_h5:menu:"):
		return []string{"dingtalk_h5:menu:%"}
	case strings.HasPrefix(key, "dingtalk_h5:button:"):
		return []string{"dingtalk_h5:button:%"}
	case strings.HasPrefix(key, "dingtalk_h5:api:"):
		return []string{"dingtalk_h5:api:%"}
	case strings.HasPrefix(key, "data:"):
		return []string{"data:%"}
	default:
		return nil
	}
}

func UserPermissionKeySetsContext(ctx context.Context, db *gorm.DB, userID uint) ([]string, []string, error) {
	return directSubjectPermissionKeySetsContext(ctx, db, SubjectUser, userID, UserPermissionPrefixes())
}

func UserApplicationMenuPermissionKeySetsContext(ctx context.Context, db *gorm.DB, userID uint) ([]string, []string, error) {
	prefixes := append(ApplicationMenuPermissionPrefixes(), ApplicationAPIPermissionPrefixes()...)
	return directSubjectPermissionKeySetsContext(ctx, db, SubjectUser, userID, prefixes)
}

func UserApplicationPermissionKeySetsContext(ctx context.Context, db *gorm.DB, userID uint) ([]string, []string, error) {
	return directSubjectPermissionKeySetsContext(ctx, db, SubjectUser, userID, ApplicationPermissionPrefixes())
}

func directSubjectPermissionKeySetsContext(ctx context.Context, db *gorm.DB, subjectType string, subjectID uint, prefixes []string) ([]string, []string, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, nil, err
	}
	if db == nil {
		return nil, nil, fmt.Errorf("数据库连接异常")
	}
	allowKeys := make([]string, 0)
	denyKeys := make([]string, 0)
	if !TablesReady(db) || subjectID == 0 {
		return allowKeys, denyKeys, nil
	}
	query := db.Where("`grant_subject_type` = ? AND `grant_subject_id` = ? AND `grant_status` = 1", subjectType, subjectID)
	if len(prefixes) > 0 {
		where, args := likeAnyClause("`grant_permission_key`", prefixes)
		query = query.Where(where, args...)
	}
	var grants []model.PermissionGrant
	if err := query.Find(&grants).Error; err != nil {
		return nil, nil, err
	}
	for _, grant := range grants {
		if grant.Effect == EffectDeny {
			denyKeys = append(denyKeys, grant.PermissionKey)
			continue
		}
		allowKeys = append(allowKeys, grant.PermissionKey)
	}
	sort.Strings(allowKeys)
	sort.Strings(denyKeys)
	return allowKeys, denyKeys, ctxErr(ctx)
}

func AdminMenuPermissionsContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]model.Permission, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, false, err
	}
	if db == nil {
		return nil, false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return nil, false, nil
	}
	allowed, denied, err := subjectAdminPermissionSets(ctx, db, userID, roleID)
	if err != nil {
		return nil, true, err
	}
	keys := make([]string, 0, len(allowed))
	for key := range allowed {
		if denied[key] || !strings.HasPrefix(key, "admin:menu:") {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return []model.Permission{}, true, nil
	}
	var rows []model.Permission
	if err := db.Where("`permission_key` IN ? AND `permission_platform` = ? AND `permission_type` IN ? AND `permission_status` = 1", keys, PlatformAdmin, []string{TypeDirectory, TypeMenu, TypeButton}).
		Order("`permission_sort` ASC, `id` ASC").
		Find(&rows).Error; err != nil {
		return nil, true, err
	}
	return rows, true, ctxErr(ctx)
}

func RoleAdminPermissionKeyMapContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (map[uint][]string, error) {
	return rolePermissionKeyMapContext(ctx, db, roleIDs, "admin:menu:%")
}

func RoleAdminAPIPermissionKeyMapContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (map[uint][]string, error) {
	return rolePermissionKeyMapContext(ctx, db, roleIDs, "admin:api:%")
}

func RoleApplicationAPIKeyMapContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (map[uint][]string, map[uint][]string, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, nil, err
	}
	clientKeysByRole := make(map[uint][]string, len(roleIDs))
	dingTalkH5KeysByRole := make(map[uint][]string, len(roleIDs))
	for _, roleID := range roleIDs {
		clientKeysByRole[roleID] = []string{}
		dingTalkH5KeysByRole[roleID] = []string{}
	}
	if db == nil || !TablesReady(db) || len(roleIDs) == 0 {
		return clientKeysByRole, dingTalkH5KeysByRole, nil
	}
	var grants []model.PermissionGrant
	where, args := likeAnyClause("`grant_permission_key`", ApplicationAPIPermissionPrefixes())
	if err := db.Select(permissionGrantKeySelectColumns).
		Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, EffectAllow).
		Where(where, args...).
		Find(&grants).Error; err != nil {
		return nil, nil, err
	}
	clientSets := make(map[uint]map[string]bool, len(roleIDs))
	dingTalkH5Sets := make(map[uint]map[string]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		clientSets[roleID] = map[string]bool{}
		dingTalkH5Sets[roleID] = map[string]bool{}
	}
	for _, grant := range grants {
		if strings.HasPrefix(grant.PermissionKey, "client:api:") {
			clientSets[grant.SubjectID][grant.PermissionKey] = true
		}
		if strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:api:") {
			dingTalkH5Sets[grant.SubjectID][grant.PermissionKey] = true
		}
	}
	for _, roleID := range roleIDs {
		clientKeysByRole[roleID] = orderedApplicationAPIKeys(clientSets[roleID], appapiperm.ClientAPIDeclarations())
		dingTalkH5KeysByRole[roleID] = orderedApplicationAPIKeys(dingTalkH5Sets[roleID], appapiperm.DingTalkH5APIDeclarations())
	}
	return clientKeysByRole, dingTalkH5KeysByRole, ctxErr(ctx)
}

func RoleAssignmentMapsContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (RoleAssignmentMaps, error) {
	result := newRoleAssignmentMaps(roleIDs)
	if err := ctxErr(ctx); err != nil {
		return result, err
	}
	if db == nil || !TablesReady(db) || len(roleIDs) == 0 {
		return result, nil
	}
	prefixes := []string{"admin:menu:%", "admin:api:%"}
	prefixes = append(prefixes, ApplicationMenuPermissionPrefixes()...)
	prefixes = append(prefixes, ApplicationAPIPermissionPrefixes()...)
	where, args := likeAnyClause("`grant_permission_key`", prefixes)
	args = append(args, DataCustomPermissionKey)

	var grants []model.PermissionGrant
	if err := db.Select(permissionGrantRoleAssignmentSelectColumns).
		Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, EffectAllow).
		Where("("+where+") OR `grant_permission_key` = ?", args...).
		Order("`id` ASC").
		Find(&grants).Error; err != nil {
		return result, err
	}

	clientMenuSets := make(map[uint]map[string]bool, len(roleIDs))
	dingTalkH5MenuSets := make(map[uint]map[string]bool, len(roleIDs))
	clientAPISets := make(map[uint]map[string]bool, len(roleIDs))
	dingTalkH5APISets := make(map[uint]map[string]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		clientMenuSets[roleID] = map[string]bool{}
		dingTalkH5MenuSets[roleID] = map[string]bool{}
		clientAPISets[roleID] = map[string]bool{}
		dingTalkH5APISets[roleID] = map[string]bool{}
	}

	for _, grant := range grants {
		switch {
		case strings.HasPrefix(grant.PermissionKey, "admin:menu:"):
			result.AdminPermissionKeys[grant.SubjectID] = append(result.AdminPermissionKeys[grant.SubjectID], grant.PermissionKey)
		case strings.HasPrefix(grant.PermissionKey, "admin:api:"):
			result.AdminAPIPermissionKeys[grant.SubjectID] = append(result.AdminAPIPermissionKeys[grant.SubjectID], grant.PermissionKey)
		case grant.PermissionKey == DataCustomPermissionKey:
			result.DeptIDs[grant.SubjectID] = decodeDeptScope(grant.ScopeValue)
		case strings.HasPrefix(grant.PermissionKey, "client:menu:"):
			clientMenuSets[grant.SubjectID][grant.PermissionKey] = true
		case strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:menu:"), strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:button:"):
			dingTalkH5MenuSets[grant.SubjectID][grant.PermissionKey] = true
		case strings.HasPrefix(grant.PermissionKey, "client:api:"):
			clientAPISets[grant.SubjectID][grant.PermissionKey] = true
		case strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:api:"):
			dingTalkH5APISets[grant.SubjectID][grant.PermissionKey] = true
		}
	}

	for _, roleID := range roleIDs {
		result.ClientMenuKeys[roleID] = orderedApplicationMenuKeys(clientMenuSets[roleID], appmenuperm.ClientMenuDeclarations())
		result.DingTalkH5MenuKeys[roleID] = orderedApplicationMenuKeys(dingTalkH5MenuSets[roleID], appmenuperm.DingTalkH5PermissionDeclarations())
		result.ClientAPIPermissionKeys[roleID] = orderedApplicationAPIKeys(clientAPISets[roleID], appapiperm.ClientAPIDeclarations())
		result.DingTalkH5APIPermissionKeys[roleID] = orderedApplicationAPIKeys(dingTalkH5APISets[roleID], appapiperm.DingTalkH5APIDeclarations())
	}
	return result, ctxErr(ctx)
}

func newRoleAssignmentMaps(roleIDs []uint) RoleAssignmentMaps {
	result := RoleAssignmentMaps{
		AdminPermissionKeys:         make(map[uint][]string, len(roleIDs)),
		AdminAPIPermissionKeys:      make(map[uint][]string, len(roleIDs)),
		DeptIDs:                     make(map[uint][]uint, len(roleIDs)),
		ClientMenuKeys:              make(map[uint][]string, len(roleIDs)),
		DingTalkH5MenuKeys:          make(map[uint][]string, len(roleIDs)),
		ClientAPIPermissionKeys:     make(map[uint][]string, len(roleIDs)),
		DingTalkH5APIPermissionKeys: make(map[uint][]string, len(roleIDs)),
	}
	for _, roleID := range roleIDs {
		result.AdminPermissionKeys[roleID] = []string{}
		result.AdminAPIPermissionKeys[roleID] = []string{}
		result.DeptIDs[roleID] = []uint{}
		result.ClientMenuKeys[roleID] = []string{}
		result.DingTalkH5MenuKeys[roleID] = []string{}
		result.ClientAPIPermissionKeys[roleID] = []string{}
		result.DingTalkH5APIPermissionKeys[roleID] = []string{}
	}
	return result
}

func rolePermissionKeyMapContext(ctx context.Context, db *gorm.DB, roleIDs []uint, like string) (map[uint][]string, error) {
	result := make(map[uint][]string, len(roleIDs))
	for _, roleID := range roleIDs {
		result[roleID] = []string{}
	}
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if db == nil || !TablesReady(db) || len(roleIDs) == 0 {
		return result, nil
	}
	var grants []model.PermissionGrant
	if err := db.Select(permissionGrantKeySelectColumns).
		Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_permission_key` LIKE ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, like, EffectAllow).
		Order("`id` ASC").
		Find(&grants).Error; err != nil {
		return nil, err
	}
	for _, grant := range grants {
		result[grant.SubjectID] = append(result[grant.SubjectID], grant.PermissionKey)
	}
	return result, nil
}

func RoleCustomDeptIDMapContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (map[uint][]uint, error) {
	result := make(map[uint][]uint, len(roleIDs))
	for _, roleID := range roleIDs {
		result[roleID] = []uint{}
	}
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if db == nil || !TablesReady(db) || len(roleIDs) == 0 {
		return result, nil
	}
	var grants []model.PermissionGrant
	if err := db.Select(permissionGrantScopeSelectColumns).
		Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_permission_key` = ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, DataCustomPermissionKey, EffectAllow).Find(&grants).Error; err != nil {
		return nil, err
	}
	for _, grant := range grants {
		result[grant.SubjectID] = decodeDeptScope(grant.ScopeValue)
	}
	return result, nil
}

func RoleCustomDeptIDsContext(ctx context.Context, db *gorm.DB, roleID uint) ([]uint, error) {
	deptIDsByRole, err := RoleCustomDeptIDMapContext(ctx, db, []uint{roleID})
	if err != nil {
		return nil, err
	}
	return deptIDsByRole[roleID], nil
}

func ActiveRoleIDsForUserContext(ctx context.Context, db *gorm.DB, userID, primaryRoleID uint) ([]uint, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if db == nil || userID == 0 || !UserRolesTableReady(db) {
		return normalizeRoleIDs(primaryRoleID), nil
	}
	roleIDs := []uint{}
	var rows []model.UserRole
	if err := db.Table("user_roles AS ur").
		Select("ur.user_role_role_id AS user_role_role_id").
		Joins("JOIN roles r ON r.id = ur.user_role_role_id AND r.role_status = 1").
		Where("ur.user_role_user_id = ? AND ur.user_role_status = 1", userID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		roleIDs = append(roleIDs, row.RoleID)
	}
	if len(roleIDs) == 0 && primaryRoleID > 0 {
		var count int64
		if err := db.Model(&model.Role{}).Where("`id` = ? AND `role_status` = 1", primaryRoleID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			roleIDs = append(roleIDs, primaryRoleID)
		}
	}
	return normalizeRoleIDs(roleIDs...), ctxErr(ctx)
}

func ClientMenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return SubjectMenuPermissionKeysContext(ctx, db, userID, roleID, PlatformClient)
}

func DingTalkH5MenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx, db, userID, []uint{roleID})
}

func DingTalkH5MenuPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) ([]string, bool, error) {
	roleIDs = normalizeRoleIDs(roleIDs...)
	if keys, ready, ok := getDingTalkH5MenuPermissionCache(userID, roleIDs); ok {
		return keys, ready, nil
	}
	keys, ready, err := SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, userID, roleIDs, PlatformDingTalkH5)
	if err == nil {
		setDingTalkH5MenuPermissionCache(userID, roleIDs, keys, ready)
	}
	return keys, ready, err
}

func DingTalkH5ButtonPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx, db, userID, []uint{roleID})
}

func DingTalkH5ButtonPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) ([]string, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, false, err
	}
	if db == nil {
		return nil, false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return nil, false, nil
	}
	allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, []string{"dingtalk_h5:button:%"})
	if err != nil {
		return nil, true, err
	}
	selected := make(map[string]bool, len(allowed))
	for key := range allowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	return orderedApplicationMenuKeys(selected, appmenuperm.DingTalkH5ButtonDeclarations()), true, nil
}

func dingtalkH5MenuPermissionCacheKey(userID uint, roleIDs []uint) string {
	roleIDs = normalizeRoleIDs(roleIDs...)
	roleParts := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		roleParts = append(roleParts, strconv.FormatUint(uint64(roleID), 10))
	}
	return strconv.FormatUint(uint64(userID), 10) + ":" + strings.Join(roleParts, ",")
}

func getDingTalkH5MenuPermissionCache(userID uint, roleIDs []uint) ([]string, bool, bool) {
	key := dingtalkH5MenuPermissionCacheKey(userID, roleIDs)
	now := time.Now()
	dingtalkH5MenuPermissionCache.RLock()
	entry, ok := dingtalkH5MenuPermissionCache.items[key]
	dingtalkH5MenuPermissionCache.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, false, false
	}
	return append([]string(nil), entry.keys...), entry.ready, true
}

func setDingTalkH5MenuPermissionCache(userID uint, roleIDs []uint, keys []string, ready bool) {
	key := dingtalkH5MenuPermissionCacheKey(userID, roleIDs)
	dingtalkH5MenuPermissionCache.Lock()
	dingtalkH5MenuPermissionCache.items[key] = dingtalkH5MenuPermissionCacheEntry{
		keys:      append([]string(nil), keys...),
		ready:     ready,
		expiresAt: time.Now().Add(dingtalkH5MenuPermissionCacheTTL),
	}
	dingtalkH5MenuPermissionCache.Unlock()
}

func invalidateDingTalkH5MenuPermissionCache() {
	dingtalkH5MenuPermissionCache.Lock()
	dingtalkH5MenuPermissionCache.items = map[string]dingtalkH5MenuPermissionCacheEntry{}
	dingtalkH5MenuPermissionCache.Unlock()
}

func InvalidateRuntimePermissionCaches() {
	invalidateSubjectPermissionSetCache()
	invalidateDingTalkH5MenuPermissionCache()
}

func subjectPermissionSetCacheKey(userID uint, roleIDs []uint, prefixes []string) string {
	roleIDs = normalizeRoleIDs(roleIDs...)
	roleParts := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		roleParts = append(roleParts, strconv.FormatUint(uint64(roleID), 10))
	}
	normalized := normalizePermissionKeys(prefixes)
	sort.Strings(normalized)
	return strconv.FormatUint(uint64(userID), 10) + ":" +
		strings.Join(roleParts, ",") + ":" +
		strings.Join(normalized, "\x1f")
}

func getSubjectPermissionSetCache(userID uint, roleIDs []uint, prefixes []string) (map[string]bool, map[string]bool, bool) {
	key := subjectPermissionSetCacheKey(userID, roleIDs, prefixes)
	now := time.Now()
	subjectPermissionSetCache.RLock()
	entry, ok := subjectPermissionSetCache.items[key]
	subjectPermissionSetCache.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return nil, nil, false
	}
	return copyBoolMap(entry.allowed), copyBoolMap(entry.denied), true
}

func setSubjectPermissionSetCache(userID uint, roleIDs []uint, prefixes []string, allowed, denied map[string]bool) {
	key := subjectPermissionSetCacheKey(userID, roleIDs, prefixes)
	subjectPermissionSetCache.Lock()
	subjectPermissionSetCache.items[key] = subjectPermissionSetCacheEntry{
		allowed:   copyBoolMap(allowed),
		denied:    copyBoolMap(denied),
		expiresAt: time.Now().Add(subjectPermissionSetCacheTTL),
	}
	subjectPermissionSetCache.Unlock()
}

func invalidateSubjectPermissionSetCache() {
	subjectPermissionSetCache.Lock()
	subjectPermissionSetCache.items = map[string]subjectPermissionSetCacheEntry{}
	subjectPermissionSetCache.Unlock()
}

func copyBoolMap(source map[string]bool) map[string]bool {
	if len(source) == 0 {
		return map[string]bool{}
	}
	copied := make(map[string]bool, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}

func SubjectMenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint, platform string) ([]string, bool, error) {
	return SubjectMenuPermissionKeysWithRoleIDsContext(ctx, db, userID, []uint{roleID}, platform)
}

func SubjectMenuPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, platform string) ([]string, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, false, err
	}
	if db == nil {
		return nil, false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return nil, false, nil
	}
	prefix := platform + ":menu:%"
	allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, []string{prefix})
	if err != nil {
		return nil, true, err
	}
	keys := make([]string, 0, len(allowed))
	for key := range allowed {
		if !denied[key] {
			keys = append(keys, key)
		}
	}
	return keys, true, nil
}

func SubjectAPIPermissionReadyContext(ctx context.Context, db *gorm.DB, userID, roleID uint, platform string) (bool, error) {
	return SubjectAPIPermissionReadyWithRoleIDsContext(ctx, db, userID, []uint{roleID}, platform)
}

func SubjectAPIPermissionReadyWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, platform string) (bool, error) {
	if err := ctxErr(ctx); err != nil {
		return false, err
	}
	if db == nil {
		return false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return false, nil
	}
	prefix := strings.TrimSpace(platform) + ":api:%"
	if prefix == ":api:%" {
		return false, nil
	}
	roleIDs = normalizeRoleIDs(roleIDs...)
	if userID == 0 && len(roleIDs) == 0 {
		return false, nil
	}
	allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, []string{prefix})
	if err != nil {
		return false, err
	}
	return len(allowed) > 0 || len(denied) > 0, ctxErr(ctx)
}

func SubjectAPIPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint, platform string) ([]string, bool, error) {
	return SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, userID, []uint{roleID}, platform)
}

func SubjectAPIPermissionKeysWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, platform string) ([]string, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, false, err
	}
	if db == nil {
		return nil, false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return nil, false, nil
	}
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return nil, false, nil
	}
	allowed, denied, err := subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, roleIDs, []string{platform + ":api:%"})
	if err != nil {
		return nil, true, err
	}
	selected := make(map[string]bool, len(allowed))
	for key := range allowed {
		if !denied[key] {
			selected[key] = true
		}
	}
	switch platform {
	case PlatformClient:
		return orderedApplicationAPIKeys(selected, appapiperm.ClientAPIDeclarations()), true, nil
	case PlatformDingTalkH5:
		return orderedApplicationAPIKeys(selected, appapiperm.DingTalkH5APIDeclarations()), true, nil
	default:
		keys := make([]string, 0, len(selected))
		for key := range selected {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return keys, true, nil
	}
}

func RoleApplicationMenuKeyMapContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (map[uint][]string, map[uint][]string, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, nil, err
	}
	clientKeysByRole := make(map[uint][]string, len(roleIDs))
	dingTalkH5KeysByRole := make(map[uint][]string, len(roleIDs))
	for _, roleID := range roleIDs {
		clientKeysByRole[roleID] = []string{}
		dingTalkH5KeysByRole[roleID] = []string{}
	}
	if db == nil || !TablesReady(db) || len(roleIDs) == 0 {
		return clientKeysByRole, dingTalkH5KeysByRole, nil
	}
	var grants []model.PermissionGrant
	where, args := likeAnyClause("`grant_permission_key`", ApplicationMenuPermissionPrefixes())
	if err := db.Select(permissionGrantKeySelectColumns).
		Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, EffectAllow).
		Where(where, args...).
		Find(&grants).Error; err != nil {
		return nil, nil, err
	}
	clientSets := make(map[uint]map[string]bool, len(roleIDs))
	dingTalkH5Sets := make(map[uint]map[string]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		clientSets[roleID] = map[string]bool{}
		dingTalkH5Sets[roleID] = map[string]bool{}
	}
	for _, grant := range grants {
		if strings.HasPrefix(grant.PermissionKey, "client:menu:") {
			clientSets[grant.SubjectID][grant.PermissionKey] = true
		}
		if strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:menu:") || strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:button:") {
			dingTalkH5Sets[grant.SubjectID][grant.PermissionKey] = true
		}
	}
	for _, roleID := range roleIDs {
		clientKeysByRole[roleID] = orderedApplicationMenuKeys(clientSets[roleID], appmenuperm.ClientMenuDeclarations())
		dingTalkH5KeysByRole[roleID] = orderedApplicationMenuKeys(dingTalkH5Sets[roleID], appmenuperm.DingTalkH5PermissionDeclarations())
	}
	return clientKeysByRole, dingTalkH5KeysByRole, ctxErr(ctx)
}

func RoleIDsByCustomDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]uint, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if db == nil || !TablesReady(db) || len(deptIDs) == 0 {
		return nil, nil
	}
	deptSet := uintSet(deptIDs)
	var grants []model.PermissionGrant
	if err := db.Select(permissionGrantScopeSelectColumns).
		Where("`grant_subject_type` = ? AND `grant_permission_key` = ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, DataCustomPermissionKey, EffectAllow).Find(&grants).Error; err != nil {
		return nil, err
	}
	seen := map[uint]bool{}
	roleIDs := make([]uint, 0)
	for _, grant := range grants {
		if seen[grant.SubjectID] || !intersectsUintSet(decodeDeptScope(grant.ScopeValue), deptSet) {
			continue
		}
		seen[grant.SubjectID] = true
		roleIDs = append(roleIDs, grant.SubjectID)
	}
	return roleIDs, nil
}

func AdminPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, false, err
	}
	if db == nil {
		return nil, false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		return nil, false, nil
	}
	allowed, denied, err := subjectAdminPermissionSets(ctx, db, userID, roleID)
	if err != nil {
		return nil, true, err
	}
	keys := make([]string, 0, len(allowed))
	for key := range allowed {
		if !denied[key] {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil, true, nil
	}
	var rows []model.Permission
	if err := db.Select("`permission_key`").
		Where("`permission_key` IN ? AND `permission_status` = 1", keys).
		Find(&rows).Error; err != nil {
		return nil, true, err
	}
	seen := map[string]bool{}
	var result []string
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result, true, nil
}

func DataScopeContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (DataScope, error) {
	roleIDs, err := ActiveRoleIDsForUserContext(ctx, db, userID, roleID)
	if err != nil {
		return DataScope{}, err
	}
	return DataScopeWithRoleIDsContext(ctx, db, userID, roleIDs)
}

func DataScopeWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) (DataScope, error) {
	if err := ctxErr(ctx); err != nil {
		return DataScope{}, err
	}
	if db == nil || !TablesReady(db) {
		return DataScope{}, nil
	}
	roleIDs = normalizeRoleIDs(roleIDs...)
	keys := []string{DataAllPermissionKey, DataDeptPermissionKey, DataSelfPermissionKey, DataCustomPermissionKey}
	subjects := make([]permissionSubjectRef, 0, len(roleIDs)+1)
	subjects = append(subjects, permissionSubjectRef{subjectType: SubjectUser, subjectID: userID})
	for _, roleID := range roleIDs {
		subjects = append(subjects, permissionSubjectRef{subjectType: SubjectRole, subjectID: roleID})
	}
	grants, err := grantsBySubjectsAndKeys(ctx, db, keys, subjects...)
	if err != nil {
		return DataScope{}, err
	}
	userGrants := make([]model.PermissionGrant, 0)
	roleGrants := make([]model.PermissionGrant, 0)
	roleIDSet := uintSet(roleIDs)
	for _, grant := range grants {
		switch {
		case grant.SubjectType == SubjectUser && grant.SubjectID == userID:
			userGrants = append(userGrants, grant)
		case grant.SubjectType == SubjectRole && roleIDSet[grant.SubjectID]:
			roleGrants = append(roleGrants, grant)
		}
	}
	if scope, ok := mergedDataScope(userGrants); ok {
		return scope, nil
	}
	if scope, ok := mergedDataScope(roleGrants); ok {
		return scope, nil
	}
	return DataScope{}, nil
}

func DataScopeExtrasContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (DataScopeExtras, error) {
	roleIDs, err := ActiveRoleIDsForUserContext(ctx, db, userID, roleID)
	if err != nil {
		return DataScopeExtras{}, err
	}
	return DataScopeExtrasWithRoleIDsContext(ctx, db, userID, roleIDs)
}

func DataScopeExtrasWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) (DataScopeExtras, error) {
	if err := ctxErr(ctx); err != nil {
		return DataScopeExtras{}, err
	}
	if db == nil || !TablesReady(db) {
		return DataScopeExtras{}, nil
	}
	roleIDs = normalizeRoleIDs(roleIDs...)
	subjects := make([]permissionSubjectRef, 0, len(roleIDs)+1)
	subjects = append(subjects, permissionSubjectRef{subjectType: SubjectUser, subjectID: userID})
	for _, roleID := range roleIDs {
		subjects = append(subjects, permissionSubjectRef{subjectType: SubjectRole, subjectID: roleID})
	}
	grants, err := grantsBySubjectsAndKeys(ctx, db, []string{DataExtraPermissionKey}, subjects...)
	if err != nil {
		return DataScopeExtras{}, err
	}
	return mergeDataScopeExtras(grants), nil
}

func DataScopeBundleContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (DataScope, DataScopeExtras, error) {
	roleIDs, err := ActiveRoleIDsForUserContext(ctx, db, userID, roleID)
	if err != nil {
		return DataScope{}, DataScopeExtras{}, err
	}
	return DataScopeBundleWithRoleIDsContext(ctx, db, userID, roleIDs)
}

func DataScopeBundleWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) (DataScope, DataScopeExtras, error) {
	if err := ctxErr(ctx); err != nil {
		return DataScope{}, DataScopeExtras{}, err
	}
	if db == nil || !TablesReady(db) {
		return DataScope{}, DataScopeExtras{}, nil
	}
	roleIDs = normalizeRoleIDs(roleIDs...)
	keys := []string{DataAllPermissionKey, DataDeptPermissionKey, DataSelfPermissionKey, DataCustomPermissionKey, DataExtraPermissionKey}
	subjects := make([]permissionSubjectRef, 0, len(roleIDs)+1)
	subjects = append(subjects, permissionSubjectRef{subjectType: SubjectUser, subjectID: userID})
	for _, roleID := range roleIDs {
		subjects = append(subjects, permissionSubjectRef{subjectType: SubjectRole, subjectID: roleID})
	}
	grants, err := grantsBySubjectsAndKeys(ctx, db, keys, subjects...)
	if err != nil {
		return DataScope{}, DataScopeExtras{}, err
	}
	baseGrants := make([]model.PermissionGrant, 0, len(grants))
	extraGrants := make([]model.PermissionGrant, 0)
	for _, grant := range grants {
		if grant.PermissionKey == DataExtraPermissionKey {
			extraGrants = append(extraGrants, grant)
			continue
		}
		baseGrants = append(baseGrants, grant)
	}
	userGrants := make([]model.PermissionGrant, 0)
	roleGrants := make([]model.PermissionGrant, 0)
	roleIDSet := uintSet(roleIDs)
	for _, grant := range baseGrants {
		switch {
		case grant.SubjectType == SubjectUser && grant.SubjectID == userID:
			userGrants = append(userGrants, grant)
		case grant.SubjectType == SubjectRole && roleIDSet[grant.SubjectID]:
			roleGrants = append(roleGrants, grant)
		}
	}
	scope := DataScope{}
	if found, ok := mergedDataScope(userGrants); ok {
		scope = found
	} else if found, ok := mergedDataScope(roleGrants); ok {
		scope = found
	}
	return scope, mergeDataScopeExtras(extraGrants), nil
}

func UserDataScopeExtrasContext(ctx context.Context, db *gorm.DB, userID uint) (DataScopeExtras, error) {
	if err := ctxErr(ctx); err != nil {
		return DataScopeExtras{}, err
	}
	if db == nil || !TablesReady(db) || userID == 0 {
		return DataScopeExtras{}, nil
	}
	grants, err := grantsBySubjectAndKeys(ctx, db, SubjectUser, userID, []string{DataExtraPermissionKey})
	if err != nil {
		return DataScopeExtras{}, err
	}
	return mergeDataScopeExtras(grants), nil
}

func SetRoleAdminPermissionKeysTx(tx *gorm.DB, roleID uint, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, dataScope int, deptIDs []uint) error {
	return setRoleAdminPermissionKeysTx(tx, roleID, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs, false)
}

func SetRoleApplicationMenuPermissionsTx(tx *gorm.DB, roleID uint, clientMenuKeys, dingtalkH5MenuKeys []string) error {
	if roleID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	keys := normalizeRoleApplicationMenuKeys(clientMenuKeys, dingtalkH5MenuKeys)
	if err := ensureApplicationPermissionCatalogForKeysTx(tx, keys); err != nil {
		return err
	}
	if err := replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationMenuPermissionPrefixes(), keys, EffectAllow, nil, "form"); err != nil {
		return err
	}
	invalidateSubjectPermissionSetCache()
	invalidateDingTalkH5MenuPermissionCache()
	return nil
}

func SetRoleApplicationAPIPermissionsTx(tx *gorm.DB, roleID uint, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys []string) error {
	if roleID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	keys := normalizeRoleApplicationAPIKeys(clientAPIPermissionKeys, dingtalkH5APIPermissionKeys)
	if err := ensureApplicationPermissionCatalogForKeysTx(tx, keys); err != nil {
		return err
	}
	if err := replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationAPIPermissionPrefixes(), keys, EffectAllow, nil, "form"); err != nil {
		return err
	}
	invalidateSubjectPermissionSetCache()
	return nil
}

func setRoleAdminPermissionKeysTx(tx *gorm.DB, roleID uint, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, dataScope int, deptIDs []uint, ensureCatalog bool) error {
	if roleID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	if ensureCatalog {
		if err := ensureBuiltinPermissions(tx); err != nil {
			return err
		}
		if err := syncAdminAPIPermissions(tx); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(adminPermissionKeys)+len(adminAPIPermissionKeys)+2)
	if allowAdminLogin != 0 {
		keys = append(keys, AdminLoginPermissionKey)
	}
	for _, key := range normalizePermissionKeys(adminPermissionKeys) {
		if strings.HasPrefix(key, "admin:menu:") {
			keys = append(keys, key)
		}
	}
	for _, key := range normalizePermissionKeys(adminAPIPermissionKeys) {
		if strings.HasPrefix(key, "admin:api:") {
			keys = append(keys, key)
		}
	}
	dataKey, scopeValue := dataScopeGrant(dataScope, deptIDs)
	if dataKey != "" {
		keys = append(keys, dataKey)
	}
	if err := replaceSubjectGrantsTx(tx, SubjectRole, roleID, roleManagedPrefixes(), keys, EffectAllow, scopeValueByKey(dataKey, scopeValue), "form"); err != nil {
		return err
	}
	invalidateSubjectPermissionSetCache()
	return nil
}

func SetUserAdminPermissionOverridesTx(tx *gorm.DB, userID uint, allowKeys, denyKeys []string) error {
	if userID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	if err := replaceSubjectGrantsByEffectsTx(tx, SubjectUser, userID, UserPermissionPrefixes(), normalizePermissionKeys(allowKeys), normalizePermissionKeys(denyKeys), "form"); err != nil {
		return err
	}
	invalidateSubjectPermissionSetCache()
	invalidateDingTalkH5MenuPermissionCache()
	return nil
}

func SetUserApplicationMenuPermissionOverridesTx(tx *gorm.DB, userID uint, allowKeys, denyKeys []string) error {
	if userID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return fmt.Errorf("权限表未初始化")
	}
	prefixes := append(ApplicationMenuPermissionPrefixes(), ApplicationAPIPermissionPrefixes()...)
	normalizedAllowKeys, normalizedDenyKeys := normalizeUserApplicationPermissionKeySets(allowKeys, denyKeys)
	applicationKeys := append(append([]string{}, normalizedAllowKeys...), normalizedDenyKeys...)
	if err := ensureApplicationPermissionCatalogForKeysTx(tx, applicationKeys); err != nil {
		return err
	}
	if err := replaceSubjectGrantsByEffectsTx(tx, SubjectUser, userID, prefixes, normalizedAllowKeys, normalizedDenyKeys, "form"); err != nil {
		return err
	}
	invalidateSubjectPermissionSetCache()
	invalidateDingTalkH5MenuPermissionCache()
	return nil
}

func SetUserDataScopeExtrasTx(tx *gorm.DB, userID uint, deptIDs, userIDs []uint) error {
	if userID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return fmt.Errorf("权限表未初始化")
	}
	if err := ensureBuiltinPermissions(tx); err != nil {
		return err
	}
	deptIDs = normalizeUintIDs(deptIDs)
	userIDs = normalizeUintIDs(userIDs)
	scopeValue := ""
	keys := []string{}
	if len(deptIDs) > 0 || len(userIDs) > 0 {
		raw, _ := json.Marshal(map[string][]uint{"deptIds": deptIDs, "userIds": userIDs})
		scopeValue = string(raw)
		keys = append(keys, DataExtraPermissionKey)
	}
	if err := replaceSubjectGrantsTx(tx, SubjectUser, userID, []string{DataExtraPermissionKey}, keys, EffectAllow, scopeValueByKey(DataExtraPermissionKey, scopeValue), "form"); err != nil {
		return err
	}
	invalidateSubjectPermissionSetCache()
	return nil
}

func AdminAPIPermissionKey(perms string) string {
	return adminrouteperm.KeyForPerms(perms)
}

func AdminPermissionPrefixes() []string {
	return []string{AdminLoginPermissionKey, "admin:menu:%", "admin:api:%", "data:%"}
}

func ApplicationMenuPermissionPrefixes() []string {
	return []string{"client:menu:%", "dingtalk_h5:menu:%", "dingtalk_h5:button:%"}
}

func ApplicationAPIPermissionPrefixes() []string {
	return []string{"client:api:%", "dingtalk_h5:api:%"}
}

func ApplicationPermissionPrefixes() []string {
	prefixes := append(ApplicationMenuPermissionPrefixes(), ApplicationAPIPermissionPrefixes()...)
	return prefixes
}

func ensureApplicationPermissionCatalogForKeysTx(tx *gorm.DB, keys []string) error {
	keys = normalizePermissionKeys(keys)
	if len(keys) == 0 {
		return nil
	}
	ctx := txContext(tx)
	needsClientMenu := false
	needsDingTalkH5Menu := false
	needsDingTalkH5Button := false
	needsClientAPI := false
	needsDingTalkH5API := false
	for _, key := range keys {
		switch {
		case strings.HasPrefix(key, "client:menu:"):
			needsClientMenu = true
		case strings.HasPrefix(key, "dingtalk_h5:menu:"):
			needsDingTalkH5Menu = true
		case strings.HasPrefix(key, "dingtalk_h5:button:"):
			needsDingTalkH5Button = true
		case strings.HasPrefix(key, "client:api:"):
			needsClientAPI = true
		case strings.HasPrefix(key, "dingtalk_h5:api:"):
			needsDingTalkH5API = true
		}
	}
	if needsClientMenu {
		if err := ensureMissingApplicationMenuPermissionsContext(ctx, tx, appmenuperm.ClientMenuDeclarations()); err != nil {
			return err
		}
	}
	if needsDingTalkH5Menu {
		if err := ensureMissingApplicationMenuPermissionsContext(ctx, tx, appmenuperm.DingTalkH5MenuDeclarations()); err != nil {
			return err
		}
	}
	if needsDingTalkH5Button {
		if err := ensureMissingApplicationButtonPermissionsContext(ctx, tx, appmenuperm.DingTalkH5ButtonDeclarations()); err != nil {
			return err
		}
	}
	if needsClientAPI {
		if err := ensureMissingApplicationAPIPermissionsContext(ctx, tx, appapiperm.ClientAPICategories(), appapiperm.ClientAPIDeclarations()); err != nil {
			return err
		}
	}
	if needsDingTalkH5API {
		if err := ensureMissingApplicationAPIPermissionsContext(ctx, tx, appapiperm.DingTalkH5APICategories(), appapiperm.DingTalkH5APIDeclarations()); err != nil {
			return err
		}
	}
	return nil
}

func txContext(tx *gorm.DB) context.Context {
	if tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		return tx.Statement.Context
	}
	return context.Background()
}

func UserPermissionPrefixes() []string {
	prefixes := append(AdminPermissionPrefixes(), ApplicationMenuPermissionPrefixes()...)
	return append(prefixes, ApplicationAPIPermissionPrefixes()...)
}

func roleManagedPrefixes() []string {
	return AdminPermissionPrefixes()
}

func ensureBuiltinPermissions(db *gorm.DB) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	items := []model.Permission{
		{Key: AdminLoginPermissionKey, Name: "后台入口", Platform: PlatformAdmin, Type: TypeLogin, Status: 1, Sort: 0, AddTime: now, EditTime: now},
		{Key: DataAllPermissionKey, Name: "全部数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 1, AddTime: now, EditTime: now},
		{Key: DataDeptPermissionKey, Name: "本部门数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 2, AddTime: now, EditTime: now},
		{Key: DataSelfPermissionKey, Name: "本人数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 3, AddTime: now, EditTime: now},
		{Key: DataCustomPermissionKey, Name: "自定义部门数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 4, AddTime: now, EditTime: now},
		{Key: DataExtraPermissionKey, Name: "用户额外数据", Platform: PlatformData, Type: TypeData, Status: 1, Sort: 5, AddTime: now, EditTime: now},
	}
	for _, item := range items {
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func SyncAdminMenuPermissionsContext(ctx context.Context, db *gorm.DB, enableExam bool) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if err := EnsurePermissionSchemaContext(ctx, db); err != nil {
		return err
	}
	return syncAdminMenuPermissions(db.WithContext(ctx), enableExam)
}

func syncAdminMenuPermissions(db *gorm.DB, enableExam bool) error {
	declarations := adminmenuperm.Declarations(enableExam)
	now := database.Now()
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         adminMenuDeclarationType(declaration.Type),
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Icon:         declaration.Icon,
			Perms:        declaration.Perms,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertAdminMenuPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func adminMenuDeclarationType(value string) string {
	switch value {
	case adminmenuperm.TypeDirectory:
		return TypeDirectory
	case adminmenuperm.TypeButton:
		return TypeButton
	default:
		return TypeMenu
	}
}

func firstBool(values []bool, fallback bool) bool {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}

func upsertAdminMenuPermission(db *gorm.DB, item model.Permission) error {
	return upsertPermission(db, item)
}

func syncAdminAPIPermissions(db *gorm.DB) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, category := range adminrouteperm.Categories() {
		item := model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: PlatformAdmin,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
			AddTime:  now,
			EditTime: now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	for index, declaration := range adminrouteperm.Declarations() {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         (index + 1) * 10,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func syncClientMenuPermissions(db *gorm.DB) error {
	return syncApplicationMenuPermissions(db, appmenuperm.ClientMenuDeclarations())
}

func syncDingTalkH5MenuPermissions(db *gorm.DB) error {
	return syncApplicationMenuPermissions(db, appmenuperm.DingTalkH5MenuDeclarations())
}

func syncDingTalkH5ButtonPermissions(db *gorm.DB) error {
	return syncApplicationButtonPermissions(db, appmenuperm.DingTalkH5ButtonDeclarations())
}

func syncClientAPIPermissions(db *gorm.DB) error {
	return syncApplicationAPIPermissions(db, appapiperm.ClientAPICategories(), appapiperm.ClientAPIDeclarations())
}

func syncDingTalkH5APIPermissions(db *gorm.DB) error {
	return syncApplicationAPIPermissions(db, appapiperm.DingTalkH5APICategories(), appapiperm.DingTalkH5APIDeclarations())
}

func syncApplicationMenuPermissions(db *gorm.DB, declarations []appmenuperm.Declaration) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, declaration := range declarations {
		permissionType := declaration.Type
		if permissionType == "" {
			permissionType = TypeMenu
		}
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         permissionType,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Icon:         declaration.Icon,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func syncApplicationButtonPermissions(db *gorm.DB, declarations []appmenuperm.Declaration) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeButton,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func syncApplicationAPIPermissions(db *gorm.DB, categories []appapiperm.Category, declarations []appapiperm.Declaration) error {
	if err := EnsurePermissionSchemaContext(context.Background(), db); err != nil {
		return err
	}
	now := database.Now()
	for _, category := range categories {
		item := model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: category.Platform,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
			AddTime:  now,
			EditTime: now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         declaration.Sort,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func ensureMissingApplicationMenuPermissionsContext(ctx context.Context, db *gorm.DB, declarations []appmenuperm.Declaration) error {
	items := make([]model.Permission, 0, len(declarations))
	for _, declaration := range declarations {
		permissionType := declaration.Type
		if permissionType == "" {
			permissionType = TypeMenu
		}
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         permissionType,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Icon:         declaration.Icon,
			Sort:         declaration.Sort,
			Status:       1,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func ensureMissingApplicationButtonPermissionsContext(ctx context.Context, db *gorm.DB, declarations []appmenuperm.Declaration) error {
	items := make([]model.Permission, 0, len(declarations))
	for _, declaration := range declarations {
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeButton,
			ParentKey:    declaration.ParentKey,
			ResourcePath: declaration.Path,
			Sort:         declaration.Sort,
			Status:       1,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func ensureMissingAdminAPIPermissionsContext(ctx context.Context, db *gorm.DB) error {
	now := database.Now()
	categories := adminrouteperm.Categories()
	declarations := adminrouteperm.Declarations()
	items := make([]model.Permission, 0, len(categories)+len(declarations))
	for _, category := range categories {
		items = append(items, model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: PlatformAdmin,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
			AddTime:  now,
			EditTime: now,
		})
	}
	for index, declaration := range declarations {
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         (index + 1) * 10,
			Status:       1,
			AddTime:      now,
			EditTime:     now,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func ensureMissingApplicationAPIPermissionsContext(ctx context.Context, db *gorm.DB, categories []appapiperm.Category, declarations []appapiperm.Declaration) error {
	items := make([]model.Permission, 0, len(categories)+len(declarations))
	for _, category := range categories {
		items = append(items, model.Permission{
			Key:      category.Key,
			Name:     category.Name,
			Platform: category.Platform,
			Type:     TypeAPICategory,
			Sort:     category.Sort,
			Status:   1,
		})
	}
	for _, declaration := range declarations {
		items = append(items, model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeAPI,
			ParentKey:    declaration.CategoryKey,
			ResourcePath: declaration.Path,
			Perms:        declaration.Perms,
			Sort:         declaration.Sort,
			Status:       1,
		})
	}
	return createMissingPermissionsContext(ctx, db, items)
}

func createMissingPermissionsContext(ctx context.Context, db *gorm.DB, items []model.Permission) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	keys := make([]string, 0, len(items))
	byKey := make(map[string]model.Permission, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		item.Key = key
		keys = append(keys, key)
		byKey[key] = item
	}
	keys = normalizePermissionKeys(keys)
	if len(keys) == 0 {
		return nil
	}
	var existingKeys []string
	if err := db.Model(&model.Permission{}).Where("`permission_key` IN ?", keys).Pluck("permission_key", &existingKeys).Error; err != nil {
		return err
	}
	existing := make(map[string]bool, len(existingKeys))
	for _, key := range existingKeys {
		existing[key] = true
	}
	now := database.Now()
	createItems := make([]model.Permission, 0)
	for _, key := range keys {
		if existing[key] {
			continue
		}
		item := byKey[key]
		item.AddTime = now
		item.EditTime = now
		createItems = append(createItems, item)
	}
	if len(createItems) == 0 {
		return ctxErr(ctx)
	}
	if err := db.CreateInBatches(createItems, 100).Error; err != nil {
		return err
	}
	return ctxErr(ctx)
}

func upsertPermission(db *gorm.DB, item model.Permission) error {
	var current model.Permission
	err := db.Where("`permission_key` = ?", item.Key).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&item).Error
	}
	if err != nil {
		return err
	}
	icon := strings.TrimSpace(item.Icon)
	if currentIcon := strings.TrimSpace(current.Icon); currentIcon != "" {
		icon = currentIcon
	}
	return db.Model(&current).Updates(map[string]interface{}{
		"permission_name":          item.Name,
		"permission_platform":      item.Platform,
		"permission_type":          item.Type,
		"permission_parent_key":    item.ParentKey,
		"permission_resource_id":   item.ResourceID,
		"permission_resource_path": item.ResourcePath,
		"permission_icon":          icon,
		"permission_perms":         item.Perms,
		"permission_sort":          item.Sort,
		"permission_status":        item.Status,
		"permission_edit_time":     database.Now(),
	}).Error
}

func replaceSubjectGrantsTx(tx *gorm.DB, subjectType string, subjectID uint, likePrefixes []string, allowKeys []string, effect string, scopeValues map[string]string, source string) error {
	if err := deleteSubjectGrantsByPrefixesTx(tx, subjectType, subjectID, likePrefixes); err != nil {
		return err
	}
	grants, err := buildSubjectGrantsTx(tx, subjectType, subjectID, normalizePermissionKeys(allowKeys), effect, scopeValues, source)
	if err != nil {
		return err
	}
	if len(grants) == 0 {
		return nil
	}
	return createGrantBatchTx(tx, grants)
}

func replaceSubjectGrantsByEffectsTx(tx *gorm.DB, subjectType string, subjectID uint, likePrefixes []string, allowKeys, denyKeys []string, source string) error {
	if err := deleteSubjectGrantsByPrefixesTx(tx, subjectType, subjectID, likePrefixes); err != nil {
		return err
	}
	grants, err := buildSubjectGrantsByEffectsTx(tx, subjectType, subjectID, allowKeys, denyKeys, source)
	if err != nil {
		return err
	}
	if len(grants) == 0 {
		return nil
	}
	return createGrantBatchTx(tx, grants)
}

func createGrantBatchTx(tx *gorm.DB, grants []model.PermissionGrant) error {
	if len(grants) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "grant_subject_type"},
			{Name: "grant_subject_id"},
			{Name: "grant_permission_key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"grant_permission_id",
			"grant_effect",
			"grant_scope_value",
			"grant_source",
			"grant_status",
			"grant_edit_time",
			"updated_at",
		}),
	}).CreateInBatches(grants, 100).Error
}

func deleteSubjectGrantsByPrefixesTx(tx *gorm.DB, subjectType string, subjectID uint, likePrefixes []string) error {
	query := tx.Where("`grant_subject_type` = ? AND `grant_subject_id` = ?", subjectType, subjectID)
	if len(likePrefixes) > 0 {
		clauses := make([]string, 0, len(likePrefixes))
		args := make([]interface{}, 0, len(likePrefixes))
		for range likePrefixes {
			clauses = append(clauses, "`grant_permission_key` LIKE ?")
		}
		for _, prefix := range likePrefixes {
			args = append(args, prefix)
		}
		query = query.Where(strings.Join(clauses, " OR "), args...)
	}
	return query.Delete(&model.PermissionGrant{}).Error
}

func createGrantTx(tx *gorm.DB, subjectType string, subjectID uint, key, effect, scopeValue, source string) error {
	grants, err := buildSubjectGrantsTx(tx, subjectType, subjectID, []string{key}, effect, map[string]string{key: scopeValue}, source)
	if err != nil {
		return err
	}
	if len(grants) == 0 {
		return nil
	}
	return tx.Create(&grants[0]).Error
}

func buildSubjectGrantsByEffectsTx(tx *gorm.DB, subjectType string, subjectID uint, allowKeys, denyKeys []string, source string) ([]model.PermissionGrant, error) {
	keys := normalizePermissionKeys(append(append([]string{}, allowKeys...), denyKeys...))
	if len(keys) == 0 {
		return nil, nil
	}
	var perms []model.Permission
	if err := tx.Select("`id`, `permission_key`").Where("`permission_key` IN ?", keys).Find(&perms).Error; err != nil {
		return nil, err
	}
	permissionByKey := make(map[string]model.Permission, len(perms))
	for _, perm := range perms {
		permissionByKey[perm.Key] = perm
	}
	now := database.Now()
	allowSet := make(map[string]struct{}, len(allowKeys))
	grants := make([]model.PermissionGrant, 0, len(keys))
	for _, key := range normalizePermissionKeys(allowKeys) {
		allowSet[key] = struct{}{}
		grant, ok, err := buildSubjectGrantFromPermission(subjectType, subjectID, key, EffectAllow, "", source, now, permissionByKey)
		if err != nil {
			return nil, err
		}
		if ok {
			grants = append(grants, grant)
		}
	}
	for _, key := range normalizePermissionKeys(denyKeys) {
		if _, ok := allowSet[key]; ok {
			continue
		}
		grant, ok, err := buildSubjectGrantFromPermission(subjectType, subjectID, key, EffectDeny, "", source, now, permissionByKey)
		if err != nil {
			return nil, err
		}
		if ok {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func buildSubjectGrantsTx(tx *gorm.DB, subjectType string, subjectID uint, keys []string, effect string, scopeValues map[string]string, source string) ([]model.PermissionGrant, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	var perms []model.Permission
	if err := tx.Where("`permission_key` IN ?", keys).Find(&perms).Error; err != nil {
		return nil, err
	}
	permissionByKey := make(map[string]model.Permission, len(perms))
	for _, perm := range perms {
		permissionByKey[perm.Key] = perm
	}
	now := database.Now()
	grants := make([]model.PermissionGrant, 0, len(keys))
	for _, key := range keys {
		grant, ok, err := buildSubjectGrantFromPermission(subjectType, subjectID, key, effect, scopeValues[key], source, now, permissionByKey)
		if err != nil {
			return nil, err
		}
		if ok {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func buildSubjectGrantFromPermission(subjectType string, subjectID uint, key, effect, scopeValue, source string, now int64, permissionByKey map[string]model.Permission) (model.PermissionGrant, bool, error) {
	perm, ok := permissionByKey[key]
	if !ok {
		if strings.HasPrefix(key, "admin:menu:") {
			return model.PermissionGrant{}, false, nil
		}
		return model.PermissionGrant{}, false, fmt.Errorf("%w: %s", gorm.ErrRecordNotFound, key)
	}
	return model.PermissionGrant{
		SubjectType:   subjectType,
		SubjectID:     subjectID,
		PermissionKey: key,
		PermissionID:  perm.ID,
		Effect:        effect,
		ScopeValue:    scopeValue,
		Source:        source,
		Status:        1,
		AddTime:       now,
		EditTime:      now,
	}, true, nil
}

func dataScopeGrant(dataScope int, deptIDs []uint) (string, string) {
	switch dataScope {
	case 2:
		return DataDeptPermissionKey, ""
	case 3:
		return DataSelfPermissionKey, ""
	case 4:
		raw, _ := json.Marshal(map[string][]uint{"deptIds": deptIDs})
		return DataCustomPermissionKey, string(raw)
	default:
		return DataAllPermissionKey, ""
	}
}

func scopeValueByKey(key, value string) map[string]string {
	if key == "" {
		return nil
	}
	return map[string]string{key: value}
}

func normalizeRoleApplicationMenuKeys(clientMenuKeys, dingtalkH5MenuKeys []string) []string {
	selected := map[string]bool{}
	for _, key := range normalizePermissionKeys(clientMenuKeys) {
		if strings.HasPrefix(key, "client:menu:") {
			selected[key] = true
		}
	}
	for _, key := range normalizePermissionKeys(dingtalkH5MenuKeys) {
		if strings.HasPrefix(key, "dingtalk_h5:menu:") {
			selected[key] = true
		}
		if strings.HasPrefix(key, "dingtalk_h5:button:") {
			selected[key] = true
		}
		if strings.HasPrefix(key, "dingtalk_h5:menu:performance:") {
			selected["dingtalk_h5:menu:performance"] = true
		}
	}
	keys := orderedApplicationMenuKeys(selected, appmenuperm.ClientMenuDeclarations())
	keys = append(keys, orderedApplicationMenuKeys(selected, appmenuperm.DingTalkH5PermissionDeclarations())...)
	return keys
}

func orderedApplicationMenuKeys(selected map[string]bool, declarations []appmenuperm.Declaration) []string {
	keys := make([]string, 0, len(declarations))
	for _, declaration := range declarations {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func normalizeRoleApplicationAPIKeys(clientAPIPermissionKeys, dingtalkH5APIPermissionKeys []string) []string {
	selected := map[string]bool{}
	for _, key := range normalizePermissionKeys(clientAPIPermissionKeys) {
		if strings.HasPrefix(key, "client:api:") {
			selected[key] = true
		}
	}
	for _, key := range normalizePermissionKeys(dingtalkH5APIPermissionKeys) {
		if strings.HasPrefix(key, "dingtalk_h5:api:") {
			selected[key] = true
		}
	}
	keys := orderedApplicationAPIKeys(selected, appapiperm.ClientAPIDeclarations())
	keys = append(keys, orderedApplicationAPIKeys(selected, appapiperm.DingTalkH5APIDeclarations())...)
	return keys
}

func normalizeUserApplicationPermissionKeys(keys []string) []string {
	clientMenuKeys := make([]string, 0)
	dingtalkH5MenuKeys := make([]string, 0)
	clientAPIKeys := make([]string, 0)
	dingtalkH5APIKeys := make([]string, 0)
	for _, key := range normalizePermissionKeys(keys) {
		switch {
		case strings.HasPrefix(key, "client:menu:"):
			clientMenuKeys = append(clientMenuKeys, key)
		case strings.HasPrefix(key, "dingtalk_h5:menu:"), strings.HasPrefix(key, "dingtalk_h5:button:"):
			dingtalkH5MenuKeys = append(dingtalkH5MenuKeys, key)
		case strings.HasPrefix(key, "client:api:"):
			clientAPIKeys = append(clientAPIKeys, key)
		case strings.HasPrefix(key, "dingtalk_h5:api:"):
			dingtalkH5APIKeys = append(dingtalkH5APIKeys, key)
		}
	}
	result := normalizeRoleApplicationMenuKeys(clientMenuKeys, dingtalkH5MenuKeys)
	result = append(result, normalizeRoleApplicationAPIKeys(clientAPIKeys, dingtalkH5APIKeys)...)
	return result
}

func normalizeUserApplicationPermissionKeySets(allowKeys, denyKeys []string) ([]string, []string) {
	allow := normalizeUserApplicationPermissionKeys(allowKeys)
	deny := normalizeUserApplicationPermissionKeys(denyKeys)
	allowed := make(map[string]bool, len(allow))
	for _, key := range allow {
		allowed[key] = true
	}
	filteredDeny := make([]string, 0, len(deny))
	for _, key := range deny {
		if !allowed[key] {
			filteredDeny = append(filteredDeny, key)
		}
	}
	return allow, filteredDeny
}

func orderedApplicationAPIKeys(selected map[string]bool, declarations []appapiperm.Declaration) []string {
	keys := make([]string, 0, len(declarations))
	for _, declaration := range declarations {
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return keys
}

func normalizePermissionKeys(keys []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}

func grantEffect(ctx context.Context, db *gorm.DB, subjectType string, subjectID uint, key string) (string, bool, error) {
	if subjectID == 0 {
		return "", false, nil
	}
	if err := ctxErr(ctx); err != nil {
		return "", false, err
	}
	var grant model.PermissionGrant
	err := db.Where("`grant_subject_type` = ? AND `grant_subject_id` = ? AND `grant_permission_key` = ? AND `grant_status` = 1", subjectType, subjectID, key).
		Order("`id` DESC").First(&grant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return grant.Effect, true, nil
}

func grantsBySubjectAndKeys(ctx context.Context, db *gorm.DB, subjectType string, subjectID uint, keys []string) ([]model.PermissionGrant, error) {
	if subjectID == 0 {
		return nil, nil
	}
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	var grants []model.PermissionGrant
	err := db.Where("`grant_subject_type` = ? AND `grant_subject_id` = ? AND `grant_permission_key` IN ? AND `grant_status` = 1", subjectType, subjectID, keys).
		Order("`id` DESC").Find(&grants).Error
	return grants, err
}

func grantsBySubjectsAndKeys(ctx context.Context, db *gorm.DB, keys []string, subjects ...permissionSubjectRef) ([]model.PermissionGrant, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	clauses := make([]string, 0, len(subjects))
	args := make([]interface{}, 0, len(subjects)*2)
	for _, subject := range subjects {
		if subject.subjectID == 0 || strings.TrimSpace(subject.subjectType) == "" {
			continue
		}
		clauses = append(clauses, "(`grant_subject_type` = ? AND `grant_subject_id` = ?)")
		args = append(args, subject.subjectType, subject.subjectID)
	}
	if len(clauses) == 0 {
		return nil, nil
	}
	var grants []model.PermissionGrant
	err := db.Where("`grant_permission_key` IN ? AND `grant_status` = 1", keys).
		Where(strings.Join(clauses, " OR "), args...).
		Order("`id` DESC").
		Find(&grants).Error
	return grants, err
}

func firstDataScope(grants []model.PermissionGrant) (DataScope, bool) {
	return mergedDataScope(grants)
}

func mergedDataScope(grants []model.PermissionGrant) (DataScope, bool) {
	hasSelf := false
	hasCustom := false
	customDeptIDs := []uint{}
	for _, grant := range grants {
		if grant.Effect == EffectDeny {
			continue
		}
		switch grant.PermissionKey {
		case DataAllPermissionKey:
			return DataScope{Mode: 1, Ready: true}, true
		case DataDeptPermissionKey:
			return DataScope{Mode: 2, Ready: true}, true
		case DataSelfPermissionKey:
			hasSelf = true
		case DataCustomPermissionKey:
			hasCustom = true
			customDeptIDs = append(customDeptIDs, decodeDeptScope(grant.ScopeValue)...)
		}
	}
	if hasCustom {
		return DataScope{Mode: 4, DeptIDs: normalizeUintIDs(customDeptIDs), Ready: true}, true
	}
	if hasSelf {
		return DataScope{Mode: 3, Ready: true}, true
	}
	return DataScope{}, false
}

func mergeDataScopeExtras(grants []model.PermissionGrant) DataScopeExtras {
	deptIDs := []uint{}
	userIDs := []uint{}
	ready := false
	for _, grant := range grants {
		if grant.Effect == EffectDeny || grant.PermissionKey != DataExtraPermissionKey {
			continue
		}
		extras := decodeDataScopeExtras(grant.ScopeValue)
		if len(extras.DeptIDs) == 0 && len(extras.UserIDs) == 0 {
			continue
		}
		ready = true
		deptIDs = append(deptIDs, extras.DeptIDs...)
		userIDs = append(userIDs, extras.UserIDs...)
	}
	return DataScopeExtras{
		DeptIDs: normalizeUintIDs(deptIDs),
		UserIDs: normalizeUintIDs(userIDs),
		Ready:   ready,
	}
}

func decodeDeptScope(raw string) []uint {
	var payload struct {
		DeptIDs []uint `json:"deptIds"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil
	}
	return payload.DeptIDs
}

func decodeDataScopeExtras(raw string) DataScopeExtras {
	var payload struct {
		DeptIDs []uint `json:"deptIds"`
		UserIDs []uint `json:"userIds"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return DataScopeExtras{}
	}
	return DataScopeExtras{
		DeptIDs: normalizeUintIDs(payload.DeptIDs),
		UserIDs: normalizeUintIDs(payload.UserIDs),
	}
}

func normalizeUintIDs(values []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeRoleIDs(values ...uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func uintSet(values []uint) map[uint]bool {
	result := make(map[uint]bool, len(values))
	for _, value := range values {
		if value > 0 {
			result[value] = true
		}
	}
	return result
}

func intersectsUintSet(values []uint, set map[uint]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

func subjectAdminPermissionSets(ctx context.Context, db *gorm.DB, userID, roleID uint) (map[string]bool, map[string]bool, error) {
	return subjectPermissionSetsByPrefixes(ctx, db, userID, roleID, []string{"admin:menu:%", "admin:api:%"})
}

func subjectPermissionSetsByPrefixes(ctx context.Context, db *gorm.DB, userID, roleID uint, prefixes []string) (map[string]bool, map[string]bool, error) {
	return subjectPermissionSetsByRoleIDsAndPrefixes(ctx, db, userID, []uint{roleID}, prefixes)
}

func subjectPermissionSetsByRoleIDsAndPrefixes(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint, prefixes []string) (map[string]bool, map[string]bool, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, nil, err
	}
	roleIDs = normalizeRoleIDs(roleIDs...)
	if cachedAllowed, cachedDenied, ok := getSubjectPermissionSetCache(userID, roleIDs, prefixes); ok {
		return cachedAllowed, cachedDenied, nil
	}
	allowed := map[string]bool{}
	denied := map[string]bool{}
	if len(roleIDs) > 0 {
		var grants []model.PermissionGrant
		query := db.Select(permissionGrantKeySelectColumns).
			Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_status` = 1", SubjectRole, roleIDs)
		if len(prefixes) > 0 {
			where, args := likeAnyClause("`grant_permission_key`", prefixes)
			query = query.Where(where, args...)
		}
		if err := query.Find(&grants).Error; err != nil {
			return nil, nil, err
		}
		for _, grant := range grants {
			if grant.Effect == EffectDeny {
				denied[grant.PermissionKey] = true
			} else if grant.Effect == EffectAllow {
				allowed[grant.PermissionKey] = true
			}
		}
	}
	if userID > 0 {
		var grants []model.PermissionGrant
		query := db.Select(permissionGrantKeySelectColumns).
			Where("`grant_subject_type` = ? AND `grant_subject_id` = ? AND `grant_status` = 1", SubjectUser, userID)
		if len(prefixes) > 0 {
			where, args := likeAnyClause("`grant_permission_key`", prefixes)
			query = query.Where(where, args...)
		}
		if err := query.Find(&grants).Error; err != nil {
			return nil, nil, err
		}
		for _, grant := range grants {
			if grant.Effect == EffectDeny {
				denied[grant.PermissionKey] = true
			} else {
				allowed[grant.PermissionKey] = true
			}
		}
	}
	setSubjectPermissionSetCache(userID, roleIDs, prefixes, allowed, denied)
	return allowed, denied, ctxErr(ctx)
}

func likeAnyClause(column string, prefixes []string) (string, []interface{}) {
	clauses := make([]string, 0, len(prefixes))
	args := make([]interface{}, 0, len(prefixes))
	for _, prefix := range prefixes {
		clauses = append(clauses, column+" LIKE ?")
		args = append(args, prefix)
	}
	return strings.Join(clauses, " OR "), args
}

func ResetPermissionTablesReadyCache() {
	permissionTablesReadyCache.Lock()
	permissionTablesReadyCache.checked = false
	permissionTablesReadyCache.ready = false
	permissionTablesReadyCache.schemaReady = false
	permissionTablesReadyCache.checkedAt = time.Time{}
	permissionTablesReadyCache.Unlock()
	userRolesTableReadyCache.Lock()
	userRolesTableReadyCache.checked = false
	userRolesTableReadyCache.ready = false
	userRolesTableReadyCache.checkedAt = time.Time{}
	userRolesTableReadyCache.Unlock()
}

func markPermissionTablesReady(ready bool) {
	permissionTablesReadyCache.Lock()
	permissionTablesReadyCache.checked = true
	permissionTablesReadyCache.ready = ready
	if !ready {
		permissionTablesReadyCache.schemaReady = false
	}
	permissionTablesReadyCache.checkedAt = time.Now()
	permissionTablesReadyCache.Unlock()
}

func permissionTablesReadyCached() bool {
	permissionTablesReadyCache.RLock()
	ready := permissionTablesReadyCache.checked && permissionTablesReadyCache.ready && permissionTablesReadyCache.schemaReady
	permissionTablesReadyCache.RUnlock()
	return ready
}

func markPermissionSchemaReady(ready bool) {
	permissionTablesReadyCache.Lock()
	permissionTablesReadyCache.checked = true
	permissionTablesReadyCache.ready = ready
	permissionTablesReadyCache.schemaReady = ready
	permissionTablesReadyCache.checkedAt = time.Now()
	permissionTablesReadyCache.Unlock()
}

func TablesReady(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	permissionTablesReadyCache.RLock()
	if permissionTablesReadyCache.checked {
		ready := permissionTablesReadyCache.ready
		checkedAt := permissionTablesReadyCache.checkedAt
		permissionTablesReadyCache.RUnlock()
		if ready || time.Since(checkedAt) < permissionTablesReadyNegativeCacheTTL {
			return ready
		}
	} else {
		permissionTablesReadyCache.RUnlock()
	}

	permissionTablesReadyCache.Lock()
	defer permissionTablesReadyCache.Unlock()
	if permissionTablesReadyCache.checked {
		if permissionTablesReadyCache.ready || time.Since(permissionTablesReadyCache.checkedAt) < permissionTablesReadyNegativeCacheTTL {
			return permissionTablesReadyCache.ready
		}
	}
	permissionTablesReadyCache.checked = true
	permissionTablesReadyCache.ready = db.Migrator().HasTable(&model.Permission{}) && db.Migrator().HasTable(&model.PermissionGrant{})
	permissionTablesReadyCache.checkedAt = time.Now()
	return permissionTablesReadyCache.ready
}

func UserRolesTableReady(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	userRolesTableReadyCache.RLock()
	if userRolesTableReadyCache.checked {
		ready := userRolesTableReadyCache.ready
		checkedAt := userRolesTableReadyCache.checkedAt
		userRolesTableReadyCache.RUnlock()
		if ready || time.Since(checkedAt) < userRolesTableReadyNegativeCacheTTL {
			return ready
		}
	} else {
		userRolesTableReadyCache.RUnlock()
	}

	userRolesTableReadyCache.Lock()
	defer userRolesTableReadyCache.Unlock()
	if userRolesTableReadyCache.checked {
		if userRolesTableReadyCache.ready || time.Since(userRolesTableReadyCache.checkedAt) < userRolesTableReadyNegativeCacheTTL {
			return userRolesTableReadyCache.ready
		}
	}
	userRolesTableReadyCache.checked = true
	userRolesTableReadyCache.ready = db.Migrator().HasTable(&model.UserRole{})
	userRolesTableReadyCache.checkedAt = time.Now()
	return userRolesTableReadyCache.ready
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
