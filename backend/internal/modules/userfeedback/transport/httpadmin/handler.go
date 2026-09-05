package httpadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
	"wecheckin/backend/pkg/response"
)

type Service interface {
	GetAdminOverview(context.Context, application.AdminListQuery) (application.Overview, error)
	ListAdminFeedbacks(context.Context, application.AdminListQuery) (application.FeedbackList, error)
	GetAdminFeedback(context.Context, uint64) (*application.FeedbackDetail, error)
	UpdateFeedbackStatus(context.Context, application.UpdateStatusCommand) (*application.FeedbackDetail, error)
}

var _ Service = (*application.Service)(nil)

type Handler struct {
	service Service
}

type updateStatusBody struct {
	Status     domain.Status `json:"status"`
	Note       string        `json:"note"`
	NotifyUser *bool         `json:"notifyUser"`
	Version    uint64        `json:"version"`
	RequestID  string        `json:"requestId"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Overview(ctx context.Context, c *app.RequestContext) {
	if _, ok := currentAdminID(c); !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	query, err := adminListQuery(c)
	if err != nil {
		respond(ctx, c, nil, application.ErrInvalidArgument)
		return
	}
	data, err := handler.service.GetAdminOverview(ctx, query)
	respond(ctx, c, data, err)
}

func (handler *Handler) List(ctx context.Context, c *app.RequestContext) {
	if _, ok := currentAdminID(c); !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	query, err := adminListQuery(c)
	if err != nil {
		respond(ctx, c, nil, application.ErrInvalidArgument)
		return
	}
	data, err := handler.service.ListAdminFeedbacks(ctx, query)
	respond(ctx, c, data, err)
}

func (handler *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	if _, ok := currentAdminID(c); !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := positivePathID(c, "id")
	if !ok {
		response.Fail(c, "反馈编号不正确")
		return
	}
	data, err := handler.service.GetAdminFeedback(ctx, id)
	respond(ctx, c, data, err)
}

func (handler *Handler) UpdateStatus(ctx context.Context, c *app.RequestContext) {
	adminID, ok := currentAdminID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := positivePathID(c, "id")
	if !ok {
		response.Fail(c, "反馈编号不正确")
		return
	}
	var body updateStatusBody
	if err := decodeExactJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数不正确")
		return
	}
	if !body.Status.Valid() {
		response.Fail(c, "反馈状态无效")
		return
	}
	if body.Version == 0 {
		response.Fail(c, "请提供当前版本号")
		return
	}
	if strings.TrimSpace(body.RequestID) == "" {
		response.Fail(c, "请提供请求标识")
		return
	}
	notifyUser := true
	if body.NotifyUser != nil {
		notifyUser = *body.NotifyUser
	}
	data, err := handler.service.UpdateFeedbackStatus(ctx, application.UpdateStatusCommand{
		FeedbackID: id,
		AdminID:    adminID,
		Status:     body.Status,
		Note:       body.Note,
		NotifyUser: notifyUser,
		Version:    body.Version,
		RequestID:  body.RequestID,
	})
	respond(ctx, c, data, err)
}

func currentAdminID(c *app.RequestContext) (uint, bool) {
	value, ok := c.Get("admin")
	if !ok {
		return 0, false
	}
	admin, ok := value.(*model.Admin)
	if !ok || admin == nil || admin.ID == 0 {
		return 0, false
	}
	return admin.ID, true
}

func positivePathID(c *app.RequestContext, key string) (uint64, bool) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param(key)), 10, 64)
	return value, err == nil && value > 0
}

func adminListQuery(c *app.RequestContext) (application.AdminListQuery, error) {
	page, err := optionalNonNegativeInt(c, "page")
	if err != nil {
		return application.AdminListQuery{}, err
	}
	pageSize, err := optionalNonNegativeInt(c, "pageSize")
	if err != nil || pageSize > application.MaxPageSize {
		return application.AdminListQuery{}, application.ErrInvalidArgument
	}
	status := domain.Status(strings.TrimSpace(c.Query("status")))
	if status != "" && !status.Valid() {
		return application.AdminListQuery{}, application.ErrInvalidArgument
	}
	submitterID, err := optionalPositiveUint(c, "submitterId")
	if err != nil {
		return application.AdminListQuery{}, err
	}
	handlerID, err := optionalPositiveUint(c, "handlerId")
	if err != nil {
		return application.AdminListQuery{}, err
	}
	submittedFrom, err := optionalNonNegativeInt64(c, "submittedFrom")
	if err != nil {
		return application.AdminListQuery{}, err
	}
	submittedTo, err := optionalNonNegativeInt64(c, "submittedTo")
	if err != nil || (submittedFrom > 0 && submittedTo > 0 && submittedFrom > submittedTo) {
		return application.AdminListQuery{}, application.ErrInvalidArgument
	}
	return application.AdminListQuery{
		Keyword:       c.Query("keyword"),
		SubmitterID:   valueOrZero(submitterID),
		HandlerID:     handlerID,
		Status:        status,
		SubmittedFrom: submittedFrom,
		SubmittedTo:   submittedTo,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

func optionalNonNegativeInt(c *app.RequestContext, key string) (int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, application.ErrInvalidArgument
	}
	return parsed, nil
}

func optionalPositiveUint(c *app.RequestContext, key string) (*uint, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(value, 10, strconv.IntSize)
	if err != nil || parsed == 0 {
		return nil, application.ErrInvalidArgument
	}
	result := uint(parsed)
	return &result, nil
}

func optionalNonNegativeInt64(c *app.RequestContext, key string) (int64, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return 0, application.ErrInvalidArgument
	}
	return parsed, nil
}

func valueOrZero(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
}

func decodeExactJSONBody(c *app.RequestContext, target interface{}) error {
	if c == nil || len(c.Request.Body()) == 0 {
		return application.ErrInvalidArgument
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Request.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return application.ErrInvalidArgument
	}
	return nil
}

func respond(ctx context.Context, c *app.RequestContext, data interface{}, err error) {
	if err == nil {
		response.JSON(c, data)
		return
	}
	if message := localizedError(err); message != "" {
		response.Fail(c, message)
		return
	}
	response.FailInternal(ctx, c, "userfeedback.admin", "反馈操作失败，请稍后重试", err)
}

func localizedError(err error) string {
	switch {
	case errors.Is(err, application.ErrAttachmentTypeNotAllowed):
		return "仅支持 JPG、PNG、WebP 图片"
	case errors.Is(err, application.ErrAttachmentTooLarge):
		return "单张图片不能超过 10MB"
	case errors.Is(err, application.ErrAttachmentLimitExceeded):
		return "图片数量超过限制"
	case errors.Is(err, application.ErrDailyLimitExceeded):
		return "今日反馈数量已达上限"
	case errors.Is(err, application.ErrFeedbackNotFound):
		return "反馈不存在"
	case errors.Is(err, application.ErrSupplementNotAllowed):
		return "当前状态不允许补充反馈"
	case errors.Is(err, application.ErrTransitionNotAllowed):
		return "不允许执行该状态变更"
	case errors.Is(err, application.ErrTransitionNoteRequired):
		return "当前状态变更需要填写处理说明"
	case errors.Is(err, application.ErrVersionConflict):
		return "反馈已更新，请刷新后重试"
	case errors.Is(err, application.ErrDuplicateRequest):
		return "请勿重复提交"
	case errors.Is(err, application.ErrTransactionOutcomeUnknown):
		return "操作结果暂未确认，请稍后刷新"
	case errors.Is(err, application.ErrInvalidArgument):
		return "请求参数不正确"
	default:
		return ""
	}
}
