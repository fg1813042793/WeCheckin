package notificationmodel

import "time"

const (
	StatusPending = "pending"
	StatusSending = "sending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusDead    = "dead"

	ChannelInternal = "internal"
	ChannelWebhook  = "webhook"
)

type Outbox struct {
	ID             uint64    `json:"id" gorm:"primaryKey;column:id"`
	IdempotencyKey string    `json:"idempotencyKey" gorm:"size:191;column:idempotency_key;uniqueIndex:uk_notification_outbox_idempotency"`
	SourceType     string    `json:"sourceType" gorm:"size:64;column:source_type;index:idx_notification_outbox_source,priority:1"`
	SourceID       string    `json:"sourceId" gorm:"size:128;column:source_id;index:idx_notification_outbox_source,priority:2"`
	Channel        string    `json:"channel" gorm:"size:32;column:notification_channel"`
	RecipientJSON  string    `json:"recipientJson" gorm:"type:mediumtext;column:recipient_json"`
	PayloadJSON    string    `json:"payloadJson" gorm:"type:mediumtext;column:payload_json"`
	Status         string    `json:"status" gorm:"size:24;column:notification_status;index:idx_notification_outbox_due,priority:1"`
	Attempts       int       `json:"attempts" gorm:"column:attempts"`
	NextRetryAt    int64     `json:"nextRetryAt" gorm:"column:next_retry_at;index:idx_notification_outbox_due,priority:2"`
	LastError      string    `json:"lastError" gorm:"size:1000;column:last_error"`
	SentAt         int64     `json:"sentAt" gorm:"column:sent_at"`
	AddTime        int64     `json:"addTime" gorm:"column:add_time"`
	EditTime       int64     `json:"editTime" gorm:"column:edit_time"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (Outbox) TableName() string { return "notification_outbox" }
