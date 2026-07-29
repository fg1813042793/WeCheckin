package adminmgr

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	onlineservice "wecheckin-backend/backend/internal/app/service/online"
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/adminaccess"
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
	RoleName string `json:"roleName"`
	LoginCnt int    `json:"loginCnt"`
	AddTime  int64  `json:"addTime"`
	EditTime int64  `json:"editTime"`
	DeptIDs  []uint `json:"deptIds"`
}

var adminManagerListColumns = []string{
	"id",
	"user_name",
	"user_admin_desc",
	"user_pic",
	"user_mobile",
	"user_status",
	"user_admin_type",
	"user_role_id",
	"user_login_cnt",
	"user_login_time",
	"user_add_time",
	"user_edit_time",
}

func adminLoginRoleFilter(db *gorm.DB) *gorm.DB {
	return adminaccess.ApplyUserAdminAccessRoleFilter(db)
}

func ensureAdminLoginRole(ctx context.Context, db *gorm.DB, roleID uint) error {
	_, err := adminaccess.RoleAllowsAdminAccessContext(ctx, db, roleID)
	return err
}

func GetList(adminID uint, keyword string, page, pageSize int) (*ListResponse, error) {
	return GetListContext(context.Background(), adminID, keyword, page, pageSize)
}

func GetListContext(ctx context.Context, adminID uint, keyword string, page, pageSize int) (*ListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	db.First(&admin, adminID)
	conditions := []func(*gorm.DB) *gorm.DB{
		func(d *gorm.DB) *gorm.DB {
			return adminLoginRoleFilter(d)
		},
	}
	if !adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) && admin.RoleID > 0 {
		var role model.Role
		if err := db.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = access.AdminDeptIDsContext(ctx, admin.ID)
				} else {
					deptIDs = access.RoleDeptIDsContext(ctx, admin.RoleID)
				}
				if len(deptIDs) > 0 {
					conditions = append(conditions, func(d *gorm.DB) *gorm.DB {
						return d.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
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
			likeKeyword := "%" + kw + "%"
			return d.Where("`user_name` LIKE ? OR `user_mobile` LIKE ? OR `user_admin_desc` LIKE ?", likeKeyword, likeKeyword, likeKeyword)
		})
	}
	var total int64
	queryBuilder := db.Model(&model.User{}).Where("`user_role_id` > 0").Scopes(conditions...)
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.User
	if err := queryBuilder.Select(adminManagerListColumns).Order("`user_add_time` DESC, `id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	roleNames, err := loadRoleNameMapContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	deptIDsByAdmin, err := loadAdminDeptIDMapContext(ctx, db, list)
	if err != nil {
		return nil, err
	}
	result := make([]ListItem, len(list))
	for i, a := range list {
		result[i] = ListItem{
			ID:       a.ID,
			Name:     a.Name,
			Desc:     a.AdminDesc,
			Pic:      media.FullURLWithStaticDomain(a.Pic),
			Phone:    a.Mobile,
			Status:   a.Status,
			Type:     a.AdminType,
			RoleID:   a.RoleID,
			RoleName: roleNames[a.RoleID],
			LoginCnt: a.LoginCnt,
			AddTime:  a.AddTime,
			EditTime: a.EditTime,
			DeptIDs:  deptIDsByAdmin[a.ID],
		}
	}
	return &ListResponse{List: result, Total: total}, nil
}

func loadRoleNameMapContext(ctx context.Context, db *gorm.DB, list []model.User) (map[uint]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	roleIDs := make([]uint, 0, len(list))
	seen := make(map[uint]struct{}, len(list))
	for _, item := range list {
		if item.RoleID == 0 {
			continue
		}
		if _, ok := seen[item.RoleID]; ok {
			continue
		}
		seen[item.RoleID] = struct{}{}
		roleIDs = append(roleIDs, item.RoleID)
	}
	roleNames := make(map[uint]string, len(roleIDs))
	if len(roleIDs) == 0 {
		return roleNames, nil
	}
	var roles []model.Role
	if err := db.Where("`id` IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	for _, role := range roles {
		roleNames[role.ID] = role.Name
	}
	return roleNames, nil
}

func loadAdminDeptIDMapContext(ctx context.Context, db *gorm.DB, list []model.User) (map[uint][]uint, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	adminIDs := make([]uint, 0, len(list))
	deptIDsByAdmin := make(map[uint][]uint, len(list))
	for _, item := range list {
		adminIDs = append(adminIDs, item.ID)
		deptIDsByAdmin[item.ID] = []uint{}
	}
	if len(adminIDs) == 0 {
		return deptIDsByAdmin, nil
	}
	var rows []model.UserDept
	if err := db.Where("`user_dept_user_id` IN ?", adminIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		deptIDsByAdmin[row.UserID] = append(deptIDsByAdmin[row.UserID], row.DeptID)
	}
	return deptIDsByAdmin, nil
}

func Insert(name, password, desc, phone, addIP string, typ int, roleID uint, deptIDs []uint) error {
	return InsertContext(context.Background(), name, password, desc, phone, addIP, typ, roleID, deptIDs)
}

func InsertContext(ctx context.Context, name, password, desc, phone, addIP string, typ int, roleID uint, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var cnt int64
	if err := ensureAdminLoginRole(ctx, db, roleID); err != nil {
		return err
	}
	if err := db.Model(&model.Admin{}).Where("`user_name` = ?", name).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("用户姓名已存在")
	}
	hash, err := passwordutil.Hash(password)
	if err != nil {
		return err
	}
	now := database.Now()
	admin := model.Admin{
		MiniOpenID: fmt.Sprintf("admin:%s:%d", name, now),
		Name:       name,
		Password:   hash,
		Desc:       desc,
		Phone:      phone,
		Status:     1,
		Type:       typ,
		RoleID:     roleID,
		AddTime:    now,
		EditTime:   now,
		AddIP:      addIP,
		EditIP:     addIP,
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
	if err := db.Where("`id` = ?", id).Scopes(adminLoginRoleFilter).First(&admin).Error; err != nil {
		return err
	}
	if adminaccess.IsReservedSuperAdminRoleContext(ctx, db, admin.RoleID) {
		return fmt.Errorf("超级管理员不可删除")
	}
	onlineservice.ForceOfflineAdmin(id, "")
	return db.Model(&model.Admin{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"user_role_id":          0,
		"user_admin_token":      "",
		"user_admin_token_time": int64(0),
		"user_edit_time":        database.Now(),
	}).Error
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
	err := db.Where("`id` = ?", id).Scopes(adminLoginRoleFilter).First(&admin).Error
	if err != nil {
		return nil, err
	}
	roleNames, err := loadRoleNameMapContext(ctx, db, []model.User{{ID: admin.ID, RoleID: admin.RoleID}})
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
		RoleName: roleNames[admin.RoleID],
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

	if err := ensureAdminLoginRole(ctx, db, roleID); err != nil {
		return err
	}
	updates := map[string]interface{}{
		"user_name":       name,
		"user_admin_desc": desc,
		"user_mobile":     phone,
		"user_role_id":    roleID,
		"user_edit_time":  database.Now(),
		"user_edit_ip":    addIP,
	}
	if pic != "" {
		updates["user_pic"] = pic
	}
	if password != "" {
		hash, err := passwordutil.Hash(password)
		if err != nil {
			return err
		}
		updates["user_password"] = hash
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

	err := db.Model(&model.Admin{}).Where("`id` = ?", id).Scopes(adminLoginRoleFilter).Update("user_status", status).Error
	if err == nil && status != 1 {
		onlineservice.ForceOfflineAdmin(id, "")
	}
	return err
}
