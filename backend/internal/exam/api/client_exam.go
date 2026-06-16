package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	examPkg "wecheckin-backend/backend/internal/formkit/exam"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type ClientExamHandler struct{}

func NewClientExamHandler() *ClientExamHandler { return &ClientExamHandler{} }

// List GET /exam/list
func (h *ClientExamHandler) List(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.Exam{}).Where("`exam_status` = 1")
	if kw := c.Query("keyword"); kw != "" {
		q = q.Where("`exam_title` LIKE ?", "%"+kw+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Exam
	q.Order("`exam_order` DESC, `exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// View GET /exam/view?id=
func (h *ClientExamHandler) View(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id == 0 {
		response.Fail(c, "id 必填")
		return
	}
	var e model.Exam
	if err := database.DB.Where("`exam_id` = ? AND `exam_status` = 1", id).First(&e).Error; err != nil {
		response.Fail(c, "考试不存在或未发布")
		return
	}
	var p model.ExamPaper
	if err := database.DB.Where("`exam_p_id` = ?", e.PaperID).First(&p).Error; err != nil {
		response.Fail(c, "试卷不存在")
		return
	}
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		database.DB.Where("`exam_q_id` IN ?", qids).Find(&qs)
	}
	safe := make([]map[string]interface{}, 0, len(qs))
	for _, q := range qs {
		safe = append(safe, map[string]interface{}{
			"id":         q.ID,
			"type":       q.Type,
			"title":      q.Title,
			"options":    q.Options,
			"score":      q.Score,
			"difficulty": q.Difficulty,
			"category":   q.Category,
		})
	}
	startAt := time.Now().UnixMilli()
	response.JSON(c, map[string]interface{}{
		"exam":      e,
		"paper":     p,
		"questions": safe,
		"startAt":   startAt,
		"session":   strconv.FormatInt(startAt, 36),
	})
}

// Start GET /exam/start?examId=
func (h *ClientExamHandler) Start(_ context.Context, c *app.RequestContext) {
	uidVal, _ := c.Get("user_id")
	uid := uint(uidVal.(int64))
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	examID, _ := strconv.Atoi(c.Query("examId"))
	if examID == 0 {
		response.Fail(c, "examId 必填")
		return
	}
	var e model.Exam
	if err := database.DB.Where("`exam_id` = ?", examID).First(&e).Error; err != nil {
		response.Fail(c, "考试不存在")
		return
	}
	if e.Status != 1 {
		response.Fail(c, "考试未发布")
		return
	}
	if e.MaxAttempts > 0 {
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ?", examID, uid).Count(&cnt)
		if int(cnt) >= e.MaxAttempts {
			response.Fail(c, "已达最大尝试次数")
			return
		}
	}
	nowMs := time.Now().UnixMilli()
	if e.StartTime > 0 && nowMs < e.StartTime {
		response.Fail(c, "考试未开始")
		return
	}
	if e.EndTime > 0 && nowMs > e.EndTime {
		response.Fail(c, "考试已结束")
		return
	}
	uidStr := strconv.FormatUint(uint64(uid), 10)
	var rec model.ExamRecord
	err := database.DB.Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ? AND `exam_r_status` IN (0,1)", examID, uidStr).Order("`exam_r_id` DESC").First(&rec).Error
	if err != nil {
		var p model.ExamPaper
		if err := database.DB.Where("`exam_p_id` = ?", e.PaperID).First(&p).Error; err != nil {
			response.Fail(c, "试卷不存在")
			return
		}
		rec = model.ExamRecord{
			ExamID:     uint(examID),
			PaperID:    e.PaperID,
			UserID:     uidStr,
			TotalScore: p.TotalScore,
			Status:     0,
			StartTime:  nowMs,
		}
		database.DB.Create(&rec)
	}
	var p model.ExamPaper
	database.DB.Where("`exam_p_id` = ?", e.PaperID).First(&p)
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		database.DB.Where("`exam_q_id` IN ?", qids).Find(&qs)
	}
	safe := make([]map[string]interface{}, 0, len(qs))
	for _, q := range qs {
		safe = append(safe, map[string]interface{}{
			"id":         q.ID,
			"type":       q.Type,
			"title":      q.Title,
			"options":    q.Options,
			"score":      q.Score,
			"difficulty": q.Difficulty,
		})
	}
	var prevAnswers map[string]interface{}
	if rec.Answers != "" {
		_ = json.Unmarshal([]byte(rec.Answers), &prevAnswers)
	}
	response.JSON(c, map[string]interface{}{
		"record":    rec,
		"paper":     p,
		"exam":      e,
		"questions": safe,
		"answers":   prevAnswers,
	})
}

// SaveAnswer POST /exam/save_answer
func (h *ClientExamHandler) SaveAnswer(_ context.Context, c *app.RequestContext) {
	uidVal, _ := c.Get("user_id")
	uid := uint(uidVal.(int64))
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	recordID, _ := strconv.Atoi(c.PostForm("recordId"))
	answersJSON := c.PostForm("answers")
	if recordID == 0 {
		response.Fail(c, "recordId 必填")
		return
	}
	var rec model.ExamRecord
	if err := database.DB.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", recordID, uid).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	if rec.Status == 2 {
		response.Fail(c, "已提交，不可修改")
		return
	}
	if err := database.DB.Model(&rec).Update("exam_r_answers", answersJSON).Error; err != nil {
		response.Fail(c, "保存失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// Submit POST /exam/submit
func (h *ClientExamHandler) Submit(_ context.Context, c *app.RequestContext) {
	uidVal, _ := c.Get("user_id")
	uid := uint(uidVal.(int64))
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	recordID, _ := strconv.Atoi(c.PostForm("recordId"))
	answersJSON := c.PostForm("answers")
	if recordID == 0 {
		response.Fail(c, "recordId 必填")
		return
	}
	var rec model.ExamRecord
	if err := database.DB.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", recordID, uid).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	if rec.Status == 2 {
		response.Fail(c, "已提交")
		return
	}
	var p model.ExamPaper
	database.DB.Where("`exam_p_id` = ?", rec.PaperID).First(&p)
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
	var answers map[string]interface{}
	if answersJSON != "" {
		_ = json.Unmarshal([]byte(answersJSON), &answers)
	}
	res := examPkg.Grade(exQs, answers)
	resultJSON, _ := json.Marshal(res)
	nowMs := time.Now().UnixMilli()
	updates := map[string]interface{}{
		"exam_r_answers":     answersJSON,
		"exam_r_score":       res.TotalScore,
		"exam_r_status":      1,
		"exam_r_submit_time": nowMs,
		"exam_r_result":      string(resultJSON),
	}
	if res.ManualCount == 0 {
		updates["exam_r_status"] = 2
		updates["exam_r_pass"] = res.TotalScore >= p.PassScore
	}
	if err := database.DB.Model(&rec).Updates(updates).Error; err != nil {
		response.Fail(c, "提交失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]interface{}{
		"score":      res.TotalScore,
		"fullScore":  res.FullScore,
		"correctCnt": res.CorrectCnt,
		"manualCnt":  res.ManualCount,
		"results":    res.Results,
	})
}

// Record GET /exam/record?id=
func (h *ClientExamHandler) Record(_ context.Context, c *app.RequestContext) {
	uidVal, _ := c.Get("user_id")
	uid := uint(uidVal.(int64))
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	id, _ := strconv.Atoi(c.Query("id"))
	var rec model.ExamRecord
	if err := database.DB.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", id, uid).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	var p model.ExamPaper
	database.DB.Where("`exam_p_id` = ?", rec.PaperID).First(&p)
	var e model.Exam
	database.DB.Where("`exam_id` = ?", rec.ExamID).First(&e)
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		database.DB.Where("`exam_q_id` IN ?", qids).Find(&qs)
	}
	safe := make([]map[string]interface{}, 0, len(qs))
	for _, q := range qs {
		item := map[string]interface{}{
			"id":      q.ID,
			"type":    q.Type,
			"title":   q.Title,
			"options": q.Options,
			"score":   q.Score,
		}
		if p.ShowAnswer == 1 || rec.Status == 2 {
			item["answer"] = q.Answer
			item["analysis"] = q.Analysis
		}
		safe = append(safe, item)
	}
	var prevAnswers map[string]interface{}
	if rec.Answers != "" {
		_ = json.Unmarshal([]byte(rec.Answers), &prevAnswers)
	}
	var results []examPkg.Result
	if rec.Result != "" {
		_ = json.Unmarshal([]byte(rec.Result), &results)
	}
	response.JSON(c, map[string]interface{}{
		"record":    rec,
		"exam":      e,
		"paper":     p,
		"questions": safe,
		"answers":   prevAnswers,
		"results":   results,
	})
}

// MyRecords GET /exam/my_records
func (h *ClientExamHandler) MyRecords(_ context.Context, c *app.RequestContext) {
	uidVal, _ := c.Get("user_id")
	uid := uint(uidVal.(int64))
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	var list []model.ExamRecord
	database.DB.Where("`exam_r_user_id` = ?", uid).Order("`exam_r_id` DESC").Limit(50).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list})
}
