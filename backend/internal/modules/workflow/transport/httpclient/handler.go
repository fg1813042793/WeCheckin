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
	"wecheckin/backend/pkg/response"
)

type RuntimeService interface {
	ListPublishedDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error)
	GetPublishedDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error)
	StartInstance(context.Context, workflowapp.StartInstanceRequest) (*workflowdomain.State, error)
	CompleteTask(context.Context, workflowapp.CompleteTaskRequest) (*workflowdomain.State, error)
	WithdrawInstance(context.Context, workflowapp.WithdrawInstanceRequest) (*workflowdomain.State, error)
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

type completeTaskBody struct {
	Action    workflowdomain.TaskAction `json:"action"`
	Comment   string                    `json:"comment"`
	Variables map[string]interface{}    `json:"variables"`
	FormData  map[string]interface{}    `json:"formData"`
}

type withdrawInstanceBody struct {
	Reason string `json:"reason"`
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

func (handler *RuntimeHandler) ListDefinitions(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedActorID(c); !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	data, err := handler.service.ListPublishedDefinitions(ctx)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) GetDefinition(ctx context.Context, c *app.RequestContext) {
	if _, ok := authenticatedActorID(c); !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || definitionID == 0 {
		response.Fail(c, "流程定义无效")
		return
	}
	data, err := handler.service.GetPublishedDefinition(ctx, uint(definitionID))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
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
		BusinessType: body.BusinessType, BusinessKey: body.BusinessKey, StarterID: actorID,
		Variables: body.Variables, FormData: body.FormData,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func (handler *RuntimeHandler) ListMyInstances(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		response.Fail(c, "未登录或权限失效")
		return
	}
	definitionID, _ := strconv.ParseUint(c.Query("definitionId"), 10, 64)
	data, err := handler.service.ListMyInstances(ctx, actorID, workflowapp.InstanceQuery{
		DefinitionID: uint(definitionID), Status: strings.TrimSpace(c.Query("status")),
		BusinessType: strings.TrimSpace(c.Query("businessType")), BusinessKey: strings.TrimSpace(c.Query("businessKey")),
		Page: queryInt(c, "page"), PageSize: queryInt(c, "pageSize"),
	})
	if err != nil {
		response.Fail(c, err.Error())
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
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
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
		response.Fail(c, err.Error())
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
		InstanceID: strings.TrimSpace(c.Query("instanceId")), Status: strings.TrimSpace(c.Query("status")),
		Page: queryInt(c, "page"), PageSize: queryInt(c, "pageSize"),
	})
	if err != nil {
		response.Fail(c, err.Error())
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
		Variables: body.Variables, FormData: body.FormData,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, newMutationResponse(state))
}

func authenticatedActorID(c *app.RequestContext) (string, bool) {
	value, ok := c.Get("user")
	if !ok {
		return "", false
	}
	user, ok := value.(*model.User)
	if !ok || user == nil || user.ID == 0 {
		return "", false
	}
	return strconv.FormatUint(uint64(user.ID), 10), true
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

func newMutationResponse(state *workflowdomain.State) mutationResponse {
	if state == nil {
		return mutationResponse{}
	}
	result := mutationResponse{
		InstanceID: state.Instance.ID, Status: string(state.Instance.Status),
		Variables: state.Variables, FormData: state.FormData,
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
