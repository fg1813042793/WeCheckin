package api

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	examPkg "wecheckin-backend/backend/internal/formkit/exam"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// ==================== Question 题库 ====================

// ListQuestion GET /admin/survey/question_list
func (h *AdminSurveyHandler) ListQuestion(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.ExamQuestion{})
	if kw := c.Query("keyword"); kw != "" {
		q = q.Where("`exam_q_title` LIKE ?", "%"+kw+"%")
	}
	if cat := c.Query("category"); cat != "" {
		q = q.Where("`exam_q_category` = ?", cat)
	}
	var total int64
	q.Count(&total)
	var list []model.ExamQuestion
	q.Order("`exam_q_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// InsertQuestion POST /admin/survey/question_insert
func (h *AdminSurveyHandler) InsertQuestion(_ context.Context, c *app.RequestContext) {
	var q model.ExamQuestion
	if err := c.BindAndValidate(&q); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if q.Title == "" || q.Type == "" {
		response.Fail(c, "标题和类型不能为空")
		return
	}
	q.AddTime = database.Now()
	if err := database.DB.Create(&q).Error; err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.JSON(c, q)
}

// EditQuestion POST /admin/survey/question_edit
func (h *AdminSurveyHandler) EditQuestion(_ context.Context, c *app.RequestContext) {
	var q model.ExamQuestion
	if err := c.BindAndValidate(&q); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if q.ID == 0 {
		response.Fail(c, "ID 不能为空")
		return
	}
	if err := database.DB.Model(&model.ExamQuestion{}).Where("`exam_q_id` = ?", q.ID).Updates(map[string]interface{}{
		"exam_q_type":       q.Type,
		"exam_q_title":      q.Title,
		"exam_q_options":    q.Options,
		"exam_q_answer":     q.Answer,
		"exam_q_score":      q.Score,
		"exam_q_category":   q.Category,
		"exam_q_tags":       q.Tags,
		"exam_q_analysis":   q.Analysis,
		"exam_q_difficulty": q.Difficulty,
		"exam_q_status":     q.Status,
	}).Error; err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// DelQuestion POST /admin/survey/question_del
func (h *AdminSurveyHandler) DelQuestion(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`exam_q_id` = ?", id).Delete(&model.ExamQuestion{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ==================== Paper 试卷 ====================

// ListPaper GET /admin/survey/paper_list
func (h *AdminSurveyHandler) ListPaper(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.ExamPaper{})
	if kw := c.Query("keyword"); kw != "" {
		q = q.Where("`exam_p_title` LIKE ?", "%"+kw+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.ExamPaper
	q.Order("`exam_p_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// GetPaperDetail GET /admin/survey/paper_detail?id=
func (h *AdminSurveyHandler) GetPaperDetail(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	var p model.ExamPaper
	if err := database.DB.Where("`exam_p_id` = ?", id).First(&p).Error; err != nil {
		response.Fail(c, "试卷不存在")
		return
	}
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		database.DB.Where("`exam_q_id` IN ?", qids).Find(&qs)
	}
	response.JSON(c, map[string]interface{}{"paper": p, "questions": qs})
}

// InsertPaper POST /admin/survey/paper_insert
func (h *AdminSurveyHandler) InsertPaper(_ context.Context, c *app.RequestContext) {
	var p model.ExamPaper
	if err := c.BindAndValidate(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if p.Title == "" {
		response.Fail(c, "标题不能为空")
		return
	}
	if p.QuestionIDs != "" {
		var qids []uint
		_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
		var qs []model.ExamQuestion
		if len(qids) > 0 {
			database.DB.Where("`exam_q_id` IN ?", qids).Find(&qs)
			total := 0
			for _, q := range qs {
				total += q.Score
			}
			p.TotalScore = total
		}
	}
	p.AddTime = database.Now()
	if err := database.DB.Create(&p).Error; err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.JSON(c, p)
}

// EditPaper POST /admin/survey/paper_edit
func (h *AdminSurveyHandler) EditPaper(_ context.Context, c *app.RequestContext) {
	var p model.ExamPaper
	if err := c.BindAndValidate(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if p.ID == 0 {
		response.Fail(c, "ID 不能为空")
		return
	}
	if err := database.DB.Model(&model.ExamPaper{}).Where("`exam_p_id` = ?", p.ID).Updates(map[string]interface{}{
		"exam_p_title":        p.Title,
		"exam_p_desc":         p.Description,
		"exam_p_question_ids": p.QuestionIDs,
		"exam_p_total_score":  p.TotalScore,
		"exam_p_time_limit":   p.TimeLimit,
		"exam_p_pass_score":   p.PassScore,
		"exam_p_shuffle":      p.Shuffle,
		"exam_p_show_answer":  p.ShowAnswer,
		"exam_p_category":     p.Category,
		"exam_p_status":       p.Status,
	}).Error; err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// DelPaper POST /admin/survey/paper_del
func (h *AdminSurveyHandler) DelPaper(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`exam_p_id` = ?", id).Delete(&model.ExamPaper{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ==================== Exam 考试 ====================

// ListExam GET /admin/survey/exam_list
func (h *AdminSurveyHandler) ListExam(_ context.Context, c *app.RequestContext) {
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
	q.Order("`exam_order` DESC, `exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// InsertExam POST /admin/survey/exam_insert
func (h *AdminSurveyHandler) InsertExam(_ context.Context, c *app.RequestContext) {
	var e model.Exam
	if err := c.BindAndValidate(&e); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if e.Title == "" || e.PaperID == 0 {
		response.Fail(c, "标题和试卷不能为空")
		return
	}
	e.AddTime = database.Now()
	if err := database.DB.Create(&e).Error; err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.JSON(c, e)
}

// EditExam POST /admin/survey/exam_edit
func (h *AdminSurveyHandler) EditExam(_ context.Context, c *app.RequestContext) {
	var e model.Exam
	if err := c.BindAndValidate(&e); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if e.ID == 0 {
		response.Fail(c, "ID 不能为空")
		return
	}
	if err := database.DB.Model(&model.Exam{}).Where("`exam_id` = ?", e.ID).Updates(map[string]interface{}{
		"exam_title":            e.Title,
		"exam_paper_id":         e.PaperID,
		"exam_start_time":       e.StartTime,
		"exam_end_time":         e.EndTime,
		"exam_duration":         e.Duration,
		"exam_max_attempts":     e.MaxAttempts,
		"exam_show_score":       e.ShowScore,
		"exam_publish_dept_ids": e.PublishDepts,
		"exam_qr":               e.QR,
		"exam_status":           e.Status,
		"exam_order":            e.Order,
	}).Error; err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// DelExam POST /admin/survey/exam_del
func (h *AdminSurveyHandler) DelExam(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`exam_id` = ?", id).Delete(&model.Exam{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ==================== Record / Grading ====================

// ListRecord GET /admin/survey/record_list?examId=xx
func (h *AdminSurveyHandler) ListRecord(_ context.Context, c *app.RequestContext) {
	examID, _ := strconv.Atoi(c.Query("examId"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examID)
	var total int64
	q.Count(&total)
	var list []model.ExamRecord
	q.Order("`exam_r_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// ManualGrade POST /admin/survey/manual_grade?recordId=&qid=&score=
// 对某题人工判分，更新每题结果 + 总分 + 状态
func (h *AdminSurveyHandler) ManualGrade(_ context.Context, c *app.RequestContext) {
	recordID, _ := strconv.Atoi(c.Query("recordId"))
	if recordID == 0 {
		recordID, _ = strconv.Atoi(c.PostForm("recordId"))
	}
	qid, _ := strconv.Atoi(c.PostForm("qid"))
	score, _ := strconv.Atoi(c.PostForm("score"))
	if recordID == 0 || qid == 0 {
		response.Fail(c, "参数错误")
		return
	}
	var rec model.ExamRecord
	if err := database.DB.Where("`exam_r_id` = ?", recordID).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	// 解析已有的每题结果
	var results []examPkg.Result
	if rec.Result != "" {
		_ = json.Unmarshal([]byte(rec.Result), &results)
	}
	found := false
	newTotal := 0
	needsManual := false
	for i, r := range results {
		if r.QuestionID == uint(qid) {
			results[i].GotScore = score
			results[i].Correct = score >= r.FullScore
			results[i].NeedManual = false
			results[i].Reason = ""
			found = true
		}
		newTotal += results[i].GotScore
		if results[i].NeedManual {
			needsManual = true
		}
	}
	if !found {
		response.Fail(c, "该题目不在本记录中")
		return
	}
	resultJSON, _ := json.Marshal(results)
	updates := map[string]interface{}{
		"exam_r_score":  newTotal,
		"exam_r_result": string(resultJSON),
		"exam_r_status": 2,
	}
	if !needsManual {
		// 全部批改完成
		var paper model.ExamPaper
		if err := database.DB.Where("`exam_p_id` = ?", rec.PaperID).First(&paper).Error; err == nil {
			if paper.PassScore > 0 && newTotal >= paper.PassScore {
				updates["exam_r_pass"] = 1
			} else if paper.PassScore > 0 {
				updates["exam_r_pass"] = 0
			}
		}
	}
	database.DB.Model(&rec).Updates(updates)
	response.JSON(c, nil)
}

// PreviewGrade POST /admin/survey/preview_grade
func (h *AdminSurveyHandler) PreviewGrade(_ context.Context, c *app.RequestContext) {
	var req struct {
		PaperID uint                   `json:"paperId"`
		Answers map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	var p model.ExamPaper
	if err := database.DB.Where("`exam_p_id` = ?", req.PaperID).First(&p).Error; err != nil {
		response.Fail(c, "试卷不存在")
		return
	}
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		database.DB.Where("`exam_q_id` IN ?", qids).Find(&qs)
	}
	exQs := make([]examPkg.Question, 0, len(qs))
	for _, q := range qs {
		exQs = append(exQs, examPkg.Question{
			ID: q.ID, Type: q.Type, Title: q.Title,
			Options: q.Options, Answer: q.Answer, Score: q.Score,
			NeedManual: examPkg.QWithType(q.Type),
		})
	}
	res := examPkg.Grade(exQs, req.Answers)
	response.JSON(c, res)
}
