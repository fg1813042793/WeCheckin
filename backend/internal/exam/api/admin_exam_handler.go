package api

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
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
	var respCnt int64
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", id).Count(&respCnt)
	var rawSchema string
	database.DB.Model(&model.Exam{}).Select("exam_schema").Where("`exam_id` = ?", id).Scan(&rawSchema)
	survey := map[string]interface{}{
		"id":          exam.ID,
		"title":       exam.Title,
		"description": exam.Description,
		"category":    exam.Category,
		"tags":        exam.Tags,
		"visibility":  exam.Visibility,
		"allowMulti":  exam.AllowMulti,
		"anonymous":   exam.Anonymous,
		"showResult":  exam.ShowResult,
		"startTime":   exam.StartTime,
		"endTime":     exam.EndTime,
		"maxResponse": exam.MaxResponse,
		"duration":    exam.Duration,
		"maxAttempts": exam.MaxAttempts,
		"showScore":   exam.ShowScore,
		"status":      exam.Status,
		"deptIds":     exam.DeptIds,
		"mode":        exam.Mode,
		"createBy":    exam.CreateBy,
		"settings":    exam.Settings,
	}
	response.JSON(c, map[string]interface{}{"survey": survey, "responseCount": respCnt, "schema": rawSchema})
}

func (h *AdminExamHandler) Save(_ context.Context, c *app.RequestContext) {
	type ExamSaveReq struct {
		ID          uint   `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Tags        string `json:"tags"`
		Visibility  int    `json:"visibility"`
		AllowMulti  int    `json:"allowMulti"`
		Anonymous   int    `json:"anonymous"`
		ShowResult  int    `json:"showResult"`
		StartTime   int64  `json:"startTime"`
		EndTime     int64  `json:"endTime"`
		MaxResponse int    `json:"maxResponse"`
		Duration    int    `json:"duration"`
		MaxAttempts int    `json:"maxAttempts"`
		ShowScore   int    `json:"showScore"`
		Status      int    `json:"status"`
		Schema      string `json:"schema"`
		DeptIds     string `json:"deptIds"`
		Mode        string `json:"mode"`
		Settings    string `json:"settings"`
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
			Description: req.Description,
			Category:    req.Category,
			Tags:        req.Tags,
			Visibility:  req.Visibility,
			AllowMulti:  req.AllowMulti,
			Anonymous:   req.Anonymous,
			ShowResult:  req.ShowResult,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			MaxResponse: req.MaxResponse,
			Schema:      req.Schema,
			DeptIds:     req.DeptIds,
			Mode:        req.Mode,
			Settings:    req.Settings,
			Duration:    req.Duration,
			MaxAttempts: req.MaxAttempts,
			ShowScore:   req.ShowScore,
			Status:      req.Status,
			AddTime:     now,
		}
		if exam.Mode == "" {
			exam.Mode = "exam"
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
			"exam_title":       req.Title,
			"exam_desc":        req.Description,
			"exam_category":    req.Category,
			"exam_tags":        req.Tags,
			"exam_visibility":  req.Visibility,
			"exam_allow_multi": req.AllowMulti,
			"exam_anonymous":   req.Anonymous,
			"exam_show_result": req.ShowResult,
			"exam_start_time":  req.StartTime,
			"exam_end_time":    req.EndTime,
			"exam_max_response": req.MaxResponse,
			"exam_schema":      req.Schema,
			"exam_dept_ids":    req.DeptIds,
			"exam_mode":        req.Mode,
			"exam_settings":      req.Settings,
			"exam_duration":      req.Duration,
			"exam_max_attempts":  req.MaxAttempts,
			"exam_show_score":    req.ShowScore,
			"exam_status":        req.Status,
			"exam_edit_time":   now,
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
	if cat := c.Query("category"); cat != "" {
		q = q.Where("`exam_category` = ?", cat)
	}
	if st := c.Query("status"); st != "" {
		if s, err := strconv.Atoi(st); err == nil && s >= 0 {
			q = q.Where("`exam_status` = ?", s)
		}
	}
	var total int64
	q.Count(&total)
	var list []model.Exam
	q.Order("`exam_id` DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

func (h *AdminExamHandler) Status(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	if id <= 0 {
		response.Fail(c, "无效的考试ID")
		return
	}
	if err := database.DB.Model(&model.Exam{}).Where("`exam_id` = ?", id).Update("exam_status", status).Error; err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
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

func (h *AdminExamHandler) RecordList(_ context.Context, c *app.RequestContext) {
	examId, _ := strconv.Atoi(c.Query("examId"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examId)
	if kw := c.Query("keyword"); kw != "" {
		q = q.Where("`exam_r_user_id` LIKE ?", "%"+kw+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.ExamRecord
	q.Order("`exam_r_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

func (h *AdminExamHandler) RecordDetail(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	var record model.ExamRecord
	if err := database.DB.Where("`exam_r_id` = ?", id).First(&record).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	var exam model.Exam
	database.DB.Where("`exam_id` = ?", record.ExamID).First(&exam)
	var answers map[string]interface{}
	json.Unmarshal([]byte(record.Answers), &answers)
	var scoring map[string]bool
	json.Unmarshal([]byte(record.Result), &scoring)
	var schema interface{}
	if exam.Schema != "" {
		json.Unmarshal([]byte(exam.Schema), &schema)
	}
	response.JSON(c, map[string]interface{}{
		"record":  record,
		"answers": answers,
		"scoring": scoring,
		"schema":  schema,
	})
}

func (h *AdminExamHandler) RecordDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	database.DB.Where("`exam_r_id` = ?", id).Delete(&model.ExamRecord{})
	response.JSON(c, nil)
}

func (h *AdminExamHandler) RecordBatchDel(_ context.Context, c *app.RequestContext) {
	ids := c.PostForm("ids")
	if ids == "" {
		response.Fail(c, "参数错误")
		return
	}
	database.DB.Where("`exam_r_id` IN ?", strings.Split(ids, ",")).Delete(&model.ExamRecord{})
	response.JSON(c, nil)
}

func (h *AdminExamHandler) Statistics(_ context.Context, c *app.RequestContext) {
	examId, _ := strconv.Atoi(c.Query("examId"))
	if examId <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	var total, submitted, passed int64
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examId).Count(&total)
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` = 1", examId).Count(&submitted)
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_pass` = 1", examId).Count(&passed)
	var passRate float64
	if submitted > 0 {
		passRate = float64(passed) / float64(submitted)
	}
	// daily trends (last 7 days)
	type dailyCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	var daily []dailyCount
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local).UnixMilli()
		end := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 999, time.Local).UnixMilli()
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_submit_time` >= ? AND `exam_r_submit_time` <= ?", examId, start, end).Count(&cnt)
		daily = append(daily, dailyCount{Date: day.Format("01-02"), Count: cnt})
	}
	// score distribution
	type scoreDist struct {
		Score int   `json:"score"`
		Count int64 `json:"count"`
	}
	var sds []scoreDist
	database.DB.Model(&model.ExamRecord{}).Select("FLOOR(`exam_r_score`/10)*10 as score, COUNT(*) as count").Where("`exam_r_exam_id` = ? AND `exam_r_status` = 1", examId).Group("score").Order("score").Find(&sds)
	scoreDistMap := make(map[string]int64)
	for _, sd := range sds {
		key := strconv.Itoa(sd.Score) + "-" + strconv.Itoa(sd.Score+9)
		scoreDistMap[key] = sd.Count
	}
	// field stats from saved schema
	var exam model.Exam
	database.DB.Where("`exam_id` = ?", examId).First(&exam)
	var schema struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	json.Unmarshal([]byte(exam.Schema), &schema)
	var fieldStats []map[string]interface{}
	for _, q := range schema.Questions {
		qid, _ := q["id"].(string)
		qtype, _ := q["type"].(string)
		title, _ := q["title"].(string)
		var nonEmpty, empty, totalCnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` = 1 AND JSON_EXTRACT(`exam_r_answers`, '$.\""+qid+"\"') IS NOT NULL AND JSON_EXTRACT(`exam_r_answers`, '$.\""+qid+"\"') != ''", examId).Count(&nonEmpty)
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` = 1", examId).Count(&totalCnt)
		empty = totalCnt - nonEmpty
		fs := map[string]interface{}{
			"questionId": qid,
			"type":       qtype,
			"title":      title,
			"nonEmpty":   nonEmpty,
			"empty":      empty,
			"totalCount": totalCnt,
		}
		fieldStats = append(fieldStats, fs)
	}
	response.JSON(c, map[string]interface{}{
		"total":     total,
		"submitted": submitted,
		"passed":    passed,
		"passRate":  passRate,
		"daily":     daily,
		"scoreDist": scoreDistMap,
		"fieldStats": fieldStats,
	})
}
