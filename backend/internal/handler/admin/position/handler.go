package position

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	positionservice "wecheckin/backend/internal/service/admin/position"
	"wecheckin/backend/pkg/response"
)

type AdminPositionHandler struct{}

func NewAdminPositionHandler() *AdminPositionHandler { return &AdminPositionHandler{} }

func (h *AdminPositionHandler) GetPositionList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	data, err := positionservice.GetListContext(ctx, c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminPositionHandler) AddPosition(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	if err := positionservice.AddContext(ctx, name, c.ClientIP(), sort); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *AdminPositionHandler) EditPosition(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	name := c.PostForm("name")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	if err := positionservice.EditContext(ctx, uint(id), name, c.ClientIP(), sort, status); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *AdminPositionHandler) DelPosition(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := positionservice.DeleteContext(ctx, uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}
