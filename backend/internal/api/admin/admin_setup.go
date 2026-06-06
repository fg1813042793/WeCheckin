package admin

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
	"wecheckin-backend/backend/pkg/tokenutil"
)

type AdminSetupHandler struct{}

func NewAdminSetupHandler() *AdminSetupHandler { return &AdminSetupHandler{} }

// @Tags 系统设置
// @Summary 设置系统配置
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
// @Router /admin/setup_set [post]
func (h *AdminSetupHandler) SetSetup(ctx context.Context, c *app.RequestContext) {
	key := c.PostForm("key")
	value := c.PostForm("value")
	addIP := c.ClientIP()
	err := service.SetSetup(key, value, "", addIP)
	if err != nil {
		response.Fail(c, "设置失败")
		return
	}
	tokenutil.InvalidateSetupCache()
	response.JSON(c, nil)
}

// @Tags 系统设置
// @Summary 设置内容配置
// @Param key formData string true "设置键名"
// @Param value formData string true "设置值"
// @Success 200 {object} response.Resp
// @Router /admin/setup_set_content [post]
func (h *AdminSetupHandler) SetContentSetup(ctx context.Context, c *app.RequestContext) {
	key := c.PostForm("key")
	value := c.PostForm("value")
	addIP := c.ClientIP()
	err := service.SetContentSetup(key, value, addIP)
	if err != nil {
		response.Fail(c, "设置失败")
		return
	}
	tokenutil.InvalidateSetupCache()
	response.JSON(c, nil)
}

// @Tags 系统设置
// @Summary 生成小程序码
// @Param page query string false "页面路径"
// @Param scene query string false "场景值"
// @Success 200 {object} response.Resp
// @Router /admin/setup_qr [get]
func (h *AdminSetupHandler) GenMiniQr(ctx context.Context, c *app.RequestContext) {
	response.Fail(c, "该功能暂不开放")
}

func (h *AdminSetupHandler) DebugTokenConfig(ctx context.Context, c *app.RequestContext) {
	userExpire, userPrefix := tokenutil.GetTokenConfig("user")
	adminExpire, adminPrefix := tokenutil.GetTokenConfig("admin")
	response.JSON(c, map[string]interface{}{
		"user_expire_seconds": int(userExpire.Seconds()),
		"user_expire_str":     userExpire.String(),
		"user_prefix":         userPrefix,
		"admin_expire_seconds": int(adminExpire.Seconds()),
		"admin_expire_str":    adminExpire.String(),
		"admin_prefix":        adminPrefix,
	})
}
