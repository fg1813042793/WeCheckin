package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestSubmitSelfSchedulesManagerDingTalkNotification(t *testing.T) {
	src, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	body := functionBody(string(src), "func SubmitSelfContext")
	for _, want := range []string{
		"notifyReviewTransitionAsync(ctx, submittedReview, user, dingtalkH5NotifyEventSelfSubmitted)",
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

func TestDingTalkH5NotificationSourceHasSwitchAndBindingLookup(t *testing.T) {
	src, err := os.ReadFile("notification.go")
	if err != nil {
		t.Fatalf("read notification.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"DINGTALK_H5_NOTIFY_ENABLED",
		"func DingTalkH5NotificationEnabledContext",
		"func notifyReviewTransitionAsync",
		"func resolveDingTalkH5NotifyRecipientContext",
		"model.DingTalkH5UserBinding",
		"manager_account",
		"SendWorkNotificationContext",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("DingTalk H5 notification should include %q", want)
		}
	}
}
