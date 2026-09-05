package infrastructure

import (
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

func TestCloneFormFieldsDeepCopiesLayoutFieldsAndHelp(t *testing.T) {
	help := &workflowcore.FormHelp{ButtonText: "查看说明", Title: "原因说明", Content: "请填写完整原因"}
	precision := 2
	fields := []workflowcore.FormField{{
		Key: "group", Label: "分组", Type: workflowcore.FormFieldTypeGroup,
		Fields: []workflowcore.FormField{
			{Key: "reason", Label: "原因", Type: workflowcore.FormFieldTypeText, Help: help},
			{Key: "total", Label: "合计", Type: workflowcore.FormFieldTypeCalculation, Calculation: &workflowcore.FormCalculation{
				Expression: "[quantity] * [price]", Display: workflowcore.CalculationDisplayField, Precision: &precision,
			}},
		},
	}}

	cloned := cloneFormFields(fields)
	fields[0].Fields[0].Label = "已修改"
	fields[0].Fields[0].Help.Content = "已修改说明"
	fields[0].Fields[1].Calculation.Expression = "0"
	*fields[0].Fields[1].Calculation.Precision = 6
	if cloned[0].Fields[0].Label != "原因" || cloned[0].Fields[0].Help.Content != "请填写完整原因" {
		t.Fatalf("layout field clone shares draft memory: %#v", cloned)
	}
	if cloned[0].Fields[1].Calculation.Expression != "[quantity] * [price]" || *cloned[0].Fields[1].Calculation.Precision != 2 {
		t.Fatalf("calculation clone shares draft memory: %#v", cloned)
	}
}
