package workflowcore

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func calculationFields() []FormField {
	precision := 2
	return []FormField{
		{Key: "quantity", Label: "数量", Type: FormFieldTypeNumber},
		{Key: "price", Label: "单价", Type: FormFieldTypeAmount},
		{Key: "discount", Label: "优惠", Type: FormFieldTypeAmount},
		{
			Key: "items", Label: "费用明细", Type: FormFieldTypeDetailList, RowKey: "id",
			Columns: []FormField{
				{Key: "quantity", Label: "数量", Type: FormFieldTypeNumber},
				{Key: "price", Label: "单价", Type: FormFieldTypeAmount},
				{Key: "remark", Label: "备注", Type: FormFieldTypeText},
			},
		},
		{
			Key: "total", Label: "合计", Type: FormFieldTypeCalculation,
			Calculation: &FormCalculation{Expression: "[quantity] * [price] - [discount]", Display: CalculationDisplayField, Precision: &precision},
		},
		{
			Key: "detailTotal", Label: "明细合计", Type: FormFieldTypeCalculation,
			Calculation: &FormCalculation{Expression: "SUM([items.quantity] * [items.price])", Display: CalculationDisplayLabel, Precision: &precision},
		},
	}
}

func TestApplyFormCalculationsSupportsScalarAndDetailRowExpressions(t *testing.T) {
	precision := 2
	fields := append(calculationFields(), FormField{
		Key: "negative", Label: "负数舍入", Type: FormFieldTypeCalculation,
		Calculation: &FormCalculation{Expression: "-1.005", Display: CalculationDisplayField, Precision: &precision},
	})
	data := map[string]interface{}{
		"quantity": 3,
		"price":    12.345,
		"discount": 2,
		"items": []interface{}{
			map[string]interface{}{"id": "row-1", "quantity": 2, "price": 10},
			map[string]interface{}{"id": "row-2", "quantity": 3, "price": 4.5},
			map[string]interface{}{"id": "row-3", "quantity": "", "price": nil},
		},
		"total":       999,
		"detailTotal": -1,
	}

	calculated, err := ApplyFormCalculations(fields, data)
	if err != nil {
		t.Fatalf("ApplyFormCalculations() error = %v", err)
	}
	if calculated["total"] != 35.04 {
		t.Fatalf("scalar calculation = %#v, want 35.04", calculated["total"])
	}
	if calculated["detailTotal"] != 33.5 {
		t.Fatalf("detail calculation = %#v, want 33.5", calculated["detailTotal"])
	}
	if calculated["negative"] != -1.01 {
		t.Fatalf("negative calculation = %#v, want -1.01", calculated["negative"])
	}
	if data["total"] != 999 {
		t.Fatalf("source data was mutated: %#v", data)
	}
}

func TestApplyFormCalculationsSupportsAggregateComposition(t *testing.T) {
	precision := 0
	fields := calculationFields()
	fields = append(fields,
		FormField{Key: "average", Label: "平均值", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "AVG([items.quantity] + [items.price])", Display: CalculationDisplayField, Precision: &precision}},
		FormField{Key: "range", Label: "极差", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "MAX([items.price]) - MIN([items.price])", Display: CalculationDisplayField, Precision: &precision}},
		FormField{Key: "rows", Label: "行数", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "COUNT([items.quantity])", Display: CalculationDisplayLabel, Precision: &precision}},
	)
	data := map[string]interface{}{
		"quantity": 0,
		"price":    0,
		"discount": 0,
		"items": []interface{}{
			map[string]interface{}{"id": "row-1", "quantity": 2, "price": 10},
			map[string]interface{}{"id": "row-2", "quantity": 4, "price": 20},
		},
	}

	calculated, err := ApplyFormCalculations(fields, data)
	if err != nil {
		t.Fatalf("ApplyFormCalculations() error = %v", err)
	}
	if calculated["average"] != float64(18) || calculated["range"] != float64(10) || calculated["rows"] != float64(2) {
		t.Fatalf("aggregate calculations = average:%#v range:%#v rows:%#v", calculated["average"], calculated["range"], calculated["rows"])
	}
}

func TestValidateDefinitionRejectsInvalidCalculationReferences(t *testing.T) {
	tests := []struct {
		name       string
		expression string
	}{
		{name: "unknown field", expression: "[missing] + 1"},
		{name: "text field", expression: "[items.remark] + 1"},
		{name: "detail outside aggregate", expression: "[items.price] * 2"},
		{name: "mixed detail lists", expression: "SUM([items.price] + [other.amount])"},
		{name: "calculation dependency", expression: "[total] + 1"},
		{name: "division by malformed expression", expression: "[price] /"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := calculationFields()
			fields = append(fields, FormField{
				Key: "other", Label: "其他明细", Type: FormFieldTypeDetailList,
				Columns: []FormField{{Key: "amount", Label: "金额", Type: FormFieldTypeAmount}},
			})
			fields[len(fields)-3].Calculation.Expression = test.expression
			definition := validLinearDefinition()
			definition.Form = fields
			if validationErrors := ValidateDefinition(definition); !hasValidationCode(validationErrors, ValidationFormFieldCalculation) {
				t.Fatalf("ValidateDefinition() errors = %#v, want %s", validationErrors, ValidationFormFieldCalculation)
			}
		})
	}
}

func TestValidateDefinitionRejectsInvalidCalculationConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]FormField)
	}{
		{name: "calculation placeholder", mutate: func(fields []FormField) { fields[4].Placeholder = "请输入" }},
		{name: "calculation in detail", mutate: func(fields []FormField) {
			fields[3].Columns = append(fields[3].Columns, fields[4])
		}},
		{name: "calculation attached to normal field", mutate: func(fields []FormField) {
			fields[0].Calculation = fields[4].Calculation
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := calculationFields()
			test.mutate(fields)
			definition := validLinearDefinition()
			definition.Form = fields
			if validationErrors := ValidateDefinition(definition); !hasValidationCode(validationErrors, ValidationFormFieldCalculation) {
				t.Fatalf("ValidateDefinition() errors = %#v, want %s", validationErrors, ValidationFormFieldCalculation)
			}
		})
	}
}

func TestCalculationFieldsAreServerOwned(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = calculationFields()
	definition.Nodes[0].FormPermissions = []FieldPermission{{Field: "total", Access: FieldAccessWrite}}
	if validationErrors := ValidateDefinition(definition); !hasValidationCode(validationErrors, ValidationFieldPermissionAccess) {
		t.Fatalf("write permission errors = %#v", validationErrors)
	}

	definition.Nodes[0].FormPermissions = nil
	input := map[string]interface{}{
		"quantity": 2,
		"price":    10,
		"discount": 1,
		"items":    []interface{}{},
	}
	if err := ValidateStartFormData(definition, input); err != nil {
		t.Fatalf("calculation omitted from client payload: %v", err)
	}
	forged := MergeFormData(input, map[string]interface{}{"total": 999})
	if err := ValidateStartFormData(definition, forged); !errors.Is(err, ErrFormDataInvalid) {
		t.Fatalf("forged calculation error = %v, want ErrFormDataInvalid", err)
	}
}

func TestApplyFormCalculationsSupportsSelectOptionCalculationValues(t *testing.T) {
	fields := []FormField{
		{Key: "levelA", Label: "等级A", Type: FormFieldTypeSelect, Options: []FormOption{
			{Label: "A组", Value: "group_a", CalculationValue: numberPointer(0), Children: []FormOption{
				{Label: "A+", Value: "a_plus", CalculationValue: numberPointer(1.5)},
			}},
			{Label: "A", Value: "a", CalculationValue: numberPointer(1)},
		}},
		{Key: "levelB", Label: "等级B", Type: FormFieldTypeRadio, Options: []FormOption{
			{Label: "B+", Value: "b_plus", CalculationValue: numberPointer(0.8)},
			{Label: "B", Value: "b", CalculationValue: numberPointer(0.6)},
		}},
		{Key: "total", Label: "综合等级", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "[levelA] + [levelB]"}},
	}
	calculated, err := ApplyFormCalculations(fields, map[string]interface{}{"levelA": "a_plus", "levelB": "b_plus"})
	if err != nil {
		t.Fatalf("select calculation error = %v", err)
	}
	if calculated["total"] != 2.3 {
		t.Fatalf("select calculation = %#v, want 2.3", calculated["total"])
	}
}

func TestApplyFormCalculationsKeepsNumericSelectValueCompatibility(t *testing.T) {
	fields := []FormField{
		{Key: "level", Label: "等级", Type: FormFieldTypeSelect, Options: []FormOption{{Label: "A+", Value: "1.5"}}},
		{Key: "total", Label: "合计", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "[level] * 2"}},
	}
	calculated, err := ApplyFormCalculations(fields, map[string]interface{}{"level": "1.5"})
	if err != nil || calculated["total"] != float64(3) {
		t.Fatalf("legacy numeric select calculation = %#v, error = %v", calculated, err)
	}
}

func TestFormOptionSourceKeepsCalculationValueFieldMapping(t *testing.T) {
	sourceJSON := []byte(`{"type":"api","url":"/api/v2/dict/items","calculationValueField":"scoreValue"}`)
	var source FormOptionSource
	if err := json.Unmarshal(sourceJSON, &source); err != nil {
		t.Fatalf("unmarshal option source: %v", err)
	}
	if source.CalculationValueField != "scoreValue" {
		t.Fatalf("calculationValueField = %q", source.CalculationValueField)
	}
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("marshal option source: %v", err)
	}
	if !strings.Contains(string(encoded), `"calculationValueField":"scoreValue"`) {
		t.Fatalf("encoded option source = %s", encoded)
	}
}

func TestValidateFormCalculationRejectsNonnumericSelectOption(t *testing.T) {
	fields := []FormField{
		{Key: "level", Label: "等级", Type: FormFieldTypeSelect, Options: []FormOption{{Label: "优秀", Value: "excellent"}}},
		{Key: "total", Label: "合计", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "[level] + 1"}},
	}
	if err := validateFormCalculation(fields[1], buildCalculationSchema(fields)); err == nil {
		t.Fatal("nonnumeric select option must not be accepted as a calculation field")
	}
}

func TestWorkflowCalculationResultLabelUsesFirstMatchingRule(t *testing.T) {
	calculation := &FormCalculation{
		ResultDisplay: &FormCalculationResultDisplay{
			Mode: "label",
			Rules: []FormCalculationResultRule{
				{Min: numberPointer(2.3), Label: "优秀"},
				{Min: numberPointer(1.5), Max: numberPointer(2.3), Label: "良好"},
			},
			Fallback: "未评级",
		},
	}
	if got := WorkflowCalculationResultLabel(2.3, calculation); got != "优秀" {
		t.Fatalf("result label = %q, want 优秀", got)
	}
	if got := WorkflowCalculationResultLabel(1.8, calculation); got != "良好" {
		t.Fatalf("result label = %q, want 良好", got)
	}
	if got := WorkflowCalculationResultLabel(0.5, calculation); got != "未评级" {
		t.Fatalf("fallback result label = %q, want 未评级", got)
	}
}
