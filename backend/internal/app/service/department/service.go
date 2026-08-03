package department

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func GetTree(adminID uint) ([]*model.Department, error) {
	return GetTreeContext(context.Background(), adminID)
}

func GetTreeContext(ctx context.Context, adminID uint) ([]*model.Department, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	db.First(&admin, adminID)
	var list []*model.Department
	if err := db.Order("`dept_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	visibleIDs := access.VisibleDeptIDsContext(ctx, &admin)
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
	return AddContext(context.Background(), name, parentID, sort)
}

func AddContext(ctx context.Context, name string, parentID uint, sort int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	dept := model.Department{
		Name:     name,
		ParentID: parentID,
		Sort:     sort,
		Status:   1,
		AddTime:  database.Now(),
	}
	return db.Create(&dept).Error
}

func Edit(id uint, name string, parentID uint, sort, status int) error {
	return EditContext(context.Background(), id, name, parentID, sort, status)
}

func EditContext(ctx context.Context, id uint, name string, parentID uint, sort, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"dept_name":      name,
		"dept_parent_id": parentID,
		"dept_sort":      sort,
		"dept_status":    status,
		"dept_edit_time": database.Now(),
	}
	return db.Model(&model.Department{}).Where("`id` = ?", id).Updates(updates).Error
}

func Delete(id uint) error {
	return DeleteContext(context.Background(), id)
}

func DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`dept_parent_id` = ?", id).Delete(&model.Department{}).Error; err != nil {
			return err
		}
		return tx.Where("`id` = ?", id).Delete(&model.Department{}).Error
	})
}
