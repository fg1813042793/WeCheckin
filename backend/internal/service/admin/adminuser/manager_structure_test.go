package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestUserDirectManagerIsExposedAndSaved(t *testing.T) {
	modelSource := readStructureFile(t, "../../../model/account/user.go")
	serviceSource := readStructureFile(t, "service.go")
	handlerSource := readStructureFile(t, "../../../handler/admin/user/handler.go")

	for _, want := range []string{
		"UserReportingRelation",
		"user_reporting_relations",
		"ReportingRelationTypeDirect",
		"EmployeeUserID",
		"ManagerUserID",
	} {
		if !strings.Contains(modelSource, want) {
			t.Fatalf("user reporting relation model must include snippet %q", want)
		}
	}
	if strings.Contains(modelSource, "column:manager_user_id;comment:直属上级用户ID") {
		t.Fatal("users model must not persist the direct manager on the users table")
	}
	for _, want := range []string{
		"ManagerUserID",
		"ManagerUserName",
		"saveUserDirectManagerRelationTx",
		"loadDirectManagerRelationsContext",
		"用户不能选择自己作为直属上级",
	} {
		if !strings.Contains(serviceSource, want) {
			t.Fatalf("user service must expose and persist direct manager snippet %q", want)
		}
	}
	if strings.Contains(serviceSource, `updates["manager_user_id"]`) {
		t.Fatal("user service must persist direct manager through user_reporting_relations")
	}
	if !strings.Contains(handlerSource, "managerUserId") {
		t.Fatalf("user handler must accept managerUserId")
	}
}

func readStructureFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
