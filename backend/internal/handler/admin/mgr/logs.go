package mgr

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	adminlogservice "wecheckin/backend/internal/service/admin/adminlog"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-管理员管理
// @Summary 获取操作日志列表
// @Success 200 {object} response.Resp
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
func (h *AdminMgrHandler) ClearLog(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr != "" {
		var ids []uint
		for _, s := range strings.Split(idsStr, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(s))
			if err == nil && id > 0 {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) == 0 {
			response.Fail(c, "请选择日志")
			return
		}
		if err := adminlogservice.DeleteIDsContext(ctx, ids); err != nil {
			response.Fail(c, "操作失败")
			return
		}
		response.JSON(c, nil)
		return
	}

	err := adminlogservice.Clear()
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
