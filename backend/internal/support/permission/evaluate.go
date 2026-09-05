package permission

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"sort"
	"strings"
	"wecheckin/backend/internal/model"
)

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
