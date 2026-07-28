package dingtalkh5

import (
	"os"
	"strings"
	"testing"

	"wecheckin-backend/backend/internal/model"
)

func TestHrbpAccessIsLimitedToAssignedOrResponsibleData(t *testing.T) {
	hrbp := &model.DingTalkH5PerfUser{
		Account:                "lucky",
		Role:                   "hrbp",
		ResponsibleDepartments: encodeJSON([]string{"产品部"}),
	}
	unrelatedEmployee := &model.DingTalkH5PerfUser{
		Account:          "lip",
		Role:             "employee",
		DepartmentLevel2: "研发部",
		HRBPAccount:      "nick",
	}
	unrelatedReview := model.DingTalkH5PerfReview{
		EmployeeAccount:  "lip",
		HRBPAccount:      "nick",
		DepartmentLevel2: "研发部",
	}
	if canViewReview(hrbp, unrelatedReview, unrelatedEmployee) {
		t.Fatalf("hrbp should not view unrelated employee review")
	}

	assignedReview := unrelatedReview
	assignedReview.HRBPAccount = "lucky"
	if !canViewReview(hrbp, assignedReview, unrelatedEmployee) {
		t.Fatalf("hrbp should view assigned review")
	}

	responsibleReview := unrelatedReview
	responsibleReview.DepartmentLevel2 = "产品部"
	if !canViewReview(hrbp, responsibleReview, unrelatedEmployee) {
		t.Fatalf("hrbp should view responsible department review")
	}
}

func TestHrbpWithoutResponsibleDepartmentDoesNotFallBackToOrgLevel(t *testing.T) {
	hrbp := &model.DingTalkH5PerfUser{
		Account:          "lucky",
		Role:             "hrbp",
		DepartmentLevel1: "M/H业务",
	}
	employee := &model.DingTalkH5PerfUser{
		Account:          "lip",
		Role:             "employee",
		DepartmentLevel1: "M/H业务",
		DepartmentLevel2: "研发部",
		HRBPAccount:      "nick",
	}
	review := model.DingTalkH5PerfReview{
		EmployeeAccount:  "lip",
		HRBPAccount:      "nick",
		DepartmentLevel1: "M/H业务",
		DepartmentLevel2: "研发部",
	}
	if canViewReview(hrbp, review, employee) {
		t.Fatalf("hrbp without explicit responsible department should not inherit org-level access")
	}
}

func TestReviewVisibilityWhereSeparatesDashboardAndSummaryScope(t *testing.T) {
	user := &model.DingTalkH5PerfUser{
		Account:          "cube",
		Role:             "manager",
		DepartmentLevel1: "M/H业务",
		DepartmentLevel2: "产品部",
	}
	where, args := reviewVisibilityWhere(user, "dashboard")
	for _, want := range []string{"employee_account", "manager_account", "hrbp_account", "hrbp_reviewer_account"} {
		if !strings.Contains(where, want) {
			t.Fatalf("dashboard scope where %q must include %s", where, want)
		}
	}
	if strings.Contains(where, "department_level") {
		t.Fatalf("dashboard scope must stay personal, got %q", where)
	}
	if len(args) != 4 {
		t.Fatalf("dashboard scope args = %d, want 4", len(args))
	}

	summaryWhere, summaryArgs := reviewVisibilityWhere(user, "summary")
	if !strings.Contains(summaryWhere, "department_level1") || !strings.Contains(summaryWhere, "department_level2") {
		t.Fatalf("summary scope should include department scope, got %q", summaryWhere)
	}
	if len(summaryArgs) <= len(args) {
		t.Fatalf("summary scope should carry personal and department args, got %#v", summaryArgs)
	}

	adminWhere, adminArgs := reviewVisibilityWhere(&model.DingTalkH5PerfUser{Account: "nick", Role: "admin"}, "summary")
	if adminWhere != "" || len(adminArgs) != 0 {
		t.Fatalf("admin summary scope should allow all data, got where=%q args=%#v", adminWhere, adminArgs)
	}
}

func TestVisiblePerfUsersFiltersPersonalDirectory(t *testing.T) {
	current := &model.DingTalkH5PerfUser{
		Account:                "lucky",
		Role:                   "hrbp",
		ResponsibleDepartments: encodeJSON([]string{"产品部"}),
	}
	users := []model.DingTalkH5PerfUser{
		{Account: "lucky", Name: "Lucky", Role: "hrbp"},
		{Account: "cube", Name: "Cube", Role: "supervisor", DepartmentLevel2: "产品部", HRBPAccount: "lucky"},
		{Account: "lip", Name: "Lip", Role: "employee", DepartmentLevel2: "研发部", HRBPAccount: "nick"},
	}
	got := visiblePerfUsers(current, users)
	accounts := map[string]bool{}
	for _, user := range got {
		accounts[user.Account] = true
	}
	if !accounts["lucky"] || !accounts["cube"] {
		t.Fatalf("visible users should include current and responsible users, got %#v", accounts)
	}
	if accounts["lip"] {
		t.Fatalf("visible users should not include unrelated user, got %#v", accounts)
	}
}

func TestDingTalkH5ReviewScopeIsPassedThroughAPI(t *testing.T) {
	handlerSrc, err := os.ReadFile("../../handler/client/dingtalkh5/handler.go")
	if err != nil {
		t.Fatalf("read handler: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	for _, snippet := range []string{
		`Scope:      strings.TrimSpace(c.Query("scope"))`,
		`dingTalkPerformanceApi.reviews({ scope: reviewScopeForContentView(), ...params })`,
		`function reviewScopeForContentView()`,
		`exportUrl({ scope: 'summary', ...summaryFilters })`,
	} {
		if !strings.Contains(string(handlerSrc)+string(pageSrc), snippet) {
			t.Fatalf("dingtalk h5 reviews API must preserve scope snippet %q", snippet)
		}
	}
}

func TestDingTalkH5WorkbenchIsStatsOnlyAndAccountMenuRemoved(t *testing.T) {
	routeSrc, err := os.ReadFile("../../../../cmd/routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes: %v", err)
	}
	catalogSrc, err := os.ReadFile("../../support/appmenuperm/catalog.go")
	if err != nil {
		t.Fatalf("read menu catalog: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench: %v", err)
	}
	combined := string(routeSrc) + string(catalogSrc) + string(pageSrc) + string(workbenchSrc)
	for _, snippet := range []string{
		`auth.GET("/workbench", handler.Workbench)`,
		`dingTalkPerformanceApi.workbench()`,
		`if (state.view === 'dashboard')`,
		`ctx.workbenchCards.value.map`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("workbench stats-only flow must include %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`dingtalk_h5:menu:account`,
		`Name: "账号设置"`,
		`Path: "account"`,
		`AccountView`,
		`state.view === 'account'`,
		`['account', '账号设置', 'account']`,
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("dingtalk h5 account menu/page must be removed, found %q", forbidden)
		}
	}
}
