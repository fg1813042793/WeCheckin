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
		`Key: "dingtalk_h5:menu:performance:mine", Name: "我的绩效", Platform: "dingtalk_h5", Path: "performance:mine", ParentKey: "dingtalk_h5:menu:performance"`,
		`Key: "dingtalk_h5:menu:performance:history", Name: "历史绩效", Platform: "dingtalk_h5", Path: "performance:history", ParentKey: "dingtalk_h5:menu:performance"`,
		`Key: "dingtalk_h5:menu:performance:manager", Name: "上级评价", Platform: "dingtalk_h5", Path: "performance:manager", ParentKey: "dingtalk_h5:menu:performance"`,
		`Key: "dingtalk_h5:menu:performance:hrbp", Name: "HRBP评价", Platform: "dingtalk_h5", Path: "performance:hrbp", ParentKey: "dingtalk_h5:menu:performance"`,
		`Key: "dingtalk_h5:menu:performance:summary", Name: "HRBP汇总", Platform: "dingtalk_h5", Path: "performance:summary", ParentKey: "dingtalk_h5:menu:performance"`,
		`Key: "dingtalk_h5:menu:performance:org", Name: "流程执行", Platform: "dingtalk_h5", Path: "performance:org", ParentKey: "dingtalk_h5:menu:performance"`,
		`Key: "dingtalk_h5:menu:performance:template", Name: "绩效模版", Platform: "dingtalk_h5", Path: "performance:template", ParentKey: "dingtalk_h5:menu:performance"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 tabs must expose parent hierarchy with %s", snippet)
		}
	}
}
