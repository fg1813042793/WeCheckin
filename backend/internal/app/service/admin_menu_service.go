package service

import (
	"strings"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ===================== Menu =====================

func GetMenuTree() ([]*model.Menu, error) {
	var list []*model.Menu
	if err := database.DB.Order("`menu_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return buildMenuTree(list, 0), nil
}

func GetMenuList() ([]model.Menu, error) {
	var list []model.Menu
	return list, database.DB.Order("`menu_sort` ASC, `id` ASC").Find(&list).Error
}

func buildMenuTree(list []*model.Menu, pid uint) []*model.Menu {
	var tree []*model.Menu
	for _, item := range list {
		if item.ParentID == pid {
			item.Children = buildMenuTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

func AddMenu(name string, parentID uint, path, perms, icon string, sort, mtype int) error {
	return database.DB.Table("menus").Create(map[string]interface{}{
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
}

func EditMenu(id uint, name string, parentID uint, path, perms, icon string, sort, status, mtype int) error {
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
	return database.DB.Model(&model.Menu{}).Where("`id` = ?", id).Updates(updates).Error
}

func DelMenu(id uint) error {
	database.DB.Where("`role_menu_menu_id` = ?", id).Delete(&model.RoleMenu{})
	database.DB.Where("`menu_parent_id` = ?", id).Delete(&model.Menu{})
	return database.DB.Where("`id` = ?", id).Delete(&model.Menu{}).Error
}

// ===================== RoleMenu =====================

func GetRoleMenuIDs(roleID uint) []uint {
	var list []model.RoleMenu
	database.DB.Where("`role_menu_role_id` = ?", roleID).Find(&list)
	ids := make([]uint, len(list))
	for i, rm := range list {
		ids[i] = rm.MenuID
	}
	return ids
}

func SetRoleMenus(roleID uint, menuIDs []uint) {
	database.DB.Where("`role_menu_role_id` = ?", roleID).Delete(&model.RoleMenu{})
	for _, menuID := range menuIDs {
		if menuID > 0 {
			database.DB.Create(&model.RoleMenu{RoleID: roleID, MenuID: menuID})
		}
	}
}

// GetAdminMenuTree returns the menu tree for an admin (filtered by role)
func GetAdminMenuTree(admin *model.Admin) ([]*model.Menu, error) {
	if admin.Type == 1 {
		var list []*model.Menu
		if err := database.DB.Where("`menu_status` = 1").Order("`menu_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
			return nil, err
		}
		return buildMenuTree(list, 0), nil
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
	return buildMenuTree(list, 0), nil
}

// GetAdminPerms returns the permission keys for an admin
func GetAdminPerms(admin *model.Admin) []string {
	if admin.Type == 1 {
		var all []model.Menu
		database.DB.Where("`menu_perms` != ''").Find(&all)
		var perms []string
		for _, m := range all {
			for _, p := range strings.Split(m.Perms, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					perms = append(perms, p)
				}
			}
		}
		return perms
	}
	if admin.RoleID == 0 {
		return nil
	}
	menuIDs := GetRoleMenuIDs(admin.RoleID)
	if len(menuIDs) == 0 {
		return nil
	}
	var menus []model.Menu
	database.DB.Where("`id` IN ? AND `menu_perms` != ''", menuIDs).Find(&menus)
	var perms []string
	for _, m := range menus {
		for _, p := range strings.Split(m.Perms, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				perms = append(perms, p)
			}
		}
	}
	return perms
}
