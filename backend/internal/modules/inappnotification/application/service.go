package application

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Service struct {
	store            Store
	dingTalkDelivery DingTalkDelivery
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func NewServiceWithDingTalk(store Store, delivery DingTalkDelivery) *Service {
	return &Service{store: store, dingTalkDelivery: delivery}
}

func (service *Service) List(ctx context.Context, currentUserID uint, page, pageSize int) (NotificationList, error) {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return NotificationList{}, err
	}
	page, pageSize = normalizePagination(page, pageSize)
	items, total, err := service.store.List(ctx, userID, page, pageSize)
	if err != nil {
		return NotificationList{}, err
	}
	return NotificationList{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (service *Service) UnreadCount(ctx context.Context, currentUserID uint) (int64, error) {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return 0, err
	}
	return service.store.UnreadCount(ctx, userID)
}

func (service *Service) MarkRead(ctx context.Context, currentUserID, notificationID uint) error {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return err
	}
	if notificationID == 0 {
		return ErrNotificationMissing
	}
	found, err := service.store.MarkRead(ctx, userID, notificationID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotificationMissing
	}
	return nil
}

func (service *Service) MarkAllRead(ctx context.Context, currentUserID uint) error {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return err
	}
	return service.store.MarkAllRead(ctx, userID)
}

func (service *Service) RecipientOptions(ctx context.Context) (RecipientOptions, error) {
	if service == nil || service.store == nil {
		return RecipientOptions{}, fmt.Errorf("in-app notification store is not initialized")
	}
	return service.store.RecipientOptions(ctx)
}

func ValidateSendInput(input SendInput) error {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return ErrTitleRequired
	}
	if utf8.RuneCountInString(title) > 255 {
		return ErrTitleTooLong
	}
	if strings.TrimSpace(input.Content) == "" {
		return ErrContentRequired
	}
	if utf8.RuneCountInString(input.Content) > 5000 {
		return ErrContentTooLong
	}
	switch input.Scope {
	case ScopeAll:
	case ScopeUsers:
		if len(normalizeIDs(input.UserIDs)) == 0 {
			return ErrRecipientsRequired
		}
	case ScopeDepartments:
		if len(normalizeIDs(input.DepartmentIDs)) == 0 {
			return ErrRecipientsRequired
		}
	default:
		return ErrInvalidScope
	}
	if strings.TrimSpace(input.SourceType) == "" || strings.TrimSpace(input.SourceID) == "" {
		return ErrSourceRequired
	}
	if len(strings.TrimSpace(input.SourceID)) > 64 {
		return ErrSourceTooLong
	}
	return nil
}

func (service *Service) Send(ctx context.Context, input SendInput) (SendResult, error) {
	if service == nil || service.store == nil {
		return SendResult{}, fmt.Errorf("in-app notification store is not initialized")
	}
	if err := ValidateSendInput(input); err != nil {
		return SendResult{}, err
	}
	input.Title = strings.TrimSpace(input.Title)
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceID = strings.TrimSpace(input.SourceID)
	rule := RecipientRule{
		Scope:         input.Scope,
		UserIDs:       normalizeIDs(input.UserIDs),
		DepartmentIDs: normalizeIDs(input.DepartmentIDs),
	}
	resolution, err := service.store.ResolveRecipients(ctx, rule)
	if err != nil {
		return SendResult{}, err
	}
	resolution.UserIDs = normalizeIDs(resolution.UserIDs)
	if len(resolution.UserIDs) == 0 {
		return SendResult{}, ErrNoRecipients
	}
	delivery, err := service.store.Deliver(ctx, DeliveryBatch{
		Title:      input.Title,
		Content:    input.Content,
		Type:       notificationType(input.SourceType),
		SourceType: input.SourceType,
		SourceID:   input.SourceID,
		UserIDs:    resolution.UserIDs,
	})
	if err != nil {
		return SendResult{}, err
	}
	return SendResult{
		SourceID:     input.SourceID,
		PlannedCount: len(resolution.UserIDs),
		SentCount:    delivery.SentCount,
		SkippedCount: resolution.SkippedCount,
		Replayed:     delivery.Replayed,
	}, nil
}

func (service *Service) SendDingTalk(ctx context.Context, input SendInput) (SendResult, error) {
	if service == nil || service.store == nil {
		return SendResult{}, fmt.Errorf("notification recipient store is not initialized")
	}
	if service.dingTalkDelivery == nil {
		return SendResult{}, ErrDingTalkDeliveryUnavailable
	}
	if err := ValidateSendInput(input); err != nil {
		return SendResult{}, err
	}
	input.Title = strings.TrimSpace(input.Title)
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceID = strings.TrimSpace(input.SourceID)
	resolution, err := service.store.ResolveRecipients(ctx, RecipientRule{
		Scope:         input.Scope,
		UserIDs:       normalizeIDs(input.UserIDs),
		DepartmentIDs: normalizeIDs(input.DepartmentIDs),
	})
	if err != nil {
		return SendResult{}, err
	}
	resolution.UserIDs = normalizeIDs(resolution.UserIDs)
	if len(resolution.UserIDs) == 0 {
		return SendResult{}, ErrNoRecipients
	}
	delivery, err := service.dingTalkDelivery.DeliverDingTalk(ctx, DingTalkDeliveryBatch{
		Title: input.Title, Content: input.Content, UserIDs: resolution.UserIDs,
	})
	if err != nil {
		return SendResult{}, err
	}
	return SendResult{
		SourceID:     input.SourceID,
		PlannedCount: len(resolution.UserIDs),
		SentCount:    delivery.SentCount,
		SkippedCount: resolution.SkippedCount + delivery.SkippedCount,
		FailedCount:  delivery.FailedCount,
	}, nil
}

func normalizeIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func notificationType(sourceType string) string {
	if sourceType == SourceScheduledTaskRun {
		return TypeScheduledTask
	}
	return TypeAdminManual
}

func requiredUserID(value uint) (string, error) {
	if value == 0 {
		return "", ErrUnauthenticated
	}
	return strconv.FormatUint(uint64(value), 10), nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
