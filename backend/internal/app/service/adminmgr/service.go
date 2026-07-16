package adminmgr

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	onlineservice "wecheckin-backend/backend/internal/app/service/online"
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
)

type ListItem struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Pic      string `json:"pic"`
	Phone    string `json:"phone"`
	Status   int    `json:"status"`
	Type     int    `json:"type"`
	RoleID   uint   `json:"roleId"`
	RoleName string `json:"roleName"`
	LoginCnt int    `json:"loginCnt"`
	AddTime  int64  `json:"addTime"`
	EditTime int64  `json:"editTime"`
	DeptIDs  []uint `json:"deptIds"`
}

type ListResponse struct {
	List  []ListItem `json:"list"`
	Total int64      `json:"total"`
}

type DetailResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Pic      string `json:"pic"`
	Phone    string `json:"phone"`
	Status   int    `json:"status"`
	Type     int    `json:"type"`
	RoleID   uint   `json:"roleId"`
	LoginCnt int    `json:"loginCnt"`
	AddTime  int64  `json:"addTime"`
	EditTime int64  `json:"editTime"`
	DeptIDs  []uint `json:"deptIds"`
}

func GetList(adminID uint, keyword string, page, pageSize int) (*ListResponse, error) {
	return GetListContext(context.Background(), adminID, keyword, page, pageSize)
}

func GetListContext(ctx context.Context, adminID uint, keyword string, page, pageSize int) (*ListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	db.First(&admin, adminID)
	var conditions []func(*gorm.DB) *gorm.DB
	if admin.Type != 1 && admin.RoleID > 0 {
		var role model.Role
		if err := db.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = access.AdminDeptIDs(admin.ID)
				} else {
					deptIDs = access.RoleDeptIDs(admin.RoleID)
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
	if err := db.Model(&model.Admin{}).Scopes(conditions...).Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Admin
	if err := db.Model(&model.Admin{}).Scopes(conditions...).Order("`admin_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]ListItem, len(list))
	for i, a := range list {
		roleName := ""
		if a.RoleID > 0 {
			var role model.Role
			if err := db.First(&role, a.RoleID).Error; err == nil {
				roleName = role.Name
			}
		}
		result[i] = ListItem{
			ID:       a.ID,
			Name:     a.Name,
			Desc:     a.Desc,
			Pic:      media.FullURLWithStaticDomain(a.Pic),
			Phone:    a.Phone,
			Status:   a.Status,
			Type:     a.Type,
			RoleID:   a.RoleID,
			RoleName: roleName,
			LoginCnt: a.LoginCnt,
			AddTime:  a.AddTime,
			EditTime: a.EditTime,
			DeptIDs:  access.AdminDeptIDs(a.ID),
		}
	}
	return &ListResponse{List: result, Total: total}, nil
}

func Insert(name, password, desc, phone, addIP string, typ int, roleID uint, deptIDs []uint) error {
	return InsertContext(context.Background(), name, password, desc, phone, addIP, typ, roleID, deptIDs)
}

func InsertContext(ctx context.Context, name, password, desc, phone, addIP string, typ int, roleID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var cnt int64
	if err := db.Model(&model.Admin{}).Where("`admin_name` = ?", name).Count(&cnt).Error; err != nil {
		return err
	}
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
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		return access.SaveAdminDeptsTx(tx, admin.ID, deptIDs)
	})
}

func Delete(id string) error {
	return DeleteContext(context.Background(), id)
}

func DeleteContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	if err := db.Where("`id` = ?", id).First(&admin).Error; err != nil {
		return err
	}
	if admin.Type == 1 {
		return fmt.Errorf("超级管理员不可删除")
	}
	onlineservice.ForceOfflineAdmin(id, "")
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`admin_dept_admin_id` = ?", id).Delete(&model.AdminDept{}).Error; err != nil {
			return err
		}
		return tx.Where("`id` = ?", id).Delete(&model.Admin{}).Error
	})
}

func BatchDelete(ids []string) error {
	return BatchDeleteContext(context.Background(), ids)
}

func BatchDeleteContext(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := DeleteContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func GetDetail(id string) (*DetailResponse, error) {
	return GetDetailContext(context.Background(), id)
}

func GetDetailContext(ctx context.Context, id string) (*DetailResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	err := db.Where("`id` = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(id)
	return &DetailResponse{
		ID:       admin.ID,
		Name:     admin.Name,
		Desc:     admin.Desc,
		Pic:      media.FullURLWithStaticDomain(admin.Pic),
		Phone:    admin.Phone,
		Status:   admin.Status,
		Type:     admin.Type,
		RoleID:   admin.RoleID,
		LoginCnt: admin.LoginCnt,
		AddTime:  admin.AddTime,
		EditTime: admin.EditTime,
		DeptIDs:  access.AdminDeptIDs(uint(uid)),
	}, nil
}

func Edit(id, name, desc, pic, phone, password, addIP string, roleID uint, deptIDs []uint) error {
	return EditContext(context.Background(), id, name, desc, pic, phone, password, addIP, roleID, deptIDs)
}

func EditContext(ctx context.Context, id, name, desc, pic, phone, password, addIP string, roleID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

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
	uid, _ := strconv.Atoi(id)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Admin{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return access.SaveAdminDeptsTx(tx, uint(uid), deptIDs)
	})
}

func SetStatus(id string, status int) error {
	return SetStatusContext(context.Background(), id, status)
}

func SetStatusContext(ctx context.Context, id string, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	err := db.Model(&model.Admin{}).Where("`id` = ?", id).Update("admin_status", status).Error
	if err == nil && status != 1 {
		onlineservice.ForceOfflineAdmin(id, "")
	}
	return err
}
