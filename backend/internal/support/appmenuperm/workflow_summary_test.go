package appmenuperm

import "testing"

func TestWorkflowSummaryButtonUsesWorkflowMenuAsParent(t *testing.T) {
	for _, declaration := range DingTalkH5ButtonDeclarations() {
		if declaration.Key != "dingtalk_h5:button:workflow:summary" {
			continue
		}
		if declaration.ParentKey != "dingtalk_h5:menu:workflow" {
			t.Fatalf("parent = %q", declaration.ParentKey)
		}
		return
	}
	t.Fatal("missing workflow summary button permission")
}
