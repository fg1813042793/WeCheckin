package client

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type EnrollHandler struct{}

func NewEnrollHandler() *EnrollHandler { return &EnrollHandler{} }

// @Tags 报名
// @Summary 获取报名列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /enroll/list [get]
func (h *EnrollHandler) GetEnrollList(ctx context.Context, c *app.RequestContext) {
	page, pageSize := 1, 10
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.Query("pageSize")); err == nil && ps > 0 {
		pageSize = ps
	}
	userID := c.Query("user_id")
	keyword := c.Query("keyword")
	data, err := service.GetEnrollList(page, pageSize, userID, keyword)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 查看报名详情
// @Param id query string true "报名ID"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/view [get]
func (h *EnrollHandler) ViewEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	userID := c.Query("user_id")
	data, err := service.ViewEnroll(id, userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 按日获取报名打卡记录
// @Param enroll_id query string true "报名ID"
// @Param day query string true "日期"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/join_day [get]
func (h *EnrollHandler) GetEnrollJoinByDay(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("id")
	if enrollID == "" {
		enrollID = c.Query("enroll_id")
	}
	day := c.Query("day")
	if day == "" {
		day = time.Now().Format("2006-01-02")
	}
	data, err := service.GetEnrollJoinByDay(enrollID, day)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 用户报名打卡
// @Param enroll_id formData string true "报名ID"
// @Param day formData string true "日期"
// @Param user_id formData string false "用户ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /enroll/join [post]
func (h *EnrollHandler) EnrollJoin(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enroll_id")
	if enrollID == "" {
		enrollID = c.PostForm("enrollId")
	}
	day := c.PostForm("day")
	if day == "" {
		day = time.Now().Format("2006-01-02")
	}
	userID := c.PostForm("user_id")
	if userID == "" {
		userID = c.PostForm("token")
	}
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	err := service.EnrollJoin(enrollID, userID, day, forms, addIP, 1)
	if err != nil {
		log.Printf("[EnrollJoin] 失败: enrollID=%s userID=%s day=%s err=%s", enrollID, userID, day, err.Error())
		response.Fail(c, err.Error())
		return
	}
	log.Printf("[EnrollJoin] 成功: enrollID=%s userID=%s day=%s", enrollID, userID, day)
	response.JSON(c, nil)
}

// @Tags 报名
// @Summary 用户报名(提交报名表单)
// @Param enroll_id formData string true "项目ID"
// @Param user_id formData string true "用户ID"
// @Param forms formData string false "报名表单数据JSON"
// @Success 200 {object} response.Resp
// @Router /enroll/enroll_submit [post]
func (h *EnrollHandler) EnrollUserSubmit(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enroll_id")
	if enrollID == "" {
		enrollID = c.PostForm("enrollId")
	}
	userID := c.PostForm("user_id")
	if userID == "" {
		userID = c.PostForm("token")
	}
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	err := service.EnrollUserSubmit(enrollID, userID, forms, addIP)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags 报名
// @Summary 获取我的打卡记录
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/my_join_list [get]
func (h *EnrollHandler) GetMyEnrollJoinList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = c.Query("id")
	}
	enrollID := c.Query("enrollId")
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	data, total, err := service.GetMyEnrollJoinList(userID, enrollID, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": data, "total": total})
}

// @Tags 报名
// @Summary 获取我的打卡记录
// @Param user_id query string false "用户ID"
// @Param page query string false "页码"
// @Param pageSize query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /enroll/my_records [get]
func (h *EnrollHandler) GetMyJoinRecords(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	list, total, err := service.GetMyJoinRecords(userID, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 报名
// @Summary 获取我的日历打卡数据
// @Param user_id query string false "用户ID"
// @Param month query string false "年月 (2026-06)"
// @Success 200 {object} response.Resp
// @Router /enroll/my_calendar [get]
func (h *EnrollHandler) GetMyCalendar(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	month := c.Query("month")
	data, err := service.GetMyCalendarDays(userID, month)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 获取我指定日期的打卡记录
// @Param user_id query string false "用户ID"
// @Param day query string false "日期 (2026-06-01)"
// @Success 200 {object} response.Resp
// @Router /enroll/my_day_records [get]
func (h *EnrollHandler) GetMyDayRecords(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	day := c.Query("day")
	data, err := service.GetMyDayRecords(userID, day)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 获取我的报名用户列表
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/my_user_list [get]
func (h *EnrollHandler) GetMyEnrollUserList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	data, err := service.GetMyEnrollUserList(userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}
