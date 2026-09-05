package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

const (
	NotificationChannelInternal    = "internal"
	NotificationTypeFeedbackStatus = "feedback_status"
	NotificationSourceUserFeedback = "user_feedback"
)

func (service *Service) UpdateFeedbackStatus(ctx context.Context, command UpdateStatusCommand) (*FeedbackDetail, error) {
	if err := service.requireClock(); err != nil {
		return nil, err
	}
	if command.FeedbackID == 0 || command.AdminID == 0 || command.Version == 0 || !command.Status.Valid() {
		return nil, ErrInvalidArgument
	}
	requestID, err := normalizeRequestID(command.RequestID)
	if err != nil {
		return nil, err
	}
	note, err := normalizeStatusNote(command.Note)
	if err != nil {
		return nil, err
	}
	key := MessageReplayKey{
		FeedbackID: command.FeedbackID,
		AuthorType: domain.AuthorTypeAdmin,
		AuthorID:   command.AdminID,
		RequestID:  requestID,
	}
	if replay, found, err := service.store.FindMessageReplay(ctx, key); err != nil {
		return nil, err
	} else if found {
		return decorateDetail(replay), nil
	}

	now := service.now().UTC().UnixMilli()
	var replay *FeedbackDetail
	concurrentReplay := false
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		var found bool
		replay, found, err = store.FindMessageReplay(ctx, key)
		if err != nil {
			return err
		}
		if found {
			concurrentReplay = true
			return nil
		}
		locked, err := store.LockFeedback(ctx, command.FeedbackID)
		if err != nil {
			return err
		}
		if locked == nil || locked.ID != command.FeedbackID {
			return ErrFeedbackNotFound
		}
		if locked.Version != command.Version {
			return ErrVersionConflict
		}
		if err := domain.ValidateTransition(locked.Status, command.Status, note); err != nil {
			return err
		}
		messageID, err := store.AppendMessage(ctx, MessageRecord{
			FeedbackID:  command.FeedbackID,
			MessageType: domain.MessageTypeStatus,
			AuthorType:  domain.AuthorTypeAdmin,
			AuthorID:    command.AdminID,
			Content:     note,
			FromStatus:  locked.Status,
			ToStatus:    command.Status,
			RequestID:   requestID,
			CreatedAt:   now,
		})
		if err != nil {
			return err
		}
		updated := *locked
		handlerID := command.AdminID
		updated.HandlerID = &handlerID
		updated.Status = command.Status
		updated.Version = locked.Version + 1
		updated.LastActivityAt = now
		updated.UpdatedAt = now
		if command.Status == domain.StatusResolved && updated.ResolvedAt == nil {
			resolvedAt := now
			updated.ResolvedAt = &resolvedAt
		}
		if command.Status == domain.StatusClosed && updated.ClosedAt == nil {
			closedAt := now
			updated.ClosedAt = &closedAt
		}
		if err := store.UpdateSnapshot(ctx, updated, locked.Version); err != nil {
			return err
		}
		if !command.NotifyUser {
			return nil
		}
		return store.EnqueueNotification(ctx, NotificationOutboxRecord{
			IdempotencyKey:   fmt.Sprintf("user-feedback-status:%d:%d", command.FeedbackID, messageID),
			Channel:          NotificationChannelInternal,
			NotificationType: NotificationTypeFeedbackStatus,
			SourceType:       NotificationSourceUserFeedback,
			SourceID:         strconv.FormatUint(command.FeedbackID, 10),
			RecipientUserID:  locked.SubmitterID,
			Title:            fmt.Sprintf("反馈 %s %s", locked.FeedbackNo, feedbackStatusName(command.Status)),
			Content:          note,
			CreatedAt:        now,
		})
	})
	if err != nil {
		if ctx.Err() == nil {
			if duplicate, found, lookupErr := service.store.FindMessageReplay(ctx, key); lookupErr == nil && found {
				return decorateDetail(duplicate), nil
			}
		}
		return nil, err
	}
	if concurrentReplay {
		return decorateDetail(replay), nil
	}
	return service.GetAdminFeedback(ctx, command.FeedbackID)
}

func normalizeStatusNote(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", ErrInvalidArgument
	}
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > MaxStatusNoteRunes {
		return "", ErrInvalidArgument
	}
	return value, nil
}

func feedbackStatusName(status domain.Status) string {
	switch status {
	case domain.StatusPending:
		return "待处理"
	case domain.StatusProcessing:
		return "处理中"
	case domain.StatusResolved:
		return "已解决"
	case domain.StatusClosed:
		return "已关闭"
	default:
		return string(status)
	}
}
