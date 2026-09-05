package httpdingtalkh5

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"path"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

const MaxMultipartBodyBytes = 31 * 1024 * 1024

var (
	errMultipartRequired     = errors.New("feedback request must use multipart/form-data")
	errMultipartBodyTooLarge = errors.New("feedback multipart body is too large")
	errTooManyImages         = errors.New("too many feedback images in one request")
	errEmptyImage            = errors.New("feedback image is empty")
)

type Service interface {
	GetUserOverview(context.Context, uint) (application.Overview, error)
	ListUserFeedbacks(context.Context, uint, application.UserListQuery) (application.FeedbackList, error)
	CreateFeedback(context.Context, application.CreateCommand) (*application.FeedbackDetail, error)
	GetUserFeedback(context.Context, uint64, uint) (*application.FeedbackDetail, error)
	SupplementFeedback(context.Context, application.SupplementCommand) (*application.FeedbackDetail, error)
}

var _ Service = (*application.Service)(nil)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Overview(ctx context.Context, c *app.RequestContext) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	data, err := handler.service.GetUserOverview(ctx, userID)
	respond(ctx, c, data, err)
}

func (handler *Handler) List(ctx context.Context, c *app.RequestContext) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	query, err := userListQuery(c)
	if err != nil {
		respond(ctx, c, nil, application.ErrInvalidArgument)
		return
	}
	data, err := handler.service.ListUserFeedbacks(ctx, userID, query)
	respond(ctx, c, data, err)
}

func (handler *Handler) Create(ctx context.Context, c *app.RequestContext) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	form, err := parseMultipartMessage(c, false)
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	data, err := handler.service.CreateFeedback(ctx, application.CreateCommand{
		SubmitterID: userID,
		Content:     form.content,
		Attachments: form.attachments,
		RequestID:   form.requestID,
	})
	respond(ctx, c, data, err)
}

func (handler *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := positivePathID(c, "id")
	if !ok {
		response.Fail(c, "反馈编号不正确")
		return
	}
	data, err := handler.service.GetUserFeedback(ctx, id, userID)
	respond(ctx, c, data, err)
}

func (handler *Handler) Supplement(ctx context.Context, c *app.RequestContext) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := positivePathID(c, "id")
	if !ok {
		response.Fail(c, "反馈编号不正确")
		return
	}
	form, err := parseMultipartMessage(c, true)
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	data, err := handler.service.SupplementFeedback(ctx, application.SupplementCommand{
		FeedbackID:  id,
		SubmitterID: userID,
		Content:     form.content,
		Attachments: form.attachments,
		Version:     form.version,
		RequestID:   form.requestID,
	})
	respond(ctx, c, data, err)
}

type multipartMessage struct {
	content     string
	requestID   string
	version     uint64
	attachments []application.AttachmentInput
}

func parseMultipartMessage(c *app.RequestContext, requireVersion bool) (multipartMessage, error) {
	if c == nil {
		return multipartMessage{}, application.ErrInvalidArgument
	}
	if err := checkMultipartBodySize(c); err != nil {
		return multipartMessage{}, err
	}
	contentType, params, err := mime.ParseMediaType(string(c.Request.Header.ContentType()))
	if err != nil || !strings.EqualFold(contentType, "multipart/form-data") || strings.TrimSpace(params["boundary"]) == "" {
		return multipartMessage{}, errMultipartRequired
	}
	form, err := c.MultipartForm()
	if err != nil {
		return multipartMessage{}, application.ErrInvalidArgument
	}
	defer c.Request.RemoveMultipartFormFiles()

	content, err := singleFormValue(form, "content")
	if err != nil {
		return multipartMessage{}, err
	}
	requestID, err := singleFormValue(form, "requestId")
	if err != nil {
		return multipartMessage{}, err
	}
	var version uint64
	if requireVersion {
		value, err := singleFormValue(form, "version")
		if err != nil {
			return multipartMessage{}, err
		}
		version, err = strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err != nil || version == 0 {
			return multipartMessage{}, application.ErrInvalidArgument
		}
	}

	for field := range form.File {
		if field != "images" {
			return multipartMessage{}, application.ErrInvalidArgument
		}
	}
	attachments, err := readAttachments(form.File["images"])
	if err != nil {
		return multipartMessage{}, err
	}
	return multipartMessage{content: content, requestID: requestID, version: version, attachments: attachments}, nil
}

func checkMultipartBodySize(c *app.RequestContext) error {
	if contentLength := c.Request.Header.ContentLength(); contentLength > MaxMultipartBodyBytes {
		return errMultipartBodyTooLarge
	}
	if len(c.Request.Body()) > MaxMultipartBodyBytes {
		return errMultipartBodyTooLarge
	}
	return nil
}

func singleFormValue(form *multipart.Form, key string) (string, error) {
	values := form.Value[key]
	if len(values) == 0 {
		return "", nil
	}
	if len(values) != 1 {
		return "", application.ErrInvalidArgument
	}
	return values[0], nil
}

func readAttachments(files []*multipart.FileHeader) ([]application.AttachmentInput, error) {
	if len(files) > application.MaxImagesPerMessage {
		return nil, errTooManyImages
	}
	for _, fileHeader := range files {
		if fileHeader == nil || fileHeader.Size == 0 {
			return nil, errEmptyImage
		}
		if fileHeader.Size < 0 {
			return nil, application.ErrInvalidArgument
		}
		if fileHeader.Size > application.MaxImageSizeBytes {
			return nil, application.ErrAttachmentTooLarge
		}
	}

	attachments := make([]application.AttachmentInput, 0, len(files))
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(io.LimitReader(file, application.MaxImageSizeBytes+1))
		closeErr := file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(content) == 0 {
			return nil, errEmptyImage
		}
		if len(content) > application.MaxImageSizeBytes {
			return nil, application.ErrAttachmentTooLarge
		}
		filename := safeBaseName(fileHeader.Filename)
		if filename == "" {
			return nil, application.ErrInvalidArgument
		}
		attachments = append(attachments, application.AttachmentInput{
			OriginalName: filename,
			ContentType:  fileHeader.Header.Get("Content-Type"),
			SizeBytes:    uint64(len(content)),
			Content:      content,
		})
	}
	return attachments, nil
}

func safeBaseName(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), `\`, "/")
	value = path.Base(value)
	if value == "." || value == "/" {
		return ""
	}
	return value
}

func currentUserID(c *app.RequestContext) (uint, bool) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok || user == nil || user.ID == 0 {
		return 0, false
	}
	return user.ID, true
}

func positivePathID(c *app.RequestContext, key string) (uint64, bool) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param(key)), 10, 64)
	return value, err == nil && value > 0
}

func userListQuery(c *app.RequestContext) (application.UserListQuery, error) {
	page, err := optionalNonNegativeInt(c, "page")
	if err != nil {
		return application.UserListQuery{}, err
	}
	pageSize, err := optionalNonNegativeInt(c, "pageSize")
	if err != nil || pageSize > application.MaxPageSize {
		return application.UserListQuery{}, application.ErrInvalidArgument
	}
	status := domain.Status(strings.TrimSpace(c.Query("status")))
	if status != "" && !status.Valid() {
		return application.UserListQuery{}, application.ErrInvalidArgument
	}
	return application.UserListQuery{
		Keyword:  c.Query("keyword"),
		Status:   status,
		Page:     page,
		PageSize: pageSize,
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

func respond(ctx context.Context, c *app.RequestContext, data interface{}, err error) {
	if err == nil {
		response.JSON(c, data)
		return
	}
	if message := localizedError(err); message != "" {
		response.Fail(c, message)
		return
	}
	response.FailInternal(ctx, c, "userfeedback.dingtalkh5", "反馈操作失败，请稍后重试", err)
}

func localizedError(err error) string {
	switch {
	case errors.Is(err, errMultipartBodyTooLarge):
		return "反馈请求总大小不能超过 31MB"
	case errors.Is(err, errMultipartRequired):
		return "请使用 multipart/form-data 提交反馈"
	case errors.Is(err, errTooManyImages):
		return "每次最多上传 6 张图片"
	case errors.Is(err, errEmptyImage):
		return "图片文件不能为空"
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
