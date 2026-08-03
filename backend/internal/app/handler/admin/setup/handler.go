package setup

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	setupservice "wecheckin/backend/internal/app/service/setup"
	"wecheckin/backend/pkg/response"
	"wecheckin/backend/pkg/tokenutil"
)

type AdminSetupHandler struct{}

func NewAdminSetupHandler() *AdminSetupHandler { return &AdminSetupHandler{} }

// @Tags PC端-系统设置
// @Summary 设置系统配置
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
func (h *AdminSetupHandler) SetSetup(ctx context.Context, c *app.RequestContext) {
	key := c.PostForm("key")
	value := c.PostForm("value")
	addIP := c.ClientIP()
	err := setupservice.SetSetup(key, value, "", addIP)
	if err != nil {
		response.Fail(c, "设置失败")
		return
	}
	tokenutil.InvalidateSetupCache()
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
	err := setupservice.SetContentSetup(key, value, addIP)
	if err != nil {
		response.Fail(c, "设置失败")
		return
	}
	tokenutil.InvalidateSetupCache()
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
	userExpire, userPrefix := tokenutil.GetTokenConfig("user")
	adminExpire, adminPrefix := tokenutil.GetTokenConfig("admin")
	dingTalkH5Expire, dingTalkH5Prefix := tokenutil.GetTokenConfig("dingtalk_h5")
	response.JSON(c, map[string]interface{}{
		"user_expire_seconds":        int(userExpire.Seconds()),
		"user_expire_str":            userExpire.String(),
		"user_prefix":                userPrefix,
		"admin_expire_seconds":       int(adminExpire.Seconds()),
		"admin_expire_str":           adminExpire.String(),
		"admin_prefix":               adminPrefix,
		"dingtalk_h5_expire_seconds": int(dingTalkH5Expire.Seconds()),
		"dingtalk_h5_expire_str":     dingTalkH5Expire.String(),
		"dingtalk_h5_prefix":         dingTalkH5Prefix,
	})
}
