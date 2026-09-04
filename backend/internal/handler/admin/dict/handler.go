package dict

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	dictservice "wecheckin/backend/internal/service/admin/dict"
	"wecheckin/backend/pkg/response"
)

type AdminDictHandler struct{}

func NewAdminDictHandler() *AdminDictHandler { return &AdminDictHandler{} }

// @Tags PC端-字典管理
// @Summary 获取全部字典类型
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
// @Summary 获取启用的公开字典类型
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) GetPublicDictTypes(ctx context.Context, c *app.RequestContext) {
	data, err := dictservice.GetActiveTypesContext(ctx)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-字典管理
// @Summary 根据类型获取全部字典项
// @Param typeCode query string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) GetDictByType(ctx context.Context, c *app.RequestContext) {
	data, err := dictservice.GetByTypeContext(ctx, c.Query("typeCode"))
	if err != nil {
		fail(c, err, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-字典管理
// @Summary 根据类型获取启用的公开字典项
// @Param typeCode query string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) GetPublicDictByType(ctx context.Context, c *app.RequestContext) {
	data, err := dictservice.GetActiveByTypeContext(ctx, c.Query("typeCode"))
	if err != nil {
		fail(c, err, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-字典管理
// @Summary 新增字典类型
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) AddDictType(ctx context.Context, c *app.RequestContext) {
	status, err := formStatus(c, 1)
	if err != nil {
		fail(c, err, "添加失败")
		return
	}
	err = dictservice.AddTypeContext(ctx, c.PostForm("typeCode"), c.PostForm("typeName"), c.PostForm("remark"), status)
	if err != nil {
		fail(c, err, "添加失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 编辑字典类型
// @Param typeCode path string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) EditDictType(ctx context.Context, c *app.RequestContext) {
	typeCode := strings.TrimSpace(c.Param("typeCode"))
	if bodyTypeCode := strings.TrimSpace(c.PostForm("typeCode")); bodyTypeCode != "" && bodyTypeCode != typeCode {
		fail(c, dictservice.ErrTypeCodeImmutable, "编辑失败")
		return
	}
	status, err := formStatus(c, 1)
	if err != nil {
		fail(c, err, "编辑失败")
		return
	}
	err = dictservice.EditTypeContext(ctx, typeCode, c.PostForm("typeName"), c.PostForm("remark"), status)
	if err != nil {
		fail(c, err, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 删除字典类型及其数据
// @Param typeCode path string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) DelDictType(ctx context.Context, c *app.RequestContext) {
	if err := dictservice.DeleteTypeContext(ctx, c.Param("typeCode")); err != nil {
		fail(c, err, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 新增字典项
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) AddDictItem(ctx context.Context, c *app.RequestContext) {
	status, err := formStatus(c, 1)
	if err != nil {
		fail(c, err, "添加失败")
		return
	}
	typeCode := c.PostForm("typeCode")
	typeName := c.PostForm("typeName")
	label := c.PostForm("label")
	value := c.PostForm("value")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))

	// Rolling deployments may still send the historical empty item used to create a type.
	if strings.TrimSpace(value) == "" && strings.TrimSpace(label) == strings.TrimSpace(typeName) {
		err = dictservice.AddItemContext(ctx, typeCode, typeName, label, value, remark, sort)
	} else {
		err = dictservice.AddItemWithStatusContext(ctx, typeCode, label, value, remark, sort, status)
	}
	if err != nil {
		fail(c, err, "添加失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 编辑字典项
// @Param id path int true "字典项ID"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) EditDictItem(ctx context.Context, c *app.RequestContext) {
	status, err := formStatus(c, 1)
	if err != nil {
		fail(c, err, "编辑失败")
		return
	}
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	id := c.Param("id")
	if id == "" {
		id = c.PostForm("id")
	}
	err = dictservice.EditItemWithStatusContext(ctx, id, c.PostForm("label"), c.PostForm("value"), c.PostForm("remark"), sort, status)
	if err != nil {
		fail(c, err, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 删除字典项
// @Param id path int true "字典项ID"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) DelDictItem(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		id = c.PostForm("id")
	}
	if err := dictservice.DeleteItemContext(ctx, id); err != nil {
		fail(c, err, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-字典管理
// @Summary 清空指定类型的字典项并保留类型
// @Param typeCode path string true "类型编码"
// @Success 200 {object} response.Resp
func (h *AdminDictHandler) DelDictByType(ctx context.Context, c *app.RequestContext) {
	if err := dictservice.ClearTypeItemsContext(ctx, c.Param("typeCode")); err != nil {
		fail(c, err, "清空失败")
		return
	}
	response.JSON(c, nil)
}

// EditDictTypeName preserves the historical handler entry while enforcing immutable type codes.
func (h *AdminDictHandler) EditDictTypeName(ctx context.Context, c *app.RequestContext) {
	oldTypeCode := c.Param("typeCode")
	if oldTypeCode == "" {
		oldTypeCode = c.PostForm("oldTypeCode")
	}
	typeCode := c.PostForm("typeCode")
	if typeCode == "" {
		typeCode = oldTypeCode
	}
	if err := dictservice.EditTypeNameContext(ctx, oldTypeCode, typeCode, c.PostForm("typeName")); err != nil {
		fail(c, err, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

func formStatus(c *app.RequestContext, fallback int) (int, error) {
	raw := strings.TrimSpace(c.PostForm("status"))
	if raw == "" {
		return fallback, nil
	}
	status, err := strconv.Atoi(raw)
	if err != nil {
		return 0, dictservice.ErrInvalidStatus
	}
	if status != 0 && status != 1 {
		return 0, dictservice.ErrInvalidStatus
	}
	return status, nil
}

func fail(c *app.RequestContext, err error, fallback string) {
	response.Fail(c, dictservice.ClientErrorMessage(err, fallback))
}
