package application

import (
	"context"
	"errors"
	"wecheckin/backend/internal/model/notification"
)

var (
	ErrTemplateInvalid             = errors.New("发送模板参数无效")
	ErrTemplateNotFound            = errors.New("发送模板已删除或不存在，请刷新后重新选择")
	ErrTemplateDisabled            = errors.New("发送模板已停用，请重新选择启用模板")
	ErrTemplateVersionConflict     = errors.New("发送模板已被修改，请刷新后重新选择或编辑")
	ErrTemplateNameExists          = errors.New("模板名称已存在，请更换名称")
	ErrTemplateWorkflowUnavailable = errors.New("模板目标流程未发布或已停用/删除，请重新选择已发布流程")
	ErrTemplateStoreUnavailable    = errors.New("发送模板服务暂不可用，请稍后重试")
	ErrTemplateSchemaMissing       = errors.New("发送模板数据表尚未就绪，请联系管理员执行 backend/init.sh 迁移")
	ErrTemplateTargetUnavailable   = errors.New("收件企业的CorpID、应用标识或页面地址配置无效，请联系管理员检查钉钉应用配置")
)

type TemplateValidationError struct{ Message string }

func (e *TemplateValidationError) Error() string { return e.Message }

func (e *TemplateValidationError) Unwrap() error { return ErrTemplateInvalid }

type DingTalkTemplateTarget struct {
	Kind           string `json:"kind"`
	DefinitionID   uint   `json:"definitionId,omitempty"`
	DefinitionName string `json:"definitionName,omitempty"`
	URL            string `json:"url,omitempty"`
}

type DingTalkTemplateInput struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"messageType"`
	Target      DingTalkTemplateTarget `json:"target"`
	ButtonTitle string                 `json:"buttonTitle"`
	HeadColor   string                 `json:"headColor"`
	Version     int                    `json:"version,omitempty"`
}

type DingTalkSendTemplate struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"messageType"`
	Target      DingTalkTemplateTarget `json:"target"`
	ButtonTitle string                 `json:"buttonTitle"`
	HeadColor   string                 `json:"headColor"`
	Version     int                    `json:"version"`
	UpdatedAt   int64                  `json:"updatedAt"`
}

type DingTalkTemplateQuery struct {
	Keyword    string
	TargetKind string
	Page       int
	PageSize   int
	ID         uint
	Enabled    *bool
}

type DingTalkTemplateList struct {
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
	List     []DingTalkSendTemplate `json:"list"`
	Total    int64                  `json:"total"`
}

type DingTalkTemplateWorkflowOption struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type DingTalkTemplateWorkflowList struct {
	List     []DingTalkTemplateWorkflowOption `json:"list"`
	Total    int64                            `json:"total"`
	Page     int                              `json:"page"`
	PageSize int                              `json:"pageSize"`
}

// TemplateStore is optional so existing notification stores and fakes remain compatible.
type TemplateStore interface {
	ListTemplates(context.Context, DingTalkTemplateQuery, bool) ([]notificationmodel.DingTalkTemplate, int64, error)
	GetTemplate(context.Context, uint) (notificationmodel.DingTalkTemplate, error)
	CreateTemplate(context.Context, *notificationmodel.DingTalkTemplate) error
	UpdateTemplate(context.Context, *notificationmodel.DingTalkTemplate, int) error
	DeleteTemplate(context.Context, uint, int) error
	PublishedTemplateWorkflow(context.Context, uint) (DingTalkTemplateWorkflowOption, error)
	TemplateWorkflowOptions(context.Context, DingTalkTemplateQuery) ([]DingTalkTemplateWorkflowOption, int64, error)
}
