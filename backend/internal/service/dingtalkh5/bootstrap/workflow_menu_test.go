package bootstrap

import "testing"

func TestWorkflowMenuRequiresExplicitPermissionKey(t *testing.T) {
	withoutWorkflow := dingTalkH5MenusByKeysWithLabelsAndIcons([]string{
		"dingtalk_h5:menu:dashboard",
	}, nil, nil)
	for _, menu := range withoutWorkflow {
		if menu.Key == "workflow" {
			t.Fatal("workflow menu must stay hidden without its permission key")
		}
	}

	withWorkflow := dingTalkH5MenusByKeysWithLabelsAndIcons([]string{
		"dingtalk_h5:menu:dashboard",
		"dingtalk_h5:menu:workflow",
	}, nil, nil)
	if len(withWorkflow) != 2 {
		t.Fatalf("top-level menus = %d, want 2: %#v", len(withWorkflow), withWorkflow)
	}
	workflow := withWorkflow[1]
	if workflow.Key != "workflow" || workflow.Label != "流程审批" {
		t.Fatalf("workflow menu = %#v", workflow)
	}
	if workflow.PermissionKey != "dingtalk_h5:menu:workflow" {
		t.Fatalf("workflow permission key = %q", workflow.PermissionKey)
	}
}
