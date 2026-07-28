package access

import (
	"context"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/app/support/adminaccess"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func AdminDeptIDs(adminID uint) []uint {
	return AdminDeptIDsContext(context.Background(), adminID)
}

func AdminDeptIDsContext(ctx context.Context, adminID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var list []model.UserDept
	db.Where("`user_dept_user_id` = ?", adminID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func SaveAdminDepts(adminID uint, deptIDs []uint) {
	_ = SaveAdminDeptsContext(context.Background(), adminID, deptIDs)
}

func SaveAdminDeptsContext(ctx context.Context, adminID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return SaveAdminDeptsTx(db, adminID, deptIDs)
}

func SaveAdminDeptsTx(tx *gorm.DB, adminID uint, deptIDs []uint) error {
	if err := tx.Where("`user_dept_user_id` = ?", adminID).Delete(&model.UserDept{}).Error; err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if deptID > 0 {
			if err := tx.Create(&model.UserDept{UserID: adminID, DeptID: deptID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func RoleDeptIDs(roleID uint) []uint {
	return RoleDeptIDsContext(context.Background(), roleID)
}

func RoleDeptIDsContext(ctx context.Context, roleID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	deptIDs, _ := permissionsupport.RoleCustomDeptIDsContext(ctx, db, roleID)
	return deptIDs
}

func DeptDescendantIDs(all []*model.Department, parentIDs []uint) []uint {
	ids := make([]uint, 0)
	idSet := make(map[uint]bool)
	for _, id := range parentIDs {
		idSet[id] = true
	}
	queue := make([]uint, len(parentIDs))
	copy(queue, parentIDs)
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, d := range all {
			if d.ParentID == pid && !idSet[d.ID] {
				idSet[d.ID] = true
				queue = append(queue, d.ID)
			}
		}
	}
	for id := range idSet {
		ids = append(ids, id)
	}
	return ids
}

func VisibleDeptIDs(admin *model.Admin) []uint {
	return VisibleDeptIDsContext(context.Background(), admin)
}

func VisibleDeptIDsContext(ctx context.Context, admin *model.Admin) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		return nil
	}
	if scope, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID); err == nil && scope.Ready {
		return visibleDeptIDsByScope(ctx, db, admin, scope.Mode, scope.DeptIDs)
	}
	var role model.Role
	if err := db.First(&role, admin.RoleID).Error; err != nil {
		return nil
	}
	var all []*model.Department
	db.Find(&all)
	switch role.DataScope {
	case 1:
		return nil
	case 2:
		deptIDs := AdminDeptIDsContext(ctx, admin.ID)
		if len(deptIDs) == 0 {
			return []uint{}
		}
		return DeptDescendantIDs(all, deptIDs)
	case 3:
		return AdminDeptIDsContext(ctx, admin.ID)
	case 4:
		deptIDs := RoleDeptIDsContext(ctx, admin.RoleID)
		if len(deptIDs) == 0 {
			return []uint{}
		}
		return DeptDescendantIDs(all, deptIDs)
	}
	return nil
}

func DataScopeFilter(admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	return DataScopeFilterContext(context.Background(), admin, deptField, createByField)
}

func DataScopeFilterContext(ctx context.Context, admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		return "", nil
	}
	if scope, err := permissionsupport.DataScopeContext(ctx, db, admin.ID, admin.RoleID); err == nil && scope.Ready {
		return dataScopeFilterByMode(ctx, admin, deptField, createByField, scope.Mode, scope.DeptIDs)
	}
	var role model.Role
	if err := db.First(&role, admin.RoleID).Error; err != nil {
		return "", nil
	}
	switch role.DataScope {
	case 1:
		return "", nil
	case 2:
		deptIDs := AdminDeptIDsContext(ctx, admin.ID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{toInterfaceSlice(deptIDs)}
	case 3:
		return createByField + " = ?", []interface{}{admin.ID}
	case 4:
		deptIDs := RoleDeptIDsContext(ctx, admin.RoleID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{toInterfaceSlice(deptIDs)}
	}
	return "", nil
}

func ScopedResourceQueryContext(ctx context.Context, db *gorm.DB, adminID uint, resource interface{}, deptField, createByField string) (*gorm.DB, error) {
	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	query := db.Model(resource)
	where, args := DataScopeFilterContext(ctx, &admin, deptField, createByField)
	if where != "" {
		query = query.Where(where, args...)
	}
	return query, nil
}

func UserDataScopeFilterContext(ctx context.Context, admin *model.Admin) (string, []interface{}) {
	deptIDs := VisibleDeptIDsContext(ctx, admin)
	if deptIDs == nil {
		return "", nil
	}
	if len(deptIDs) == 0 {
		return "1 = 0", nil
	}
	return "`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", []interface{}{deptIDs}
}

func RequireRowsAffected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func visibleDeptIDsByScope(ctx context.Context, db *gorm.DB, admin *model.Admin, mode int, customDeptIDs []uint) []uint {
	var all []*model.Department
	db.Find(&all)
	switch mode {
	case 1:
		return nil
	case 2:
		deptIDs := AdminDeptIDsContext(ctx, admin.ID)
		if len(deptIDs) == 0 {
			return []uint{}
		}
		return DeptDescendantIDs(all, deptIDs)
	case 3:
		return AdminDeptIDsContext(ctx, admin.ID)
	case 4:
		if len(customDeptIDs) == 0 {
			return []uint{}
		}
		return DeptDescendantIDs(all, customDeptIDs)
	}
	return nil
}

func dataScopeFilterByMode(ctx context.Context, admin *model.Admin, deptField, createByField string, mode int, customDeptIDs []uint) (string, []interface{}) {
	switch mode {
	case 1:
		return "", nil
	case 2:
		deptIDs := AdminDeptIDsContext(ctx, admin.ID)
		if len(deptIDs) == 0 {
			return deptField + " = 0", nil
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{toInterfaceSlice(deptIDs)}
	case 3:
		return createByField + " = ?", []interface{}{admin.ID}
	case 4:
		if len(customDeptIDs) == 0 {
			return deptField + " = 0", nil
		}
		return "(" + deptField + " IN ? OR " + deptField + " = 0)", []interface{}{toInterfaceSlice(customDeptIDs)}
	}
	return "", nil
}

func toInterfaceSlice(ids []uint) []interface{} {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}
