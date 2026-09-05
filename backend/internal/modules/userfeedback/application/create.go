package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func (service *Service) CreateFeedback(ctx context.Context, command CreateCommand) (*FeedbackDetail, error) {
	if err := service.requireWritable(); err != nil {
		return nil, err
	}
	if command.SubmitterID == 0 {
		return nil, ErrInvalidArgument
	}
	requestID, err := normalizeRequestID(command.RequestID)
	if err != nil {
		return nil, err
	}
	message, err := ValidateInitialMessage(command.Content, command.Attachments)
	if err != nil {
		return nil, err
	}
	key := CreateReplayKey{SubmitterID: command.SubmitterID, RequestID: requestID}
	if replay, found, err := service.store.FindCreateReplay(ctx, key); err != nil {
		return nil, err
	} else if found {
		return decorateDetail(replay), nil
	}

	storedImages, err := service.saveImages(ctx, message.Images)
	if err != nil {
		return nil, err
	}

	now := service.now()
	localNow := now.In(service.location)
	startLocal := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, service.location)
	startMs := startLocal.UTC().UnixMilli()
	endMs := startLocal.AddDate(0, 0, 1).UTC().UnixMilli()
	var feedbackID uint64
	var replay *FeedbackDetail
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		var found bool
		replay, found, err = store.FindCreateReplay(ctx, key)
		if err != nil || found {
			return err
		}
		sequence, err := store.NextFeedbackNumber(ctx, localNow.Format("2006-01-02"))
		if err != nil {
			return err
		}
		createdCount, err := store.CountCreatedByUserOnDate(ctx, command.SubmitterID, startMs, endMs)
		if err != nil {
			return err
		}
		if createdCount >= MaxFeedbacksPerUserPerDay {
			return ErrDailyLimitExceeded
		}
		timestamp := now.UTC().UnixMilli()
		feedbackID, err = store.CreateFeedback(ctx, FeedbackRecord{
			FeedbackNo:      fmt.Sprintf("FB-%s-%04d", localNow.Format("20060102"), sequence),
			SubmitterID:     command.SubmitterID,
			CreateRequestID: requestID,
			Status:          domain.StatusPending,
			Version:         1,
			LastActivityAt:  timestamp,
			CreatedAt:       timestamp,
			UpdatedAt:       timestamp,
		})
		if err != nil {
			return err
		}
		messageID, err := store.AppendMessage(ctx, MessageRecord{
			FeedbackID:  feedbackID,
			MessageType: domain.MessageTypeInitial,
			AuthorType:  domain.AuthorTypeUser,
			AuthorID:    command.SubmitterID,
			Content:     message.Content,
			RequestID:   requestID,
			CreatedAt:   timestamp,
		})
		if err != nil {
			return err
		}
		attachments := attachmentRecords(feedbackID, messageID, storedImages, timestamp)
		if len(attachments) > 0 {
			return store.AppendAttachments(ctx, attachments)
		}
		return nil
	})
	if err != nil {
		service.deleteImages(ctx, storedImages)
		if errors.Is(err, ErrDuplicateRequest) {
			if duplicate, found, lookupErr := service.store.FindCreateReplay(ctx, key); lookupErr == nil && found {
				return decorateDetail(duplicate), nil
			}
		}
		return nil, err
	}
	if replay != nil {
		service.deleteImages(ctx, storedImages)
		return decorateDetail(replay), nil
	}
	return service.GetUserFeedback(ctx, feedbackID, command.SubmitterID)
}

func (service *Service) saveImages(ctx context.Context, images []ValidatedImage) ([]StoredImage, error) {
	stored := make([]StoredImage, 0, len(images))
	for _, image := range images {
		item, err := service.imageStorage.Save(ctx, image)
		if err != nil {
			service.deleteImages(ctx, stored)
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return nil, fmt.Errorf("%w: %w", ErrStorageFailed, err)
		}
		stored = append(stored, item)
	}
	return stored, nil
}

func (service *Service) deleteImages(ctx context.Context, images []StoredImage) {
	for _, image := range images {
		_ = service.imageStorage.Delete(ctx, image)
	}
}

func attachmentRecords(feedbackID, messageID uint64, images []StoredImage, createdAt int64) []AttachmentRecord {
	records := make([]AttachmentRecord, 0, len(images))
	for index, image := range images {
		records = append(records, AttachmentRecord{
			FeedbackID:      feedbackID,
			MessageID:       messageID,
			StorageProvider: image.StorageProvider,
			ObjectKey:       image.ObjectKey,
			OriginalName:    image.OriginalName,
			ContentType:     image.ContentType,
			SizeBytes:       image.SizeBytes,
			SortOrder:       index,
			CreatedAt:       createdAt,
		})
	}
	return records
}
