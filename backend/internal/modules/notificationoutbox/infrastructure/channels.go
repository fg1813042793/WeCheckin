package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	notificationmodel "wecheckin/backend/internal/model/notification"
	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/internal/support/outboundhttp"
)

type InAppSender interface {
	Send(context.Context, inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error)
}

type InternalChannel struct {
	sender InAppSender
}

func NewInternalChannel(sender InAppSender) *InternalChannel {
	return &InternalChannel{sender: sender}
}

func (*InternalChannel) Name() string { return notificationmodel.ChannelInternal }

func (channel *InternalChannel) Deliver(ctx context.Context, row notificationmodel.Outbox) error {
	if channel == nil || channel.sender == nil {
		return errors.New("in-app notification sender is not initialized")
	}
	var recipient notificationoutboxapp.InternalRecipient
	if err := json.Unmarshal([]byte(row.RecipientJSON), &recipient); err != nil {
		return fmt.Errorf("decode in-app notification recipient: %w", err)
	}
	payload, err := decodeMessagePayload(row)
	if err != nil {
		return err
	}
	scope := inappnotificationapp.ScopeUsers
	if recipient.NotifyAdmin {
		scope = inappnotificationapp.ScopeAll
	}
	sourceType := strings.TrimSpace(payload.SourceType)
	if sourceType == "" {
		sourceType = row.SourceType
	}
	sourceID := strings.TrimSpace(payload.SourceID)
	if sourceID == "" {
		sourceID = row.SourceID
	}
	_, err = channel.sender.Send(ctx, inappnotificationapp.SendInput{
		Title: payload.Title, Content: payload.Content, Scope: scope, UserIDs: recipient.UserIDs,
		SourceType: sourceType, SourceID: sourceID, DeliveryKey: row.IdempotencyKey,
	})
	return err
}

type OutboundClient interface {
	Do(context.Context, outboundhttp.Request) (outboundhttp.Response, error)
}

type WebhookChannel struct {
	client OutboundClient
}

func NewWebhookChannel(client OutboundClient) *WebhookChannel {
	return &WebhookChannel{client: client}
}

func (*WebhookChannel) Name() string { return notificationmodel.ChannelWebhook }

func (channel *WebhookChannel) Deliver(ctx context.Context, row notificationmodel.Outbox) error {
	if channel == nil || channel.client == nil {
		return errors.New("webhook HTTP client is not initialized")
	}
	var recipient notificationoutboxapp.WebhookRecipient
	if err := json.Unmarshal([]byte(row.RecipientJSON), &recipient); err != nil {
		return fmt.Errorf("decode webhook recipient: %w", err)
	}
	payload, err := decodeMessagePayload(row)
	if err != nil {
		return err
	}
	body, err := webhookPayload(recipient.Type, payload.Title, payload.Content)
	if err != nil {
		return err
	}
	response, err := channel.client.Do(ctx, outboundhttp.Request{
		Method: http.MethodPost, URL: recipient.URL,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    body, Timeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("webhook returned HTTP status %d", response.StatusCode)
	}
	return nil
}

func decodeMessagePayload(row notificationmodel.Outbox) (notificationoutboxapp.MessagePayload, error) {
	var payload notificationoutboxapp.MessagePayload
	if err := json.Unmarshal([]byte(row.PayloadJSON), &payload); err != nil {
		return payload, fmt.Errorf("decode notification payload: %w", err)
	}
	return payload, nil
}

func webhookPayload(webhookType, title, content string) ([]byte, error) {
	content = strings.ReplaceAll(content, "\n", "\n\n")
	var payload any
	switch strings.ToLower(strings.TrimSpace(webhookType)) {
	case "dingtalk":
		payload = map[string]any{
			"msgtype":  "markdown",
			"markdown": map[string]string{"title": title, "text": content},
		}
	case "wecom":
		payload = map[string]any{
			"msgtype":  "markdown",
			"markdown": map[string]string{"content": content},
		}
	case "lark":
		payload = map[string]any{
			"msg_type": "interactive",
			"card": map[string]any{
				"header": map[string]any{
					"title":    map[string]string{"tag": "plain_text", "content": title},
					"template": "blue",
				},
				"elements": []map[string]any{{"tag": "markdown", "content": content}},
			},
		}
	default:
		payload = map[string]any{"title": title, "content": content, "type": "survey_stat"}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode webhook payload: %w", err)
	}
	return body, nil
}

var _ notificationoutboxapp.Channel = (*InternalChannel)(nil)
var _ notificationoutboxapp.Channel = (*WebhookChannel)(nil)
