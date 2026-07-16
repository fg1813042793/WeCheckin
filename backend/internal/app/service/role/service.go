package role

import (
	"context"

	"gorm.io/gorm"

	menuservice "wecheckin-backend/backend/internal/app/service/menu"
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type ListItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Remark    string `json:"remark"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	DataScope int    `json:"dataScope"`
	AddTime   int64  `json:"addTime"`
	EditTime  int64  `json:"editTime"`
	MenuIDs   []uint `json:"menuIds"`
	DeptIDs   []uint `json:"deptIds"`
}

type ListResponse struct {
	List  []ListItem `json:"list"`
	Total int64      `json:"total"`
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
	if err := db.Model(&model.Role{}).Scopes(conditions...).Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Role
	if err := db.Model(&model.Role{}).Scopes(conditions...).Order("`role_sort` ASC, `id` ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]ListItem, len(list))
	for i, r := range list {
		result[i] = ListItem{
			ID:        r.ID,
			Name:      r.Name,
			Remark:    r.Remark,
			Sort:      r.Sort,
			Status:    r.Status,
			DataScope: r.DataScope,
			AddTime:   r.AddTime,
			EditTime:  r.EditTime,
			MenuIDs:   menuservice.GetRoleMenuIDs(r.ID),
			DeptIDs:   access.RoleDeptIDs(r.ID),
		}
	}
	return &ListResponse{List: result, Total: total}, nil
}

func Add(name, remark, addIP string, sort, dataScope int) (uint, error) {
	return AddContext(context.Background(), name, remark, addIP, sort, dataScope)
}

func AddContext(ctx context.Context, name, remark, addIP string, sort, dataScope int) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	role := model.Role{
		Name:      name,
		Remark:    remark,
		Sort:      sort,
		Status:    1,
		DataScope: dataScope,
		AddTime:   database.Now(),
		AddIP:     addIP,
	}
	if err := db.Create(&role).Error; err != nil {
		return 0, err
	}
	return role.ID, nil
}

func AddWithAssignmentsContext(ctx context.Context, name, remark, addIP string, sort, dataScope int, menuIDs, deptIDs []uint) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	role := model.Role{
		Name:      name,
		Remark:    remark,
		Sort:      sort,
		Status:    1,
		DataScope: dataScope,
		AddTime:   database.Now(),
		AddIP:     addIP,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		if err := menuservice.SetRoleMenusTx(tx, role.ID, menuIDs); err != nil {
			return err
		}
		return access.SetRoleDeptsTx(tx, role.ID, deptIDs)
	})
	if err != nil {
		return 0, err
	}
	menuservice.InvalidateAdminPermCacheForRole(role.ID)
	return role.ID, nil
}

func Edit(id uint, name, remark, addIP string, sort, status, dataScope int) error {
	return EditContext(context.Background(), id, name, remark, addIP, sort, status, dataScope)
}

func EditContext(ctx context.Context, id uint, name, remark, addIP string, sort, status, dataScope int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	updates := map[string]interface{}{
		"role_name":       name,
		"role_remark":     remark,
		"role_sort":       sort,
		"role_status":     status,
		"role_data_scope": dataScope,
		"role_edit_time":  database.Now(),
		"role_edit_ip":    addIP,
	}
	return db.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error
}

func EditWithAssignmentsContext(ctx context.Context, id uint, name, remark, addIP string, sort, status, dataScope int, menuIDs, deptIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	updates := map[string]interface{}{
		"role_name":       name,
		"role_remark":     remark,
		"role_sort":       sort,
		"role_status":     status,
		"role_data_scope": dataScope,
		"role_edit_time":  database.Now(),
		"role_edit_ip":    addIP,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Role{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := menuservice.SetRoleMenusTx(tx, id, menuIDs); err != nil {
			return err
		}
		return access.SetRoleDeptsTx(tx, id, deptIDs)
	})
	if err == nil {
		menuservice.InvalidateAdminPermCacheForRole(id)
	}
	return err
}

func Delete(id uint) error {
	return DeleteContext(context.Background(), id)
}

func DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`role_menu_role_id` = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`role_dept_role_id` = ?", id).Delete(&model.RoleDept{}).Error; err != nil {
			return err
		}
		return tx.Where("`id` = ?", id).Delete(&model.Role{}).Error
	})
	if err == nil {
		menuservice.InvalidateAdminPermCacheForRole(id)
	}
	return err
}

func BatchDelete(ids []uint) error {
	return BatchDeleteContext(context.Background(), ids)
}

func BatchDeleteContext(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		if err := DeleteContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func GetDeptIDs(roleID uint) []uint {
	return access.RoleDeptIDs(roleID)
}

func SetDepts(roleID uint, deptIDs []uint) {
	access.SetRoleDepts(roleID, deptIDs)
}

func SetDeptsContext(ctx context.Context, roleID uint, deptIDs []uint) error {
	return access.SetRoleDeptsContext(ctx, roleID, deptIDs)
}
