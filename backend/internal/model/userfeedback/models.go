package userfeedbackmodel

import "time"

type Feedback struct {
	ID              uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	FeedbackNo      string `gorm:"column:feedback_no;size:32"`
	SubmitterID     uint   `gorm:"column:submitter_id"`
	CreateRequestID string `gorm:"column:create_request_id;size:64"`
	Status          string `gorm:"column:feedback_status;size:24;default:pending"`
	HandlerID       *uint  `gorm:"column:handler_id"`
	Version         uint64 `gorm:"column:version;default:1"`
	LastActivityAt  int64  `gorm:"column:last_activity_at"`
	ResolvedAt      *int64 `gorm:"column:resolved_at"`
	ClosedAt        *int64 `gorm:"column:closed_at"`
	CreatedAt       int64  `gorm:"column:created_at;autoCreateTime:false;autoUpdateTime:false"`
	UpdatedAt       int64  `gorm:"column:updated_at;autoCreateTime:false;autoUpdateTime:false"`
}

func (Feedback) TableName() string { return "user_feedbacks" }

type Message struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	FeedbackID  uint64 `gorm:"column:feedback_id"`
	MessageType string `gorm:"column:message_type;size:24"`
	AuthorType  string `gorm:"column:author_type;size:16"`
	AuthorID    uint   `gorm:"column:author_id"`
	Content     string `gorm:"column:content;type:text"`
	FromStatus  string `gorm:"column:from_status;size:24"`
	ToStatus    string `gorm:"column:to_status;size:24"`
	RequestID   string `gorm:"column:request_id;size:64"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:false;autoUpdateTime:false"`
}

func (Message) TableName() string { return "user_feedback_messages" }

type Attachment struct {
	ID              uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	FeedbackID      uint64 `gorm:"column:feedback_id"`
	MessageID       uint64 `gorm:"column:message_id"`
	StorageProvider string `gorm:"column:storage_provider;size:32"`
	ObjectKey       string `gorm:"column:object_key;size:512"`
	OriginalName    string `gorm:"column:original_name;size:255"`
	ContentType     string `gorm:"column:content_type;size:127"`
	SizeBytes       uint64 `gorm:"column:size_bytes"`
	SortOrder       int    `gorm:"column:sort_order"`
	CreatedAt       int64  `gorm:"column:created_at;autoCreateTime:false;autoUpdateTime:false"`
}

func (Attachment) TableName() string { return "user_feedback_attachments" }

type DailySequence struct {
	SequenceDate time.Time `gorm:"column:sequence_date;type:date;primaryKey"`
	CurrentValue uint64    `gorm:"column:current_value"`
	UpdatedAt    int64     `gorm:"column:updated_at;autoCreateTime:false;autoUpdateTime:false"`
}

func (DailySequence) TableName() string { return "user_feedback_daily_sequences" }
