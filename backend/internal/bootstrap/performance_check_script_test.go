package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPerformanceCheckScriptIsDocumentedAndRunnable(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	scriptPath := filepath.Join(root, "scripts", "check-performance.mjs")
	scriptSrc, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("performance check script must exist at %s: %v", scriptPath, err)
	}
	script := string(scriptSrc)
	for _, snippet := range []string{
		"WECHECKIN_PERF_BASE_URL",
		"WECHECKIN_ADMIN_TOKEN",
		"WECHECKIN_USER_TOKEN",
		"WECHECKIN_DINGTALK_TOKEN",
		"WECHECKIN_PERF_STRICT",
		"validateBusinessResponse",
		"business code",
		"normalizeAuthToken",
		".replace(/^Bearer\\s+/i, '')",
		"p95",
		"/api/v2/admin/users?page=1&pageSize=20",
		"/api/v2/dingtalk/h5/bootstrap",
		"/api/v2/exams?page=1&pageSize=20",
	} {
		if !strings.Contains(script, snippet) {
			t.Fatalf("performance script should contain %q", snippet)
		}
	}

	packagePath := filepath.Join(root, "package.json")
	packageSrc, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("root package.json must exist at %s: %v", packagePath, err)
	}
	if !strings.Contains(string(packageSrc), `"check:performance"`) {
		t.Fatalf("root package.json should expose check:performance")
	}

	readmeSrc, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeSrc)
	for _, snippet := range []string{
		"npm run check:performance",
		"WECHECKIN_PERF_STRICT=1",
	} {
		if !strings.Contains(readme, snippet) {
			t.Fatalf("README should document performance check snippet %q", snippet)
		}
	}
}
