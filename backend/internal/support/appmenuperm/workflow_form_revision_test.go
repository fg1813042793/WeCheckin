package appmenuperm

import "testing"

func TestWorkflowCompletedRevisionButtonsUseWorkflowMenuAsParent(t *testing.T) {
	want := map[string]string{
		"dingtalk_h5:button:workflow:form-revision-create": "发起完成后表单修订",
		"dingtalk_h5:button:workflow:form-revision-handle": "处理完成后表单修订",
	}
	for _, declaration := range DingTalkH5ButtonDeclarations() {
		name, ok := want[declaration.Key]
		if !ok {
			continue
		}
		if declaration.Name != name || declaration.ParentKey != "dingtalk_h5:menu:workflow" {
			t.Fatalf("revision button declaration=%#v", declaration)
		}
		delete(want, declaration.Key)
	}
	if len(want) != 0 {
		t.Fatalf("missing revision buttons: %#v", want)
	}
}
