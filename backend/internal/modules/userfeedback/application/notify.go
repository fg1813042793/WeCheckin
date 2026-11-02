package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

const NotificationChannelDingTalk = "dingtalk"

var ErrDingTalkRecipientUnavailable = errors.New("feedback submitter has no available DingTalk notification binding")

// NotifySubmitter records a message and queues delivery without changing the
// handling status or assignee. Timeline and notification commit together.
func (service *Service) NotifySubmitter(ctx context.Context, command NotifySubmitterCommand) (*FeedbackDetail, error) {
	if err := service.requireClock(); err != nil {
		return nil, err
	}
	if command.FeedbackID == 0 || command.AdminID == 0 || command.Version == 0 {
		return nil, ErrInvalidArgument
	}
	channel := strings.TrimSpace(command.Channel)
	if channel == "" {
		channel = NotificationChannelInternal
	}
	if channel != NotificationChannelInternal && channel != NotificationChannelDingTalk {
		return nil, ErrInvalidArgument
	}
	requestID, err := normalizeRequestID(command.RequestID)
	if err != nil {
		return nil, err
	}
	content, err := normalizeStatusNote(command.Content)
	if err != nil || content == "" {
		return nil, ErrInvalidArgument
	}
	key := MessageReplayKey{FeedbackID: command.FeedbackID, AuthorType: domain.AuthorTypeAdmin, AuthorID: command.AdminID, RequestID: requestID}
	if replay, found, err := service.store.FindMessageReplay(ctx, key); err != nil {
		return nil, err
	} else if found {
		return decorateDetail(replay), nil
	}
	// Resolve readiness before acquiring the feedback lock. The checker may use
	// another database connection; never hold a transaction while waiting for it.
	var checkedSubmitterID uint
	if channel == NotificationChannelDingTalk {
		if service.dingTalkRecipientChecker == nil {
			return nil, ErrDingTalkRecipientUnavailable
		}
		detail, err := service.GetAdminFeedback(ctx, command.FeedbackID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, ErrFeedbackNotFound
		}
		checkedSubmitterID = detail.SubmitterID
		available, err := service.dingTalkRecipientChecker.CanDeliverTo(ctx, checkedSubmitterID)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, ErrDingTalkRecipientUnavailable
		}
	}

	now := service.now().UTC().UnixMilli()
	var replay *FeedbackDetail
	var concurrentReplay bool
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		var err error
		replay, concurrentReplay, err = store.FindMessageReplay(ctx, key)
		if err != nil || concurrentReplay {
			return err
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
		if channel == NotificationChannelDingTalk && locked.SubmitterID != checkedSubmitterID {
			return ErrVersionConflict
		}
		messageID, err := store.AppendMessage(ctx, MessageRecord{
			FeedbackID: command.FeedbackID, MessageType: domain.MessageTypeNotification,
			AuthorType: domain.AuthorTypeAdmin, AuthorID: command.AdminID,
			Content: content, RequestID: requestID, CreatedAt: now,
		})
		if err != nil {
			return err
		}
		updated := *locked
		updated.Version++
		updated.LastActivityAt, updated.UpdatedAt = now, now
		if err := store.UpdateSnapshot(ctx, updated, locked.Version); err != nil {
			return err
		}
		return store.EnqueueNotification(ctx, NotificationOutboxRecord{
			IdempotencyKey: fmt.Sprintf("user-feedback-notify:%d:%d", command.FeedbackID, messageID),
			Channel:        channel, NotificationType: NotificationTypeFeedbackStatus,
			SourceType: NotificationSourceUserFeedback, SourceID: strconv.FormatUint(command.FeedbackID, 10),
			RecipientUserID: locked.SubmitterID, Title: fmt.Sprintf("反馈 %s 新消息", locked.FeedbackNo),
			Content: content, CreatedAt: now,
		})
	})
	if err != nil {
		// A lost response or uncertain commit must not send another notification.
		reconciliationCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeoutOrDefault(service.replayReconciliationTimeout, defaultReplayReconciliationTimeout))
		duplicate, found, lookupErr := service.store.FindMessageReplay(reconciliationCtx, key)
		cancel()
		if lookupErr == nil && found {
			return decorateDetail(duplicate), nil
		}
		return nil, err
	}
	if concurrentReplay {
		return decorateDetail(replay), nil
	}
	return service.GetAdminFeedback(ctx, command.FeedbackID)
}
