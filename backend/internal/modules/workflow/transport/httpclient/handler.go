package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowhttperror "wecheckin/backend/internal/modules/workflow/transport/httperror"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/response"
)

type RuntimeService interface {
	ListPublishedDefinitionsForStarter(context.Context, string) ([]workflowapp.PublishedDefinition, error)
	ListPublishedDefinitionCategories(context.Context) ([]string, error)
	GetPublishedDefinitionForStarter(context.Context, uint, string) (*workflowapp.PublishedDefinition, error)
	GetStartDraft(context.Context, uint, string) (*workflowapp.StartDraft, error)
	SaveStartDraft(context.Context, workflowapp.SaveStartDraftRequest) (*workflowapp.StartDraft, error)
	DeleteStartDraft(context.Context, uint, string) error
	StartInstance(context.Context, workflowapp.StartInstanceRequest) (*workflowdomain.State, error)
	CompleteTask(context.Context, workflowapp.CompleteTaskRequest) (*workflowdomain.State, error)
	WithdrawInstance(context.Context, workflowapp.WithdrawInstanceRequest) (*workflowdomain.State, error)
	CommentInstance(context.Context, workflowapp.CommentInstanceRequest) error
	RemindInstance(context.Context, workflowapp.RemindInstanceRequest) (*workflowapp.RemindInstanceResult, error)
	ReviseInstanceForm(context.Context, workflowapp.ReviseInstanceFormRequest) (*workflowdomain.State, error)
	DeleteMyInstance(context.Context, string, string) error
	GetMyOverview(context.Context, string) (*workflowapp.WorkflowOverview, error)
	ListMyInstances(context.Context, string, workflowapp.InstanceQuery) (*workflowapp.InstanceList, error)
	GetMyInstance(context.Context, string, string) (*workflowapp.InstanceDetail, error)
	ListMyTasks(context.Context, string, workflowapp.TaskQuery) (*workflowapp.TaskList, error)
}

type RuntimeHandler struct {
	service RuntimeService
}

func NewRuntimeHandler(service RuntimeService) *RuntimeHandler {
	return &RuntimeHandler{service: service}
}

type startInstanceBody struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	BusinessType      string                 `json:"businessType"`
	BusinessKey       string                 `json:"businessKey"`
	Variables         map[string]interface{} `json:"variables"`
	FormData          map[string]interface{} `json:"formData"`
}

type saveStartDraftBody struct {
	DefinitionVersion int                    `json:"definitionVersion"`
	FormData          map[string]interface{} `json:"formData"`
}

type completeTaskBody struct {
	Action             workflowdomain.TaskAction     `json:"action"`
	Comment            string                        `json:"comment"`
	Images             []workflowcore.FormAttachment `json:"images"`
	ReturnTargetNodeID string                        `json:"returnTargetNodeId"`
	Variables          map[string]interface{}        `json:"variables"`
	FormData           map[string]interface{}        `json:"formData"`
}

type withdrawInstanceBody struct {
	Reason string `json:"reason"`
}

type commentInstanceBody struct {
	Comment      string                                  `json:"comment"`
	Images       []workflowcore.FormAttachment           `json:"images"`
	Notification *workflowapp.CommentNotificationRequest `json:"notification"`
}

type remindInstanceBody struct {
	NodeID string `json:"nodeId"`
}

type reviseInstanceFormBody struct {
	ExpectedRevision int64                                        `json:"expectedRevision"`
	FormData         map[string]interface{}                       `json:"formData"`
	Reason           string                                       `json:"reason"`
	Notification     *workflowapp.FormRevisionNotificationRequest `json:"notification"`
}

type mutationResponse struct {
	InstanceID   string                 `json:"instanceId"`
	Status       string                 `json:"status"`
	Variables    map[string]interface{} `json:"variables"`
	FormData     map[string]interface{} `json:"formData"`
	FormRevision int64                  `json:"formRevision"`
	PendingTasks []mutationTask         `json:"pendingTasks"`
}

type mutationTask struct {
	ID         string `json:"id"`
	NodeID     string `json:"nodeId"`
	NodeName   string `json:"nodeName"`
	AssigneeID string `json:"assigneeId"`
	Status     string `json:"status"`
}

func (handler *RuntimeHandler) ListDefinitions(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	data, err := handler.service.ListPublishedDefinitionsForStarter(ctx, actorID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.list_definitions", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) ListDefinitionCategories(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedActorID(c); !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	data, err := handler.service.ListPublishedDefinitionCategories(ctx)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.list_categories", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) GetDefinition(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || definitionID == 0 {
		response.Fail(c, "流程定义无效")
		return
	}
	data, err := handler.service.GetPublishedDefinitionForStarter(ctx, uint(definitionID), actorID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.get_definition", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) GetStartDraft(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, ok := parseDefinitionID(c, "definitionId")
	if !ok {
		response.Fail(c, "流程定义无效")
		return
	}
	draft, err := handler.service.GetStartDraft(ctx, definitionID, actorID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.get_draft", err)
		return
	}
	response.JSON(c, draft)
}

func (handler *RuntimeHandler) SaveStartDraft(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, ok := parseDefinitionID(c, "definitionId")
	if !ok {
		response.Fail(c, "流程定义无效")
		return
	}
	var body saveStartDraftBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	draft, err := handler.service.SaveStartDraft(ctx, workflowapp.SaveStartDraftRequest{
		DefinitionID: definitionID, DefinitionVersion: body.DefinitionVersion,
		StarterID: actorID, FormData: body.FormData,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.save_draft", err)
		return
	}
	response.JSON(c, draft)
}

func (handler *RuntimeHandler) DeleteStartDraft(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, ok := parseDefinitionID(c, "definitionId")
	if !ok {
		response.Fail(c, "流程定义无效")
		return
	}
	if err := handler.service.DeleteStartDraft(ctx, definitionID, actorID); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.delete_draft", err)
		return
	}
	response.JSON(c, nil)
}

func (handler *RuntimeHandler) StartInstance(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	var body startInstanceBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	state, err := handler.service.StartInstance(ctx, workflowapp.StartInstanceRequest{
		DefinitionID: body.DefinitionID, DefinitionVersion: body.DefinitionVersion,
		BusinessType: body.BusinessType, BusinessKey: body.BusinessKey,
		StarterID: actorID, OperatorID: actorID,
		ClearStartDraft: true,
		Variables:       body.Variables, FormData: body.FormData,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.start", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) GetMyOverview(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	data, err := handler.service.GetMyOverview(ctx, actorID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.overview", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) ListMyInstances(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, _ := strconv.ParseUint(c.Query("definitionId"), 10, 64)
	data, err := handler.service.ListMyInstances(ctx, actorID, workflowapp.InstanceQuery{
		DefinitionID:       uint(definitionID),
		DefinitionName:     strings.TrimSpace(c.Query("definitionName")),
		DefinitionCategory: strings.TrimSpace(c.Query("definitionCategory")),
		StarterName:        strings.TrimSpace(c.Query("starterName")),
		Status:             strings.TrimSpace(c.Query("status")),
		BusinessType:       strings.TrimSpace(c.Query("businessType")),
		BusinessKey:        strings.TrimSpace(c.Query("businessKey")),
		Scope:              strings.TrimSpace(c.Query("scope")),
		StartTimeFrom:      queryInt64(c, "startTimeFrom"),
		StartTimeTo:        queryInt64(c, "startTimeTo"),
		EndTimeFrom:        queryInt64(c, "endTimeFrom"),
		EndTimeTo:          queryInt64(c, "endTimeTo"),
		Page:               queryInt(c, "page"),
		PageSize:           queryInt(c, "pageSize"),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.list_instances", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) GetMyInstance(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	data, err := handler.service.GetMyInstance(ctx, actorID, instanceID)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.get_instance", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) DeleteMyInstance(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	if err := handler.service.DeleteMyInstance(ctx, actorID, instanceID); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.delete_instance", err)
		return
	}
	response.JSON(c, nil)
}

func (handler *RuntimeHandler) WithdrawInstance(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	var body withdrawInstanceBody
	if len(c.Request.Body()) > 0 {
		if err := json.Unmarshal(c.Request.Body(), &body); err != nil {
			response.Fail(c, "请求参数格式无效")
			return
		}
	}
	state, err := handler.service.WithdrawInstance(ctx, workflowapp.WithdrawInstanceRequest{
		InstanceID: instanceID, ActorID: actorID, Reason: body.Reason,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.withdraw", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) CommentInstance(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	var body commentInstanceBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if err := handler.service.CommentInstance(ctx, workflowapp.CommentInstanceRequest{
		InstanceID: instanceID, ActorID: actorID, Comment: body.Comment, Images: body.Images,
		Notification: body.Notification,
	}); err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.comment", err)
		return
	}
	response.JSON(c, nil)
}

func (handler *RuntimeHandler) RemindInstance(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	var body remindInstanceBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	result, err := handler.service.RemindInstance(ctx, workflowapp.RemindInstanceRequest{
		InstanceID: instanceID, ActorID: actorID, NodeID: body.NodeID,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.remind", err)
		return
	}
	response.JSON(c, result)
}

func (handler *RuntimeHandler) ReviseInstanceForm(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	instanceID := strings.TrimSpace(c.Param("id"))
	if instanceID == "" {
		response.Fail(c, "流程实例不能为空")
		return
	}
	var body reviseInstanceFormBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	state, err := handler.service.ReviseInstanceForm(ctx, workflowapp.ReviseInstanceFormRequest{
		InstanceID: instanceID, ActorID: actorID, ExpectedRevision: body.ExpectedRevision,
		FormData: body.FormData, Reason: body.Reason, Notification: body.Notification,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.revise_form", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) ListMyTasks(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	data, err := handler.service.ListMyTasks(ctx, actorID, workflowapp.TaskQuery{
		InstanceID:         strings.TrimSpace(c.Query("instanceId")),
		Status:             strings.TrimSpace(c.Query("status")),
		DefinitionName:     strings.TrimSpace(c.Query("definitionName")),
		DefinitionCategory: strings.TrimSpace(c.Query("definitionCategory")),
		StarterName:        strings.TrimSpace(c.Query("starterName")),
		StartTimeFrom:      queryInt64(c, "startTimeFrom"),
		StartTimeTo:        queryInt64(c, "startTimeTo"),
		Page:               queryInt(c, "page"),
		PageSize:           queryInt(c, "pageSize"),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.list_tasks", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) CompleteTask(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		response.Fail(c, "流程任务不能为空")
		return
	}
	var body completeTaskBody
	if err := decodeJSONBody(c, &body); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	state, err := handler.service.CompleteTask(ctx, workflowapp.CompleteTaskRequest{
		TaskID: taskID, ActorID: actorID, Action: body.Action, Comment: body.Comment,
		Images: body.Images, ReturnTargetNodeID: body.ReturnTargetNodeID,
		Variables: body.Variables, FormData: body.FormData,
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.client.complete_task", err)
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func authenticatedActorID(c *app.RequestContext) (string, bool) {
	if value, ok := c.Get("user"); ok {
		if user, ok := value.(*model.User); ok && user != nil && user.ID > 0 {
			return strconv.FormatUint(uint64(user.ID), 10), true
		}
	}
	if user, ok := dingtalkh5session.CurrentUser(c); ok && user != nil && user.ID > 0 {
		return strconv.FormatUint(uint64(user.ID), 10), true
	}
	return "", false
}

func parseDefinitionID(c *app.RequestContext, key string) (uint, bool) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param(key)), 10, 64)
	return uint(value), err == nil && value > 0
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
		InstanceID: state.Instance.ID, Status: string(state.Instance.Status),
		Variables: state.Variables, FormData: state.FormData, FormRevision: state.Instance.FormRevision,
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
