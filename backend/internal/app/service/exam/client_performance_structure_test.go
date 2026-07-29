package exam

import (
	"os"
	"strings"
	"testing"
)

func TestPublishedExamListBatchLoadsLimitInfo(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	if strings.Contains(text, "s.examListLimitInfo(db, item, deviceID, clientIP)") {
		t.Fatalf("published exam list should not query limit info per exam")
	}
	required := []string{
		"loadExamListLimitInfoContext(ctx, db, list, deviceID, clientIP)",
		"`exam_r_exam_id` IN ?",
		"Group(\"`exam_r_exam_id`\")",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("published exam list should batch load limit info with %q", snippet)
		}
	}
}

func TestPublishedExamListSelectsLightweightColumns(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"var publishedExamListColumns = []string{",
		"Select(publishedExamListColumns)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("published exam list should use lightweight columns with %q", snippet)
		}
	}
	start := strings.Index(text, "var publishedExamListColumns = []string{")
	if start < 0 {
		t.Fatalf("publishedExamListColumns declaration missing")
	}
	end := strings.Index(text[start:], "}")
	if end < 0 {
		t.Fatalf("publishedExamListColumns declaration incomplete")
	}
	columnsBlock := text[start : start+end]
	for _, forbidden := range []string{"exam_schema"} {
		if strings.Contains(columnsBlock, forbidden) {
			t.Fatalf("published exam list columns should not include heavy field %q", forbidden)
		}
	}
}
