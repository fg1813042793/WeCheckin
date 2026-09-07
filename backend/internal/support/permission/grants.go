package permission

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
	"strings"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
	"wecheckin/backend/pkg/database"
)

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
	result := make([]string, 0, len(clientMenuKeys)+len(dingtalkH5MenuKeys))
	seen := make(map[string]bool, cap(result))
	result = appendPermissionKeysWithPrefixes(result, seen, clientMenuKeys, "client:menu:")
	for _, key := range normalizePermissionKeys(dingtalkH5MenuKeys) {
		if !strings.HasPrefix(key, "dingtalk_h5:menu:") && !strings.HasPrefix(key, "dingtalk_h5:button:") {
			continue
		}
		if strings.HasPrefix(key, "dingtalk_h5:menu:performance:") && !seen["dingtalk_h5:menu:performance"] {
			seen["dingtalk_h5:menu:performance"] = true
			result = append(result, "dingtalk_h5:menu:performance")
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}

func normalizeRoleApplicationAPIKeys(clientAPIPermissionKeys, dingtalkH5APIPermissionKeys []string) []string {
	result := make([]string, 0, len(clientAPIPermissionKeys)+len(dingtalkH5APIPermissionKeys))
	seen := make(map[string]bool, cap(result))
	result = appendPermissionKeysWithPrefixes(result, seen, clientAPIPermissionKeys, "client:api:")
	return appendPermissionKeysWithPrefixes(result, seen, dingtalkH5APIPermissionKeys, "dingtalk_h5:api:")
}

func appendPermissionKeysWithPrefixes(result []string, seen map[string]bool, keys []string, prefixes ...string) []string {
	for _, key := range normalizePermissionKeys(keys) {
		matched := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				matched = true
				break
			}
		}
		if matched && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}

func orderedApplicationMenuKeys(selected map[string]bool, declarations []appmenuperm.Declaration) []string {
	keys := make([]string, 0, len(declarations))
	declared := make(map[string]bool, len(declarations))
	for _, declaration := range declarations {
		declared[declaration.Key] = true
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return appendCustomPermissionKeys(keys, selected, declared)
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
	declared := make(map[string]bool, len(declarations))
	for _, declaration := range declarations {
		declared[declaration.Key] = true
		if selected[declaration.Key] {
			keys = append(keys, declaration.Key)
		}
	}
	return appendCustomPermissionKeys(keys, selected, declared)
}

func appendCustomPermissionKeys(keys []string, selected, declared map[string]bool) []string {
	custom := make([]string, 0)
	for key, checked := range selected {
		if checked && !declared[key] {
			custom = append(custom, key)
		}
	}
	sort.Strings(custom)
	return append(keys, custom...)
}
