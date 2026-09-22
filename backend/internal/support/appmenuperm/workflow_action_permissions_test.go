package appmenuperm

import "testing"

func TestDingTalkH5WorkflowActionButtonsAreDeclared(t *testing.T) {
	want := map[string]struct {
		name string
		path string
	}{
		"dingtalk_h5:button:workflow:approve": {name: "通过审批", path: "workflow:approve"},
		"dingtalk_h5:button:workflow:reject":  {name: "驳回流程", path: "workflow:reject"},
		"dingtalk_h5:button:workflow:return":  {name: "退回流程", path: "workflow:return"},
		"dingtalk_h5:button:workflow:submit":  {name: "提交办理", path: "workflow:submit"},
	}
	for _, item := range DingTalkH5ButtonDeclarations() {
		expected, ok := want[item.Key]
		if !ok {
			continue
		}
		if item.Name != expected.name || item.Path != expected.path || item.ParentKey != "dingtalk_h5:menu:workflow" {
			t.Fatalf("workflow action button %#v", item)
		}
		delete(want, item.Key)
	}
	if len(want) != 0 {
		t.Fatalf("missing workflow action buttons: %#v", want)
	}
}
