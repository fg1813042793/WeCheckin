package service

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

func GetMgrList(adminID uint, keyword string, page, pageSize int) (map[string]interface{}, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var conditions []func(*gorm.DB) *gorm.DB
	// Data scope filter for admins
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
					conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
						return d.Where("`id` IN (SELECT `admin_dept_admin_id` FROM `admin_depts` WHERE `admin_dept_dept_id` IN ?)", deptIDs)
					})
				}
			} else if role.DataScope == 3 {
				id := admin.ID
				conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
					return d.Where("`id` = ?", id)
				})
			}
		}
	}
	if keyword != "" {
		kw := keyword
		conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
			return d.Where("`admin_name` LIKE ? OR `admin_phone` LIKE ?", "%"+kw+"%", "%"+kw+"%")
		})
	}
	var total int64
	database.DB.Model(&model.Admin{}).Scopes(conditions...).Count(&total)
	var list []model.Admin
	database.DB.Model(&model.Admin{}).Scopes(conditions...).Order("`admin_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	result := make([]map[string]interface{}, len(list))
	for i, a := range list {
		roleName := ""
		if a.RoleID > 0 {
			var role model.Role
			if err := database.DB.First(&role, a.RoleID).Error; err == nil {
				roleName = role.Name
			}
		}
		result[i] = map[string]interface{}{
			"id": a.ID, "name": a.Name, "desc": a.Desc, "pic": GetFullURL(a.Pic),
			"phone": a.Phone, "status": a.Status, "type": a.Type,
			"roleId": a.RoleID, "roleName": roleName, "loginCnt": a.LoginCnt,
			"addTime": a.AddTime, "editTime": a.EditTime,
			"deptIds": getAdminDeptIDs(a.ID),
		}
	}
	return map[string]interface{}{"list": result, "total": total}, nil
}

func getAdminDeptIDs(adminID uint) []uint {
	var list []model.AdminDept
	database.DB.Where("`admin_dept_admin_id` = ?", adminID).Find(&list)
	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.DeptID
	}
	return ids
}

func saveAdminDepts(adminID uint, deptIDs []uint) {
	database.DB.Where("`admin_dept_admin_id` = ?", adminID).Delete(&model.AdminDept{})
	for _, deptID := range deptIDs {
		if deptID > 0 {
			database.DB.Create(&model.AdminDept{AdminID: adminID, DeptID: deptID})
		}
	}
}

func InsertMgr(name, password, desc, phone, addIP string, typ int, roleID uint, deptIDs []uint) error {
	var cnt int64
	database.DB.Model(&model.Admin{}).Where("`admin_name` = ?", name).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("管理员已存在")
	}
	hash, err := passwordutil.Hash(password)
	if err != nil {
		return err
	}
	admin := model.Admin{
		Name:     name,
		Password: hash,
		Desc:     desc,
		Phone:    phone,
		Status:   1,
		Type:     typ,
		RoleID:   roleID,
		AddTime:  database.Now(),
		AddIP:    addIP,
	}
	if err := database.DB.Create(&admin).Error; err != nil {
		return err
	}
	saveAdminDepts(admin.ID, deptIDs)
	return nil
}

func DelMgr(id string) error {
	var admin model.Admin
	if err := database.DB.Where("`id` = ?", id).First(&admin).Error; err != nil {
		return err
	}
	if admin.Type == 1 {
		return fmt.Errorf("超级管理员不可删除")
	}
	ForceOfflineAdmin(id, "")
	database.DB.Where("`admin_dept_admin_id` = ?", id).Delete(&model.AdminDept{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Admin{}).Error
}

func DelMgrs(ids []string) error {
	for _, id := range ids {
		if err := DelMgr(id); err != nil {
			return err
		}
	}
	return nil
}

func GetMgrDetail(id string) (map[string]interface{}, error) {
	var admin model.Admin
	err := database.DB.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(id)
	return map[string]interface{}{
		"id": admin.ID, "name": admin.Name, "desc": admin.Desc,
		"pic": GetFullURL(admin.Pic), "phone": admin.Phone, "status": admin.Status,
		"type": admin.Type, "roleId": admin.RoleID,
		"loginCnt": admin.LoginCnt,
		"addTime":  admin.AddTime, "editTime": admin.EditTime,
		"deptIds": getAdminDeptIDs(uint(uid)),
	}, nil
}

func EditMgr(id, name, desc, pic, phone, password, addIP string, roleID uint, deptIDs []uint) error {
	updates := map[string]interface{}{
		"admin_name":      name,
		"admin_desc":      desc,
		"admin_phone":     phone,
		"admin_role_id":   roleID,
		"admin_edit_time": database.Now(),
		"admin_edit_ip":   addIP,
	}
	if pic != "" {
		updates["admin_pic"] = pic
	}
	if password != "" {
		hash, err := passwordutil.Hash(password)
		if err != nil {
			return err
		}
		updates["admin_password"] = hash
	}
	if err := database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	uid, _ := strconv.Atoi(id)
	saveAdminDepts(uint(uid), deptIDs)
	return nil
}

func StatusMgr(id string, status int) error {
	err := database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Update("admin_status", status).Error
	if err == nil && status != 1 {
		ForceOfflineAdmin(id, "")
	}
	return err
}

func PwdMgr(id, oldPassword, newPassword string) error {
	var admin model.Admin
	err := database.DB.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return fmt.Errorf("管理员不存在")
	}
	if !passwordutil.Verify(admin.Password, oldPassword) {
		return fmt.Errorf("旧密码错误")
	}
	hash, err := passwordutil.Hash(newPassword)
	if err != nil {
		return err
	}
	return database.DB.Model(&model.Admin{}).Where("`id` = ?", id).Update("admin_password", hash).Error
}

func GetLogList(keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Log
	var total int64
	query := database.DB.Model(&model.Log{})
	if keyword != "" {
		query = query.Where("`log_content` LIKE ? OR `log_admin_name` LIKE ? OR `log_admin_desc` LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	// Data scope
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

func ClearLog() error {
	return database.DB.Where("1 = 1").Delete(&model.Log{}).Error
}
