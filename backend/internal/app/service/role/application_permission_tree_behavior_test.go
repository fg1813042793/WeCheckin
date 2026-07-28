package role

import "testing"

func TestApplicationPermissionTreeContainsBuiltInMenus(t *testing.T) {
	tree := ApplicationPermissionTree()
	if len(tree.Client) == 0 {
		t.Fatalf("client application permission tree must not be empty")
	}
	if len(tree.DingTalkH5) == 0 {
		t.Fatalf("dingtalk h5 application permission tree must not be empty")
	}
	var performance *ApplicationPermissionNode
	for i := range tree.DingTalkH5 {
		if tree.DingTalkH5[i].Key == "dingtalk_h5:menu:performance" {
			performance = &tree.DingTalkH5[i]
			break
		}
	}
	if performance == nil {
		t.Fatalf("dingtalk h5 tree must contain performance root")
	}
	if len(performance.Children) == 0 {
		t.Fatalf("performance root must contain tab children")
	}
}
