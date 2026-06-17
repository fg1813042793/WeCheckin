package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	examPkg "wecheckin-backend/backend/internal/formkit/exam"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/response"
)

type ClientExamHandler struct{}

func NewClientExamHandler() *ClientExamHandler { return &ClientExamHandler{} }

// List GET /exam/list
// @Tags 考试-客户端
// @Summary 获取考试列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /exam/list [get]
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
// @Tags 考试-客户端
// @Summary 查看考试详情
// @Param id query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /exam/view [get]
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
	// 登录可见 / 部门限定：检查用户登录
	if e.Visibility == 1 || e.Visibility == 2 {
		token := ""
		auth := c.GetHeader("Authorization")
		if len(auth) > 0 {
			token = string(auth)
		}
		if token == "" {
			response.Fail(c, "请先登录")
			return
		}
		rdKey := "user_token:a:" + token
		jsonStr, err := rd.RDB.Get(rd.Ctx, rdKey).Result()
		if err != nil || jsonStr == "" {
			response.Fail(c, "请先登录")
			return
		}
		// 部门限定：校验用户部门
		if e.Visibility == 2 && e.DeptIds != "" {
			var userInfo map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &userInfo)
			uid := uint(0)
			if id, ok := userInfo["id"].(float64); ok {
				uid = uint(id)
			}
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", uid).First(&ud)
			deptIds := strings.Split(e.DeptIds, ",")
			allowed := false
			for _, did := range deptIds {
				d, _ := strconv.Atoi(strings.TrimSpace(did))
				if uint(d) == ud.DeptID {
					allowed = true
					break
				}
			}
			if !allowed {
				response.Fail(c, "您不在该考试的可见部门中")
				return
			}
		}
	}
	// 兼容两种存储模式：
	//   PaperID > 0 : 独立试卷（ExamPaper + ExamQuestion）
	//   PaperID == 0: Schema JSON 内联（ExamDesigner 新建）
	if e.PaperID > 0 {
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
				"id":                 q.ID,
				"type":               q.Type,
				"title":              q.Title,
				"options":            q.Options,
				"score":              q.Score,
				"difficulty":         q.Difficulty,
				"category":           q.Category,
				"examCorrectAnswer":  q.Answer,
				"examAnalysis":       q.Analysis,
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
		return
	}
	// PaperID == 0: 直接解析 schema（Survey 风格）
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Schema), &schMap)
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
	if settingsMap == nil {
		settingsMap = make(map[string]interface{})
	}
	if _, ok := settingsMap["answerVisible"]; !ok {
		settingsMap["answerVisible"] = true
	}
	now := time.Now().UnixMilli()
	session := c.Query("session")
	if session == "" {
		session = fmt.Sprintf("%x", time.Now().UnixNano()+rand.Int63())
	}
	redisKey := fmt.Sprintf("exam_session:%d:%s", e.ID, session)
	var startAt int64
	exists, _ := rd.RDB.Exists(rd.Ctx, redisKey).Result()
	if exists == 0 {
		rd.RDB.Set(rd.Ctx, redisKey, now, 24*time.Hour)
		startAt = now
	} else {
		v, err := rd.RDB.Get(rd.Ctx, redisKey).Int64()
		if err == nil {
			startAt = v
		} else {
			startAt = now
		}
	}
	resp := map[string]interface{}{
		"id":            e.ID,
		"title":         e.Title,
		"description":   e.Description,
		"visibility":    e.Visibility,
		"anonymous":     e.Anonymous,
		"showResult":    e.ShowResult,
		"showScore":     e.ShowScore,
		"duration":      e.Duration,
		"maxAttempts":   e.MaxAttempts,
		"startTime":     e.StartTime,
		"endTime":       e.EndTime,
		"schema":        schMap,
		"settings":      settingsMap,
		"startAt":       startAt,
		"session":       session,
		"deptIds":       e.DeptIds,
		"mode":          e.Mode,
	}
	response.JSON(c, resp)
}

// Start GET /exam/start?examId=
// @Tags 考试-客户端
// @Summary 开始考试
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /exam/start [get]
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
// @Tags 考试-客户端
// @Summary 保存答案
// @Param recordId formData int true "记录ID"
// @Param answers formData string true "答案JSON"
// @Success 200 {object} response.Resp
// @Router /exam/save_answer [post]
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
// @Tags 考试-客户端
// @Summary 提交考试
// @Param recordId formData int true "记录ID"
// @Param answers formData string true "答案JSON"
// @Success 200 {object} response.Resp
// @Router /exam/submit [post]
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
// @Tags 考试-客户端
// @Summary 查看考试记录
// @Param id query int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /exam/record [get]
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
// @Tags 考试-客户端
// @Summary 我的考试记录
// @Success 200 {object} response.Resp
// @Router /exam/my_records [get]
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
