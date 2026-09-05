package permission

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"sort"
	"strings"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
)

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
