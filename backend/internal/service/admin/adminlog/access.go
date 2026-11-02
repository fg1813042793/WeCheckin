package adminlog

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

func visibleQueryContext(ctx context.Context, db *gorm.DB, category string, adminID uint, export bool) (*gorm.DB, error) {
	if !oneOf(category, CategoryOperation, CategoryLogin) {
		return nil, invalid("category", "不支持的日志类型")
	}
	if adminID == 0 {
		return nil, ErrForbidden
	}
	if db == nil {
		return nil, ErrUnavailable
	}
	var admin model.Admin
	// An absent or disabled querying actor must not fall through to an empty scope.
	if err := db.Where("`id` = ? AND `user_status` = 1", adminID).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return nil, ErrForbidden
	}
	if _, err := adminaccess.UserAllowsAdminAccessWithRoleIDsContext(ctx, db, admin.ID, roleIDs); err != nil {
		return nil, err
	}
	query := categoryQuery(db, category)
	if adminaccess.HasReservedSuperAdminRoleWithRoleIDsContext(ctx, db, roleIDs) {
		return query, nil
	}
	keys, ready, err := permissionsupport.SubjectAPIPermissionKeysWithRoleIDsContext(ctx, db, admin.ID, roleIDs, permissionsupport.PlatformAdmin)
	if err != nil {
		return nil, err
	}
	if !ready || !hasReadPermissions(keys, category, export) {
		return nil, ErrForbidden
	}
	if category == CategoryOperation {
		// Anonymous operation attempts are reserved for the superadmin branch
		// above, even when an ordinary reader holds data:all.
		query = query.Where("`create_by` > 0")
		// Scope lookup failures must not be hidden by the legacy scope helper.
		if _, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID); err != nil {
			return nil, err
		}
		if _, err := permissionsupport.UserDataScopeExtrasContext(ctx, db, admin.ID); err != nil {
			return nil, err
		}
		where, args := access.UserIDDataScopeFilterWithDBContext(ctx, db, &admin, access.AdminLogAuditFields.CreateByField())
		if where != "" {
			query = query.Where(where, args...)
		}
	}
	// Security readers intentionally see anonymous attempts across actor scopes.
	return query, nil
}

func hasReadPermissions(keys []string, category string, export bool) bool {
	readKey, exportKey := "log:list", "log:export"
	if category == CategoryLogin {
		readKey, exportKey = "log:security", "log:security-export"
	}
	allowed := make(map[string]bool, len(keys))
	for _, key := range keys {
		allowed[key] = true
	}
	return allowed[permissionsupport.AdminAPIPermissionKey(readKey)] && (!export || allowed[permissionsupport.AdminAPIPermissionKey(exportKey)])
}
