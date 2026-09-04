package dingtalk

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	admindingtalkservice "wecheckin/backend/internal/service/admin/dingtalk"
	"wecheckin/backend/pkg/response"
)

func (h *AdminDingTalkHandler) GetUserBindings(ctx context.Context, c *app.RequestContext) {
	data, err := h.service.ListUserBindings(ctx, admindingtalkservice.UserBindingQuery{
		Page: parsePositiveInt(c.Query("page"), 1), PageSize: parsePositiveInt(c.Query("pageSize"), 20),
		CorpID: c.Query("corpId"), Keyword: c.Query("keyword"), Enabled: c.Query("enabled"),
	})
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminDingTalkHandler) SaveUserBinding(ctx context.Context, c *app.RequestContext) {
	corpID := strings.TrimSpace(c.PostForm("corpId"))
	dingTalkUserID := strings.TrimSpace(c.PostForm("dingTalkUserId"))
	userID := parseUint(c.PostForm("userId"))
	if corpID == "" {
		response.Fail(c, "请选择钉钉企业")
		return
	}
	if dingTalkUserID == "" {
		response.Fail(c, "请输入钉钉 UserId")
		return
	}
	if userID == 0 {
		response.Fail(c, "请选择本地用户")
		return
	}
	enabled := 1
	if strings.TrimSpace(c.PostForm("enabled")) == "0" {
		enabled = 0
	}
	err := h.service.SaveUserBinding(ctx, admindingtalkservice.SaveUserBindingInput{
		ID: parseUint(c.PostForm("id")), CorpID: corpID, DingTalkUserID: dingTalkUserID,
		UnionID: c.PostForm("unionId"), UserID: userID, Enabled: enabled,
	})
	if err != nil {
		response.FailInternal(ctx, c, "admin.dingtalk.bindings", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) StatusUserBinding(ctx context.Context, c *app.RequestContext) {
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	enabled := 0
	if strings.TrimSpace(c.PostForm("enabled")) == "1" {
		enabled = 1
	}
	if err := h.service.SetUserBindingEnabled(ctx, id, enabled); err != nil {
		if errors.Is(err, admindingtalkservice.ErrBindingNotFound) {
			response.Fail(c, "绑定不存在")
			return
		}
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) DeleteUserBinding(ctx context.Context, c *app.RequestContext) {
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.service.DeleteUserBinding(ctx, id); err != nil {
		if errors.Is(err, admindingtalkservice.ErrBindingNotFound) {
			response.Fail(c, "绑定不存在")
			return
		}
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
