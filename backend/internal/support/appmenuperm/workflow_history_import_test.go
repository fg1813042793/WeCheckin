package appmenuperm

import "testing"

func TestWorkflowHistoryImportButtonIsDeclared(t *testing.T) {
	for _, declaration := range DingTalkH5ButtonDeclarations() {
		if declaration.Key != "dingtalk_h5:button:workflow:history-import" {
			continue
		}
		if declaration.Name != "导入历史表单" {
			t.Fatalf("name = %q", declaration.Name)
		}
		if declaration.Platform != "dingtalk_h5" || declaration.Path != "workflow:history-import" {
			t.Fatalf("history import declaration = %#v", declaration)
		}
		if declaration.ParentKey != "dingtalk_h5:menu:workflow" || declaration.Sort != 74 {
			t.Fatalf("history import hierarchy = %#v", declaration)
		}
		return
	}
	t.Fatal("missing workflow history import button permission")
}
