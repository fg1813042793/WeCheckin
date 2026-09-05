package application

import (
	"context"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func (service *Service) SupplementFeedback(ctx context.Context, command SupplementCommand) (*FeedbackDetail, error) {
	if err := service.requireWritable(); err != nil {
		return nil, err
	}
	if command.FeedbackID == 0 || command.SubmitterID == 0 || command.Version == 0 {
		return nil, ErrInvalidArgument
	}
	requestID, err := normalizeRequestID(command.RequestID)
	if err != nil {
		return nil, err
	}
	message, err := ValidateSupplementMessage(command.Content, command.Attachments, 0)
	if err != nil {
		return nil, err
	}
	key := MessageReplayKey{
		FeedbackID: command.FeedbackID,
		AuthorType: domain.AuthorTypeUser,
		AuthorID:   command.SubmitterID,
		RequestID:  requestID,
	}
	cleanupReference := imageCleanupReference{RequestID: requestID, FeedbackID: command.FeedbackID}
	if replay, found, err := service.store.FindMessageReplay(ctx, key); err != nil {
		return nil, err
	} else if found {
		return decorateDetail(replay), nil
	}

	storedImages, err := service.saveImages(ctx, message.Images, cleanupReference)
	if err != nil {
		return nil, err
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
		if locked == nil || locked.ID != command.FeedbackID || locked.SubmitterID != command.SubmitterID {
			return ErrFeedbackNotFound
		}
		cleanupReference.FeedbackNo = locked.FeedbackNo
		if !locked.Status.AllowsSupplement() {
			return ErrSupplementNotAllowed
		}
		if locked.Version != command.Version {
			return ErrVersionConflict
		}
		if locked.AttachmentCount < 0 {
			return ErrInvalidArgument
		}
		if int64(len(storedImages)) > MaxImagesPerFeedback-locked.AttachmentCount {
			return ErrAttachmentLimitExceeded
		}
		messageID, err := store.AppendMessage(ctx, MessageRecord{
			FeedbackID:  command.FeedbackID,
			MessageType: domain.MessageTypeSupplement,
			AuthorType:  domain.AuthorTypeUser,
			AuthorID:    command.SubmitterID,
			Content:     message.Content,
			RequestID:   requestID,
			CreatedAt:   now,
		})
		if err != nil {
			return err
		}
		attachments := attachmentRecords(command.FeedbackID, messageID, storedImages, now)
		if len(attachments) > 0 {
			if err := store.AppendAttachments(ctx, attachments); err != nil {
				return err
			}
		}
		updated := *locked
		updated.Version = locked.Version + 1
		updated.AttachmentCount = locked.AttachmentCount + int64(len(storedImages))
		updated.LastActivityAt = now
		updated.UpdatedAt = now
		return store.UpdateSnapshot(ctx, updated, locked.Version)
	})
	if err != nil {
		if duplicate, found := service.reconcileReplay(ctx, storedImages, cleanupReference, err, func(reconciliationCtx context.Context) (*FeedbackDetail, bool, error) {
			return service.store.FindMessageReplay(reconciliationCtx, key)
		}); found {
			return decorateDetail(duplicate), nil
		}
		return nil, err
	}
	if concurrentReplay {
		if replay != nil {
			cleanupReference.FeedbackNo = replay.FeedbackNo
		}
		service.deleteImagesNotReferencedByReplay(ctx, storedImages, cleanupReference, replay)
		return decorateDetail(replay), nil
	}
	return service.GetUserFeedback(ctx, command.FeedbackID, command.SubmitterID)
}
