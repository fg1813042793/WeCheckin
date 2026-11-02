package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	notificationmodel "wecheckin/backend/internal/model/notification"
	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
)

// DingTalkChannel delivers one feedback notification per outbox attempt. Delivery
// is at least once: a crash after the external send but before MarkSent can resend.
type DingTalkChannel struct {
	delivery inappnotificationapp.DingTalkDelivery
}

func NewDingTalkChannel(delivery inappnotificationapp.DingTalkDelivery) *DingTalkChannel {
	return &DingTalkChannel{delivery: delivery}
}

func (*DingTalkChannel) Name() string { return notificationmodel.ChannelDingTalk }

func (channel *DingTalkChannel) Deliver(ctx context.Context, row notificationmodel.Outbox) (notificationoutboxapp.DeliveryResult, error) {
	if channel == nil || channel.delivery == nil {
		return notificationoutboxapp.DeliveryResult{}, inappnotificationapp.ErrDingTalkDeliveryUnavailable
	}
	var recipient notificationoutboxapp.InternalRecipient
	decoder := json.NewDecoder(strings.NewReader(row.RecipientJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&recipient); err != nil {
		return notificationoutboxapp.DeliveryResult{}, fmt.Errorf("decode DingTalk notification recipient: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return notificationoutboxapp.DeliveryResult{}, errors.New("DingTalk notification recipient must be a single JSON object")
	}
	if recipient.NotifyAdmin || len(recipient.UserIDs) != 1 || recipient.UserIDs[0] == 0 {
		return notificationoutboxapp.DeliveryResult{}, errors.New("DingTalk feedback notification requires exactly one nonzero user ID")
	}
	payload, err := decodeMessagePayload(row)
	if err != nil {
		return notificationoutboxapp.DeliveryResult{}, err
	}
	sourceID := strings.TrimSpace(row.SourceID)
	feedbackID, sourceErr := strconv.ParseUint(sourceID, 10, 64)
	if row.Channel != notificationmodel.ChannelDingTalk || strings.TrimSpace(row.SourceType) != "user_feedback" ||
		strings.TrimSpace(payload.SourceType) != "user_feedback" || sourceErr != nil || feedbackID == 0 ||
		strings.TrimSpace(payload.SourceID) != sourceID || strings.TrimSpace(payload.NotificationType) != "feedback_status" {
		return notificationoutboxapp.DeliveryResult{}, errors.New("invalid DingTalk feedback notification source")
	}
	if strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Content) == "" {
		return notificationoutboxapp.DeliveryResult{}, errors.New("DingTalk feedback notification title and content are required")
	}
	// Call the delivery boundary directly; SendDingTalk would create another record.
	delivered, deliveryErr := channel.delivery.DeliverDingTalk(ctx, inappnotificationapp.DingTalkDeliveryBatch{
		Title: payload.Title, Content: payload.Content, NotificationType: strings.TrimSpace(payload.NotificationType),
		UserIDs: recipient.UserIDs,
	})
	counts := struct {
		PlannedCount int `json:"plannedCount"`
		SentCount    int `json:"sentCount"`
		SkippedCount int `json:"skippedCount"`
		FailedCount  int `json:"failedCount"`
	}{1, delivered.SentCount, delivered.SkippedCount, delivered.FailedCount}
	body, err := json.Marshal(counts)
	if err != nil {
		return notificationoutboxapp.DeliveryResult{}, fmt.Errorf("encode DingTalk delivery counts: %w", err)
	}
	result := notificationoutboxapp.DeliveryResult{ResponseBody: string(body)}
	if deliveryErr != nil {
		return result, deliveryErr
	}
	if delivered.SentCount != 1 || delivered.SkippedCount != 0 || delivered.FailedCount != 0 {
		return result, fmt.Errorf("DingTalk feedback delivery incomplete: planned=1 sent=%d skipped=%d failed=%d", delivered.SentCount, delivered.SkippedCount, delivered.FailedCount)
	}
	return result, nil
}

var _ notificationoutboxapp.Channel = (*DingTalkChannel)(nil)
