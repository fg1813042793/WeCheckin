package infrastructure

import (
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

func TestCloneFormFieldsDeepCopiesLayoutFieldsAndHelp(t *testing.T) {
	help := &workflowcore.FormHelp{ButtonText: "查看说明", Title: "原因说明", Content: "请填写完整原因"}
	fields := []workflowcore.FormField{{
		Key: "group", Label: "分组", Type: workflowcore.FormFieldTypeGroup,
		Fields: []workflowcore.FormField{{Key: "reason", Label: "原因", Type: workflowcore.FormFieldTypeText, Help: help}},
	}}

	cloned := cloneFormFields(fields)
	fields[0].Fields[0].Label = "已修改"
	fields[0].Fields[0].Help.Content = "已修改说明"
	if cloned[0].Fields[0].Label != "原因" || cloned[0].Fields[0].Help.Content != "请填写完整原因" {
		t.Fatalf("layout field clone shares draft memory: %#v", cloned)
	}
}
