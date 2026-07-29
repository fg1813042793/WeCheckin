package enroll

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	admincontentservice "wecheckin-backend/backend/internal/app/service/admincontent"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-打卡管理
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
	list, total, err := admincontentservice.GetAdminEnrollListContext(ctx, keyword, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, pagedListResponse{List: newEnrollListItems(list), Total: total})
}

// @Tags PC端-打卡管理
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
	err := admincontentservice.InsertEnrollContext(ctx, title, cateID, cateName, enrollForms, joinForms, "", addIP, publishDeptIds, 1, sort, dayCnt, start, end, string(objBytes), allowRepeat == "1" || allowRepeat == "true", dailyLimit, uint(deptID), admin.ID)
	if err != nil {
		response.Fail(c, "创建失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 获取打卡项目详情
// @Param id query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_detail [get]
func (h *AdminEnrollHandler) GetEnrollDetail(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.Query("id")
	data, err := admincontentservice.GetEnrollDetailForAdminContext(ctx, id, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-打卡管理
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
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
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
	err := admincontentservice.EditEnrollForAdminContext(ctx, id, title, cateID, cateName, enrollForms, joinForms, "", addIP, publishDeptIds, 1, sort, dayCnt, start, end, string(objBytes), allowRepeat == "1" || allowRepeat == "true", dailyLimit, uint(deptID), admin.ID)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}
