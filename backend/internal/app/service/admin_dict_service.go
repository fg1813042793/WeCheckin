package service

import (
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ===================== SysDict =====================

func GetDictTypes() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	rows, err := database.DB.Model(&model.SysDict{}).
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
		results = append(results, map[string]interface{}{
			"typeCode": typeCode,
			"typeName": typeName,
			"itemCnt":  itemCnt,
		})
	}
	return results, nil
}

func GetDictByType(typeCode string) ([]model.SysDict, error) {
	var list []model.SysDict
	return list, database.DB.Where("`dict_type_code` = ?", typeCode).Order("`dict_sort` ASC, `id` ASC").Find(&list).Error
}

func AddDictItem(typeCode, typeName, label, value, remark string, sort int) error {
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
	return database.DB.Create(&d).Error
}

func EditDictItem(id, label, value, remark string, sort int) error {
	return database.DB.Model(&model.SysDict{}).Where("`id` = ?", id).Updates(map[string]interface{}{
		"dict_label":     label,
		"dict_value":     value,
		"dict_sort":      sort,
		"dict_remark":    remark,
		"dict_edit_time": database.Now(),
	}).Error
}

func DelDictItem(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.SysDict{}).Error
}

func DelDictByType(typeCode string) error {
	return database.DB.Where("`dict_type_code` = ?", typeCode).Delete(&model.SysDict{}).Error
}

func EditDictTypeName(oldTypeCode, typeCode, typeName string) error {
	updates := map[string]interface{}{"dict_type_name": typeName}
	if oldTypeCode != typeCode {
		updates["dict_type_code"] = typeCode
	}
	return database.DB.Model(&model.SysDict{}).Where("`dict_type_code` = ?", oldTypeCode).Updates(updates).Error
}
