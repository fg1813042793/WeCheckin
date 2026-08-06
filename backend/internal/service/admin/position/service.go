package position

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type ListResponse struct {
	List  []model.Position `json:"list"`
	Total int64            `json:"total"`
}

func GetListContext(ctx context.Context, keyword string, page, pageSize int) (*ListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()

	query := db.Model(&model.Position{})
	if text := strings.TrimSpace(keyword); text != "" {
		query = query.Where("`position_name` LIKE ?", "%"+text+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Position
	if err := query.Order("`position_sort` ASC, `id` ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &ListResponse{List: list, Total: total}, nil
}

func AddContext(ctx context.Context, name, addIP string, sort int) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("岗位名称不能为空")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	now := database.Now()
	item := model.Position{
		Name:     name,
		Sort:     sort,
		Status:   1,
		AddTime:  now,
		EditTime: now,
		AddIP:    addIP,
		EditIP:   addIP,
	}
	return db.Create(&item).Error
}

func EditContext(ctx context.Context, id uint, name, editIP string, sort, status int) error {
	name = strings.TrimSpace(name)
	if id == 0 || name == "" {
		return errors.New("参数错误")
	}
	if status != 0 {
		status = 1
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Position{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"position_name":      name,
		"position_sort":      sort,
		"position_status":    status,
		"position_edit_time": database.Now(),
		"position_edit_ip":   editIP,
	}).Error
}

func DeleteContext(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("参数错误")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var count int64
	if err := db.Model(&model.User{}).Where("`user_position_id` = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("岗位已被用户使用，不能删除")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("`id` = ?", id).Delete(&model.Position{}).Error
	})
}
