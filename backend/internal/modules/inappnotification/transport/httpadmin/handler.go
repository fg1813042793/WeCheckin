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
	"wecheckin/backend/internal/support/notificationstyle"
	"wecheckin/backend/pkg/response"
)

type Service interface {
	Send(context.Context, application.SendInput) (application.SendResult, error)
	SendDingTalk(context.Context, application.SendInput) (application.SendResult, error)
	ListRecords(context.Context, application.NotificationRecordQuery) (application.NotificationList, error)
	DeleteRecord(context.Context, uint, uint) error
	UnreadCount(context.Context, uint) (int64, error)
	MarkRead(context.Context, uint, uint) error
	MarkAllRead(context.Context, uint) error
	RecipientOptions(context.Context) (application.RecipientOptions, error)
	NotificationStyles(context.Context) (notificationstyle.Config, error)
	SaveNotificationStyles(context.Context, notificationstyle.Config) (notificationstyle.Config, error)
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

type StyleTestRequest struct {
	RequestID        string `json:"requestId"`
	NotificationType string `json:"notificationType"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	UserIDs          []uint `json:"userIds"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service, newSourceID: randomSourceID}
}

// List godoc
// @Summary 查询通知投递记录
// @Tags API v2-后台管理-通知记录
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param title query string false "通知标题，模糊匹配"
// @Param recipientName query string false "接收人用户名，模糊匹配"
// @Param sourceType query string false "来源类型"
// @Param type query string false "通知类型"
// @Param isRead query int false "阅读状态：0未读，1已读"
// @Param addTimeFrom query int false "发送开始时间，毫秒时间戳"
// @Param addTimeTo query int false "发送结束时间，毫秒时间戳"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications [get]
func (handler *Handler) List(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	query, err := notificationRecordQuery(c)
	if err != nil {
		response.Fail(c, "通知记录筛选参数格式无效")
		return
	}
	result, err := handler.service.ListRecords(ctx, query)
	respond(ctx, c, result, err)
}

// DeleteRecord godoc
// @Summary 删除通知投递记录
// @Description 仅从后台通知记录列表中移除，不撤回接收人已经收到的站内信
// @Tags API v2-后台管理-通知记录
// @Produce json
// @Param id path int true "通知记录ID"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/in-app-notifications/{id} [delete]
func (handler *Handler) DeleteRecord(ctx context.Context, c *app.RequestContext) {
	admin, ok := authenticatedAdmin(c)
	if !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	notificationID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || notificationID == 0 {
		response.Fail(c, "通知记录ID无效")
		return
	}
	if err := handler.service.DeleteRecord(ctx, admin.ID, uint(notificationID)); err != nil {
		respond(ctx, c, nil, err)
		return
	}
	response.JSON(c, map[string]uint{"id": uint(notificationID)})
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
	respond(ctx, c, result, err)
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
	respond(ctx, c, result, err)
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
	respond(ctx, c, result, err)
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
	respond(ctx, c, nil, handler.service.MarkRead(ctx, admin.ID, uint(notificationID)))
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
	respond(ctx, c, nil, handler.service.MarkAllRead(ctx, admin.ID))
}

// NotificationStyles godoc
// @Summary 查询通知消息样式
// @Tags API v2-后台管理-通知样式
// @Produce json
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/notification-styles [get]
func (handler *Handler) NotificationStyles(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	result, err := handler.service.NotificationStyles(ctx)
	respond(ctx, c, result, err)
}

// SaveNotificationStyles godoc
// @Summary 保存通知消息样式
// @Tags API v2-后台管理-通知样式
// @Accept json
// @Produce json
// @Param request body notificationstyle.Config true "消息样式配置"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/notification-styles [put]
func (handler *Handler) SaveNotificationStyles(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	var request notificationstyle.Config
	if err := decodeJSONBody(c, &request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	result, err := handler.service.SaveNotificationStyles(ctx, request)
	respond(ctx, c, result, err)
}

// SendInAppStyleTest godoc
// @Summary 发送站内信样式测试
// @Tags API v2-后台管理-通知样式
// @Accept json
// @Produce json
// @Param request body StyleTestRequest true "测试消息与指定收件人"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/notification-styles/test/in-app [post]
func (handler *Handler) SendInAppStyleTest(ctx context.Context, c *app.RequestContext) {
	handler.sendStyleTest(ctx, c, false)
}

// SendDingTalkStyleTest godoc
// @Summary 发送钉钉样式测试
// @Tags API v2-后台管理-通知样式
// @Accept json
// @Produce json
// @Param request body StyleTestRequest true "测试消息与指定收件人"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/notification-styles/test/dingtalk [post]
func (handler *Handler) SendDingTalkStyleTest(ctx context.Context, c *app.RequestContext) {
	handler.sendStyleTest(ctx, c, true)
}

func (handler *Handler) sendStyleTest(ctx context.Context, c *app.RequestContext, dingTalk bool) {
	if _, ok := authenticatedAdmin(c); !ok {
		response.Fail(c, "管理员未登录")
		return
	}
	var request StyleTestRequest
	if err := decodeJSONBody(c, &request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if strings.TrimSpace(request.Title) == "" {
		response.Fail(c, localizedError(application.ErrTitleRequired))
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
	input := application.SendInput{
		Title:            styleTestTitle(request.Title),
		Content:          request.Content,
		Scope:            application.ScopeUsers,
		UserIDs:          request.UserIDs,
		NotificationType: request.NotificationType,
		SourceType:       application.SourceAdminManual,
		SourceID:         sourceID,
	}
	if dingTalk {
		input.SourceType = application.SourceAdminManualDingTalk
		result, err := handler.service.SendDingTalk(ctx, input)
		respond(ctx, c, result, err)
		return
	}
	result, err := handler.service.Send(ctx, input)
	respond(ctx, c, result, err)
}

func styleTestTitle(value string) string {
	const prefix = "[样式测试] "
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, prefix) {
		return value
	}
	return prefix + value
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

func notificationRecordQuery(c *app.RequestContext) (application.NotificationRecordQuery, error) {
	isRead, err := optionalQueryInt(c, "isRead")
	if err != nil {
		return application.NotificationRecordQuery{}, err
	}
	addTimeFrom, err := optionalQueryInt64(c, "addTimeFrom")
	if err != nil {
		return application.NotificationRecordQuery{}, err
	}
	addTimeTo, err := optionalQueryInt64(c, "addTimeTo")
	if err != nil {
		return application.NotificationRecordQuery{}, err
	}
	return application.NotificationRecordQuery{
		Title: c.Query("title"), RecipientName: c.Query("recipientName"),
		SourceType: c.Query("sourceType"), Type: c.Query("type"), IsRead: isRead,
		AddTimeFrom: addTimeFrom, AddTimeTo: addTimeTo,
		Page: queryInt(c, "page"), PageSize: queryInt(c, "pageSize"),
	}, nil
}

func optionalQueryInt(c *app.RequestContext, key string) (*int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalQueryInt64(c *app.RequestContext, key string) (int64, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
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

func respond(ctx context.Context, c *app.RequestContext, data interface{}, err error) {
	if err != nil {
		if message := localizedError(err); message != "" {
			response.Fail(c, message)
		} else {
			response.FailInternal(ctx, c, "inappnotification.admin", "通知操作失败，请稍后重试", err)
		}
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
	case errors.Is(err, application.ErrInvalidNotificationType):
		return "消息类型无效"
	case errors.Is(err, application.ErrInvalidReadStatus):
		return "阅读状态筛选无效"
	case errors.Is(err, application.ErrInvalidTimeRange):
		return "发送时间范围无效"
	case errors.Is(err, notificationstyle.ErrUnknownType):
		return "消息样式中包含不支持的消息类型"
	case errors.Is(err, notificationstyle.ErrDuplicateType):
		return "消息样式中存在重复的消息类型"
	case errors.Is(err, notificationstyle.ErrInvalidLabel):
		return "消息样式标签必须为1至20个字符"
	case errors.Is(err, notificationstyle.ErrInvalidIcon):
		return "消息样式图标无效"
	case errors.Is(err, notificationstyle.ErrInvalidTone):
		return "消息样式色调无效"
	case errors.Is(err, notificationstyle.ErrInvalidDingTalkTemplate):
		return "钉钉消息模板配置无效"
	default:
		return ""
	}
}
