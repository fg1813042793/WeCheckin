package httpadmin

import (
	"os"
	"strings"
	"testing"
)

func TestScheduledTaskAdminHandlerExposesManagementEndpoints(t *testing.T) {
	source, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, want := range []string{
		"func (handler *Handler) ListTasks(",
		"func (handler *Handler) CreateTask(",
		"func (handler *Handler) UpdateTask(",
		"func (handler *Handler) DeleteTask(",
		"func (handler *Handler) SetTaskStatus(",
		"func (handler *Handler) RunTask(",
		"func (handler *Handler) PreviewCron(",
		"func (handler *Handler) ListHandlers(",
		"func (handler *Handler) ListRuns(",
		"func (handler *Handler) GetRun(",
		"func (handler *Handler) RetryRun(",
		"func (handler *Handler) CancelRun(",
		"func (handler *Handler) ListWorkers(",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("scheduled task admin handler missing %q", want)
		}
	}
}
