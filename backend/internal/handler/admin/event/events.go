package event

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	eventservice "wecheckin/backend/internal/service/client/event"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-赛事活动管理
// @Summary 获取活动列表(管理端)
// @Param keyword query string false "搜索关键词"
// @Param type query string false "活动类型"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param sort query string false "排序"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) GetAdminEventList(ctx context.Context, c *app.RequestContext) {
	keyword := c.Query("keyword")
	typ := c.Query("type")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("pageSize"))
	sortStr := c.Query("sort")
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	list, total, err := eventservice.GetAdminEventListContext(ctx, keyword, typ, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, pagedListResponse{List: newEventListItems(list), Total: total})
}

// @Tags PC端-赛事活动管理
// @Summary 获取活动详情(管理端)
// @Param id query string true "活动ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) GetAdminEventDetail(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.Query("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	event, err := eventservice.GetAdminEventDetailForAdminContext(ctx, id, admin.ID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	response.JSON(c, event)
}

// @Tags PC端-赛事活动管理
// @Summary 新增活动
// @Param title formData string true "活动标题"
// @Param type formData int false "活动类型(1=活动 2=赛事)"
// @Param cateId formData string false "分类ID"
// @Param cateName formData string false "分类名称"
// @Param status formData int false "状态"
// @Param order formData int false "排序"
// @Param regStart formData int false "报名开始时间(时间戳)"
// @Param regEnd formData int false "报名结束时间(时间戳)"
// @Param eventStart formData int false "活动开始时间(时间戳)"
// @Param eventEnd formData int false "活动结束时间(时间戳)"
// @Param forms formData string false "报名表单(JSON)"
// @Param scoreFields formData string false "评分字段(JSON)"
// @Param qr formData string false "二维码URL"
// @Param obj formData string false "扩展对象(JSON)"
// @Param publishDeptIds formData string false "发布部门IDs(逗号分隔)"
// @Param deptId formData int false "所属部门ID"
// @Param organizers formData string false "组织者列表(JSON或逗号分隔)"
// @Param assistants formData string false "协助者列表(JSON或逗号分隔)"
// @Param referees formData string false "裁判列表(JSON或逗号分隔)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) InsertEvent(ctx context.Context, c *app.RequestContext) {
	title := c.PostForm("title")
	typ, _ := strconv.Atoi(c.PostForm("type"))
	cateID := c.PostForm("cateId")
	cateName := c.PostForm("cateName")
	status, _ := strconv.Atoi(c.PostForm("status"))
	order, _ := strconv.Atoi(c.PostForm("order"))
	regStart, _ := strconv.ParseInt(c.PostForm("regStart"), 10, 64)
	regEnd, _ := strconv.ParseInt(c.PostForm("regEnd"), 10, 64)
	eventStart, _ := strconv.ParseInt(c.PostForm("eventStart"), 10, 64)
	eventEnd, _ := strconv.ParseInt(c.PostForm("eventEnd"), 10, 64)
	forms := c.PostForm("forms")
	scoreFields := c.PostForm("scoreFields")
	qr := c.PostForm("qr")
	obj := c.PostForm("obj")
	publishDeptIds := c.PostForm("publishDeptIds")
	deptID, _ := strconv.Atoi(c.PostForm("deptId"))
	addIP := c.ClientIP()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)

	organizers := parseUserArray(c.PostForm("organizers"))
	assistants := parseUserArray(c.PostForm("assistants"))
	referees := parseUserArray(c.PostForm("referees"))

	err := eventservice.InsertEventContext(ctx, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds,
		typ, status, order, regStart, regEnd, eventStart, eventEnd, obj,
		uint(deptID), admin.ID, organizers, assistants, referees)
	if err != nil {
		response.Fail(c, "创建失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动
// @Param id formData string true "活动ID"
// @Param title formData string false "活动标题"
// @Param type formData int false "活动类型"
// @Param cateId formData string false "分类ID"
// @Param cateName formData string false "分类名称"
// @Param status formData int false "状态"
// @Param order formData int false "排序"
// @Param regStart formData int false "报名开始时间(时间戳)"
// @Param regEnd formData int false "报名结束时间(时间戳)"
// @Param eventStart formData int false "活动开始时间(时间戳)"
// @Param eventEnd formData int false "活动结束时间(时间戳)"
// @Param forms formData string false "报名表单(JSON)"
// @Param scoreFields formData string false "评分字段(JSON)"
// @Param qr formData string false "二维码URL"
// @Param obj formData string false "扩展对象(JSON)"
// @Param publishDeptIds formData string false "发布部门IDs(逗号分隔)"
// @Param deptId formData int false "所属部门ID"
// @Param organizers formData string false "组织者列表(JSON或逗号分隔)"
// @Param assistants formData string false "协助者列表(JSON或逗号分隔)"
// @Param referees formData string false "裁判列表(JSON或逗号分隔)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) EditEvent(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	title := c.PostForm("title")
	typ, _ := strconv.Atoi(c.PostForm("type"))
	cateID := c.PostForm("cateId")
	cateName := c.PostForm("cateName")
	status, _ := strconv.Atoi(c.PostForm("status"))
	order, _ := strconv.Atoi(c.PostForm("order"))
	regStart, _ := strconv.ParseInt(c.PostForm("regStart"), 10, 64)
	regEnd, _ := strconv.ParseInt(c.PostForm("regEnd"), 10, 64)
	eventStart, _ := strconv.ParseInt(c.PostForm("eventStart"), 10, 64)
	eventEnd, _ := strconv.ParseInt(c.PostForm("eventEnd"), 10, 64)
	forms := c.PostForm("forms")
	scoreFields := c.PostForm("scoreFields")
	qr := c.PostForm("qr")
	obj := c.PostForm("obj")
	publishDeptIds := c.PostForm("publishDeptIds")
	deptID, _ := strconv.Atoi(c.PostForm("deptId"))
	addIP := c.ClientIP()

	organizers := parseUserArray(c.PostForm("organizers"))
	assistants := parseUserArray(c.PostForm("assistants"))
	referees := parseUserArray(c.PostForm("referees"))

	err := eventservice.EditEventForAdminContext(ctx, id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds,
		typ, status, order, regStart, regEnd, eventStart, eventEnd, obj,
		uint(deptID), admin.ID, organizers, assistants, referees)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 删除活动
// @Param id formData string true "活动ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) DelEvent(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := eventservice.DelEventForAdminContext(ctx, id, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 批量删除活动
// @Param ids formData string true "活动ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) DelEvents(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := eventservice.DelEventsForAdminContext(ctx, ids, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 设置活动状态
// @Param id formData string true "活动ID"
// @Param status formData int true "状态(1=启用 0=禁用)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) StatusEvent(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := eventservice.StatusEventForAdminContext(ctx, id, status, admin.ID); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 设置活动推荐
// @Param id formData string true "活动ID"
// @Param vouch formData int true "推荐(1=推荐 0=取消)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) VouchEvent(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	vouch, _ := strconv.Atoi(c.PostForm("vouch"))
	err := eventservice.VouchEventForAdminContext(ctx, id, vouch, admin.ID)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 置顶活动
// @Param id formData string true "活动ID"
// @Param top formData int true "置顶(1=置顶 0=取消)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) TopEvent(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	top, _ := strconv.Atoi(c.PostForm("top"))
	err := eventservice.TopEventForAdminContext(ctx, id, top, admin.ID)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
