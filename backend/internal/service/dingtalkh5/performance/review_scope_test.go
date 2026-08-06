package performance

import (
	"os"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

func cssRule(src, selector string) string {
	start := strings.Index(src, selector+" {")
	if start < 0 {
		return ""
	}
	rest := src[start:]
	end := strings.Index(rest, "\n}")
	if end < 0 {
		return rest
	}
	return rest[:end+2]
}

func TestHrbpAccessIsLimitedToAssignedOrResponsibleData(t *testing.T) {
	hrbp := &model.DingTalkH5PerfUser{
		Account:                "lucky",
		ResponsibleDepartments: encodeJSON([]string{"产品部"}),
	}
	unrelatedEmployee := &model.DingTalkH5PerfUser{
		Account:          "lip",
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
		DepartmentLevel1: "M/H业务",
	}
	employee := &model.DingTalkH5PerfUser{
		Account:          "lip",
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
	if summaryWhere != "" || len(summaryArgs) != 0 {
		t.Fatalf("summary scope should be delegated to unified data scope, got where=%q args=%#v", summaryWhere, summaryArgs)
	}
}

func TestMineReviewScopeBypassesUnifiedDataScopeForIndexedSelfQuery(t *testing.T) {
	src, err := os.ReadFile("review_scope.go")
	if err != nil {
		t.Fatalf("read review_scope.go: %v", err)
	}
	body := functionBody(string(src), "func applyReviewVisibilityScopeContext")
	mineGuard := strings.Index(body, "normalizedScope == reviewScopeMine")
	dataScopeCall := strings.Index(body, "reviewDataScopeWhereContext")
	if mineGuard < 0 {
		t.Fatalf("mine review scope should have a dedicated fast path")
	}
	if dataScopeCall >= 0 && mineGuard > dataScopeCall {
		t.Fatalf("mine review scope must bypass unified data scope before querying permissions")
	}
	if !strings.Contains(body, "return query.Where(where, args...), nil") {
		t.Fatalf("mine review scope should apply only the indexed employee_account predicate")
	}
}

func TestDashboardReviewScopeBypassesUnifiedDataScopeForPersonalList(t *testing.T) {
	src, err := os.ReadFile("review_scope.go")
	if err != nil {
		t.Fatalf("read review_scope.go: %v", err)
	}
	body := functionBody(string(src), "func applyReviewVisibilityScopeContext")
	dashboardGuard := strings.Index(body, "normalizedScope == reviewScopeDashboard")
	dataScopeCall := strings.Index(body, "reviewDataScopeWhereContext")
	if dashboardGuard < 0 {
		t.Fatalf("dashboard review scope should have a dedicated personal fast path")
	}
	if dataScopeCall >= 0 && dashboardGuard > dataScopeCall {
		t.Fatalf("dashboard review scope must bypass unified data scope before querying permissions")
	}
}

func TestVisiblePerfUsersFiltersPersonalDirectory(t *testing.T) {
	current := &model.DingTalkH5PerfUser{
		Account:                "lucky",
		ResponsibleDepartments: encodeJSON([]string{"产品部"}),
	}
	users := []model.DingTalkH5PerfUser{
		{Account: "lucky", Name: "Lucky"},
		{Account: "cube", Name: "Cube", DepartmentLevel2: "产品部", HRBPAccount: "lucky"},
		{Account: "lip", Name: "Lip", DepartmentLevel2: "研发部", HRBPAccount: "nick"},
	}
	got := visiblePerfUsers(current, users, permissionsupport.DataScope{Mode: 2, Ready: true})
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

func TestVisiblePerfUsersRequiresUnifiedDataScopeForFullDirectory(t *testing.T) {
	current := &model.DingTalkH5PerfUser{Account: "nick", RoleID: 0}
	users := []model.DingTalkH5PerfUser{
		{Account: "nick", Name: "Nick"},
		{Account: "lip", Name: "Lip"},
	}
	if got := visiblePerfUsers(current, users, permissionsupport.DataScope{}); len(got) != 0 {
		t.Fatalf("user without unified data scope must not see all users, got %#v", got)
	}

	allScope := permissionsupport.DataScope{Mode: 1, Ready: true}
	if got := visiblePerfUsers(current, users, allScope); len(got) != 2 {
		t.Fatalf("data:all should see full directory, got %#v", got)
	}

	selfScope := permissionsupport.DataScope{Mode: 3, Ready: true}
	got := visiblePerfUsers(current, users, selfScope)
	if len(got) != 1 || got[0].Account != "nick" {
		t.Fatalf("data:self should see current user only, got %#v", got)
	}

	deptScope := permissionsupport.DataScope{Mode: 2, Ready: true}
	got = visiblePerfUsers(current, users, deptScope)
	if len(got) != 1 || got[0].Account != "nick" {
		t.Fatalf("data:dept without matching department should not see all users, got %#v", got)
	}
}

func TestReviewDataScopeUsesAuditOwnerFields(t *testing.T) {
	src, err := os.ReadFile("data_scope.go")
	if err != nil {
		t.Fatalf("read data_scope.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`fields.CreateByField() + " = ?"`,
		`fields.CreateDeptField() + " IN ?"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("review data scope should include audit owner field snippet %q", snippet)
		}
	}
}

func TestDingTalkH5BackendDoesNotUsePerformanceRoleForPermissions(t *testing.T) {
	checks := map[string][]string{
		"access.go": {
			"peopleLeaderRoles",
			"editableRoles",
			"user.Role",
			"leader.Role",
			"employee.Role",
		},
		"review_scope.go": {
			"user.Role",
		},
		"workbench.go": {
			"user.Role",
		},
		"users.go": {
			"payload.Role",
			"manager.Role",
			"editableRoles",
			"peopleLeaderRoles",
			"users[i].Role",
			"users[j].Role",
		},
		"review_helpers.go": {
			"employee.Role ==",
		},
	}
	for file, snippets := range checks {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, snippet := range snippets {
			if strings.Contains(string(src), snippet) {
				t.Fatalf("%s must not use performance role for permission flow, found %q", file, snippet)
			}
		}
	}
}

func TestDingTalkH5FrontendDoesNotExposePerformanceRole(t *testing.T) {
	files := []string{
		"../../../../../dingtalk-h5/pages/index/index.vue",
		"../../../../../dingtalk-h5/components/performance/AppShell.vue",
		"../../../../../dingtalk-h5/components/performance/OrgView.vue",
		"../../../../../dingtalk-h5/components/performance/AccountView.vue",
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		body := string(src)
		for _, snippet := range []string{
			"roleName",
			"roleText",
			"state.user?.role",
			"state.user.role",
			"user.role",
			"selectedUser.role",
			"configForm.role",
		} {
			if strings.Contains(body, snippet) {
				t.Fatalf("%s must not expose performance role in H5 UI, found %q", file, snippet)
			}
		}
	}
}

func TestDingTalkH5ReviewScopeIsPassedThroughAPI(t *testing.T) {
	handlerSrc, err := os.ReadFile("../../../handler/dingtalkh5/performance/review/handler.go")
	if err != nil {
		t.Fatalf("read review handler: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	for _, snippet := range []string{
		`strings.TrimSpace(c.Query("scope"))`,
		`dingTalkPerformanceApi.reviews({ ...reviewQueryParamsForContentView(), ...params })`,
		`function reviewQueryParamsForContentView()`,
		`function reviewScopeForContentView()`,
		`exportSummary`,
	} {
		if !strings.Contains(string(handlerSrc)+string(pageSrc), snippet) {
			t.Fatalf("dingtalk h5 reviews API must preserve scope snippet %q", snippet)
		}
	}
}

func TestDingTalkH5NotificationLinkDeepLinksToReview(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	page := string(pageSrc)
	for _, snippet := range []string{
		`function initialReviewDeepLink()`,
		`async function applyReviewDeepLinkIfNeeded()`,
		`queryParam('reviewNo')`,
		`queryParam('view')`,
		`dingTalkPerformanceApi.reviewDetail(deepLink.reviewNo)`,
		`await openWorkbenchTodo(detail, { preferredView: deepLink.view, reviewTab: deepLink.reviewTab })`,
		`if (!(await applyReviewDeepLinkIfNeeded()))`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("dingtalk h5 notification link must deep-link to review with %q", snippet)
		}
	}
}

func TestDingTalkH5SummaryOnlyLoadsCompletedReviews(t *testing.T) {
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	combined := string(reviewsSrc) + string(pageSrc)
	for _, snippet := range []string{
		`filters.Scope = reviewScopeSummary`,
		`filters.Status = ReviewStatusCompleted`,
		`} else if (view === 'summary') {`,
		`params.status = 'completed'`,
		`if (item.status !== 'completed') return false`,
		`status: 'completed'`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("HRBP summary must only load archived reviews; missing snippet %q", snippet)
		}
	}
}

func TestDingTalkH5SummaryFiltersSupportEmployeeAndDepartmentName(t *testing.T) {
	handlerSrc, err := os.ReadFile("../../../handler/dingtalkh5/performance/review/handler.go")
	if err != nil {
		t.Fatalf("read review handler: %v", err)
	}
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	summarySrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/SummaryView.js")
	if err != nil {
		t.Fatalf("read summary view: %v", err)
	}
	combined := string(handlerSrc) + string(typesSrc) + string(reviewsSrc) + string(pageSrc) + string(summarySrc)
	for _, snippet := range []string{
		`EmployeeName    string`,
		`DepartmentName  string`,
		`DepartmentNames []string`,
		`strings.TrimSpace(c.Query("employeeName"))`,
		`strings.TrimSpace(c.Query("departmentName"))`,
		`DepartmentNames: splitQueryList(c.Query("departmentNames"))`,
		`filters.EmployeeName`,
		`filters.DepartmentName`,
		`filters.DepartmentNames`,
		`applyReviewDepartmentNamesQuery`,
		`ctx.summaryFilters.employeeName`,
		`ctx.summaryFilters.departmentName`,
		`placeholder: '员工姓名'`,
		`placeholder: '搜索部门名称'`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("summary filters must support employee/department name snippet %q", snippet)
		}
	}
}

func TestDingTalkH5SummaryDepartmentFilterUsesSearchableMultiSelectTree(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	summarySrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/SummaryView.js")
	if err != nil {
		t.Fatalf("read summary view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	page := string(pageSrc)
	summary := string(summarySrc)
	style := string(styleSrc)
	combined := page + summary + style
	for _, snippet := range []string{
		`departmentNames: []`,
		`const departmentDropdownOpen = ref(false)`,
		`const departmentSearchKeyword = ref('')`,
		`const summaryDepartmentTree = computed(() => buildSummaryDepartmentTree(ctx.state.users))`,
		`const summaryDepartmentRows = computed(() => flattenSummaryDepartmentTree(filterSummaryDepartmentTree(summaryDepartmentTree.value, departmentSearchKeyword.value), summaryDepartmentExpandedKeys.value, departmentSearchKeyword.value))`,
		`class: 'summary-department-filter'`,
		`'summary-department-trigger'`,
		`placeholder: '搜索部门名称'`,
		`class: 'summary-department-panel'`,
		`class: 'summary-department-search'`,
		`class: 'summary-department-tree'`,
		`renderDepartmentRow(row)`,
		`toggleSummaryDepartment(row)`,
		`summaryDepartmentCheckState(row)`,
		`'summary-department-check'`,
		`'summary-department-check-indeterminate'`,
		`selectedDepartmentLabels.value`,
		`ctx.summaryFilters.departmentNames`,
		`summaryDepartmentMatches(item.department, summaryFilters.departmentName, summaryFilters.departmentNames)`,
		`Object.assign(summaryFilters, { employeeName: '', departmentName: '', departmentNames: [], period: '', status: '' })`,
		`.summary-department-filter`,
		`.summary-department-panel`,
		`.summary-department-search`,
		`.summary-department-tree`,
		`.summary-department-check`,
		`.summary-department-check-indeterminate`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("summary department filter should use searchable multi-select tree with %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`placeholder: '部门名称', onInput: (event) => { ctx.summaryFilters.departmentName = event.detail?.value ?? event.target.value }`,
	} {
		if strings.Contains(summary, forbidden) {
			t.Fatalf("summary department filter should not keep legacy plain input %q", forbidden)
		}
	}
}

func TestDingTalkH5SummaryPeriodFilterUsesMonthPicker(t *testing.T) {
	summarySrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/SummaryView.js")
	if err != nil {
		t.Fatalf("read summary view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	summary := string(summarySrc)
	combined := summary + string(styleSrc)
	for _, snippet := range []string{
		`renderSummaryPeriodFilter()`,
		`function renderSummaryPeriodFilter()`,
		`h('picker', {`,
		`mode: 'date'`,
		`fields: 'month'`,
		`value: ctx.summaryFilters.period`,
		`ctx.summaryFilters.period = event.detail?.value || ''`,
		`class: 'summary-month-picker'`,
		`ctx.summaryFilters.period || '考评月份'`,
		`.summary-month-picker`,
		`.summary-month-value`,
		`.summary-month-arrow`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("summary period filter should use dropdown month picker with %q", snippet)
		}
	}
	if strings.Contains(summary, `h('input', { class: 'field-input', type: 'month', value: ctx.summaryFilters.period`) {
		t.Fatalf("summary period filter should not use plain month input")
	}
}

func TestDingTalkH5WorkbenchShowsTodoListAndAccountMenuRemoved(t *testing.T) {
	routeSrc, err := os.ReadFile("../../../../cmd/routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes: %v", err)
	}
	catalogSrc, err := os.ReadFile("../../../support/appmenuperm/catalog.go")
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
		`auth.GET("/workbench", handler.Bootstrap.Workbench)`,
		`if (state.view === 'dashboard')`,
		`loadReviews({ pageSize: 100 }, { autoSelectReview: false })`,
		`function openWorkbenchTodo(review)`,
		`openWorkbenchTodo,`,
		`function renderWorkbenchTodoList(ctx)`,
		`ctx.queueReviews()`,
		`class: 'workbench-todo-list'`,
		`onClick: () => ctx.openWorkbenchTodo(review)`,
		`class: 'review-detail-only'`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("workbench todo-list flow must include %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`ctx.workbenchCards.value.map`,
		`class: 'stats-grid workbench-stats-grid'`,
		`dingtalk_h5:menu:account`,
		`Name: "账号设置"`,
		`Path: "account"`,
		`AccountView`,
		`state.view === 'account'`,
		`['account', '账号设置', 'account']`,
		`class: 'panel list-panel'`,
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("dingtalk h5 account menu/page must be removed, found %q", forbidden)
		}
	}
}

func TestDingTalkH5HistoryPerformanceUsesTableList(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	workbench := string(workbenchSrc)
	style := string(styleSrc)
	for _, snippet := range []string{
		`if (ctx.contentView.value === 'history')`,
		`class: 'history-table-wrap table-wrap'`,
		`class: 'history-performance-table summary-table'`,
		`class: 'history-empty-row'`,
		`.history-performance-table`,
	} {
		if !strings.Contains(workbench+style, snippet) {
			t.Fatalf("history performance table view must include %q", snippet)
		}
	}
	columnsStart := strings.Index(workbench, `const HISTORY_TABLE_COLUMNS = [`)
	columnsEnd := strings.Index(workbench, `const HRBP_REVIEW_TABLE_COLUMNS = [`)
	if columnsStart < 0 || columnsEnd < 0 || columnsStart > columnsEnd {
		t.Fatalf("history table columns must be defined before hrbp review table columns")
	}
	columnsBody := workbench[columnsStart:columnsEnd]
	for _, snippet := range []string{
		`{ key: 'period', label: '考评月份', mobile: true }`,
		`{ key: 'status', label: '状态', mobile: true }`,
		`{ key: 'objectiveScore', label: '目标得分' }`,
		`{ key: 'managerGrade', label: '上级分档' }`,
		`{ key: 'hrbpGrade', label: 'HRBP分档' }`,
		`{ key: 'finalGrade', label: '最终分档', mobile: true }`,
	} {
		if !strings.Contains(columnsBody, snippet) {
			t.Fatalf("history performance table columns must include %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`key: 'department'`,
		`label: '部门'`,
		`key: 'employeeConfirm'`,
		`label: '员工确认'`,
	} {
		if strings.Contains(columnsBody, forbidden) {
			t.Fatalf("history performance table columns should not include %q", forbidden)
		}
	}
	historyIndex := strings.Index(workbench, `if (ctx.contentView.value === 'history')`)
	detailIndex := strings.Index(workbench, `class: 'review-detail-only'`)
	if historyIndex < 0 || detailIndex < 0 || historyIndex > detailIndex {
		t.Fatalf("history table branch must run before default review detail branch")
	}
	historyBody := workbench[historyIndex:detailIndex]
	for _, duplicateSnippet := range []string{
		`h('text', { class: 'panel-title' }, '历史绩效列表')`,
		`h('text', { class: 'count-pill' }, String(historyRows.length))`,
	} {
		if strings.Contains(historyBody, duplicateSnippet) {
			t.Fatalf("history table should not render duplicate panel header snippet %q", duplicateSnippet)
		}
	}
	for _, forbidden := range []string{
		`目标月份`,
		`review.nextPeriod || '-'`,
	} {
		if strings.Contains(historyBody, forbidden) {
			t.Fatalf("history table should not render target month field %q", forbidden)
		}
	}
}

func TestDingTalkH5HrbpReviewUsesTabTable(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(pageSrc) + string(workbenchSrc) + string(styleSrc)
	for _, snippet := range []string{
		`const hrbpReviewTab = ref('pending')`,
		`if (ctx.contentView.value === 'hrbp')`,
		`const pendingRows = ctx.currentReviews.value.filter((review) => review.status === 'hrbp_review')`,
		`const reviewedRows = ctx.currentReviews.value.filter((review) => ['employee_confirm', 'hr_final', 'completed'].includes(review.status))`,
		`class: 'hrbp-review-tabs-bar'`,
		`class: 'hrbp-review-tabs'`,
		`'待评'`,
		`'已评'`,
		`class: 'hrbp-review-table summary-table'`,
		`class: 'hrbp-empty-row'`,
		`{ key: 'actions', label: '操作' }`,
		`column.key === 'actions' ? 'table-actions-cell' : ''`,
		`function renderHrbpReviewCell(ctx, column, review, onOpenRow)`,
		`ctx.canHrbpHandle(review) ? 'dt-btn-primary' : 'dt-btn-light'`,
		`onOpenRow(review)`,
		`'评价' : '查看'`,
		`['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(item.status)`,
		`.hrbp-review-tabs-bar`,
		`.hrbp-review-tabs`,
		`.hrbp-review-table`,
		`justify-content: flex-start;`,
		`display: inline-flex;`,
		`.hrbp-review-tab::after`,
		`.hrbp-review-tab.active::after`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("hrbp review table view must include %q", snippet)
		}
	}
	hrbpTabsRule := cssRule(string(styleSrc), ".hrbp-review-tabs")
	if strings.Contains(hrbpTabsRule, `grid-template-columns: repeat(2, minmax(0, 1fr));`) {
		t.Fatalf("hrbp tabs should not stretch into two full-width grid columns")
	}
	tabsBarRule := cssRule(string(styleSrc), ".hrbp-review-tabs-bar")
	if !strings.Contains(tabsBarRule, `border-bottom: 1px solid #e5e6eb;`) {
		t.Fatalf("hrbp tabs bar should be a standalone bottom-border tab strip")
	}
	hrbpIndex := strings.Index(string(workbenchSrc), `if (ctx.contentView.value === 'hrbp')`)
	detailIndex := strings.Index(string(workbenchSrc), `class: 'review-detail-only'`)
	if hrbpIndex < 0 || detailIndex < 0 || hrbpIndex > detailIndex {
		t.Fatalf("hrbp table branch must run before default review detail branch")
	}
	pendingTabIndex := strings.Index(string(workbenchSrc), `ctx.hrbpReviewTab.value === 'pending' ? 'active' : ''], onClick: () => ctx.switchHrbpReviewTab('pending') }, '待评')`)
	reviewedTabIndex := strings.Index(string(workbenchSrc), `ctx.hrbpReviewTab.value === 'reviewed' ? 'active' : ''], onClick: () => ctx.switchHrbpReviewTab('reviewed') }, '已评')`)
	if pendingTabIndex < 0 || reviewedTabIndex < 0 || pendingTabIndex > reviewedTabIndex {
		t.Fatalf("hrbp review tabs must render pending on the left before reviewed")
	}
	tabsIndex := strings.Index(string(workbenchSrc), `class: 'hrbp-review-tabs'`)
	tabsBarIndex := strings.Index(string(workbenchSrc), `class: 'hrbp-review-tabs-bar'`)
	panelIndex := strings.Index(string(workbenchSrc)[hrbpIndex:], `h('section', { class: 'panel' }`)
	tableIndex := strings.Index(string(workbenchSrc), `class: 'hrbp-table-wrap table-wrap'`)
	if tabsBarIndex < 0 || tabsIndex < 0 || tableIndex < 0 || panelIndex < 0 || tabsBarIndex > hrbpIndex+panelIndex || tabsIndex > tableIndex {
		t.Fatalf("hrbp tabs must render as a standalone strip before the table panel")
	}
	for _, legacySnippet := range []string{
		`class: 'hrbp-table-toolbar'`,
		`h('text', { class: 'panel-title' }, 'HRBP评价列表')`,
		`h('text', { class: 'count-pill' }, String(rows.length))`,
		`h('text', { class: 'count-pill' }, String(reviewedRows.length))`,
		`h('text', { class: 'count-pill' }, String(pendingRows.length))`,
		`.hrbp-table-toolbar`,
		".hrbp-review-tab.active {\n  border-color: #91caff;",
	} {
		if strings.Contains(combined, legacySnippet) {
			t.Fatalf("hrbp review tab style should not include legacy snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5ManagerReviewUsesTabTable(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(pageSrc) + string(workbenchSrc) + string(styleSrc)
	for _, snippet := range []string{
		`'performance:manager'`,
		`manager: '上级评价'`,
		`const managerReviewTab = ref('pending')`,
		`if (ctx.contentView.value === 'manager')`,
		`const pendingRows = ctx.currentReviews.value.filter((review) => review.status === 'manager_review')`,
		`const reviewedRows = ctx.currentReviews.value.filter((review) => ['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(review.status))`,
		`class: 'workbench hrbp-review-page manager-review-page'`,
		`class: 'hrbp-review-tabs-bar'`,
		`class: 'hrbp-review-tabs'`,
		`ctx.managerReviewTab.value === 'pending' ? 'active' : ''], onClick: () => ctx.switchManagerReviewTab('pending') }, '待评')`,
		`ctx.managerReviewTab.value === 'reviewed' ? 'active' : ''], onClick: () => ctx.switchManagerReviewTab('reviewed') }, '已评')`,
		`class: 'hrbp-review-table summary-table'`,
		`class: 'hrbp-empty-row'`,
		`{ key: 'position', label: '岗位' }`,
		`{ key: 'actions', label: '操作' }`,
		`column.key === 'actions' ? 'table-actions-cell' : ''`,
		`ctx.canManager(review) ? '评价' : '查看'`,
		`ctx.selectReview(review.id)`,
		`ctx.reviewTab.value = 'manager'`,
		`const selectedManagerReview = ctx.selectedReviewId.value ? ctx.selectedReview.value : null`,
		`if (selectedManagerReview) {`,
		`class: 'workbench manager-review-detail-page'`,
		`ctx.selectedReviewId.value = ''`,
		`'返回列表'`,
		`.detail-page-toolbar`,
		`h(ReviewForm, { review: selectedManagerReview })`,
		`class: 'panel detail-panel manager-review-detail-panel'`,
		`['manager_review', 'hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(item.status)`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("manager review table view must include %q", snippet)
		}
	}
	managerIndex := strings.Index(string(workbenchSrc), `if (ctx.contentView.value === 'manager')`)
	detailIndex := strings.Index(string(workbenchSrc), `class: 'review-detail-only'`)
	if managerIndex < 0 || detailIndex < 0 || managerIndex > detailIndex {
		t.Fatalf("manager table branch must run before default review detail branch")
	}
	managerBody := string(workbenchSrc)[managerIndex:detailIndex]
	for _, forbidden := range []string{
		`目标月份`,
		`review.nextPeriod || '-'`,
		`selectedManagerReview ? h('section', { class: 'panel detail-panel manager-review-detail-panel' }`,
	} {
		if strings.Contains(managerBody, forbidden) {
			t.Fatalf("manager review table should not render target month field %q", forbidden)
		}
	}
}

func TestDingTalkH5PcTablePagesRenderPagination(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench: %v", err)
	}
	summarySrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/SummaryView.js")
	if err != nil {
		t.Fatalf("read summary view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	paginationSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/tablePagination.js")
	if err != nil {
		t.Fatalf("read table pagination helper: %v", err)
	}
	combined := string(workbenchSrc) + string(summarySrc) + string(styleSrc) + string(paginationSrc)
	for _, snippet := range []string{
		`function useTablePagination(rowsRef`,
		`function renderTablePagination(pagination`,
		`class: 'table-pagination'`,
		`table-pagination-btn`,
		`const historyPagination = useTablePagination(historyTableRows)`,
		`const hrbpPagination = useTablePagination(hrbpTableRows)`,
		`const managerPagination = useTablePagination(managerTableRows)`,
		`const summaryPagination = useTablePagination(summaryTableRows)`,
		`renderTablePagination(historyPagination)`,
		`renderTablePagination(hrbpPagination)`,
		`renderTablePagination(managerPagination)`,
		`renderTablePagination(summaryPagination)`,
		`.table-pagination {`,
		`.table-pagination-actions`,
		`.table-pagination-btn`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("pc table pages must render shared pagination, missing %q", snippet)
		}
	}
}

func TestDingTalkH5ReviewListTabsDoNotAutoOpenDetail(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	page := string(pageSrc)
	for _, snippet := range []string{
		`function shouldAutoSelectReview(options = {})`,
		`if (options.autoSelectReview === false) return false`,
		`return contentView.value === 'mine'`,
		`async function loadReviews(params = {}, options = {})`,
		`ensureSelectedReview(options)`,
		`await loadReviews({}, options)`,
		`async function switchManagerReviewTab(tab) {
  if (managerReviewTab.value === tab) return
  managerReviewTab.value = tab
  selectedReviewId.value = ''
  await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
}`,
		`async function switchHrbpReviewTab(tab) {
  if (hrbpReviewTab.value === tab) return
  hrbpReviewTab.value = tab
  selectedReviewId.value = ''
  await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
}`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("review list tabs must avoid auto-opening detail; missing %q", snippet)
		}
	}
}

func TestDingTalkH5TemplateSupportsPermissionGuardedEditing(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	apiSrc, err := os.ReadFile("../../../../../dingtalk-h5/services/dingtalkH5Api.js")
	if err != nil {
		t.Fatalf("read h5 api: %v", err)
	}
	templateSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/TemplateView.js")
	if err != nil {
		t.Fatalf("read template view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(pageSrc) + string(apiSrc) + string(templateSrc) + string(styleSrc)
	for _, snippet := range []string{
		`saveTemplate(data)`,
		"return put(`${DINGTALK_H5_API}/template`, data)",
		`async function saveTemplate(data)`,
		`state.template = res.data`,
		`saveTemplate,`,
		`const canEdit = () => ctx.canEditTemplate()`,
		`const editing = ref(false)`,
		`template-panel-actions`,
		`template-edit-row`,
		`template-editor-textarea`,
		`template-editor-input`,
		`template-delete-btn`,
		`.template-delete-btn`,
		`template-rubric-description`,
		`template-rubric-preview`,
		`white-space: nowrap;`,
		`.template-grid.editing`,
		`grid-template-columns: minmax(0, 1fr) auto;`,
		`grid-template-columns: minmax(0, 1fr) minmax(70px, 90px) minmax(70px, 90px) auto;`,
		`grid-template-columns: minmax(120px, 160px) minmax(70px, 88px) minmax(260px, 1fr) auto;`,
		`ctx.saveTemplate(draft.value)`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("template edit flow must include %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`h('button', { class: 'dt-btn dt-btn-danger-light small', onClick: onRemove }, '删除')`,
		`h('button', { class: 'dt-btn dt-btn-danger-light small', onClick: () => removeItem(item.rubric, rubricIndex) }, '删除')`,
	} {
		if strings.Contains(string(templateSrc), legacySnippet) {
			t.Fatalf("template delete buttons should use compact template style instead of legacy snippet %q", legacySnippet)
		}
	}
	editIndex := strings.Index(string(templateSrc), `canEdit() && !editing.value ? h('button', { class: 'dt-btn dt-btn-primary', onClick: startEdit }, '编辑') : null`)
	refreshIndex := strings.Index(string(templateSrc), `!editing.value ? h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新') : null`)
	if editIndex < 0 || refreshIndex < 0 || editIndex > refreshIndex {
		t.Fatalf("template toolbar should render edit button before refresh button")
	}
}

func TestDingTalkH5TemplateTitleUsesActiveMenuLabel(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	templateSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/TemplateView.js")
	if err != nil {
		t.Fatalf("read template view: %v", err)
	}
	for _, snippet := range []string{
		`activeMenuItem.value?.label || titleForContent(contentView.value)`,
		`sectionTitle,`,
	} {
		if !strings.Contains(string(pageSrc), snippet) {
			t.Fatalf("page context must expose active menu label through sectionTitle, missing %q", snippet)
		}
	}
	if !strings.Contains(string(templateSrc), `h('text', { class: 'page-title' }, ctx.sectionTitle.value)`) {
		t.Fatalf("template page title must follow current menu label from permissions")
	}
	if strings.Contains(string(templateSrc), `h('text', { class: 'page-title' }, '绩效模版')`) {
		t.Fatalf("template page title must not be hardcoded because admin can rename the menu")
	}
}

func TestDingTalkH5PageTitlesFollowActiveMenuLabel(t *testing.T) {
	files := map[string]string{
		"template":  "../../../../../dingtalk-h5/components/performance/TemplateView.js",
		"summary":   "../../../../../dingtalk-h5/components/performance/SummaryView.js",
		"workbench": "../../../../../dingtalk-h5/components/performance/WorkbenchView.js",
		"org":       "../../../../../dingtalk-h5/components/performance/OrgView.vue",
		"account":   "../../../../../dingtalk-h5/components/performance/AccountView.vue",
	}
	for name, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s view: %v", name, err)
		}
		body := string(src)
		if !strings.Contains(body, "ctx.sectionTitle.value") {
			t.Fatalf("%s page title must use active menu label from ctx.sectionTitle", name)
		}
		for _, hardcoded := range []string{
			`class="page-title">流程执行`,
			`class="page-title">绩效模版`,
			`class: 'page-title' }, '流程执行'`,
			`class: 'page-title' }, '绩效模版'`,
			`class: 'page-title' }, 'HRBP汇总'`,
			`class: 'page-title' }, '工作台'`,
			`class: 'page-title' }, '我的绩效'`,
			`class: 'page-title' }, '历史绩效'`,
			`class: 'page-title' }, '上级评价'`,
			`class: 'page-title' }, 'HRBP评价'`,
		} {
			if strings.Contains(body, hardcoded) {
				t.Fatalf("%s page title should not hardcode menu label snippet %q", name, hardcoded)
			}
		}
	}
}

func TestDingTalkH5ActionsUseUnifiedPermissionsInsteadOfRoleNames(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	summarySrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/SummaryView.js")
	if err != nil {
		t.Fatalf("read summary view: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews service: %v", err)
	}
	usersSrc, err := os.ReadFile("users.go")
	if err != nil {
		t.Fatalf("read users service: %v", err)
	}
	page := string(pageSrc)
	summary := string(summarySrc)
	for _, snippet := range []string{
		`function hasMenuPermission(key)`,
		`function canCreateReview()`,
		`function canDeleteReview()`,
		`function canExportReviews()`,
		`hasApiPermission('dingtalk_h5:api:review:create')`,
		`hasApiPermission('dingtalk_h5:api:review:delete')`,
		`hasApiPermission('dingtalk_h5:api:review:export')`,
		`hasApiPermission('dingtalk_h5:api:user:list')`,
		`hasApiPermission('dingtalk_h5:api:template:view')`,
		`'delete-review': {`,
		`confirmReviewAction('delete-review')`,
		`toast(error?.msg || '删除失败')`,
		`hasMenuPermission,`,
		`canCreateReview,`,
		`canDeleteReview,`,
		`canExportReviews,`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("dingtalk h5 page should use unified permissions with %q", snippet)
		}
	}
	for _, snippet := range []string{
		`ctx.canExportReviews()`,
		`SUMMARY_TABLE_COLUMNS.filter((column) => column.key !== 'actions' || ctx.canDeleteReview())`,
		`function handleDeleteReview(event, review)`,
		`await ctx.deleteReview(review.id)`,
		`onClick: (event) => handleDeleteReview(event, review)`,
	} {
		if !strings.Contains(summary, snippet) {
			t.Fatalf("summary actions should use permission helpers with %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`ctx.canCreateReview()`,
		`ctx.openCreateReviewDialog`,
		`['admin', 'hrbp', 'hrbp_manager'].includes(ctx.state.user.role)`,
		`ctx.state.user.role === 'admin'`,
	} {
		if strings.Contains(summary, forbidden) {
			t.Fatalf("summary actions must not use role hardcode %q", forbidden)
		}
	}
	for _, forbidden := range []string{
		`if !canCreateReview(user)`,
		`if !isAdmin(user)`,
		`if !isAdmin(current)`,
	} {
		if strings.Contains(string(reviewsSrc)+string(usersSrc), forbidden) {
			t.Fatalf("dingtalk h5 services should trust unified ApiPerm gate instead of %q", forbidden)
		}
	}
}

func TestDingTalkH5CreateReviewUsesButtonPermissionAndDepartmentMultiSelect(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(pageSrc) + string(workbenchSrc) + string(styleSrc)
	for _, snippet := range []string{
		`buttonPermissionKeys`,
		`function hasButtonPermission(key)`,
		`hasButtonPermission('dingtalk_h5:button:review:create')`,
		`createReviewForm.employeeIds`,
		`async function openCreateReviewDialog()`,
		`employeeIds: createReviewForm.employeeIds`,
		`ctx.openCreateReviewDialog`,
		`review-create-modal`,
		`department-user-tree`,
		`const createTargetUserTreeRows = computed(() => flattenCreateTargetTree(createTargetUserTree.value, createReviewExpandedDeptKeys.value))`,
		`function buildCreateTargetUserTree(users = [])`,
		`function flattenCreateTargetTree(nodes = [], expandedKeys, depth = 1)`,
		`function createTargetDepartmentUserIds(row)`,
		`function createTargetDepartmentCheckState(row)`,
		`function toggleCreateReviewDepartment(row)`,
		`function createTargetDepartmentLevels(user)`,
		`v-for="row in createTargetUserTreeRows"`,
		`row.type === 'department'`,
		`row.type === 'employee'`,
		`@click.stop="toggleCreateReviewDepartment(row)"`,
		`createTargetDepartmentCheckState(row)`,
		`create-target-dept-check`,
		`create-target-dept-check-indeterminate`,
		`create-target-user-tree`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("dingtalk h5 create review must use button permission and department multi-select with %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`const createTargetUserTree = computed(() => departmentUserTree(createReviewTargetUsers.value))`,
		`function departmentUserTree(users = [])`,
		`const groups = new Map()`,
		`v-for="dept in createTargetUserTree"`,
		`:key="dept.key"`,
		`dept.users.length`,
		`v-for="user in dept.users"`,
	} {
		if strings.Contains(combined, legacySnippet) {
			t.Fatalf("create review target selector should not keep flat department group snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5CreateReviewOnlyShowsPeriodMonthPicker(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	page := string(pageSrc)
	createForm := page
	if start := strings.Index(createForm, `<view class="review-create-form">`); start >= 0 {
		createForm = createForm[start:]
		if end := strings.Index(createForm, `<view class="review-create-actions">`); end >= 0 {
			createForm = createForm[:end]
		}
	}
	for _, snippet := range []string{
		`const createReviewForm = reactive({ employeeIds: [], period: currentMonth() })`,
		`<view class="review-create-inline-fields single">`,
		`<text class="review-create-label">考评月份</text>`,
		`<input class="field-input" type="month" v-model="createReviewForm.period" />`,
		`function nextMonthFromPeriod(period)`,
		`nextPeriod: nextMonthFromPeriod(createReviewForm.period)`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("create review should keep only period month picker with %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`<text class="review-create-label">目标月份</text>`,
		`v-model="createReviewForm.nextPeriod"`,
		`createReviewForm.nextPeriod`,
		`nextPeriod: nextMonth()`,
	} {
		if strings.Contains(createForm, legacySnippet) || strings.Contains(page, `createReviewForm.nextPeriod`) {
			t.Fatalf("create review form should not expose target month field snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5CreateReviewUsesPreviousNextObjectives(t *testing.T) {
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	helpersSrc, err := os.ReadFile("review_helpers.go")
	if err != nil {
		t.Fatalf("read review_helpers.go: %v", err)
	}
	combined := string(reviewsSrc) + string(helpersSrc)
	for _, snippet := range []string{
		`loadPreviousNextObjectivesForCreate(ctx, db, employee.Account, period)`,
		`currentObjectivesForNewReview(reviewNo, tpl.ObjectiveDefaults, previousReview)`,
		`ObjectiveSourceReviewNo: objectiveSource.reviewNo`,
		`ObjectiveSourcePeriod:   objectiveSource.period`,
		`ObjectivesJSON:          encodeJSON(objectives)`,
		"`employee_account` = ? AND `next_period` = ?",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("create review should inherit previous next objectives with snippet %q", snippet)
		}
	}
}

func TestDingTalkH5OrgConfigUsesButtonPermission(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	orgSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/OrgView.vue")
	if err != nil {
		t.Fatalf("read org view: %v", err)
	}
	combined := string(pageSrc) + string(orgSrc)
	for _, snippet := range []string{
		`function canEditUsers()`,
		`hasButtonPermission('dingtalk_h5:button:user:config')`,
		`hasApiPermission('dingtalk_h5:api:user:edit')`,
		`canEditUsers,`,
		`const canEditUsers = computed(() => ctx.canEditUsers())`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("org config button must use unified button and api permission with %q", snippet)
		}
	}
	if strings.Contains(string(orgSrc), `const canEditUsers = computed(() => ctx.hasApiPermission('dingtalk_h5:api:user:edit'))`) {
		t.Fatalf("org config button must not be visible from api permission alone")
	}
}

func TestDingTalkH5DeleteReviewUsesSoftDelete(t *testing.T) {
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	helpersSrc, err := os.ReadFile("review_helpers.go")
	if err != nil {
		t.Fatalf("read review_helpers.go: %v", err)
	}
	auditSrc, err := os.ReadFile("audit.go")
	if err != nil {
		t.Fatalf("read audit.go: %v", err)
	}
	combined := string(reviewsSrc) + string(helpersSrc) + string(auditSrc)
	for _, snippet := range []string{
		`func notDeletedReviewQuery(db *gorm.DB) *gorm.DB`,
		"`deleted_at` = 0",
		`func dingtalkH5DeleteAuditUpdateValues(meta dingtalkH5AuditMeta) map[string]interface{}`,
		`dingtalkH5DeleteAuditUpdateValues(audit)`,
		`addHistoryWithDB(tx, &review, user, "删除考评单")`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("delete review should soft delete and filter deleted rows with %q", snippet)
		}
	}
	for _, hardDelete := range []string{
		`Delete(&model.DingTalkH5PerfHistory{})`,
		`tx.Delete(&review)`,
		`Delete(&model.DingTalkH5PerfReview{})`,
	} {
		if strings.Contains(string(reviewsSrc), hardDelete) {
			t.Fatalf("delete review must not hard delete records with %q", hardDelete)
		}
	}
}

func TestDingTalkH5DeleteUserDisablesSharedUser(t *testing.T) {
	usersSrc, err := os.ReadFile("users.go")
	if err != nil {
		t.Fatalf("read users.go: %v", err)
	}
	storeSrc, err := os.ReadFile("user_store.go")
	if err != nil {
		t.Fatalf("read user_store.go: %v", err)
	}
	combined := string(usersSrc) + string(storeSrc)
	for _, snippet := range []string{
		`"user_status":    0`,
		"`user_status` = 1",
		`Where("` + "`user_mini_openid` = ?" + `", target.Account).Updates(updates)`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("delete user should disable shared users row and list active users with %q", snippet)
		}
	}
	for _, hardDelete := range []string{
		`Delete(&model.User{})`,
		`userHasReferences(db, target.Account)`,
	} {
		if strings.Contains(string(usersSrc), hardDelete) {
			t.Fatalf("delete user must not hard delete or block soft delete with %q", hardDelete)
		}
	}
}

func TestCreateReviewSupportsBatchAndEmployeeAuditOwner(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	handlerSrc, err := os.ReadFile("../../../handler/dingtalkh5/performance/review/handler.go")
	if err != nil {
		t.Fatalf("read review handler.go: %v", err)
	}
	combined := string(typesSrc) + string(reviewsSrc) + string(handlerSrc)
	for _, snippet := range []string{
		`EmployeeIDs`,
		"`json:\"employeeIds\"`",
		`type CreateReviewBatchResponse struct`,
		`func CreateReviewsContext(ctx context.Context, user *model.DingTalkH5PerfUser, payload ReviewPayload)`,
		`reviewPayloadEmployeeIDs(payload)`,
		`accessScope, err := createReviewAccessScopeContext(ctx, db, user)`,
		`createReviewForEmployeeContext(ctx, db, user, accessScope, employeeAccount, period, nextPeriod, tpl)`,
		`func createReviewAccessScopeContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) (createReviewAccessScope, error)`,
		`scope, err := permissionsupport.DataScopeContext(ctx, db, user.ID, user.RoleID)`,
		`allowed, all, err := dataScopeUserAccountsContext(ctx, db, user, scope)`,
		`ownerAudit := dingtalkH5AuditMetaForUserContext(ctx, db, employee, now)`,
		`operatorAudit := dingtalkH5AuditMetaForUserContext(ctx, db, user, now)`,
		`response.JSON(c, data)`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("create review must support batch targets and employee-owned create audit with %q", snippet)
		}
	}
	createOneBody := string(reviewsSrc)
	if start := strings.Index(createOneBody, "func createReviewForEmployeeContext"); start >= 0 {
		createOneBody = createOneBody[start:]
		if end := strings.Index(createOneBody, "\n}\n\nfunc "); end >= 0 {
			createOneBody = createOneBody[:end+3]
		}
	}
	if strings.Contains(createOneBody, "canAccessPerfUserAccountContext(ctx, db, user, employee.Account)") {
		t.Fatalf("batch create should not recompute data scope for each employee")
	}
}

func TestDingTalkH5CurrentObjectivesSupportAddAndDelete(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(workbenchSrc) + string(styleSrc)
	for _, snippet := range []string{
		`function createObjective`,
		`function addCurrentObjective(review)`,
		`function removeCurrentObjective(review, index)`,
		`function confirmObjectiveDelete(index)`,
		`function confirmRemoveCurrentObjective(review, index, onRemoved)`,
		`uni.showModal`,
		`confirmText: '删除'`,
		`ctx.canEditObjectiveDimension(props.review)`,
		`const showDeleteActions = ref(false)`,
		`function toggleDeleteActions()`,
		`onClick: () => addCurrentObjective(props.review)`,
		`onClick: toggleDeleteActions`,
		`showDeleteActions.value ? '隐藏删除' : '显示删除'`,
		`showDeleteActions.value && editableObjectives`,
		`confirmRemoveCurrentObjective(props.review, index`,
		`'增加目标'`,
		`'删除'`,
		`class: 'section-actions'`,
		`class: 'objective-head-actions'`,
		`class: 'objective-empty'`,
		`class: ['dt-btn small objective-delete-toggle'`,
		`class: 'dt-btn dt-btn-danger-light small objective-delete-btn'`,
		`.section-actions`,
		`.objective-head-actions`,
		`.objective-empty`,
		`.objective-delete-toggle.active`,
		`.objective-delete-btn`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("current objective add/delete flow must include %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`editableObjectives ? h('button', { class: 'dt-btn dt-btn-danger-light small', onClick: () => removeCurrentObjective(props.review, index) }, '删除') : null`,
		`window.confirm('确认删除目标`,
	} {
		if strings.Contains(combined, legacySnippet) {
			t.Fatalf("current objective delete flow should not keep legacy snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5CurrentReviewSubmitRequiresConfirmAndWithdrawRequiresReason(t *testing.T) {
	src, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read index page: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`const reviewActionConfirmCopy = {`,
		`'submit-self': {`,
		`title: '提交自评'`,
		`content: '确认提交当前绩效？提交后将进入上级评价流程。'`,
		`async function confirmReviewAction(action)`,
		`uni.showModal({`,
		`const confirmed = await confirmReviewAction(action)`,
		`if (!confirmed) return`,
		`const withdrawDialog = reactive({ visible: false, loading: false, reason: '' })`,
		`const returnDialog = reactive({`,
		`class="review-action-modal"`,
		`v-model="withdrawDialog.reason"`,
		`v-model="returnDialog.reason"`,
		`async function submitWithdrawReview()`,
		`async function submitReturnReview()`,
		`toast('请填写撤回理由')`,
		`toast('请填写退回原因')`,
		`returnReason: reason`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("submit confirm and withdraw reason flow should include snippet %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`window.confirm('确认撤销提交？撤销后会回到上一阶段，可重新编辑。')`,
		`const confirmed = await confirmReviewAction('withdraw')`,
		"window.prompt(`${label}原因`)",
	} {
		if strings.Contains(text, legacySnippet) {
			t.Fatalf("withdraw reason flow should not keep legacy snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5ProcessModalShowsWithdrawReason(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(workbenchSrc) + string(styleSrc)
	for _, snippet := range []string{
		`function historyReasonForStep(review, step, selectedHistory, actionParts)`,
		`if (actionParts.reason) return actionParts.reason`,
		`historyActionNeedsReason(selectedTitle) ? missingHistoryReasonText(selectedTitle) : ''`,
		`return '未记录撤销理由'`,
		`function historyReasonLabel(title)`,
		`return '撤销理由'`,
		`reason,`,
		`reasonLabel,`,
		`step.reason ? h('view', { class: 'process-step-reason' }`,
		`h('text', { class: 'process-step-reason-label' }, step.reasonLabel || '理由')`,
		`.process-step-reason`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("process modal should show withdraw reason with snippet %q", snippet)
		}
	}
}

func TestDingTalkH5ManagerReviewShowsPendingTextWhenEmpty(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	workbench := string(workbenchSrc)
	combined := workbench + string(styleSrc)
	for _, snippet := range []string{
		`function hasManagerEvaluation(review)`,
		`function reviewGradeBadge(label, grade)`,
		`function managerReviewTitleMeta(ctx, review)`,
		"`上级：${name}`",
		`'上级未设置'`,
		`const managerPendingText = '上级待评'`,
		`const managerPending = !hasManagerEvaluation(props.review)`,
		`h('section', { class: 'form-section review-evaluation-section manager-review-section' }, [`,
		`h('text', { class: 'section-meta' }, managerReviewTitleMeta(ctx, props.review))`,
		`reviewGradeBadge('上级分档', props.review.managerGrade)`,
		`props.editable ? h('view', { class: 'field-block manager-grade-field' }`,
		`h('text', { class: 'field-label' }, '上级分档')`,
		`class: props.editable ? 'form-grid manager-review-grid manager-review-block' : 'manager-review-readonly manager-review-block'`,
		`placeholder: managerPending && !props.editable ? managerPendingText : '填写上级评价'`,
		`.section-meta`,
		`.review-grade-badge`,
		`.manager-review-block`,
		`.review-evaluation-section > .section-title`,
		`.review-evaluation-section > .manager-review-block,`,
		`.review-evaluation-section > .hrbp-review-block {`,
		`margin-top: 10px;`,
		`padding: 0 20px;`,
		`box-sizing: border-box;`,
		`.review-evaluation-section > .section-title {`,
		`flex-wrap: nowrap;`,
		`.review-evaluation-section .section-title-main {`,
		`width: auto;`,
		`.review-evaluation-section .review-grade-badge {`,
		`white-space: nowrap;`,
		`.manager-review-readonly`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("manager review pending display must include %q", snippet)
		}
	}
	legacySnippet := `grade || (managerPending && !props.editable ? managerPendingText : '未选择')`
	if strings.Contains(workbench, legacySnippet) {
		t.Fatalf("readonly manager pending state should not render the old grade select placeholder %q", legacySnippet)
	}
	managerGridRule := cssRule(string(styleSrc), ".manager-review-grid")
	if !strings.Contains(managerGridRule, `grid-template-columns: 1fr;`) {
		t.Fatalf("editable manager review area should stack grade and comment vertically on PC, got %q", managerGridRule)
	}
	managerGradeRule := cssRule(string(styleSrc), ".manager-review-grid .manager-grade-field")
	if !strings.Contains(managerGradeRule, `width: min(220px, 100%);`) {
		t.Fatalf("manager grade selector should keep a compact width in vertical layout, got %q", managerGradeRule)
	}
}

func TestDingTalkH5HrbpReviewShowsReviewerNameInTitle(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	workbench := string(workbenchSrc)
	combined := workbench + string(styleSrc)
	for _, snippet := range []string{
		`function hasHrbpEvaluation(review)`,
		`function reviewGradeBadge(label, grade)`,
		`function hrbpReviewTitleMeta(ctx, review)`,
		`ctx.userName(review?.hrbpReviewerId || review?.hrbpId)`,
		"`HRBP：${name}`",
		`'HRBP未设置'`,
		`const hrbpPendingText = 'HRBP待评'`,
		`const hrbpPending = !hasHrbpEvaluation(props.review)`,
		`h('section', { class: 'form-section review-evaluation-section hrbp-review-section' }, [`,
		`h('text', { class: 'section-meta' }, hrbpReviewTitleMeta(ctx, props.review))`,
		`reviewGradeBadge('HRBP分档', props.review.hrbpGrade)`,
		`props.editable ? h('view', { class: 'field-block hrbp-grade-field' }`,
		`h('text', { class: 'field-label' }, 'HRBP分档')`,
		`class: props.editable ? 'form-grid hrbp-review-grid hrbp-review-block' : 'manager-review-readonly hrbp-review-block'`,
		`placeholder: hrbpPending && !props.editable ? hrbpPendingText : '填写 HRBP 评价'`,
		`.hrbp-review-block`,
		`.review-grade-badge`,
		`.review-evaluation-section > .section-title`,
		`.review-evaluation-section > .manager-review-block,`,
		`.review-evaluation-section > .hrbp-review-block {`,
		`margin-top: 10px;`,
		`padding: 0 20px;`,
		`box-sizing: border-box;`,
		`.review-evaluation-section > .section-title {`,
		`flex-wrap: nowrap;`,
		`.review-evaluation-section .section-title-main {`,
		`width: auto;`,
		`.review-evaluation-section .review-grade-badge {`,
		`white-space: nowrap;`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("hrbp review title should show reviewer name and hide readonly grade select with %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`h('view', { class: 'section-title' }, [h('text', {}, 'HRBP评价')])`,
		`h('select', { class: 'field-select', value: props.review.hrbpGrade, disabled: !props.editable`,
	} {
		if strings.Contains(workbench, legacySnippet) {
			t.Fatalf("hrbp review should not keep legacy title or always-visible grade select %q", legacySnippet)
		}
	}
	hrbpGridRule := cssRule(string(styleSrc), ".hrbp-review-grid")
	if !strings.Contains(hrbpGridRule, `grid-template-columns: 1fr;`) {
		t.Fatalf("editable hrbp review area should stack grade and comment vertically on PC, got %q", hrbpGridRule)
	}
	hrbpGradeRule := cssRule(string(styleSrc), ".hrbp-review-grid .hrbp-grade-field")
	if !strings.Contains(hrbpGradeRule, `width: min(220px, 100%);`) {
		t.Fatalf("hrbp grade selector should keep a compact width in vertical layout, got %q", hrbpGradeRule)
	}
}

func TestDingTalkH5ReviewFormUsesSectionTabsInsteadOfLongList(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	workbench := string(workbenchSrc)
	combined := workbench + string(styleSrc) + string(pageSrc)
	for _, snippet := range []string{
		`const reviewFormTabs = [`,
		`['currentTargets', '本月目标']`,
		`['selfSummary', '思考总结']`,
		`['selfValues', '价值观自评']`,
		`['manager', '上级评价']`,
		`['hrbp', 'HRBP评价']`,
		`['nextTargets', '下月目标']`,
		`function normalizeReviewFormTab(ctx, review)`,
		`function renderReviewFormPane(ctx, review, editableSelf, editableManager, editableHrbp, readonly = false)`,
		`class: 'review-form-pane'`,
		`ctx.reviewTab.value = item.key`,
		`h(CurrentObjectiveSection, { review, editableSelf, readonly })`,
		`h(SelfSummarySection, { review, editableSelf })`,
		`h(ValueSection, { review, field: 'self', title: '价值观自评', editable: editableSelf })`,
		`h(ManagerSection, { review, editable: editableManager })`,
		`h(HrbpSection, { review, editable: editableHrbp })`,
		`h(NextObjectiveSection, { review, editable: editableSelf })`,
		`const reviewTab = ref('currentTargets')`,
		`reviewTab.value = 'currentTargets'`,
		`class: 'value-grid value-list'`,
		`const rubrics = valueRubricItems(tpl, item)`,
		`class: 'value-title-row'`,
		`class: 'value-score-tag'`,
		`ValueStandardModal`,
		`评分标准`,
		`renderValueRubricList(rubrics)`,
		`class: 'value-standard-modal-mask'`,
		`class: 'value-score-guide'`,
		`.value-list {`,
		`grid-template-columns: 1fr;`,
		`.review-form-pane`,
		`overflow-x: auto;`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("review form tabbed sections must include %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`const activeNext = ctx.reviewTab.value === 'next'`,
		`'当月绩效'`,
		`ctx.reviewTab.value = 'current'`,
		`activeNext`,
	} {
		if strings.Contains(workbench+string(pageSrc), legacySnippet) {
			t.Fatalf("review form should not keep legacy long-list tab snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5SubmitSelfRequiresRequiredSectionsBeforeConfirm(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	text := string(pageSrc)
	for _, snippet := range []string{
		`function selfSubmitRequiredMessage(review)`,
		`function validateSelfSubmitReview(review)`,
		`missing.push('本月目标')`,
		`missing.push('思考总结')`,
		`missing.push('价值观自评')`,
		`missing.push('下月目标')`,
		`return missing.length ? ` + "`请完善：${missing.join('、')}`" + ` : ''`,
		`if (action === 'submit-self') {`,
		`const validationMessage = validateSelfSubmitReview(selectedReview.value)`,
		`if (validationMessage) {`,
		`toast(validationMessage)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("submit self required validation should include %q", snippet)
		}
	}
	validationIndex := strings.Index(text, `const validationMessage = validateSelfSubmitReview(selectedReview.value)`)
	confirmIndex := strings.Index(text, `const confirmed = await confirmReviewAction(action)`)
	requestIndex := strings.Index(text, `dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action`)
	if validationIndex < 0 || confirmIndex < 0 || requestIndex < 0 || validationIndex > confirmIndex || validationIndex > requestIndex {
		t.Fatalf("submit self validation must run before confirm modal and request, validation=%d confirm=%d request=%d", validationIndex, confirmIndex, requestIndex)
	}
}

func TestDingTalkH5ManagerSubmitRequiresRequiredFieldsAndConfirm(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	backendSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	page := string(pageSrc)
	backend := string(backendSrc)
	for _, snippet := range []string{
		`'submit-manager': {`,
		`title: '提交给HRBP'`,
		`content: '确认提交给 HRBP？提交后将进入 HRBP 评价流程。'`,
		`confirmText: '提交'`,
		`'return-employee': {`,
		`title: '退回员工'`,
		`content: '确认退回员工修改？退回后员工可重新编辑自评内容。'`,
		`confirmText: '退回'`,
		`function hasRequiredManagerValues(review)`,
		`function managerSubmitRequiredMessage(review)`,
		`function validateManagerSubmitReview(review)`,
		`missing.push('上级分档')`,
		`missing.push('评价内容')`,
		`missing.push('上级价值观评分')`,
		`return missing.length ? ` + "`请完善：${missing.join('、')}`" + ` : ''`,
		`if (action === 'submit-manager') {`,
		`const managerValidationMessage = validateManagerSubmitReview(selectedReview.value)`,
		`toast(managerValidationMessage)`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("manager submit validation and confirm flow should include %q", snippet)
		}
	}
	for _, snippet := range []string{
		`return fmt.Errorf("请先选择绩效分档")`,
		`return fmt.Errorf("请先填写上级评价")`,
		`return fmt.Errorf("请先填写上级价值观评分")`,
	} {
		if !strings.Contains(backend, snippet) {
			t.Fatalf("backend manager submit guard should include %q", snippet)
		}
	}

	performStart := strings.Index(page, `async function performReviewAction`)
	returnStart := strings.Index(page, `async function returnReview`)
	if performStart < 0 || returnStart < 0 || returnStart <= performStart {
		t.Fatalf("expected performReviewAction before returnReview, perform=%d return=%d", performStart, returnStart)
	}
	performBody := page[performStart:returnStart]
	validationIndex := strings.Index(performBody, `const managerValidationMessage = validateManagerSubmitReview(selectedReview.value)`)
	confirmIndex := strings.Index(performBody, `const confirmed = await confirmReviewAction(action)`)
	requestIndex := strings.Index(performBody, `dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action`)
	if validationIndex < 0 || confirmIndex < 0 || requestIndex < 0 || validationIndex > confirmIndex || validationIndex > requestIndex {
		t.Fatalf("manager submit validation must run before confirm modal and request, validation=%d confirm=%d request=%d", validationIndex, confirmIndex, requestIndex)
	}

	submitReturnStart := strings.Index(page, `async function submitReturnReview`)
	if submitReturnStart < 0 || submitReturnStart <= returnStart {
		t.Fatalf("expected returnReview before submitReturnReview, return=%d submitReturn=%d", returnStart, submitReturnStart)
	}
	withdrawStart := strings.Index(page, `async function withdrawReview`)
	if withdrawStart < 0 || withdrawStart <= submitReturnStart {
		t.Fatalf("expected submitReturnReview before withdrawReview, submitReturn=%d withdraw=%d", submitReturnStart, withdrawStart)
	}
	returnBody := page[returnStart:submitReturnStart]
	if !strings.Contains(returnBody, `openReturnDialog(action, label)`) {
		t.Fatalf("return employee should open the return reason dialog before submitting")
	}
	submitReturnBody := page[submitReturnStart:withdrawStart]
	for _, snippet := range []string{
		`const reason = String(returnDialog.reason || '').trim()`,
		`toast('请填写退回原因')`,
		`returnReason: reason`,
	} {
		if !strings.Contains(submitReturnBody, snippet) {
			t.Fatalf("return employee reason submit should include %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`const confirmed = await confirmReviewAction(action)`,
		"window.prompt(`${label}原因`)",
	} {
		if strings.Contains(returnBody, legacySnippet) || strings.Contains(submitReturnBody, legacySnippet) {
			t.Fatalf("return employee reason flow should not keep legacy snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5HrbpSubmitRequiresRequiredFieldsAndWorkflowConfirm(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	backendSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	page := string(pageSrc)
	backend := string(backendSrc)
	for _, snippet := range []string{
		`'submit-hrbp': {`,
		`title: '提交给员工'`,
		`content: '确认提交给员工确认？提交后员工将查看并确认绩效结果。'`,
		`confirmText: '提交'`,
		`'return-manager': {`,
		`title: '退回上级'`,
		`content: '确认退回上级修改？退回后上级可重新调整评价。'`,
		`confirmText: '退回'`,
		`'return-hrbp': {`,
		`title: '退回HRBP'`,
		`content: '确认退回 HRBP 修改？退回后 HRBP 可重新处理评价。'`,
		`finalize: {`,
		`title: '绩效归档'`,
		`content: '确认归档绩效结果？归档后流程将完成。'`,
		`function hasRequiredHrbpValues(review)`,
		`function hrbpSubmitRequiredMessage(review)`,
		`function hrbpSubmitGradeMismatchMessage(review)`,
		`function validateHrbpSubmitReview(review)`,
		`missing.push('HRBP分档')`,
		`missing.push('评价内容')`,
		`missing.push('HRBP价值观评分')`,
		`return missing.length ? ` + "`请完善：${missing.join('、')}`" + ` : ''`,
		`HRBP分档需与上级分档一致`,
		`function showValidationModal(title, content)`,
		`showCancel: false`,
		`if (action === 'submit-hrbp') {`,
		`const hrbpValidationMessage = validateHrbpSubmitReview(selectedReview.value)`,
		`await showValidationModal(hrbpValidationMessage.title || '提示', hrbpValidationMessage.message)`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("hrbp submit validation and workflow confirm flow should include %q", snippet)
		}
	}
	for _, snippet := range []string{
		`return fmt.Errorf("请先选择 HRBP绩效分档")`,
		`return fmt.Errorf("上级绩效分档为空，不能提交 HRBP 评价")`,
		`return fmt.Errorf("HRBP绩效分档与上级绩效分档不一致，不能提交给员工确认")`,
		`return fmt.Errorf("请先填写 HRBP 评价")`,
		`return fmt.Errorf("请先填写 HRBP 价值观评分")`,
	} {
		if !strings.Contains(backend, snippet) {
			t.Fatalf("backend hrbp submit guard should include %q", snippet)
		}
	}

	performStart := strings.Index(page, `async function performReviewAction`)
	returnStart := strings.Index(page, `async function returnReview`)
	if performStart < 0 || returnStart < 0 || returnStart <= performStart {
		t.Fatalf("expected performReviewAction before returnReview, perform=%d return=%d", performStart, returnStart)
	}
	performBody := page[performStart:returnStart]
	validationIndex := strings.Index(performBody, `const hrbpValidationMessage = validateHrbpSubmitReview(selectedReview.value)`)
	confirmIndex := strings.Index(performBody, `const confirmed = await confirmReviewAction(action)`)
	requestIndex := strings.Index(performBody, `dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action`)
	if validationIndex < 0 || confirmIndex < 0 || requestIndex < 0 || validationIndex > confirmIndex || validationIndex > requestIndex {
		t.Fatalf("hrbp submit validation must run before confirm modal and request, validation=%d confirm=%d request=%d", validationIndex, confirmIndex, requestIndex)
	}
}

func TestDingTalkH5RefreshButtonsReloadCurrentViewDependencies(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	orgSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/OrgView.vue")
	if err != nil {
		t.Fatalf("read org view: %v", err)
	}
	summarySrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/SummaryView.js")
	if err != nil {
		t.Fatalf("read summary view: %v", err)
	}
	templateSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/TemplateView.js")
	if err != nil {
		t.Fatalf("read template view: %v", err)
	}
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	page := string(pageSrc)
	components := string(orgSrc) + string(summarySrc) + string(templateSrc) + string(workbenchSrc)
	for _, snippet := range []string{
		`@click="ctx.refreshData"`,
		`onClick: ctx.refreshData`,
	} {
		if !strings.Contains(components, snippet) {
			t.Fatalf("visible refresh buttons must call ctx.refreshData, missing %q", snippet)
		}
	}
	for _, snippet := range []string{
		`function needsUserDirectoryForContentView()`,
		`return contentView.value === 'summary'`,
		`async function ensureReferenceData(options = {})`,
		`const force = Boolean(options.force)`,
		`const needsUsers = Boolean(options.users)`,
		`const needsTemplate = options.template !== false`,
		`if (needsUsers && (force || !state.users.length) && hasApiPermission('dingtalk_h5:api:user:list')) tasks.push(loadUsers())`,
		`if (needsTemplate && (force || !state.template) && hasApiPermission('dingtalk_h5:api:template:view')) tasks.push(loadTemplate())`,
		`async function refreshData(options = {})`,
		`const forceReference = Boolean(options.forceReference)`,
		`await ensureReferenceData({ force: forceReference, users: needsUserDirectoryForContentView(), template: true })`,
		`return refreshDataSafely({ forceReference: true, contentLoading: true })`,
		`async function refreshWithUserFeedback()`,
		`if (refreshing.value) return false`,
		`showRefreshLoading()`,
		`hideRefreshLoading()`,
		`toast('已刷新')`,
		`toast('刷新失败，请稍后重试')`,
		`refreshData: refreshWithUserFeedback`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("manual refresh should reload current view dependencies with %q", snippet)
		}
	}
	for _, snippet := range []string{
		`function reviewPersonName(review, id)`,
		`function reviewPersonNameFromReviews(id)`,
		`return reviewPersonNameFromReviews(id) || state.users.find((item) => item.id === id)?.name || id || '无'`,
	} {
		if !strings.Contains(page, snippet) {
			t.Fatalf("current performance should read display names from review payload with %q", snippet)
		}
	}
	refreshBody := page
	if start := strings.Index(refreshBody, `async function refreshData(options = {})`); start >= 0 {
		refreshBody = refreshBody[start:]
		if end := strings.Index(refreshBody, "\n}\n\nfunction "); end >= 0 {
			refreshBody = refreshBody[:end+3]
		}
	}
	for _, legacySnippet := range []string{
		`await ensureReferenceData(forceReference)`,
		`Promise.all([loadReviews(), loadUsers()`,
	} {
		if strings.Contains(refreshBody, legacySnippet) {
			t.Fatalf("current performance refresh must not keep user directory load snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5OrgViewUsesPermissionScopedDepartmentTree(t *testing.T) {
	orgSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/OrgView.vue")
	if err != nil {
		t.Fatalf("read org view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	org := string(orgSrc)
	style := string(styleSrc)
	combined := org + style
	for _, snippet := range []string{
		`import { computed, reactive, ref, watch } from 'vue'`,
		`const departmentTree = computed(() => buildDepartmentTree(filteredUsers.value))`,
		`const expandedDepartmentKeys = ref(new Set())`,
		`const treeRows = computed(() => flattenDepartmentTree(departmentTree.value, expandedDepartmentKeys.value))`,
		`watch(departmentTree, (nodes) => syncExpandedDepartments(nodes), { immediate: true })`,
		`v-for="row in treeRows"`,
		`class="org-tree-table"`,
		`class="org-tree-table-head"`,
		`部门名称`,
		`class="org-tree-list"`,
		`org-tree-row`,
		`row.type === 'department'`,
		`row.type === 'employee'`,
		`realDepartmentLevels(user)`,
		`class="org-employee-list"`,
		`class="org-employee-row"`,
		`org-config-btn`,
		`openConfig(user)`,
		`selectedUser`,
		`configForm`,
		`class="org-config-modal-mask"`,
		`class="org-config-modal"`,
		`配置流程审批人`,
		`const activeRelationPicker = ref('')`,
		`const managerPickerTree = computed(() => buildDepartmentTree(managerOptions()))`,
		`const hrbpPickerTree = computed(() => buildDepartmentTree(hrbpOptions.value))`,
		`const managerPickerRows = computed(() => flattenDepartmentTree(managerPickerTree.value, managerPickerExpandedKeys.value))`,
		`const hrbpPickerRows = computed(() => flattenDepartmentTree(hrbpPickerTree.value, hrbpPickerExpandedKeys.value))`,
		`class="relation-picker-trigger"`,
		`class="relation-picker-panel"`,
		`class="relation-picker-tree"`,
		`class="relation-picker-user"`,
		`class="relation-picker-radio"`,
		`toggleRelationPicker('manager')`,
		`toggleRelationPicker('hrbp')`,
		`selectRelationUser('manager', row.user.id)`,
		`selectRelationUser('hrbp', row.user.id)`,
		`toggleRelationDepartment('manager', row)`,
		`toggleRelationDepartment('hrbp', row)`,
		`relationPickerText(configForm.managerId, '无直属上级')`,
		`relationPickerText(configForm.hrbpId, '无HRBP')`,
		`ctx.saveUser(configForm)`,
		`ctx.canEditUsers()`,
		`已按权限过滤`,
		`@click="toggleDepartment(row)"`,
		`org-tree-chevron`,
		`:class="['org-tree-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"`,
		`.org-tree-chevron::before`,
		`.org-tree-chevron.expanded::before`,
		`transform: rotate(-45deg);`,
		`transform: rotate(45deg);`,
		`hasChildren: node.children.length > 0 || node.users.length > 0`,
		`if (!expandedKeys.has(node.key)) continue`,
		`function syncExpandedDepartments(nodes = [])`,
		`const availableKeys = new Set(collectDepartmentKeys(nodes))`,
		`for (const key of expandedDepartmentKeys.value)`,
		`if (availableKeys.has(key)) next.add(key)`,
		`function toggleDepartment(row)`,
		`.org-tree-table`,
		`.org-tree-table-head`,
		`.org-tree-chevron`,
		`.org-tree-row.depth-2 .org-department-title`,
		`.org-tree-row.depth-3 .org-person-cell`,
		`border-bottom: 1px solid #edf0f5;`,
		`border-width: 0 0 1px;`,
		`background: #fafafa;`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("org view must use permission-scoped department tree with %q", snippet)
		}
	}
	for _, selector := range []string{`.org-tree-row::before`, `.org-tree-row::after`} {
		if strings.Contains(style, selector) {
			t.Fatalf("org tree should avoid noisy pseudo connector selector %q", selector)
		}
	}
	for _, ruleCheck := range []struct {
		name      string
		rule      string
		forbidden string
	}{
		{name: "tree list", rule: cssRule(style, ".org-tree-list"), forbidden: "background: #fbfcff;"},
		{name: "employee row", rule: cssRule(style, ".org-employee-row"), forbidden: "border: 1px solid #e8edf5;"},
		{name: "department row", rule: cssRule(style, ".org-department-row"), forbidden: "border-left: 3px solid #1677ff;"},
	} {
		if strings.Contains(ruleCheck.rule, ruleCheck.forbidden) {
			t.Fatalf("org %s should not keep old noisy style %q", ruleCheck.name, ruleCheck.forbidden)
		}
	}
	for _, ruleCheck := range []struct {
		name     string
		rule     string
		required []string
	}{
		{
			name: "config modal",
			rule: cssRule(style, ".org-config-modal"),
			required: []string{
				`height: min(640px, calc(100vh - 48px));`,
				`display: grid;`,
				`grid-template-rows: auto auto minmax(0, 1fr) auto;`,
				`overflow: hidden;`,
			},
		},
		{
			name: "config form",
			rule: cssRule(style, ".org-config-form"),
			required: []string{
				`align-content: start;`,
				`overflow: visible;`,
			},
		},
		{
			name: "relation picker panel",
			rule: cssRule(style, ".relation-picker-panel"),
			required: []string{
				`max-height: min(360px, calc(100vh - 320px));`,
			},
		},
	} {
		for _, required := range ruleCheck.required {
			if !strings.Contains(ruleCheck.rule, required) {
				t.Fatalf("org %s should keep enough modal height and picker room with %q", ruleCheck.name, required)
			}
		}
	}
	for _, legacySnippet := range []string{
		`新增人员`,
		`class="panel create-panel"`,
		`ctx.newUser`,
		`ctx.createUser`,
		`class="org-people-grid"`,
		`class="org-user-card"`,
		`class="org-edit-grid"`,
		`class="org-card-actions"`,
		`v-for="level2 in level1.children"`,
		`v-for="level3 in level2.children"`,
		`未分配二级部门`,
		`未分配小组`,
		`ctx.deleteUser`,
		`<table class="org-table">`,
		`v-for="user in ctx.state.users"`,
		`<select v-model="configForm.managerId"`,
		`<select v-model="configForm.hrbpId"`,
		`class="org-tree-dot"`,
		`class="org-department-path"`,
		`row.expanded ? '⌄' : '›'`,
		`const departmentTree = computed(() => buildDepartmentTree(ctx.state.users))`,
		"for (const key of collectDepartmentKeys(nodes)) {\n    next.add(key)\n  }",
	} {
		if strings.Contains(org, legacySnippet) {
			t.Fatalf("org view should not keep legacy flat user maintenance snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5OrgViewSupportsSearchAndSetupFilters(t *testing.T) {
	orgSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/OrgView.vue")
	if err != nil {
		t.Fatalf("read org view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	org := string(orgSrc)
	style := string(styleSrc)
	combined := org + style
	for _, snippet := range []string{
		`const orgFilters = reactive({`,
		`employeeKeyword: ''`,
		`departmentNames: []`,
		`managerStatus: ''`,
		`hrbpStatus: ''`,
		`const filteredUsers = computed(() => ctx.state.users.filter(matchesOrgFilters))`,
		`const orgFilterDepartmentTree = computed(() => buildDepartmentTree(ctx.state.users))`,
		`const orgFilterDepartmentRows = computed(() => flattenDepartmentSelectionTree(filterDepartmentTree(orgFilterDepartmentTree.value, orgDepartmentSearchKeyword.value), orgFilterDepartmentExpandedKeys.value, orgDepartmentSearchKeyword.value))`,
		`placeholder="搜索员工姓名/账号"`,
		`class="org-filter-bar"`,
		`class="org-department-filter"`,
		`class="org-department-trigger"`,
		`class="org-department-panel"`,
		`placeholder="搜索部门名称"`,
		`orgFilterDepartmentRows`,
		`toggleOrgFilterDepartment(row)`,
		`orgDepartmentCheckState(row)`,
		`orgFilters.managerStatus`,
		`orgFilters.hrbpStatus`,
		`全部上级状态`,
		`已设置上级`,
		`未设置上级`,
		`全部HRBP状态`,
		`已设置HRBP`,
		`未设置HRBP`,
		`matchesOrgFilters(user)`,
		`matchesOrgDepartmentFilter(user)`,
		`hasConfiguredManager(user)`,
		`hasConfiguredHrbp(user)`,
		`resetOrgFilters`,
		`.org-filter-bar`,
		`.org-department-filter`,
		`.org-department-panel`,
		`.org-department-check`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("org view search filters must include %q", snippet)
		}
	}
	if strings.Contains(org, `const departmentTree = computed(() => buildDepartmentTree(ctx.state.users))`) {
		t.Fatalf("org tree should be built from filtered users, not raw visible users")
	}
}

func TestDingTalkH5OrgViewExpandsFilteredTree(t *testing.T) {
	orgSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/OrgView.vue")
	if err != nil {
		t.Fatalf("read org view: %v", err)
	}
	org := string(orgSrc)
	for _, snippet := range []string{
		`const orgHasActiveFilters = computed(() => Boolean(`,
		`normalizeText(orgFilters.employeeKeyword)`,
		`selectedOrgDepartmentLabels.value.length > 0`,
		`orgFilters.managerStatus`,
		`orgFilters.hrbpStatus`,
		`if (orgHasActiveFilters.value) {`,
		`for (const key of availableKeys) next.add(key)`,
	} {
		if !strings.Contains(org, snippet) {
			t.Fatalf("org filtered tree should auto-expand visible departments with %q", snippet)
		}
	}
}

func TestDingTalkH5OrgConfigDoesNotExposeResponsibleDepartments(t *testing.T) {
	orgSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/OrgView.vue")
	if err != nil {
		t.Fatalf("read org view: %v", err)
	}
	pageSrc, err := os.ReadFile("../../../../../dingtalk-h5/pages/index/index.vue")
	if err != nil {
		t.Fatalf("read h5 page: %v", err)
	}
	usersSrc, err := os.ReadFile("users.go")
	if err != nil {
		t.Fatalf("read users.go: %v", err)
	}
	for _, forbidden := range []string{
		`负责部门`,
		`configForm.responsibleDepartments`,
		`responsibleDepartments: ''`,
		`responsibleValue(user)`,
		`function responsibleValue(user)`,
	} {
		if strings.Contains(string(orgSrc), forbidden) {
			t.Fatalf("org config should not expose responsible departments snippet %q", forbidden)
		}
	}
	if strings.Contains(string(pageSrc), `responsibleDepartments:`) {
		t.Fatalf("h5 save user payload should not submit hidden responsible departments")
	}
	for _, snippet := range []string{
		`responsibleDepartments := existingResponsibleDepartments(payload.ResponsibleDepartments, existing)`,
		`ResponsibleDepartments: responsibleDepartments`,
		`func existingResponsibleDepartments(value interface{}, existing *model.DingTalkH5PerfUser) string`,
	} {
		if !strings.Contains(string(usersSrc), snippet) {
			t.Fatalf("backend should preserve omitted responsible departments with %q", snippet)
		}
	}
}

func TestDingTalkH5ProcessModalUsesCompactCurrentStatusCard(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	combined := string(workbenchSrc) + string(styleSrc)
	processModal := string(workbenchSrc)
	if start := strings.Index(processModal, `const ProcessModal = {`); start >= 0 {
		processModal = processModal[start:]
		if end := strings.Index(processModal, `const CurrentObjectiveSection = {`); end >= 0 {
			processModal = processModal[:end]
		}
	}
	for _, snippet := range []string{
		`class: 'process-current-card compact'`,
		`class: 'process-current-index'`,
		`currentProgress.indexText`,
		`class: 'process-current-main'`,
		`class: 'process-current-title-line'`,
		`class: 'process-current-title'`,
		`currentProgress.title`,
		`class: 'process-current-meta'`,
		`class: 'process-current-progress'`,
		`class: 'process-current-handler'`,
		`.process-current-card.compact`,
		`.process-current-index`,
		`.process-current-title-line`,
		`.process-current-meta`,
		`.process-current-handler`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("process modal current status should use compact summary snippet %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		`h('text', { class: 'process-kicker' }, '当前流程状态')`,
		`h('view', { class: 'process-status-line' }, [`,
		`h('text', { class: 'process-handler' }, currentProgress.progressText)`,
	} {
		if strings.Contains(processModal, legacySnippet) {
			t.Fatalf("process modal should not keep loose legacy current status snippet %q", legacySnippet)
		}
	}
}

func TestDingTalkH5HistoryDetailModalScrollsFullReviewForm(t *testing.T) {
	workbenchSrc, err := os.ReadFile("../../../../../dingtalk-h5/components/performance/WorkbenchView.js")
	if err != nil {
		t.Fatalf("read workbench view: %v", err)
	}
	styleSrc, err := os.ReadFile("../../../../../dingtalk-h5/styles/performance.css")
	if err != nil {
		t.Fatalf("read performance css: %v", err)
	}
	workbench := string(workbenchSrc)
	style := string(styleSrc)
	combined := workbench + style
	for _, snippet := range []string{
		`class: 'history-detail-modal-body'`,
		`h(ReviewForm, { review: props.review, readonly: true })`,
		`.history-detail-modal-body`,
		`flex-direction: column;`,
		`overflow: hidden;`,
		`overflow-y: auto;`,
		`min-height: 0;`,
		`padding-bottom: 24px;`,
		`-webkit-overflow-scrolling: touch;`,
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("history detail modal should keep full review form scrollable with %q", snippet)
		}
	}
	modalRule := cssRule(style, ".history-detail-modal")
	if !strings.Contains(modalRule, `display: flex;`) || !strings.Contains(modalRule, `overflow: hidden;`) {
		t.Fatalf("history detail modal must be a fixed-height flex container with hidden outer overflow, got %q", modalRule)
	}
	bodyRule := cssRule(style, ".history-detail-modal-body")
	if !strings.Contains(bodyRule, `flex: 1 1 auto;`) || !strings.Contains(bodyRule, `overflow-y: auto;`) {
		t.Fatalf("history detail modal body must own vertical scrolling, got %q", bodyRule)
	}
}
