package documentation_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCurrentDocumentationDoesNotReferenceRemovedPaths(t *testing.T) {
	repoRoot := repositoryRoot(t)
	files := []string{
		"docs/PERMISSION_CODE_FRONTEND_SYNC.md",
		"docs/API_V2.md",
		"docs/project-maintenance.md",
		"docs/SCHEDULED_TASKS.md",
		"backend/docs/development-guidelines.md",
	}
	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(repoRoot, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(body)
		for _, removed := range []string{"backend/internal/app/support/", "dingtalk-h5/"} {
			if strings.Contains(text, removed) {
				t.Errorf("%s still references removed path %q", name, removed)
			}
		}
	}
}

func TestPermissionGuideNamesCurrentBackendAndDingTalkSources(t *testing.T) {
	repoRoot := repositoryRoot(t)
	body, err := os.ReadFile(filepath.Join(repoRoot, "docs/PERMISSION_CODE_FRONTEND_SYNC.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, current := range []string{"backend/internal/support/", "h5app/"} {
		if !strings.Contains(text, current) {
			t.Errorf("permission guide must reference current source path %q", current)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../.."))
}
