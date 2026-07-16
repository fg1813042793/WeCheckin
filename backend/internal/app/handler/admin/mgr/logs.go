package mgr

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	adminlogservice "wecheckin-backend/backend/internal/app/service/adminlog"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-管理员管理
// @Summary 获取操作日志列表
// @Success 200 {object} response.Resp
// @Router /admin/log_list [get]
func (h *AdminMgrHandler) GetLogList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	keyword := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	list, total, err := adminlogservice.GetList(keyword, page, pageSize, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, pagedListResponse{List: list, Total: total})
}

// @Tags PC端-管理员管理
// @Summary 清除操作日志
// @Success 200 {object} response.Resp
// @Router /admin/log_clear [get]
func (h *AdminMgrHandler) ClearLog(ctx context.Context, c *app.RequestContext) {
	err := adminlogservice.Clear()
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
