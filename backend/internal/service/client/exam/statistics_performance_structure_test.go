package exam

import (
	"os"
	"strings"
	"testing"
)

func TestExamStatisticsUsesAggregatedAndInMemoryAnswerStats(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"Count(&cnt)",
		"JSON_EXTRACT(`exam_r_answers`",
		"Count(&submittedTotal)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("exam statistics should avoid repeated SQL snippet %q", snippet)
		}
	}
	required := []string{
		"SUM(CASE WHEN `exam_r_status` >= 1 THEN 1 ELSE 0 END)",
		"DATE_FORMAT(FROM_UNIXTIME(`exam_r_submit_time` / 1000), '%m-%d')",
		"Select(\"exam_r_answers\")",
		"answerValueNonEmpty",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("exam statistics should use aggregated/in-memory stats with %q", snippet)
		}
	}
}
