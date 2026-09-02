package appmenuperm

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5DeclarationsExposePerformanceTabParent(t *testing.T) {
	src, err := os.ReadFile("catalog.go")
	if err != nil {
		t.Fatalf("read catalog.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"ParentKey string",
		`Key: "dingtalk_h5:menu:performance:mine", Name: "我的绩效", Platform: "dingtalk_h5", Path: "performance:mine", Icon: "mine", ParentKey: "dingtalk_h5:menu:performance", Sort: 30`,
		`Key: "dingtalk_h5:menu:performance:history", Name: "历史绩效", Platform: "dingtalk_h5", Path: "performance:history", Icon: "history", ParentKey: "dingtalk_h5:menu:performance", Sort: 40`,
		`Key: "dingtalk_h5:menu:performance:manager", Name: "上级评价", Platform: "dingtalk_h5", Path: "performance:manager", Icon: "manager", ParentKey: "dingtalk_h5:menu:performance", Sort: 50`,
		`Key: "dingtalk_h5:menu:performance:hrbp", Name: "HRBP评价", Platform: "dingtalk_h5", Path: "performance:hrbp", Icon: "hrbp", ParentKey: "dingtalk_h5:menu:performance", Sort: 60`,
		`Key: "dingtalk_h5:menu:performance:summary", Name: "HRBP汇总", Platform: "dingtalk_h5", Path: "performance:summary", Icon: "summary", ParentKey: "dingtalk_h5:menu:performance", Sort: 70`,
		`Key: "dingtalk_h5:menu:performance:org", Name: "流程执行", Platform: "dingtalk_h5", Path: "performance:org", Icon: "org", ParentKey: "dingtalk_h5:menu:performance", Sort: 80`,
		`Key: "dingtalk_h5:menu:performance:template", Name: "绩效模版", Platform: "dingtalk_h5", Path: "performance:template", Icon: "template", ParentKey: "dingtalk_h5:menu:performance", Sort: 90`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 tabs must expose parent hierarchy with %s", snippet)
		}
	}
}

func TestDingTalkH5DeclarationsExposeWorkflowMenu(t *testing.T) {
	src, err := os.ReadFile("catalog.go")
	if err != nil {
		t.Fatalf("read catalog.go: %v", err)
	}
	const workflowMenu = "Key: \"dingtalk_h5:menu:workflow\", Name: \"流程审批\", Platform: \"dingtalk_h5\", Path: \"workflow\", Icon: \"workflow\""
	if !strings.Contains(string(src), workflowMenu) {
		t.Fatalf("dingtalk h5 permissions must expose the workflow menu with %s", workflowMenu)
	}
}
