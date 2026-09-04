package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5NotificationRoutesUseAuthenticatedGroup(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		`authed.GET("/notifications", handler.Notification.List)`,
		`authed.GET("/notifications/unread-count", handler.Notification.UnreadCount)`,
		`authed.PATCH("/notifications/read-all", handler.Notification.MarkAllRead)`,
		`authed.PATCH("/notifications/:id/read", handler.Notification.MarkRead)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk H5 routes missing authenticated notification registration %q", snippet)
		}
	}
	for _, forbidden := range []string{
		`auth.GET("/notifications"`,
		`auth.PATCH("/notifications`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("notification routes must not require the app permission middleware: found %q", forbidden)
		}
	}
}
