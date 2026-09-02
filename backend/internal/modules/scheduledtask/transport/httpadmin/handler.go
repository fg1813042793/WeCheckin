package httpadmin

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	"wecheckin/backend/internal/modules/scheduledtask/application"
	"wecheckin/backend/pkg/response"
)

type WorkerLister interface {
	ListWorkers(context.Context) ([]application.WorkerHeartbeat, error)
}

type RiskAuthorizer interface {
	HasPermission(context.Context, *model.Admin, string) (bool, error)
}

type Handler struct {
	service    *application.Service
	workers    WorkerLister
	authorizer RiskAuthorizer
}

func NewHandler(service *application.Service, workers WorkerLister, authorizer RiskAuthorizer) *Handler {
	return &Handler{service: service, workers: workers, authorizer: authorizer}
}

type taskStatusBody struct {
	Enabled bool  `json:"enabled"`
	Version int64 `json:"version"`
}

type cronPreviewBody struct {
	Expression  string `json:"expression"`
	Precision   string `json:"precision"`
	Timezone    string `json:"timezone"`
	Count       int    `json:"count"`
	AfterMillis int64  `json:"afterMillis"`
}

func (handler *Handler) ListTasks(ctx context.Context, c *app.RequestContext) {
	query := application.TaskQuery{
		Keyword: c.Query("keyword"), HandlerType: c.Query("handlerType"),
		Page: queryInt(c, "page"), PageSize: queryInt(c, "pageSize"),
	}
	if raw := strings.TrimSpace(c.Query("enabled")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			response.Fail(c, "启用状态参数无效")
			return
		}
		query.Enabled = &value
	}
	data, err := handler.service.ListTasks(ctx, query)
	respond(c, data, err)
}

func (handler *Handler) CreateTask(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	var request application.CreateTaskRequest
	if err := decodeJSONBody(c, &request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if !handler.authorizeRisk(ctx, c, admin, request.HandlerType, request.HandlerConfigJSON) {
		return
	}
	data, err := handler.service.CreateTask(ctx, uint64(admin.ID), request)
	respond(c, data, err)
}

func (handler *Handler) GetTask(ctx context.Context, c *app.RequestContext) {
	id, ok := taskID(c)
	if !ok {
		response.Fail(c, "定时任务 ID 无效")
		return
	}
	data, err := handler.service.GetTask(ctx, id)
	respond(c, data, err)
}

func (handler *Handler) UpdateTask(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := taskID(c)
	if !ok {
		response.Fail(c, "定时任务 ID 无效")
		return
	}
	var request application.UpdateTaskRequest
	if err := decodeJSONBody(c, &request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if !handler.authorizeRisk(ctx, c, admin, request.HandlerType, request.HandlerConfigJSON) {
		return
	}
	data, err := handler.service.UpdateTask(ctx, id, uint64(admin.ID), request)
	respond(c, data, err)
}

func (handler *Handler) DeleteTask(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := taskID(c)
	if !ok {
		response.Fail(c, "定时任务 ID 无效")
		return
	}
	err := handler.service.DeleteTask(ctx, id, uint64(admin.ID))
	respond(c, nil, err)
}

func (handler *Handler) SetTaskStatus(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := taskID(c)
	if !ok {
		response.Fail(c, "定时任务 ID 无效")
		return
	}
	var body taskStatusBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	data, err := handler.service.SetTaskEnabled(ctx, id, uint64(admin.ID), body.Enabled, body.Version)
	respond(c, data, err)
}

func (handler *Handler) RunTask(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	id, ok := taskID(c)
	if !ok {
		response.Fail(c, "定时任务 ID 无效")
		return
	}
	data, err := handler.service.RunNow(ctx, id, uint64(admin.ID))
	respond(c, data, err)
}

func (handler *Handler) PreviewCron(_ context.Context, c *app.RequestContext) {
	var body cronPreviewBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	after := time.Time{}
	if body.AfterMillis > 0 {
		after = time.UnixMilli(body.AfterMillis).UTC()
	}
	data, err := handler.service.PreviewCron(application.CronPreviewRequest{
		Expression: body.Expression, Precision: body.Precision, Timezone: body.Timezone,
		Count: body.Count, After: after,
	})
	respond(c, data, err)
}

func (handler *Handler) ListHandlers(_ context.Context, c *app.RequestContext) {
	response.JSON(c, handler.service.HandlerMetadata())
}

func (handler *Handler) ListRuns(ctx context.Context, c *app.RequestContext) {
	taskIDValue, _ := strconv.ParseUint(c.Query("taskId"), 10, 64)
	data, err := handler.service.ListRuns(ctx, application.RunQuery{
		TaskID: taskIDValue, Status: c.Query("status"), TriggerType: c.Query("triggerType"),
		WorkerID: c.Query("workerId"), StartTime: queryInt64(c, "startTime"), EndTime: queryInt64(c, "endTime"),
		Page: queryInt(c, "page"), PageSize: queryInt(c, "pageSize"),
	})
	respond(c, data, err)
}

func (handler *Handler) GetRun(ctx context.Context, c *app.RequestContext) {
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		response.Fail(c, "运行记录 ID 无效")
		return
	}
	data, err := handler.service.GetRunDetail(ctx, runID)
	respond(c, data, err)
}

func (handler *Handler) RetryRun(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		response.Fail(c, "运行记录 ID 无效")
		return
	}
	data, err := handler.service.RetryRun(ctx, runID, uint64(admin.ID))
	respond(c, data, err)
}

func (handler *Handler) CancelRun(ctx context.Context, c *app.RequestContext) {
	admin, ok := currentAdmin(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		response.Fail(c, "运行记录 ID 无效")
		return
	}
	data, err := handler.service.CancelRun(ctx, runID, uint64(admin.ID))
	respond(c, data, err)
}

func (handler *Handler) ListWorkers(ctx context.Context, c *app.RequestContext) {
	if handler.workers == nil {
		response.JSON(c, []application.WorkerHeartbeat{})
		return
	}
	data, err := handler.workers.ListWorkers(ctx)
	respond(c, data, err)
}

func (handler *Handler) authorizeRisk(ctx context.Context, c *app.RequestContext, admin *model.Admin, handlerType string, raw json.RawMessage) bool {
	permission := riskPermission(handlerType, raw)
	if permission == "" {
		return true
	}
	if handler.authorizer == nil {
		response.Fail(c, "高风险任务权限校验不可用")
		return false
	}
	allowed, err := handler.authorizer.HasPermission(ctx, admin, permission)
	if err != nil {
		response.Fail(c, "高风险任务权限校验失败")
		return false
	}
	if !allowed {
		response.Fail(c, "无权限配置该任务处理器")
		return false
	}
	return true
}

func riskPermission(handlerType string, raw json.RawMessage) string {
	switch strings.TrimSpace(handlerType) {
	case scheduledtaskmodel.HandlerTypeHTTP:
		return "scheduled-task:http"
	case scheduledtaskmodel.HandlerTypeShell:
		return "scheduled-task:shell"
	case scheduledtaskmodel.HandlerTypeSQL:
		var value struct {
			Mode string `json:"mode"`
		}
		if json.Unmarshal(raw, &value) == nil && strings.TrimSpace(value.Mode) == "write" {
			return "scheduled-task:sql:write"
		}
		return "scheduled-task:sql:read"
	default:
		return ""
	}
}

func respond(c *app.RequestContext, data interface{}, err error) {
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func currentAdmin(c *app.RequestContext) (*model.Admin, bool) {
	value, ok := c.Get("admin")
	if !ok {
		return nil, false
	}
	admin, ok := value.(*model.Admin)
	return admin, ok && admin != nil && admin.ID > 0
}

func taskID(c *app.RequestContext) (uint64, bool) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	return value, err == nil && value > 0
}

func decodeJSONBody(c *app.RequestContext, target interface{}) error {
	if c == nil || len(c.Request.Body()) == 0 {
		return errors.New("请求体不能为空")
	}
	return json.Unmarshal(c.Request.Body(), target)
}

func queryInt(c *app.RequestContext, key string) int {
	value, _ := strconv.Atoi(c.Query(key))
	return value
}

func queryInt64(c *app.RequestContext, key string) int64 {
	value, _ := strconv.ParseInt(c.Query(key), 10, 64)
	return value
}
