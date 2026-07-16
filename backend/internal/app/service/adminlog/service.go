package adminlog

import (
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetList(keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Log
	var total int64
	query := database.DB.Model(&model.Log{})
	if keyword != "" {
		query = query.Where("`log_content` LIKE ? OR `log_admin_name` LIKE ? OR `log_admin_desc` LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = access.AdminDeptIDs(admin.ID)
				} else {
					deptIDs = access.RoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					query = query.Where("`log_admin_id` IN (SELECT `admin_dept_admin_id` FROM `admin_depts` WHERE `admin_dept_dept_id` IN ?)", deptIDs)
				}
			} else if role.DataScope == 3 {
				query = query.Where("`log_admin_id` = ?", admin.ID)
			}
		}
	}
	query.Count(&total)
	err := query.Order("`log_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func Clear() error {
	return database.DB.Where("1 = 1").Delete(&model.Log{}).Error
}
