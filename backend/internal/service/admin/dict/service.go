package dict

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

var (
	ErrInvalidTypeCode   = errors.New("字典类型编码只能使用小写字母、数字、点、下划线和连字符，且必须以小写字母开头")
	ErrInvalidStatus     = errors.New("字典状态只能为启用或停用")
	ErrTypeNameRequired  = errors.New("字典类型名称不能为空")
	ErrTypeAlreadyExists = errors.New("字典类型编码已存在")
	ErrTypeNotFound      = errors.New("字典类型不存在")
	ErrTypeCodeImmutable = errors.New("字典类型编码创建后不可修改")
	ErrItemRequired      = errors.New("字典标签和值不能为空")
	ErrItemAlreadyExists = errors.New("当前字典类型下已存在相同值")
	ErrItemNotFound      = errors.New("字典项不存在")
)

var typeCodePattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,49}$`)

const legacyPlaceholderPredicateSQL = "`dict_value` = '' AND `dict_label` = `dict_type_name` AND COALESCE(`dict_sort`, 0) = 0 AND COALESCE(`dict_remark`, '') = ''"
const legacyPlaceholderSQL = "NOT (" + legacyPlaceholderPredicateSQL + ")"
const dictTypeJoinSQL = "LEFT JOIN sys_dicts AS d ON d.dict_type_code COLLATE utf8mb4_general_ci = t.dict_type_code COLLATE utf8mb4_general_ci"

type TypeSummary struct {
	TypeCode string `json:"typeCode"`
	TypeName string `json:"typeName"`
	Status   int    `json:"status"`
	Remark   string `json:"remark"`
	ItemCnt  int64  `json:"itemCnt"`
	AddTime  int64  `json:"addTime"`
	EditTime int64  `json:"editTime"`
}

func validateTypeCode(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !typeCodePattern.MatchString(value) {
		return "", ErrInvalidTypeCode
	}
	return value, nil
}

func normalizeExistingTypeCode(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 50 {
		return "", ErrInvalidTypeCode
	}
	return value, nil
}

func normalizeStatus(status int) (int, error) {
	if status != 0 && status != 1 {
		return 0, ErrInvalidStatus
	}
	return status, nil
}

func isLegacyTypePlaceholder(item model.SysDict) bool {
	return item.Value == "" && item.Label == item.TypeName && item.Sort == 0 && item.Remark == ""
}

func validateTypeFields(typeName, remark string) (string, string, error) {
	typeName = strings.TrimSpace(typeName)
	remark = strings.TrimSpace(remark)
	if typeName == "" {
		return "", "", ErrTypeNameRequired
	}
	if utf8.RuneCountInString(typeName) > 100 || utf8.RuneCountInString(remark) > 500 {
		return "", "", errors.New("字典类型名称或备注超过长度限制")
	}
	return typeName, remark, nil
}

func validateItemFields(label, value, remark string) (string, string, string, error) {
	label = strings.TrimSpace(label)
	value = strings.TrimSpace(value)
	remark = strings.TrimSpace(remark)
	if label == "" || value == "" {
		return "", "", "", ErrItemRequired
	}
	if utf8.RuneCountInString(label) > 100 || utf8.RuneCountInString(value) > 200 || utf8.RuneCountInString(remark) > 500 {
		return "", "", "", errors.New("字典标签、值或备注超过长度限制")
	}
	return label, value, remark, nil
}

func GetTypes() ([]TypeSummary, error) {
	return GetTypesContext(context.Background())
}

func GetTypesContext(ctx context.Context) ([]TypeSummary, error) {
	return getTypesContext(ctx, false)
}

func GetActiveTypesContext(ctx context.Context) ([]TypeSummary, error) {
	return getTypesContext(ctx, true)
}

func getTypesContext(ctx context.Context, activeOnly bool) ([]TypeSummary, error) {
	now := time.Now()
	if cached, ok := getScopedDictTypesCache(activeOnly, now); ok {
		return cached, nil
	}
	results := make([]TypeSummary, 0)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	itemCount := "SUM(CASE WHEN d.id IS NULL OR (d.dict_value = '' AND d.dict_label = d.dict_type_name AND COALESCE(d.dict_sort, 0) = 0 AND COALESCE(d.dict_remark, '') = '') THEN 0 ELSE 1 END)"
	if activeOnly {
		itemCount = "SUM(CASE WHEN d.id IS NULL OR d.dict_status <> 1 OR (d.dict_value = '' AND d.dict_label = d.dict_type_name AND COALESCE(d.dict_sort, 0) = 0 AND COALESCE(d.dict_remark, '') = '') THEN 0 ELSE 1 END)"
	}
	query := db.Table("sys_dict_types AS t").
		Select("t.dict_type_code AS type_code, t.dict_type_name AS type_name, t.dict_type_status AS status, t.dict_type_remark AS remark, t.dict_add_time AS add_time, t.dict_edit_time AS edit_time, " + itemCount + " AS item_cnt").
		Joins(dictTypeJoinSQL).
		Group("t.dict_type_code, t.dict_type_name, t.dict_type_status, t.dict_type_remark, t.dict_add_time, t.dict_edit_time").
		Order("t.dict_type_name ASC, t.dict_type_code ASC")
	if activeOnly {
		query = query.Where("t.dict_type_status = ?", 1)
	}
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}
	setScopedDictTypesCache(activeOnly, results, now)
	return results, nil
}

func GetByType(typeCode string) ([]model.SysDict, error) {
	return GetByTypeContext(context.Background(), typeCode)
}

func GetByTypeContext(ctx context.Context, typeCode string) ([]model.SysDict, error) {
	return getByTypeContext(ctx, typeCode, false)
}

func GetActiveByTypeContext(ctx context.Context, typeCode string) ([]model.SysDict, error) {
	return getByTypeContext(ctx, typeCode, true)
}

func getByTypeContext(ctx context.Context, typeCode string, activeOnly bool) ([]model.SysDict, error) {
	typeCode = strings.TrimSpace(typeCode)
	if typeCode == "" {
		return nil, ErrInvalidTypeCode
	}
	now := time.Now()
	if cached, ok := getScopedDictItemsCache(typeCode, activeOnly, now); ok {
		return cached, nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if activeOnly {
		var count int64
		if err := db.Model(&model.SysDictType{}).Where("dict_type_code = ? AND dict_type_status = ?", typeCode, 1).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return []model.SysDict{}, nil
		}
	}
	list := make([]model.SysDict, 0)
	query := db.Where("`dict_type_code` = ?", typeCode).Where(legacyPlaceholderSQL)
	if activeOnly {
		query = query.Where("`dict_status` = ?", 1)
	}
	if err := query.Order("`dict_sort` ASC, `id` ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	setScopedDictItemsCache(typeCode, activeOnly, list, now)
	return list, nil
}

func AddTypeContext(ctx context.Context, typeCode, typeName, remark string, status int) error {
	var err error
	if typeCode, err = validateTypeCode(typeCode); err != nil {
		return err
	}
	if typeName, remark, err = validateTypeFields(typeName, remark); err != nil {
		return err
	}
	if status, err = normalizeStatus(status); err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var count int64
	if err := db.Model(&model.SysDictType{}).Where("dict_type_code = ?", typeCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrTypeAlreadyExists
	}
	now := database.Now()
	if err := db.Create(&model.SysDictType{TypeCode: typeCode, TypeName: typeName, Status: status, Remark: remark, AddTime: now, EditTime: now}).Error; err != nil {
		return err
	}
	invalidateDictServiceCache()
	return nil
}

func EditTypeContext(ctx context.Context, typeCode, typeName, remark string, status int) error {
	var err error
	if typeCode, err = normalizeExistingTypeCode(typeCode); err != nil {
		return err
	}
	if typeName, remark, err = validateTypeFields(typeName, remark); err != nil {
		return err
	}
	if status, err = normalizeStatus(status); err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err = db.Transaction(func(tx *gorm.DB) error {
		var existing model.SysDictType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("dict_type_code = ?", typeCode).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTypeNotFound
			}
			return err
		}
		result := tx.Model(&model.SysDictType{}).Where("dict_type_code = ?", typeCode).Updates(map[string]interface{}{
			"dict_type_name": typeName, "dict_type_status": status, "dict_type_remark": remark, "dict_edit_time": database.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		if err := tx.Model(&model.SysDict{}).
			Where("dict_type_code = ?", typeCode).
			Where(legacyPlaceholderPredicateSQL).
			Updates(map[string]interface{}{"dict_label": typeName, "dict_type_name": typeName, "dict_edit_time": database.Now()}).Error; err != nil {
			return err
		}
		return tx.Model(&model.SysDict{}).Where("dict_type_code = ?", typeCode).Update("dict_type_name", typeName).Error
	})
	if err == nil {
		invalidateDictServiceCache()
	}
	return err
}

func DeleteTypeContext(ctx context.Context, typeCode string) error {
	var err error
	if typeCode, err = normalizeExistingTypeCode(typeCode); err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err = db.Transaction(func(tx *gorm.DB) error {
		var dictType model.SysDictType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("dict_type_code = ?", typeCode).First(&dictType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTypeNotFound
			}
			return err
		}
		if err := tx.Where("dict_type_code = ?", typeCode).Delete(&model.SysDict{}).Error; err != nil {
			return err
		}
		result := tx.Where("dict_type_code = ?", typeCode).Delete(&model.SysDictType{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrTypeNotFound
		}
		return nil
	})
	if err == nil {
		invalidateDictServiceCache()
	}
	return err
}

func AddItem(typeCode, typeName, label, value, remark string, sort int) error {
	return AddItemContext(context.Background(), typeCode, typeName, label, value, remark, sort)
}

func AddItemContext(ctx context.Context, typeCode, typeName, label, value, remark string, sort int) error {
	if strings.TrimSpace(value) == "" && strings.TrimSpace(label) == strings.TrimSpace(typeName) {
		return AddTypeContext(ctx, typeCode, typeName, remark, 1)
	}
	return AddItemWithStatusContext(ctx, typeCode, label, value, remark, sort, 1)
}

func AddItemWithStatusContext(ctx context.Context, typeCode, label, value, remark string, sort, status int) error {
	var err error
	if typeCode, err = normalizeExistingTypeCode(typeCode); err != nil {
		return err
	}
	if label, value, remark, err = validateItemFields(label, value, remark); err != nil {
		return err
	}
	if status, err = normalizeStatus(status); err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err = db.Transaction(func(tx *gorm.DB) error {
		var dictType model.SysDictType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("dict_type_code = ?", typeCode).First(&dictType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTypeNotFound
			}
			return err
		}
		var duplicate int64
		if err := tx.Model(&model.SysDict{}).Where("dict_type_code = ? AND dict_value = ?", typeCode, value).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return ErrItemAlreadyExists
		}
		now := database.Now()
		return tx.Create(&model.SysDict{TypeCode: typeCode, TypeName: dictType.TypeName, Label: label, Value: value, Sort: sort, Status: status, Remark: remark, AddTime: now, EditTime: now}).Error
	})
	if err == nil {
		invalidateDictServiceCache()
	}
	return err
}

func EditItem(id, label, value, remark string, sort int) error {
	return EditItemContext(context.Background(), id, label, value, remark, sort)
}

func EditItemContext(ctx context.Context, id, label, value, remark string, sort int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var item model.SysDict
	if err := db.Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrItemNotFound
		}
		return err
	}
	return EditItemWithStatusContext(ctx, id, label, value, remark, sort, item.Status)
}

func EditItemWithStatusContext(ctx context.Context, id, label, value, remark string, sort, status int) error {
	var err error
	if label, value, remark, err = validateItemFields(label, value, remark); err != nil {
		return err
	}
	if status, err = normalizeStatus(status); err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err = db.Transaction(func(tx *gorm.DB) error {
		var item model.SysDict
		if err := tx.Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrItemNotFound
			}
			return err
		}
		if isLegacyTypePlaceholder(item) {
			return ErrItemNotFound
		}
		var dictType model.SysDictType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("dict_type_code = ?", item.TypeCode).First(&dictType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTypeNotFound
			}
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrItemNotFound
			}
			return err
		}
		if isLegacyTypePlaceholder(item) {
			return ErrItemNotFound
		}
		if item.Value != value {
			var duplicate int64
			if err := tx.Model(&model.SysDict{}).Where("dict_type_code = ? AND dict_value = ? AND id <> ?", item.TypeCode, value, item.ID).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return ErrItemAlreadyExists
			}
		}
		result := tx.Model(&model.SysDict{}).Where("id = ?", id).Updates(map[string]interface{}{
			"dict_label": label, "dict_value": value, "dict_sort": sort, "dict_status": status, "dict_remark": remark, "dict_edit_time": database.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		return nil
	})
	if err == nil {
		invalidateDictServiceCache()
	}
	return err
}

func DeleteItem(id string) error {
	return DeleteItemContext(context.Background(), id)
}

func DeleteItemContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var item model.SysDict
	if err := db.Where("id = ?", id).First(&item).Error; err != nil || isLegacyTypePlaceholder(item) {
		if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrItemNotFound
		}
		return err
	}
	result := db.Where("id = ?", id).Delete(&model.SysDict{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	invalidateDictServiceCache()
	return nil
}

func DeleteByType(typeCode string) error {
	return DeleteByTypeContext(context.Background(), typeCode)
}

func DeleteByTypeContext(ctx context.Context, typeCode string) error {
	return ClearTypeItemsContext(ctx, typeCode)
}

func ClearTypeItemsContext(ctx context.Context, typeCode string) error {
	var err error
	if typeCode, err = normalizeExistingTypeCode(typeCode); err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	err = db.Transaction(func(tx *gorm.DB) error {
		var dictType model.SysDictType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("dict_type_code = ?", typeCode).First(&dictType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTypeNotFound
			}
			return err
		}
		return tx.Where("dict_type_code = ?", typeCode).Delete(&model.SysDict{}).Error
	})
	if err != nil {
		return err
	}
	invalidateDictServiceCache()
	return nil
}

func EditTypeName(oldTypeCode, typeCode, typeName string) error {
	return EditTypeNameContext(context.Background(), oldTypeCode, typeCode, typeName)
}

func EditTypeNameContext(ctx context.Context, oldTypeCode, typeCode, typeName string) error {
	oldTypeCode = strings.TrimSpace(oldTypeCode)
	typeCode = strings.TrimSpace(typeCode)
	if oldTypeCode != typeCode {
		return ErrTypeCodeImmutable
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var dictType model.SysDictType
	if err := db.Where("dict_type_code = ?", oldTypeCode).First(&dictType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTypeNotFound
		}
		return err
	}
	return EditTypeContext(ctx, oldTypeCode, typeName, dictType.Remark, dictType.Status)
}

func IsClientError(err error) bool {
	for _, target := range []error{
		ErrInvalidTypeCode, ErrInvalidStatus, ErrTypeNameRequired, ErrTypeAlreadyExists, ErrTypeNotFound,
		ErrTypeCodeImmutable, ErrItemRequired, ErrItemAlreadyExists, ErrItemNotFound,
	} {
		if errors.Is(err, target) {
			return true
		}
	}
	return err != nil && strings.HasPrefix(err.Error(), "字典")
}

func ClientErrorMessage(err error, fallback string) string {
	if IsClientError(err) {
		return err.Error()
	}
	return fallback
}
