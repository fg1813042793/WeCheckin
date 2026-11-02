package httpadmin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"
	"wecheckin/backend/internal/modules/inappnotification/application"
	"wecheckin/backend/internal/support/auditlog"
)

func TestTemplateHandlersUseBoundedQueriesAndTypedResponses(t *testing.T) {
	service := &templateServiceStub{}
	handler := NewHandler(service)
	for _, endpoint := range []struct {
		method func(context.Context, *app.RequestContext)
		kind   string
	}{{handler.ListDingTalkTemplates, "list"}, {handler.DingTalkTemplateOptions, "options"}, {handler.DingTalkTemplateWorkflowOptions, "workflows"}} {
		c := newAdminContext(9)
		c.Request.SetRequestURI("/templates?keyword=abc&targetKind=workflow_start&page=2&pageSize=50&id=17")
		endpoint.method(context.Background(), c)
		if service.kind != endpoint.kind || service.query.Page != 2 || service.query.PageSize != 50 || service.query.ID != 17 || service.query.Keyword != "abc" || service.query.TargetKind != "workflow_start" {
			t.Fatalf("query=%#v kind=%s", service.query, service.kind)
		}
		var result struct {
			Code int                          `json:"code"`
			Data struct{ Page, PageSize int } `json:"data"`
		}
		if err := json.Unmarshal(c.Response.Body(), &result); err != nil || result.Code != 0 || result.Data.PageSize != 50 {
			t.Fatalf("response=%s err=%v", c.Response.Body(), err)
		}
	}
	for _, query := range []string{"page=-1", "page=x", "pageSize=101", "pageSize=0", "id=-1", "id=0", "enabled=maybe", "targetKind=unknown", "page=1&page=2"} {
		c := newAdminContext(9)
		c.Request.SetRequestURI("/templates?" + query)
		before := service.calls
		handler.ListDingTalkTemplates(context.Background(), c)
		if service.calls != before || !strings.Contains(string(c.Response.Body()), `"code":1`) {
			t.Fatalf("query=%s calls=%d response=%s", query, service.calls, c.Response.Body())
		}
	}
}

func TestTemplateHandlersCreateUpdateDeleteAndRequireAuthVersion(t *testing.T) {
	service := &templateServiceStub{}
	handler := NewHandler(service)
	body := `{"name":"模板","enabled":true,"title":"标题","content":"内容","messageType":"oa","target":{"kind":"workflow_start","definitionId":17},"version":3}`
	c := newAdminContext(9)
	c.Request.SetBodyString(body)
	handler.CreateDingTalkTemplate(context.Background(), c)
	if service.actor != 9 || service.input.Name != "模板" || !strings.Contains(string(c.Response.Body()), `"definitionName":"可信流程"`) {
		t.Fatalf("response=%s", c.Response.Body())
	}
	c = newAdminContext(9)
	c.Params = param.Params{{Key: "id", Value: "7"}}
	c.Request.SetBodyString(body)
	handler.UpdateDingTalkTemplate(context.Background(), c)
	if service.id != 7 || service.input.Version != 3 || !strings.Contains(string(c.Response.Body()), `"version":4`) {
		t.Fatalf("update=%s", c.Response.Body())
	}
	for _, query := range []string{"", "?version=0", "?version=-1", "?version=no"} {
		c = newAdminContext(9)
		c.Params = param.Params{{Key: "id", Value: "7"}}
		c.Request.SetRequestURI("/templates/7" + query)
		before := service.calls
		handler.DeleteDingTalkTemplate(context.Background(), c)
		if service.calls != before || !strings.Contains(string(c.Response.Body()), `"code":1`) {
			t.Fatalf("delete=%s", c.Response.Body())
		}
	}
	c = newAdminContext(9)
	c.Params = param.Params{{Key: "id", Value: "7"}}
	c.Request.SetRequestURI("/templates/7?version=3")
	handler.DeleteDingTalkTemplate(context.Background(), c)
	if service.id != 7 || service.version != 3 || !strings.Contains(string(c.Response.Body()), `"id":7`) {
		t.Fatalf("delete=%s", c.Response.Body())
	}
	for _, method := range []func(context.Context, *app.RequestContext){handler.ListDingTalkTemplates, handler.DingTalkTemplateOptions, handler.DingTalkTemplateWorkflowOptions, handler.CreateDingTalkTemplate, handler.UpdateDingTalkTemplate, handler.DeleteDingTalkTemplate} {
		before := service.calls
		c = app.NewContext(1)
		method(context.Background(), c)
		if service.calls != before || !strings.Contains(string(c.Response.Body()), "未登录") {
			t.Fatalf("auth=%s", c.Response.Body())
		}
	}
}

func TestTemplateHandlerRejectsOversizedOrUnknownFieldsAndClassifiesErrors(t *testing.T) {
	service := &templateServiceStub{}
	handler := NewHandler(service)
	for _, body := range []string{`{"id":7}`, `{"name":"` + strings.Repeat("x", 65536) + `"}`, `null`} {
		c := newAdminContext(9)
		c.Request.SetBodyString(body)
		handler.CreateDingTalkTemplate(context.Background(), c)
		if service.calls != 0 || !strings.Contains(string(c.Response.Body()), `"code":1`) {
			t.Fatalf("body accepted=%s", c.Response.Body())
		}
	}
	ctx, recorder := auditlog.WithRecorder(context.Background())
	c := newAdminContext(9)
	respond(ctx, c, nil, application.ErrTemplateVersionConflict)
	if !strings.Contains(string(c.Response.Body()), "刷新") || recorder.Snapshot().ReasonCode != "template_version_conflict" || recorder.Snapshot().Outcome != "failure" {
		t.Fatalf("error=%s audit=%#v", c.Response.Body(), recorder.Snapshot())
	}
}

func TestDingTalkSendPassesOnlyTemplateReference(t *testing.T) {
	service := &serviceStub{}
	c := newAdminContext(9)
	c.Request.SetBodyString(`{"title":"临时标题","content":"临时正文","scope":"all","templateId":7,"templateVersion":3,"target":{"kind":"url","url":"https://forged.example"},"messageType":"text"}`)
	NewHandler(service).SendDingTalk(context.Background(), c)
	if service.dingTalkSendInput.TemplateID != 7 || service.dingTalkSendInput.TemplateVersion != 3 || service.dingTalkSendInput.NotificationType != "" {
		t.Fatalf("input=%#v", service.dingTalkSendInput)
	}
}

type templateServiceStub struct {
	serviceStub
	query          application.DingTalkTemplateQuery
	input          application.DingTalkTemplateInput
	actor, id      uint
	version, calls int
	kind           string
}

func (s *templateServiceStub) ListDingTalkTemplates(_ context.Context, q application.DingTalkTemplateQuery) (application.DingTalkTemplateList, error) {
	s.query = q
	s.calls++
	s.kind = "list"
	return application.DingTalkTemplateList{List: []application.DingTalkSendTemplate{}, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *templateServiceStub) DingTalkTemplateOptions(ctx context.Context, q application.DingTalkTemplateQuery) (application.DingTalkTemplateList, error) {
	r, e := s.ListDingTalkTemplates(ctx, q)
	s.kind = "options"
	return r, e
}

func (s *templateServiceStub) DingTalkTemplateWorkflowOptions(_ context.Context, q application.DingTalkTemplateQuery) (application.DingTalkTemplateWorkflowList, error) {
	s.query = q
	s.calls++
	s.kind = "workflows"
	return application.DingTalkTemplateWorkflowList{List: []application.DingTalkTemplateWorkflowOption{}, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *templateServiceStub) CreateDingTalkTemplate(_ context.Context, actor uint, in application.DingTalkTemplateInput) (application.DingTalkSendTemplate, error) {
	s.actor = actor
	s.input = in
	s.calls++
	return application.DingTalkSendTemplate{ID: 7, Version: 4, Target: application.DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17, DefinitionName: "可信流程"}}, nil
}

func (s *templateServiceStub) UpdateDingTalkTemplate(ctx context.Context, actor, id uint, in application.DingTalkTemplateInput) (application.DingTalkSendTemplate, error) {
	s.id = id
	return s.CreateDingTalkTemplate(ctx, actor, in)
}

func (s *templateServiceStub) DeleteDingTalkTemplate(_ context.Context, actor, id uint, version int) error {
	s.actor = actor
	s.id = id
	s.version = version
	s.calls++
	return nil
}
