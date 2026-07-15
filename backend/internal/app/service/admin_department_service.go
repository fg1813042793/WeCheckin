package service

import (
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// Department

func getDeptDescendantIDs(all []*model.Department, parentIDs []uint) []uint {
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

func getDeptVisibleIDs(admin *model.Admin) []uint {
	if admin.Type == 1 {
		return nil
	}
	var role model.Role
	if err := database.DB.First(&role, admin.RoleID).Error; err != nil {
		return nil
	}
	var all []*model.Department
	database.DB.Find(&all)
	switch role.DataScope {
	case 1:
		return nil
	case 2:
		deptIDs := getAdminDeptIDs(admin.ID)
		if len(deptIDs) == 0 {
			return nil
		}
		return getDeptDescendantIDs(all, deptIDs)
	case 3:
		return getAdminDeptIDs(admin.ID)
	case 4:
		deptIDs := GetRoleDeptIDs(admin.RoleID)
		if len(deptIDs) == 0 {
			return nil
		}
		return getDeptDescendantIDs(all, deptIDs)
	}
	return nil
}

func GetDeptTree(adminID uint) ([]*model.Department, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []*model.Department
	if err := database.DB.Order("`dept_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	visibleIDs := getDeptVisibleIDs(&admin)
	if visibleIDs != nil {
		visibleSet := make(map[uint]bool)
		for _, id := range visibleIDs {
			visibleSet[id] = true
		}
		var filtered []*model.Department
		for _, d := range list {
			if visibleSet[d.ID] {
				filtered = append(filtered, d)
			}
		}
		// Detach from non-visible parents - set ParentID=0 so they appear as root
		for _, d := range filtered {
			if d.ParentID != 0 && !visibleSet[d.ParentID] {
				d.ParentID = 0
			}
		}
		list = filtered
	}
	return buildDeptTree(list, 0), nil
}

func buildDeptTree(list []*model.Department, pid uint) []*model.Department {
	var tree []*model.Department
	for _, item := range list {
		if item.ParentID == pid {
			item.Children = buildDeptTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

func AddDept(name string, parentID uint, sort int) error {
	dept := model.Department{
		Name:     name,
		ParentID: parentID,
		Sort:     sort,
		Status:   1,
		AddTime:  database.Now(),
	}
	return database.DB.Create(&dept).Error
}

func EditDept(id uint, name string, parentID uint, sort, status int) error {
	updates := map[string]interface{}{
		"dept_name":      name,
		"dept_parent_id": parentID,
		"dept_sort":      sort,
		"dept_status":    status,
		"dept_edit_time": database.Now(),
	}
	return database.DB.Model(&model.Department{}).Where("`id` = ?", id).Updates(updates).Error
}

func DelDept(id uint) error {
	tx := database.DB.Begin()
	tx.Where("`dept_parent_id` = ?", id).Delete(&model.Department{})
	tx.Where("`id` = ?", id).Delete(&model.Department{})
	return tx.Commit().Error
}
