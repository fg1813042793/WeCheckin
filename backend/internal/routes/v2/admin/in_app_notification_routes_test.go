package admin

import (
	"os"
	"strings"
	"testing"
)

func TestAdminRoutesRegisterInAppNotificationEndpoints(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		"notificationService := registerNotificationRoutes(admin)",
		`admin.GET("/in-app-notifications", handler.List)`,
		`admin.GET("/in-app-notifications/unread-count", handler.UnreadCount)`,
		`admin.GET("/in-app-notifications/recipient-options", handler.RecipientOptions)`,
		`admin.POST("/in-app-notifications", handler.Send)`,
		`admin.DELETE("/in-app-notifications/:id", handler.DeleteRecord)`,
		`admin.GET("/dingtalk-notifications/recipient-options", handler.RecipientOptions)`,
		`admin.POST("/dingtalk-notifications", handler.SendDingTalk)`,
		`admin.GET("/notification-styles", handler.NotificationStyles)`,
		`admin.PUT("/notification-styles", handler.SaveNotificationStyles)`,
		`admin.POST("/notification-styles/test/in-app", handler.SendInAppStyleTest)`,
		`admin.POST("/notification-styles/test/dingtalk", handler.SendDingTalkStyleTest)`,
		`admin.PATCH("/in-app-notifications/read-all", handler.MarkAllRead)`,
		`admin.PATCH("/in-app-notifications/:id/read", handler.MarkRead)`,
		"scheduledtaskinfra.NewInAppNotificationJob(notificationService)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("routes.go missing %q", snippet)
		}
	}
}
