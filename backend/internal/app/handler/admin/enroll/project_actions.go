package enroll

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	admincontentservice "wecheckin-backend/backend/internal/app/service/admincontent"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-打卡管理
// @Summary 更新打卡表单
// @Param id formData string true "项目ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_update_forms [post]
func (h *AdminEnrollHandler) UpdateEnrollForms(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	err := admincontentservice.UpdateEnrollForms(id, forms)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 清除打卡全部数据
// @Param id formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_clear [post]
func (h *AdminEnrollHandler) ClearEnrollAll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := admincontentservice.ClearEnrollAll(id)
	if err != nil {
		response.Fail(c, "清除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 删除打卡项目
// @Param id formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_del [post]
func (h *AdminEnrollHandler) DelEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := admincontentservice.DelEnroll(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) DelEnrolls(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := admincontentservice.DelEnrolls(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 打卡项目排序
// @Param id formData string true "项目ID"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_sort [post]
func (h *AdminEnrollHandler) SortEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	sortStr := c.PostForm("sort")
	err := admincontentservice.SortEnroll(id, sortStr)
	if err != nil {
		response.Fail(c, "排序失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 设置打卡推荐
// @Param id formData string true "项目ID"
// @Param vouch formData string true "推荐值"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_vouch [post]
func (h *AdminEnrollHandler) VouchEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	vouch, _ := strconv.Atoi(c.PostForm("vouch"))
	err := admincontentservice.VouchEnroll(id, vouch)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 设置打卡状态
// @Param id formData string true "项目ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_status [post]
func (h *AdminEnrollHandler) StatusEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := admincontentservice.StatusEnroll(id, status)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
