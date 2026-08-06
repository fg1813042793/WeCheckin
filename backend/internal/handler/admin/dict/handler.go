package dict

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	dictservice "wecheckin/backend/internal/service/admin/dict"
	"wecheckin/backend/pkg/response"
)

type AdminDictHandler struct{}

func NewAdminDictHandler() *AdminDictHandler { return &AdminDictHandler{} }

// @Tags PC端-字典管理
// @Summary 获取字典类型列表
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) GetDictTypes(ctx context.Context, c *app.RequestContext) {
	data, err := dictservice.GetTypesContext(ctx)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-字典管理
// @Summary 根据类型获取字典项
// @Param typeCode query string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) GetDictByType(ctx context.Context, c *app.RequestContext) {
	typeCode := c.Query("typeCode")
	data, err := dictservice.GetByTypeContext(ctx, typeCode)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-字典管理
// @Summary 新增字典项
// @Param typeCode formData string true "类型编码"
// @Param typeName formData string false "类型名称"
// @Param label formData string true "显示标签"
// @Param value formData string false "字典值"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) AddDictItem(ctx context.Context, c *app.RequestContext) {
	typeCode := c.PostForm("typeCode")
	typeName := c.PostForm("typeName")
	label := c.PostForm("label")
	value := c.PostForm("value")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))

	if typeCode == "" || label == "" {
		response.Fail(c, "类型编码和标签不能为空")
		return
	}
	if err := dictservice.AddItemContext(ctx, typeCode, typeName, label, value, remark, sort); err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 编辑字典项
// @Param id formData string true "字典项ID"
// @Param label formData string false "显示标签"
// @Param value formData string false "字典值"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) EditDictItem(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	label := c.PostForm("label")
	value := c.PostForm("value")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))

	if err := dictservice.EditItemContext(ctx, id, label, value, remark, sort); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 删除字典项
// @Param id formData string true "字典项ID"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) DelDictItem(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if err := dictservice.DeleteItemContext(ctx, id); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 清空指定类型的字典项
// @Param typeCode formData string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) DelDictByType(ctx context.Context, c *app.RequestContext) {
	typeCode := c.PostForm("typeCode")
	if err := dictservice.DeleteByTypeContext(ctx, typeCode); err != nil {
		response.Fail(c, "清空失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 编辑字典类型名称
// @Param oldTypeCode formData string true "原类型编码"
// @Param typeCode formData string true "新类型编码"
// @Param typeName formData string true "类型名称"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) EditDictTypeName(ctx context.Context, c *app.RequestContext) {
	oldTypeCode := c.PostForm("oldTypeCode")
	typeCode := c.PostForm("typeCode")
	typeName := c.PostForm("typeName")
	if typeCode == "" || typeName == "" {
		response.Fail(c, "参数不能为空")
		return
	}
	if err := dictservice.EditTypeNameContext(ctx, oldTypeCode, typeCode, typeName); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}
