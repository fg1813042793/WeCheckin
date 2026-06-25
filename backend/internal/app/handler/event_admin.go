package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/app/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminEventHandler struct{}

func NewAdminEventHandler() *AdminEventHandler { return &AdminEventHandler{} }

// @Tags PC端-赛事活动管理
// @Summary 获取活动列表(管理端)
// @Param keyword query string false "搜索关键词"
// @Param type query string false "活动类型"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param sort query string false "排序"
// @Success 200 {object} response.Resp
// @Router /admin/event_list [get]
func (h *AdminEventHandler) GetAdminEventList(ctx context.Context, c *app.RequestContext) {
	keyword := c.Query("keyword")
	typ := c.Query("type")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("pageSize"))
	sortStr := c.Query("sort")
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	list, total, err := service.GetAdminEventList(keyword, typ, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags PC端-赛事活动管理
// @Summary 获取活动详情(管理端)
// @Param id query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_detail [get]
func (h *AdminEventHandler) GetAdminEventDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	event, err := service.GetAdminEventDetail(id)
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
// @Router /admin/event_insert [post]
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

	// Parse role arrays from JSON
	organizers := parseUserArray(c.PostForm("organizers"))
	assistants := parseUserArray(c.PostForm("assistants"))
	referees := parseUserArray(c.PostForm("referees"))

	err := service.InsertEvent(title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds,
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
// @Router /admin/event_edit [post]
func (h *AdminEventHandler) EditEvent(ctx context.Context, c *app.RequestContext) {
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

	err := service.EditEvent(id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds,
		typ, status, order, regStart, regEnd, eventStart, eventEnd, obj,
		uint(deptID), organizers, assistants, referees)
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
// @Router /admin/event_del [post]
func (h *AdminEventHandler) DelEvent(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelEvent(id); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 批量删除活动
// @Param ids formData string true "活动ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/event_dels [post]
func (h *AdminEventHandler) DelEvents(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelEvents(ids); err != nil {
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
// @Router /admin/event_status [post]
func (h *AdminEventHandler) StatusEvent(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.StatusEvent(id, status); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 获取活动参与成员列表
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_participant_list [get]
func (h *AdminEventHandler) GetEventParticipantList(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := service.GetEventParticipantList(eventID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list})
}

// @Tags PC端-赛事活动管理
// @Summary 删除活动参与成员
// @Param id formData string true "参与记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_participant_del [post]
func (h *AdminEventHandler) DelEventParticipant(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelEventParticipant(id); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动参与成员信息
// @Param id formData string true "参与记录ID"
// @Param forms formData string false "表单数据(JSON)"
// @Success 200 {object} response.Resp
// @Router /admin/event_participant_edit [post]
func (h *AdminEventHandler) EditEventParticipant(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.EditEventParticipant(id, forms); err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 批量删除活动参与成员
// @Param ids formData string true "参与记录ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/event_participant_dels [post]
func (h *AdminEventHandler) DelEventParticipants(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelEventParticipants(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 发布活动动态(管理端)
// @Param eventId formData string true "活动ID"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
// @Router /admin/event_dynamic_add [post]
func (h *AdminEventHandler) PostEventDynamic(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("eventId")
	title := c.PostForm("title")
	content := c.PostForm("content")
	images := c.PostForm("images")
	videos := c.PostForm("videos")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	userID := "admin_" + admin.Name
	addIP := c.ClientIP()
	if err := service.PostEventDynamic(eventID, userID, title, content, images, videos, addIP); err != nil {
		response.Fail(c, "发布失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 获取活动动态列表(管理端)
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_dynamics [get]
func (h *AdminEventHandler) GetEventDynamics(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := service.GetEventDynamics(eventID, 1, 100)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动动态
// @Param id formData string true "动态ID"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
// @Router /admin/event_dynamic_edit [post]
func (h *AdminEventHandler) EditEventDynamic(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	title := c.PostForm("title")
	content := c.PostForm("content")
	images := c.PostForm("images")
	videos := c.PostForm("videos")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	editIP := c.ClientIP()
	if err := service.EditEventDynamic(id, title, content, images, videos, editIP); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 删除活动动态
// @Param id formData string true "动态ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_dynamic_del [post]
func (h *AdminEventHandler) DelEventDynamic(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelEventDynamic(id); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 批量删除活动动态
// @Param ids formData string true "动态ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/event_dynamic_dels [post]
func (h *AdminEventHandler) DelEventDynamics(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelEventDynamics(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 获取活动评分列表(管理端)
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_scores [get]
func (h *AdminEventHandler) GetEventScores(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := service.GetEventScores(eventID, 1, 100)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动评分
// @Param id formData string false "评分记录ID(为空时新增)"
// @Param score formData string true "评分"
// @Param eventId formData string false "活动ID(新增时必填)"
// @Param participantId formData string false "参赛者ID(新增时必填)"
// @Success 200 {object} response.Resp
// @Router /admin/event_score_edit [post]
func (h *AdminEventHandler) EditEventScore(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	score := c.PostForm("score")
	if id == "" {
		// Create new
		eventID := c.PostForm("eventId")
		participantID := c.PostForm("participantId")
		if eventID == "" || participantID == "" {
			response.Fail(c, "参数错误")
			return
		}
		if err := service.SaveEventScore(eventID, participantID, score, "admin"); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	} else {
		if err := service.AdminEditEventScore(id, score); err != nil {
			response.Fail(c, "编辑失败")
			return
		}
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 获取部门用户列表
// @Param deptIds query string true "部门ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/dept_users [get]
func (h *AdminEventHandler) GetDeptUsers(ctx context.Context, c *app.RequestContext) {
	deptIDsStr := c.Query("deptIds")
	if deptIDsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	var deptIDs []uint
	for _, s := range strings.Split(deptIDsStr, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(s))
		if err == nil && id > 0 {
			deptIDs = append(deptIDs, uint(id))
		}
	}
	users, err := service.GetDeptUsers(deptIDs)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": users})
}

// @Tags PC端-赛事活动管理
// @Summary 设置活动推荐
// @Param id formData string true "活动ID"
// @Param vouch formData int true "推荐(1=推荐 0=取消)"
// @Success 200 {object} response.Resp
// @Router /admin/event_vouch [post]
func (h *AdminEventHandler) VouchEvent(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	vouch, _ := strconv.Atoi(c.PostForm("vouch"))
	err := service.VouchEvent(id, vouch)
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
// @Router /admin/event_top [post]
func (h *AdminEventHandler) TopEvent(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	top, _ := strconv.Atoi(c.PostForm("top"))
	err := service.TopEvent(id, top)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

func parseUserArray(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	// Try JSON array first
	if err := json.Unmarshal([]byte(s), &arr); err == nil {
		return arr
	}
	// Fall back to comma-separated
	return strings.Split(s, ",")
}
