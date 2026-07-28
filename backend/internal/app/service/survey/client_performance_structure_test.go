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
