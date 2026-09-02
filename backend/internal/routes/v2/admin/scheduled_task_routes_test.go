package admin

import (
	"os"
	"strings"
	"testing"
)

func TestScheduledTaskSpecialRoutesAreRegisteredBeforeIDRoutes(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	preview := strings.Index(text, `admin.POST("/scheduled-tasks/cron-preview"`)
	handlers := strings.Index(text, `admin.GET("/scheduled-task-handlers"`)
	workers := strings.Index(text, `admin.GET("/scheduled-task-workers"`)
	detail := strings.Index(text, `admin.GET("/scheduled-tasks/:id"`)
	if preview < 0 || handlers < 0 || workers < 0 || detail < 0 || preview > detail || handlers > detail || workers > detail {
		t.Fatalf("scheduled task special routes must be registered before ID routes")
	}
}
