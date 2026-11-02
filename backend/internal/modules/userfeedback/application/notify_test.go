package application

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func notificationCommand() NotifySubmitterCommand {
	return NotifySubmitterCommand{FeedbackID: 101, AdminID: 77, Content: "  请确认页面是否已恢复  ", Version: 3, RequestID: " notify-1 "}
}

func TestNotifySubmitterPreservesStateAndQueuesOnlySubmitter(t *testing.T) {
	for _, status := range []domain.Status{domain.StatusPending, domain.StatusProcessing, domain.StatusResolved, domain.StatusClosed} {
		t.Run(string(status), func(t *testing.T) {
			store := newFakeStore()
			events := []string{}
			store.events = &events
			handlerID := uint(88)
			resolvedAt, closedAt := int64(100), int64(200)
			store.locked = &FeedbackSnapshot{ID: 101, FeedbackNo: "FB-TEST", SubmitterID: 9, Status: status, Version: 3, HandlerID: &handlerID, ResolvedAt: &resolvedAt, ClosedAt: &closedAt}
			store.adminDetail = feedbackDetail(101, 9, status, 4)
			service := newServiceWithClock(store, nil, func() time.Time { return time.UnixMilli(300) }, mustShanghai(t))
			detail, err := service.NotifySubmitter(context.Background(), notificationCommand())
			if err != nil || detail == nil || detail.Version != 4 {
				t.Fatalf("NotifySubmitter() = %#v, %v", detail, err)
			}
			wantSnapshot := *store.locked
			wantSnapshot.Version, wantSnapshot.LastActivityAt, wantSnapshot.UpdatedAt = 4, 300, 300
			if !reflect.DeepEqual(store.updatedSnapshot, wantSnapshot) || store.updateExpectedVersion != 3 {
				t.Fatalf("notification changed business state: %#v", store.updatedSnapshot)
			}
			message := store.appendedMessage
			if message.MessageType != domain.MessageTypeNotification || message.AuthorType != domain.AuthorTypeAdmin || message.AuthorID != 77 || message.Content != "请确认页面是否已恢复" || message.FromStatus != "" || message.ToStatus != "" || message.RequestID != "notify-1" {
				t.Fatalf("notification history = %#v", message)
			}
			if len(store.enqueued) != 1 {
				t.Fatalf("outbox = %#v", store.enqueued)
			}
			outbox := store.enqueued[0]
			if outbox.RecipientUserID != 9 || outbox.SourceID != "101" || outbox.SourceType != NotificationSourceUserFeedback || outbox.Channel != NotificationChannelInternal || outbox.IdempotencyKey != "user-feedback-notify:101:201" || outbox.Title != "反馈 FB-TEST 新消息" || outbox.Content != message.Content {
				t.Fatalf("notification routing/content = %#v", outbox)
			}
			wantEvents := []string{"find-message-replay", "transaction", "find-message-replay", "lock-feedback", "append-message", "update-snapshot", "enqueue-notification"}
			if store.writeOutsideTransaction || !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("writes outside transaction=%v events=%v", store.writeOutsideTransaction, events)
			}
		})
	}
}

func TestNotifySubmitterRejectsInvalidInputBeforeWrites(t *testing.T) {
	for _, modify := range []func(*NotifySubmitterCommand){
		func(c *NotifySubmitterCommand) { c.FeedbackID = 0 },
		func(c *NotifySubmitterCommand) { c.AdminID = 0 },
		func(c *NotifySubmitterCommand) { c.Version = 0 },
		func(c *NotifySubmitterCommand) { c.RequestID = "" },
		func(c *NotifySubmitterCommand) { c.RequestID = strings.Repeat("a", 65) },
		func(c *NotifySubmitterCommand) { c.Content = " \n " },
		func(c *NotifySubmitterCommand) { c.Content = strings.Repeat("字", 5001) },
		func(c *NotifySubmitterCommand) { c.Content = string([]byte{0xff}) },
		func(c *NotifySubmitterCommand) { c.Channel = "webhook" },
	} {
		store := newFakeStore()
		service := newServiceWithClock(store, nil, time.Now, mustShanghai(t))
		command := notificationCommand()
		modify(&command)
		if _, err := service.NotifySubmitter(context.Background(), command); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("invalid notification accepted: %v", err)
		}
		if store.transactionCalls != 0 {
			t.Fatal("invalid input started transaction")
		}
	}
}

type recipientCheckerFake struct {
	available bool
	err       error
	userID    uint
	store     *fakeStore
	t         *testing.T
}

func (checker *recipientCheckerFake) CanDeliverTo(_ context.Context, userID uint) (bool, error) {
	checker.userID = userID
	if checker.store.inTransaction {
		checker.t.Fatal("binding lookup inside feedback transaction")
	}
	return checker.available, checker.err
}

func TestNotifySubmitterDingTalkPreflightAndChannel(t *testing.T) {
	for _, available := range []bool{false, true} {
		store := newFakeStore()
		store.locked = &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusPending, Version: 3}
		store.adminDetail = feedbackDetail(101, 9, domain.StatusPending, 3)
		checker := &recipientCheckerFake{available: available, store: store, t: t}
		service := newServiceWithClock(store, nil, time.Now, mustShanghai(t))
		service.dingTalkRecipientChecker = checker
		command := notificationCommand()
		command.Channel = "dingtalk"
		_, err := service.NotifySubmitter(context.Background(), command)
		if checker.userID != 9 {
			t.Fatalf("checked user=%d", checker.userID)
		}
		if !available {
			if !errors.Is(err, ErrDingTalkRecipientUnavailable) || store.transactionCalls != 0 {
				t.Fatalf("unavailable recipient: error=%v tx=%d", err, store.transactionCalls)
			}
		} else if err != nil || len(store.enqueued) != 1 || store.enqueued[0].Channel != "dingtalk" || store.enqueued[0].RecipientUserID != 9 {
			t.Fatalf("DingTalk notification: error=%v outbox=%#v", err, store.enqueued)
		}
	}
}

func TestNotifySubmitterRejectsMissingFeedbackOrStaleVersion(t *testing.T) {
	for _, missing := range []bool{true, false} {
		store := newFakeStore()
		want := ErrFeedbackNotFound
		if !missing {
			store.locked = &FeedbackSnapshot{ID: 101, SubmitterID: 9, Version: 4, Status: domain.StatusPending}
			want = ErrVersionConflict
		}
		service := newServiceWithClock(store, nil, time.Now, mustShanghai(t))
		if _, err := service.NotifySubmitter(context.Background(), notificationCommand()); !errors.Is(err, want) {
			t.Fatalf("error=%v want=%v", err, want)
		}
		if store.appendMessageCalls != 0 || store.updateCalls != 0 || len(store.enqueued) != 0 {
			t.Fatal("rejected operation wrote data")
		}
	}
}

func TestNotifySubmitterReplayDoesNotSendAgain(t *testing.T) {
	for _, concurrent := range []bool{false, true} {
		store := newFakeStore()
		replay := feedbackDetail(101, 9, domain.StatusClosed, 5)
		store.messageReplays = []*FeedbackDetail{replay}
		if concurrent {
			store.messageReplays = []*FeedbackDetail{nil, replay}
		}
		service := newServiceWithClock(store, nil, time.Now, mustShanghai(t))
		got, err := service.NotifySubmitter(context.Background(), notificationCommand())
		if err != nil || got != replay {
			t.Fatalf("replay=%#v error=%v", got, err)
		}
		if store.appendMessageCalls != 0 || store.updateCalls != 0 || len(store.enqueued) != 0 {
			t.Fatal("replay wrote data")
		}
	}
}

func TestNotifySubmitterPropagatesTransactionFailure(t *testing.T) {
	want := errors.New("write failed")
	for _, step := range []string{"message", "snapshot", "outbox", "commit"} {
		t.Run(step, func(t *testing.T) {
			store := newFakeStore()
			store.locked = &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusPending, Version: 3}
			switch step {
			case "message":
				store.appendMessageErr = want
			case "snapshot":
				store.updateErr = want
			case "outbox":
				store.enqueueErr = want
			case "commit":
				store.transactionCommitErr = want
			}
			service := newServiceWithClock(store, nil, time.Now, mustShanghai(t))
			if _, err := service.NotifySubmitter(context.Background(), notificationCommand()); !errors.Is(err, want) {
				t.Fatalf("error=%v", err)
			}
			if store.writeOutsideTransaction {
				t.Fatal("write outside transaction")
			}
			if (step == "message" || step == "snapshot") && len(store.enqueued) != 0 {
				t.Fatal("notification queued after failure")
			}
		})
	}
}

func TestNotifySubmitterReconcilesUncertainCommit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := newFakeStore()
	store.locked = &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusPending, Version: 3}
	store.transactionCommitErr = ErrTransactionOutcomeUnknown
	store.transactionReturnHook = cancel
	replay := feedbackDetail(101, 9, domain.StatusPending, 4)
	store.messageReplays = []*FeedbackDetail{nil, nil, replay}
	service := newServiceWithClock(store, nil, time.Now, mustShanghai(t))
	got, err := service.NotifySubmitter(ctx, notificationCommand())
	if err != nil || got != replay || len(store.enqueued) != 1 {
		t.Fatalf("reconciliation=%#v, %v outbox=%d", got, err, len(store.enqueued))
	}
}
