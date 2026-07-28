package enroll

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	admincontentservice "wecheckin-backend/backend/internal/app/service/admincontent"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-打卡管理
// @Summary 获取打卡数据导出链接
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_get [get]
func (h *AdminEnrollHandler) EnrollJoinDataGet(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	enrollID := c.Query("enrollId")
	if err := admincontentservice.EnsureEnrollVisibleForAdminContext(ctx, enrollID, admin.ID); err != nil {
		response.Fail(c, "获取失败")
		return
	}
	data, err := admincontentservice.GetEnrollJoinDataURL(enrollID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-打卡管理
// @Summary 导出打卡数据
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_export [get]
func (h *AdminEnrollHandler) EnrollJoinDataExport(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	enrollID := c.Query("enrollId")
	startDay := c.Query("startTime")
	endDay := c.Query("endTime")
	if err := admincontentservice.EnsureEnrollVisibleForAdminContext(ctx, enrollID, admin.ID); err != nil {
		response.Fail(c, "导出失败")
		return
	}
	filename, err := admincontentservice.ExportEnrollJoinDataExcelContext(ctx, enrollID, startDay, endDay)
	if err != nil {
		response.Fail(c, "导出失败")
		return
	}
	fileURL := fmt.Sprintf("http://%s/uploads/%s", c.Request.Host(), filename)
	response.JSON(c, fileURL)
}

// @Tags PC端-打卡管理
// @Summary 删除打卡导出数据
// @Param enrollId formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_del [post]
func (h *AdminEnrollHandler) EnrollJoinDataDel(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	enrollID := c.PostForm("enrollId")
	if err := admincontentservice.EnsureEnrollVisibleForAdminContext(ctx, enrollID, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	err := admincontentservice.DeleteEnrollJoinDataExcel(enrollID)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
