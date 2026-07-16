package department

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	departmentservice "wecheckin-backend/backend/internal/app/service/department"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type AdminDeptHandler struct{}

func NewAdminDeptHandler() *AdminDeptHandler { return &AdminDeptHandler{} }

// @Tags PC端-部门管理
// @Summary 获取部门树
// @Success 200 {object} response.Resp
// @Router /admin/dept/tree [get]
func (h *AdminDeptHandler) GetDeptTree(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	data, err := departmentservice.GetTree(admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-部门管理
// @Summary 新增部门
// @Param name formData string true "部门名称"
// @Param parentId formData int false "父部门ID"
// @Param sort formData int false "排序"
// @Success 200 {object} response.Resp
// @Router /admin/dept/add [post]
func (h *AdminDeptHandler) AddDept(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	parentID, _ := strconv.Atoi(c.PostForm("parentId"))
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	if name == "" {
		response.Fail(c, "部门名称不能为空")
		return
	}
	if err := departmentservice.Add(name, uint(parentID), sort); err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-部门管理
// @Summary 编辑部门
// @Param id formData string true "部门ID"
// @Param name formData string true "部门名称"
// @Param parentId formData int false "父部门ID"
// @Param sort formData int false "排序"
// @Param status formData int false "状态(1=启用 0=禁用)"
// @Success 200 {object} response.Resp
// @Router /admin/dept/edit [post]
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
	if err := departmentservice.Edit(uint(id), name, uint(parentID), sort, status); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-部门管理
// @Summary 删除部门
// @Param id formData string true "部门ID"
// @Success 200 {object} response.Resp
// @Router /admin/dept/del [post]
func (h *AdminDeptHandler) DelDept(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := departmentservice.Delete(uint(id)); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
