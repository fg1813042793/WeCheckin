package admin

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminDeptHandler struct{}

func NewAdminDeptHandler() *AdminDeptHandler { return &AdminDeptHandler{} }

func (h *AdminDeptHandler) GetDeptTree(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	data, err := service.GetDeptTree(admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminDeptHandler) AddDept(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	parentID, _ := strconv.Atoi(c.PostForm("parentId"))
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	if name == "" {
		response.Fail(c, "部门名称不能为空")
		return
	}
	if err := service.AddDept(name, uint(parentID), sort); err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDeptHandler) EditDept(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	name := c.PostForm("name")
	parentID, _ := strconv.Atoi(c.PostForm("parentId"))
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	if name == "" || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.EditDept(uint(id), name, uint(parentID), sort, status); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDeptHandler) DelDept(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelDept(uint(id)); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
