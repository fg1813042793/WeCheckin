package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	examPkg "wecheckin-backend/backend/internal/app/formkit/exam"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/response"
	"wecheckin-backend/backend/pkg/tokenutil"
)

// Submit POST /exam/submit
// @Tags 客户端-考试
// @Summary 提交考试
// @Router /exam/submit [post]
func (h *ClientExamHandler) Submit(ctx context.Context, c *app.RequestContext) {
	raw, _ := c.Body()
	var req struct {
		RecordID   int                    `json:"recordId"`
		ExamID     int                    `json:"examId"`
		Answers    map[string]interface{} `json:"answers"`
		Session    string                 `json:"session"`
		Device     string                 `json:"device"`
		DeviceID   string                 `json:"deviceId"`
		AutoSubmit bool                   `json:"autoSubmit"`
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	answersJSON, _ := json.Marshal(req.Answers)
	clientIP := c.ClientIP()
	db, dbCancel := database.WithContext(ctx)
	defer dbCancel()

	if req.RecordID > 0 {
		uid := getUID(c)
		if uid == 0 {
			response.Fail(c, "未登录")
			return
		}
		var rec model.ExamRecord
		if err := db.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", req.RecordID, uid).First(&rec).Error; err != nil {
			response.Fail(c, "记录不存在")
			return
		}
		if rec.Status == 2 {
			response.Fail(c, "已提交")
			return
		}
		var p model.ExamPaper
		db.Where("`exam_p_id` = ?", rec.PaperID).First(&p)
		var qids []uint
		_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
		var qs []model.ExamQuestion
		if len(qids) > 0 {
			db.Where("`exam_q_id` IN ?", qids).Find(&qs)
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
		resultJSON, _ := json.Marshal(res)
		nowMs := time.Now().UnixMilli()
		autoSubmitVal := 0
		if req.AutoSubmit {
			autoSubmitVal = 1
		}
		updates := map[string]interface{}{
			"exam_r_answers":     string(answersJSON),
			"exam_r_score":       res.TotalScore,
			"exam_r_status":      1,
			"exam_r_submit_time": nowMs,
			"exam_r_result":      string(resultJSON),
			"exam_r_auto_submit": autoSubmitVal,
			"exam_r_device_id":   req.DeviceID,
			"exam_r_add_ip":      clientIP,
		}
		if rec.StartTime > 0 && nowMs > rec.StartTime {
			updates["exam_r_time_spent"] = int((nowMs - rec.StartTime) / 1000)
		}
		if res.ManualCount == 0 {
			updates["exam_r_status"] = 2
			updates["exam_r_pass"] = res.TotalScore >= p.PassScore
		}
		if err := db.Model(&rec).Updates(updates).Error; err != nil {
			logger.Logger.Printf("[ExamSubmit] PaperID模式提交失败 examId=%d recordId=%d uid=%d err=%s", rec.ExamID, req.RecordID, uid, err.Error())
			response.Fail(c, "提交失败: "+err.Error())
			return
		}
		logger.Logger.Printf("[ExamSubmit] PaperID模式提交成功 examId=%d recordId=%d uid=%d score=%d fullScore=%d", rec.ExamID, req.RecordID, uid, res.TotalScore, res.FullScore)
		response.JSON(c, examSubmitResponse{Score: res.TotalScore, FullScore: res.FullScore, CorrectCnt: res.CorrectCnt, ManualCnt: res.ManualCount, Results: res.Results})
		return
	}

	if req.ExamID == 0 || req.Session == "" {
		response.Fail(c, "参数错误")
		return
	}
	redisKey := fmt.Sprintf("exam_session:%d:%s", req.ExamID, req.Session)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	exist, _ := rd.RDB.Exists(redisCtx, redisKey).Result()
	if exist == 0 {
		logger.Logger.Printf("[ExamSubmit] 会话过期 examId=%d session=%s", req.ExamID, req.Session)
		response.Fail(c, "会话不存在或已过期")
		return
	}
	var e model.Exam
	if err := db.Where("`exam_id` = ? AND `exam_status` = 1", req.ExamID).First(&e).Error; err != nil {
		logger.Logger.Printf("[ExamSubmit] Schema模式考试不存在或未发布 examId=%d", req.ExamID)
		response.Fail(c, "考试不存在或未发布")
		return
	}
	if e.PaperID > 0 {
		response.Fail(c, "请使用 recordId 模式提交")
		return
	}
	uidStr := ""
	auth := string(c.GetHeader("Authorization"))
	if auth != "" {
		rdKey := tokenutil.TokenAuthKey("user", auth)
		if jsonStr, err := rd.RDB.Get(redisCtx, rdKey).Result(); err == nil && jsonStr != "" {
			var userInfo struct {
				ID uint `json:"id"`
			}
			if json.Unmarshal([]byte(jsonStr), &userInfo) == nil && userInfo.ID > 0 {
				uidStr = fmt.Sprintf("%d", userInfo.ID)
			}
		}
	}
	sessionStart, _ := rd.RDB.Get(redisCtx, redisKey).Int64()
	clientIP = c.ClientIP()
	if msg := checkExamLimitContext(ctx, &e, uidStr, req.Device, req.DeviceID, clientIP); msg != "" {
		response.Fail(c, msg)
		return
	}
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Schema), &schMap)
	rawQuestions, _ := schMap["questions"].([]interface{})
	var qs []examPkg.Question
	idMap := make(map[string]uint)
	seqID := uint(1)
	for _, raw := range rawQuestions {
		q, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := q["id"].(string)
		if id == "" {
			continue
		}
		idMap[id] = seqID
		typ, _ := q["type"].(string)
		title, _ := q["title"].(string)
		score := 0
		if s, ok := q["examScore"].(float64); ok {
			score = int(s)
		}
		answer, _ := q["examCorrectAnswer"].(string)
		qs = append(qs, examPkg.Question{
			ID: seqID, Type: typ, Title: title,
			Answer: answer, Score: score,
			NeedManual: examPkg.QWithType(typ),
		})
		seqID++
	}
	mapped := make(map[string]interface{})
	for k, v := range req.Answers {
		if sid, ok := idMap[k]; ok {
			mapped[fmt.Sprintf("%d", sid)] = v
		}
	}
	res := examPkg.Grade(qs, mapped)
	resultJSON, _ := json.Marshal(res)
	nowMs := time.Now().UnixMilli()
	timeSpent := 0
	if sessionStart > 0 && nowMs > sessionStart {
		timeSpent = int((nowMs - sessionStart) / 1000)
	}
	autoSubmitVal := 0
	if req.AutoSubmit {
		autoSubmitVal = 1
	}
	rec := model.ExamRecord{
		ExamID:       uint(req.ExamID),
		UserID:       uidStr,
		Answers:      string(answersJSON),
		Score:        res.TotalScore,
		TotalScore:   res.FullScore,
		Status:       2,
		StartTime:    sessionStart,
		SubmitTime:   nowMs,
		TimeSpent:    timeSpent,
		IsAutoSubmit: autoSubmitVal,
		Device:       req.Device,
		DeviceID:     req.DeviceID,
		AddIP:        clientIP,
		Session:      req.Session,
		Result:       string(resultJSON),
	}
	if err := db.Create(&rec).Error; err != nil {
		logger.Logger.Printf("[ExamSubmit] Schema模式持久化失败 examId=%d uid=%s err=%s", req.ExamID, uidStr, err.Error())
		response.Fail(c, "提交失败: "+err.Error())
		return
	}
	logger.Logger.Printf("[ExamSubmit] Schema模式提交成功 examId=%d uid=%s score=%d fullScore=%d ip=%s device=%s", req.ExamID, uidStr, res.TotalScore, res.FullScore, clientIP, req.Device)
	rd.RDB.Del(redisCtx, redisKey)
	response.JSON(c, examSubmitResponse{Score: res.TotalScore, FullScore: res.FullScore, CorrectCnt: res.CorrectCnt, ManualCnt: res.ManualCount, Results: res.Results})
}
