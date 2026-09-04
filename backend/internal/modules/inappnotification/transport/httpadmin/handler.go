package httpadmin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/inappnotification/application"
	"wecheckin/backend/pkg/response"
)

type Service interface {
	Send(context.Context, application.SendInput) (application.SendResult, error)
	SendDingTalk(context.Context, application.SendInput) (application.SendResult, error)
	List(context.Context, uint, int, int) (application.NotificationList, error)
	UnreadCount(context.Context, uint) (int64, error)
	MarkRead(context.Context, uint, uint) error
	MarkAllRead(context.Context, uint) error
	RecipientOptions(context.Context) (application.RecipientOptions, error)
}

type Handler struct {
	service     Service
	newSourceID func() (string, error)
}

type SendRequest struct {
	RequestID     string                     `json:"requestId"`
	Title         string                     `json:"title"`
	Content       string                     `json:"content"`
	Scope         application.RecipientScope `json:"scope"`
	UserIDs       []uint                     `json:"userIds"`
	DepartmentIDs []uint                     `json:"departmentIds"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service, newSourceID: randomSourceID}
}

// List godoc
// @Summary 查询当前管理员站内信
// @Tags Admin站内信
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications [get]
func (handler *Handler) List(ctx context.Context, c *app.RequestContext) {
	admin, ok := authenticatedAdmin(c)
	if !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	result, err := handler.service.List(ctx, admin.ID, queryInt(c, "page"), queryInt(c, "pageSize"))
	respond(c, result, err)
}

// UnreadCount godoc
// @Summary 查询当前管理员未读站内信数量
// @Tags Admin站内信
// @Produce json
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/unread-count [get]
func (handler *Handler) UnreadCount(ctx context.Context, c *app.RequestContext) {
	admin, ok := authenticatedAdmin(c)
	if !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	count, err := handler.service.UnreadCount(ctx, admin.ID)
	if err != nil {
		response.Fail(c, localizedError(err))
		return
	}
	response.JSON(c, map[string]int64{"count": count})
}

// RecipientOptions godoc
// @Summary 查询可选站内信收件范围
// @Tags Admin站内信
// @Produce json
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/recipient-options [get]
func (handler *Handler) RecipientOptions(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	result, err := handler.service.RecipientOptions(ctx)
	respond(c, result, err)
}

// Send godoc
// @Summary 手动发送站内信
// @Tags Admin站内信
// @Accept json
// @Produce json
// @Param request body SendRequest true "发送内容与收件范围"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications [post]
func (handler *Handler) Send(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	var request SendRequest
	if err := decodeJSONBody(c, &request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	sourceID := strings.TrimSpace(request.RequestID)
	if sourceID == "" {
		var err error
		sourceID, err = handler.newSourceID()
		if err != nil {
			response.Fail(c, "生成发送标识失败")
			return
		}
	}
	result, err := handler.service.Send(ctx, application.SendInput{
		Title:         request.Title,
		Content:       request.Content,
		Scope:         request.Scope,
		UserIDs:       request.UserIDs,
		DepartmentIDs: request.DepartmentIDs,
		SourceType:    application.SourceAdminManual,
		SourceID:      sourceID,
	})
	respond(c, result, err)
}

// SendDingTalk godoc
// @Summary 手动发送钉钉通知
// @Tags Admin钉钉通知
// @Accept json
// @Produce json
// @Param request body SendRequest true "发送内容与收件范围"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/dingtalk-notifications [post]
func (handler *Handler) SendDingTalk(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	var request SendRequest
	if err := decodeJSONBody(c, &request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	sourceID := strings.TrimSpace(request.RequestID)
	if sourceID == "" {
		var err error
		sourceID, err = handler.newSourceID()
		if err != nil {
			response.Fail(c, "生成发送标识失败")
			return
		}
	}
	result, err := handler.service.SendDingTalk(ctx, application.SendInput{
		Title:         request.Title,
		Content:       request.Content,
		Scope:         request.Scope,
		UserIDs:       request.UserIDs,
		DepartmentIDs: request.DepartmentIDs,
		SourceType:    application.SourceAdminManualDingTalk,
		SourceID:      sourceID,
	})
	respond(c, result, err)
}

// MarkRead godoc
// @Summary 标记当前管理员的一条站内信为已读
// @Tags Admin站内信
// @Produce json
// @Param id path int true "站内信ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/{id}/read [patch]
func (handler *Handler) MarkRead(ctx context.Context, c *app.RequestContext) {
	admin, ok := authenticatedAdmin(c)
	if !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	notificationID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || notificationID == 0 {
		response.Fail(c, "站内信ID无效")
		return
	}
	respond(c, nil, handler.service.MarkRead(ctx, admin.ID, uint(notificationID)))
}

// MarkAllRead godoc
// @Summary 标记当前管理员的全部站内信为已读
// @Tags Admin站内信
// @Produce json
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/read-all [patch]
func (handler *Handler) MarkAllRead(ctx context.Context, c *app.RequestContext) {
	admin, ok := authenticatedAdmin(c)
	if !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	respond(c, nil, handler.service.MarkAllRead(ctx, admin.ID))
}

func authenticatedAdmin(c *app.RequestContext) (*model.Admin, bool) {
	value, ok := c.Get("admin")
	if !ok {
		return nil, false
	}
	admin, ok := value.(*model.Admin)
	return admin, ok && admin != nil && admin.ID > 0
}

func queryInt(c *app.RequestContext, key string) int {
	value, _ := strconv.Atoi(c.Query(key))
	return value
}

func decodeJSONBody(c *app.RequestContext, target interface{}) error {
	if c == nil || len(c.Request.Body()) == 0 {
		return errors.New("request body is required")
	}
	return json.Unmarshal(c.Request.Body(), target)
}

func randomSourceID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func respond(c *app.RequestContext, data interface{}, err error) {
	if err != nil {
		response.Fail(c, localizedError(err))
		return
	}
	response.JSON(c, data)
}

func localizedError(err error) string {
	switch {
	case errors.Is(err, application.ErrTitleRequired):
		return "请输入通知标题"
	case errors.Is(err, application.ErrTitleTooLong):
		return "通知标题不能超过255个字符"
	case errors.Is(err, application.ErrContentRequired):
		return "请输入通知正文"
	case errors.Is(err, application.ErrContentTooLong):
		return "通知正文不能超过5000个字符"
	case errors.Is(err, application.ErrInvalidScope):
		return "收件范围无效"
	case errors.Is(err, application.ErrRecipientsRequired):
		return "请选择收件人或部门"
	case errors.Is(err, application.ErrNoRecipients):
		return "当前范围内没有可接收通知的启用用户"
	case errors.Is(err, application.ErrNotificationMissing):
		return "站内信不存在"
	case errors.Is(err, application.ErrDingTalkDeliveryUnavailable):
		return "钉钉通知发送通道未初始化"
	default:
		return err.Error()
	}
}
