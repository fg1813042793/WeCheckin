package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminNewsHandler struct{}

func NewAdminNewsHandler() *AdminNewsHandler { return &AdminNewsHandler{} }

// @Tags 新闻管理
// @Summary 获取新闻列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/news_list [get]
func (h *AdminNewsHandler) GetAdminNewsList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	page, _ := strconv.Atoi(c.Query("page"))
	sizeStr := c.Query("pageSize")
	if sizeStr == "" {
		sizeStr = c.Query("size")
	}
	size, _ := strconv.Atoi(sizeStr)
	keyword := c.Query("keyword")
	sortStr := c.Query("sort")
	list, total, err := service.GetAdminNewsList(keyword, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 新闻管理
// @Summary 新增新闻
// @Success 200 {object} response.Resp
// @Router /admin/news_insert [post]
func (h *AdminNewsHandler) InsertNews(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	title := c.PostForm("title")
	desc := c.PostForm("desc")
	cateID := c.PostForm("cateId")
	cateName := c.PostForm("cateName")
	content := c.PostForm("content")
	img := c.PostForm("img")
	orderStr := c.PostForm("order")
	if orderStr == "" {
		orderStr = c.PostForm("sortOrder")
	}
	addIP := c.ClientIP()
	order, _ := strconv.Atoi(orderStr)
	deptID, _ := strconv.ParseUint(c.PostForm("deptId"), 10, 64)
	publishDeptIds := c.PostForm("publishDeptIds")

	err := service.InsertNews(title, desc, cateID, cateName, content, "", img, "", addIP, publishDeptIds, 1, order, uint(deptID), admin.ID)
	if err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 获取新闻详情
// @Param id query string true "新闻ID"
// @Success 200 {object} response.Resp
// @Router /admin/news_detail [get]
func (h *AdminNewsHandler) GetNewsDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.GetNewsDetail(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 新闻管理
// @Summary 编辑新闻
// @Success 200 {object} response.Resp
// @Router /admin/news_edit [post]
func (h *AdminNewsHandler) EditNews(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	title := c.PostForm("title")
	desc := c.PostForm("desc")
	cateID := c.PostForm("cateId")
	cateName := c.PostForm("cateName")
	content := c.PostForm("content")
	img := c.PostForm("img")
	orderStr := c.PostForm("order")
	if orderStr == "" {
		orderStr = c.PostForm("sortOrder")
	}
	addIP := c.ClientIP()
	order, _ := strconv.Atoi(orderStr)
	deptID, _ := strconv.ParseUint(c.PostForm("deptId"), 10, 64)
	publishDeptIds := c.PostForm("publishDeptIds")

	err := service.EditNews(id, title, desc, cateID, cateName, content, "", addIP, publishDeptIds, 1, order, uint(deptID))
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	if img != "" {
		service.UpdateNewsPic(id, img)
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 更新新闻表单
// @Param id formData string true "新闻ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /admin/news_update_forms [post]
func (h *AdminNewsHandler) UpdateNewsForms(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	err := service.UpdateNewsForms(id, forms)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 更新新闻图片
// @Param id formData string true "新闻ID"
// @Param pic formData string false "图片数据"
// @Success 200 {object} response.Resp
// @Router /admin/news_update_pic [post]
func (h *AdminNewsHandler) UpdateNewsPic(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	pic := c.PostForm("pic")
	err := service.UpdateNewsPic(id, pic)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 更新新闻内容
// @Param id formData string true "新闻ID"
// @Param content formData string false "内容"
// @Success 200 {object} response.Resp
// @Router /admin/news_update_content [post]
func (h *AdminNewsHandler) UpdateNewsContent(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	content := c.PostForm("content")
	err := service.UpdateNewsContent(id, content)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 删除新闻
// @Param id formData string true "新闻ID"
// @Success 200 {object} response.Resp
// @Router /admin/news_del [post]
func (h *AdminNewsHandler) DelNews(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelNews(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminNewsHandler) DelNewsList(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelNewsList(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 新闻排序
// @Param id formData string true "新闻ID"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /admin/news_sort [post]
func (h *AdminNewsHandler) SortNews(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	sortStr := c.PostForm("sort")
	err := service.SortNews(id, sortStr)
	if err != nil {
		response.Fail(c, "排序失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 新闻管理
// @Summary 设置新闻状态
// @Param id formData string true "新闻ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/news_status [post]
func (h *AdminNewsHandler) StatusNews(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := service.StatusNews(id, status)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
