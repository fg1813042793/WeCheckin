package notification

import (
	"context"
	"errors"
	"testing"

	"wecheckin/backend/internal/model"
)

type fakeRepository struct {
	listUserID      string
	listPage        int
	listPageSize    int
	listItems       []model.Notify
	listTotal       int64
	unreadUserID    string
	unreadCount     int64
	markReadUserID  string
	markReadID      uint
	markReadUpdated bool
	markAllUserID   string
}

func (r *fakeRepository) List(_ context.Context, userID string, page, pageSize int) ([]model.Notify, int64, error) {
	r.listUserID = userID
	r.listPage = page
	r.listPageSize = pageSize
	return r.listItems, r.listTotal, nil
}

func (r *fakeRepository) UnreadCount(_ context.Context, userID string) (int64, error) {
	r.unreadUserID = userID
	return r.unreadCount, nil
}

func (r *fakeRepository) MarkRead(_ context.Context, userID string, id uint) (bool, error) {
	r.markReadUserID = userID
	r.markReadID = id
	return r.markReadUpdated, nil
}

func (r *fakeRepository) MarkAllRead(_ context.Context, userID string) error {
	r.markAllUserID = userID
	return nil
}

func TestServiceListScopesToCurrentUserAndNormalizesPagination(t *testing.T) {
	repository := &fakeRepository{
		listItems: []model.Notify{{
			ID:         7,
			Title:      "待办提醒",
			Content:    "你有一项待处理任务",
			Type:       "workflow",
			SourceType: "workflow_instance",
			SourceID:   "instance-1",
			UserID:     "42",
			IsRead:     0,
			AddTime:    1788422400000,
		}},
		listTotal: 1,
	}
	service := NewServiceWithRepository(repository)

	result, err := service.List(context.Background(), 42, 0, 101)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.listUserID != "42" {
		t.Fatalf("repository user id = %q, want 42", repository.listUserID)
	}
	if repository.listPage != 1 || repository.listPageSize != 20 {
		t.Fatalf("pagination = (%d, %d), want (1, 20)", repository.listPage, repository.listPageSize)
	}
	if result.Total != 1 || len(result.List) != 1 {
		t.Fatalf("result = %#v, want one notification", result)
	}
	item := result.List[0]
	if item.ID != 7 || item.SourceType != "workflow_instance" || item.SourceID != "instance-1" {
		t.Fatalf("notification dto = %#v", item)
	}
}

func TestServiceUnreadCountScopesToCurrentUser(t *testing.T) {
	repository := &fakeRepository{unreadCount: 3}
	service := NewServiceWithRepository(repository)

	count, err := service.UnreadCount(context.Background(), 66)
	if err != nil {
		t.Fatalf("UnreadCount() error = %v", err)
	}
	if count != 3 || repository.unreadUserID != "66" {
		t.Fatalf("count/user = %d/%q, want 3/66", count, repository.unreadUserID)
	}
}

func TestServiceMarkReadRejectsNotificationOutsideCurrentUser(t *testing.T) {
	repository := &fakeRepository{markReadUpdated: false}
	service := NewServiceWithRepository(repository)

	err := service.MarkRead(context.Background(), 66, 9)
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf("MarkRead() error = %v, want ErrNotificationNotFound", err)
	}
	if repository.markReadUserID != "66" || repository.markReadID != 9 {
		t.Fatalf("mark read scope = %q/%d, want 66/9", repository.markReadUserID, repository.markReadID)
	}
}

func TestServiceMarkAllReadScopesToCurrentUser(t *testing.T) {
	repository := &fakeRepository{}
	service := NewServiceWithRepository(repository)

	if err := service.MarkAllRead(context.Background(), 88); err != nil {
		t.Fatalf("MarkAllRead() error = %v", err)
	}
	if repository.markAllUserID != "88" {
		t.Fatalf("mark all user id = %q, want 88", repository.markAllUserID)
	}
}

func TestServiceRejectsMissingCurrentUser(t *testing.T) {
	service := NewServiceWithRepository(&fakeRepository{})

	if _, err := service.List(context.Background(), 0, 1, 20); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("List() error = %v, want ErrUnauthenticated", err)
	}
	if err := service.MarkRead(context.Background(), 0, 1); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("MarkRead() error = %v, want ErrUnauthenticated", err)
	}
}
