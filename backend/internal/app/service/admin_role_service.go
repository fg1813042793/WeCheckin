package service

import (
	"gorm.io/gorm"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ===================== Role =====================

func GetRoleList(adminID uint, keyword string, page, pageSize int) (map[string]interface{}, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)

	var conditions []func(*gorm.DB) *gorm.DB

	// Data scope filter
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = getAdminDeptIDs(admin.ID)
				} else {
					deptIDs = GetRoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					ids := deptIDs
					rid := admin.RoleID
					conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
						return d.Where("`id` IN (SELECT `role_dept_role_id` FROM `role_depts` WHERE `role_dept_dept_id` IN ?) OR `id` = ?", ids, rid)
					})
				}
			} else if role.DataScope == 3 {
				rid := admin.RoleID
				conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
					return d.Where("`id` = ?", rid)
				})
			}
		}
	}
	if keyword != "" {
		kw := keyword
		conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
			return d.Where("`role_name` LIKE ?", "%"+kw+"%")
		})
	}
	var total int64
	database.DB.Model(&model.Role{}).Scopes(conditions...).Count(&total)
	var list []model.Role
	database.DB.Model(&model.Role{}).Scopes(conditions...).Order("`role_sort` ASC, `id` ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	result := make([]map[string]interface{}, len(list))
	for i, r := range list {
		result[i] = map[string]interface{}{
			"id": r.ID, "name": r.Name, "remark": r.Remark,
			"sort": r.Sort, "status": r.Status, "dataScope": r.DataScope,
			"addTime": r.AddTime, "editTime": r.EditTime,
			"menuIds": GetRoleMenuIDs(r.ID),
			"deptIds": GetRoleDeptIDs(r.ID),
		}
	}
	return map[string]interface{}{"list": result, "total": total}, nil
}

func AddRole(name, remark, addIP string, sort, dataScope int) (uint, error) {
	role := model.Role{
		Name:      name,
		Remark:    remark,
		Sort:      sort,
		Status:    1,
		DataScope: dataScope,
		AddTime:   database.Now(),
		AddIP:     addIP,
	}
	err := database.DB.Create(&role).Error
	if err != nil {
		return 0, err
	}
	return role.ID, nil
}

func EditRole(id uint, name, remark, addIP string, sort, status, dataScope int) error {
	updates := map[string]interface{}{
		"role_name":       name,
		"role_remark":     remark,
		"role_sort":       sort,
		"role_status":     status,
		"role_data_scope": dataScope,
		"role_edit_time":  database.Now(),
		"role_edit_ip":    addIP,
	}
	return database.DB.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error
}

func DelRole(id uint) error {
	database.DB.Where("`role_menu_role_id` = ?", id).Delete(&model.RoleMenu{})
	database.DB.Where("`role_dept_role_id` = ?", id).Delete(&model.RoleDept{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Role{}).Error
}

func DelRoles(ids []uint) error {
	for _, id := range ids {
		if err := DelRole(id); err != nil {
			return err
		}
	}
	return nil
}

// ===================== RoleDept =====================

func GetRoleDeptIDs(roleID uint) []uint {
	var list []model.RoleDept
	database.DB.Where("`role_dept_role_id` = ?", roleID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func SetRoleDepts(roleID uint, deptIDs []uint) {
	database.DB.Where("`role_dept_role_id` = ?", roleID).Delete(&model.RoleDept{})
	for _, deptID := range deptIDs {
		if deptID > 0 {
			database.DB.Create(&model.RoleDept{RoleID: roleID, DeptID: deptID})
		}
	}
}
