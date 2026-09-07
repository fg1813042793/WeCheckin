package httpclient

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowhttperror "wecheckin/backend/internal/modules/workflow/transport/httperror"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/response"
)

type previewCompletedFormRevisionBody struct {
	SourceNodeID     string                 `json:"sourceNodeId"`
	ExpectedRevision int64                  `json:"expectedRevision"`
	FormData         map[string]interface{} `json:"formData"`
}

type createCompletedFormRevisionBody struct {
	SourceNodeID     string                 `json:"sourceNodeId"`
	ExpectedRevision int64                  `json:"expectedRevision"`
	FormData         map[string]interface{} `json:"formData"`
	Reason           string                 `json:"reason"`
}

type completeFormRevisionTaskBody struct {
	Action  workflowdomain.RevisionAction `json:"action"`
	Comment string                        `json:"comment"`
	Images  []workflowcore.FormAttachment `json:"images"`
}

func (handler *RuntimeHandler) ListFormRevisions(ctx context.Context, c *app.RequestContext) {
	actorID, instanceID, ok := formRevisionRouteIdentity(c, "流程实例不能为空")
	if !ok {
		return
	}
	data, err := handler.service.ListFormRevisions(ctx, instanceID, actorID)
	if err != nil {
		workflowhttperror.RespondWithStatus(ctx, c, "workflow.client.list_form_revisions", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) PreviewCompletedFormRevision(ctx context.Context, c *app.RequestContext) {
	actorID, instanceID, ok := formRevisionRouteIdentity(c, "流程实例不能为空")
	if !ok {
		return
	}
	var body previewCompletedFormRevisionBody
	if err := decodeJSONBody(c, &body); err != nil {
		formRevisionBadRequest(c, "请求参数格式无效")
		return
	}
	data, err := handler.service.PreviewCompletedFormRevision(ctx, workflowapp.PreviewCompletedFormRevisionRequest{
		InstanceID: instanceID, SourceNodeID: body.SourceNodeID, RequesterID: actorID,
		ExpectedRevision: body.ExpectedRevision, FormData: body.FormData,
	})
	if err != nil {
		workflowhttperror.RespondWithStatus(ctx, c, "workflow.client.preview_completed_form_revision", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) CreateCompletedFormRevision(ctx context.Context, c *app.RequestContext) {
	actorID, instanceID, ok := formRevisionRouteIdentity(c, "流程实例不能为空")
	if !ok {
		return
	}
	var body createCompletedFormRevisionBody
	if err := decodeJSONBody(c, &body); err != nil {
		formRevisionBadRequest(c, "请求参数格式无效")
		return
	}
	data, err := handler.service.CreateCompletedFormRevision(ctx, workflowapp.CreateCompletedFormRevisionRequest{
		InstanceID: instanceID, SourceNodeID: body.SourceNodeID, RequesterID: actorID,
		ExpectedRevision: body.ExpectedRevision, FormData: body.FormData, Reason: body.Reason,
	})
	if err != nil {
		workflowhttperror.RespondWithStatus(ctx, c, "workflow.client.create_completed_form_revision", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) GetFormRevision(ctx context.Context, c *app.RequestContext) {
	actorID, revisionID, ok := formRevisionRouteIdentity(c, "表单修订请求不能为空")
	if !ok {
		return
	}
	data, err := handler.service.GetFormRevision(ctx, revisionID, actorID)
	if err != nil {
		workflowhttperror.RespondWithStatus(ctx, c, "workflow.client.get_form_revision", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) CancelCompletedFormRevision(ctx context.Context, c *app.RequestContext) {
	actorID, revisionID, ok := formRevisionRouteIdentity(c, "表单修订请求不能为空")
	if !ok {
		return
	}
	data, err := handler.service.CancelCompletedFormRevision(ctx, workflowapp.CancelCompletedFormRevisionRequest{
		RevisionID: revisionID, ActorID: actorID,
	})
	if err != nil {
		workflowhttperror.RespondWithStatus(ctx, c, "workflow.client.cancel_completed_form_revision", err)
		return
	}
	response.JSON(c, data)
}

func (handler *RuntimeHandler) CompleteFormRevisionTask(ctx context.Context, c *app.RequestContext) {
	actorID, taskID, ok := formRevisionRouteIdentity(c, "表单修订任务不能为空")
	if !ok {
		return
	}
	var body completeFormRevisionTaskBody
	if err := decodeJSONBody(c, &body); err != nil {
		formRevisionBadRequest(c, "请求参数格式无效")
		return
	}
	if body.Action != workflowdomain.RevisionActionApprove && body.Action != workflowdomain.RevisionActionReject {
		formRevisionBadRequest(c, "表单修订任务操作无效")
		return
	}
	data, err := handler.service.CompleteFormRevisionTask(ctx, workflowapp.CompleteFormRevisionTaskRequest{
		TaskID: taskID, ActorID: actorID, Action: body.Action, Comment: body.Comment, Images: body.Images,
	})
	if err != nil {
		workflowhttperror.RespondWithStatus(ctx, c, "workflow.client.complete_form_revision_task", err)
		return
	}
	response.JSON(c, data)
}

func formRevisionRouteIdentity(c *app.RequestContext, missingIDMessage string) (string, string, bool) {
	actorID, ok := authenticatedActorID(c)
	if !ok {
		c.JSON(consts.StatusForbidden, response.Resp{Code: 1, Msg: "未登录或权限失效"})
		return "", "", false
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		formRevisionBadRequest(c, missingIDMessage)
		return "", "", false
	}
	return actorID, id, true
}

func formRevisionBadRequest(c *app.RequestContext, message string) {
	c.JSON(consts.StatusBadRequest, response.Resp{Code: 1, Msg: message})
}
