package service

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestServicePackageHasNavigationDocs(t *testing.T) {
	for _, file := range []string{"README.md", "doc.go"} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("service package must keep %s for navigation and package ownership: %v", file, err)
		}
	}
}

func TestServiceRootFilesStayFocused(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file == "doc.go" {
			continue
		}
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		t.Fatalf("service root should only keep doc.go and tests; move %s into a domain subpackage or support package", file)
	}
}

func TestServiceDomainFilesStaySplitByResponsibility(t *testing.T) {
	groups := map[string][]string{}
	for name, files := range groups {
		for _, file := range files {
			if _, err := os.Stat(file); err != nil {
				t.Fatalf("%s service code must keep %s split by responsibility: %v", name, file, err)
			}
		}
	}
}

func TestExamServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("client", "exam", "service.go")); err != nil {
		t.Fatalf("ExamService implementation must live in service/client/exam subpackage: %v", err)
	}
	assertNoRootShim(t, "exam_service.go")
}

func TestPassportServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("client", "passport", "service.go")); err != nil {
		t.Fatalf("passport implementation must live in service/client/passport subpackage: %v", err)
	}
	for _, file := range []string{"login.go", "profile.go", "register.go", "phone.go"} {
		if _, err := os.Stat(filepath.Join("client", "passport", file)); err != nil {
			t.Fatalf("passport implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	assertNoRootShim(t, "passport.go")
}

func TestSetupServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "setup", "service.go")); err != nil {
		t.Fatalf("setup implementation must live in service/admin/setup subpackage: %v", err)
	}
	assertNoRootShim(t, "setup_service.go")
}

func TestNewsServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("client", "news", "service.go")); err != nil {
		t.Fatalf("news implementation must live in service/client/news subpackage: %v", err)
	}
	assertNoRootShim(t, "news.go")
}

func TestHomeServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"list.go", "enroll.go"} {
		if _, err := os.Stat(filepath.Join("client", "home", file)); err != nil {
			t.Fatalf("home implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{"home_list.go", "home_enroll.go", "home_media.go"} {
		assertNoRootShim(t, file)
	}
}

func TestEnrollServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"client.go", "detail.go", "fields.go", "records.go", "submission.go", "user_records.go"} {
		if _, err := os.Stat(filepath.Join("client", "enroll", file)); err != nil {
			t.Fatalf("enroll implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{
		"enroll_client.go",
		"enroll_detail.go",
		"enroll_fields.go",
		"enroll_records.go",
		"enroll_submission.go",
		"enroll_user_records.go",
	} {
		assertNoRootShim(t, file)
	}
}

func TestEventServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"admin.go", "client.go", "detail.go", "dynamic.go", "helpers.go", "my.go", "participation.go", "score.go"} {
		if _, err := os.Stat(filepath.Join("client", "event", file)); err != nil {
			t.Fatalf("event implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{
		"event_admin.go",
		"event_client.go",
		"event_detail.go",
		"event_dynamic.go",
		"event_helpers.go",
		"event_my.go",
		"event_participation.go",
		"event_score.go",
	} {
		assertNoRootShim(t, file)
	}
}

func TestFavoriteServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("client", "favorite", "service.go")); err != nil {
		t.Fatalf("favorite implementation must live in service/client/favorite subpackage: %v", err)
	}
	assertNoRootShim(t, "fav.go")
}

func TestDictServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "dict", "service.go")); err != nil {
		t.Fatalf("dict implementation must live in service/admin/dict subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_dict_service.go")
}

func TestMenuServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "menu", "service.go")); err != nil {
		t.Fatalf("menu implementation must live in service/admin/menu subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_menu_service.go")
}

func TestOnlineServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"sessions.go", "user.go", "admin.go"} {
		if _, err := os.Stat(filepath.Join("admin", "online", file)); err != nil {
			t.Fatalf("online implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{"admin_online_sessions.go", "admin_online_user.go", "admin_online_admin.go"} {
		assertNoRootShim(t, file)
	}
}

func TestSurveyPostStatServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"service.go", "submitter.go", "rules.go", "result.go", "webhook.go", "notify.go"} {
		if _, err := os.Stat(filepath.Join("client", "poststat", file)); err != nil {
			t.Fatalf("survey poststat implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{
		"survey_poststat.go",
		"survey_poststat_submitter.go",
		"survey_poststat_rules.go",
		"survey_poststat_result.go",
		"survey_poststat_webhook.go",
		"survey_poststat_notify.go",
	} {
		assertNoRootShim(t, file)
	}
}

func TestSurveyServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{
		"service.go",
		"response.go",
		"response_query.go",
		"response_submit.go",
		"response_draft.go",
		"response_device.go",
		"response_validation.go",
	} {
		if _, err := os.Stat(filepath.Join("client", "survey", file)); err != nil {
			t.Fatalf("survey implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{
		"survey_service.go",
		"survey_response.go",
		"survey_response_query.go",
		"survey_response_submit.go",
		"survey_response_draft.go",
		"survey_response_device.go",
		"survey_response_validation.go",
	} {
		assertNoRootShim(t, file)
	}
}

func TestDingTalkH5ServiceUsesBusinessSubpackages(t *testing.T) {
	for _, file := range []string{
		filepath.Join("dingtalkh5", "performance", "reviews.go"),
		filepath.Join("dingtalkh5", "performance", "users.go"),
		filepath.Join("dingtalkh5", "performance", "defaults.go"),
		filepath.Join("dingtalkh5", "performance", "auth.go"),
		filepath.Join("dingtalkh5", "performance", "corp_config.go"),
	} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("dingtalk h5 service implementation must live under business subpackages: %v", err)
		}
	}

	rootFiles, err := filepath.Glob(filepath.Join("dingtalkh5", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range rootFiles {
		t.Fatalf("dingtalk h5 service root should not keep implementation file %s; move it into a business subpackage", file)
	}
}

func TestRoleServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "role", "service.go")); err != nil {
		t.Fatalf("role implementation must live in service/role subpackage: %v", err)
	}
	for _, file := range []string{"admin_role_service.go", "admin_mgr_dept.go"} {
		assertNoRootShim(t, file)
	}
}

func TestDepartmentServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "department", "service.go")); err != nil {
		t.Fatalf("department implementation must live in service/department subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_department_service.go")
}

func TestAdminAuthServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "adminauth", "service.go")); err != nil {
		t.Fatalf("admin auth implementation must live in service/adminauth subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_auth.go")
}

func TestAdminManagerServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"service.go", "password.go"} {
		if _, err := os.Stat(filepath.Join("admin", "adminmgr", file)); err != nil {
			t.Fatalf("admin manager implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{"admin_mgr_service.go", "admin_mgr_password.go"} {
		assertNoRootShim(t, file)
	}
}

func TestHighFrequencyAdminServicesUseTypedResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "role", "service.go"),
		filepath.Join("admin", "adminmgr", "service.go"),
		filepath.Join("client", "news", "service.go"),
		filepath.Join("admin", "adminauth", "service.go"),
		filepath.Join("client", "passport", "login.go"),
		filepath.Join("client", "passport", "profile.go"),
		filepath.Join("client", "passport", "register.go"),
		filepath.Join("admin", "dict", "service.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			") (map[string]interface{}, error)",
			") ([]map[string]interface{}, error)",
			"result := make([]map[string]interface{}",
			"return map[string]interface{}{",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s high-frequency service responses must use typed DTOs instead of %q", file, snippet)
			}
		}
	}
}

func TestHighFrequencyAdminServicesUseRequestContext(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "role", "service.go"),
		filepath.Join("admin", "adminmgr", "service.go"),
		filepath.Join("admin", "adminmgr", "password.go"),
		filepath.Join("admin", "adminauth", "service.go"),
		filepath.Join("client", "passport", "login.go"),
		filepath.Join("client", "passport", "profile.go"),
		filepath.Join("client", "passport", "register.go"),
		filepath.Join("client", "news", "service.go"),
		filepath.Join("admin", "dict", "service.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "database.DB.") {
			t.Fatalf("%s must use database.WithContext or transaction handles instead of direct database.DB calls", file)
		}
	}
}

func TestClientEventEnrollServicesUseTypedContextResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "event", "client.go"),
		filepath.Join("client", "enroll", "client.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			") (map[string]interface{}, error)",
			"return map[string]interface{}{",
			"database.DB.",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s client list services must use typed responses and request context instead of %q", file, snippet)
			}
		}
	}
}

func TestEventSecondaryServicesUseContextAndTypedResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "event", "my.go"),
		filepath.Join("client", "event", "dynamic.go"),
		filepath.Join("client", "event", "score.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			") (map[string]interface{}, error)",
			"return map[string]interface{}{",
			"database.DB.",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must use context-aware queries and typed responses instead of %q", file, snippet)
			}
		}
	}
}

func TestEventAdminDetailServicesUseRequestContext(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "event", "admin.go"),
		filepath.Join("client", "event", "detail.go"),
		filepath.Join("client", "event", "helpers.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "database.DB.") {
			t.Fatalf("%s must use database.WithContext or transaction handles instead of direct database.DB calls", file)
		}
	}
}

func TestEventHelpersUseTypedDeptUsers(t *testing.T) {
	src := readServiceSource(t, filepath.Join("client", "event", "helpers.go"))
	for _, snippet := range []string{
		"[]map[string]interface{}",
		"map[string]interface{}{",
		".Find(&users)\n",
		"First(&user)",
	} {
		if strings.Contains(src, snippet) {
			t.Fatalf("event/helpers.go must type dept-user responses and check query errors instead of %q", snippet)
		}
	}
}

func TestExamAdminServiceUsesContextAndTypedResponses(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("client", "exam", "service.go"))
	if err != nil {
		t.Fatalf("read exam/service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func (s *Service) Statistics(examID int) map[string]interface{}",
		"return map[string]interface{}{",
		"database.DB.",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("exam/service.go must use context-aware queries and typed responses instead of %q", snippet)
		}
	}

	for _, fn := range []string{"RecordDelete", "RecordBatchDelete"} {
		body := functionBody(t, text, fn)
		if !strings.Contains(body, fn+"Context(") {
			t.Fatalf("%s must delegate to a context-aware implementation", fn)
		}
	}
}

func TestExamClientServiceUsesTypedQuestionResponses(t *testing.T) {
	src := readServiceSource(t, filepath.Join("client", "exam", "client.go"))
	for _, snippet := range []string{
		"Questions []map[string]interface{}",
		"func safeExamQuestions(questions []model.ExamQuestion, options PaperQuestionOptions) []map[string]interface{}",
		"safe := make([]map[string]interface{}",
		"item := map[string]interface{}",
	} {
		if strings.Contains(src, snippet) {
			t.Fatalf("exam/client.go must type client question responses instead of %q", snippet)
		}
	}
}

func TestUserFavoriteDashboardServicesUseContextAndTypedResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "favorite", "service.go"),
		filepath.Join("admin", "dashboard", "service.go"),
		filepath.Join("admin", "adminuser", "service.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			") (map[string]interface{}, error)",
			") ([]map[string]interface{}, error)",
			"return map[string]interface{}{",
			"database.DB.",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must use context-aware queries and typed responses instead of %q", file, snippet)
			}
		}
	}
}

func TestHomeListServiceUsesContextAndTypedResponses(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("client", "home", "list.go"))
	if err != nil {
		t.Fatalf("read home/list.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		") (map[string]interface{}, error)",
		"[]map[string]interface{}",
		"map[string]interface{}{",
		"database.DB.",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("home/list.go must use context-aware queries and typed responses instead of %q", snippet)
		}
	}
}

func TestSurveySubmitPostStatUseRequestContext(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "survey", "response_submit.go"),
		filepath.Join("client", "survey", "response_draft.go"),
		filepath.Join("client", "poststat", "service.go"),
		filepath.Join("client", "poststat", "submitter.go"),
		filepath.Join("client", "poststat", "notify.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "database.DB.") {
			t.Fatalf("%s must use database.WithContext instead of direct database.DB calls", file)
		}
	}
}

func TestSurveyAnswerServicesUseTypedResults(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "survey", "service.go"),
		filepath.Join("client", "survey", "response_query.go"),
	} {
		src := readServiceSource(t, file)
		for _, snippet := range []string{
			") map[string]interface{}",
			") (map[string]interface{}, error)",
		} {
			if strings.Contains(src, snippet) {
				t.Fatalf("%s must wrap answer maps in typed result DTOs instead of %q", file, snippet)
			}
		}
	}
}

func TestOnlineServicesUseRequestContext(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "online", "admin.go"),
		filepath.Join("admin", "online", "sessions.go"),
		filepath.Join("admin", "online", "user.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		if strings.Contains(text, "database.DB.") {
			t.Fatalf("%s must use database.WithContext instead of direct database.DB calls", file)
		}
		if strings.Contains(text, "OperationContext(context.Background())") {
			t.Fatalf("%s must derive redis operation context from caller context", file)
		}
		for _, snippet := range []string{
			"[]map[string]interface{}",
			"map[string]interface{}{",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must use typed online response DTOs instead of %q", file, snippet)
			}
		}
	}
}

func TestEnrollRecordsUseRequestContext(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("client", "enroll", "records.go"))
	if err != nil {
		t.Fatalf("read enroll/records.go: %v", err)
	}
	if strings.Contains(string(src), "database.DB.") {
		t.Fatalf("enroll/records.go must use database.WithContext instead of direct database.DB calls")
	}
	text := string(src)
	for _, snippet := range []string{
		"func GetEnrollJoinByDay(enrollID, day string) ([]map[string]interface{}, error)",
		"func GetEnrollJoinByDayContext(ctx context.Context, enrollID, day string) ([]map[string]interface{}, error)",
		"func GetMyDayRecords(userID, day string) ([]map[string]interface{}, error)",
		"func GetMyDayRecordsContext(ctx context.Context, userID, day string) ([]map[string]interface{}, error)",
		"func GetMyEnrollJoinList(userID, enrollID string, page, pageSize int) (interface{}, int64, error)",
		"func GetMyEnrollJoinListContext(ctx context.Context, userID, enrollID string, page, pageSize int) (interface{}, int64, error)",
		"var allUsers []model.User",
		"db.Find(&allUsers)",
		"First(&e)",
		"First(&enroll)",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("enroll/records.go must type common record responses instead of %q", snippet)
		}
	}
}

func TestAdminConfigAndEnrollDetailServicesUseRequestContext(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "department", "service.go"),
		filepath.Join("admin", "menu", "service.go"),
		filepath.Join("admin", "adminlog", "service.go"),
		filepath.Join("client", "enroll", "detail.go"),
		filepath.Join("client", "enroll", "user_records.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "database.DB.") {
			t.Fatalf("%s must use database.WithContext or transaction handles instead of direct database.DB calls", file)
		}
	}
}

func TestAdminContentServicesUseRequestContext(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "admincontent", "enroll.go"),
		filepath.Join("admin", "admincontent", "enroll_records.go"),
		filepath.Join("admin", "admincontent", "export.go"),
		filepath.Join("admin", "admincontent", "news.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "database.DB.") {
			t.Fatalf("%s must use database.WithContext or transaction handles instead of direct database.DB calls", file)
		}
	}
}

func TestCriticalDeleteCountersUseTransactions(t *testing.T) {
	adminEnrollRecords := readServiceSource(t, filepath.Join("admin", "admincontent", "enroll_records.go"))
	for _, fn := range []string{"DelEnrollJoinContext", "RemoveEnrollUserContext"} {
		body := functionBody(t, adminEnrollRecords, fn)
		if !strings.Contains(body, ".Transaction(") {
			t.Fatalf("%s must wrap delete and counter updates in one transaction", fn)
		}
	}

	eventAdmin := readServiceSource(t, filepath.Join("client", "event", "admin.go"))
	delEvent := functionBody(t, eventAdmin, "DelEventContext")
	if !strings.Contains(delEvent, ".Transaction(") || strings.Contains(delEvent, ".Begin()") {
		t.Fatalf("DelEventContext must use database transaction helper with checked errors")
	}
	delParticipant := functionBody(t, eventAdmin, "DelEventParticipantContext")
	for _, snippet := range []string{".Transaction(", "event_user_cnt"} {
		if !strings.Contains(delParticipant, snippet) {
			t.Fatalf("DelEventParticipantContext must keep participant delete and event_user_cnt update in one transaction")
		}
	}
}

func TestAdminUserServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "adminuser", "service.go")); err != nil {
		t.Fatalf("admin user implementation must live in service/adminuser subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_user_service.go")
}

func TestAdminContentServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"enroll.go", "enroll_records.go", "export.go", "news.go"} {
		if _, err := os.Stat(filepath.Join("admin", "admincontent", file)); err != nil {
			t.Fatalf("admin content implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{
		"admin_enroll_service.go",
		"admin_enroll_records.go",
		"admin_enroll_export.go",
		"admin_news_service.go",
	} {
		assertNoRootShim(t, file)
	}
}

func TestAdminLogServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "adminlog", "service.go")); err != nil {
		t.Fatalf("admin log implementation must live in service/adminlog subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_log_service.go")
}

func TestDashboardServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("admin", "dashboard", "service.go")); err != nil {
		t.Fatalf("dashboard implementation must live in service/dashboard subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_dashboard.go")
}

func countLines(t *testing.T, file string) int {
	t.Helper()
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(content), "\n") + 1
}

func readServiceSource(t *testing.T, file string) string {
	t.Helper()
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return string(src)
}

func functionBody(t *testing.T, src, name string) string {
	t.Helper()
	start := strings.Index(src, "func "+name+"(")
	if start < 0 {
		methodPattern := regexp.MustCompile(`func\s+\([^)]*\)\s+` + regexp.QuoteMeta(name) + `\s*\(`)
		if loc := methodPattern.FindStringIndex(src); loc != nil {
			start = loc[0]
		}
	}
	if start < 0 {
		t.Fatalf("function %s not found", name)
	}
	next := strings.Index(src[start+len("func "):], "\nfunc ")
	if next < 0 {
		return src[start:]
	}
	return src[start : start+len("func ")+next]
}

func assertNoRootShim(t *testing.T, file string) {
	t.Helper()
	if _, err := os.Stat(file); err == nil {
		t.Fatalf("%s should not remain as a root compatibility shim; import the domain subpackage directly", file)
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}
