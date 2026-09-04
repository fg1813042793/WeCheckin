package setup

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/pkg/response"
	"wecheckin/backend/pkg/tokenutil"
)

type AdminSetupHandler struct{}

func NewAdminSetupHandler() *AdminSetupHandler { return &AdminSetupHandler{} }

type DebugTokenConfigResponse struct {
	UserExpireSeconds       int    `json:"user_expire_seconds"`
	UserExpireString        string `json:"user_expire_str"`
	UserPrefix              string `json:"user_prefix"`
	AdminExpireSeconds      int    `json:"admin_expire_seconds"`
	AdminExpireString       string `json:"admin_expire_str"`
	AdminPrefix             string `json:"admin_prefix"`
	DingTalkH5ExpireSeconds int    `json:"dingtalk_h5_expire_seconds"`
	DingTalkH5ExpireString  string `json:"dingtalk_h5_expire_str"`
	DingTalkH5Prefix        string `json:"dingtalk_h5_prefix"`
}

// @Tags PC端-系统设置
// @Summary 获取内容配置
// @Param key query string true "设置键名"
// @Success 200 {object} response.Resp
func (h *AdminSetupHandler) GetContentSetup(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	if key == "" {
		response.Fail(c, "key 必填")
		return
	}
	setup, err := setupservice.GetSetupContext(ctx, key)
	if err != nil || setup == nil {
		response.JSON(c, nil)
		return
	}
	response.JSON(c, setup.Value)
}

// @Tags PC端-系统设置
// @Summary 设置系统配置
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
func (h *AdminSetupHandler) SetSetup(ctx context.Context, c *app.RequestContext) {
	key := c.PostForm("key")
	value := c.PostForm("value")
	addIP := c.ClientIP()
	err := setupservice.SetSetupContext(ctx, key, value, "", addIP)
	if err != nil {
		response.Fail(c, "设置失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-系统设置
// @Summary 设置内容配置
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
func (h *AdminSetupHandler) SetContentSetup(ctx context.Context, c *app.RequestContext) {
	key := c.PostForm("key")
	value := c.PostForm("value")
	addIP := c.ClientIP()
	err := setupservice.SetContentSetupContext(ctx, key, value, addIP)
	if err != nil {
		response.Fail(c, "设置失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-系统设置
// @Summary 生成小程序码
// @Param page query string false "页面路径"
// @Param scene query string false "场景值"
// @Success 200 {object} response.Resp
func (h *AdminSetupHandler) GenMiniQr(ctx context.Context, c *app.RequestContext) {
	response.Fail(c, "该功能暂不开放")
}

func (h *AdminSetupHandler) DebugTokenConfig(ctx context.Context, c *app.RequestContext) {
	userExpire, userPrefix := tokenutil.GetTokenConfigContext(ctx, "user")
	adminExpire, adminPrefix := tokenutil.GetTokenConfigContext(ctx, "admin")
	dingTalkH5Expire, dingTalkH5Prefix := tokenutil.GetTokenConfigContext(ctx, "dingtalk_h5")
	response.JSON(c, DebugTokenConfigResponse{
		UserExpireSeconds:       int(userExpire.Seconds()),
		UserExpireString:        userExpire.String(),
		UserPrefix:              userPrefix,
		AdminExpireSeconds:      int(adminExpire.Seconds()),
		AdminExpireString:       adminExpire.String(),
		AdminPrefix:             adminPrefix,
		DingTalkH5ExpireSeconds: int(dingTalkH5Expire.Seconds()),
		DingTalkH5ExpireString:  dingTalkH5Expire.String(),
		DingTalkH5Prefix:        dingTalkH5Prefix,
	})
}
