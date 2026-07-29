package event

import (
	"os"
	"strings"
	"testing"
)

func TestEventClientListLimitsParticipationLookupToCurrentPage(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"db.Where(\"`event_part_mini_openid` = ?\", userID).Find(&parts)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("event client list should not load all user participation history: %q", snippet)
		}
	}
	required := []string{
		"eventListIDs(list)",
		"`event_part_event_id` IN ?",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("event client list should limit participation lookup to current page with %q", snippet)
		}
	}
}

func TestEventClientListSelectsLightweightColumns(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"var clientEventListColumns = []string{",
		"Select(clientEventListColumns)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("event client list should use lightweight columns with %q", snippet)
		}
	}
	start := strings.Index(text, "var clientEventListColumns = []string{")
	if start < 0 {
		t.Fatalf("clientEventListColumns declaration missing")
	}
	end := strings.Index(text[start:], "}")
	if end < 0 {
		t.Fatalf("clientEventListColumns declaration incomplete")
	}
	columnsBlock := text[start : start+end]
	for _, forbidden := range []string{"event_forms", "event_score_fields"} {
		if strings.Contains(columnsBlock, forbidden) {
			t.Fatalf("event client list columns should not include heavy field %q", forbidden)
		}
	}
}
