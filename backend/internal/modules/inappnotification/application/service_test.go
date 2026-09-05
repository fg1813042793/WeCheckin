package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"wecheckin/backend/internal/support/notificationstyle"
)

func TestValidateSendInput(t *testing.T) {
	tests := []struct {
		name  string
		input SendInput
		want  error
	}{
		{name: "title required", input: SendInput{Content: "content", Scope: ScopeAll, SourceType: SourceAdminManual, SourceID: "request-1"}, want: ErrTitleRequired},
		{name: "title too long", input: SendInput{Title: string(make([]byte, 256)), Content: "content", Scope: ScopeAll, SourceType: SourceAdminManual, SourceID: "request-1"}, want: ErrTitleTooLong},
		{name: "content required", input: SendInput{Title: "title", Scope: ScopeAll, SourceType: SourceAdminManual, SourceID: "request-1"}, want: ErrContentRequired},
		{name: "invalid scope", input: SendInput{Title: "title", Content: "content", Scope: "team", SourceType: SourceAdminManual, SourceID: "request-1"}, want: ErrInvalidScope},
		{name: "users required", input: SendInput{Title: "title", Content: "content", Scope: ScopeUsers, SourceType: SourceAdminManual, SourceID: "request-1"}, want: ErrRecipientsRequired},
		{name: "departments required", input: SendInput{Title: "title", Content: "content", Scope: ScopeDepartments, SourceType: SourceAdminManual, SourceID: "request-1"}, want: ErrRecipientsRequired},
		{name: "source required", input: SendInput{Title: "title", Content: "content", Scope: ScopeAll}, want: ErrSourceRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateSendInput(tt.input); !errors.Is(err, tt.want) {
				t.Fatalf("ValidateSendInput() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestServiceSendNormalizesRecipientsAndReturnsDeliveryResult(t *testing.T) {
	store := &fakeStore{
		resolution: RecipientResolution{UserIDs: []uint{9, 4, 9}, SkippedCount: 2},
		delivery:   DeliveryResult{SentCount: 2},
	}
	service := NewService(store)

	result, err := service.Send(context.Background(), SendInput{
		Title: "  系统通知  ", Content: "内容", Scope: ScopeUsers,
		UserIDs: []uint{9, 4, 9, 0}, SourceType: SourceAdminManual, SourceID: "request-1", DeliveryKey: "delivery-1",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !reflect.DeepEqual(store.rule.UserIDs, []uint{4, 9}) {
		t.Fatalf("resolved user IDs = %v, want [4 9]", store.rule.UserIDs)
	}
	if store.batch.Title != "系统通知" || store.batch.DeliveryKey != "delivery-1" || !reflect.DeepEqual(store.batch.UserIDs, []uint{4, 9}) {
		t.Fatalf("delivery batch = %#v", store.batch)
	}
	if result.PlannedCount != 2 || result.SentCount != 2 || result.SkippedCount != 2 || result.Replayed {
		t.Fatalf("send result = %#v", result)
	}
}

func TestServiceSendUsesExplicitNotificationTypeForStylePreview(t *testing.T) {
	store := &fakeStore{
		resolution: RecipientResolution{UserIDs: []uint{4}},
		delivery:   DeliveryResult{SentCount: 1},
	}
	_, err := NewService(store).Send(context.Background(), SendInput{
		Title: "退回通知测试", Content: "测试正文", Scope: ScopeUsers, UserIDs: []uint{4},
		SourceType: SourceAdminManual, SourceID: "request-style-1", NotificationType: "approval_result_returned",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if store.batch.Type != "approval_result_returned" {
		t.Fatalf("delivery type = %q, want approval_result_returned", store.batch.Type)
	}
}

func TestValidateSendInputRejectsUnsupportedExplicitNotificationType(t *testing.T) {
	err := ValidateSendInput(SendInput{
		Title: "title", Content: "content", Scope: ScopeAll,
		SourceType: SourceAdminManual, SourceID: "request-1", NotificationType: "unknown_type",
	})
	if !errors.Is(err, ErrInvalidNotificationType) {
		t.Fatalf("ValidateSendInput() error = %v, want ErrInvalidNotificationType", err)
	}
}

func TestServiceSendDingTalkCombinesRecipientAndChannelResults(t *testing.T) {
	store := &fakeStore{
		resolution: RecipientResolution{UserIDs: []uint{9, 4, 9}, SkippedCount: 2},
	}
	delivery := &fakeDingTalkDelivery{
		result: DingTalkDeliveryResult{SentCount: 1, SkippedCount: 1, FailedCount: 1},
	}
	service := NewServiceWithDingTalk(store, delivery)

	result, err := service.SendDingTalk(context.Background(), SendInput{
		Title: "  钉钉通知  ", Content: "内容", Scope: ScopeUsers,
		UserIDs: []uint{9, 4, 9, 0}, SourceType: SourceAdminManualDingTalk, SourceID: "request-1",
		NotificationType: notificationstyle.TypeTaskReminder,
	})
	if err != nil {
		t.Fatalf("SendDingTalk() error = %v", err)
	}
	if !reflect.DeepEqual(store.rule.UserIDs, []uint{4, 9}) {
		t.Fatalf("resolved user IDs = %v, want [4 9]", store.rule.UserIDs)
	}
	if delivery.batch.Title != "钉钉通知" || delivery.batch.NotificationType != notificationstyle.TypeTaskReminder || !reflect.DeepEqual(delivery.batch.UserIDs, []uint{4, 9}) {
		t.Fatalf("delivery batch = %#v", delivery.batch)
	}
	if result.PlannedCount != 2 || result.SentCount != 1 || result.SkippedCount != 3 || result.FailedCount != 1 {
		t.Fatalf("send result = %#v", result)
	}
}

func TestServiceSendDingTalkRequiresConfiguredDelivery(t *testing.T) {
	service := NewService(&fakeStore{resolution: RecipientResolution{UserIDs: []uint{4}}})
	_, err := service.SendDingTalk(context.Background(), SendInput{
		Title: "通知", Content: "内容", Scope: ScopeAll,
		SourceType: SourceAdminManualDingTalk, SourceID: "request-1",
	})
	if !errors.Is(err, ErrDingTalkDeliveryUnavailable) {
		t.Fatalf("SendDingTalk() error = %v, want ErrDingTalkDeliveryUnavailable", err)
	}
}

func TestServiceSendRejectsEmptyResolvedRecipients(t *testing.T) {
	service := NewService(&fakeStore{})
	_, err := service.Send(context.Background(), SendInput{
		Title: "title", Content: "content", Scope: ScopeAll,
		SourceType: SourceScheduledTaskRun, SourceID: "run-1",
	})
	if !errors.Is(err, ErrNoRecipients) {
		t.Fatalf("Send() error = %v, want ErrNoRecipients", err)
	}
}

func TestServiceSendReturnsReplay(t *testing.T) {
	store := &fakeStore{
		resolution: RecipientResolution{UserIDs: []uint{3}},
		delivery:   DeliveryResult{SentCount: 1, Replayed: true},
	}
	result, err := NewService(store).Send(context.Background(), SendInput{
		Title: "title", Content: "content", Scope: ScopeAll,
		SourceType: SourceScheduledTaskRun, SourceID: "run-1",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !result.Replayed || result.SentCount != 1 {
		t.Fatalf("send replay result = %#v", result)
	}
}

func TestServiceSendPropagatesStoreFailure(t *testing.T) {
	want := errors.New("database unavailable")
	store := &fakeStore{resolveErr: want}
	_, err := NewService(store).Send(context.Background(), SendInput{
		Title: "title", Content: "content", Scope: ScopeAll,
		SourceType: SourceScheduledTaskRun, SourceID: "run-1",
	})
	if !errors.Is(err, want) {
		t.Fatalf("Send() error = %v, want %v", err, want)
	}
}

func TestServiceListRecordsNormalizesFiltersAndPagination(t *testing.T) {
	store := &fakeStore{notifications: []Notification{{ID: 3, Title: "通知"}}, total: 1}
	readStatus := 0
	result, err := NewService(store).ListRecords(context.Background(), NotificationRecordQuery{
		Title: "  系统通知  ", RecipientName: "  张三  ", SourceType: " workflow ", Type: " task_arrived ",
		IsRead: &readStatus, AddTimeFrom: 100, AddTimeTo: 200, Page: 0, PageSize: 999,
	})
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
	if store.recordQuery.Title != "系统通知" || store.recordQuery.RecipientName != "张三" ||
		store.recordQuery.SourceType != "workflow" || store.recordQuery.Type != "task_arrived" ||
		store.recordQuery.IsRead == nil || *store.recordQuery.IsRead != 0 ||
		store.recordQuery.AddTimeFrom != 100 || store.recordQuery.AddTimeTo != 200 ||
		store.recordQuery.Page != 1 || store.recordQuery.PageSize != 20 {
		t.Fatalf("record query = %#v", store.recordQuery)
	}
	if result.Total != 1 || len(result.List) != 1 || result.Page != 1 || result.PageSize != 20 {
		t.Fatalf("list result = %#v", result)
	}
}

func TestServiceListRecordsRejectsInvalidReadStatusAndTimeRange(t *testing.T) {
	invalidReadStatus := 2
	service := NewService(&fakeStore{})
	if _, err := service.ListRecords(context.Background(), NotificationRecordQuery{IsRead: &invalidReadStatus}); !errors.Is(err, ErrInvalidReadStatus) {
		t.Fatalf("invalid read status error = %v", err)
	}
	if _, err := service.ListRecords(context.Background(), NotificationRecordQuery{AddTimeFrom: 200, AddTimeTo: 100}); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("invalid time range error = %v", err)
	}
}

func TestServiceDeleteRecordSoftDeletesWithAdminIdentity(t *testing.T) {
	store := &fakeStore{deleteRecordFound: true}
	service := NewService(store)

	if err := service.DeleteRecord(context.Background(), 66, 7); err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}
	if store.deleteRecordID != 7 || store.deleteRecordActorID != "66" || store.deleteRecordAt <= 0 {
		t.Fatalf("delete record id=%d actor=%q at=%d", store.deleteRecordID, store.deleteRecordActorID, store.deleteRecordAt)
	}
}

func TestServiceDeleteRecordRejectsMissingRecordAndUnauthenticatedAdmin(t *testing.T) {
	service := NewService(&fakeStore{})
	if err := service.DeleteRecord(context.Background(), 66, 7); !errors.Is(err, ErrNotificationMissing) {
		t.Fatalf("DeleteRecord() missing error = %v, want ErrNotificationMissing", err)
	}
	if err := service.DeleteRecord(context.Background(), 0, 7); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("DeleteRecord() unauthenticated error = %v, want ErrUnauthenticated", err)
	}
}

func TestServiceMarkReadRejectsNotificationOutsideCurrentUsersInbox(t *testing.T) {
	store := &fakeStore{markReadFound: false}
	err := NewService(store).MarkRead(context.Background(), 66, 7)
	if !errors.Is(err, ErrNotificationMissing) {
		t.Fatalf("MarkRead() error = %v, want ErrNotificationMissing", err)
	}
	if store.inboxUserID != "66" || store.notificationID != 7 {
		t.Fatalf("mark read query user=%q notification=%d", store.inboxUserID, store.notificationID)
	}
}

func TestServiceInboxRequiresAuthenticatedUser(t *testing.T) {
	service := NewService(&fakeStore{})
	if _, err := service.UnreadCount(context.Background(), 0); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("UnreadCount() error = %v, want ErrUnauthenticated", err)
	}
}

type fakeStore struct {
	rule                RecipientRule
	batch               DeliveryBatch
	resolution          RecipientResolution
	delivery            DeliveryResult
	resolveErr          error
	deliverErr          error
	notifications       []Notification
	total               int64
	recordQuery         NotificationRecordQuery
	inboxUserID         string
	notificationID      uint
	markReadFound       bool
	deleteRecordID      uint
	deleteRecordActorID string
	deleteRecordAt      int64
	deleteRecordFound   bool
}

type fakeDingTalkDelivery struct {
	batch  DingTalkDeliveryBatch
	result DingTalkDeliveryResult
	err    error
}

func (delivery *fakeDingTalkDelivery) DeliverDingTalk(_ context.Context, batch DingTalkDeliveryBatch) (DingTalkDeliveryResult, error) {
	delivery.batch = batch
	return delivery.result, delivery.err
}

func (store *fakeStore) ResolveRecipients(_ context.Context, rule RecipientRule) (RecipientResolution, error) {
	store.rule = rule
	return store.resolution, store.resolveErr
}

func (store *fakeStore) Deliver(_ context.Context, batch DeliveryBatch) (DeliveryResult, error) {
	store.batch = batch
	return store.delivery, store.deliverErr
}

func (store *fakeStore) ListRecords(_ context.Context, query NotificationRecordQuery) ([]Notification, int64, error) {
	store.recordQuery = query
	return store.notifications, store.total, nil
}

func (store *fakeStore) SoftDeleteRecord(_ context.Context, notificationID uint, actorID string, deletedAt int64) (bool, error) {
	store.deleteRecordID = notificationID
	store.deleteRecordActorID = actorID
	store.deleteRecordAt = deletedAt
	return store.deleteRecordFound, nil
}

func (store *fakeStore) UnreadCount(_ context.Context, userID string) (int64, error) {
	store.inboxUserID = userID
	return store.total, nil
}

func (store *fakeStore) MarkRead(_ context.Context, userID string, notificationID uint) (bool, error) {
	store.inboxUserID = userID
	store.notificationID = notificationID
	return store.markReadFound, nil
}

func (store *fakeStore) MarkAllRead(_ context.Context, userID string) error {
	store.inboxUserID = userID
	return nil
}

func (store *fakeStore) RecipientOptions(context.Context) (RecipientOptions, error) {
	return RecipientOptions{}, nil
}

func (store *fakeStore) NotificationStyles(context.Context) (notificationstyle.Config, error) {
	return notificationstyle.DefaultConfig(), nil
}

func (store *fakeStore) SaveNotificationStyles(_ context.Context, config notificationstyle.Config) (notificationstyle.Config, error) {
	return config, nil
}
