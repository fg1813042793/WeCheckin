package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

const (
	defaultImageCleanupTimeout         = 10 * time.Second
	defaultReplayReconciliationTimeout = 10 * time.Second
	cleanupErrorMaxRunes               = 1000
	cleanupErrorMaxDepth               = 16
	cleanupErrorMaxNodes               = 64
	cleanupRequestIDMaxRunes           = MaxRequestIDRunes
	cleanupFeedbackNoMaxRunes          = 64
	cleanupObjectKeyMaxRunes           = 600
)

var (
	cleanupSecretAssignmentPattern = regexp.MustCompile(`(?i)(access[_-]?token|app[_-]?secret|secret|token)(\s*[=:]\s*)([^&\s"']+)`)
	cleanupBearerPattern           = regexp.MustCompile(`(?i)(bearer\s+)([a-z0-9._~+/=-]+)`)
)

type imageCleanupReference struct {
	RequestID  string
	FeedbackID uint64
	FeedbackNo string
}

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
	cleanupReference := imageCleanupReference{RequestID: requestID}
	if replay, found, err := service.store.FindCreateReplay(ctx, key); err != nil {
		return nil, err
	} else if found {
		return decorateDetail(replay), nil
	}

	storedImages, err := service.saveImages(ctx, message.Images, cleanupReference)
	if err != nil {
		return nil, err
	}

	now := service.now()
	localNow := now.In(service.location)
	startLocal := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, service.location)
	startMs := startLocal.UTC().UnixMilli()
	endMs := startLocal.AddDate(0, 0, 1).UTC().UnixMilli()
	var feedbackID uint64
	var feedbackNo string
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
		feedbackNo = fmt.Sprintf("FB-%s-%04d", localNow.Format("20060102"), sequence)
		cleanupReference.FeedbackNo = feedbackNo
		feedbackID, err = store.CreateFeedback(ctx, FeedbackRecord{
			FeedbackNo:      feedbackNo,
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
		cleanupReference.FeedbackID = feedbackID
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
		if duplicate, found := service.reconcileReplay(ctx, storedImages, cleanupReference, err, func(reconciliationCtx context.Context) (*FeedbackDetail, bool, error) {
			return service.store.FindCreateReplay(reconciliationCtx, key)
		}); found {
			return decorateDetail(duplicate), nil
		}
		return nil, err
	}
	if replay != nil {
		cleanupReference.FeedbackID = replay.ID
		cleanupReference.FeedbackNo = replay.FeedbackNo
		service.deleteImagesNotReferencedByReplay(ctx, storedImages, cleanupReference, replay)
		return decorateDetail(replay), nil
	}
	return service.GetUserFeedback(ctx, feedbackID, command.SubmitterID)
}

func (service *Service) saveImages(ctx context.Context, images []ValidatedImage, reference imageCleanupReference) ([]StoredImage, error) {
	stored := make([]StoredImage, 0, len(images))
	for _, image := range images {
		item, err := service.imageStorage.Save(ctx, image)
		if err != nil {
			service.deleteImages(ctx, stored, reference)
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return nil, fmt.Errorf("%w: %w", ErrStorageFailed, err)
		}
		stored = append(stored, item)
	}
	return stored, nil
}

func (service *Service) deleteImages(ctx context.Context, images []StoredImage, reference imageCleanupReference) {
	for _, image := range images {
		cleanupCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			timeoutOrDefault(service.imageCleanupTimeout, defaultImageCleanupTimeout),
		)
		if err := service.imageStorage.Delete(cleanupCtx, image); err != nil {
			service.logImageCleanupFailure(reference, image.ObjectKey, err)
		}
		cancel()
	}
}

func (service *Service) reconcileReplay(
	ctx context.Context,
	images []StoredImage,
	reference imageCleanupReference,
	transactionErr error,
	findReplay func(context.Context) (*FeedbackDetail, bool, error),
) (*FeedbackDetail, bool) {
	reconciliationCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		timeoutOrDefault(service.replayReconciliationTimeout, defaultReplayReconciliationTimeout),
	)
	replay, found, err := findReplay(reconciliationCtx)
	cancel()
	if err != nil || !found || replay == nil {
		if errors.Is(transactionErr, ErrTransactionOutcomeUnknown) {
			service.logDeferredImageCleanup(reference, images, transactionErr)
		} else {
			service.deleteImages(ctx, images, reference)
		}
		return nil, false
	}
	reference.FeedbackID = replay.ID
	reference.FeedbackNo = replay.FeedbackNo
	service.deleteImagesNotReferencedByReplay(ctx, images, reference, replay)
	return replay, true
}

func (service *Service) deleteImagesNotReferencedByReplay(
	ctx context.Context,
	images []StoredImage,
	reference imageCleanupReference,
	replay *FeedbackDetail,
) {
	referencedObjectKeys := make(map[string]struct{})
	if replay != nil {
		for _, message := range replay.Messages {
			for _, attachment := range message.Attachments {
				if attachment.ObjectKey != "" {
					referencedObjectKeys[attachment.ObjectKey] = struct{}{}
				}
			}
		}
	}
	unreferencedImages := make([]StoredImage, 0, len(images))
	for _, image := range images {
		if _, referenced := referencedObjectKeys[image.ObjectKey]; !referenced {
			unreferencedImages = append(unreferencedImages, image)
		}
	}
	service.deleteImages(ctx, unreferencedImages, reference)
}

func (service *Service) logImageCleanupFailure(reference imageCleanupReference, objectKey string, err error) {
	service.logImageCleanupEvent("user_feedback_image_cleanup_failed", reference, objectKey, err)
}

func (service *Service) logDeferredImageCleanup(reference imageCleanupReference, images []StoredImage, err error) {
	for _, image := range images {
		service.logImageCleanupEvent("user_feedback_image_cleanup_deferred", reference, image.ObjectKey, err)
	}
}

func (service *Service) logImageCleanupEvent(event string, reference imageCleanupReference, objectKey string, err error) {
	if service == nil || service.logger == nil {
		return
	}
	parts := []string{
		"event=" + event,
		"requestId=" + strconv.Quote(sanitizeCleanupText(reference.RequestID, cleanupRequestIDMaxRunes)),
	}
	if reference.FeedbackID > 0 {
		parts = append(parts, fmt.Sprintf("feedbackId=%d", reference.FeedbackID))
	}
	if reference.FeedbackNo != "" {
		parts = append(parts, "feedbackNo="+strconv.Quote(sanitizeCleanupText(reference.FeedbackNo, cleanupFeedbackNoMaxRunes)))
	}
	parts = append(parts,
		"objectKey="+strconv.Quote(sanitizeCleanupText(objectKey, cleanupObjectKeyMaxRunes)),
		"err="+strconv.Quote(sanitizeCleanupError(err)),
	)
	service.logger.Printf("%s", strings.Join(parts, " "))
}

func sanitizeCleanupError(err error) string {
	if err == nil {
		return "-"
	}
	type errorNode struct {
		err   error
		depth int
	}
	nodes := []errorNode{{err: err}}
	parts := make([]string, 0, 4)
	seen := make(map[string]struct{})
	for visited := 0; len(nodes) > 0 && visited < cleanupErrorMaxNodes; visited++ {
		node := nodes[0]
		nodes = nodes[1:]
		if node.err == nil {
			continue
		}
		children := unwrapCleanupErrors(node.err)
		part := sanitizeCleanupText(cleanupErrorOwnMessage(node.err, children), cleanupErrorMaxRunes)
		part = strings.TrimSpace(part)
		if part != "" {
			if _, exists := seen[part]; !exists {
				seen[part] = struct{}{}
				parts = append(parts, part)
			}
		}
		if node.depth >= cleanupErrorMaxDepth {
			continue
		}
		for _, child := range children {
			if child != nil {
				nodes = append(nodes, errorNode{err: child, depth: node.depth + 1})
			}
		}
	}
	message := sanitizeCleanupText(strings.Join(parts, " | "), cleanupErrorMaxRunes)
	if message == "" {
		return "-"
	}
	return message
}

func unwrapCleanupErrors(err error) []error {
	switch value := err.(type) {
	case interface{ Unwrap() []error }:
		return value.Unwrap()
	case interface{ Unwrap() error }:
		if child := value.Unwrap(); child != nil {
			return []error{child}
		}
	}
	return nil
}

func cleanupErrorOwnMessage(err error, children []error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" || len(children) == 0 {
		return message
	}
	childMessages := make([]string, 0, len(children))
	for _, child := range children {
		if child != nil {
			if childMessage := strings.TrimSpace(child.Error()); childMessage != "" {
				childMessages = append(childMessages, childMessage)
			}
		}
	}
	if len(childMessages) == 1 && message != childMessages[0] && strings.HasSuffix(message, childMessages[0]) {
		message = strings.TrimSpace(strings.TrimSuffix(message, childMessages[0]))
		message = strings.TrimSpace(strings.TrimSuffix(message, ":"))
		return message
	}
	if len(childMessages) > 1 && message == strings.Join(childMessages, "\n") {
		return ""
	}
	return message
}

func sanitizeCleanupText(value string, maxRunes int) string {
	value = cleanupSecretAssignmentPattern.ReplaceAllString(value, "${1}${2}[REDACTED]")
	value = cleanupBearerPattern.ReplaceAllString(value, "${1}[REDACTED]")
	if !utf8.ValidString(value) {
		value = strings.ToValidUTF8(value, "?")
	}
	runes := []rune(value)
	if maxRunes >= 0 && len(runes) > maxRunes {
		runes = runes[:maxRunes]
	}
	return string(runes)
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
