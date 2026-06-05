package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminEnrollHandler struct{}

func NewAdminEnrollHandler() *AdminEnrollHandler { return &AdminEnrollHandler{} }

// @Tags 打卡管理
// @Summary 获取打卡项目列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_list [get]
func (h *AdminEnrollHandler) GetAdminEnrollList(ctx context.Context, c *app.RequestContext) {
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
	list, total, err := service.GetAdminEnrollList(keyword, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 打卡管理
// @Summary 新增打卡项目
// @Success 200 {object} response.Resp
// @Router /admin/enroll_insert [post]
func (h *AdminEnrollHandler) InsertEnroll(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	title := c.PostForm("title")
	cateID := c.PostForm("cateId")
	cateName := c.PostForm("cateName")
	startTime := c.PostForm("startTime")
	endTime := c.PostForm("endTime")
	sortStr := c.PostForm("sort")
	cover := c.PostForm("cover")
	desc := c.PostForm("desc")
	addIP := c.ClientIP()
	joinForms := c.PostForm("joinForms")
	enrollForms := c.PostForm("enrollForms")
	allowRepeat := c.PostForm("allowRepeat")
	dailyLimitStr := c.PostForm("dailyLimit")

	sort, _ := strconv.Atoi(sortStr)
	if sort <= 0 {
		sort = 9999
	}

	objMap := map[string]interface{}{}
	if cover != "" {
		objMap["cover"] = []string{cover}
	}
	if desc != "" {
		objMap["desc"] = desc
	}
	objBytes, _ := json.Marshal(objMap)

	var start, end int64
	if startTime != "" {
		t, err := time.Parse("2006-01-02", startTime)
		if err == nil {
			start = t.UnixMilli()
		}
	}
	if endTime != "" {
		t, err := time.Parse("2006-01-02", endTime)
		if err == nil {
			end = t.UnixMilli()
		}
	}

	dayCnt := 0
	if start > 0 && end > start {
		dayCnt = int((end - start) / (24 * 60 * 60 * 1000))
	}

	dailyLimit, _ := strconv.Atoi(dailyLimitStr)
	if dailyLimit <= 0 {
		dailyLimit = 1
	}

	deptID, _ := strconv.ParseUint(c.PostForm("deptId"), 10, 64)
	publishDeptIds := c.PostForm("publishDeptIds")
	err := service.InsertEnroll(title, cateID, cateName, enrollForms, joinForms, "", addIP, publishDeptIds, 1, sort, dayCnt, start, end, string(objBytes), allowRepeat == "1" || allowRepeat == "true", dailyLimit, uint(deptID), admin.ID)
	if err != nil {
		response.Fail(c, "创建失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 获取打卡项目详情
// @Param id query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_detail [get]
func (h *AdminEnrollHandler) GetEnrollDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.GetEnrollDetail(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 打卡管理
// @Summary 编辑打卡项目
// @Param id formData string true "项目ID"
// @Param title formData string false "标题"
// @Param cateId formData string false "分类ID"
// @Param cateName formData string false "分类名称"
// @Param startTime formData string false "开始时间"
// @Param endTime formData string false "结束时间"
// @Param sort formData string false "排序"
// @Param cover formData string false "封面图URL"
// @Param desc formData string false "描述"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_edit [post]
func (h *AdminEnrollHandler) EditEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	title := c.PostForm("title")
	cateID := c.PostForm("cateId")
	cateName := c.PostForm("cateName")
	startTime := c.PostForm("startTime")
	endTime := c.PostForm("endTime")
	sortStr := c.PostForm("sort")
	cover := c.PostForm("cover")
	desc := c.PostForm("desc")
	allowRepeat := c.PostForm("allowRepeat")
	dailyLimitStr := c.PostForm("dailyLimit")
	joinForms := c.PostForm("joinForms")
	enrollForms := c.PostForm("enrollForms")
	addIP := c.ClientIP()

	sort, _ := strconv.Atoi(sortStr)
	if sort <= 0 {
		sort = 9999
	}

	dailyLimit, _ := strconv.Atoi(dailyLimitStr)
	if dailyLimit <= 0 {
		dailyLimit = 1
	}

	objMap := map[string]interface{}{}
	if cover != "" {
		objMap["cover"] = []string{cover}
	}
	if desc != "" {
		objMap["desc"] = desc
	}
	objBytes, _ := json.Marshal(objMap)

	var start, end int64
	if startTime != "" {
		t, err := time.Parse("2006-01-02", startTime)
		if err == nil {
			start = t.UnixMilli()
		}
	}
	if endTime != "" {
		t, err := time.Parse("2006-01-02", endTime)
		if err == nil {
			end = t.UnixMilli()
		}
	}

	dayCnt := 0
	if start > 0 && end > start {
		dayCnt = int((end - start) / (24 * 60 * 60 * 1000))
	}

	deptID, _ := strconv.ParseUint(c.PostForm("deptId"), 10, 64)
	publishDeptIds := c.PostForm("publishDeptIds")
	err := service.EditEnroll(id, title, cateID, cateName, enrollForms, joinForms, "", addIP, publishDeptIds, 1, sort, dayCnt, start, end, string(objBytes), allowRepeat == "1" || allowRepeat == "true", dailyLimit, uint(deptID))
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 更新打卡表单
// @Param id formData string true "项目ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_update_forms [post]
func (h *AdminEnrollHandler) UpdateEnrollForms(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	err := service.UpdateEnrollForms(id, forms)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 清除打卡全部数据
// @Param id formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_clear [post]
func (h *AdminEnrollHandler) ClearEnrollAll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.ClearEnrollAll(id)
	if err != nil {
		response.Fail(c, "清除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 删除打卡项目
// @Param id formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_del [post]
func (h *AdminEnrollHandler) DelEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelEnroll(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) DelEnrolls(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelEnrolls(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 打卡项目排序
// @Param id formData string true "项目ID"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_sort [post]
func (h *AdminEnrollHandler) SortEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	sortStr := c.PostForm("sort")
	err := service.SortEnroll(id, sortStr)
	if err != nil {
		response.Fail(c, "排序失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 设置打卡推荐
// @Param id formData string true "项目ID"
// @Param vouch formData string true "推荐值"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_vouch [post]
func (h *AdminEnrollHandler) VouchEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	vouch, _ := strconv.Atoi(c.PostForm("vouch"))
	err := service.VouchEnroll(id, vouch)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 设置打卡状态
// @Param id formData string true "项目ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_status [post]
func (h *AdminEnrollHandler) StatusEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := service.StatusEnroll(id, status)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 获取打卡记录列表
// @Param enrollId query string true "项目ID"
// @Tags 打卡管理
// @Summary 获取参与用户列表(含报名表单数据)
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_user_list [get]
func (h *AdminEnrollHandler) GetEnrollUserList(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	keyword := c.Query("keyword")
	list, err := service.GetEnrollUserList(enrollID, keyword)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

func (h *AdminEnrollHandler) GetEnrollStats(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	startDay := c.Query("startTime")
	endDay := c.Query("endTime")
	if enrollID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := service.GetEnrollStats(enrollID, startDay, endDay)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_list [get]
func (h *AdminEnrollHandler) GetEnrollJoinList(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("pageSize"))
	if size == 0 {
		size, _ = strconv.Atoi(c.Query("size"))
	}
	list, total, err := service.GetEnrollJoinList(enrollID, keyword, page, size)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 打卡管理
// @Summary 从打卡项目中移除用户（删除用户及所有打卡记录）
// @Param enrollId formData string true "项目ID"
// @Param userId formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_remove_user [post]
func (h *AdminEnrollHandler) RemoveEnrollUser(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	userID := c.PostForm("userId")
	if enrollID == "" || userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	err := service.RemoveEnrollUser(enrollID, userID)
	if err != nil {
		response.Fail(c, "移除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) RemoveEnrollUsers(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	userIDsStr := c.PostForm("userIds")
	if enrollID == "" || userIDsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	userIDs := strings.Split(userIDsStr, ",")
	if err := service.RemoveEnrollUsers(enrollID, userIDs); err != nil {
		response.Fail(c, "移除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 删除打卡记录
// @Param id formData string true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_del [post]
func (h *AdminEnrollHandler) DelEnrollJoin(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("enrollJoinId")
	if id == "" {
		id = c.PostForm("id")
	}
	err := service.DelEnrollJoin(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) DelEnrollJoins(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelEnrollJoins(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 获取打卡数据导出链接
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_get [get]
func (h *AdminEnrollHandler) EnrollJoinDataGet(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	data, err := service.GetEnrollJoinDataURL(enrollID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 打卡管理
// @Summary 导出打卡数据
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_export [get]
func (h *AdminEnrollHandler) EnrollJoinDataExport(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	startDay := c.Query("startTime")
	endDay := c.Query("endTime")
	filename, err := service.ExportEnrollJoinDataExcel(enrollID, startDay, endDay)
	if err != nil {
		response.Fail(c, "导出失败")
		return
	}
	fileURL := fmt.Sprintf("http://%s/uploads/%s", c.Request.Host(), filename)
	response.JSON(c, fileURL)
}

// @Tags 打卡管理
// @Summary 删除打卡导出数据
// @Param enrollId formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_del [post]
func (h *AdminEnrollHandler) EnrollJoinDataDel(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	err := service.DeleteEnrollJoinDataExcel(enrollID)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
