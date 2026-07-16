package department

import (
	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetTree(adminID uint) ([]*model.Department, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []*model.Department
	if err := database.DB.Order("`dept_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	visibleIDs := access.VisibleDeptIDs(&admin)
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
		for _, d := range filtered {
			if d.ParentID != 0 && !visibleSet[d.ParentID] {
				d.ParentID = 0
			}
		}
		list = filtered
	}
	return buildTree(list, 0), nil
}

func buildTree(list []*model.Department, pid uint) []*model.Department {
	var tree []*model.Department
	for _, item := range list {
		if item.ParentID == pid {
			item.Children = buildTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

func Add(name string, parentID uint, sort int) error {
	dept := model.Department{
		Name:     name,
		ParentID: parentID,
		Sort:     sort,
		Status:   1,
		AddTime:  database.Now(),
	}
	return database.DB.Create(&dept).Error
}

func Edit(id uint, name string, parentID uint, sort, status int) error {
	updates := map[string]interface{}{
		"dept_name":      name,
		"dept_parent_id": parentID,
		"dept_sort":      sort,
		"dept_status":    status,
		"dept_edit_time": database.Now(),
	}
	return database.DB.Model(&model.Department{}).Where("`id` = ?", id).Updates(updates).Error
}

func Delete(id uint) error {
	tx := database.DB.Begin()
	tx.Where("`dept_parent_id` = ?", id).Delete(&model.Department{})
	tx.Where("`id` = ?", id).Delete(&model.Department{})
	return tx.Commit().Error
}
