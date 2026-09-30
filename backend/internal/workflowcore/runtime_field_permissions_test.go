package workflowcore

import "testing"

func TestRuntimeFieldPermissionsApplyMatchingNodesAndFields(t *testing.T) {
	source := Definition{Form: []FormField{{Key: "old", Type: FormFieldTypeText}, {Key: "shared", Type: FormFieldTypeText}}, Nodes: []Node{
		{ID: "manager", Type: NodeTypeApproval, FormPermissions: []FieldPermission{{Field: "shared", Access: FieldAccessWrite}, {Field: "new", Access: FieldAccessWrite}}},
		{ID: "new-node", Type: NodeTypeApproval, FormPermissions: []FieldPermission{{Field: "shared", Access: FieldAccessWrite}}},
	}}
	raw, err := EncodeRuntimeFieldPermissions(source)
	if err != nil {
		t.Fatal(err)
	}
	target := Definition{Form: []FormField{{Key: "old", Type: FormFieldTypeText}, {Key: "shared", Type: FormFieldTypeText}}, Nodes: []Node{
		{ID: "manager", Type: NodeTypeApproval, FormPermissions: []FieldPermission{{Field: "old", Access: FieldAccessWrite}}},
		{ID: "legacy-node", Type: NodeTypeApproval, FormPermissions: []FieldPermission{{Field: "old", Access: FieldAccessWrite}}},
	}}
	if err := ApplyRuntimeFieldPermissionsJSON(&target, raw); err != nil {
		t.Fatal(err)
	}
	if len(target.Nodes[0].FormPermissions) != 1 || target.Nodes[0].FormPermissions[0].Field != "shared" {
		t.Fatalf("manager permissions = %#v", target.Nodes[0].FormPermissions)
	}
	if len(target.Nodes[1].FormPermissions) != 1 || target.Nodes[1].FormPermissions[0].Field != "old" {
		t.Fatalf("unmatched legacy node must keep version permissions: %#v", target.Nodes[1].FormPermissions)
	}
}

func TestRuntimeFieldPermissionsCloneActions(t *testing.T) {
	definition := Definition{Form: []FormField{{Key: "items", Type: FormFieldTypeDetailList}}, Nodes: []Node{{ID: "manager", Type: NodeTypeApproval, FormPermissions: []FieldPermission{{Field: "items", Access: FieldAccessWrite, Actions: []string{"add"}}}}}}
	config := RuntimeFieldPermissionsFromDefinition(definition)
	config.Nodes["manager"][0].Actions[0] = "delete"
	if definition.Nodes[0].FormPermissions[0].Actions[0] != "add" {
		t.Fatal("runtime config must be defensively cloned")
	}
}
