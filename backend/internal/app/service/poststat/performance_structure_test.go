package poststat

import (
	"os"
	"strings"
	"testing"
)

func TestPostStatReadsOnlyNeededResponseData(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"Find(&allResp)",
		"len(allResp)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("post stat should not always load full response rows: %q", snippet)
		}
	}
	required := []string{
		"postStatNeedsAggregateResponses(rules)",
		"Select(\"survey_resp_answers\")",
		"Count(&totalCount)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("post stat should split count-only and answer aggregation paths with %q", snippet)
		}
	}
}
