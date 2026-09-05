package application

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func TestCreateFeedbackUsesShanghaiCalendarDayAndBuildsNumber(t *testing.T) {
	now := time.Date(2026, time.September, 5, 16, 0, 0, 0, time.UTC)
	store := newFakeStore()
	store.nextSequence = 7
	store.userDetail = feedbackDetail(101, 9, domain.StatusPending, 1)
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, func() time.Time { return now }, mustShanghai(t))

	detail, err := service.CreateFeedback(context.Background(), CreateCommand{
		SubmitterID: 9,
		Content:     "  页面无法提交  ",
		RequestID:   "  create-1  ",
	})
	if err != nil {
		t.Fatalf("CreateFeedback() error = %v", err)
	}
	if detail == nil || detail.ID != 101 || !detail.AllowsSupplement {
		t.Fatalf("CreateFeedback() detail = %#v", detail)
	}
	if store.nextDateKey != "2026-09-06" {
		t.Fatalf("NextFeedbackNumber dateKey = %q, want 2026-09-06", store.nextDateKey)
	}
	wantStart := time.Date(2026, time.September, 6, 0, 0, 0, 0, mustShanghai(t)).UTC().UnixMilli()
	wantEnd := time.Date(2026, time.September, 7, 0, 0, 0, 0, mustShanghai(t)).UTC().UnixMilli()
	if store.countUserID != 9 || store.countStartMs != wantStart || store.countEndMs != wantEnd {
		t.Fatalf("CountCreatedByUserOnDate args = (%d, %d, %d), want (9, %d, %d)", store.countUserID, store.countStartMs, store.countEndMs, wantStart, wantEnd)
	}
	if store.createdRecord.FeedbackNo != "FB-20260906-0007" {
		t.Fatalf("FeedbackNo = %q, want FB-20260906-0007", store.createdRecord.FeedbackNo)
	}
	if store.createdRecord.CreateRequestID != "create-1" || store.appendedMessage.Content != "页面无法提交" {
		t.Fatalf("normalized records = %#v / %#v", store.createdRecord, store.appendedMessage)
	}
	if store.createdRecord.CreatedAt != now.UnixMilli() || store.createdRecord.LastActivityAt != now.UnixMilli() {
		t.Fatalf("created timestamps = %#v, want %d", store.createdRecord, now.UnixMilli())
	}
}

func TestCreateFeedbackRejectsEmptyContentAndDailyLimitBeforeStorage(t *testing.T) {
	store := newFakeStore()
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

	_, err := service.CreateFeedback(context.Background(), CreateCommand{SubmitterID: 1, Content: " \n ", RequestID: "req"})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("empty content error = %v, want ErrInvalidArgument", err)
	}
	if store.transactionCalls != 0 || storage.saveCalls != 0 {
		t.Fatalf("invalid request side effects: transactions=%d saves=%d", store.transactionCalls, storage.saveCalls)
	}

	store.createdOnDate = MaxFeedbacksPerUserPerDay
	_, err = service.CreateFeedback(context.Background(), CreateCommand{
		SubmitterID: 1,
		Content:     "feedback",
		RequestID:   "req-limit",
		Attachments: []AttachmentInput{validImageInput("limit.png", "image/png")},
	})
	if !errors.Is(err, ErrDailyLimitExceeded) {
		t.Fatalf("daily limit error = %v, want ErrDailyLimitExceeded", err)
	}
	if storage.saveCalls != 1 || storage.deleteCalls != 1 {
		t.Fatalf("daily-limit compensation saves=%d deletes=%d, want 1/1", storage.saveCalls, storage.deleteCalls)
	}
	if store.createCalls != 0 {
		t.Fatalf("CreateFeedback store calls = %d, want 0", store.createCalls)
	}
}

func TestCreateFeedbackReplayReturnsFirstDetailWithoutWrites(t *testing.T) {
	store := newFakeStore()
	replay := feedbackDetail(55, 3, domain.StatusProcessing, 4)
	store.createReplays = []*FeedbackDetail{replay}
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

	detail, err := service.CreateFeedback(context.Background(), CreateCommand{
		SubmitterID: 3,
		Content:     "retry",
		RequestID:   " same-request ",
		Attachments: []AttachmentInput{validImageInput("retry.png", "image/png")},
	})
	if err != nil {
		t.Fatalf("CreateFeedback(replay) error = %v", err)
	}
	if detail != replay || detail.AllowsSupplement != domain.StatusProcessing.AllowsSupplement() {
		t.Fatalf("replay detail = %#v", detail)
	}
	if storage.saveCalls != 0 || storage.deleteCalls != 0 || store.transactionCalls != 0 || store.createCalls != 0 || store.appendMessageCalls != 0 || store.appendAttachmentCalls != 0 {
		t.Fatalf("replay caused writes: storage=%d/%d tx=%d create=%d message=%d attachment=%d", storage.saveCalls, storage.deleteCalls, store.transactionCalls, store.createCalls, store.appendMessageCalls, store.appendAttachmentCalls)
	}
}

func TestCreateFeedbackConcurrentReplayDeletesOnlyImagesSavedByThisCall(t *testing.T) {
	store := newFakeStore()
	replay := feedbackDetail(56, 3, domain.StatusPending, 1)
	store.createReplays = []*FeedbackDetail{nil, replay}
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

	detail, err := service.CreateFeedback(context.Background(), CreateCommand{
		SubmitterID: 3,
		Content:     "retry",
		RequestID:   "race-request",
		Attachments: []AttachmentInput{validImageInput("race.png", "image/png")},
	})
	if err != nil {
		t.Fatalf("CreateFeedback(concurrent replay) error = %v", err)
	}
	if detail != replay {
		t.Fatalf("detail = %#v, want replay", detail)
	}
	if storage.saveCalls != 1 || storage.deleteCalls != 1 {
		t.Fatalf("storage calls = save %d delete %d, want 1/1", storage.saveCalls, storage.deleteCalls)
	}
	if !reflect.DeepEqual(storage.deleted, storage.saved) {
		t.Fatalf("deleted images = %#v, want only this call's saved images %#v", storage.deleted, storage.saved)
	}
	if store.createCalls != 0 || store.appendMessageCalls != 0 || store.appendAttachmentCalls != 0 {
		t.Fatalf("concurrent replay caused DB writes: create=%d message=%d attachment=%d", store.createCalls, store.appendMessageCalls, store.appendAttachmentCalls)
	}
}

func TestCreateFeedbackCompensatesPartialSaveAndTransactionCommitFailure(t *testing.T) {
	t.Run("partial image save failure", func(t *testing.T) {
		store := newFakeStore()
		storage := &fakeImageStorage{saveErrAt: 2, saveErr: errors.New("storage unavailable")}
		service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

		_, err := service.CreateFeedback(context.Background(), CreateCommand{
			SubmitterID: 8,
			Content:     "with images",
			RequestID:   "partial-save",
			Attachments: validImageInputs(2),
		})
		if !errors.Is(err, ErrStorageFailed) {
			t.Fatalf("CreateFeedback() error = %v, want ErrStorageFailed", err)
		}
		if storage.saveCalls != 2 || storage.deleteCalls != 1 || len(storage.deleted) != 1 || storage.deleted[0] != storage.saved[0] {
			t.Fatalf("partial-save compensation saved=%#v deleted=%#v calls=%d/%d", storage.saved, storage.deleted, storage.saveCalls, storage.deleteCalls)
		}
		if store.transactionCalls != 0 {
			t.Fatalf("transactions = %d, want 0", store.transactionCalls)
		}
	})

	t.Run("transaction commit failure", func(t *testing.T) {
		store := newFakeStore()
		store.transactionCommitErr = errors.New("commit failed")
		storage := &fakeImageStorage{}
		service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

		_, err := service.CreateFeedback(context.Background(), CreateCommand{
			SubmitterID: 8,
			Content:     "with image",
			RequestID:   "commit-failure",
			Attachments: validImageInputs(1),
		})
		if !errors.Is(err, store.transactionCommitErr) {
			t.Fatalf("CreateFeedback() error = %v, want commit failure", err)
		}
		if storage.saveCalls != 1 || storage.deleteCalls != 1 || !reflect.DeepEqual(storage.deleted, storage.saved) {
			t.Fatalf("commit-failure compensation saved=%#v deleted=%#v", storage.saved, storage.deleted)
		}
	})
}

func TestCreateFeedbackRejectsInvalidIdentityAndRequestID(t *testing.T) {
	service := newServiceWithClock(newFakeStore(), &fakeImageStorage{}, time.Now, mustShanghai(t))
	for _, command := range []CreateCommand{
		{SubmitterID: 0, Content: "feedback", RequestID: "request"},
		{SubmitterID: 1, Content: "feedback", RequestID: " \n "},
		{SubmitterID: 1, Content: "feedback", RequestID: string(make([]byte, MaxRequestIDRunes+1))},
		{SubmitterID: 1, Content: "feedback", RequestID: "bad" + string([]byte{0xff})},
	} {
		if _, err := service.CreateFeedback(context.Background(), command); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("CreateFeedback(%#v) error = %v, want ErrInvalidArgument", command, err)
		}
	}
}

func TestCreateFeedbackReturnsDiagnosticErrorForUninitializedDependencies(t *testing.T) {
	for _, service := range []*Service{nil, NewService(nil, &fakeImageStorage{}), NewService(newFakeStore(), nil)} {
		_, err := service.CreateFeedback(context.Background(), CreateCommand{SubmitterID: 1, Content: "feedback", RequestID: "request"})
		if !errors.Is(err, ErrServiceUnavailable) {
			t.Fatalf("CreateFeedback() error = %v, want ErrServiceUnavailable", err)
		}
	}
}

func TestUserQueriesAlwaysPassAuthenticatedUserID(t *testing.T) {
	store := newFakeStore()
	store.userOverview = Overview{Pending: 2}
	store.userList = FeedbackList{List: []FeedbackSummary{{ID: 17}}, Total: 1}
	store.userDetail = feedbackDetail(17, 42, domain.StatusResolved, 3)
	service := NewService(store, &fakeImageStorage{})

	overview, err := service.GetUserOverview(context.Background(), 42)
	if err != nil || overview.Pending != 2 || store.lastUserID != 42 {
		t.Fatalf("GetUserOverview() = %#v, %v; store user = %d", overview, err, store.lastUserID)
	}
	list, err := service.ListUserFeedbacks(context.Background(), 42, UserListQuery{
		Keyword: "  登录失败  ", Status: domain.StatusProcessing,
	})
	if err != nil {
		t.Fatalf("ListUserFeedbacks() error = %v", err)
	}
	if store.lastUserID != 42 || store.lastUserQuery.Keyword != "登录失败" || store.lastUserQuery.Page != 1 || store.lastUserQuery.PageSize != DefaultPageSize {
		t.Fatalf("ListUserFeedbacks store args = user %d query %#v", store.lastUserID, store.lastUserQuery)
	}
	if list.Page != 1 || list.PageSize != DefaultPageSize {
		t.Fatalf("ListUserFeedbacks result paging = %#v", list)
	}
	detail, err := service.GetUserFeedback(context.Background(), 17, 42)
	if err != nil || detail == nil {
		t.Fatalf("GetUserFeedback() = %#v, %v", detail, err)
	}
	if store.lastFeedbackID != 17 || store.lastUserID != 42 {
		t.Fatalf("GetUserFeedback store args = (%d, %d)", store.lastFeedbackID, store.lastUserID)
	}
	if detail.AllowsSupplement {
		t.Fatal("resolved detail AllowsSupplement = true, want false")
	}
}

func TestGetUserFeedbackHidesMissingAndUnauthorizedRecords(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, &fakeImageStorage{})

	for _, userID := range []uint{7, 8} {
		store.userDetail = nil
		_, err := service.GetUserFeedback(context.Background(), 99, userID)
		if !errors.Is(err, ErrFeedbackNotFound) {
			t.Fatalf("GetUserFeedback(user=%d) error = %v, want ErrFeedbackNotFound", userID, err)
		}
	}
	store.userDetail = feedbackDetail(99, 8, domain.StatusPending, 1)
	if _, err := service.GetUserFeedback(context.Background(), 99, 7); !errors.Is(err, ErrFeedbackNotFound) {
		t.Fatalf("GetUserFeedback(other owner's detail) error = %v, want ErrFeedbackNotFound", err)
	}
	if store.getAdminCalls != 0 {
		t.Fatalf("client detail used admin query %d times", store.getAdminCalls)
	}
}

func TestAdminQueriesNormalizeFiltersAndDecorateDetail(t *testing.T) {
	handlerID := uint(12)
	store := newFakeStore()
	store.adminOverview = Overview{Processing: 4}
	store.adminList = FeedbackList{Total: 3}
	store.adminDetail = feedbackDetail(88, 9, domain.StatusPending, 2)
	service := NewService(store, &fakeImageStorage{})
	query := AdminListQuery{
		Keyword:       "  上传  ",
		SubmitterID:   9,
		HandlerID:     &handlerID,
		Status:        domain.StatusPending,
		SubmittedFrom: 100,
		SubmittedTo:   200,
	}

	overview, err := service.GetAdminOverview(context.Background(), query)
	if err != nil || overview.Processing != 4 {
		t.Fatalf("GetAdminOverview() = %#v, %v", overview, err)
	}
	if store.lastAdminQuery.Keyword != "上传" || store.lastAdminQuery.Page != 1 || store.lastAdminQuery.PageSize != DefaultPageSize {
		t.Fatalf("GetAdminOverview normalized query = %#v", store.lastAdminQuery)
	}
	list, err := service.ListAdminFeedbacks(context.Background(), query)
	if err != nil || list.Page != 1 || list.PageSize != DefaultPageSize {
		t.Fatalf("ListAdminFeedbacks() = %#v, %v", list, err)
	}
	detail, err := service.GetAdminFeedback(context.Background(), 88)
	if err != nil || detail == nil || !detail.AllowsSupplement {
		t.Fatalf("GetAdminFeedback() = %#v, %v", detail, err)
	}
}

func TestQueriesValidateIDsKeywordsStatusesPagingAndAdminTimes(t *testing.T) {
	store := newFakeStore()
	storage := &fakeImageStorage{}
	service := NewService(store, storage)
	longKeyword := string(make([]byte, MaxKeywordRunes+1))
	zeroHandler := uint(0)

	userCases := []UserListQuery{
		{Keyword: longKeyword},
		{Keyword: "bad" + string([]byte{0xff})},
		{Status: domain.Status("invalid")},
		{Page: -1},
		{PageSize: -1},
		{PageSize: MaxPageSize + 1},
	}
	for _, query := range userCases {
		if _, err := service.ListUserFeedbacks(context.Background(), 1, query); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("ListUserFeedbacks(%#v) error = %v, want ErrInvalidArgument", query, err)
		}
	}
	adminCases := []AdminListQuery{
		{Keyword: longKeyword},
		{Status: domain.Status("invalid")},
		{HandlerID: &zeroHandler},
		{SubmittedFrom: -1},
		{SubmittedTo: -1},
		{SubmittedFrom: 200, SubmittedTo: 100},
		{Page: -1},
		{PageSize: MaxPageSize + 1},
	}
	for _, query := range adminCases {
		if _, err := service.ListAdminFeedbacks(context.Background(), query); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("ListAdminFeedbacks(%#v) error = %v, want ErrInvalidArgument", query, err)
		}
	}
	for _, call := range []func() error{
		func() error { _, err := service.GetUserOverview(context.Background(), 0); return err },
		func() error {
			_, err := service.ListUserFeedbacks(context.Background(), 0, UserListQuery{})
			return err
		},
		func() error { _, err := service.GetUserFeedback(context.Background(), 0, 1); return err },
		func() error { _, err := service.GetUserFeedback(context.Background(), 1, 0); return err },
		func() error { _, err := service.GetAdminFeedback(context.Background(), 0); return err },
	} {
		if err := call(); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("invalid query ID error = %v, want ErrInvalidArgument", err)
		}
	}
	if store.listUserCalls != 0 || store.listAdminCalls != 0 || storage.saveCalls != 0 || storage.deleteCalls != 0 {
		t.Fatalf("invalid query side effects: user=%d admin=%d storage=%d/%d", store.listUserCalls, store.listAdminCalls, storage.saveCalls, storage.deleteCalls)
	}
}

func TestQueryFailuresDoNotUseImageStorage(t *testing.T) {
	store := newFakeStore()
	store.queryErr = errors.New("query failed")
	storage := &fakeImageStorage{}
	service := NewService(store, storage)

	if _, err := service.GetUserOverview(context.Background(), 1); !errors.Is(err, store.queryErr) {
		t.Fatalf("GetUserOverview() error = %v", err)
	}
	if storage.saveCalls != 0 || storage.deleteCalls != 0 {
		t.Fatalf("query failure used storage: %d/%d", storage.saveCalls, storage.deleteCalls)
	}
}

func TestSupplementFeedbackLocksAndUpdatesPendingOrProcessingFeedback(t *testing.T) {
	for _, status := range []domain.Status{domain.StatusPending, domain.StatusProcessing} {
		t.Run(string(status), func(t *testing.T) {
			now := time.Date(2026, time.September, 5, 3, 4, 5, 0, time.UTC)
			events := make([]string, 0)
			store := newFakeStore()
			store.events = &events
			store.locked = &FeedbackSnapshot{
				ID: 101, FeedbackNo: "FB-20260905-0001", SubmitterID: 9,
				Status: status, Version: 3, AttachmentCount: 28,
			}
			store.userDetail = feedbackDetail(101, 9, status, 4)
			storage := &fakeImageStorage{events: &events}
			service := newServiceWithClock(store, storage, func() time.Time { return now }, mustShanghai(t))

			detail, err := service.SupplementFeedback(context.Background(), SupplementCommand{
				FeedbackID: 101, SubmitterID: 9, Version: 3,
				Content: "  更多信息  ", RequestID: " supplement-1 ", Attachments: validImageInputs(2),
			})
			if err != nil {
				t.Fatalf("SupplementFeedback() error = %v", err)
			}
			if detail == nil || detail.Version != 4 || detail.AllowsSupplement != status.AllowsSupplement() {
				t.Fatalf("SupplementFeedback() detail = %#v", detail)
			}
			wantEvents := []string{"find-message-replay", "save-image", "save-image", "transaction", "find-message-replay", "lock-feedback", "append-message", "append-attachments", "update-snapshot"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("operation order = %#v, want %#v", events, wantEvents)
			}
			if store.appendedMessage.MessageType != domain.MessageTypeSupplement || store.appendedMessage.AuthorType != domain.AuthorTypeUser || store.appendedMessage.AuthorID != 9 || store.appendedMessage.Content != "更多信息" {
				t.Fatalf("supplement message = %#v", store.appendedMessage)
			}
			if store.updateExpectedVersion != 3 || store.updatedSnapshot.Version != 4 || store.updatedSnapshot.AttachmentCount != 30 {
				t.Fatalf("snapshot update = %#v expectedVersion=%d", store.updatedSnapshot, store.updateExpectedVersion)
			}
			if store.updatedSnapshot.LastActivityAt != now.UnixMilli() || store.updatedSnapshot.UpdatedAt != now.UnixMilli() {
				t.Fatalf("snapshot timestamps = %#v, want %d", store.updatedSnapshot, now.UnixMilli())
			}
			if len(store.appendedAttachments) != 2 || store.appendedAttachments[0].ObjectKey != storage.saved[0].ObjectKey {
				t.Fatalf("attachment records = %#v", store.appendedAttachments)
			}
		})
	}
}

func TestSupplementFeedbackValidatesBeforeReplayOrStorage(t *testing.T) {
	store := newFakeStore()
	store.messageReplays = []*FeedbackDetail{feedbackDetail(1, 2, domain.StatusPending, 1)}
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))
	invalidImage := validImageInput("bad.png", "image/png")
	invalidImage.Content = []byte("not a png")

	cases := []SupplementCommand{
		{FeedbackID: 0, SubmitterID: 2, Version: 1, Content: "text", RequestID: "req"},
		{FeedbackID: 1, SubmitterID: 0, Version: 1, Content: "text", RequestID: "req"},
		{FeedbackID: 1, SubmitterID: 2, Version: 0, Content: "text", RequestID: "req"},
		{FeedbackID: 1, SubmitterID: 2, Version: 1, Content: "text", RequestID: " "},
		{FeedbackID: 1, SubmitterID: 2, Version: 1, Content: " ", RequestID: "req"},
		{FeedbackID: 1, SubmitterID: 2, Version: 1, Content: "text", RequestID: "req", Attachments: []AttachmentInput{invalidImage}},
	}
	for _, command := range cases {
		if _, err := service.SupplementFeedback(context.Background(), command); !errors.Is(err, ErrInvalidArgument) && !errors.Is(err, ErrAttachmentTypeNotAllowed) {
			t.Fatalf("SupplementFeedback(%#v) error = %v, want validation error", command, err)
		}
	}
	if store.messageReplayCalls != 0 || storage.saveCalls != 0 || store.transactionCalls != 0 {
		t.Fatalf("invalid supplement side effects: replay=%d save=%d tx=%d", store.messageReplayCalls, storage.saveCalls, store.transactionCalls)
	}
}

func TestSupplementFeedbackReplayDoesNotSaveOrWrite(t *testing.T) {
	store := newFakeStore()
	replay := feedbackDetail(101, 9, domain.StatusProcessing, 7)
	store.messageReplays = []*FeedbackDetail{replay}
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

	detail, err := service.SupplementFeedback(context.Background(), SupplementCommand{
		FeedbackID: 101, SubmitterID: 9, Version: 3,
		Content: "retry", RequestID: "same", Attachments: validImageInputs(1),
	})
	if err != nil || detail != replay {
		t.Fatalf("SupplementFeedback(replay) = %#v, %v", detail, err)
	}
	if storage.saveCalls != 0 || storage.deleteCalls != 0 || store.transactionCalls != 0 || store.appendMessageCalls != 0 || store.appendAttachmentCalls != 0 {
		t.Fatalf("replay writes: storage=%d/%d tx=%d message=%d attachments=%d", storage.saveCalls, storage.deleteCalls, store.transactionCalls, store.appendMessageCalls, store.appendAttachmentCalls)
	}
}

func TestSupplementFeedbackConcurrentReplayCompensatesNewImages(t *testing.T) {
	store := newFakeStore()
	replay := feedbackDetail(101, 9, domain.StatusPending, 4)
	store.messageReplays = []*FeedbackDetail{nil, replay}
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

	detail, err := service.SupplementFeedback(context.Background(), SupplementCommand{
		FeedbackID: 101, SubmitterID: 9, Version: 3,
		Content: "race", RequestID: "same", Attachments: validImageInputs(1),
	})
	if err != nil || detail != replay {
		t.Fatalf("SupplementFeedback(concurrent replay) = %#v, %v", detail, err)
	}
	if !reflect.DeepEqual(storage.deleted, storage.saved) || store.lockCalls != 0 || store.appendMessageCalls != 0 {
		t.Fatalf("concurrent replay state: saved=%#v deleted=%#v locks=%d messages=%d", storage.saved, storage.deleted, store.lockCalls, store.appendMessageCalls)
	}
}

func TestSupplementFeedbackRechecksOwnershipStatusVersionAndAttachmentLimitAfterLock(t *testing.T) {
	tests := []struct {
		name   string
		locked *FeedbackSnapshot
		want   error
	}{
		{name: "missing", locked: nil, want: ErrFeedbackNotFound},
		{name: "other owner", locked: &FeedbackSnapshot{ID: 101, SubmitterID: 10, Status: domain.StatusPending, Version: 3}, want: ErrFeedbackNotFound},
		{name: "resolved", locked: &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusResolved, Version: 3}, want: ErrSupplementNotAllowed},
		{name: "closed", locked: &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusClosed, Version: 3}, want: ErrSupplementNotAllowed},
		{name: "version changed", locked: &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusPending, Version: 4}, want: ErrVersionConflict},
		{name: "cumulative image limit", locked: &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusPending, Version: 3, AttachmentCount: 30}, want: ErrAttachmentLimitExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore()
			store.locked = test.locked
			storage := &fakeImageStorage{}
			service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

			_, err := service.SupplementFeedback(context.Background(), SupplementCommand{
				FeedbackID: 101, SubmitterID: 9, Version: 3,
				Content: "more", RequestID: "check", Attachments: validImageInputs(1),
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("SupplementFeedback() error = %v, want %v", err, test.want)
			}
			if storage.saveCalls != 1 || storage.deleteCalls != 1 || !reflect.DeepEqual(storage.deleted, storage.saved) {
				t.Fatalf("failure compensation saved=%#v deleted=%#v", storage.saved, storage.deleted)
			}
			if store.appendMessageCalls != 0 || store.appendAttachmentCalls != 0 || store.updateCalls != 0 {
				t.Fatalf("failure wrote message=%d attachments=%d update=%d", store.appendMessageCalls, store.appendAttachmentCalls, store.updateCalls)
			}
		})
	}
}

func TestSupplementFeedbackCompensatesTransactionAndWriteFailures(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*fakeStore)
		want      error
	}{
		{name: "append message", configure: func(store *fakeStore) { store.appendMessageErr = errors.New("message failed") }, want: errors.New("message failed")},
		{name: "append attachments", configure: func(store *fakeStore) { store.appendAttachmentErr = errors.New("attachments failed") }, want: errors.New("attachments failed")},
		{name: "snapshot", configure: func(store *fakeStore) { store.updateErr = ErrVersionConflict }, want: ErrVersionConflict},
		{name: "commit", configure: func(store *fakeStore) { store.transactionCommitErr = errors.New("commit failed") }, want: errors.New("commit failed")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore()
			store.locked = &FeedbackSnapshot{ID: 101, SubmitterID: 9, Status: domain.StatusPending, Version: 3}
			test.configure(store)
			storage := &fakeImageStorage{}
			service := newServiceWithClock(store, storage, time.Now, mustShanghai(t))

			_, err := service.SupplementFeedback(context.Background(), SupplementCommand{
				FeedbackID: 101, SubmitterID: 9, Version: 3,
				Content: "more", RequestID: "failure", Attachments: validImageInputs(1),
			})
			if test.name == "snapshot" {
				if !errors.Is(err, ErrVersionConflict) {
					t.Fatalf("SupplementFeedback() error = %v, want ErrVersionConflict", err)
				}
			} else if err == nil || err.Error() != test.want.Error() {
				t.Fatalf("SupplementFeedback() error = %v, want %v", err, test.want)
			}
			if !reflect.DeepEqual(storage.deleted, storage.saved) {
				t.Fatalf("failure compensation saved=%#v deleted=%#v", storage.saved, storage.deleted)
			}
		})
	}
}

func TestUpdateFeedbackStatusWritesMessageSnapshotAndNotificationInOneTransaction(t *testing.T) {
	now := time.Date(2026, time.September, 5, 8, 9, 10, 0, time.UTC)
	events := make([]string, 0)
	store := newFakeStore()
	store.events = &events
	store.locked = &FeedbackSnapshot{
		ID: 101, FeedbackNo: "FB-20260905-0007", SubmitterID: 9,
		Status: domain.StatusPending, Version: 3,
	}
	store.adminDetail = feedbackDetail(101, 9, domain.StatusProcessing, 4)
	storage := &fakeImageStorage{}
	service := newServiceWithClock(store, storage, func() time.Time { return now }, mustShanghai(t))

	detail, err := service.UpdateFeedbackStatus(context.Background(), UpdateStatusCommand{
		FeedbackID: 101,
		AdminID:    77,
		Status:     domain.StatusProcessing,
		Note:       "  已安排处理  ",
		NotifyUser: true,
		Version:    3,
		RequestID:  " status-1 ",
	})
	if err != nil {
		t.Fatalf("UpdateFeedbackStatus() error = %v", err)
	}
	if detail == nil || detail.Version != 4 {
		t.Fatalf("UpdateFeedbackStatus() detail = %#v", detail)
	}
	wantEvents := []string{"find-message-replay", "transaction", "find-message-replay", "lock-feedback", "append-message", "update-snapshot", "enqueue-notification"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("operation order = %#v, want %#v", events, wantEvents)
	}
	if store.writeOutsideTransaction {
		t.Fatal("status write occurred outside Store.InTransaction callback")
	}
	message := store.appendedMessage
	if message.MessageType != domain.MessageTypeStatus || message.AuthorType != domain.AuthorTypeAdmin || message.AuthorID != 77 || message.Content != "已安排处理" || message.FromStatus != domain.StatusPending || message.ToStatus != domain.StatusProcessing || message.RequestID != "status-1" {
		t.Fatalf("status message = %#v", message)
	}
	if store.updateExpectedVersion != 3 || store.updatedSnapshot.Version != 4 || store.updatedSnapshot.Status != domain.StatusProcessing || store.updatedSnapshot.HandlerID == nil || *store.updatedSnapshot.HandlerID != 77 {
		t.Fatalf("status snapshot = %#v expectedVersion=%d", store.updatedSnapshot, store.updateExpectedVersion)
	}
	if store.updatedSnapshot.LastActivityAt != now.UnixMilli() || store.updatedSnapshot.UpdatedAt != now.UnixMilli() {
		t.Fatalf("snapshot timestamps = %#v", store.updatedSnapshot)
	}
	if len(store.enqueued) != 1 {
		t.Fatalf("notifications = %#v, want one", store.enqueued)
	}
	outbox := store.enqueued[0]
	if outbox.Channel != NotificationChannelInternal || outbox.NotificationType != NotificationTypeFeedbackStatus || outbox.SourceType != NotificationSourceUserFeedback || outbox.SourceID != "101" || outbox.RecipientUserID != 9 {
		t.Fatalf("outbox routing = %#v", outbox)
	}
	if outbox.IdempotencyKey != "user-feedback-status:101:201" || outbox.Title != "反馈 FB-20260905-0007 处理中" || outbox.Content != "已安排处理" || outbox.CreatedAt != now.UnixMilli() {
		t.Fatalf("outbox content = %#v", outbox)
	}
	if storage.saveCalls != 0 || storage.deleteCalls != 0 {
		t.Fatalf("status update used image storage: %d/%d", storage.saveCalls, storage.deleteCalls)
	}
}

func TestUpdateFeedbackStatusWithoutNotificationDoesNotEnqueue(t *testing.T) {
	store := newFakeStore()
	store.locked = &FeedbackSnapshot{ID: 101, FeedbackNo: "FB-1", SubmitterID: 9, Status: domain.StatusPending, Version: 1}
	store.adminDetail = feedbackDetail(101, 9, domain.StatusProcessing, 2)
	service := newServiceWithClock(store, &fakeImageStorage{}, time.Now, mustShanghai(t))

	_, err := service.UpdateFeedbackStatus(context.Background(), UpdateStatusCommand{
		FeedbackID: 101, AdminID: 7, Status: domain.StatusProcessing,
		Version: 1, RequestID: "without-notify", NotifyUser: false,
	})
	if err != nil {
		t.Fatalf("UpdateFeedbackStatus() error = %v", err)
	}
	if len(store.enqueued) != 0 {
		t.Fatalf("notifications = %#v, want none", store.enqueued)
	}
}

func TestUpdateFeedbackStatusReplayDoesNotLockOrWrite(t *testing.T) {
	store := newFakeStore()
	replay := feedbackDetail(101, 9, domain.StatusResolved, 6)
	store.messageReplays = []*FeedbackDetail{replay}
	service := newServiceWithClock(store, &fakeImageStorage{}, time.Now, mustShanghai(t))

	detail, err := service.UpdateFeedbackStatus(context.Background(), UpdateStatusCommand{
		FeedbackID: 101, AdminID: 7, Status: domain.StatusResolved,
		Note: "done", NotifyUser: true, Version: 5, RequestID: "same-status",
	})
	if err != nil || detail != replay {
		t.Fatalf("UpdateFeedbackStatus(replay) = %#v, %v", detail, err)
	}
	if store.transactionCalls != 0 || store.lockCalls != 0 || store.appendMessageCalls != 0 || store.updateCalls != 0 || len(store.enqueued) != 0 {
		t.Fatalf("replay writes: tx=%d lock=%d message=%d update=%d outbox=%d", store.transactionCalls, store.lockCalls, store.appendMessageCalls, store.updateCalls, len(store.enqueued))
	}
}

func TestUpdateFeedbackStatusValidatesCommandBeforeStoreCalls(t *testing.T) {
	store := newFakeStore()
	service := newServiceWithClock(store, &fakeImageStorage{}, time.Now, mustShanghai(t))
	cases := []UpdateStatusCommand{
		{FeedbackID: 0, AdminID: 1, Status: domain.StatusProcessing, Version: 1, RequestID: "req"},
		{FeedbackID: 1, AdminID: 0, Status: domain.StatusProcessing, Version: 1, RequestID: "req"},
		{FeedbackID: 1, AdminID: 1, Status: domain.StatusProcessing, Version: 0, RequestID: "req"},
		{FeedbackID: 1, AdminID: 1, Status: domain.Status("bad"), Version: 1, RequestID: "req"},
		{FeedbackID: 1, AdminID: 1, Status: domain.StatusProcessing, Version: 1, RequestID: " "},
		{FeedbackID: 1, AdminID: 1, Status: domain.StatusProcessing, Version: 1, RequestID: "req", Note: "bad" + string([]byte{0xff})},
		{FeedbackID: 1, AdminID: 1, Status: domain.StatusProcessing, Version: 1, RequestID: "req", Note: strings.Repeat("中", MaxStatusNoteRunes+1)},
	}
	for _, command := range cases {
		if _, err := service.UpdateFeedbackStatus(context.Background(), command); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("UpdateFeedbackStatus(%#v) error = %v, want ErrInvalidArgument", command, err)
		}
	}
	if store.messageReplayCalls != 0 || store.transactionCalls != 0 {
		t.Fatalf("invalid status command called store: replay=%d tx=%d", store.messageReplayCalls, store.transactionCalls)
	}
}

func TestUpdateFeedbackStatusRechecksLockedVersionAndTransition(t *testing.T) {
	tests := []struct {
		name   string
		locked *FeedbackSnapshot
		status domain.Status
		note   string
		want   error
	}{
		{name: "missing", locked: nil, status: domain.StatusProcessing, want: ErrFeedbackNotFound},
		{name: "version", locked: &FeedbackSnapshot{ID: 101, Status: domain.StatusPending, Version: 2}, status: domain.StatusProcessing, want: ErrVersionConflict},
		{name: "invalid transition", locked: &FeedbackSnapshot{ID: 101, Status: domain.StatusPending, Version: 1}, status: domain.StatusResolved, note: "done", want: ErrTransitionNotAllowed},
		{name: "note required", locked: &FeedbackSnapshot{ID: 101, Status: domain.StatusProcessing, Version: 1}, status: domain.StatusResolved, note: " ", want: ErrTransitionNoteRequired},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore()
			store.locked = test.locked
			service := newServiceWithClock(store, &fakeImageStorage{}, time.Now, mustShanghai(t))

			_, err := service.UpdateFeedbackStatus(context.Background(), UpdateStatusCommand{
				FeedbackID: 101, AdminID: 7, Status: test.status, Note: test.note,
				Version: 1, RequestID: "transition",
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("UpdateFeedbackStatus() error = %v, want %v", err, test.want)
			}
			if store.appendMessageCalls != 0 || store.updateCalls != 0 || len(store.enqueued) != 0 {
				t.Fatalf("failed transition writes: message=%d update=%d outbox=%d", store.appendMessageCalls, store.updateCalls, len(store.enqueued))
			}
		})
	}
}

func TestUpdateFeedbackStatusKeepsFirstResolvedAndClosedHistoryAcrossReopen(t *testing.T) {
	now := time.Date(2026, time.September, 5, 10, 0, 0, 0, time.UTC).UnixMilli()
	oldResolved := now - 3000
	oldClosed := now - 2000
	tests := []struct {
		name         string
		from         domain.Status
		to           domain.Status
		resolvedAt   *int64
		closedAt     *int64
		wantResolved int64
		wantClosed   int64
	}{
		{name: "first resolve", from: domain.StatusProcessing, to: domain.StatusResolved, wantResolved: now},
		{name: "first close", from: domain.StatusResolved, to: domain.StatusClosed, resolvedAt: &oldResolved, wantResolved: oldResolved, wantClosed: now},
		{name: "reopen resolved", from: domain.StatusResolved, to: domain.StatusProcessing, resolvedAt: &oldResolved, wantResolved: oldResolved},
		{name: "reopen closed", from: domain.StatusClosed, to: domain.StatusProcessing, resolvedAt: &oldResolved, closedAt: &oldClosed, wantResolved: oldResolved, wantClosed: oldClosed},
		{name: "close again keeps first close", from: domain.StatusResolved, to: domain.StatusClosed, resolvedAt: &oldResolved, closedAt: &oldClosed, wantResolved: oldResolved, wantClosed: oldClosed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore()
			store.locked = &FeedbackSnapshot{
				ID: 101, FeedbackNo: "FB-1", SubmitterID: 9, Status: test.from,
				Version: 1, ResolvedAt: test.resolvedAt, ClosedAt: test.closedAt,
			}
			store.adminDetail = feedbackDetail(101, 9, test.to, 2)
			service := newServiceWithClock(store, &fakeImageStorage{}, func() time.Time { return time.UnixMilli(now) }, mustShanghai(t))

			_, err := service.UpdateFeedbackStatus(context.Background(), UpdateStatusCommand{
				FeedbackID: 101, AdminID: 7, Status: test.to, Note: "handled",
				Version: 1, RequestID: "history-" + test.name,
			})
			if err != nil {
				t.Fatalf("UpdateFeedbackStatus() error = %v", err)
			}
			if got := dereferenceTime(store.updatedSnapshot.ResolvedAt); got != test.wantResolved {
				t.Fatalf("ResolvedAt = %d, want %d", got, test.wantResolved)
			}
			if got := dereferenceTime(store.updatedSnapshot.ClosedAt); got != test.wantClosed {
				t.Fatalf("ClosedAt = %d, want %d", got, test.wantClosed)
			}
		})
	}
}

func TestUpdateFeedbackStatusFailsAtomicallyWhenAnyTransactionalWriteFails(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*fakeStore)
		want      error
	}{
		{name: "message", configure: func(store *fakeStore) { store.appendMessageErr = errors.New("message failed") }, want: errors.New("message failed")},
		{name: "snapshot", configure: func(store *fakeStore) { store.updateErr = ErrVersionConflict }, want: ErrVersionConflict},
		{name: "outbox", configure: func(store *fakeStore) { store.enqueueErr = errors.New("outbox failed") }, want: errors.New("outbox failed")},
		{name: "commit", configure: func(store *fakeStore) { store.transactionCommitErr = errors.New("commit failed") }, want: errors.New("commit failed")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore()
			store.locked = &FeedbackSnapshot{ID: 101, FeedbackNo: "FB-1", SubmitterID: 9, Status: domain.StatusPending, Version: 1}
			test.configure(store)
			service := newServiceWithClock(store, &fakeImageStorage{}, time.Now, mustShanghai(t))

			_, err := service.UpdateFeedbackStatus(context.Background(), UpdateStatusCommand{
				FeedbackID: 101, AdminID: 7, Status: domain.StatusProcessing, Note: "started",
				NotifyUser: true, Version: 1, RequestID: "atomic-" + test.name,
			})
			if test.name == "snapshot" {
				if !errors.Is(err, ErrVersionConflict) {
					t.Fatalf("UpdateFeedbackStatus() error = %v", err)
				}
			} else if err == nil || err.Error() != test.want.Error() {
				t.Fatalf("UpdateFeedbackStatus() error = %v, want %v", err, test.want)
			}
			if store.writeOutsideTransaction {
				t.Fatal("transactional write occurred outside callback")
			}
			if test.name == "message" && (store.updateCalls != 0 || len(store.enqueued) != 0) {
				t.Fatalf("message failure continued: update=%d outbox=%d", store.updateCalls, len(store.enqueued))
			}
			if test.name == "snapshot" && len(store.enqueued) != 0 {
				t.Fatalf("snapshot failure enqueued %#v", store.enqueued)
			}
		})
	}
}

func dereferenceTime(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

type fakeStore struct {
	events                  *[]string
	inTransaction           bool
	writeOutsideTransaction bool
	transactionCalls        int
	transactionCommitErr    error
	createReplays           []*FeedbackDetail
	messageReplays          []*FeedbackDetail
	createReplayCalls       int
	messageReplayCalls      int
	nextSequence            uint64
	nextDateKey             string
	createdOnDate           int64
	countUserID             uint
	countStartMs            int64
	countEndMs              int64
	createdRecord           FeedbackRecord
	createCalls             int
	createdID               uint64
	locked                  *FeedbackSnapshot
	lockErr                 error
	lockCalls               int
	appendedMessage         MessageRecord
	appendMessageCalls      int
	appendedMessageID       uint64
	appendMessageErr        error
	appendedAttachments     []AttachmentRecord
	appendAttachmentCalls   int
	appendAttachmentErr     error
	updatedSnapshot         FeedbackSnapshot
	updateExpectedVersion   uint64
	updateCalls             int
	updateErr               error
	enqueued                []NotificationOutboxRecord
	enqueueErr              error
	userOverview            Overview
	userList                FeedbackList
	userDetail              *FeedbackDetail
	adminOverview           Overview
	adminList               FeedbackList
	adminDetail             *FeedbackDetail
	queryErr                error
	lastUserID              uint
	lastFeedbackID          uint64
	lastUserQuery           UserListQuery
	lastAdminQuery          AdminListQuery
	getUserOverviewCalls    int
	listUserCalls           int
	getUserCalls            int
	getAdminOverviewCalls   int
	listAdminCalls          int
	getAdminCalls           int
}

func newFakeStore() *fakeStore {
	return &fakeStore{nextSequence: 1, createdID: 101, appendedMessageID: 201}
}

func (store *fakeStore) InTransaction(ctx context.Context, fn func(TransactionStore) error) error {
	store.transactionCalls++
	store.recordEvent("transaction")
	store.inTransaction = true
	err := fn(store)
	store.inTransaction = false
	if err != nil {
		return err
	}
	return store.transactionCommitErr
}

func (store *fakeStore) FindCreateReplay(_ context.Context, _ CreateReplayKey) (*FeedbackDetail, bool, error) {
	index := store.createReplayCalls
	store.createReplayCalls++
	if index >= len(store.createReplays) || store.createReplays[index] == nil {
		return nil, false, nil
	}
	return store.createReplays[index], true, nil
}

func (store *fakeStore) FindMessageReplay(_ context.Context, _ MessageReplayKey) (*FeedbackDetail, bool, error) {
	store.recordEvent("find-message-replay")
	index := store.messageReplayCalls
	store.messageReplayCalls++
	if index >= len(store.messageReplays) || store.messageReplays[index] == nil {
		return nil, false, nil
	}
	return store.messageReplays[index], true, nil
}

func (store *fakeStore) NextFeedbackNumber(_ context.Context, dateKey string) (uint64, error) {
	store.nextDateKey = dateKey
	return store.nextSequence, nil
}

func (store *fakeStore) CountCreatedByUserOnDate(_ context.Context, userID uint, startMs, endMs int64) (int64, error) {
	store.countUserID, store.countStartMs, store.countEndMs = userID, startMs, endMs
	return store.createdOnDate, nil
}

func (store *fakeStore) CreateFeedback(_ context.Context, record FeedbackRecord) (uint64, error) {
	store.recordTransactionalWrite()
	store.createCalls++
	store.createdRecord = record
	return store.createdID, nil
}

func (store *fakeStore) LockFeedback(_ context.Context, _ uint64) (*FeedbackSnapshot, error) {
	store.lockCalls++
	store.recordEvent("lock-feedback")
	return store.locked, store.lockErr
}

func (store *fakeStore) AppendMessage(_ context.Context, record MessageRecord) (uint64, error) {
	store.recordTransactionalWrite()
	store.appendMessageCalls++
	store.recordEvent("append-message")
	store.appendedMessage = record
	return store.appendedMessageID, store.appendMessageErr
}

func (store *fakeStore) AppendAttachments(_ context.Context, records []AttachmentRecord) error {
	store.recordTransactionalWrite()
	store.appendAttachmentCalls++
	store.recordEvent("append-attachments")
	store.appendedAttachments = append([]AttachmentRecord(nil), records...)
	return store.appendAttachmentErr
}

func (store *fakeStore) UpdateSnapshot(_ context.Context, snapshot FeedbackSnapshot, expectedVersion uint64) error {
	store.recordTransactionalWrite()
	store.updateCalls++
	store.recordEvent("update-snapshot")
	store.updatedSnapshot = snapshot
	store.updateExpectedVersion = expectedVersion
	return store.updateErr
}

func (store *fakeStore) EnqueueNotification(_ context.Context, record NotificationOutboxRecord) error {
	store.recordTransactionalWrite()
	store.recordEvent("enqueue-notification")
	store.enqueued = append(store.enqueued, record)
	return store.enqueueErr
}

func (store *fakeStore) GetUserOverview(_ context.Context, userID uint) (Overview, error) {
	store.getUserOverviewCalls++
	store.lastUserID = userID
	return store.userOverview, store.queryErr
}

func (store *fakeStore) ListUserFeedbacks(_ context.Context, userID uint, query UserListQuery) (FeedbackList, error) {
	store.listUserCalls++
	store.lastUserID, store.lastUserQuery = userID, query
	return store.userList, store.queryErr
}

func (store *fakeStore) GetUserFeedback(_ context.Context, id uint64, userID uint) (*FeedbackDetail, error) {
	store.getUserCalls++
	store.lastFeedbackID, store.lastUserID = id, userID
	return store.userDetail, store.queryErr
}

func (store *fakeStore) GetAdminOverview(_ context.Context, query AdminListQuery) (Overview, error) {
	store.getAdminOverviewCalls++
	store.lastAdminQuery = query
	return store.adminOverview, store.queryErr
}

func (store *fakeStore) ListAdminFeedbacks(_ context.Context, query AdminListQuery) (FeedbackList, error) {
	store.listAdminCalls++
	store.lastAdminQuery = query
	return store.adminList, store.queryErr
}

func (store *fakeStore) GetAdminFeedback(_ context.Context, id uint64) (*FeedbackDetail, error) {
	store.getAdminCalls++
	store.lastFeedbackID = id
	return store.adminDetail, store.queryErr
}

func (store *fakeStore) recordEvent(event string) {
	if store.events != nil {
		*store.events = append(*store.events, event)
	}
}

func (store *fakeStore) recordTransactionalWrite() {
	if !store.inTransaction {
		store.writeOutsideTransaction = true
	}
}

type fakeImageStorage struct {
	events      *[]string
	saveCalls   int
	deleteCalls int
	saveErrAt   int
	saveErr     error
	saved       []StoredImage
	deleted     []StoredImage
}

func (storage *fakeImageStorage) Save(_ context.Context, image ValidatedImage) (StoredImage, error) {
	storage.saveCalls++
	if storage.events != nil {
		*storage.events = append(*storage.events, "save-image")
	}
	if storage.saveErrAt == storage.saveCalls {
		return StoredImage{}, storage.saveErr
	}
	stored := StoredImage{
		StorageProvider: "fake",
		ObjectKey:       fmt.Sprintf("uploads/feedback/%d-%s", storage.saveCalls, image.OriginalName),
		OriginalName:    image.OriginalName,
		ContentType:     image.ContentType,
		SizeBytes:       image.SizeBytes,
		URL:             fmt.Sprintf("https://static.example/%d", storage.saveCalls),
	}
	storage.saved = append(storage.saved, stored)
	return stored, nil
}

func (storage *fakeImageStorage) Delete(_ context.Context, image StoredImage) error {
	storage.deleteCalls++
	storage.deleted = append(storage.deleted, image)
	return nil
}

func (*fakeImageStorage) PublicURL(_ context.Context, objectKey string) string { return objectKey }

func feedbackDetail(id uint64, submitterID uint, status domain.Status, version uint64) *FeedbackDetail {
	return &FeedbackDetail{FeedbackSummary: FeedbackSummary{ID: id, SubmitterID: submitterID, Status: status, Version: version}}
}

func mustShanghai(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	return location
}
