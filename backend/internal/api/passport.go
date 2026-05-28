package api

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type PassportHandler struct{}

func NewPassportHandler() *PassportHandler { return &PassportHandler{} }

// @Tags 通行证
// @Summary 用户登录
// @Param user_id formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /passport/login [post]
func (h *PassportHandler) Login(ctx context.Context, c *app.RequestContext) {
	userID := c.PostForm("user_id")
	data, err := service.LoginUser(userID)
	if err != nil {
		response.Fail(c, "登录失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 通行证
// @Summary 获取手机号
// @Param cloud_id formData string true "云ID"
// @Success 200 {object} response.Resp
// @Router /passport/phone [post]
func (h *PassportHandler) GetPhone(ctx context.Context, c *app.RequestContext) {
	cloudID := c.PostForm("cloud_id")
	data, err := service.GetPhone(cloudID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 通行证
// @Summary 用户注册
// @Param user_id formData string true "用户ID"
// @Param name formData string true "姓名"
// @Param mobile formData string true "手机号"
// @Param pic formData string false "头像"
// @Success 200 {object} response.Resp
// @Router /passport/register [post]
func (h *PassportHandler) Register(ctx context.Context, c *app.RequestContext) {
	userID := c.PostForm("user_id")
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	data, err := service.RegisterUser(userID, mobile, name, pic, nil, 1)
	if err != nil {
		response.Fail(c, "注册失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 通行证
// @Summary 获取我的详情
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /passport/my_detail [get]
func (h *PassportHandler) GetMyDetail(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	data, err := service.GetMyDetail(userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 通行证
// @Summary 编辑基本信息
// @Param name formData string false "姓名"
// @Param mobile formData string false "手机号"
// @Param pic formData string false "头像"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /passport/edit_base [post]
func (h *PassportHandler) EditBase(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	userID := c.PostForm("user_id")
	err := service.EditBase(userID, mobile, name, pic, nil)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}
