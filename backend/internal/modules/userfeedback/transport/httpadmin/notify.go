package httpadmin

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/pkg/response"
)

type notifySubmitterBody struct {
	Content   string `json:"content"`
	Channel   string `json:"channel"`
	Version   uint64 `json:"version"`
	RequestID string `json:"requestId"`
}

func (handler *Handler) NotifySubmitter(ctx context.Context, c *app.RequestContext) {
	adminID, ok := currentAdminID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	if !handler.requireService(ctx, c) {
		return
	}
	id, ok := positivePathID(c, "id")
	if !ok {
		response.Fail(c, "反馈编号不正确")
		return
	}
	var body notifySubmitterBody
	if err := decodeExactJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数不正确")
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
	content := strings.TrimSpace(body.Content)
	if content == "" || !utf8.ValidString(content) || utf8.RuneCountInString(content) > application.MaxStatusNoteRunes {
		response.Fail(c, "请填写通知内容，最多5000字")
		return
	}
	data, err := handler.service.NotifySubmitter(ctx, application.NotifySubmitterCommand{
		FeedbackID: id, AdminID: adminID, Content: content, Channel: body.Channel, Version: body.Version, RequestID: body.RequestID,
	})
	respond(ctx, c, data, err)
}
