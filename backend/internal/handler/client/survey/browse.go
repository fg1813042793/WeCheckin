package survey

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/pkg/logger"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/response"
	"wecheckin/backend/pkg/tokenutil"
)

// List GET /survey/list
// @Tags 客户端-问卷
// @Summary 获取问卷列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Success 200 {object} response.Resp
// @Router /survey/list [get]
func (h *ClientSurveyHandler) List(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	deviceId := c.Query("deviceId")
	clientIP := c.ClientIP()
	list, total, serviceLimits, err := h.survey.PublishedListWithLimitsContext(ctx, keyword, category, page, pageSize, deviceId, clientIP)
	if err != nil {
		response.FailInternal(ctx, c, "client.survey.browse", "查询失败，请稍后重试", err)
		return
	}
	limitsMap := make(map[uint]limitInfo, len(serviceLimits))
	for id, limit := range serviceLimits {
		limitsMap[id] = limitInfo{DeviceFull: limit.DeviceFull, IPFull: limit.IPFull}
	}
	response.JSON(c, listResponse{List: list, Total: total, Page: page, Size: pageSize, Limits: limitsMap})
}

// Detail GET /survey/view?id=
// @Tags 客户端-问卷
// @Summary 查看问卷详情
// @Param id query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /survey/view [get]
func (h *ClientSurveyHandler) Detail(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	sv, err := h.survey.GetContext(ctx, uint(id))
	if err != nil {
		logger.Logger.Printf("[SurveyDetail] 问卷不存在 id=%d", id)
		response.Fail(c, "问卷不存在")
		return
	}
	if sv.Status != 1 {
		logger.Logger.Printf("[SurveyDetail] 问卷已停用 id=%d title=%s", id, sv.Title)
		response.Fail(c, "问卷已停用")
		return
	}
	if sv.Visibility == 1 || sv.Visibility == 2 {
		auth := c.GetHeader("Authorization")
		token := ""
		if len(auth) > 0 {
			token = string(auth)
		}
		if token == "" {
			logger.Logger.Printf("[SurveyDetail] 未登录 id=%d", id)
			response.Fail(c, "请先登录")
			return
		}
		rdKey := tokenutil.TokenAuthKeyContext(ctx, "user", token)
		redisCtx, cancel := rd.OperationContext(ctx)
		defer cancel()
		jsonStr, err := rd.RDB.Get(redisCtx, rdKey).Result()
		if err != nil || jsonStr == "" {
			logger.Logger.Printf("[SurveyDetail] token无效 id=%d", id)
			response.Fail(c, "请先登录")
			return
		}
		if sv.Visibility == 2 && sv.DeptIDs != "" {
			var userInfo map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &userInfo)
			uid := uint(0)
			if id, ok := userInfo["id"].(float64); ok {
				uid = uint(id)
			}
			deptID, _ := h.survey.UserDeptIDContext(ctx, uid)
			deptIds := strings.Split(sv.DeptIDs, ",")
			allowed := false
			for _, did := range deptIds {
				d, _ := strconv.Atoi(strings.TrimSpace(did))
				if uint(d) == deptID {
					allowed = true
					break
				}
			}
			if !allowed {
				logger.Logger.Printf("[SurveyDetail] 部门无权限 id=%d uid=%d deptId=%d", id, uid, deptID)
				response.Fail(c, "您不在该问卷的可见部门中")
				return
			}
		}
	}
	now := time.Now().UnixMilli()
	if sv.StartTime > 0 && now < sv.StartTime {
		logger.Logger.Printf("[SurveyDetail] 问卷未开始 id=%d startTime=%d", id, sv.StartTime)
		response.Fail(c, "问卷未开始")
		return
	}
	if sv.EndTime > 0 && now > sv.EndTime {
		logger.Logger.Printf("[SurveyDetail] 问卷已结束 id=%d endTime=%d", id, sv.EndTime)
		response.Fail(c, "问卷已结束")
		return
	}
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(sv.Schema), &schMap)
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(sv.Settings), &settingsMap)
	if raw, ok := settingsMap["logicRules"]; ok {
		var rules []interface{}
		switch v := raw.(type) {
		case string:
			json.Unmarshal([]byte(v), &rules)
		case []interface{}:
			rules = v
		}
		if len(rules) > 0 {
			filtered := make([]interface{}, 0, len(rules))
			for _, item := range rules {
				if m, ok := item.(map[string]interface{}); ok {
					if scope, _ := m["scope"].(string); scope == "backend" {
						continue
					}
				}
				filtered = append(filtered, item)
			}
			settingsMap["logicRules"] = filtered
		}
	}
	session := c.Query("session")
	if session == "" {
		session = fmt.Sprintf("%x", time.Now().UnixNano()+rand.Int63())
	}
	redisKey := fmt.Sprintf("survey_session:%d:%s", sv.ID, session)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
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
	logger.Logger.Printf("[SurveyDetail] 成功 id=%d title=%s", id, sv.Title)
	response.JSON(c, detailResponse{
		ID:          sv.ID,
		Title:       sv.Title,
		Description: sv.Desc,
		Category:    sv.Category,
		Cover:       sv.Cover,
		Visibility:  sv.Visibility,
		Anonymous:   sv.Anonymous,
		AllowMulti:  sv.AllowMulti,
		StartTime:   sv.StartTime,
		EndTime:     sv.EndTime,
		MaxResponse: sv.MaxResponse,
		ShowResult:  sv.ShowResult,
		Schema:      schMap,
		Settings:    settingsMap,
		Session:     session,
		StartAt:     startAt,
		DeptIDs:     sv.DeptIDs,
	})
}
