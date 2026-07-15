package service

import (
	"os"
	"testing"
)

func TestEventServiceCodeStaysSplitByResponsibility(t *testing.T) {
	required := []string{"event_client.go", "event_admin.go", "event_helpers.go"}
	for _, file := range required {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("event service code must keep %s split by responsibility: %v", file, err)
		}
	}
}
