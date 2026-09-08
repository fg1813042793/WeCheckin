package admin

import (
	"os"
	"strings"
	"testing"
)

func TestAdminRoutesRegisterWorkflowNotificationDelete(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	if snippet := `admin.DELETE("/workflow-notifications/:id", runtimeHandler.DeleteNotification)`; !strings.Contains(string(source), snippet) {
		t.Fatalf("workflow notification routes missing %s", snippet)
	}
}
