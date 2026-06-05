package admin

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminDictHandler struct{}

func NewAdminDictHandler() *AdminDictHandler { return &AdminDictHandler{} }

func (h *AdminDictHandler) GetDictTypes(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetDictTypes()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminDictHandler) GetDictByType(ctx context.Context, c *app.RequestContext) {
	typeCode := c.Query("typeCode")
	data, err := service.GetDictByType(typeCode)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

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
	if err := service.AddDictItem(typeCode, typeName, label, value, remark, sort); err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDictHandler) EditDictItem(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	label := c.PostForm("label")
	value := c.PostForm("value")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))

	if err := service.EditDictItem(id, label, value, remark, sort); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDictHandler) DelDictItem(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if err := service.DelDictItem(id); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDictHandler) DelDictByType(ctx context.Context, c *app.RequestContext) {
	typeCode := c.PostForm("typeCode")
	if err := service.DelDictByType(typeCode); err != nil {
		response.Fail(c, "清空失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDictHandler) EditDictTypeName(ctx context.Context, c *app.RequestContext) {
	oldTypeCode := c.PostForm("oldTypeCode")
	typeCode := c.PostForm("typeCode")
	typeName := c.PostForm("typeName")
	if typeCode == "" || typeName == "" {
		response.Fail(c, "参数不能为空")
		return
	}
	if err := service.EditDictTypeName(oldTypeCode, typeCode, typeName); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}
