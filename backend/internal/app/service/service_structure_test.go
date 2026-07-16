package service

import (
	"errors"
	"os"
	"path/filepath"
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
	if _, err := os.Stat(filepath.Join("exam", "service.go")); err != nil {
		t.Fatalf("ExamService implementation must live in service/exam subpackage: %v", err)
	}
	assertNoRootShim(t, "exam_service.go")
}

func TestPassportServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("passport", "service.go")); err != nil {
		t.Fatalf("passport implementation must live in service/passport subpackage: %v", err)
	}
	for _, file := range []string{"login.go", "profile.go", "register.go", "phone.go"} {
		if _, err := os.Stat(filepath.Join("passport", file)); err != nil {
			t.Fatalf("passport implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	assertNoRootShim(t, "passport.go")
}

func TestSetupServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("setup", "service.go")); err != nil {
		t.Fatalf("setup implementation must live in service/setup subpackage: %v", err)
	}
	assertNoRootShim(t, "setup_service.go")
}

func TestNewsServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("news", "service.go")); err != nil {
		t.Fatalf("news implementation must live in service/news subpackage: %v", err)
	}
	assertNoRootShim(t, "news.go")
}

func TestHomeServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"list.go", "enroll.go"} {
		if _, err := os.Stat(filepath.Join("home", file)); err != nil {
			t.Fatalf("home implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{"home_list.go", "home_enroll.go", "home_media.go"} {
		assertNoRootShim(t, file)
	}
}

func TestEnrollServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"client.go", "detail.go", "fields.go", "records.go", "submission.go", "user_records.go"} {
		if _, err := os.Stat(filepath.Join("enroll", file)); err != nil {
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
		if _, err := os.Stat(filepath.Join("event", file)); err != nil {
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
	if _, err := os.Stat(filepath.Join("favorite", "service.go")); err != nil {
		t.Fatalf("favorite implementation must live in service/favorite subpackage: %v", err)
	}
	assertNoRootShim(t, "fav.go")
}

func TestDictServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("dict", "service.go")); err != nil {
		t.Fatalf("dict implementation must live in service/dict subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_dict_service.go")
}

func TestMenuServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("menu", "service.go")); err != nil {
		t.Fatalf("menu implementation must live in service/menu subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_menu_service.go")
}

func TestOnlineServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"sessions.go", "user.go", "admin.go"} {
		if _, err := os.Stat(filepath.Join("online", file)); err != nil {
			t.Fatalf("online implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{"admin_online_sessions.go", "admin_online_user.go", "admin_online_admin.go"} {
		assertNoRootShim(t, file)
	}
}

func TestSurveyPostStatServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"service.go", "submitter.go", "rules.go", "result.go", "webhook.go", "notify.go"} {
		if _, err := os.Stat(filepath.Join("poststat", file)); err != nil {
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
		if _, err := os.Stat(filepath.Join("survey", file)); err != nil {
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

func TestRoleServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("role", "service.go")); err != nil {
		t.Fatalf("role implementation must live in service/role subpackage: %v", err)
	}
	for _, file := range []string{"admin_role_service.go", "admin_mgr_dept.go"} {
		assertNoRootShim(t, file)
	}
}

func TestDepartmentServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("department", "service.go")); err != nil {
		t.Fatalf("department implementation must live in service/department subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_department_service.go")
}

func TestAdminAuthServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("adminauth", "service.go")); err != nil {
		t.Fatalf("admin auth implementation must live in service/adminauth subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_auth.go")
}

func TestAdminManagerServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"service.go", "password.go"} {
		if _, err := os.Stat(filepath.Join("adminmgr", file)); err != nil {
			t.Fatalf("admin manager implementation must keep %s split by responsibility: %v", file, err)
		}
	}
	for _, file := range []string{"admin_mgr_service.go", "admin_mgr_password.go"} {
		assertNoRootShim(t, file)
	}
}

func TestHighFrequencyAdminServicesUseTypedResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("role", "service.go"),
		filepath.Join("adminmgr", "service.go"),
		filepath.Join("news", "service.go"),
		filepath.Join("adminauth", "service.go"),
		filepath.Join("passport", "login.go"),
		filepath.Join("passport", "profile.go"),
		filepath.Join("passport", "register.go"),
		filepath.Join("dict", "service.go"),
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
		filepath.Join("role", "service.go"),
		filepath.Join("adminmgr", "service.go"),
		filepath.Join("adminmgr", "password.go"),
		filepath.Join("adminauth", "service.go"),
		filepath.Join("passport", "login.go"),
		filepath.Join("passport", "profile.go"),
		filepath.Join("passport", "register.go"),
		filepath.Join("news", "service.go"),
		filepath.Join("dict", "service.go"),
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
		filepath.Join("event", "client.go"),
		filepath.Join("enroll", "client.go"),
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

func TestUserFavoriteDashboardServicesUseContextAndTypedResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("favorite", "service.go"),
		filepath.Join("dashboard", "service.go"),
		filepath.Join("adminuser", "service.go"),
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

func TestAdminUserServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("adminuser", "service.go")); err != nil {
		t.Fatalf("admin user implementation must live in service/adminuser subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_user_service.go")
}

func TestAdminContentServiceUsesDedicatedSubpackage(t *testing.T) {
	for _, file := range []string{"enroll.go", "enroll_records.go", "export.go", "news.go"} {
		if _, err := os.Stat(filepath.Join("admincontent", file)); err != nil {
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
	if _, err := os.Stat(filepath.Join("adminlog", "service.go")); err != nil {
		t.Fatalf("admin log implementation must live in service/adminlog subpackage: %v", err)
	}
	assertNoRootShim(t, "admin_log_service.go")
}

func TestDashboardServiceUsesDedicatedSubpackage(t *testing.T) {
	if _, err := os.Stat(filepath.Join("dashboard", "service.go")); err != nil {
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

func assertNoRootShim(t *testing.T, file string) {
	t.Helper()
	if _, err := os.Stat(file); err == nil {
		t.Fatalf("%s should not remain as a root compatibility shim; import the domain subpackage directly", file)
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}
