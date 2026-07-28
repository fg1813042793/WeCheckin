package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartScriptDoesNotRunMaintenanceTasks(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "start.sh"))
	if err != nil {
		t.Fatalf("read start.sh: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"WECHECKIN_AUTO_MIGRATE",
		"自动迁移",
		"InitBusiness",
		"init.sh",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("start.sh must not run or advertise maintenance tasks with %s", snippet)
		}
	}
}

func TestInitScriptRunsStandaloneMaintenanceCommand(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "init.sh"))
	if err != nil {
		t.Fatalf("read init.sh: %v", err)
	}
	text := string(src)
	required := []string{
		"go run ./cmd/maintenance",
		"-migrations",
		"schema_migrations",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("init.sh must include %s", snippet)
		}
	}
}
