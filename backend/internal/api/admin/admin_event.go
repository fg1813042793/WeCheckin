package admin

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminEventHandler struct{}

func NewAdminEventHandler() *AdminEventHandler { return &AdminEventHandler{} }

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
