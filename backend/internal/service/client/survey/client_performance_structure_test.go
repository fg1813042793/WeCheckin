package survey

import (
	"os"
	"strings"
	"testing"
)

func TestSurveyPublishedListBatchesLimitCounts(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"for _, survey := range list {\n\t\tlimitInfo, hasLimit, err := s.surveyListLimitInfo",
		"surveyListLimitInfo(db, survey",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("survey published list should not query limit counts one survey at a time: %q", snippet)
		}
	}
	required := []string{
		"loadSurveyListLimitInfoContext(ctx, db, list, deviceID, clientIP)",
		"loadSurveyResponseCountsByFilter",
		"Group(\"`survey_resp_survey_id`\")",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("survey published list should batch limit counts with %q", snippet)
		}
	}
}

func TestSurveyPublishedListSelectsLightweightColumns(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"var publishedSurveyListColumns = []string{",
		"Select(publishedSurveyListColumns)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("survey published list should use lightweight columns with %q", snippet)
		}
	}
	start := strings.Index(text, "var publishedSurveyListColumns = []string{")
	if start < 0 {
		t.Fatalf("publishedSurveyListColumns declaration missing")
	}
	end := strings.Index(text[start:], "}")
	if end < 0 {
		t.Fatalf("publishedSurveyListColumns declaration incomplete")
	}
	columnsBlock := text[start : start+end]
	for _, forbidden := range []string{"survey_schema"} {
		if strings.Contains(columnsBlock, forbidden) {
			t.Fatalf("survey published list columns should not include heavy field %q", forbidden)
		}
	}
}
