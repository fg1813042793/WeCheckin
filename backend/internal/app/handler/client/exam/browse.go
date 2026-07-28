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

	examservice "wecheckin-backend/backend/internal/app/service/exam"
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
	deviceId := c.Query("deviceId")
	clientIP := c.ClientIP()
	list, total, serviceLimits, err := h.service().PublishedListWithLimitsContext(ctx, c.Query("keyword"), page, pageSize, deviceId, clientIP)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	limitsMap := make(map[uint]examLimitInfo, len(serviceLimits))
	for id, limit := range serviceLimits {
		limitsMap[id] = examLimitInfo{DeviceFull: limit.DeviceFull, IPFull: limit.IPFull}
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
	e, err := h.service().PublishedExamContext(ctx, uint(id))
	if err != nil {
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
			deptID, _ := h.service().UserDeptIDContext(ctx, uid)
			deptIds := strings.Split(e.DeptIds, ",")
			allowed := false
			for _, did := range deptIds {
				d, _ := strconv.Atoi(strings.TrimSpace(did))
				if uint(d) == deptID {
					allowed = true
					break
				}
			}
			if !allowed {
				logger.Logger.Printf("[ExamView] 部门未授权 examId=%d userId=%d deptId=%d", e.ID, uid, deptID)
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
		paperResult, err := h.service().PaperQuestionsContext(ctx, e.PaperID, examservice.PaperQuestionOptions{
			IncludeExamAnswer:   true,
			IncludeExamAnalysis: true,
			IncludeCategory:     true,
			IncludeDifficulty:   true,
		})
		if err != nil {
			response.Fail(c, "试卷不存在")
			return
		}
		response.JSON(c, examViewPaperResponse{Exam: *e, Paper: paperResult.Paper, Questions: paperResult.Questions, StartAt: startAt, Session: session})
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
