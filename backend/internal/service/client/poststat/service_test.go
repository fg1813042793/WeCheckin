package poststat

import (
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/formkit/report"
)

func TestRenderNotificationMessageReplacesAvailablePlaceholders(t *testing.T) {
	now := time.Date(2026, 9, 4, 11, 22, 33, 0, time.Local)
	message := renderNotificationMessage(
		"{title}|{submitter}|{date}|{total}|{questionCount}|{result}",
		"季度问卷", "张三", now, 4,
		[]report.FieldStat{{Type: "radio", Title: "意见", Dist: map[string]int{"同意": 1}}}, "value",
	)
	for _, expected := range []string{"季度问卷", "张三", "2026-09-04 11:22:33", "4", "1", "同意"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("message %q does not contain %q", message, expected)
		}
	}
}
