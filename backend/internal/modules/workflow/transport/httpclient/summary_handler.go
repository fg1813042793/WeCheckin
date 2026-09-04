package httpclient

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowhttperror "wecheckin/backend/internal/modules/workflow/transport/httperror"
	workflowsummary "wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

type SummaryService interface {
	ListDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error)
	GetDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error)
	ListInstances(context.Context, *model.DingTalkH5PerfUser, workflowsummary.InstanceQuery) (*workflowapp.InstanceList, error)
	GetInstance(context.Context, *model.DingTalkH5PerfUser, string) (*workflowapp.InstanceDetail, error)
	Export(context.Context, *model.DingTalkH5PerfUser, workflowsummary.ExportRequest) (*workflowsummary.ExportResult, error)
}

type SummaryHandler struct {
	service SummaryService
}

func NewSummaryHandler(service SummaryService) *SummaryHandler {
	return &SummaryHandler{service: service}
}

func (handler *SummaryHandler) ListDefinitions(ctx context.Context, c *app.RequestContext) {
	if !handler.authenticated(c) {
		return
	}
	data, err := handler.service.ListDefinitions(ctx)
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.summary.list_definitions", err)
		return
	}
	response.JSON(c, data)
}

func (handler *SummaryHandler) GetDefinition(ctx context.Context, c *app.RequestContext) {
	if !handler.authenticated(c) {
		return
	}
	definitionID, _ := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	data, err := handler.service.GetDefinition(ctx, uint(definitionID))
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.summary.get_definition", err)
		return
	}
	response.JSON(c, data)
}

func (handler *SummaryHandler) ListInstances(ctx context.Context, c *app.RequestContext) {
	user, ok := handler.currentUser(c)
	if !ok {
		return
	}
	definitionID, _ := strconv.ParseUint(c.Query("definitionId"), 10, 64)
	definitionVersion, _ := strconv.Atoi(strings.TrimSpace(c.Query("definitionVersion")))
	data, err := handler.service.ListInstances(ctx, user, workflowsummary.InstanceQuery{
		DefinitionID:      uint(definitionID),
		DefinitionVersion: definitionVersion,
		DefinitionName:    strings.TrimSpace(c.Query("definitionName")),
		StarterName:       strings.TrimSpace(c.Query("starterName")),
		Status:            strings.TrimSpace(c.Query("status")),
		StartTimeFrom:     queryInt64(c, "startTimeFrom"),
		StartTimeTo:       queryInt64(c, "startTimeTo"),
		EndTimeFrom:       queryInt64(c, "endTimeFrom"),
		EndTimeTo:         queryInt64(c, "endTimeTo"),
		Page:              queryInt(c, "page"),
		PageSize:          queryInt(c, "pageSize"),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.summary.list_instances", err)
		return
	}
	response.JSON(c, data)
}

func (handler *SummaryHandler) GetInstance(ctx context.Context, c *app.RequestContext) {
	user, ok := handler.currentUser(c)
	if !ok {
		return
	}
	data, err := handler.service.GetInstance(ctx, user, strings.TrimSpace(c.Param("id")))
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.summary.get_instance", err)
		return
	}
	response.JSON(c, data)
}

func (handler *SummaryHandler) Export(ctx context.Context, c *app.RequestContext) {
	user, ok := handler.currentUser(c)
	if !ok {
		return
	}
	definitionID, _ := strconv.ParseUint(c.Query("definitionId"), 10, 64)
	data, err := handler.service.Export(ctx, user, workflowsummary.ExportRequest{
		DefinitionID: uint(definitionID),
		InstanceIDs:  splitSummaryInstanceIDs(c.Query("instanceIds")),
		Format:       workflowsummary.ExportFormat(strings.ToLower(strings.TrimSpace(c.Query("format")))),
	})
	if err != nil {
		workflowhttperror.Respond(ctx, c, "workflow.summary.export", err)
		return
	}
	c.Header("Content-Type", data.ContentType)
	c.Header("Content-Disposition", summaryContentDisposition(data.Filename))
	c.Header("Cache-Control", "no-store")
	c.Write(data.Body)
}

func (handler *SummaryHandler) authenticated(c *app.RequestContext) bool {
	_, ok := handler.currentUser(c)
	return ok
}

func (handler *SummaryHandler) currentUser(c *app.RequestContext) (*model.DingTalkH5PerfUser, bool) {
	if handler == nil || handler.service == nil {
		response.Fail(c, "流程汇总服务未初始化")
		return nil, false
	}
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok || user.ID == 0 {
		response.Fail(c, "未登录或权限失效")
		return nil, false
	}
	return user, true
}

func splitSummaryInstanceIDs(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' })
}

func summaryContentDisposition(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "workflow-export"
	}
	return "attachment; filename*=UTF-8''" + strings.ReplaceAll(url.QueryEscape(filename), "+", "%20")
}
