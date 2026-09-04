package poststat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	notificationmodel "wecheckin/backend/internal/model/notification"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
)

var ErrNotificationDispatcherUnavailable = errors.New("post-submit notification dispatcher is not initialized")

type NotificationRequest struct {
	IdempotencyKey string
	SourceType     string
	SourceID       string
	Channel        string
	Recipient      any
	Payload        any
}

type NotificationDispatcher interface {
	Enqueue(context.Context, NotificationRequest) error
}

type OutboxEnqueuer interface {
	Enqueue(context.Context, notificationoutboxapp.EnqueueRequest) error
}

type outboxDispatcher struct {
	enqueuer OutboxEnqueuer
}

func NewOutboxDispatcher(enqueuer OutboxEnqueuer) NotificationDispatcher {
	return &outboxDispatcher{enqueuer: enqueuer}
}

func (dispatcher *outboxDispatcher) Enqueue(ctx context.Context, request NotificationRequest) error {
	if dispatcher == nil || dispatcher.enqueuer == nil {
		return ErrNotificationDispatcherUnavailable
	}
	return dispatcher.enqueuer.Enqueue(ctx, notificationoutboxapp.EnqueueRequest{
		IdempotencyKey: request.IdempotencyKey,
		SourceType:     request.SourceType,
		SourceID:       request.SourceID,
		Channel:        request.Channel,
		Recipient:      request.Recipient,
		Payload:        request.Payload,
	})
}

var dispatcherRegistry struct {
	sync.RWMutex
	dispatcher NotificationDispatcher
}

func ConfigureNotificationDispatcher(dispatcher NotificationDispatcher) {
	dispatcherRegistry.Lock()
	dispatcherRegistry.dispatcher = dispatcher
	dispatcherRegistry.Unlock()
}

func configuredNotificationDispatcher() NotificationDispatcher {
	dispatcherRegistry.RLock()
	defer dispatcherRegistry.RUnlock()
	return dispatcherRegistry.dispatcher
}

type ruleNotificationInput struct {
	SurveyID    uint
	ResponseID  uint
	RuleIndex   int
	SurveyTitle string
	Message     string
	Rule        Rule
}

func dispatchRuleNotifications(ctx context.Context, dispatcher NotificationDispatcher, input ruleNotificationInput) error {
	requests := notificationRequests(input)
	if len(requests) == 0 {
		return nil
	}
	if dispatcher == nil {
		return ErrNotificationDispatcherUnavailable
	}
	for _, request := range requests {
		if err := dispatcher.Enqueue(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

func notificationRequests(input ruleNotificationInput) []NotificationRequest {
	sourceID := strconv.FormatUint(uint64(input.ResponseID), 10)
	payload := notificationoutboxapp.MessagePayload{
		Title:   "问卷统计通知: " + input.SurveyTitle,
		Content: input.Message, SourceType: "survey", SourceID: strconv.FormatUint(uint64(input.SurveyID), 10),
	}
	requests := make([]NotificationRequest, 0, 2)
	if (input.Rule.NotifyChannel == "webhook" || input.Rule.NotifyChannel == "both") && strings.TrimSpace(input.Rule.WebhookURL) != "" {
		requests = append(requests, NotificationRequest{
			IdempotencyKey: postStatIdempotencyKey(input, notificationmodel.ChannelWebhook),
			SourceType:     "survey_response", SourceID: sourceID, Channel: notificationmodel.ChannelWebhook,
			Recipient: notificationoutboxapp.WebhookRecipient{Type: input.Rule.WebhookType, URL: strings.TrimSpace(input.Rule.WebhookURL)},
			Payload: notificationoutboxapp.MessagePayload{
				Title: input.SurveyTitle, Content: input.Message,
			},
		})
	}
	userIDs := parseNotificationUserIDs(input.Rule.NotifyUserIds)
	if (input.Rule.NotifyChannel == "internal" || input.Rule.NotifyChannel == "both") && (input.Rule.NotifyAdmin || len(userIDs) > 0) {
		requests = append(requests, NotificationRequest{
			IdempotencyKey: postStatIdempotencyKey(input, notificationmodel.ChannelInternal),
			SourceType:     "survey_response", SourceID: sourceID, Channel: notificationmodel.ChannelInternal,
			Recipient: notificationoutboxapp.InternalRecipient{NotifyAdmin: input.Rule.NotifyAdmin, UserIDs: userIDs},
			Payload:   payload,
		})
	}
	return requests
}

func postStatIdempotencyKey(input ruleNotificationInput, channel string) string {
	ruleID := strings.TrimSpace(input.Rule.ID)
	if ruleID == "" {
		ruleID = fmt.Sprintf("index-%d", input.RuleIndex)
	}
	raw := fmt.Sprintf("survey:%d:answer:%d:rule:%s:channel:%s", input.SurveyID, input.ResponseID, ruleID, channel)
	sum := sha256.Sum256([]byte(raw))
	return "poststat:" + hex.EncodeToString(sum[:])
}

func parseNotificationUserIDs(raw string) []uint {
	seen := make(map[uint]struct{})
	result := make([]uint, 0)
	for _, part := range strings.Split(raw, ",") {
		value, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
		if err != nil || value == 0 {
			continue
		}
		id := uint(value)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
