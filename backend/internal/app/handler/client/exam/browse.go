package exam

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

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/response"
	"wecheckin-backend/backend/pkg/tokenutil"
)

// List GET /exam/list
// @Tags 客户端-考试
// @Summary 获取考试列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /exam/list [get]
func (h *ClientExamHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Exam{}).Where("`exam_status` = 1")
	if kw := c.Query("keyword"); kw != "" {
		q = q.Where("`exam_title` LIKE ?", "%"+kw+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Exam
	q.Order("`exam_order` DESC, `exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	deviceId := c.Query("deviceId")
	clientIP := c.ClientIP()
	limitsMap := make(map[uint]examLimitInfo)
	for _, e := range list {
		li := examLimitInfo{}
		var settingsMap map[string]interface{}
		_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
		if deviceLimit, _ := settingsMap["deviceLimit"].(float64); deviceLimit > 0 && deviceId != "" {
			var cnt int64
			db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_device_id` = ? AND `exam_r_status` >= 1", e.ID, deviceId).Count(&cnt)
			if int(cnt) >= int(deviceLimit) {
				li.DeviceFull = true
			}
		}
		if ipLimit, _ := settingsMap["ipLimit"].(float64); ipLimit > 0 && clientIP != "" {
			var cnt int64
			db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_add_ip` = ? AND `exam_r_status` >= 1", e.ID, clientIP).Count(&cnt)
			if int(cnt) >= int(ipLimit) {
				li.IPFull = true
			}
		}
		if li.DeviceFull || li.IPFull {
			limitsMap[e.ID] = li
		}
	}
	response.JSON(c, examListResponse{List: list, Total: total, Page: page, Size: pageSize, Limits: limitsMap})
}

// View GET /exam/view?id=
// @Tags 客户端-考试
// @Summary 查看考试详情
// @Param id query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /exam/view [get]
func (h *ClientExamHandler) View(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id == 0 {
		response.Fail(c, "id 必填")
		return
	}
	var e model.Exam
	db, dbCancel := database.WithContext(ctx)
	defer dbCancel()
	if err := db.Where("`exam_id` = ? AND `exam_status` = 1", id).First(&e).Error; err != nil {
		logger.Logger.Printf("[ExamView] 考试不存在或未发布 examId=%d", id)
		response.Fail(c, "考试不存在或未发布")
		return
	}
	logger.Logger.Printf("[ExamView] 考试已发布 examId=%d title=%s visibility=%d", e.ID, e.Title, e.Visibility)
	nowMs := time.Now().UnixMilli()
	if e.StartTime > 0 && nowMs < e.StartTime {
		logger.Logger.Printf("[ExamView] 考试未开始 examId=%d startTime=%d", e.ID, e.StartTime)
		c.JSON(consts.StatusOK, struct {
			Code int                `json:"code"`
			Msg  string             `json:"msg"`
			Data examNotStartedData `json:"data"`
		}{
			Code: 1,
			Msg:  "考试未开始",
			Data: examNotStartedData{StartTime: e.StartTime, EndTime: e.EndTime, Title: e.Title},
		})
		return
	}
	if e.EndTime > 0 && nowMs > e.EndTime {
		logger.Logger.Printf("[ExamView] 考试已结束 examId=%d endTime=%d", e.ID, e.EndTime)
		response.Fail(c, "考试已结束")
		return
	}
	loginRequired := false
	if e.Visibility == 1 || e.Visibility == 2 {
		loginRequired = true
	} else {
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
		redisCtx, cancel := rd.OperationContext(ctx)
		defer cancel()
		jsonStr, err := rd.RDB.Get(redisCtx, rdKey).Result()
		if err != nil || jsonStr == "" {
			logger.Logger.Printf("[ExamView] token无效 examId=%d", e.ID)
			response.Fail(c, "请先登录")
			return
		}
		if e.Visibility == 2 && e.DeptIds != "" {
			var userInfo map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &userInfo)
			uid := uint(0)
			if id, ok := userInfo["id"].(float64); ok {
				uid = uint(id)
			}
			var ud model.UserDept
			db.Where("`user_dept_user_id` = ?", uid).First(&ud)
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
	session := c.Query("session")
	if session == "" {
		session = fmt.Sprintf("%x", time.Now().UnixNano()+rand.Int63())
	}
	redisKey := fmt.Sprintf("exam_session:%d:%s", e.ID, session)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	now := time.Now().UnixMilli()
	var startAt int64
	exists, _ := rd.RDB.Exists(redisCtx, redisKey).Result()
	if exists == 0 {
		rd.RDB.Set(redisCtx, redisKey, now, 24*time.Hour)
		startAt = now
	} else {
		v, err := rd.RDB.Get(redisCtx, redisKey).Int64()
		if err == nil {
			startAt = v
		} else {
			startAt = now
		}
	}
	if e.PaperID > 0 {
		var p model.ExamPaper
		if err := db.Where("`exam_p_id` = ?", e.PaperID).First(&p).Error; err != nil {
			response.Fail(c, "试卷不存在")
			return
		}
		var qids []uint
		_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
		var qs []model.ExamQuestion
		if len(qids) > 0 {
			db.Where("`exam_q_id` IN ?", qids).Find(&qs)
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
		response.JSON(c, examViewPaperResponse{Exam: e, Paper: p, Questions: safe, StartAt: startAt, Session: session})
		return
	}
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
	if showAnalysis, _ := settingsMap["showAnalysis"].(bool); !showAnalysis {
		if questions, ok := schMap["questions"].([]interface{}); ok {
			for _, q := range questions {
				if qm, ok := q.(map[string]interface{}); ok {
					delete(qm, "examAnalysis")
				}
			}
		}
	}
	response.JSON(c, examViewSchemaResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Visibility:  e.Visibility,
		Anonymous:   e.Anonymous,
		ShowResult:  e.ShowResult,
		ShowScore:   e.ShowScore,
		Duration:    e.Duration,
		MaxAttempts: e.MaxAttempts,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		Schema:      schMap,
		Settings:    settingsMap,
		StartAt:     startAt,
		Session:     session,
		DeptIDs:     e.DeptIds,
		Mode:        e.Mode,
	})
}
