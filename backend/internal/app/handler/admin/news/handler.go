package news

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	admincontentservice "wecheckin-backend/backend/internal/app/service/admincontent"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type AdminNewsHandler struct{}

func NewAdminNewsHandler() *AdminNewsHandler { return &AdminNewsHandler{} }

// @Tags PC端-通知公告
// @Summary 获取通知公告列表
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
	list, total, err := admincontentservice.GetAdminNewsListContext(ctx, keyword, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, pagedListResponse{List: list, Total: total})
}

// @Tags PC端-通知公告
// @Summary 新增通知公告
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

	err := admincontentservice.InsertNewsContext(ctx, title, desc, cateID, cateName, content, "", img, "", addIP, publishDeptIds, 1, order, uint(deptID), admin.ID)
	if err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 获取通知公告详情
// @Param id query string true "通知公告ID"
// @Success 200 {object} response.Resp
// @Router /admin/news_detail [get]
func (h *AdminNewsHandler) GetNewsDetail(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.Query("id")
	data, err := admincontentservice.GetNewsDetailForAdminContext(ctx, id, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-通知公告
// @Summary 编辑通知公告
// @Success 200 {object} response.Resp
// @Router /admin/news_edit [post]
func (h *AdminNewsHandler) EditNews(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
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

	err := admincontentservice.EditNewsForAdminContext(ctx, id, title, desc, cateID, cateName, content, "", addIP, publishDeptIds, 1, order, uint(deptID), admin.ID)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	if img != "" {
		_ = admincontentservice.UpdateNewsPicForAdminContext(ctx, id, img, admin.ID)
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 更新通知公告表单
// @Param id formData string true "通知公告ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /admin/news_update_forms [post]
func (h *AdminNewsHandler) UpdateNewsForms(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	err := admincontentservice.UpdateNewsFormsForAdminContext(ctx, id, forms, admin.ID)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 更新通知公告图片
// @Param id formData string true "通知公告ID"
// @Param pic formData string false "图片数据"
// @Success 200 {object} response.Resp
// @Router /admin/news_update_pic [post]
func (h *AdminNewsHandler) UpdateNewsPic(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	pic := c.PostForm("pic")
	err := admincontentservice.UpdateNewsPicForAdminContext(ctx, id, pic, admin.ID)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 更新通知公告内容
// @Param id formData string true "通知公告ID"
// @Param content formData string false "内容"
// @Success 200 {object} response.Resp
// @Router /admin/news_update_content [post]
func (h *AdminNewsHandler) UpdateNewsContent(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	content := c.PostForm("content")
	err := admincontentservice.UpdateNewsContentForAdminContext(ctx, id, content, admin.ID)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 删除通知公告
// @Param id formData string true "通知公告ID"
// @Success 200 {object} response.Resp
// @Router /admin/news_del [post]
func (h *AdminNewsHandler) DelNews(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	err := admincontentservice.DelNewsForAdminContext(ctx, id, admin.ID)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminNewsHandler) DelNewsList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := admincontentservice.DelNewsListForAdminContext(ctx, ids, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 通知公告排序
// @Param id formData string true "通知公告ID"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /admin/news_sort [post]
func (h *AdminNewsHandler) SortNews(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	sortStr := c.PostForm("sort")
	err := admincontentservice.SortNewsForAdminContext(ctx, id, sortStr, admin.ID)
	if err != nil {
		response.Fail(c, "排序失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 设置通知公告状态
// @Param id formData string true "通知公告ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/news_status [post]
func (h *AdminNewsHandler) StatusNews(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := admincontentservice.StatusNewsForAdminContext(ctx, id, status, admin.ID)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-通知公告
// @Summary 设置通知公告推荐
// @Param id formData string true "通知公告ID"
// @Param vouch formData int true "推荐(1=推荐 0=取消)"
// @Success 200 {object} response.Resp
// @Router /admin/news_vouch [post]
func (h *AdminNewsHandler) VouchNews(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	vouch, _ := strconv.Atoi(c.PostForm("vouch"))
	err := admincontentservice.VouchNewsForAdminContext(ctx, id, vouch, admin.ID)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
