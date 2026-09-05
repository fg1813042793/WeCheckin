package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"wecheckin/backend/internal/support/adminrouteperm"
	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
)

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
