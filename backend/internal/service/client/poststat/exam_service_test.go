package poststat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	exportlinkapp "wecheckin/backend/internal/modules/examrecordexportlink/application"
)

func TestRenderExamNotificationMessageControlsTotalScoreOutput(t *testing.T) {
	now := time.Date(2026, 10, 31, 9, 30, 0, 0, time.Local)
	withScore := renderExamNotificationMessage("{title}|{submitter}|{date}|{total}|{result}", "月考", "张三", now, 3, 88, 100, true, false, "")
	for _, want := range []string{"月考", "张三", "2026-10-31 09:30:00", "3", "得分 88 / 100"} {
		if !strings.Contains(withScore, want) {
			t.Fatalf("message %q does not contain %q", withScore, want)
		}
	}

	withoutScore := renderExamNotificationMessage("{title}|{result}", "月考", "张三", now, 3, 88, 100, false, false, "")
	if strings.Contains(withoutScore, "88") || strings.Contains(withoutScore, "100") {
		t.Fatalf("score must be omitted when calculateTotalScore is disabled: %q", withoutScore)
	}
}

func TestRenderExamNotificationMessageUsesSelectedTemplateForManualAnswers(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.Local)
	message := renderExamNotificationMessage("{title}|{submitter}|{date}|{total}|{score}|{fullScore}|{result}", "月考", "张三", now, 3, 12, 100, true, true, "")
	want := "月考|张三|2026-11-02 10:00:00|3|待人工批改|待人工批改|待人工批改"
	if message != want {
		t.Fatalf("manual notification = %q, want %q", message, want)
	}
}

func TestRenderExamNotificationMessageUsesReviewPromptWithoutSelectedTemplate(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.Local)
	message := renderExamNotificationMessage("", "月考", "张三", now, 3, 12, 100, true, true, "")
	if message != "收到新答卷，请查看" {
		t.Fatalf("manual notification = %q", message)
	}
}

func TestRenderExamNotificationMessageIncludesAnswerDownloadURL(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.Local)
	message := renderExamNotificationMessage("考试 {title}\n下载：{answerDownloadUrl}", "月考", "张三", now, 1, 0, 100, true, true, "https://exam.example.com/download?token=secret")
	if message != "考试 月考\n下载：https://exam.example.com/download?token=secret" {
		t.Fatalf("message = %q", message)
	}
}

type downloadLinkIssuerStub struct {
	calls int
	input exportlinkapp.IssueInput
	link  string
}

func (issuer *downloadLinkIssuerStub) Issue(_ context.Context, input exportlinkapp.IssueInput) (string, error) {
	issuer.calls++
	issuer.input = input
	return issuer.link, nil
}

func TestIssueExamAnswerDownloadURLOnlyWhenTemplateUsesPlaceholder(t *testing.T) {
	issuer := &downloadLinkIssuerStub{link: "https://exam.example.com/download?token=secret"}
	settings := `{"publicUrl":"https://exam.example.com/ef1/5","answerDownloadBaseUrl":"https://api.example.com:8086","answerDownloadFormat":"xlsx","answerDownloadExpireDays":14}`
	link, err := issueExamAnswerDownloadURL(context.Background(), issuer, 5, 19, settings, []Rule{{MessageTemplate: "下载 {answerDownloadUrl}"}})
	if err != nil || link != issuer.link || issuer.calls != 1 {
		t.Fatalf("link=%q calls=%d err=%v", link, issuer.calls, err)
	}
	want := exportlinkapp.IssueInput{ExamID: 5, RecordID: 19, Format: "xlsx", ExpireDays: 14, BaseURL: "https://api.example.com:8086"}
	if issuer.input != want {
		t.Fatalf("issue input = %#v, want %#v", issuer.input, want)
	}

	issuer.calls = 0
	link, err = issueExamAnswerDownloadURL(context.Background(), issuer, 5, 19, settings, []Rule{{MessageTemplate: "普通通知"}})
	if err != nil || link != "" || issuer.calls != 0 {
		t.Fatalf("unused placeholder link=%q calls=%d err=%v", link, issuer.calls, err)
	}
}

func TestIssueExamAnswerDownloadURLUsesFormatAndExpiryDefaults(t *testing.T) {
	issuer := &downloadLinkIssuerStub{link: "link"}
	_, err := issueExamAnswerDownloadURL(context.Background(), issuer, 5, 19, `{"answerDownloadBaseUrl":"https://api.example.com"}`, []Rule{{MessageTemplate: "{answerDownloadUrl}"}})
	if err != nil {
		t.Fatal(err)
	}
	if issuer.input.Format != "pdf" || issuer.input.ExpireDays != 7 {
		t.Fatalf("legacy defaults = %#v", issuer.input)
	}
}

func TestIssueExamAnswerDownloadURLDoesNotFallbackToExamPublicURL(t *testing.T) {
	issuer := &downloadLinkIssuerStub{link: "link"}
	_, err := issueExamAnswerDownloadURL(context.Background(), issuer, 5, 19, `{"publicUrl":"https://exam.example.com/ef/5"}`, []Rule{{MessageTemplate: "{answerDownloadUrl}"}})
	if !errors.Is(err, exportlinkapp.ErrInvalidGrant) {
		t.Fatalf("missing download service address error = %v", err)
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
