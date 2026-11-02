package httpadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/modules/inappnotification/application"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/response"
)

type TemplateService interface {
	ListDingTalkTemplates(context.Context, application.DingTalkTemplateQuery) (application.DingTalkTemplateList, error)
	DingTalkTemplateOptions(context.Context, application.DingTalkTemplateQuery) (application.DingTalkTemplateList, error)
	DingTalkTemplateWorkflowOptions(context.Context, application.DingTalkTemplateQuery) (application.DingTalkTemplateWorkflowList, error)
	CreateDingTalkTemplate(context.Context, uint, application.DingTalkTemplateInput) (application.DingTalkSendTemplate, error)
	UpdateDingTalkTemplate(context.Context, uint, uint, application.DingTalkTemplateInput) (application.DingTalkSendTemplate, error)
	DeleteDingTalkTemplate(context.Context, uint, uint, int) error
}

type DingTalkTemplateDeleteResult struct {
	ID uint `json:"id"`
}

func (handler *Handler) ListDingTalkTemplates(ctx context.Context, c *app.RequestContext) {
	handler.listTemplates(ctx, c, false)
}

func (handler *Handler) DingTalkTemplateOptions(ctx context.Context, c *app.RequestContext) {
	handler.listTemplates(ctx, c, true)
}

func (handler *Handler) listTemplates(ctx context.Context, c *app.RequestContext, options bool) {
	service, _, ok := handler.templateService(ctx, c)
	if !ok {
		return
	}
	query, err := templateQuery(c)
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	var result application.DingTalkTemplateList
	if options {
		result, err = service.DingTalkTemplateOptions(ctx, query)
	} else {
		result, err = service.ListDingTalkTemplates(ctx, query)
	}
	respond(ctx, c, result, err)
}

func (handler *Handler) DingTalkTemplateWorkflowOptions(ctx context.Context, c *app.RequestContext) {
	service, _, ok := handler.templateService(ctx, c)
	if !ok {
		return
	}
	query, err := templateQuery(c)
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	result, err := service.DingTalkTemplateWorkflowOptions(ctx, query)
	respond(ctx, c, result, err)
}

func (handler *Handler) CreateDingTalkTemplate(ctx context.Context, c *app.RequestContext) {
	service, actor, ok := handler.templateService(ctx, c)
	if !ok {
		return
	}
	var input application.DingTalkTemplateInput
	if err := decodeTemplateInput(c, &input); err != nil {
		respond(ctx, c, nil, err)
		return
	}
	result, err := service.CreateDingTalkTemplate(ctx, actor, input)
	respond(ctx, c, result, err)
}

func (handler *Handler) UpdateDingTalkTemplate(ctx context.Context, c *app.RequestContext) {
	service, actor, ok := handler.templateService(ctx, c)
	if !ok {
		return
	}
	id, err := templateID(c.Param("id"))
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	var input application.DingTalkTemplateInput
	if err := decodeTemplateInput(c, &input); err != nil {
		respond(ctx, c, nil, err)
		return
	}
	if input.Version <= 0 || input.Version >= application.MaxTemplateVersion {
		respond(ctx, c, nil, invalidTemplateRequest("请提供有效的模板版本号并刷新后重试"))
		return
	}
	result, err := service.UpdateDingTalkTemplate(ctx, actor, id, input)
	respond(ctx, c, result, err)
}

func (handler *Handler) DeleteDingTalkTemplate(ctx context.Context, c *app.RequestContext) {
	service, actor, ok := handler.templateService(ctx, c)
	if !ok {
		return
	}
	id, err := templateID(c.Param("id"))
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	raw, err := uniqueTemplateQuery(c, "version")
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version <= 0 || version > application.MaxTemplateVersion {
		respond(ctx, c, nil, invalidTemplateRequest("删除模板必须提供有效的version版本号，请刷新后重试"))
		return
	}
	err = service.DeleteDingTalkTemplate(ctx, actor, id, version)
	respond(ctx, c, DingTalkTemplateDeleteResult{ID: id}, err)
}

func (handler *Handler) templateService(ctx context.Context, c *app.RequestContext) (TemplateService, uint, bool) {
	admin, ok := authenticatedAdmin(c)
	if !ok {
		response.Fail(c, "管理员未登录")
		return nil, 0, false
	}
	if handler != nil {
		if service, ok := handler.service.(TemplateService); ok {
			return service, admin.ID, true
		}
	}
	respond(ctx, c, nil, application.ErrTemplateStoreUnavailable)
	return nil, 0, false
}

func templateID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || id > application.MaxTemplateID {
		return 0, invalidTemplateRequest("模板或流程ID无效")
	}
	return uint(id), nil
}

func uniqueTemplateQuery(c *app.RequestContext, key string) (string, error) {
	values := c.QueryArgs().PeekAll(key)
	if len(values) > 1 {
		return "", invalidTemplateRequest("筛选参数不能重复")
	}
	if len(values) == 0 {
		return "", nil
	}
	if len(values[0]) == 0 {
		return "", invalidTemplateRequest("筛选参数不能为空")
	}
	return string(values[0]), nil
}

func templateQuery(c *app.RequestContext) (application.DingTalkTemplateQuery, error) {
	query := application.DingTalkTemplateQuery{Page: 1, PageSize: 20}
	values := make(map[string]string, 7)
	for _, key := range []string{"keyword", "targetKind", "page", "pageSize", "id", "enabled"} {
		// An empty search term is equivalent to no search.
		if key == "keyword" && len(c.QueryArgs().PeekAll(key)) <= 1 {
			values[key] = c.Query(key)
			continue
		}
		value, err := uniqueTemplateQuery(c, key)
		if err != nil {
			return query, err
		}
		values[key] = value
	}
	query.Keyword = values["keyword"]
	query.TargetKind = values["targetKind"]
	for _, field := range []struct {
		key    string
		target *int
	}{{"page", &query.Page}, {"pageSize", &query.PageSize}} {
		if raw := values[field.key]; raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n <= 0 {
				return query, invalidTemplateRequest("分页参数必须为正整数")
			}
			*field.target = n
		}
	}
	if raw := values["id"]; raw != "" {
		id, err := templateID(raw)
		if err != nil {
			return query, err
		}
		query.ID = id
	}
	if raw := values["enabled"]; raw != "" {
		if raw != "true" && raw != "false" {
			return query, invalidTemplateRequest("enabled必须为true或false")
		}
		enabled := raw == "true"
		query.Enabled = &enabled
	}
	return application.NormalizeDingTalkTemplateQuery(query)
}

func decodeTemplateInput(c *app.RequestContext, input *application.DingTalkTemplateInput) error {
	body := c.Request.Body()
	if len(body) == 0 || len(body) > 64*1024 || bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
		return invalidTemplateRequest("模板请求体为空或超过64KB限制")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		return invalidTemplateRequest("模板参数格式无效或包含未支持字段")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return invalidTemplateRequest("模板请求只能包含一个JSON对象")
	}
	return nil
}

func invalidTemplateRequest(message string) error {
	return &application.TemplateValidationError{Message: message}
}

func templateErrorReason(err error) (string, string) {
	var validation *application.TemplateValidationError
	if errors.As(err, &validation) {
		return "template_invalid", validation.Error()
	}
	for _, entry := range []struct {
		err  error
		code string
	}{
		{application.ErrTemplateInvalid, "template_invalid"}, {application.ErrTemplateNotFound, "template_not_found"},
		{application.ErrTemplateDisabled, "template_disabled"}, {application.ErrTemplateVersionConflict, "template_version_conflict"},
		{application.ErrTemplateNameExists, "template_name_exists"}, {application.ErrTemplateWorkflowUnavailable, "template_workflow_unavailable"},
		{application.ErrTemplateStoreUnavailable, "template_store_unavailable"}, {application.ErrTemplateSchemaMissing, "template_schema_missing"},
		{application.ErrTemplateTargetUnavailable, "template_target_unavailable"},
	} {
		if errors.Is(err, entry.err) {
			return entry.code, entry.err.Error()
		}
	}
	return "", ""
}

func recordTemplateError(ctx context.Context, err error) {
	if code, message := templateErrorReason(err); code != "" {
		auditlog.Update(ctx, func(d *auditlog.Details) { d.Outcome = "failure"; d.ReasonCode = code; d.Reason = message })
	}
}

var _ TemplateService = (*application.Service)(nil)
