package poststat

import (
	"strings"
	"testing"
	"time"
)

func TestRenderExamNotificationMessageControlsTotalScoreOutput(t *testing.T) {
	now := time.Date(2026, 10, 31, 9, 30, 0, 0, time.Local)
	withScore := renderExamNotificationMessage("{title}|{submitter}|{date}|{total}|{result}", "月考", "张三", now, 3, 88, 100, true)
	for _, want := range []string{"月考", "张三", "2026-10-31 09:30:00", "3", "得分 88 / 100"} {
		if !strings.Contains(withScore, want) {
			t.Fatalf("message %q does not contain %q", withScore, want)
		}
	}

	withoutScore := renderExamNotificationMessage("{title}|{result}", "月考", "张三", now, 3, 88, 100, false)
	if strings.Contains(withoutScore, "88") || strings.Contains(withoutScore, "100") {
		t.Fatalf("score must be omitted when calculateTotalScore is disabled: %q", withoutScore)
	}
}

func TestExamNotificationRequestsUseExamSources(t *testing.T) {
	requests := examNotificationRequests(examRuleNotificationInput{
		ExamID: 5, RecordID: 9, RuleIndex: 1, ExamTitle: "月考", Message: "得分 88 / 100",
		Rule: Rule{ID: "r1", NotifyChannel: "internal", NotifyAdmin: true},
	})
	if len(requests) != 1 {
		t.Fatalf("requests = %#v", requests)
	}
	request := requests[0]
	if request.SourceType != "exam_record" || request.SourceID != "9" {
		t.Fatalf("request source = %s/%s", request.SourceType, request.SourceID)
	}
}
