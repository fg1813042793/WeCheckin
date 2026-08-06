package survey

import (
	"os"
	"strings"
	"testing"
)

func TestSurveyStatisticsUsesAggregatedQueries(t *testing.T) {
	src, err := os.ReadFile("statistics.go")
	if err != nil {
		t.Fatalf("read statistics.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"Count(&count)",
		"Count(&mobileCount)",
		"Count(&pcCount)",
		"Find(&allResponses)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("survey statistics should avoid repeated/full-row query snippet %q", snippet)
		}
	}
	required := []string{
		"DATE_FORMAT(FROM_UNIXTIME(`survey_resp_add_time` / 1000), '%m-%d')",
		"SUM(CASE WHEN `survey_resp_device` LIKE ? THEN 1 ELSE 0 END)",
		"Select(\"survey_resp_answers\")",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("survey statistics should use aggregated/lightweight query with %q", snippet)
		}
	}
}
