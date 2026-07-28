package adminaccess

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
)

const ReservedSuperAdminRoleName = "超级管理员"

func IsReservedSuperAdminRole(role model.Role) bool {
	return strings.TrimSpace(role.Name) == ReservedSuperAdminRoleName
}

func IsReservedSuperAdminRoleContext(ctx context.Context, db *gorm.DB, roleID uint) bool {
	if roleID == 0 || db == nil {
		return false
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return false
		}
	}
	var role model.Role
	err := db.Select("id", "role_name", "role_status").
		Where("`id` = ? AND `role_status` = 1", roleID).
		First(&role).Error
	return err == nil && IsReservedSuperAdminRole(role)
}

func RoleAllowsAdminAccessContext(ctx context.Context, db *gorm.DB, roleID uint) (model.Role, error) {
	return userOrRoleAllowsAdminAccessContext(ctx, db, 0, roleID)
}

func UserAllowsAdminAccessContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (model.Role, error) {
	return userOrRoleAllowsAdminAccessContext(ctx, db, userID, roleID)
}

func userOrRoleAllowsAdminAccessContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (model.Role, error) {
	if roleID == 0 {
		return model.Role{}, fmt.Errorf("当前用户未分配后台角色")
	}
	if db == nil {
		return model.Role{}, fmt.Errorf("数据库连接异常")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return model.Role{}, err
		}
	}
	var role model.Role
	if err := db.Where("`id` = ? AND `role_status` = 1", roleID).First(&role).Error; err != nil {
		return model.Role{}, fmt.Errorf("当前角色已停用或不存在")
	}
	if IsReservedSuperAdminRole(role) {
		return role, nil
	}
	var ok bool
	var err error
	if userID == 0 {
		ok, err = permissionsupport.RoleHasPermissionContext(ctx, db, roleID, permissionsupport.AdminLoginPermissionKey)
	} else {
		ok, err = permissionsupport.SubjectHasPermissionContext(ctx, db, userID, roleID, permissionsupport.AdminLoginPermissionKey)
	}
	if err != nil {
		return model.Role{}, err
	}
	if !ok {
		return model.Role{}, fmt.Errorf("当前用户未获得后台入口权限")
	}
	return role, nil
}

func ApplyUserAdminAccessRoleFilter(db *gorm.DB) *gorm.DB {
	if !permissionsupport.TablesReady(db) {
		return db.Where("`user_role_id` IN (SELECT r.`id` FROM `roles` r WHERE r.`role_status` = 1)")
	}
	where, args := UserAdminAccessRoleFilter()
	return db.Where(where, args...)
}

func UserAdminAccessRoleFilter() (string, []interface{}) {
	return "`user_role_id` IN (SELECT r.`id` FROM `roles` r WHERE r.`role_status` = 1 AND (r.`role_name` = ? OR EXISTS (SELECT 1 FROM `permission_grants` pg WHERE pg.`grant_subject_type` = 'role' AND pg.`grant_subject_id` = r.`id` AND pg.`grant_permission_key` = ? AND pg.`grant_effect` = 'allow' AND pg.`grant_status` = 1)))",
		[]interface{}{ReservedSuperAdminRoleName, permissionsupport.AdminLoginPermissionKey}
}
