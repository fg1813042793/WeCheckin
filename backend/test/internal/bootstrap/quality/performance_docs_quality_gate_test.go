package quality_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPerformanceDocsAndQualityGateAreWired(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	checkScript, err := os.ReadFile(filepath.Join(root, "scripts", "check.sh"))
	if err != nil {
		t.Fatalf("read scripts/check.sh: %v", err)
	}
	for _, snippet := range []string{
		"CHECK_PERFORMANCE",
		"npm run check:performance",
	} {
		if !strings.Contains(string(checkScript), snippet) {
			t.Fatalf("scripts/check.sh should wire performance check snippet %q", snippet)
		}
	}

	qualityGate, err := os.ReadFile(filepath.Join(root, "scripts", "check-quality-gates.mjs"))
	if err != nil {
		t.Fatalf("read scripts/check-quality-gates.mjs: %v", err)
	}
	for _, snippet := range []string{
		"check:performance",
		"docs/performance/README.md",
		"docs/performance/mysql-indexes.md",
		"docs/performance/api-baseline.md",
	} {
		if !strings.Contains(string(qualityGate), snippet) {
			t.Fatalf("scripts/check-quality-gates.mjs should guard performance artifact %q", snippet)
		}
	}

	for _, doc := range []string{
		"docs/performance/README.md",
		"docs/performance/mysql-indexes.md",
		"docs/performance/api-baseline.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, doc))
		if err != nil {
			t.Fatalf("performance document %s must exist: %v", doc, err)
		}
		for _, snippet := range []string{"EXPLAIN", "npm run check:performance", "slow query"} {
			if !strings.Contains(string(content), snippet) {
				t.Fatalf("performance document %s should mention %q", doc, snippet)
			}
		}
	}
}
