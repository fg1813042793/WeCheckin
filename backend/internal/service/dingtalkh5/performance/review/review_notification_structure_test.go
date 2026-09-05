package review

import (
	"os"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	reviewnotification "wecheckin/backend/internal/service/dingtalkh5/performance/review/notification"
)

func TestSubmitSelfSchedulesManagerDingTalkNotification(t *testing.T) {
	src, err := os.ReadFile("review_employee.go")
	if err != nil {
		t.Fatalf("read review_employee.go: %v", err)
	}
	body := functionBody(string(src), "func SubmitSelfContext")
	for _, want := range []string{
		"reviewnotification.TransitionAsync(ctx, submittedReview, user, reviewnotification.EventSelfSubmitted)",
		"nextStatusAfterSelfSubmit",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("SubmitSelfContext should schedule DingTalk manager notification with %q", want)
		}
	}
	if strings.Contains(body, "SendWorkNotificationContext") {
		t.Fatalf("SubmitSelfContext should not call DingTalk network API synchronously")
	}
}

func TestReviewNextHandlerAccountUsesCurrentStatus(t *testing.T) {
	base := model.DingTalkH5PerfReview{
		EmployeeAccount:     "employee",
		ManagerAccount:      "manager",
		HRBPAccount:         "hrbp",
		HRBPReviewerAccount: "reviewer",
	}
	tests := []struct {
		name     string
		status   string
		override func(*model.DingTalkH5PerfReview)
		want     string
	}{
		{name: "draft returns to employee", status: ReviewStatusDraft, want: "employee"},
		{name: "manager review uses manager", status: ReviewStatusManagerReview, want: "manager"},
		{name: "hrbp review prefers reviewer", status: ReviewStatusHRBPReview, want: "reviewer"},
		{name: "employee confirm uses employee", status: ReviewStatusEmployeeConfirm, want: "employee"},
		{name: "hr final prefers reviewer", status: ReviewStatusHRFinal, want: "reviewer"},
		{name: "completed has no next handler", status: ReviewStatusCompleted, want: ""},
		{
			name:     "hrbp review falls back to configured hrbp",
			status:   ReviewStatusHRBPReview,
			override: func(review *model.DingTalkH5PerfReview) { review.HRBPReviewerAccount = "" },
			want:     "hrbp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := base
			review.Status = tt.status
			if tt.override != nil {
				tt.override(&review)
			}
			if got := reviewnotification.NextHandlerAccount(review); got != tt.want {
				t.Fatalf("NextHandlerAccount() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReviewTransitionActionsScheduleNextHandlerNotification(t *testing.T) {
	src, err := os.ReadFile("review_flow.go")
	if err != nil {
		t.Fatalf("read review_flow.go: %v", err)
	}
	reviewHelpers, err := os.ReadFile("review_helpers.go")
	if err != nil {
		t.Fatalf("read review_helpers.go: %v", err)
	}
	for _, fn := range []string{
		"func SubmitManagerContext",
		"func SubmitHRBPContext",
		"func ConfirmResultContext",
		"func DisputeResultContext",
		"func ReturnManagerContext",
		"func ReturnHRBPContext",
	} {
		body := functionBody(string(src), fn)
		if !strings.Contains(body, "reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)") {
			t.Fatalf("%s should schedule next-handler DingTalk notification after transition", fn)
		}
	}
	returnBody := functionBody(string(reviewHelpers), "func returnReview")
	if !strings.Contains(returnBody, "reviewnotification.TransitionAsync(ctx, transitionedReview, user, reviewnotification.EventFlowMoved)") {
		t.Fatalf("returnReview should schedule next-handler DingTalk notification after returning a review")
	}
}

func TestCreateReviewSchedulesEmployeeNotification(t *testing.T) {
	src, err := os.ReadFile("review_create.go")
	if err != nil {
		t.Fatalf("read review_create.go: %v", err)
	}
	body := functionBody(string(src), "func createReviewForEmployeeContext")
	if !strings.Contains(body, "Status:                  ReviewStatusDraft") {
		t.Fatalf("created review should start as draft so the employee is the next handler")
	}
	if !strings.Contains(body, "reviewnotification.TransitionAsync(ctx, review, user, reviewnotification.EventFlowMoved)") {
		t.Fatalf("createReviewForEmployeeContext should schedule employee DingTalk notification after creating a review")
	}
	if strings.Contains(body, "SendWorkNotificationContext") {
		t.Fatalf("createReviewForEmployeeContext should not call DingTalk network API synchronously")
	}
}

func TestDingTalkH5NotificationSourceHasSwitchAndBindingLookup(t *testing.T) {
	src, err := os.ReadFile("notification/notification.go")
	if err != nil {
		t.Fatalf("read notification/notification.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"DINGTALK_H5_NOTIFY_ENABLED",
		"func EnabledContext",
		"func TransitionAsync",
		"func resolveRecipientContext",
		"func NextHandlerAccount",
		"model.DingTalkH5UserBinding",
		"recipient.Config.NotifyEnabled",
		"SendWorkNotificationContext",
		"[DingTalkH5Notify] schedule",
		"[DingTalkH5Notify] send config",
		"[DingTalkH5Notify] sent",
		"[DingTalkH5Notify] skip",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("DingTalk H5 notification should include %q", want)
		}
	}
	notifyBody := functionBody(text, "func TransitionAsync")
	if strings.Contains(notifyBody, "EnabledContext(ctx)") {
		t.Fatalf("notification scheduling should not use one global switch before resolving the enterprise app config")
	}
	sendBody := functionBody(text, "func sendReviewTransitionContext")
	for _, want := range []string{
		"nextAccount := NextHandlerAccount(review)",
		"resolveRecipientContext(ctx, review.EmployeeAccount, nextAccount)",
	} {
		if !strings.Contains(sendBody, want) {
			t.Fatalf("DingTalk H5 notification send path should use next handler account with %q", want)
		}
	}
}
