package dict

import (
	"context"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type TypeSummary struct {
	TypeCode string `json:"typeCode"`
	TypeName string `json:"typeName"`
	ItemCnt  int64  `json:"itemCnt"`
}

func GetTypes() ([]TypeSummary, error) {
	return GetTypesContext(context.Background())
}

func GetTypesContext(ctx context.Context) ([]TypeSummary, error) {
	var results []TypeSummary
	db, cancel := database.WithContext(ctx)
	defer cancel()
	rows, err := db.Model(&model.SysDict{}).
		Select("`dict_type_code` as type_code, `dict_type_name` as type_name, COUNT(*) as item_cnt").
		Group("`dict_type_code`, `dict_type_name`").
		Order("MIN(`dict_sort`) ASC, MIN(`dict_add_time`) ASC").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var typeCode, typeName string
		var itemCnt int64
		if err := rows.Scan(&typeCode, &typeName, &itemCnt); err != nil {
			continue
		}
		results = append(results, TypeSummary{TypeCode: typeCode, TypeName: typeName, ItemCnt: itemCnt})
	}
	return results, nil
}

func GetByType(typeCode string) ([]model.SysDict, error) {
	return GetByTypeContext(context.Background(), typeCode)
}

func GetByTypeContext(ctx context.Context, typeCode string) ([]model.SysDict, error) {
	var list []model.SysDict
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return list, db.Where("`dict_type_code` = ?", typeCode).Order("`dict_sort` ASC, `id` ASC").Find(&list).Error
}

func AddItem(typeCode, typeName, label, value, remark string, sort int) error {
	return AddItemContext(context.Background(), typeCode, typeName, label, value, remark, sort)
}

func AddItemContext(ctx context.Context, typeCode, typeName, label, value, remark string, sort int) error {
	now := database.Now()
	d := model.SysDict{
		TypeCode: typeCode,
		TypeName: typeName,
		Label:    label,
		Value:    value,
		Sort:     sort,
		Status:   1,
		Remark:   remark,
		AddTime:  now,
		EditTime: now,
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Create(&d).Error
}

func EditItem(id, label, value, remark string, sort int) error {
	return EditItemContext(context.Background(), id, label, value, remark, sort)
}

func EditItemContext(ctx context.Context, id, label, value, remark string, sort int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.SysDict{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"dict_label":     label,
		"dict_value":     value,
		"dict_sort":      sort,
		"dict_remark":    remark,
		"dict_edit_time": database.Now(),
	}).Error
}

func DeleteItem(id string) error {
	return DeleteItemContext(context.Background(), id)
}

func DeleteItemContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`id` = ?", id).Delete(&model.SysDict{}).Error
}

func DeleteByType(typeCode string) error {
	return DeleteByTypeContext(context.Background(), typeCode)
}

func DeleteByTypeContext(ctx context.Context, typeCode string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`dict_type_code` = ?", typeCode).Delete(&model.SysDict{}).Error
}

func EditTypeName(oldTypeCode, typeCode, typeName string) error {
	return EditTypeNameContext(context.Background(), oldTypeCode, typeCode, typeName)
}

func EditTypeNameContext(ctx context.Context, oldTypeCode, typeCode, typeName string) error {
	updates := map[string]interface{}{"dict_type_name": typeName}
	if oldTypeCode != typeCode {
		updates["dict_type_code"] = typeCode
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.SysDict{}).Where("`dict_type_code` = ?", oldTypeCode).Updates(updates).Error
}
