package exam

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// Start GET /exam/start?examId=
// @Tags 客户端-考试
// @Summary 开始考试
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /exam/start [get]
func (h *ClientExamHandler) Start(ctx context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	examID, _ := strconv.Atoi(c.Query("examId"))
	if examID == 0 {
		response.Fail(c, "examId 必填")
		return
	}
	deviceId := c.Query("deviceId")
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var e model.Exam
	if err := db.Where("`exam_id` = ?", examID).First(&e).Error; err != nil {
		response.Fail(c, "考试不存在")
		return
	}
	if e.Status != 1 {
		response.Fail(c, "考试未发布")
		return
	}
	if e.MaxAttempts > 0 {
		var cnt int64
		db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ?", examID, uid).Count(&cnt)
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
	err := db.Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ? AND `exam_r_status` IN (0,1)", examID, uidStr).Order("`exam_r_id` DESC").First(&rec).Error
	if err != nil {
		var p model.ExamPaper
		if err := db.Where("`exam_p_id` = ?", e.PaperID).First(&p).Error; err != nil {
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
			DeviceID:   deviceId,
		}
		db.Create(&rec)
	}
	var p model.ExamPaper
	db.Where("`exam_p_id` = ?", e.PaperID).First(&p)
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		db.Where("`exam_q_id` IN ?", qids).Find(&qs)
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
	response.JSON(c, examStartResponse{Record: rec, Paper: p, Exam: e, Questions: safe, Answers: prevAnswers})
}
