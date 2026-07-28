package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5UsesTwoTopLevelMenusAndPerformanceTabs(t *testing.T) {
	catalogSrc, err := os.ReadFile("../../support/appmenuperm/catalog.go")
	if err != nil {
		t.Fatalf("read menu catalog: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	appShellSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/AppShell.vue")
	if err != nil {
		t.Fatalf("read app shell: %v", err)
	}
	combined := string(catalogSrc) + string(reviewsSrc) + string(pageSrc) + string(appShellSrc)
	for _, snippet := range []string{
		`Key: "dingtalk_h5:menu:dashboard", Name: "工作台", Platform: "dingtalk_h5", Path: "dashboard"`,
		`Key: "dingtalk_h5:menu:performance", Name: "绩效管理", Platform: "dingtalk_h5", Path: "performance"`,
		`Key: "dingtalk_h5:menu:performance:mine", Name: "本月绩效", Platform: "dingtalk_h5", Path: "performance:mine"`,
		`Key: "dingtalk_h5:menu:performance:history", Name: "历史绩效", Platform: "dingtalk_h5", Path: "performance:history"`,
		`Key: "dingtalk_h5:menu:performance:hrbp", Name: "HRBP评价", Platform: "dingtalk_h5", Path: "performance:hrbp"`,
		`Key: "dingtalk_h5:menu:performance:summary", Name: "HRBP汇总", Platform: "dingtalk_h5", Path: "performance:summary"`,
		`Key: "dingtalk_h5:menu:performance:org", Name: "流程执行", Platform: "dingtalk_h5", Path: "performance:org"`,
		`Key: "dingtalk_h5:menu:performance:template", Name: "绩效模版", Platform: "dingtalk_h5", Path: "performance:template"`,
		`const activePerformanceTab = ref('mine')`,
		`const performanceTabs = computed`,
		`<view v-if="activeView === 'performance' && performanceTabs.length" class="performance-tabs">`,
		`:performance-tabs="performanceTabs"`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("two-level navigation must include %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`Key: "dingtalk_h5:menu:mine", Name: "我的绩效"`,
		`Key: "dingtalk_h5:menu:hrbp", Name: "HRBP评价"`,
		`Key: "dingtalk_h5:menu:summary", Name: "汇总"`,
		`Key: "dingtalk_h5:menu:org", Name: "组织架构"`,
		`Key: "dingtalk_h5:menu:template", Name: "模板"`,
		`['mine', '我的绩效', 'mine']`,
		`['hrbp', 'HRBP评价', 'hrbp']`,
		`['summary', 'HRBP汇总', 'summary']`,
		`['org', '组织架构', 'org']`,
		`['template', '模板', 'template']`,
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("old top-level navigation must be removed, found %q", forbidden)
		}
	}
}
