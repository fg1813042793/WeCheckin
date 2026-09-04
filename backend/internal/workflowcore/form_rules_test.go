package workflowcore

import (
	"strings"
	"testing"
)

func TestValidateFormDataSupportsStructuredRules(t *testing.T) {
	fields := []FormField{
		{Key: "expenseType", Label: "报销类型", Type: FormFieldTypeSelect, Options: []FormOption{{Label: "差旅", Value: "travel"}, {Label: "其他", Value: "other"}}},
		{Key: "reason", Label: "具体说明", Type: FormFieldTypeTextarea, Rules: []FormValidationRule{{ID: "reason_required", Type: FormRuleConditionalRequired, When: &FormRuleCondition{Field: "expenseType", Operator: FormRuleOperatorEQ, Value: "other"}, Message: "选择其他类型时，请填写具体说明"}}},
		{Key: "code", Label: "单据编号", Type: FormFieldTypeText, Rules: []FormValidationRule{{ID: "code_length", Type: FormRuleMinLength, Min: numberPointer(6)}, {ID: "code_pattern", Type: FormRulePattern, Pattern: `^[A-Z]{2}[0-9]{4}$`, Message: "单据编号格式不正确"}}},
		{Key: "amount", Label: "报销金额", Type: FormFieldTypeAmount, Rules: []FormValidationRule{{ID: "amount_range", Type: FormRuleNumberRange, Min: numberPointer(100), Max: numberPointer(5000)}, {ID: "amount_scale", Type: FormRuleDecimalPlaces, Precision: intPointer(2)}}},
		{Key: "attachments", Label: "附件", Type: FormFieldTypeAttachment, Rules: []FormValidationRule{{ID: "attachment_count", Type: FormRuleSelectionCount, Min: numberPointer(1), Max: numberPointer(3)}}},
		{Key: "startDate", Label: "开始日期", Type: FormFieldTypeDate},
		{Key: "endDate", Label: "结束日期", Type: FormFieldTypeDate, Rules: []FormValidationRule{{ID: "date_order", Type: FormRuleCompareField, Field: "startDate", Operator: FormRuleOperatorGTE, Message: "结束日期不能早于开始日期"}}},
	}
	valid := map[string]interface{}{
		"expenseType": "other", "reason": "临时采购", "code": "BX1024", "amount": 1280.5,
		"attachments": []interface{}{"receipt.pdf"}, "startDate": "2026-09-01", "endDate": "2026-09-03",
	}
	if err := ValidateFormData(fields, valid, false); err != nil {
		t.Fatalf("valid structured form rules rejected: %v", err)
	}

	tests := []struct {
		name       string
		field      string
		value      interface{}
		expectText string
	}{
		{name: "conditional required", field: "reason", value: "", expectText: "选择其他类型时，请填写具体说明"},
		{name: "minimum length", field: "code", value: "A1", expectText: "长度不能少于6"},
		{name: "pattern", field: "code", value: "bx1024", expectText: "单据编号格式不正确"},
		{name: "number range", field: "amount", value: 99, expectText: "不能小于100"},
		{name: "decimal places", field: "amount", value: 100.123, expectText: "小数位不能超过2位"},
		{name: "selection count", field: "attachments", value: []interface{}{}, expectText: "至少选择1项"},
		{name: "compare field", field: "endDate", value: "2026-08-31", expectText: "结束日期不能早于开始日期"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := cloneFormData(valid)
			data[test.field] = test.value
			err := ValidateFormData(fields, data, false)
			if err == nil || !strings.Contains(err.Error(), test.expectText) {
				t.Fatalf("expected error containing %q, got %v", test.expectText, err)
			}
		})
	}

	travel := cloneFormData(valid)
	travel["expenseType"] = "travel"
	travel["reason"] = ""
	if err := ValidateFormData(fields, travel, false); err != nil {
		t.Fatalf("inactive conditional rule must not reject data: %v", err)
	}
}

func TestValidateFormDataSupportsDetailColumnSumRule(t *testing.T) {
	field := detailColumnSumField()
	rows := []interface{}{
		map[string]interface{}{"id": "row_1", "weight": 40, "budget": 0.1, "target": "完成方案"},
		map[string]interface{}{"id": "row_2", "weight": 60, "budget": 0.2, "target": "完成上线"},
		map[string]interface{}{"id": "row_3", "weight": "", "budget": nil, "target": "持续优化"},
	}
	operators := []struct {
		name     string
		operator string
		target   float64
	}{
		{name: "equal", operator: FormRuleOperatorEQ, target: 100},
		{name: "not equal", operator: FormRuleOperatorNE, target: 99},
		{name: "greater than", operator: FormRuleOperatorGT, target: 99},
		{name: "greater than or equal", operator: FormRuleOperatorGTE, target: 100},
		{name: "less than", operator: FormRuleOperatorLT, target: 101},
		{name: "less than or equal", operator: FormRuleOperatorLTE, target: 100},
	}
	for _, test := range operators {
		t.Run(test.name, func(t *testing.T) {
			current := field
			current.Rules = []FormValidationRule{{
				ID: "weight_sum", Type: FormRuleColumnSum, Column: "weight",
				Operator: test.operator, Value: numberPointer(test.target),
			}}
			if err := ValidateFormData([]FormField{current}, map[string]interface{}{"objectives": rows}, false); err != nil {
				t.Fatalf("valid detail column sum rejected: %v", err)
			}
		})
	}

	amount := field
	amount.Rules = []FormValidationRule{{
		ID: "budget_sum", Type: FormRuleColumnSum, Column: "budget",
		Operator: FormRuleOperatorEQ, Value: numberPointer(0.3),
	}}
	if err := ValidateFormData([]FormField{amount}, map[string]interface{}{"objectives": rows}, false); err != nil {
		t.Fatalf("floating point detail column sum rejected: %v", err)
	}

	invalid := field
	invalid.Rules = []FormValidationRule{{
		ID: "weight_sum", Type: FormRuleColumnSum, Column: "weight",
		Operator: FormRuleOperatorEQ, Value: numberPointer(99), Message: "权重合计必须为99",
	}}
	err := ValidateFormData([]FormField{invalid}, map[string]interface{}{"objectives": rows}, false)
	if err == nil || !strings.Contains(err.Error(), "权重合计必须为99") {
		t.Fatalf("expected custom column sum error, got %v", err)
	}

	invalid.Rules[0].Message = ""
	err = ValidateFormData([]FormField{invalid}, map[string]interface{}{"objectives": rows}, false)
	if err == nil || !strings.Contains(err.Error(), "我的目标“权重”列合计必须等于99") {
		t.Fatalf("expected default column sum error, got %v", err)
	}

	required := field
	required.Columns[0].Required = true
	required.Rules = []FormValidationRule{{
		ID: "weight_sum", Type: FormRuleColumnSum, Column: "weight",
		Operator: FormRuleOperatorEQ, Value: numberPointer(100), Message: "不应优先显示的合计错误",
	}}
	err = ValidateFormData([]FormField{required}, map[string]interface{}{"objectives": rows}, false)
	if err == nil || !strings.Contains(err.Error(), "权重不能为空") || strings.Contains(err.Error(), "不应优先显示的合计错误") {
		t.Fatalf("expected required column error before column sum error, got %v", err)
	}
}

func TestValidateDefinitionRejectsInvalidStructuredRules(t *testing.T) {
	validDefinition := validLinearDefinition()
	validDefinition.Form = []FormField{detailColumnSumFieldWithRule(FormValidationRule{
		ID: "sum", Type: FormRuleColumnSum, Column: "weight",
		Operator: FormRuleOperatorEQ, Value: numberPointer(100),
	})}
	if errors := ValidateDefinition(validDefinition); hasValidationCode(errors, ValidationFormFieldRules) {
		t.Fatalf("valid detail column sum rule rejected: %#v", errors)
	}

	tests := []struct {
		name  string
		field FormField
	}{
		{name: "duplicate rule id", field: FormField{Key: "code", Label: "编号", Type: FormFieldTypeText, Rules: []FormValidationRule{{ID: "same", Type: FormRuleMinLength, Min: numberPointer(2)}, {ID: "same", Type: FormRuleMaxLength, Max: numberPointer(8)}}}},
		{name: "invalid pattern", field: FormField{Key: "code", Label: "编号", Type: FormFieldTypeText, Rules: []FormValidationRule{{ID: "pattern", Type: FormRulePattern, Pattern: "["}}}},
		{name: "unknown compare field", field: FormField{Key: "endDate", Label: "结束日期", Type: FormFieldTypeDate, Rules: []FormValidationRule{{ID: "compare", Type: FormRuleCompareField, Field: "missing", Operator: FormRuleOperatorGTE}}}},
		{name: "invalid selection count type", field: FormField{Key: "reason", Label: "原因", Type: FormFieldTypeText, Rules: []FormValidationRule{{ID: "count", Type: FormRuleSelectionCount, Min: numberPointer(1)}}}},
		{name: "invalid decimal precision", field: FormField{Key: "amount", Label: "金额", Type: FormFieldTypeAmount, Rules: []FormValidationRule{{ID: "scale", Type: FormRuleDecimalPlaces, Precision: intPointer(11)}}}},
		{name: "column sum on scalar field", field: FormField{Key: "amount", Label: "金额", Type: FormFieldTypeAmount, Rules: []FormValidationRule{{ID: "sum", Type: FormRuleColumnSum, Column: "amount", Operator: FormRuleOperatorEQ, Value: numberPointer(100)}}}},
		{name: "column sum missing column", field: detailColumnSumFieldWithRule(FormValidationRule{ID: "sum", Type: FormRuleColumnSum, Column: "missing", Operator: FormRuleOperatorEQ, Value: numberPointer(100)})},
		{name: "column sum text column", field: detailColumnSumFieldWithRule(FormValidationRule{ID: "sum", Type: FormRuleColumnSum, Column: "target", Operator: FormRuleOperatorEQ, Value: numberPointer(100)})},
		{name: "column sum invalid operator", field: detailColumnSumFieldWithRule(FormValidationRule{ID: "sum", Type: FormRuleColumnSum, Column: "weight", Operator: "contains", Value: numberPointer(100)})},
		{name: "column sum missing target", field: detailColumnSumFieldWithRule(FormValidationRule{ID: "sum", Type: FormRuleColumnSum, Column: "weight", Operator: FormRuleOperatorEQ})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Form = []FormField{test.field}
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationFormFieldRules) {
				t.Fatalf("expected %s, got %#v", ValidationFormFieldRules, errors)
			}
		})
	}
}

func TestValidateDefinitionSupportsCompatibleCrossFieldRules(t *testing.T) {
	tests := []struct {
		name     string
		owner    FormField
		target   FormField
		operator string
	}{
		{
			name:     "number and amount ordering",
			owner:    FormField{Key: "actual", Label: "实际金额", Type: FormFieldTypeAmount},
			target:   FormField{Key: "budget", Label: "预算", Type: FormFieldTypeNumber},
			operator: FormRuleOperatorLTE,
		},
		{
			name:     "text equality",
			owner:    FormField{Key: "confirmEmail", Label: "确认邮箱", Type: FormFieldTypeText},
			target:   FormField{Key: "email", Label: "邮箱", Type: FormFieldTypeEmail},
			operator: FormRuleOperatorEQ,
		},
		{
			name:     "select and radio equality",
			owner:    FormField{Key: "result", Label: "复核结果", Type: FormFieldTypeSelect, Options: crossFieldOptions()},
			target:   FormField{Key: "grade", Label: "上级分档", Type: FormFieldTypeRadio, Options: crossFieldOptions()},
			operator: FormRuleOperatorEQ,
		},
		{
			name:     "boolean inequality",
			owner:    FormField{Key: "confirmed", Label: "已确认", Type: FormFieldTypeBoolean},
			target:   FormField{Key: "approved", Label: "已同意", Type: FormFieldTypeBoolean},
			operator: FormRuleOperatorNE,
		},
		{
			name:     "user equality",
			owner:    FormField{Key: "reviewer", Label: "复核人", Type: FormFieldTypeUser},
			target:   FormField{Key: "applicant", Label: "申请人", Type: FormFieldTypeUser},
			operator: FormRuleOperatorNE,
		},
		{
			name:     "date ordering",
			owner:    FormField{Key: "endDate", Label: "结束日期", Type: FormFieldTypeDate},
			target:   FormField{Key: "startDate", Label: "开始日期", Type: FormFieldTypeDate},
			operator: FormRuleOperatorGTE,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			test.owner.Rules = []FormValidationRule{{
				ID: "cross_field", Type: FormRuleCompareField,
				Field: test.target.Key, Operator: test.operator,
			}}
			definition.Form = []FormField{test.target, test.owner}
			if errors := ValidateDefinition(definition); hasValidationCode(errors, ValidationFormFieldRules) {
				t.Fatalf("compatible cross-field rule rejected: %#v", errors)
			}
		})
	}
}

func TestValidateDefinitionRejectsIncompatibleCrossFieldRules(t *testing.T) {
	tests := []struct {
		name     string
		owner    FormField
		target   FormField
		operator string
	}{
		{
			name:     "select ordering",
			owner:    FormField{Key: "result", Label: "复核结果", Type: FormFieldTypeSelect, Options: crossFieldOptions()},
			target:   FormField{Key: "grade", Label: "上级分档", Type: FormFieldTypeRadio, Options: crossFieldOptions()},
			operator: FormRuleOperatorGT,
		},
		{
			name:     "user and department",
			owner:    FormField{Key: "reviewer", Label: "复核人", Type: FormFieldTypeUser},
			target:   FormField{Key: "department", Label: "部门", Type: FormFieldTypeDepartment},
			operator: FormRuleOperatorEQ,
		},
		{
			name:     "select and text",
			owner:    FormField{Key: "result", Label: "复核结果", Type: FormFieldTypeSelect, Options: crossFieldOptions()},
			target:   FormField{Key: "remark", Label: "备注", Type: FormFieldTypeText},
			operator: FormRuleOperatorEQ,
		},
		{
			name:     "date and datetime",
			owner:    FormField{Key: "endDate", Label: "结束日期", Type: FormFieldTypeDate},
			target:   FormField{Key: "startAt", Label: "开始时间", Type: FormFieldTypeDateTime},
			operator: FormRuleOperatorGTE,
		},
		{
			name:     "attachments",
			owner:    FormField{Key: "receipts", Label: "票据", Type: FormFieldTypeAttachment},
			target:   FormField{Key: "contracts", Label: "合同", Type: FormFieldTypeAttachment},
			operator: FormRuleOperatorEQ,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			test.owner.Rules = []FormValidationRule{{
				ID: "cross_field", Type: FormRuleCompareField,
				Field: test.target.Key, Operator: test.operator,
			}}
			definition.Form = []FormField{test.target, test.owner}
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationFormFieldRules) {
				t.Fatalf("incompatible cross-field rule should be rejected: %#v", errors)
			}
		})
	}
}

func TestValidateFormDataComparesChoiceValuesAsStrings(t *testing.T) {
	fields := []FormField{
		{Key: "grade", Label: "上级分档", Type: FormFieldTypeRadio, Options: []FormOption{{Label: "01", Value: "01"}, {Label: "1", Value: "1"}}},
		{
			Key: "result", Label: "复核结果", Type: FormFieldTypeSelect,
			Options: []FormOption{{Label: "01", Value: "01"}, {Label: "1", Value: "1"}},
			Rules: []FormValidationRule{{
				ID: "same_grade", Type: FormRuleCompareField, Field: "grade", Operator: FormRuleOperatorEQ,
			}},
		},
	}

	if err := ValidateFormData(fields, map[string]interface{}{"grade": "1", "result": "01"}, false); err == nil {
		t.Fatal("choice values 01 and 1 must not compare as equal numbers")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"grade": "01", "result": "01"}, false); err != nil {
		t.Fatalf("identical choice values should compare as equal: %v", err)
	}
}

func crossFieldOptions() []FormOption {
	return []FormOption{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}}
}

func detailColumnSumField() FormField {
	return FormField{
		Key: "objectives", Label: "我的目标", Type: FormFieldTypeDetailList, RowKey: "id",
		Columns: []FormField{
			{Key: "weight", Label: "权重", Type: FormFieldTypeNumber},
			{Key: "budget", Label: "预算", Type: FormFieldTypeAmount},
			{Key: "target", Label: "目标", Type: FormFieldTypeText},
		},
	}
}

func detailColumnSumFieldWithRule(rule FormValidationRule) FormField {
	field := detailColumnSumField()
	field.Rules = []FormValidationRule{rule}
	return field
}

func cloneFormData(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func intPointer(value int) *int { return &value }
