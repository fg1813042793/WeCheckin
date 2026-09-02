package workflowcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrFormDataInvalid = errors.New("流程表单数据无效")

var phonePattern = regexp.MustCompile(`^\+?[0-9][0-9 -]{5,19}$`)

func ValidateFormData(fields []FormField, data map[string]interface{}, partial bool) error {
	dataFields := dataFormFields(fields)
	fieldByKey := make(map[string]FormField, len(dataFields))
	for _, field := range dataFields {
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
		return validateSubmittedFormRules(dataFields, data, true)
	}
	for _, field := range dataFields {
		if !field.Required {
			continue
		}
		value, ok := data[field.Key]
		if !ok || isEmptyFormValue(value) {
			return fmt.Errorf("%w：%s不能为空", ErrFormDataInvalid, field.Label)
		}
	}
	return validateSubmittedFormRules(dataFields, data, false)
}

func ValidateNodeFormPatch(definition Definition, nodeID string, current, patch map[string]interface{}) error {
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
	if node == nil || (node.Type != NodeTypeApproval && node.Type != NodeTypeHandle) {
		return fmt.Errorf("%w：人工任务节点 %s 不存在", ErrFormDataInvalid, nodeID)
	}
	permissionByField := make(map[string]FieldPermission, len(node.FormPermissions))
	for _, permission := range node.FormPermissions {
		permissionByField[permission.Field] = permission
	}
	dataFields := dataFormFields(definition.Form)
	fieldByKey := make(map[string]FormField, len(dataFields))
	for _, field := range dataFields {
		fieldByKey[field.Key] = field
	}
	for field := range patch {
		permission, ok := permissionByField[field]
		if !ok || permission.Access != FieldAccessWrite {
			return fmt.Errorf("%w：当前节点无权修改字段 %s", ErrFormDataInvalid, field)
		}
		formField, ok := fieldByKey[field]
		if ok && formField.Type == FormFieldTypeDetailList {
			if err := validateDetailListPatchActions(formField, current[field], patch[field], permission); err != nil {
				return err
			}
		}
	}
	if err := ValidateFormData(definition.Form, patch, true); err != nil {
		return err
	}
	return ValidateFormData(definition.Form, MergeFormData(current, patch), false)
}

func dataFormFields(fields []FormField) []FormField {
	result := make([]FormField, 0, len(fields))
	for _, field := range fields {
		if field.Type == FormFieldTypeGroup {
			result = append(result, dataFormFields(field.Fields)...)
			continue
		}
		if isFormLayoutFieldType(field.Type) {
			continue
		}
		result = append(result, field)
	}
	return result
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
	if field.Type == FormFieldTypeDetailList {
		return validateDetailListValue(field, value)
	}
	if value == nil || isEmptyFormValue(value) {
		return nil
	}
	switch field.Type {
	case FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypeDate, FormFieldTypeDateTime,
		FormFieldTypeUser, FormFieldTypeDepartment, FormFieldTypeSelect, FormFieldTypePhone,
		FormFieldTypeEmail, FormFieldTypeRadio, FormFieldTypeTime:
		text, ok := value.(string)
		if !ok {
			return fieldTypeError(field)
		}
		if field.MaxLength > 0 && len([]rune(text)) > field.MaxLength {
			return fmt.Errorf("%w：%s长度不能超过%d", ErrFormDataInvalid, field.Label, field.MaxLength)
		}
		if (field.Type == FormFieldTypeSelect || field.Type == FormFieldTypeRadio) && !formFieldUsesAPIOptions(field) && !optionContains(field.Options, text) {
			return fmt.Errorf("%w：%s选项无效", ErrFormDataInvalid, field.Label)
		}
		if field.Type == FormFieldTypePhone && !phonePattern.MatchString(strings.TrimSpace(text)) {
			return fmt.Errorf("%w：%s格式无效", ErrFormDataInvalid, field.Label)
		}
		if field.Type == FormFieldTypeEmail {
			address, err := mail.ParseAddress(strings.TrimSpace(text))
			if err != nil || address.Address != strings.TrimSpace(text) {
				return fmt.Errorf("%w：%s格式无效", ErrFormDataInvalid, field.Label)
			}
		}
		if field.Type == FormFieldTypeTime && !validClockTime(text) {
			return fmt.Errorf("%w：%s格式无效", ErrFormDataInvalid, field.Label)
		}
	case FormFieldTypeNumber, FormFieldTypeAmount:
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
	case FormFieldTypeMultiSelect, FormFieldTypeAttachment, FormFieldTypeCheckbox,
		FormFieldTypeDateRange, FormFieldTypeUserMulti, FormFieldTypeDepartmentMulti:
		values, ok := stringSlice(value)
		if !ok {
			return fieldTypeError(field)
		}
		if (field.Type == FormFieldTypeMultiSelect || field.Type == FormFieldTypeCheckbox) && !formFieldUsesAPIOptions(field) {
			for _, item := range values {
				if !optionContains(field.Options, item) {
					return fmt.Errorf("%w：%s选项无效", ErrFormDataInvalid, field.Label)
				}
			}
		}
		if field.Type == FormFieldTypeDateRange && !validDateRange(values) {
			return fmt.Errorf("%w：%s日期区间无效", ErrFormDataInvalid, field.Label)
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

func validateSubmittedFormRules(fields []FormField, data map[string]interface{}, partial bool) error {
	for _, field := range fields {
		value, submitted := data[field.Key]
		if partial && !submitted {
			continue
		}
		for _, rule := range field.Rules {
			if rule.When != nil && !formRuleConditionMatches(*rule.When, data) {
				continue
			}
			if err := validateSubmittedFormRule(field, value, rule, data); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSubmittedFormRule(field FormField, value interface{}, rule FormValidationRule, data map[string]interface{}) error {
	failed := false
	message := strings.TrimSpace(rule.Message)
	switch rule.Type {
	case FormRuleConditionalRequired:
		failed = isEmptyFormValue(value)
		if message == "" {
			message = field.Label + "不能为空"
		}
	case FormRuleMinLength:
		if isEmptyFormValue(value) {
			return nil
		}
		text, ok := value.(string)
		failed = !ok || rule.Min == nil || float64(len([]rune(text))) < *rule.Min
		if message == "" && rule.Min != nil {
			message = fmt.Sprintf("%s长度不能少于%d", field.Label, int(*rule.Min))
		}
	case FormRuleMaxLength:
		if isEmptyFormValue(value) {
			return nil
		}
		text, ok := value.(string)
		failed = !ok || rule.Max == nil || float64(len([]rune(text))) > *rule.Max
		if message == "" && rule.Max != nil {
			message = fmt.Sprintf("%s长度不能超过%d", field.Label, int(*rule.Max))
		}
	case FormRulePattern:
		if isEmptyFormValue(value) {
			return nil
		}
		text, ok := value.(string)
		pattern, patternErr := regexp.Compile(rule.Pattern)
		failed = !ok || patternErr != nil || !pattern.MatchString(text)
		if message == "" {
			message = field.Label + "格式不正确"
		}
	case FormRuleNumberRange:
		if isEmptyFormValue(value) {
			return nil
		}
		number, ok := formNumber(value)
		if !ok {
			failed = true
		} else if rule.Min != nil && number < *rule.Min {
			failed = true
			if message == "" {
				message = fmt.Sprintf("%s不能小于%v", field.Label, *rule.Min)
			}
		} else if rule.Max != nil && number > *rule.Max {
			failed = true
			if message == "" {
				message = fmt.Sprintf("%s不能大于%v", field.Label, *rule.Max)
			}
		}
	case FormRuleDecimalPlaces:
		if isEmptyFormValue(value) {
			return nil
		}
		places, ok := formDecimalPlaces(value)
		failed = !ok || rule.Precision == nil || places > *rule.Precision
		if message == "" && rule.Precision != nil {
			message = fmt.Sprintf("%s小数位不能超过%d位", field.Label, *rule.Precision)
		}
	case FormRuleSelectionCount:
		values, ok := stringSlice(value)
		failed = !ok
		if ok && rule.Min != nil && len(values) < int(*rule.Min) {
			failed = true
			if message == "" {
				message = fmt.Sprintf("%s至少选择%d项", field.Label, int(*rule.Min))
			}
		} else if ok && rule.Max != nil && len(values) > int(*rule.Max) {
			failed = true
			if message == "" {
				message = fmt.Sprintf("%s最多选择%d项", field.Label, int(*rule.Max))
			}
		}
	case FormRuleCompareField:
		other := data[rule.Field]
		if isEmptyFormValue(value) || isEmptyFormValue(other) {
			return nil
		}
		failed = !formRuleOperatorMatches(value, other, rule.Operator)
		if message == "" {
			message = fmt.Sprintf("%s与%s的关系不符合要求", field.Label, rule.Field)
		}
	case FormRuleColumnSum:
		column, exists := detailListColumn(field, rule.Column)
		rows, rowsOK := detailListRows(value)
		sum := 0.0
		failed = !exists || !rowsOK || rule.Value == nil
		if !failed {
			for _, row := range rows {
				item, submitted := row[column.Key]
				if !submitted || isEmptyFormValue(item) {
					continue
				}
				number, ok := formNumber(item)
				if !ok {
					failed = true
					break
				}
				sum += number
			}
		}
		if !failed {
			failed = !formRuleNumberOperatorMatches(sum, *rule.Value, rule.Operator)
		}
		if message == "" && rule.Value != nil {
			columnLabel := rule.Column
			if exists {
				columnLabel = column.Label
			}
			message = fmt.Sprintf("%s“%s”列合计必须%s%v", field.Label, columnLabel, formRuleOperatorLabel(rule.Operator), *rule.Value)
		}
	default:
		failed = true
	}
	if !failed {
		return nil
	}
	if message == "" {
		message = field.Label + "校验不通过"
	}
	return fmt.Errorf("%w：%s", ErrFormDataInvalid, message)
}

func formRuleConditionMatches(condition FormRuleCondition, data map[string]interface{}) bool {
	value := data[condition.Field]
	switch condition.Operator {
	case FormRuleOperatorEmpty:
		return isEmptyFormValue(value)
	case FormRuleOperatorNotEmpty:
		return !isEmptyFormValue(value)
	default:
		return formRuleOperatorMatches(value, condition.Value, condition.Operator)
	}
}

func formRuleOperatorMatches(left, right interface{}, operator string) bool {
	comparison, comparable := compareFormRuleValues(left, right)
	switch operator {
	case FormRuleOperatorEQ:
		return comparable && comparison == 0
	case FormRuleOperatorNE:
		return !comparable || comparison != 0
	case FormRuleOperatorGT:
		return comparable && comparison > 0
	case FormRuleOperatorGTE:
		return comparable && comparison >= 0
	case FormRuleOperatorLT:
		return comparable && comparison < 0
	case FormRuleOperatorLTE:
		return comparable && comparison <= 0
	default:
		return false
	}
}

func formRuleNumberOperatorMatches(left, right float64, operator string) bool {
	if math.IsNaN(left) || math.IsNaN(right) || math.IsInf(left, 0) || math.IsInf(right, 0) {
		return false
	}
	tolerance := 1e-9 * math.Max(1, math.Max(math.Abs(left), math.Abs(right)))
	equal := math.Abs(left-right) <= tolerance
	switch operator {
	case FormRuleOperatorEQ:
		return equal
	case FormRuleOperatorNE:
		return !equal
	case FormRuleOperatorGT:
		return left > right && !equal
	case FormRuleOperatorGTE:
		return left > right || equal
	case FormRuleOperatorLT:
		return left < right && !equal
	case FormRuleOperatorLTE:
		return left < right || equal
	default:
		return false
	}
}

func formRuleOperatorLabel(operator string) string {
	switch operator {
	case FormRuleOperatorEQ:
		return "等于"
	case FormRuleOperatorNE:
		return "不等于"
	case FormRuleOperatorGT:
		return "大于"
	case FormRuleOperatorGTE:
		return "大于等于"
	case FormRuleOperatorLT:
		return "小于"
	case FormRuleOperatorLTE:
		return "小于等于"
	default:
		return ""
	}
}

func compareFormRuleValues(left, right interface{}) (int, bool) {
	if leftNumber, leftOK := formNumber(left); leftOK {
		if rightNumber, rightOK := formNumber(right); rightOK {
			switch {
			case leftNumber < rightNumber:
				return -1, true
			case leftNumber > rightNumber:
				return 1, true
			default:
				return 0, true
			}
		}
	}
	leftText, leftOK := left.(string)
	rightText, rightOK := right.(string)
	if leftOK && rightOK {
		switch {
		case leftText < rightText:
			return -1, true
		case leftText > rightText:
			return 1, true
		default:
			return 0, true
		}
	}
	leftBool, leftOK := left.(bool)
	rightBool, rightOK := right.(bool)
	if leftOK && rightOK {
		if leftBool == rightBool {
			return 0, true
		}
		if !leftBool && rightBool {
			return -1, true
		}
		return 1, true
	}
	return 0, false
}

func formDecimalPlaces(value interface{}) (int, bool) {
	number, ok := formNumber(value)
	if !ok {
		return 0, false
	}
	text := strconv.FormatFloat(number, 'f', -1, 64)
	separator := strings.IndexByte(text, '.')
	if separator < 0 {
		return 0, true
	}
	return len(strings.TrimRight(text[separator+1:], "0")), true
}

func validateDetailListValue(field FormField, value interface{}) error {
	rows, ok := detailListRows(value)
	if !ok {
		return fieldTypeError(field)
	}
	if field.MinRows > 0 && len(rows) < field.MinRows {
		return fmt.Errorf("%w：%s至少需要%d行", ErrFormDataInvalid, field.Label, field.MinRows)
	}
	if field.MaxRows > 0 && len(rows) > field.MaxRows {
		return fmt.Errorf("%w：%s最多允许%d行", ErrFormDataInvalid, field.Label, field.MaxRows)
	}
	rowKey := detailListRowKey(field)
	columnByKey := make(map[string]FormField, len(field.Columns))
	for _, column := range field.Columns {
		columnByKey[column.Key] = column
	}
	seenRows := make(map[string]struct{}, len(rows))
	for index, row := range rows {
		rowID, ok := detailRowID(row, rowKey)
		if !ok {
			return fmt.Errorf("%w：%s第%d行缺少行标识", ErrFormDataInvalid, field.Label, index+1)
		}
		if _, exists := seenRows[rowID]; exists {
			return fmt.Errorf("%w：%s第%d行行标识重复", ErrFormDataInvalid, field.Label, index+1)
		}
		seenRows[rowID] = struct{}{}
		rowData := make(map[string]interface{}, len(row))
		for key, item := range row {
			if key == rowKey {
				continue
			}
			if _, ok := columnByKey[key]; !ok {
				return fmt.Errorf("%w：%s第%d行字段 %s 未定义", ErrFormDataInvalid, field.Label, index+1, key)
			}
			rowData[key] = item
		}
		if err := ValidateFormData(field.Columns, rowData, false); err != nil {
			return err
		}
	}
	return nil
}

func validateDetailListPatchActions(field FormField, currentValue, nextValue interface{}, permission FieldPermission) error {
	currentRows, ok := detailListRows(currentValue)
	if !ok {
		return fieldTypeError(field)
	}
	nextRows, ok := detailListRows(nextValue)
	if !ok {
		return fieldTypeError(field)
	}
	rowKey := detailListRowKey(field)
	currentIDs, ok := detailRowIDSet(currentRows, rowKey)
	if !ok {
		return fieldTypeError(field)
	}
	nextIDs, ok := detailRowIDSet(nextRows, rowKey)
	if !ok {
		return fieldTypeError(field)
	}
	actions := fieldActionSet(permission.Actions)
	for id := range nextIDs {
		if _, exists := currentIDs[id]; !exists && !actions[FieldActionAdd] {
			return fmt.Errorf("%w：当前节点无权新增%s明细行", ErrFormDataInvalid, field.Label)
		}
	}
	for id := range currentIDs {
		if _, exists := nextIDs[id]; !exists && !actions[FieldActionDelete] {
			return fmt.Errorf("%w：当前节点无权删除%s明细行", ErrFormDataInvalid, field.Label)
		}
	}
	return nil
}

func detailListRows(value interface{}) ([]map[string]interface{}, bool) {
	if value == nil {
		return []map[string]interface{}{}, true
	}
	switch typed := value.(type) {
	case []map[string]interface{}:
		result := make([]map[string]interface{}, 0, len(typed))
		for _, row := range typed {
			if row == nil {
				return nil, false
			}
			result = append(result, row)
		}
		return result, true
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(typed))
		for _, item := range typed {
			row, ok := item.(map[string]interface{})
			if !ok || row == nil {
				return nil, false
			}
			result = append(result, row)
		}
		return result, true
	default:
		return nil, false
	}
}

func detailListRowKey(field FormField) string {
	rowKey := strings.TrimSpace(field.RowKey)
	if rowKey == "" {
		return "id"
	}
	return rowKey
}

func detailRowID(row map[string]interface{}, rowKey string) (string, bool) {
	value, ok := row[rowKey]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.TrimSpace(text)
	return text, text != ""
}

func detailRowIDSet(rows []map[string]interface{}, rowKey string) (map[string]struct{}, bool) {
	result := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		rowID, ok := detailRowID(row, rowKey)
		if !ok {
			return nil, false
		}
		result[rowID] = struct{}{}
	}
	return result, true
}

func fieldActionSet(actions []string) map[string]bool {
	result := make(map[string]bool, len(actions))
	for _, action := range actions {
		result[strings.TrimSpace(action)] = true
	}
	return result
}

func validClockTime(value string) bool {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"15:04", "15:04:05"} {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}
	return false
}

func validDateRange(values []string) bool {
	if len(values) != 2 {
		return false
	}
	start, startErr := time.Parse("2006-01-02", strings.TrimSpace(values[0]))
	end, endErr := time.Parse("2006-01-02", strings.TrimSpace(values[1]))
	return startErr == nil && endErr == nil && !start.After(end)
}

func fieldTypeError(field FormField) error {
	return fmt.Errorf("%w：%s类型无效", ErrFormDataInvalid, field.Label)
}

func optionContains(options []FormOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
		if optionContains(option.Children, value) {
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
