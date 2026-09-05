package application

import (
	"context"
	"errors"

	"wecheckin/backend/internal/support/notificationstyle"
)

type RecipientScope string

const (
	ScopeAll         RecipientScope = "all"
	ScopeDepartments RecipientScope = "departments"
	ScopeUsers       RecipientScope = "users"

	SourceAdminManual         = "admin_manual"
	SourceAdminManualDingTalk = "admin_manual_dingtalk"
	SourceScheduledTaskRun    = "scheduled_task_run"
	SourceSurvey              = "survey"

	TypeAdminManual   = "admin_manual"
	TypeScheduledTask = "scheduled_task"
	TypeSurveyStat    = "survey_stat"
)

var (
	ErrUnauthenticated             = errors.New("notification user is not authenticated")
	ErrNotificationMissing         = errors.New("notification not found")
	ErrTitleRequired               = errors.New("notification title is required")
	ErrTitleTooLong                = errors.New("notification title must not exceed 255 characters")
	ErrContentRequired             = errors.New("notification content is required")
	ErrContentTooLong              = errors.New("notification content must not exceed 5000 characters")
	ErrInvalidScope                = errors.New("notification recipient scope is invalid")
	ErrRecipientsRequired          = errors.New("notification recipients are required")
	ErrSourceRequired              = errors.New("notification source is required")
	ErrSourceTooLong               = errors.New("notification source ID must not exceed 64 characters")
	ErrNoRecipients                = errors.New("notification has no active recipients")
	ErrDingTalkDeliveryUnavailable = errors.New("DingTalk notification delivery is unavailable")
	ErrInvalidNotificationType     = errors.New("notification type is invalid")
	ErrInvalidReadStatus           = errors.New("notification read status is invalid")
	ErrInvalidTimeRange            = errors.New("notification time range is invalid")
)

type RecipientRule struct {
	Scope         RecipientScope `json:"scope"`
	UserIDs       []uint         `json:"userIds,omitempty"`
	DepartmentIDs []uint         `json:"departmentIds,omitempty"`
}

type SendInput struct {
	Title            string         `json:"title"`
	Content          string         `json:"content"`
	Scope            RecipientScope `json:"scope"`
	UserIDs          []uint         `json:"userIds,omitempty"`
	DepartmentIDs    []uint         `json:"departmentIds,omitempty"`
	NotificationType string         `json:"notificationType,omitempty"`
	SourceType       string         `json:"-"`
	SourceID         string         `json:"-"`
	DeliveryKey      string         `json:"-"`
}

type RecipientResolution struct {
	UserIDs      []uint
	SkippedCount int
}

type DeliveryBatch struct {
	Title       string
	Content     string
	Type        string
	SourceType  string
	SourceID    string
	DeliveryKey string
	UserIDs     []uint
}

type DeliveryResult struct {
	SentCount int
	Replayed  bool
}

type DingTalkDeliveryBatch struct {
	Title            string
	Content          string
	NotificationType string
	UserIDs          []uint
}

type DingTalkDeliveryResult struct {
	SentCount    int
	SkippedCount int
	FailedCount  int
}

type DingTalkDelivery interface {
	DeliverDingTalk(context.Context, DingTalkDeliveryBatch) (DingTalkDeliveryResult, error)
}

type SendResult struct {
	SourceID     string `json:"sourceId"`
	PlannedCount int    `json:"plannedCount"`
	SentCount    int    `json:"sentCount"`
	SkippedCount int    `json:"skippedCount"`
	FailedCount  int    `json:"failedCount"`
	Replayed     bool   `json:"replayed"`
}

type Notification struct {
	ID              uint   `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	Type            string `json:"type"`
	SourceType      string `json:"sourceType"`
	SourceID        string `json:"sourceId"`
	RecipientUserID string `json:"recipientUserId"`
	RecipientName   string `json:"recipientName"`
	IsRead          int    `json:"isRead"`
	AddTime         int64  `json:"addTime"`
}

type NotificationRecordQuery struct {
	Title         string
	RecipientName string
	SourceType    string
	Type          string
	IsRead        *int
	AddTimeFrom   int64
	AddTimeTo     int64
	Page          int
	PageSize      int
}

type NotificationList struct {
	List     []Notification `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type RecipientUserOption struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Mobile  string `json:"mobile"`
	Status  int    `json:"status"`
	DeptIDs []uint `json:"deptIds"`
}

type RecipientDepartmentOption struct {
	ID       uint                         `json:"id"`
	Name     string                       `json:"name"`
	Children []*RecipientDepartmentOption `json:"children,omitempty"`
}

type RecipientOptions struct {
	Users       []RecipientUserOption        `json:"users"`
	Departments []*RecipientDepartmentOption `json:"departments"`
}

type Store interface {
	ResolveRecipients(context.Context, RecipientRule) (RecipientResolution, error)
	Deliver(context.Context, DeliveryBatch) (DeliveryResult, error)
	ListRecords(context.Context, NotificationRecordQuery) ([]Notification, int64, error)
	SoftDeleteRecord(context.Context, uint, string, int64) (bool, error)
	UnreadCount(context.Context, string) (int64, error)
	MarkRead(context.Context, string, uint) (bool, error)
	MarkAllRead(context.Context, string) error
	RecipientOptions(context.Context) (RecipientOptions, error)
	NotificationStyles(context.Context) (notificationstyle.Config, error)
	SaveNotificationStyles(context.Context, notificationstyle.Config) (notificationstyle.Config, error)
}
