package main

import (
	"os"
	"strings"
	"testing"
)

func TestTaskdRegistersInAppNotificationJob(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		"inappnotificationinfra.NewGormStore(db)",
		"inappnotificationapp.NewService(notificationStore)",
		"scheduledtaskinfra.NewInAppNotificationJob(notificationService)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("taskd main missing %q", snippet)
		}
	}
}
