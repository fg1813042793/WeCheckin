package passport

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	passportservice "wecheckin/backend/internal/service/client/passport"
	"wecheckin/backend/pkg/response"
)

type dingTalkAuthorizationRequest struct {
	CorpID      string `json:"corpId"`
	RedirectURI string `json:"redirectUri"`
}

type dingTalkLoginRequest struct {
	CorpID   string `json:"corpId"`
	AuthCode string `json:"authCode"`
}

func (h *PassportHandler) DingTalkLoginConfig(ctx context.Context, c *app.RequestContext) {
	data, err := passportservice.DingTalkLoginConfigContext(ctx)
	if err != nil {
		response.FailInternal(ctx, c, "client.passport.dingtalk.config", "钉钉登录配置加载失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *PassportHandler) DingTalkAuthorization(ctx context.Context, c *app.RequestContext) {
	var req dingTalkAuthorizationRequest
	_ = c.BindAndValidate(&req)
	if req.CorpID == "" {
		req.CorpID = c.PostForm("corpId")
	}
	if req.RedirectURI == "" {
		req.RedirectURI = c.PostForm("redirectUri")
	}
	data, err := passportservice.CreateDingTalkAuthorizationContext(ctx, req.CorpID, req.RedirectURI)
	if err != nil {
		if message, ok := dingTalkLoginPublicError(err); ok {
			response.Fail(c, message)
			return
		}
		response.FailInternal(ctx, c, "client.passport.dingtalk.authorization", "钉钉登录暂不可用，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *PassportHandler) LoginByDingTalk(ctx context.Context, c *app.RequestContext) {
	var req dingTalkLoginRequest
	_ = c.BindAndValidate(&req)
	if req.CorpID == "" {
		req.CorpID = c.PostForm("corpId")
	}
	if req.AuthCode == "" {
		req.AuthCode = c.PostForm("authCode")
	}
	data, err := passportservice.LoginByDingTalkOAuthCodeContext(
		ctx,
		strings.TrimSpace(req.CorpID),
		strings.TrimSpace(req.AuthCode),
		c.ClientIP(),
		string(c.UserAgent()),
	)
	if err != nil {
		if message, ok := dingTalkLoginPublicError(err); ok {
			response.Fail(c, message)
			return
		}
		response.FailInternal(ctx, c, "client.passport.dingtalk.login", "钉钉验证失败，请重新授权", err)
		return
	}
	response.JSON(c, data)
}

func dingTalkLoginPublicError(err error) (string, bool) {
	switch {
	case errors.Is(err, passportservice.ErrDingTalkLoginNotConfigured):
		return "钉钉登录尚未配置，请联系管理员", true
	case errors.Is(err, passportservice.ErrDingTalkInvalidRedirect):
		return "钉钉登录回调地址无效，请联系管理员", true
	case errors.Is(err, passportservice.ErrDingTalkBindingRequired):
		return "当前钉钉账号尚未绑定系统用户，请联系管理员", true
	case errors.Is(err, passportservice.ErrDingTalkBindingDisabled):
		return "当前钉钉账号绑定已停用，请联系管理员", true
	case errors.Is(err, passportservice.ErrDingTalkBindingAmbiguous):
		return "当前钉钉账号存在重复绑定，请联系管理员", true
	case errors.Is(err, passportservice.ErrDingTalkUserDisabled):
		return "系统账号已停用，请联系管理员", true
	default:
		return "", false
	}
}
