package api

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type AdminExamHandler struct{}

func NewAdminExamHandler() *AdminExamHandler { return &AdminExamHandler{} }

func (h *AdminExamHandler) Detail(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.Fail(c, "无效的考试ID")
		return
	}
	var exam model.Exam
	if err := database.DB.Where("`exam_id` = ?", id).First(&exam).Error; err != nil {
		response.Fail(c, "考试不存在")
		return
	}
	response.JSON(c, exam)
}

func (h *AdminExamHandler) Save(_ context.Context, c *app.RequestContext) {
	type ExamSaveReq struct {
		ID          uint   `json:"id"`
		Title       string `json:"title"`
		Schema      string `json:"schema"`
		Settings    string `json:"settings"`
		StartTime   int64  `json:"startTime"`
		EndTime     int64  `json:"endTime"`
		Duration    int    `json:"duration"`
		MaxAttempts int    `json:"maxAttempts"`
		ShowScore   int    `json:"showScore"`
		Status      int    `json:"status"`
	}
	var req ExamSaveReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if req.Title == "" {
		response.Fail(c, "标题不能为空")
		return
	}
	now := time.Now().UnixMilli()
	if req.ID == 0 {
		exam := model.Exam{
			Title:       req.Title,
			Schema:      req.Schema,
			Settings:    req.Settings,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Duration:    req.Duration,
			MaxAttempts: req.MaxAttempts,
			ShowScore:   req.ShowScore,
			Status:      req.Status,
			AddTime:     now,
		}
		if exam.Schema == "" {
			exam.Schema = `{"version":"2.0","questions":[],"setting":{}}`
		}
		if exam.Settings == "" {
			exam.Settings = "{}"
		}
		if err := database.DB.Create(&exam).Error; err != nil {
			response.Fail(c, "创建失败: "+err.Error())
			return
		}
		response.JSON(c, exam)
	} else {
		updates := map[string]interface{}{
			"exam_title":        req.Title,
			"exam_schema":       req.Schema,
			"exam_settings":     req.Settings,
			"exam_start_time":   req.StartTime,
			"exam_end_time":     req.EndTime,
			"exam_duration":     req.Duration,
			"exam_max_attempts": req.MaxAttempts,
			"exam_show_score":   req.ShowScore,
			"exam_status":       req.Status,
			"exam_edit_time":    now,
		}
		if err := database.DB.Model(&model.Exam{}).Where("`exam_id` = ?", req.ID).Updates(updates).Error; err != nil {
			response.Fail(c, "更新失败: "+err.Error())
			return
		}
		response.JSON(c, map[string]interface{}{"id": req.ID})
	}
}

func (h *AdminExamHandler) List(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.Exam{})
	if kw := c.Query("keyword"); kw != "" {
		q = q.Where("`exam_title` LIKE ?", "%"+kw+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Exam
	q.Order("`exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

func (h *AdminExamHandler) Delete(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`exam_id` = ?", id).Delete(&model.Exam{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	database.DB.Where("`exam_r_exam_id` = ?", id).Delete(&model.ExamRecord{})
	response.JSON(c, nil)
}
