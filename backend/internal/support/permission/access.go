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
