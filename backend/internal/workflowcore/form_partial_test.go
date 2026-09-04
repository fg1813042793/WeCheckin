package workflowcore

import "testing"

func TestValidateFormDataPartialAllowsIncompleteDetailListDraft(t *testing.T) {
	sumTarget := 100.0
	fields := []FormField{{
		Key: "targets", Label: "目标", Type: FormFieldTypeDetailList,
		RowKey: "id", MinRows: 2, MaxRows: 3,
		Columns: []FormField{
			{Key: "target", Label: "目标内容", Type: FormFieldTypeTextarea, Required: true},
			{Key: "weight", Label: "权重", Type: FormFieldTypeNumber, Required: true},
		},
		Rules: []FormValidationRule{{
			ID: "weight_sum", Type: FormRuleColumnSum, Column: "weight",
			Operator: FormRuleOperatorEQ, Value: &sumTarget,
		}},
	}}
	draft := map[string]interface{}{
		"targets": []interface{}{
			map[string]interface{}{"id": "row-1", "target": "进行中", "weight": ""},
		},
	}

	if err := ValidateFormData(fields, draft, true); err != nil {
		t.Fatalf("partial detail-list draft rejected: %v", err)
	}
	if err := ValidateFormData(fields, draft, false); err == nil {
		t.Fatal("complete submission must still reject the incomplete detail list")
	}

	unknownColumn := map[string]interface{}{
		"targets": []interface{}{
			map[string]interface{}{"id": "row-1", "target": "进行中", "unknown": true},
		},
	}
	if err := ValidateFormData(fields, unknownColumn, true); err == nil {
		t.Fatal("partial draft must reject unknown detail-list columns")
	}

	invalidType := map[string]interface{}{
		"targets": []interface{}{
			map[string]interface{}{"id": "row-1", "target": "进行中", "weight": "not-a-number"},
		},
	}
	if err := ValidateFormData(fields, invalidType, true); err == nil {
		t.Fatal("partial draft must reject invalid non-empty column values")
	}
}
