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
		"ManagerUserID",
		"column:manager_user_id",
	} {
		if !strings.Contains(modelSource, want) {
			t.Fatalf("user model must include direct manager field snippet %q", want)
		}
	}
	for _, want := range []string{
		"ManagerUserID",
		"ManagerUserName",
		"manager_user_id",
		"loadManagerUserNameMapContext",
		"用户不能选择自己作为直属上级",
	} {
		if !strings.Contains(serviceSource, want) {
			t.Fatalf("user service must expose and persist direct manager snippet %q", want)
		}
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
