package adminlog

import (
	"context"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/pkg/database"
)

func GetList(keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	return GetListContext(context.Background(), keyword, page, pageSize, adminID)
}

func GetListContext(ctx context.Context, keyword string, page, pageSize int, adminID uint) ([]model.Log, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	db.First(&admin, adminID)
	var list []model.Log
	var total int64
	query := db.Model(&model.Log{})
	if keyword != "" {
		query = query.Where("`log_content` LIKE ? OR `log_admin_name` LIKE ? OR `log_admin_desc` LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	where, args := access.UserIDDataScopeFilterWithDBContext(ctx, db, &admin, access.AdminLogAuditFields.CreateByField())
	if where != "" {
		query = query.Where(where, args...)
	}
	query.Count(&total)
	err := query.Order("`add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func Clear() error {
	return ClearContext(context.Background())
}

func ClearContext(ctx context.Context) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("1 = 1").Delete(&model.Log{}).Error
}

func DeleteIDs(ids []uint) error {
	return DeleteIDsContext(context.Background(), ids)
}

func DeleteIDsContext(ctx context.Context, ids []uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`id` IN ?", ids).Delete(&model.Log{}).Error
}
