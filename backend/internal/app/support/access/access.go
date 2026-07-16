package access

import (
	"context"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func AdminDeptIDs(adminID uint) []uint {
	return AdminDeptIDsContext(context.Background(), adminID)
}

func AdminDeptIDsContext(ctx context.Context, adminID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var list []model.AdminDept
	db.Where("`admin_dept_admin_id` = ?", adminID).Find(&list)
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
	if err := tx.Where("`admin_dept_admin_id` = ?", adminID).Delete(&model.AdminDept{}).Error; err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if deptID > 0 {
			if err := tx.Create(&model.AdminDept{AdminID: adminID, DeptID: deptID}).Error; err != nil {
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

	var list []model.RoleDept
	db.Where("`role_dept_role_id` = ?", roleID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func SetRoleDepts(roleID uint, deptIDs []uint) {
	_ = SetRoleDeptsContext(context.Background(), roleID, deptIDs)
}

func SetRoleDeptsContext(ctx context.Context, roleID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return SetRoleDeptsTx(db, roleID, deptIDs)
}

func SetRoleDeptsTx(tx *gorm.DB, roleID uint, deptIDs []uint) error {
	if err := tx.Where("`role_dept_role_id` = ?", roleID).Delete(&model.RoleDept{}).Error; err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if deptID > 0 {
			if err := tx.Create(&model.RoleDept{RoleID: roleID, DeptID: deptID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
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
	if admin.Type == 1 {
		return nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
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
			return nil
		}
		return DeptDescendantIDs(all, deptIDs)
	case 3:
		return AdminDeptIDsContext(ctx, admin.ID)
	case 4:
		deptIDs := RoleDeptIDsContext(ctx, admin.RoleID)
		if len(deptIDs) == 0 {
			return nil
		}
		return DeptDescendantIDs(all, deptIDs)
	}
	return nil
}

func DataScopeFilter(admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	return DataScopeFilterContext(context.Background(), admin, deptField, createByField)
}

func DataScopeFilterContext(ctx context.Context, admin *model.Admin, deptField, createByField string) (string, []interface{}) {
	if admin.Type == 1 {
		return "", nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
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

func toInterfaceSlice(ids []uint) []interface{} {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}
