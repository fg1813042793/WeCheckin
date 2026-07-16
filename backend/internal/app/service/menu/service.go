package menu

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetTree() ([]*model.Menu, error) {
	var list []*model.Menu
	if err := database.DB.Order("`menu_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return buildTree(list, 0), nil
}

func GetList() ([]model.Menu, error) {
	var list []model.Menu
	return list, database.DB.Order("`menu_sort` ASC, `id` ASC").Find(&list).Error
}

func buildTree(list []*model.Menu, pid uint) []*model.Menu {
	var tree []*model.Menu
	for _, item := range list {
		if item.ParentID == pid {
			item.Children = buildTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

func Add(name string, parentID uint, path, perms, icon string, sort, mtype int) error {
	err := database.DB.Table("menus").Create(map[string]interface{}{
		"menu_name":      name,
		"menu_parent_id": parentID,
		"menu_path":      path,
		"menu_perms":     perms,
		"menu_icon":      icon,
		"menu_sort":      sort,
		"menu_status":    1,
		"menu_type":      mtype,
		"menu_add_time":  database.Now(),
	}).Error
	if err == nil {
		InvalidateAdminPermCache()
	}
	return err
}

func Edit(id uint, name string, parentID uint, path, perms, icon string, sort, status, mtype int) error {
	updates := map[string]interface{}{
		"menu_name":      name,
		"menu_parent_id": parentID,
		"menu_path":      path,
		"menu_perms":     perms,
		"menu_icon":      icon,
		"menu_sort":      sort,
		"menu_status":    status,
		"menu_type":      mtype,
		"menu_edit_time": database.Now(),
	}
	err := database.DB.Model(&model.Menu{}).Where("`id` = ?", id).Updates(updates).Error
	if err == nil {
		InvalidateAdminPermCache()
	}
	return err
}

func Delete(id uint) error {
	database.DB.Where("`role_menu_menu_id` = ?", id).Delete(&model.RoleMenu{})
	database.DB.Where("`menu_parent_id` = ?", id).Delete(&model.Menu{})
	err := database.DB.Where("`id` = ?", id).Delete(&model.Menu{}).Error
	if err == nil {
		InvalidateAdminPermCache()
	}
	return err
}

func GetRoleMenuIDs(roleID uint) []uint {
	return GetRoleMenuIDsContext(context.Background(), roleID)
}

func GetRoleMenuIDsContext(ctx context.Context, roleID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.RoleMenu
	db.Where("`role_menu_role_id` = ?", roleID).Find(&list)
	ids := make([]uint, len(list))
	for i, rm := range list {
		ids[i] = rm.MenuID
	}
	return ids
}

func SetRoleMenus(roleID uint, menuIDs []uint) {
	_ = SetRoleMenusContext(context.Background(), roleID, menuIDs)
}

func SetRoleMenusContext(ctx context.Context, roleID uint, menuIDs []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := SetRoleMenusTx(db, roleID, menuIDs); err != nil {
		return err
	}
	InvalidateAdminPermCacheForRole(roleID)
	return nil
}

func SetRoleMenusTx(tx *gorm.DB, roleID uint, menuIDs []uint) error {
	if err := tx.Where("`role_menu_role_id` = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
		return err
	}
	for _, menuID := range menuIDs {
		if menuID > 0 {
			if err := tx.Create(&model.RoleMenu{RoleID: roleID, MenuID: menuID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func GetAdminMenuTree(admin *model.Admin) ([]*model.Menu, error) {
	if admin.Type == 1 {
		var list []*model.Menu
		if err := database.DB.Where("`menu_status` = 1").Order("`menu_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
			return nil, err
		}
		return buildTree(list, 0), nil
	}
	if admin.RoleID == 0 {
		return []*model.Menu{}, nil
	}
	menuIDs := GetRoleMenuIDs(admin.RoleID)
	if len(menuIDs) == 0 {
		return []*model.Menu{}, nil
	}
	var list []*model.Menu
	database.DB.Where("`id` IN ? AND `menu_status` = 1", menuIDs).Order("`menu_sort` ASC, `id` ASC").Find(&list)
	return buildTree(list, 0), nil
}

func GetAdminPerms(admin *model.Admin) []string {
	return GetAdminPermsContext(context.Background(), admin)
}

func GetAdminPermsContext(ctx context.Context, admin *model.Admin) []string {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if admin.Type == 1 {
		cacheKey := adminPermCacheKeyForSuperAdmin()
		if perms, ok := getAdminPermCache(cacheKey); ok {
			return perms
		}
		var all []model.Menu
		db.Where("`menu_perms` != ''").Find(&all)
		var perms []string
		for _, m := range all {
			for _, p := range strings.Split(m.Perms, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					perms = append(perms, p)
				}
			}
		}
		setAdminPermCache(cacheKey, perms)
		return perms
	}
	if admin.RoleID == 0 {
		return nil
	}
	cacheKey := adminPermCacheKeyForRole(admin.RoleID)
	if perms, ok := getAdminPermCache(cacheKey); ok {
		return perms
	}
	menuIDs := GetRoleMenuIDsContext(ctx, admin.RoleID)
	if len(menuIDs) == 0 {
		return nil
	}
	var menus []model.Menu
	db.Where("`id` IN ? AND `menu_perms` != ''", menuIDs).Find(&menus)
	var perms []string
	for _, m := range menus {
		for _, p := range strings.Split(m.Perms, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				perms = append(perms, p)
			}
		}
	}
	setAdminPermCache(cacheKey, perms)
	return perms
}
