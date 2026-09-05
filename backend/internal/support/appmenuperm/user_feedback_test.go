package appmenuperm

import "testing"

func TestDingTalkH5FeedbackMenuMetadataMatchesBootstrapContract(t *testing.T) {
	for _, declaration := range DingTalkH5MenuDeclarations() {
		if declaration.Key != "dingtalk_h5:menu:feedback" {
			continue
		}
		if declaration.Name != "我的反馈" {
			t.Fatalf("name = %q, want 我的反馈", declaration.Name)
		}
		if declaration.Icon != "chat" {
			t.Fatalf("icon = %q, want chat", declaration.Icon)
		}
		if declaration.Platform != "dingtalk_h5" || declaration.Path != "feedback" || declaration.Sort != 110 {
			t.Fatalf("feedback menu metadata = %#v", declaration)
		}
		return
	}
	t.Fatal("missing dingtalk h5 feedback menu declaration")
}
