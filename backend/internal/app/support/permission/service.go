package permission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/app/support/adminmenuperm"
	"wecheckin-backend/backend/internal/app/support/adminrouteperm"
	"wecheckin-backend/backend/internal/app/support/appapiperm"
	"wecheckin-backend/backend/internal/app/support/appmenuperm"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
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
)

type DataScope struct {
	Mode    int
	DeptIDs []uint
	Ready   bool
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
	if err := syncClientAPIPermissions(db); err != nil {
		return err
	}
	if err := syncDingTalkH5APIPermissions(db); err != nil {
		return err
	}
	if err := syncLegacyRoleGrants(db); err != nil {
		return err
	}
	return ctxErr(ctx)
}

func EnsurePermissionSchemaContext(ctx context.Context, db *gorm.DB) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("数据库连接异常")
	}
	if !db.Migrator().HasTable(&model.Permission{}) {
		return ctxErr(ctx)
	}
	if !db.Migrator().HasColumn(&model.Permission{}, "Icon") {
		if err := db.Migrator().AddColumn(&model.Permission{}, "Icon"); err != nil {
			return err
		}
	}
	return ctxErr(ctx)
}

func SubjectHasPermissionContext(ctx context.Context, db *gorm.DB, userID, roleID uint, key string) (bool, error) {
	if err := ctxErr(ctx); err != nil {
		return false, err
	}
	if db == nil {
		return false, fmt.Errorf("数据库连接异常")
	}
	if !TablesReady(db) {
		if key == AdminLoginPermissionKey && roleID > 0 {
			return false, nil
		}
		return false, nil
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
	if roleID == 0 {
		return false, nil
	}
	effect, ok, err := grantEffect(ctx, db, SubjectRole, roleID, key)
	if err != nil || !ok {
		return false, err
	}
	return effect == EffectAllow, nil
}

func RoleHasPermissionContext(ctx context.Context, db *gorm.DB, roleID uint, key string) (bool, error) {
	return SubjectHasPermissionContext(ctx, db, 0, roleID, key)
}

func SubjectPermissionEffectContext(ctx context.Context, db *gorm.DB, subjectType string, subjectID uint, key string) (string, bool, error) {
	return grantEffect(ctx, db, subjectType, subjectID, key)
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
	if err := db.Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, EffectAllow).
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
	if err := db.Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_permission_key` LIKE ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, like, EffectAllow).
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
	if err := db.Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_permission_key` = ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, DataCustomPermissionKey, EffectAllow).Find(&grants).Error; err != nil {
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

func ClientMenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return SubjectMenuPermissionKeysContext(ctx, db, userID, roleID, PlatformClient)
}

func DingTalkH5MenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
	return SubjectMenuPermissionKeysContext(ctx, db, userID, roleID, PlatformDingTalkH5)
}

func SubjectMenuPermissionKeysContext(ctx context.Context, db *gorm.DB, userID, roleID uint, platform string) ([]string, bool, error) {
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
	allowed, denied, err := subjectPermissionSetsByPrefixes(ctx, db, userID, roleID, []string{prefix})
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
	query := db.Model(&model.PermissionGrant{}).
		Where("`grant_permission_key` LIKE ? AND `grant_effect` = ? AND `grant_status` = 1", prefix, EffectAllow)
	switch {
	case userID > 0 && roleID > 0:
		query = query.Where("(`grant_subject_type` = ? AND `grant_subject_id` = ?) OR (`grant_subject_type` = ? AND `grant_subject_id` = ?)", SubjectUser, userID, SubjectRole, roleID)
	case userID > 0:
		query = query.Where("`grant_subject_type` = ? AND `grant_subject_id` = ?", SubjectUser, userID)
	case roleID > 0:
		query = query.Where("`grant_subject_type` = ? AND `grant_subject_id` = ?", SubjectRole, roleID)
	default:
		return false, nil
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, ctxErr(ctx)
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
	if err := db.Where("`grant_subject_type` = ? AND `grant_subject_id` IN ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, roleIDs, EffectAllow).
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
		if strings.HasPrefix(grant.PermissionKey, "dingtalk_h5:menu:") {
			dingTalkH5Sets[grant.SubjectID][grant.PermissionKey] = true
		}
	}
	for _, roleID := range roleIDs {
		clientKeysByRole[roleID] = orderedApplicationMenuKeys(clientSets[roleID], appmenuperm.ClientMenuDeclarations())
		dingTalkH5KeysByRole[roleID] = orderedApplicationMenuKeys(dingTalkH5Sets[roleID], appmenuperm.DingTalkH5MenuDeclarations())
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
	if err := db.Where("`grant_subject_type` = ? AND `grant_permission_key` = ? AND `grant_effect` = ? AND `grant_status` = 1", SubjectRole, DataCustomPermissionKey, EffectAllow).Find(&grants).Error; err != nil {
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

func AdminPermCodesContext(ctx context.Context, db *gorm.DB, userID, roleID uint) ([]string, bool, error) {
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
	if err := db.Where("`permission_key` IN ? AND `permission_status` = 1 AND `permission_perms` <> ''", keys).Find(&rows).Error; err != nil {
		return nil, true, err
	}
	seen := map[string]bool{}
	var result []string
	for _, row := range rows {
		for _, p := range strings.Split(row.Perms, ",") {
			p = strings.TrimSpace(p)
			if p != "" && !seen[p] {
				seen[p] = true
				result = append(result, p)
			}
		}
	}
	return result, true, nil
}

func DataScopeContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (DataScope, error) {
	if err := ctxErr(ctx); err != nil {
		return DataScope{}, err
	}
	if db == nil || !TablesReady(db) {
		return DataScope{}, nil
	}
	keys := []string{DataAllPermissionKey, DataDeptPermissionKey, DataSelfPermissionKey, DataCustomPermissionKey}
	userGrants, err := grantsBySubjectAndKeys(ctx, db, SubjectUser, userID, keys)
	if err != nil {
		return DataScope{}, err
	}
	roleGrants, err := grantsBySubjectAndKeys(ctx, db, SubjectRole, roleID, keys)
	if err != nil {
		return DataScope{}, err
	}
	if scope, ok := firstDataScope(userGrants); ok {
		return scope, nil
	}
	if scope, ok := firstDataScope(roleGrants); ok {
		return scope, nil
	}
	return DataScope{}, nil
}

func SetRoleAdminPermissionKeysTx(tx *gorm.DB, roleID uint, allowAdminLogin int, adminPermissionKeys, adminAPIPermissionKeys []string, dataScope int, deptIDs []uint) error {
	return setRoleAdminPermissionKeysTx(tx, roleID, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, dataScope, deptIDs, true)
}

func SetRoleApplicationMenuPermissionsTx(tx *gorm.DB, roleID uint, clientMenuKeys, dingtalkH5MenuKeys []string) error {
	if roleID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	if err := syncClientMenuPermissions(tx); err != nil {
		return err
	}
	if err := syncDingTalkH5MenuPermissions(tx); err != nil {
		return err
	}
	keys := normalizeRoleApplicationMenuKeys(clientMenuKeys, dingtalkH5MenuKeys)
	return replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationMenuPermissionPrefixes(), keys, EffectAllow, nil, "form")
}

func SetRoleApplicationAPIPermissionsTx(tx *gorm.DB, roleID uint, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys []string) error {
	if roleID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	if err := syncClientAPIPermissions(tx); err != nil {
		return err
	}
	if err := syncDingTalkH5APIPermissions(tx); err != nil {
		return err
	}
	keys := normalizeRoleApplicationAPIKeys(clientAPIPermissionKeys, dingtalkH5APIPermissionKeys)
	return replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationAPIPermissionPrefixes(), keys, EffectAllow, nil, "form")
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
	return nil
}

func SetUserAdminPermissionOverridesTx(tx *gorm.DB, userID uint, allowKeys, denyKeys []string) error {
	if userID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	if err := ensureBuiltinPermissions(tx); err != nil {
		return err
	}
	if err := syncAdminAPIPermissions(tx); err != nil {
		return err
	}
	if err := syncClientMenuPermissions(tx); err != nil {
		return err
	}
	if err := syncDingTalkH5MenuPermissions(tx); err != nil {
		return err
	}
	if err := deleteSubjectGrantsByPrefixesTx(tx, SubjectUser, userID, UserPermissionPrefixes()); err != nil {
		return err
	}
	for _, key := range normalizePermissionKeys(allowKeys) {
		if err := createGrantTx(tx, SubjectUser, userID, key, EffectAllow, "", "form"); err != nil {
			return err
		}
	}
	for _, key := range normalizePermissionKeys(denyKeys) {
		if err := createGrantTx(tx, SubjectUser, userID, key, EffectDeny, "", "form"); err != nil {
			return err
		}
	}
	return nil
}

func SetUserApplicationMenuPermissionOverridesTx(tx *gorm.DB, userID uint, allowKeys, denyKeys []string) error {
	if userID == 0 {
		return nil
	}
	if !TablesReady(tx) {
		return nil
	}
	if err := syncClientMenuPermissions(tx); err != nil {
		return err
	}
	if err := syncDingTalkH5MenuPermissions(tx); err != nil {
		return err
	}
	if err := syncClientAPIPermissions(tx); err != nil {
		return err
	}
	if err := syncDingTalkH5APIPermissions(tx); err != nil {
		return err
	}
	prefixes := append(ApplicationMenuPermissionPrefixes(), ApplicationAPIPermissionPrefixes()...)
	if err := deleteSubjectGrantsByPrefixesTx(tx, SubjectUser, userID, prefixes); err != nil {
		return err
	}
	for _, key := range normalizePermissionKeys(allowKeys) {
		if err := createGrantTx(tx, SubjectUser, userID, key, EffectAllow, "", "form"); err != nil {
			return err
		}
	}
	for _, key := range normalizePermissionKeys(denyKeys) {
		if err := createGrantTx(tx, SubjectUser, userID, key, EffectDeny, "", "form"); err != nil {
			return err
		}
	}
	return nil
}

func MenuPermissionKey(menuID uint) string {
	return "admin:menu:" + strconv.FormatUint(uint64(menuID), 10)
}

func AdminAPIPermissionKey(perms string) string {
	return adminrouteperm.KeyForPerms(perms)
}

func AdminPermissionPrefixes() []string {
	return []string{"admin:menu:%", "admin:api:%", "data:%"}
}

func ApplicationMenuPermissionPrefixes() []string {
	return []string{"client:menu:%", "dingtalk_h5:menu:%"}
}

func ApplicationAPIPermissionPrefixes() []string {
	return []string{"client:api:%", "dingtalk_h5:api:%"}
}

func ApplicationPermissionPrefixes() []string {
	prefixes := append(ApplicationMenuPermissionPrefixes(), ApplicationAPIPermissionPrefixes()...)
	return prefixes
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
	if err := syncLegacyMenuTablePermissions(db); err != nil {
		return err
	}
	declarations := adminmenuperm.Declarations(enableExam)
	keyMap, err := resolveAdminMenuDeclarationKeys(db, declarations)
	if err != nil {
		return err
	}
	now := database.Now()
	for _, declaration := range declarations {
		item := model.Permission{
			Key:          keyMap[declaration.Key],
			Name:         declaration.Name,
			Platform:     PlatformAdmin,
			Type:         adminMenuDeclarationType(declaration.Type),
			ParentKey:    keyMap[declaration.ParentKey],
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

type legacyAdminMenuRow struct {
	ID       uint   `gorm:"column:id"`
	Name     string `gorm:"column:menu_name"`
	ParentID uint   `gorm:"column:menu_parent_id"`
	Path     string `gorm:"column:menu_path"`
	Perms    string `gorm:"column:menu_perms"`
	Icon     string `gorm:"column:menu_icon"`
	Sort     int    `gorm:"column:menu_sort"`
	Status   int    `gorm:"column:menu_status"`
	Type     int    `gorm:"column:menu_type"`
	AddTime  int64  `gorm:"column:menu_add_time"`
	EditTime int64  `gorm:"column:menu_edit_time"`
}

func syncLegacyMenuTablePermissions(db *gorm.DB) error {
	if !db.Migrator().HasTable("menus") {
		return nil
	}
	var rows []legacyAdminMenuRow
	if err := db.Table("menus").Find(&rows).Error; err != nil {
		return err
	}
	parentKeyByID := map[uint]string{}
	for _, row := range rows {
		parentKeyByID[row.ID] = MenuPermissionKey(row.ID)
	}
	now := database.Now()
	for _, row := range rows {
		item := model.Permission{
			Key:          MenuPermissionKey(row.ID),
			Name:         row.Name,
			Platform:     PlatformAdmin,
			Type:         legacyMenuType(row.Type),
			ParentKey:    parentKeyByID[row.ParentID],
			ResourceID:   row.ID,
			ResourcePath: row.Path,
			Icon:         row.Icon,
			Perms:        row.Perms,
			Sort:         row.Sort,
			Status:       row.Status,
			AddTime:      valueOrNow(row.AddTime, now),
			EditTime:     valueOrNow(row.EditTime, now),
		}
		if err := upsertPermission(db, item); err != nil {
			return err
		}
	}
	return nil
}

func legacyMenuType(value int) string {
	switch value {
	case 0:
		return TypeDirectory
	case 2:
		return TypeButton
	default:
		return TypeMenu
	}
}

func valueOrNow(value, now int64) int64 {
	if value > 0 {
		return value
	}
	return now
}

func resolveAdminMenuDeclarationKeys(db *gorm.DB, declarations []adminmenuperm.Declaration) (map[string]string, error) {
	var existing []model.Permission
	if err := db.Where("`permission_platform` = ? AND `permission_key` LIKE ?", PlatformAdmin, "admin:menu:%").Find(&existing).Error; err != nil {
		return nil, err
	}
	byKey := make(map[string]model.Permission, len(existing))
	for _, item := range existing {
		byKey[item.Key] = item
	}
	result := make(map[string]string, len(declarations)+1)
	result[""] = ""
	for _, declaration := range declarations {
		if current, ok := byKey[declaration.Key]; ok {
			result[declaration.Key] = current.Key
			continue
		}
		if matched, ok := matchAdminMenuDeclaration(existing, declaration); ok {
			result[declaration.Key] = matched.Key
			continue
		}
		result[declaration.Key] = declaration.Key
	}
	return result, nil
}

func matchAdminMenuDeclaration(existing []model.Permission, declaration adminmenuperm.Declaration) (model.Permission, bool) {
	for _, item := range existing {
		if item.Name == declaration.Name && item.Perms == declaration.Perms && item.ResourcePath == declaration.Path {
			return item, true
		}
	}
	for _, item := range existing {
		if declaration.Type == adminmenuperm.TypeButton && item.Name == declaration.Name && item.Perms == declaration.Perms {
			return item, true
		}
		if declaration.Type != adminmenuperm.TypeButton && declaration.Path != "" && item.Name == declaration.Name && item.ResourcePath == declaration.Path {
			return item, true
		}
	}
	return model.Permission{}, false
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
			Key:       declaration.Key,
			Name:      declaration.Name,
			Platform:  PlatformAdmin,
			Type:      TypeAPI,
			ParentKey: declaration.CategoryKey,
			Perms:     declaration.Perms,
			Sort:      (index + 1) * 10,
			Status:    1,
			AddTime:   now,
			EditTime:  now,
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
		item := model.Permission{
			Key:          declaration.Key,
			Name:         declaration.Name,
			Platform:     declaration.Platform,
			Type:         TypeMenu,
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

func syncLegacyRoleGrants(db *gorm.DB) error {
	roleMenuTableReady := db.Migrator().HasTable("role_menus")
	roleDeptTableReady := db.Migrator().HasTable("role_depts")
	if !roleMenuTableReady && !roleDeptTableReady {
		return nil
	}
	var roles []model.Role
	if err := db.Find(&roles).Error; err != nil {
		return err
	}
	for _, role := range roles {
		adminPermissionKeys := make([]string, 0)
		if roleMenuTableReady {
			var menuRows []struct {
				MenuID uint `gorm:"column:role_menu_menu_id"`
			}
			if err := db.Table("role_menus").Select("role_menu_menu_id").Where("`role_menu_role_id` = ?", role.ID).Find(&menuRows).Error; err != nil {
				return err
			}
			for _, row := range menuRows {
				if row.MenuID > 0 {
					adminPermissionKeys = append(adminPermissionKeys, MenuPermissionKey(row.MenuID))
				}
			}
		}
		deptIDs := make([]uint, 0)
		if roleDeptTableReady {
			var deptRows []struct {
				DeptID uint `gorm:"column:role_dept_dept_id"`
			}
			if err := db.Table("role_depts").Select("role_dept_dept_id").Where("`role_dept_role_id` = ?", role.ID).Find(&deptRows).Error; err != nil {
				return err
			}
			for _, row := range deptRows {
				deptIDs = append(deptIDs, row.DeptID)
			}
		}
		if err := setRoleAdminPermissionKeysTx(db, role.ID, role.AllowAdminLogin, adminPermissionKeys, nil, role.DataScope, deptIDs, false); err != nil {
			return err
		}
	}
	return nil
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
	return db.Model(&current).Updates(map[string]interface{}{
		"permission_name":          item.Name,
		"permission_platform":      item.Platform,
		"permission_type":          item.Type,
		"permission_parent_key":    item.ParentKey,
		"permission_resource_id":   item.ResourceID,
		"permission_resource_path": item.ResourcePath,
		"permission_icon":          item.Icon,
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
	return tx.CreateInBatches(grants, 100).Error
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
		perm, ok := permissionByKey[key]
		if !ok {
			if strings.HasPrefix(key, "admin:menu:") {
				continue
			}
			return nil, fmt.Errorf("%w: %s", gorm.ErrRecordNotFound, key)
		}
		grants = append(grants, model.PermissionGrant{
			SubjectType:   subjectType,
			SubjectID:     subjectID,
			PermissionKey: key,
			PermissionID:  perm.ID,
			Effect:        effect,
			ScopeValue:    scopeValues[key],
			Source:        source,
			Status:        1,
			AddTime:       now,
			EditTime:      now,
		})
	}
	return grants, nil
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
		if strings.HasPrefix(key, "dingtalk_h5:menu:performance:") {
			selected["dingtalk_h5:menu:performance"] = true
		}
	}
	keys := orderedApplicationMenuKeys(selected, appmenuperm.ClientMenuDeclarations())
	keys = append(keys, orderedApplicationMenuKeys(selected, appmenuperm.DingTalkH5MenuDeclarations())...)
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

func firstDataScope(grants []model.PermissionGrant) (DataScope, bool) {
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
			return DataScope{Mode: 3, Ready: true}, true
		case DataCustomPermissionKey:
			return DataScope{Mode: 4, DeptIDs: decodeDeptScope(grant.ScopeValue), Ready: true}, true
		}
	}
	return DataScope{}, false
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
	allowed := map[string]bool{}
	denied := map[string]bool{}
	if roleID > 0 {
		var grants []model.PermissionGrant
		query := db.Where("`grant_subject_type` = ? AND `grant_subject_id` = ? AND `grant_status` = 1", SubjectRole, roleID)
		if len(prefixes) > 0 {
			where, args := likeAnyClause("`grant_permission_key`", prefixes)
			query = query.Where(where, args...)
		}
		if err := query.Find(&grants).Error; err != nil {
			return nil, nil, err
		}
		for _, grant := range grants {
			if grant.Effect == EffectAllow {
				allowed[grant.PermissionKey] = true
			}
		}
	}
	if userID > 0 {
		var grants []model.PermissionGrant
		query := db.Where("`grant_subject_type` = ? AND `grant_subject_id` = ? AND `grant_status` = 1", SubjectUser, userID)
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

func TablesReady(db *gorm.DB) bool {
	return db != nil && db.Migrator().HasTable(&model.Permission{}) && db.Migrator().HasTable(&model.PermissionGrant{})
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
