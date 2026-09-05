package workflowcore

import (
	"strings"
	"testing"
)

func TestPostHandleFormPermissionsPreserveVisibilityAndDowngradeRoutingFields(t *testing.T) {
	definition := Definition{
		Form: []FormField{
			{Key: "visible", Label: "可见字段", Type: FormFieldTypeText},
			{Key: "editable", Label: "可编辑字段", Type: FormFieldTypeText},
			{Key: "hidden", Label: "隐藏字段", Type: FormFieldTypeText},
			{Key: "routing", Label: "分支字段", Type: FormFieldTypeNumber},
		},
		Nodes: []Node{{
			ID: "approval", Type: NodeTypeApproval,
			PostHandleEdit: &PostHandleEditConfig{Enabled: true},
			FormPermissions: []FieldPermission{
				{Field: "visible", Access: FieldAccessRead},
				{Field: "editable", Access: FieldAccessWrite},
				{Field: "hidden", Access: FieldAccessHidden},
				{Field: "routing", Access: FieldAccessWrite},
			},
		}},
		Edges: []Edge{{Condition: &Condition{Field: "routing", Operator: ConditionGT, Value: 100}}},
	}

	permissions := PostHandleFormPermissions(definition, []string{"approval"})
	want := map[string]string{
		"visible":  FieldAccessRead,
		"editable": FieldAccessWrite,
		"hidden":   FieldAccessHidden,
		"routing":  FieldAccessRead,
	}
	if len(permissions) != len(want) {
		t.Fatalf("form permissions = %#v", permissions)
	}
	for _, permission := range permissions {
		if permission.Access != want[permission.Field] {
			t.Fatalf("permission %q = %q, want %q", permission.Field, permission.Access, want[permission.Field])
		}
	}
}

func TestPostHandleEditablePermissionsUseEnabledHandledNodesAndExcludeRoutingFields(t *testing.T) {
	definition := Definition{
		Form: []FormField{
			{Key: "amount", Label: "金额", Type: FormFieldTypeNumber},
			{Key: "summary", Label: "说明", Type: FormFieldTypeTextarea},
			{Key: "result", Label: "结果", Type: FormFieldTypeText},
		},
		Nodes: []Node{
			{
				ID: "approval", Type: NodeTypeApproval,
				PostHandleEdit: &PostHandleEditConfig{Enabled: true},
				FormPermissions: []FieldPermission{
					{Field: "amount", Access: FieldAccessWrite},
					{Field: "summary", Access: FieldAccessWrite},
				},
			},
			{
				ID: "handle", Type: NodeTypeHandle,
				PostHandleEdit:  &PostHandleEditConfig{Enabled: false},
				FormPermissions: []FieldPermission{{Field: "result", Access: FieldAccessWrite}},
			},
		},
		Edges: []Edge{{
			ID: "condition", Source: "gateway", Target: "end",
			Condition: &Condition{Field: "amount", Operator: ConditionGT, Value: 100},
		}},
	}

	permissions := PostHandleEditablePermissions(definition, []string{"approval", "handle"})
	if len(permissions) != 1 || permissions[0].Field != "summary" || permissions[0].Access != FieldAccessWrite {
		t.Fatalf("editable permissions = %#v, want only summary", permissions)
	}
}

func TestValidatePostHandleFormPatchMergesPermissionsAcrossHandledNodes(t *testing.T) {
	definition := Definition{
		Form: []FormField{
			{Key: "summary", Label: "说明", Type: FormFieldTypeTextarea, Required: true},
			{Key: "result", Label: "结果", Type: FormFieldTypeText},
		},
		Nodes: []Node{
			{
				ID: "approval", Type: NodeTypeApproval,
				PostHandleEdit:  &PostHandleEditConfig{Enabled: true},
				FormPermissions: []FieldPermission{{Field: "summary", Access: FieldAccessWrite}},
			},
			{
				ID: "handle", Type: NodeTypeHandle,
				PostHandleEdit:  &PostHandleEditConfig{Enabled: true},
				FormPermissions: []FieldPermission{{Field: "result", Access: FieldAccessWrite}},
			},
		},
	}
	current := map[string]interface{}{"summary": "原说明", "result": "原结果"}

	if err := ValidatePostHandleFormPatch(definition, []string{"approval", "handle"}, current, map[string]interface{}{
		"summary": "新说明",
		"result":  "新结果",
	}); err != nil {
		t.Fatalf("merged post-handle patch rejected: %v", err)
	}

	err := ValidatePostHandleFormPatch(definition, []string{"approval"}, current, map[string]interface{}{"result": "越权修改"})
	if err == nil || !strings.Contains(err.Error(), "无权修改字段 result") {
		t.Fatalf("unauthorized patch error = %v", err)
	}
}
