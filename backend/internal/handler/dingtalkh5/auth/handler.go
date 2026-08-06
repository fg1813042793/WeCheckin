package auth

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) PublicConfig(ctx context.Context, c *app.RequestContext) {
	data, err := dingtalkh5service.PublicConfigContext(ctx)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) Login(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		AuthCode string `json:"authCode"`
	}
	_ = c.BindAndValidate(&req)
	if req.Name == "" {
		req.Name = c.PostForm("name")
	}
	if req.Password == "" {
		req.Password = c.PostForm("password")
	}
	data, err := dingtalkh5service.LoginContext(ctx, req.Name, req.Password, c.ClientIP(), string(c.UserAgent()))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) SSOLogin(ctx context.Context, c *app.RequestContext) {
	var req struct {
		CorpID   string `json:"corpId"`
		AuthCode string `json:"authCode"`
	}
	_ = c.BindAndValidate(&req)
	if req.CorpID == "" {
		req.CorpID = c.PostForm("corpId")
	}
	if req.CorpID == "" {
		req.CorpID = c.Query("corpId")
	}
	if req.AuthCode == "" {
		req.AuthCode = c.PostForm("authCode")
	}
	data, err := dingtalkh5service.LoginByAuthCodeContext(ctx, req.CorpID, req.AuthCode, c.ClientIP(), string(c.UserAgent()))
	if err != nil {
		if bindData, ok := dingtalkh5service.DingTalkH5BindRequiredData(err); ok {
			c.JSON(consts.StatusOK, response.Resp{
				Code: dingtalkh5service.DingTalkH5BindRequiredCode,
				Msg:  err.Error(),
				Data: bindData,
			})
			return
		}
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) BindSelf(ctx context.Context, c *app.RequestContext) {
	var req struct {
		BindTicket string `json:"bindTicket"`
		Account    string `json:"account"`
		Password   string `json:"password"`
	}
	_ = c.BindAndValidate(&req)
	if req.BindTicket == "" {
		req.BindTicket = c.PostForm("bindTicket")
	}
	if req.Account == "" {
		req.Account = c.PostForm("account")
	}
	if req.Password == "" {
		req.Password = c.PostForm("password")
	}
	data, err := dingtalkh5service.BindSelfContext(ctx, req.BindTicket, req.Account, req.Password, c.ClientIP(), string(c.UserAgent()))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) Logout(ctx context.Context, c *app.RequestContext) {
	user, _ := dingtalkh5session.CurrentUser(c)
	if err := dingtalkh5service.LogoutContext(ctx, user, dingtalkh5session.CurrentToken(c)); err != nil {
		response.Fail(c, "退出失败")
		return
	}
	response.JSON(c, nil)
}
