package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	examPkg "wecheckin-backend/backend/internal/app/formkit/exam"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/response"
	"wecheckin-backend/backend/pkg/tokenutil"
)

type ClientExamHandler struct{}

func NewClientExamHandler() *ClientExamHandler { return &ClientExamHandler{} }

// List GET /exam/list
// @Tags 客户端-考试
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

	deviceId := c.Query("deviceId")
	clientIP := c.ClientIP()
	limitsMap := make(map[uint]map[string]bool)
	for _, e := range list {
		li := make(map[string]bool)
		var settingsMap map[string]interface{}
		_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
		if deviceLimit, _ := settingsMap["deviceLimit"].(float64); deviceLimit > 0 && deviceId != "" {
			var cnt int64
			database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_device_id` = ? AND `exam_r_status` >= 1", e.ID, deviceId).Count(&cnt)
			if int(cnt) >= int(deviceLimit) {
				li["deviceFull"] = true
			}
		}
		if ipLimit, _ := settingsMap["ipLimit"].(float64); ipLimit > 0 && clientIP != "" {
			var cnt int64
			database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_add_ip` = ? AND `exam_r_status` >= 1", e.ID, clientIP).Count(&cnt)
			if int(cnt) >= int(ipLimit) {
				li["ipFull"] = true
			}
		}
		if len(li) > 0 {
			limitsMap[e.ID] = li
		}
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize, "limits": limitsMap})
}

// View GET /exam/view?id=
// @Tags 客户端-考试
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
		logger.Logger.Printf("[ExamView] 考试不存在或未发布 examId=%d", id)
		response.Fail(c, "考试不存在或未发布")
		return
	}
	logger.Logger.Printf("[ExamView] 考试已发布 examId=%d title=%s visibility=%d", e.ID, e.Title, e.Visibility)
	nowMs := time.Now().UnixMilli()
	// 未到开考时间，拒绝访问（不创建 session）
	if e.StartTime > 0 && nowMs < e.StartTime {
		logger.Logger.Printf("[ExamView] 考试未开始 examId=%d startTime=%d", e.ID, e.StartTime)
		c.JSON(consts.StatusOK, map[string]interface{}{
			"code": 1,
			"msg":  "考试未开始",
			"data": map[string]interface{}{
				"startTime": e.StartTime,
				"endTime":   e.EndTime,
				"title":     e.Title,
			},
		})
		return
	}
	// 已过结束时间，拒绝访问
	if e.EndTime > 0 && nowMs > e.EndTime {
		logger.Logger.Printf("[ExamView] 考试已结束 examId=%d endTime=%d", e.ID, e.EndTime)
		response.Fail(c, "考试已结束")
		return
	}
	// 登录可见 / 部门限定：检查用户登录
	loginRequired := false
	if e.Visibility == 1 || e.Visibility == 2 {
		loginRequired = true
	} else {
		// 检查 settings 中的 loginRequired
		var settingsMap map[string]interface{}
		_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
		if v, ok := settingsMap["loginRequired"]; ok {
			if b, _ := v.(bool); b {
				loginRequired = true
			}
		}
	}
	if loginRequired {
		token := ""
		auth := c.GetHeader("Authorization")
		if len(auth) > 0 {
			token = string(auth)
		}
		if token == "" {
			logger.Logger.Printf("[ExamView] 未登录 examId=%d", e.ID)
			response.Fail(c, "请先登录")
			return
		}
		rdKey := tokenutil.TokenAuthKey("user", token)
		jsonStr, err := rd.RDB.Get(rd.Ctx, rdKey).Result()
		if err != nil || jsonStr == "" {
			logger.Logger.Printf("[ExamView] token无效 examId=%d", e.ID)
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
				logger.Logger.Printf("[ExamView] 部门未授权 examId=%d userId=%d deptId=%d", e.ID, uid, ud.DeptID)
				response.Fail(c, "您不在该考试的可见部门中")
				return
			}
		}
	}
	// 前置条件满足，创建或复用 session
	session := c.Query("session")
	if session == "" {
		session = fmt.Sprintf("%x", time.Now().UnixNano()+rand.Int63())
	}
	redisKey := fmt.Sprintf("exam_session:%d:%s", e.ID, session)
	now := time.Now().UnixMilli()
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
				"id":                q.ID,
				"type":              q.Type,
				"title":             q.Title,
				"options":           q.Options,
				"score":             q.Score,
				"difficulty":        q.Difficulty,
				"category":          q.Category,
				"examCorrectAnswer": q.Answer,
				"examAnalysis":      q.Analysis,
			})
		}
		response.JSON(c, map[string]interface{}{
			"exam":      e,
			"paper":     p,
			"questions": safe,
			"startAt":   startAt,
			"session":   session,
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
	// 未开启显示答案时，从 schema 中移除答案相关字段
	if answerVisible, _ := settingsMap["answerVisible"].(bool); !answerVisible {
		if questions, ok := schMap["questions"].([]interface{}); ok {
			for _, q := range questions {
				if qm, ok := q.(map[string]interface{}); ok {
					delete(qm, "examCorrectAnswer")
					delete(qm, "examCorrect")
				}
			}
		}
	}
	// 未开启分析时移除解析字段
	if showAnalysis, _ := settingsMap["showAnalysis"].(bool); !showAnalysis {
		if questions, ok := schMap["questions"].([]interface{}); ok {
			for _, q := range questions {
				if qm, ok := q.(map[string]interface{}); ok {
					delete(qm, "examAnalysis")
				}
			}
		}
	}
	resp := map[string]interface{}{
		"id":          e.ID,
		"title":       e.Title,
		"description": e.Description,
		"visibility":  e.Visibility,
		"anonymous":   e.Anonymous,
		"showResult":  e.ShowResult,
		"showScore":   e.ShowScore,
		"duration":    e.Duration,
		"maxAttempts": e.MaxAttempts,
		"startTime":   e.StartTime,
		"endTime":     e.EndTime,
		"schema":      schMap,
		"settings":    settingsMap,
		"startAt":     startAt,
		"session":     session,
		"deptIds":     e.DeptIds,
		"mode":        e.Mode,
	}
	response.JSON(c, resp)
}

func checkExamLimit(e *model.Exam, uidStr string, device string, deviceId string, ip string) string {
	if e.MaxResponse > 0 {
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", e.ID).Count(&cnt)
		if int(cnt) >= e.MaxResponse {
			logger.Logger.Printf("[ExamCheckLimit] 答卷上限 examId=%d max=%d current=%d", e.ID, e.MaxResponse, cnt)
			return "已达最大答卷数"
		}
	}
	if e.MaxAttempts > 0 && uidStr != "" {
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ? AND `exam_r_status` >= 1", e.ID, uidStr).Count(&cnt)
		if int(cnt) >= e.MaxAttempts {
			logger.Logger.Printf("[ExamCheckLimit] 个人答题次数上限 examId=%d uid=%s max=%d current=%d", e.ID, uidStr, e.MaxAttempts, cnt)
			return "已达最大答题次数"
		}
	}
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
	if deviceLimit, _ := settingsMap["deviceLimit"].(float64); deviceLimit > 0 && deviceId != "" {
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_device_id` = ? AND `exam_r_status` >= 1", e.ID, deviceId).Count(&cnt)
		if int(cnt) >= int(deviceLimit) {
			logger.Logger.Printf("[ExamCheckLimit] 设备次数上限 examId=%d limit=%d current=%d deviceId=%s", e.ID, int(deviceLimit), cnt, deviceId)
			return "已达每台设备最大答题次数"
		}
	}
	if ipLimit, _ := settingsMap["ipLimit"].(float64); ipLimit > 0 && ip != "" {
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_add_ip` = ? AND `exam_r_status` >= 1", e.ID, ip).Count(&cnt)
		if int(cnt) >= int(ipLimit) {
			logger.Logger.Printf("[ExamCheckLimit] IP次数上限 examId=%d limit=%d current=%d ip=%s", e.ID, int(ipLimit), cnt, ip)
			return "已达每个IP最大答题次数"
		}
	}
	return ""
}

// Validate POST /exam/validate
// @Tags 客户端-考试
// @Summary 校验答案（必填项等）
// @Router /exam/validate [post]
func (h *ClientExamHandler) Validate(_ context.Context, c *app.RequestContext) {
	raw, _ := c.Body()
	var req struct {
		ExamID   int                    `json:"examId"`
		Answers  map[string]interface{} `json:"answers"`
		Device   string                 `json:"device"`
		DeviceID string                 `json:"deviceId"`
	}
	if err := json.Unmarshal(raw, &req); err != nil || req.ExamID == 0 {
		response.Fail(c, "参数错误")
		return
	}
	var e model.Exam
	if err := database.DB.Where("`exam_id` = ? AND `exam_status` = 1", req.ExamID).First(&e).Error; err != nil {
		logger.Logger.Printf("[ExamValidate] 考试不存在或未发布 examId=%d", req.ExamID)
		response.Fail(c, "考试不存在或未发布")
		return
	}
	uid := getUID(c)
	uidStr := ""
	if uid > 0 {
		uidStr = strconv.FormatUint(uint64(uid), 10)
	}
	clientIP := c.ClientIP()
	if msg := checkExamLimit(&e, uidStr, req.Device, req.DeviceID, clientIP); msg != "" {
		response.JSON(c, map[string]interface{}{"ok": false, "errors": []map[string]string{{"questionId": "", "message": msg}}})
		return
	}
	type fieldErr struct {
		QuestionID string `json:"questionId"`
		Message    string `json:"message"`
	}
	var errs []fieldErr
	if e.PaperID > 0 {
		// PaperID 模式跳过校验（ExamQuestion 无 Required 字段）
	} else {
		var schMap map[string]interface{}
		_ = json.Unmarshal([]byte(e.Schema), &schMap)
		questions, _ := schMap["questions"].([]interface{})
		for _, qRaw := range questions {
			q, ok := qRaw.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := q["id"].(string)
			if id == "" {
				continue
			}
			required, _ := q["required"].(bool)
			val, ok := req.Answers[id]
			if !ok || val == nil || fmt.Sprintf("%v", val) == "" {
				if required {
					errs = append(errs, fieldErr{QuestionID: id, Message: "此项为必填"})
				}
				continue
			}
			typ, _ := q["type"].(string)
			if typ == "judge" {
				s := fmt.Sprintf("%v", val)
				if s != "true" && s != "false" {
					errs = append(errs, fieldErr{QuestionID: id, Message: "判断题答案格式错误"})
				}
			}
		}
	}
	response.JSON(c, map[string]interface{}{"ok": len(errs) == 0, "errors": errs})
}

// Start GET /exam/start?examId=
// @Tags 客户端-考试
// @Summary 开始考试
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /exam/start [get]
func (h *ClientExamHandler) Start(_ context.Context, c *app.RequestContext) {
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
			DeviceID:   deviceId,
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
// @Tags 客户端-考试
// @Summary 保存答案
// @Param recordId formData int true "记录ID"
// @Param answers formData string true "答案JSON"
// @Success 200 {object} response.Resp
// @Router /exam/save_answer [post]
func (h *ClientExamHandler) SaveAnswer(_ context.Context, c *app.RequestContext) {
	uid := getUID(c)
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
// @Tags 客户端-考试
// @Summary 提交考试
// @Router /exam/submit [post]
func (h *ClientExamHandler) Submit(_ context.Context, c *app.RequestContext) {
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

	if req.RecordID > 0 {
		// ── PaperID 模式（记录式） ──
		uid := getUID(c)
		if uid == 0 {
			response.Fail(c, "未登录")
			return
		}
		var rec model.ExamRecord
		if err := database.DB.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", req.RecordID, uid).First(&rec).Error; err != nil {
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
		if err := database.DB.Model(&rec).Updates(updates).Error; err != nil {
			logger.Logger.Printf("[ExamSubmit] PaperID模式提交失败 examId=%d recordId=%d uid=%d err=%s", rec.ExamID, req.RecordID, uid, err.Error())
			response.Fail(c, "提交失败: "+err.Error())
			return
		}
		logger.Logger.Printf("[ExamSubmit] PaperID模式提交成功 examId=%d recordId=%d uid=%d score=%d fullScore=%d", rec.ExamID, req.RecordID, uid, res.TotalScore, res.FullScore)
		response.JSON(c, map[string]interface{}{
			"score": res.TotalScore, "fullScore": res.FullScore,
			"correctCnt": res.CorrectCnt, "manualCnt": res.ManualCount,
			"results": res.Results,
		})
		return
	}

	// ── Schema 模式（PaperID == 0） ──
	if req.ExamID == 0 || req.Session == "" {
		response.Fail(c, "参数错误")
		return
	}
	redisKey := fmt.Sprintf("exam_session:%d:%s", req.ExamID, req.Session)
	exist, _ := rd.RDB.Exists(rd.Ctx, redisKey).Result()
	if exist == 0 {
		logger.Logger.Printf("[ExamSubmit] 会话过期 examId=%d session=%s", req.ExamID, req.Session)
		response.Fail(c, "会话不存在或已过期")
		return
	}
	var e model.Exam
	if err := database.DB.Where("`exam_id` = ? AND `exam_status` = 1", req.ExamID).First(&e).Error; err != nil {
		logger.Logger.Printf("[ExamSubmit] Schema模式考试不存在或未发布 examId=%d", req.ExamID)
		response.Fail(c, "考试不存在或未发布")
		return
	}
	if e.PaperID > 0 {
		response.Fail(c, "请使用 recordId 模式提交")
		return
	}
	// 尝试解析用户登录态（submit 为公开路由，需手动检测）
	uidStr := ""
	auth := string(c.GetHeader("Authorization"))
	if auth != "" {
		rdKey := tokenutil.TokenAuthKey("user", auth)
		if jsonStr, err := rd.RDB.Get(rd.Ctx, rdKey).Result(); err == nil && jsonStr != "" {
			var userInfo struct {
				ID uint `json:"id"`
			}
			if json.Unmarshal([]byte(jsonStr), &userInfo) == nil && userInfo.ID > 0 {
				uidStr = fmt.Sprintf("%d", userInfo.ID)
			}
		}
	}
	// 从 Redis 读取开始时间
	sessionStart, _ := rd.RDB.Get(rd.Ctx, redisKey).Int64()
	clientIP = c.ClientIP()
	if msg := checkExamLimit(&e, uidStr, req.Device, req.DeviceID, clientIP); msg != "" {
		response.Fail(c, msg)
		return
	}
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Schema), &schMap)
	rawQuestions, _ := schMap["questions"].([]interface{})
	var qs []examPkg.Question
	idMap := make(map[string]uint) // schema string ID → sequential uint ID
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
	// 将 answers 的 key 从 schema string ID 映射为 sequential uint ID
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
	// 写入 ExamRecord
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
	if err := database.DB.Create(&rec).Error; err != nil {
		logger.Logger.Printf("[ExamSubmit] Schema模式持久化失败 examId=%d uid=%s err=%s", req.ExamID, uidStr, err.Error())
		response.Fail(c, "提交失败: "+err.Error())
		return
	}
	logger.Logger.Printf("[ExamSubmit] Schema模式提交成功 examId=%d uid=%s score=%d fullScore=%d ip=%s device=%s", req.ExamID, uidStr, res.TotalScore, res.FullScore, clientIP, req.Device)
	rd.RDB.Del(rd.Ctx, redisKey)
	response.JSON(c, map[string]interface{}{
		"score": res.TotalScore, "fullScore": res.FullScore,
		"correctCnt": res.CorrectCnt, "manualCnt": res.ManualCount,
		"results": res.Results,
	})
}

// Record GET /exam/record?id=
// @Tags 客户端-考试
// @Summary 查看考试记录
// @Param id query int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /exam/record [get]
func (h *ClientExamHandler) Record(_ context.Context, c *app.RequestContext) {
	uid := getUID(c)
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
// @Tags 客户端-考试
// @Summary 我的考试记录
// @Success 200 {object} response.Resp
// @Router /exam/my_records [get]
func (h *ClientExamHandler) MyRecords(_ context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	var list []model.ExamRecord
	database.DB.Where("`exam_r_user_id` = ?", uid).Order("`exam_r_id` DESC").Limit(50).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list})
}

// ResultBySession GET /exam/result?session=
// @Tags 客户端-考试
// @Summary 通过 session 查看考试结果
// @Router /exam/result [get]
func (h *ClientExamHandler) ResultBySession(_ context.Context, c *app.RequestContext) {
	session := c.Query("session")
	if session == "" {
		response.Fail(c, "参数错误")
		return
	}
	var rec model.ExamRecord
	if err := database.DB.Where("`exam_r_session` = ?", session).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	var e model.Exam
	database.DB.Where("`exam_id` = ?", rec.ExamID).First(&e)
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
	if settingsMap == nil {
		settingsMap = make(map[string]interface{})
	}
	if _, ok := settingsMap["answerVisible"]; !ok {
		settingsMap["answerVisible"] = true
	}
	var questions []map[string]interface{}
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Schema), &schMap)
	rawQuestions, _ := schMap["questions"].([]interface{})
	for _, raw := range rawQuestions {
		q, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := q["id"].(string)
		if id == "" {
			continue
		}
		item := map[string]interface{}{
			"id":    id,
			"type":  q["type"],
			"title": q["title"],
		}
		if showAnalysis, _ := settingsMap["showAnalysis"].(bool); showAnalysis {
			item["examCorrectAnswer"] = q["examCorrectAnswer"]
			item["examAnalysis"] = q["examAnalysis"]
		}
		questions = append(questions, item)
	}
	var answers map[string]interface{}
	if rec.Answers != "" {
		_ = json.Unmarshal([]byte(rec.Answers), &answers)
	}
	var results []examPkg.Result
	if rec.Result != "" {
		_ = json.Unmarshal([]byte(rec.Result), &results)
	}
	response.JSON(c, map[string]interface{}{
		"exam":      e,
		"record":    rec,
		"questions": questions,
		"answers":   answers,
		"results":   results,
		"settings":  settingsMap,
	})
}
