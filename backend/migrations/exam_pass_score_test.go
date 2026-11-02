package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestExamPassScoreMigration(t *testing.T) {
	data, err := os.ReadFile("20261031130000_add_exam_pass_score.sql")
	if err != nil {
		t.Fatalf("read exam pass score migration: %v", err)
	}
	sql := string(data)
	for _, snippet := range []string{"exam_pass_score", "ALTER TABLE `exam`", "DEFAULT 0"} {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("exam pass score migration missing %q", snippet)
		}
	}
}
