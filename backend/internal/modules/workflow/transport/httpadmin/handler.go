package httpadmin

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/workflow/adminscope"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowhttperror "wecheckin/backend/internal/modules/workflow/transport/httperror"
	"wecheckin/backend/pkg/response"
)

type RuntimeService interface {
	ListPublishedDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error)
	GetPublishedDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error)
	StartInstance(context.Context, workflowapp.StartInstanceRequest) (*workflowdomain.State, error)
	CompleteTask(context.Context, workflowapp.CompleteTaskRequest) (*workflowdomain.State, error)
	CancelInstance(context.Context, workflowapp.CancelInstanceRequest) (*workflowdomain.State, error)
	ResumeTimers(context.Context, string, string) (*workflowdomain.State, int, error)
	ListInstances(context.Context, workflowapp.InstanceQuery) (*workflowapp.InstanceList, error)
	GetInstance(context.Context, string) (*workflowapp.InstanceDetail, error)
	DeleteInstances(context.Context, string, []string) (int, error)
	ListTasks(context.Context, workflowapp.TaskQuery) (*workflowapp.TaskList, error)
	DeleteTask(context.Context, string, string) error
	ListNotifications(context.Context, workflowapp.NotificationQuery) (*workflowapp.NotificationList, error)
	RetryNotification(context.Context, string) error
	SendNotification(context.Context, string) error
	DeleteNotification(context.Context, string, string) error
	DispatchDueNotifications(context.Context, int) (int, error)
}

type RuntimeHandler struct {
	service RuntimeService
	access  ScopeResolver
}

type ScopeResolver interface {
	Resolve(context.Context, *model.Admin) (adminscope.Scope, error)
}

func NewRuntimeHandlerWithAccess(service RuntimeService, access ScopeResolver) *RuntimeHandler {
	return &RuntimeHandler{service: service, access: access}
}

type startInstanceBody struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	BusinessType      string                 `json:"businessType"`
	BusinessKey       string                 `json:"businessKey"`
	StarterID         string                 `json:"starterId"`
	Variables         map[string]interface{} `json:"variables"`
	FormData          map[string]interface{} `json:"formData"`
}

type completeTaskBody struct {
	Action             workflowdomain.TaskAction `json:"action"`
	Comment            string                    `json:"comment"`
	ReturnTargetNodeID string                    `json:"returnTargetNodeId"`
	Variables          map[string]interface{}    `json:"variables"`
	FormData           map[string]interface{}    `json:"formData"`
}

type cancelInstanceBody struct {
	Reason string `json:"reason"`
}

type dispatchDueNotificationsBody struct {
	Limit int `json:"limit"`
}

type deleteInstancesBody struct {
	IDs []string `json:"ids"`
}

type mutationResponse struct {
	InstanceID   string                 `json:"instanceId"`
	Status       string                 `json:"status"`
	Variables    map[string]interface{} `json:"variables"`
	FormData     map[string]interface{} `json:"formData"`
	PendingTasks []mutationTask         `json:"pendingTasks"`
}

type mutationTask struct {
	ID         string `json:"id"`
	NodeID     string `json:"nodeId"`
	NodeName   string `json:"nodeName"`
	AssigneeID string `json:"assigneeId"`
	Status     string `json:"status"`
}

type resumeTimersResponse struct {
	mutationResponse
	Advanced int `json:"advanced"`
}

func (handler *RuntimeHandler) ListDefinitions(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	data, err := handler.service.ListPublishedDefinitions(ctx)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.list_definitions", err)
		return
	}
	response.JSON(c, filterDefinitions(data, scope))
}

func (handler *RuntimeHandler) GetDefinition(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "流程定义不能为空")
		return
	}
	if !scope.AllowsDefinition(uint(id)) {
		workflowhttperror.Respond(ctx, c, "workflow.admin.get_definition", workflowapp.ErrInstanceAccessDenied)
		return
	}
	data, err := handler.service.GetPublishedDefinition(ctx, uint(id))
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.get_definition", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) StartInstance(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	var body startInstanceBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if !scope.AllowsDefinition(body.DefinitionID) {
		workflowhttperror.Respond(ctx, c, "workflow.admin.start", workflowapp.ErrInstanceAccessDenied)
		return
	}
	state, err := handler.service.StartInstance(ctx, workflowapp.StartInstanceRequest{
		DefinitionID:      body.DefinitionID,
		DefinitionVersion: body.DefinitionVersion,
		BusinessType:      body.BusinessType,
		BusinessKey:       body.BusinessKey,
		StarterID:         body.StarterID,
		OperatorID:        actorID,
		AdminInitiated:    true,
		Variables:         body.Variables,
		FormData:          body.FormData,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.start", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) CompleteTask(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		response.Fail(c, "流程任务不能为空")
		return
	}
	if err := handler.ensureTasksVisible(ctx, []string{taskID}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.complete_task", err)
		return
	}
	var body completeTaskBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	state, err := handler.service.CompleteTask(ctx, workflowapp.CompleteTaskRequest{
		TaskID: taskID, ActorID: actorID, Action: body.Action,
		Comment: body.Comment, ReturnTargetNodeID: body.ReturnTargetNodeID,
		Variables: body.Variables, FormData: body.FormData,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.complete_task", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) CancelInstance(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	if err := handler.ensureInstancesVisible(ctx, []string{instanceID}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.cancel", err)
		return
	}
	var body cancelInstanceBody
	if len(c.Request.Body()) > 0 {
		if err := json.Unmarshal(c.Request.Body(), &body); err != nil {
			response.Fail(c, "请求参数格式无效")
			return
		}
	}
	state, err := handler.service.CancelInstance(ctx, workflowapp.CancelInstanceRequest{
		InstanceID: instanceID, ActorID: actorID, Reason: body.Reason,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.cancel", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) ResumeTimers(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	if err := handler.ensureInstancesVisible(ctx, []string{instanceID}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.resume_timers", err)
		return
	}
	state, advanced, err := handler.service.ResumeTimers(ctx, instanceID, actorID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.resume_timers", err)
		return
	}
	response.JSON(c, resumeTimersResponse{mutationResponse: newMutationResponse(state), Advanced: advanced})
}

func (handler *RuntimeHandler) ListInstances(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	definitionID, _ := strconv.ParseUint(c.Query("definitionId"), 10, 64)
	visibility := scope.Visibility
	data, err := handler.service.ListInstances(ctx, workflowapp.InstanceQuery{
		DefinitionID:       uint(definitionID),
		DefinitionName:     strings.TrimSpace(c.Query("definitionName")),
		InstanceTitle:      strings.TrimSpace(c.Query("instanceTitle")),
		BusinessPeriodKey:  strings.TrimSpace(c.Query("businessPeriodKey")),
		DefinitionCategory: strings.TrimSpace(c.Query("definitionCategory")),
		Status:             strings.TrimSpace(c.Query("status")),
		BusinessType:       strings.TrimSpace(c.Query("businessType")),
		BusinessKey:        strings.TrimSpace(c.Query("businessKey")),
		StarterID:          strings.TrimSpace(c.Query("starterId")),
		StartTimeFrom:      queryInt64(c, "startTimeFrom"),
		StartTimeTo:        queryInt64(c, "startTimeTo"),
		EndTimeFrom:        queryInt64(c, "endTimeFrom"),
		EndTimeTo:          queryInt64(c, "endTimeTo"),
		Visibility:         &visibility,
		Page:               queryInt(c, "page"),
		PageSize:           queryInt(c, "pageSize"),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.list_instances", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) GetInstance(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	if err := handler.ensureInstancesVisible(ctx, []string{instanceID}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.get_instance", err)
		return
	}
	data, err := handler.service.GetInstance(ctx, instanceID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.get_instance", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) DeleteInstance(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	if err := handler.ensureInstancesVisible(ctx, []string{instanceID}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_instance", err)
		return
	}
	deleted, err := handler.service.DeleteInstances(ctx, actorID, []string{instanceID})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_instance", err)
		return
	}
	response.JSON(c, map[string]int{"deleted": deleted})
}

func (handler *RuntimeHandler) DeleteInstances(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	var body deleteInstancesBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if err := handler.ensureInstancesVisible(ctx, body.IDs, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_instances", err)
		return
	}
	deleted, err := handler.service.DeleteInstances(ctx, actorID, body.IDs)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_instances", err)
		return
	}
	response.JSON(c, map[string]int{"deleted": deleted})
}

func (handler *RuntimeHandler) ListTasks(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	visibility := scope.Visibility
	data, err := handler.service.ListTasks(ctx, workflowapp.TaskQuery{
		InstanceID:       strings.TrimSpace(c.Query("instanceId")),
		AssigneeID:       strings.TrimSpace(c.Query("assigneeId")),
		Status:           strings.TrimSpace(c.Query("status")),
		HideAdminDeleted: true,
		Visibility:       &visibility,
		Page:             queryInt(c, "page"),
		PageSize:         queryInt(c, "pageSize"),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.list_tasks", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) DeleteTask(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		response.Fail(c, "流程任务不能为空")
		return
	}
	if err := handler.ensureTasksVisible(ctx, []string{taskID}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_task", err)
		return
	}
	if err := handler.service.DeleteTask(ctx, actorID, taskID); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_task", err)
		return
	}
	response.JSON(c, map[string]string{"id": taskID})
}

func (handler *RuntimeHandler) ListNotifications(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	visibility := scope.Visibility
	data, err := handler.service.ListNotifications(ctx, workflowapp.NotificationQuery{
		InstanceID:      strings.TrimSpace(c.Query("instanceId")),
		RecipientUserID: strings.TrimSpace(c.Query("recipientUserId")),
		Kind:            strings.TrimSpace(c.Query("kind")),
		Channel:         strings.TrimSpace(c.Query("channel")),
		Status:          strings.TrimSpace(c.Query("status")),
		Visibility:      &visibility,
		Page:            queryInt(c, "page"),
		PageSize:        queryInt(c, "pageSize"),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.list_notifications", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) RetryNotification(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Fail(c, "通知投递记录不能为空")
		return
	}
	if err := handler.ensureNotificationsVisible(ctx, []string{id}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.retry_notification", err)
		return
	}
	if err := handler.service.RetryNotification(ctx, id); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.retry_notification", err)
		return
	}
	response.JSON(c, map[string]string{"id": id})
}

func (handler *RuntimeHandler) SendNotification(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Fail(c, "通知投递记录不能为空")
		return
	}
	if err := handler.ensureNotificationsVisible(ctx, []string{id}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.send_notification", err)
		return
	}
	if err := handler.service.SendNotification(ctx, id); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.send_notification", err)
		return
	}
	response.JSON(c, map[string]string{"id": id})
}

func (handler *RuntimeHandler) DeleteNotification(ctx context.Context, c *app.RequestContext) {
	admin, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	actorID := strconv.FormatUint(uint64(admin.ID), 10)
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Fail(c, "通知投递记录不能为空")
		return
	}
	if err := handler.ensureNotificationsVisible(ctx, []string{id}, scope); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_notification", err)
		return
	}
	if err := handler.service.DeleteNotification(ctx, actorID, id); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.delete_notification", err)
		return
	}
	response.JSON(c, map[string]string{"id": id})
}

func (handler *RuntimeHandler) DispatchDueNotifications(ctx context.Context, c *app.RequestContext) {
	_, scope, ok := handler.resolveScope(ctx, c)
	if !ok {
		return
	}
	if !scope.Visibility.All {
		workflowhttperror.Respond(ctx, c, "workflow.admin.dispatch_notifications", workflowapp.ErrInstanceAccessDenied)
		return
	}
	var body dispatchDueNotificationsBody
	if len(c.Request.Body()) > 0 {
		if err := json.Unmarshal(c.Request.Body(), &body); err != nil {
			response.Fail(c, "请求参数格式无效")
			return
		}
	}
	count, err := handler.service.DispatchDueNotifications(ctx, body.Limit)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.dispatch_notifications", err)
		return
	}
	response.JSON(c, map[string]int{"dispatched": count})
}

func (handler *RuntimeHandler) resolveScope(ctx context.Context, c *app.RequestContext) (*model.Admin, adminscope.Scope, bool) {
	value, ok := c.Get("admin")
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return nil, adminscope.Scope{}, false
	}
	admin, ok := value.(*model.Admin)
	if !ok || admin == nil || admin.ID == 0 {
		response.Fail(c, "未登录或权限失效")
		return nil, adminscope.Scope{}, false
	}
	if handler.access == nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.resolve_scope", errors.New("工作流数据权限服务未初始化"))
		return nil, adminscope.Scope{}, false
	}
	scope, err := handler.access.Resolve(ctx, admin)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.admin.resolve_scope", err)
		return nil, adminscope.Scope{}, false
	}
	return admin, scope, true
}

func filterDefinitions(definitions []workflowapp.PublishedDefinition, scope adminscope.Scope) []workflowapp.PublishedDefinition {
	if scope.AllDefinitions {
		return definitions
	}
	result := make([]workflowapp.PublishedDefinition, 0, len(definitions))
	for _, definition := range definitions {
		if scope.AllowsDefinition(definition.ID) {
			result = append(result, definition)
		}
	}
	return result
}

func (handler *RuntimeHandler) ensureInstancesVisible(ctx context.Context, ids []string, scope adminscope.Scope) error {
	if scope.Visibility.All {
		return nil
	}
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return workflowapp.ErrInstanceIDRequired
	}
	visibility := scope.Visibility
	list, err := handler.service.ListInstances(ctx, workflowapp.InstanceQuery{
		InstanceIDs: ids,
		Visibility:  &visibility,
		Page:        1,
		PageSize:    len(ids),
	})
	if err != nil {
		return err
	}
	if list == nil || len(list.List) != len(ids) {
		return workflowapp.ErrInstanceAccessDenied
	}
	return nil
}

func (handler *RuntimeHandler) ensureTasksVisible(ctx context.Context, ids []string, scope adminscope.Scope) error {
	if scope.Visibility.All {
		return nil
	}
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return workflowapp.ErrTaskIDRequired
	}
	visibility := scope.Visibility
	list, err := handler.service.ListTasks(ctx, workflowapp.TaskQuery{
		IDs:                  ids,
		Visibility:           &visibility,
		HideAdminDeleted:     true,
		IncludeFormRevisions: true,
		Page:                 1,
		PageSize:             len(ids),
	})
	if err != nil {
		return err
	}
	if list == nil || len(list.List) != len(ids) {
		return workflowapp.ErrInstanceAccessDenied
	}
	return nil
}

func (handler *RuntimeHandler) ensureNotificationsVisible(ctx context.Context, ids []string, scope adminscope.Scope) error {
	if scope.Visibility.All {
		return nil
	}
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return workflowapp.ErrInstanceAccessDenied
	}
	visibility := scope.Visibility
	list, err := handler.service.ListNotifications(ctx, workflowapp.NotificationQuery{
		IDs:        ids,
		Visibility: &visibility,
		Page:       1,
		PageSize:   len(ids),
	})
	if err != nil {
		return err
	}
	if list == nil || len(list.List) != len(ids) {
		return workflowapp.ErrInstanceAccessDenied
	}
	return nil
}

func uniqueIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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

func newMutationResponse(state *workflowdomain.State) mutationResponse {
	if state == nil {
		return mutationResponse{}
	}
	result := mutationResponse{
		InstanceID: state.Instance.ID,
		Status:     string(state.Instance.Status),
		Variables:  state.Variables,
		FormData:   state.FormData,
	}
	for _, task := range state.PendingTasks() {
		result.PendingTasks = append(result.PendingTasks, mutationTask{
			ID: task.ID, NodeID: task.NodeID, NodeName: task.NodeName,
			AssigneeID: task.AssigneeID, Status: string(task.Status),
		})
	}
	if result.PendingTasks == nil {
		result.PendingTasks = make([]mutationTask, 0)
	}
	return result
}
