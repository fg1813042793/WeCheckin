package application

import "wecheckin/backend/internal/modules/userfeedback/domain"

type Overview struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Resolved   int64 `json:"resolved"`
	Closed     int64 `json:"closed"`
}

type Attachment struct {
	ID           uint64 `json:"id,string" swaggertype:"string"`
	ObjectKey    string `json:"-"`
	OriginalName string `json:"originalName"`
	ContentType  string `json:"contentType,omitempty"`
	SizeBytes    uint64 `json:"sizeBytes"`
	URL          string `json:"url"`
	SortOrder    int    `json:"sortOrder"`
	CreatedAt    int64  `json:"createdAt"`
}

type Message struct {
	ID          uint64             `json:"id,string" swaggertype:"string"`
	MessageType domain.MessageType `json:"messageType"`
	AuthorType  domain.AuthorType  `json:"authorType"`
	AuthorID    uint               `json:"authorId"`
	AuthorName  string             `json:"authorName,omitempty"`
	Content     string             `json:"content"`
	FromStatus  domain.Status      `json:"fromStatus,omitempty"`
	ToStatus    domain.Status      `json:"toStatus,omitempty"`
	Attachments []Attachment       `json:"attachments"`
	CreatedAt   int64              `json:"createdAt"`
}

type FeedbackSummary struct {
	ID             uint64        `json:"id,string" swaggertype:"string"`
	FeedbackNo     string        `json:"feedbackNo"`
	SubmitterID    uint          `json:"submitterId"`
	SubmitterName  string        `json:"submitterName,omitempty"`
	Summary        string        `json:"summary"`
	ImageCount     int64         `json:"imageCount"`
	FirstImageURL  string        `json:"firstImageUrl,omitempty"`
	Status         domain.Status `json:"status"`
	HandlerID      *uint         `json:"handlerId,omitempty"`
	HandlerName    string        `json:"handlerName,omitempty"`
	Version        uint64        `json:"version"`
	LastActivityAt int64         `json:"lastActivityAt"`
	ResolvedAt     *int64        `json:"resolvedAt,omitempty"`
	ClosedAt       *int64        `json:"closedAt,omitempty"`
	CreatedAt      int64         `json:"createdAt"`
	UpdatedAt      int64         `json:"updatedAt"`
}

type FeedbackDetail struct {
	FeedbackSummary
	Messages         []Message `json:"messages"`
	AllowsSupplement bool      `json:"allowsSupplement"`
}

type FeedbackList struct {
	List     []FeedbackSummary `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

type UserListQuery struct {
	Keyword  string
	Status   domain.Status
	Page     int
	PageSize int
}

type AdminListQuery struct {
	Keyword       string
	SubmitterID   uint
	HandlerID     *uint
	Status        domain.Status
	SubmittedFrom int64
	SubmittedTo   int64
	Page          int
	PageSize      int
}

type AttachmentInput struct {
	OriginalName string
	ContentType  string
	SizeBytes    uint64
	Content      []byte
}

type CreateCommand struct {
	SubmitterID uint
	Content     string
	Attachments []AttachmentInput
	RequestID   string
}

type SupplementCommand struct {
	FeedbackID  uint64
	SubmitterID uint
	Content     string
	Attachments []AttachmentInput
	Version     uint64
	RequestID   string
}

type UpdateStatusCommand struct {
	FeedbackID uint64
	AdminID    uint
	Status     domain.Status
	Note       string
	NotifyUser bool
	Version    uint64
	RequestID  string
}
