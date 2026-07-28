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
