package adminaccess

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

const ReservedSuperAdminRoleName = "超级管理员"

func IsReservedSuperAdminRole(role model.Role) bool {
	return strings.TrimSpace(role.Name) == ReservedSuperAdminRoleName
}

func IsReservedSuperAdminRoleContext(ctx context.Context, db *gorm.DB, roleID uint) bool {
	if roleID == 0 || db == nil {
		return false
	}
	role, err := activeRoleByIDContext(ctx, db, roleID)
	return err == nil && IsReservedSuperAdminRole(role)
}

func HasReservedSuperAdminRoleWithRoleIDsContext(ctx context.Context, db *gorm.DB, roleIDs []uint) bool {
	roleIDs = normalizeRoleIDs(roleIDs...)
	if len(roleIDs) == 0 || db == nil {
		return false
	}
	roles, err := activeRolesByIDsContext(ctx, db, roleIDs)
	if err != nil {
		return false
	}
	for _, roleID := range roleIDs {
		if role, ok := roles[roleID]; ok && IsReservedSuperAdminRole(role) {
			return true
		}
	}
	return false
}

func RoleAllowsAdminAccessContext(ctx context.Context, db *gorm.DB, roleID uint) (model.Role, error) {
	return userOrRoleAllowsAdminAccessContext(ctx, db, 0, roleID)
}

func UserAllowsAdminAccessContext(ctx context.Context, db *gorm.DB, userID, roleID uint) (model.Role, error) {
	return userOrRoleAllowsAdminAccessContext(ctx, db, userID, roleID)
}

func UserAllowsAdminAccessWithRoleIDsContext(ctx context.Context, db *gorm.DB, userID uint, roleIDs []uint) (model.Role, error) {
	roleIDs = normalizeRoleIDs(roleIDs...)
	if len(roleIDs) == 0 {
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
	roles, err := activeRolesByIDsContext(ctx, db, roleIDs)
	if err != nil {
		return model.Role{}, err
	}
	if len(roles) == 0 {
		return model.Role{}, fmt.Errorf("当前角色已停用或不存在")
	}
	for _, roleID := range roleIDs {
		role, ok := roles[roleID]
		if ok && IsReservedSuperAdminRole(role) {
			return role, nil
		}
	}
	if userID == 0 {
		for _, roleID := range roleIDs {
			if _, ok := roles[roleID]; !ok {
				continue
			}
			ok, err := permissionsupport.RoleHasPermissionContext(ctx, db, roleID, permissionsupport.AdminLoginPermissionKey)
			if err != nil {
				return model.Role{}, err
			}
			if ok {
				return roles[roleID], nil
			}
		}
		return model.Role{}, fmt.Errorf("当前用户未获得后台入口权限")
	}
	ok, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, userID, roleIDs, permissionsupport.AdminLoginPermissionKey)
	if err != nil {
		return model.Role{}, err
	}
	if !ok {
		return model.Role{}, fmt.Errorf("当前用户未获得后台入口权限")
	}
	for _, roleID := range roleIDs {
		if role, ok := roles[roleID]; ok {
			return role, nil
		}
	}
	return model.Role{}, fmt.Errorf("当前角色已停用或不存在")
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
	role, err := activeRoleByIDContext(ctx, db, roleID)
	if err != nil {
		return model.Role{}, fmt.Errorf("当前角色已停用或不存在")
	}
	if IsReservedSuperAdminRole(role) {
		return role, nil
	}
	var ok bool
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

func activeRolesByIDsContext(ctx context.Context, db *gorm.DB, roleIDs []uint) (map[uint]model.Role, error) {
	result := make(map[uint]model.Role, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, nil
	}
	missing := make([]uint, 0, len(roleIDs))
	now := time.Now()
	for _, roleID := range roleIDs {
		if role, ok := getRoleAccessCache(roleID, now); ok {
			result[roleID] = role
			continue
		}
		missing = append(missing, roleID)
	}
	if len(missing) == 0 {
		return result, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	var roles []model.Role
	if err := db.Select("id", "role_name", "role_status", "role_data_scope").
		Where("`id` IN ? AND `role_status` = 1", missing).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	for _, role := range roles {
		result[role.ID] = role
		setRoleAccessCache(role, now)
	}
	return result, nil
}

func activeRoleByIDContext(ctx context.Context, db *gorm.DB, roleID uint) (model.Role, error) {
	now := time.Now()
	if role, ok := getRoleAccessCache(roleID, now); ok {
		return role, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return model.Role{}, err
		}
	}
	var role model.Role
	err := db.Select("id", "role_name", "role_status", "role_data_scope").
		Where("`id` = ? AND `role_status` = 1", roleID).
		Take(&role).Error
	if err != nil {
		return model.Role{}, err
	}
	setRoleAccessCache(role, now)
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
	roleWhere := "(r.`role_status` = 1 AND (r.`role_name` = ? OR EXISTS (SELECT 1 FROM `permission_grants` pg WHERE pg.`grant_subject_type` = 'role' AND pg.`grant_subject_id` = r.`id` AND pg.`grant_permission_key` = ? AND pg.`grant_effect` = 'allow' AND pg.`grant_status` = 1)))"
	return "(`user_role_id` IN (SELECT r.`id` FROM `roles` r WHERE " + roleWhere + ") OR EXISTS (SELECT 1 FROM `user_roles` ur JOIN `roles` r ON r.`id` = ur.`user_role_role_id` WHERE ur.`user_role_user_id` = `users`.`id` AND ur.`user_role_status` = 1 AND " + roleWhere + "))",
		[]interface{}{
			ReservedSuperAdminRoleName,
			permissionsupport.AdminLoginPermissionKey,
			ReservedSuperAdminRoleName,
			permissionsupport.AdminLoginPermissionKey,
		}
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
	return result
}
