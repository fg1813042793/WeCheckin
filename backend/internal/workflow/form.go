package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var ErrFormDataInvalid = errors.New("流程表单数据无效")

func ValidateFormData(fields []FormField, data map[string]interface{}, partial bool) error {
	fieldByKey := make(map[string]FormField, len(fields))
	for _, field := range fields {
		fieldByKey[field.Key] = field
	}
	for key, value := range data {
		field, ok := fieldByKey[key]
		if !ok {
			return fmt.Errorf("%w：字段 %s 未在流程表单中定义", ErrFormDataInvalid, key)
		}
		if err := validateFieldValue(field, value); err != nil {
			return err
		}
	}
	if partial {
		return nil
	}
	for _, field := range fields {
		if !field.Required {
			continue
		}
		value, ok := data[field.Key]
		if !ok || isEmptyFormValue(value) {
			return fmt.Errorf("%w：%s不能为空", ErrFormDataInvalid, field.Label)
		}
	}
	return nil
}

func ValidateNodeFormPatch(definition Definition, nodeID string, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	var node *Node
	for index := range definition.Nodes {
		if definition.Nodes[index].ID == nodeID {
			node = &definition.Nodes[index]
			break
		}
	}
	if node == nil || node.Type != NodeTypeApproval {
		return fmt.Errorf("%w：审批节点 %s 不存在", ErrFormDataInvalid, nodeID)
	}
	accessByField := make(map[string]string, len(node.FormPermissions))
	for _, permission := range node.FormPermissions {
		accessByField[permission.Field] = permission.Access
	}
	for field := range patch {
		if accessByField[field] != FieldAccessWrite {
			return fmt.Errorf("%w：当前节点无权修改字段 %s", ErrFormDataInvalid, field)
		}
	}
	return ValidateFormData(definition.Form, patch, true)
}

func MergeFormData(current, patch map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(current)+len(patch))
	for key, value := range current {
		result[key] = value
	}
	for key, value := range patch {
		result[key] = value
	}
	return result
}

func validateFieldValue(field FormField, value interface{}) error {
	if value == nil || isEmptyFormValue(value) {
		return nil
	}
	switch field.Type {
	case FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypeDate, FormFieldTypeDateTime,
		FormFieldTypeUser, FormFieldTypeDepartment, FormFieldTypeSelect:
		text, ok := value.(string)
		if !ok {
			return fieldTypeError(field)
		}
		if field.MaxLength > 0 && len([]rune(text)) > field.MaxLength {
			return fmt.Errorf("%w：%s长度不能超过%d", ErrFormDataInvalid, field.Label, field.MaxLength)
		}
		if field.Type == FormFieldTypeSelect && !optionContains(field.Options, text) {
			return fmt.Errorf("%w：%s选项无效", ErrFormDataInvalid, field.Label)
		}
	case FormFieldTypeNumber:
		number, ok := formNumber(value)
		if !ok {
			return fieldTypeError(field)
		}
		if field.Min != nil && number < *field.Min {
			return fmt.Errorf("%w：%s不能小于%v", ErrFormDataInvalid, field.Label, *field.Min)
		}
		if field.Max != nil && number > *field.Max {
			return fmt.Errorf("%w：%s不能大于%v", ErrFormDataInvalid, field.Label, *field.Max)
		}
	case FormFieldTypeMultiSelect, FormFieldTypeAttachment:
		values, ok := stringSlice(value)
		if !ok {
			return fieldTypeError(field)
		}
		if field.Type == FormFieldTypeMultiSelect {
			for _, item := range values {
				if !optionContains(field.Options, item) {
					return fmt.Errorf("%w：%s选项无效", ErrFormDataInvalid, field.Label)
				}
			}
		}
	case FormFieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fieldTypeError(field)
		}
	default:
		return fieldTypeError(field)
	}
	return nil
}

func fieldTypeError(field FormField) error {
	return fmt.Errorf("%w：%s类型无效", ErrFormDataInvalid, field.Label)
}

func optionContains(options []FormOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func isEmptyFormValue(value interface{}) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	reflected := reflect.ValueOf(value)
	return (reflected.Kind() == reflect.Slice || reflected.Kind() == reflect.Array || reflected.Kind() == reflect.Map) && reflected.Len() == 0
}

func formNumber(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func stringSlice(value interface{}) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return typed, true
	case []interface{}:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, false
			}
			result = append(result, text)
		}
		return result, true
	default:
		return nil, false
	}
}
