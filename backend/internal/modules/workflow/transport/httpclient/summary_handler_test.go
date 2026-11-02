package httpclient

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowsummary "wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
	"wecheckin/backend/internal/support/dingtalkh5session"
)

type summaryServiceQueryStub struct {
	query         workflowsummary.InstanceQuery
	exportRequest workflowsummary.ExportRequest
	exportErr     error
	exportCalls   int
}

func (stub *summaryServiceQueryStub) ListDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error) {
	return nil, nil
}

func (stub *summaryServiceQueryStub) GetDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error) {
	return nil, nil
}

func (stub *summaryServiceQueryStub) ListInstances(_ context.Context, _ *model.DingTalkH5PerfUser, query workflowsummary.InstanceQuery) (*workflowapp.InstanceList, error) {
	stub.query = query
	return &workflowapp.InstanceList{}, nil
}

func (stub *summaryServiceQueryStub) GetInstance(context.Context, *model.DingTalkH5PerfUser, string) (*workflowapp.InstanceDetail, error) {
	return nil, nil
}

func (stub *summaryServiceQueryStub) Export(_ context.Context, _ *model.DingTalkH5PerfUser, request workflowsummary.ExportRequest) (*workflowsummary.ExportResult, error) {
	stub.exportRequest = request
	stub.exportCalls++
	if stub.exportErr != nil {
		return nil, stub.exportErr
	}
	return &workflowsummary.ExportResult{Filename: "summary.xlsx", ContentType: "application/test", Body: []byte("exported")}, nil
}

func TestSummaryExportPropagatesContent(t *testing.T) {
	for _, content := range []string{"", "full", "form"} {
		t.Run(content, func(t *testing.T) {
			stub := &summaryServiceQueryStub{}
			requestContext := newUserContext(42)
			dingtalkh5session.SetAuth(requestContext, &model.DingTalkH5PerfUser{ID: 42}, "token")
			requestContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/summary/export?instanceIds=one,two&format=docx&content=" + content)
			NewSummaryHandler(stub).Export(context.Background(), requestContext)
			if stub.exportRequest.Content != workflowsummary.ExportContent(content) || stub.exportRequest.Format != workflowsummary.ExportFormatDOCX || !reflect.DeepEqual(stub.exportRequest.InstanceIDs, []string{"one", "two"}) {
				t.Fatalf("export request = %+v", stub.exportRequest)
			}
			if string(requestContext.Response.Body()) != "exported" || string(requestContext.Response.Header.Peek("Cache-Control")) != "no-store" {
				t.Fatalf("export response = %s", requestContext.Response.Body())
			}
		})
	}
}

func TestSummaryExportInvalidContentReturnsParameterError(t *testing.T) {
	// Validation must reject invalid content without requiring the runtime or a database.
	service := workflowsummary.NewServiceWithAccess(nil, nil)
	requestContext := newUserContext(42)
	dingtalkh5session.SetAuth(requestContext, &model.DingTalkH5PerfUser{ID: 42}, "token")
	requestContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/summary/export?instanceIds=one&format=pdf&content=unknown")
	NewSummaryHandler(service).Export(context.Background(), requestContext)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(requestContext.Response.Body(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code == 0 || result.Msg != "导出内容仅支持 full（完整记录）或 form（仅表单）" {
		t.Fatalf("invalid content response = %+v", result)
	}
	if len(requestContext.Response.Header.Peek("Content-Disposition")) != 0 {
		t.Fatal("invalid content must not return a download")
	}
}

func TestSummaryExportContentAccessDenied(t *testing.T) {
	for _, content := range []string{"", "full", "form"} {
		for _, loggedIn := range []bool{false, true} {
			stub := &summaryServiceQueryStub{exportErr: workflowsummary.ErrSummaryAccessDenied}
			requestContext := newUserContext(42)
			if loggedIn {
				dingtalkh5session.SetAuth(requestContext, &model.DingTalkH5PerfUser{ID: 42}, "token")
			}
			requestContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/summary/export?instanceIds=one&format=xlsx&content=" + content)
			NewSummaryHandler(stub).Export(context.Background(), requestContext)
			wantMessage := "未登录或权限失效"
			if loggedIn {
				wantMessage = "无权执行该流程操作"
			} else if stub.exportCalls != 0 {
				t.Fatal("unauthenticated export reached service")
			}
			if !strings.Contains(string(requestContext.Response.Body()), wantMessage) || len(requestContext.Response.Header.Peek("Content-Disposition")) != 0 {
				t.Fatalf("denied response = %s", requestContext.Response.Body())
			}
		}
	}
}

func TestSummaryExportIgnoresDefinitionIDAndParsesSelectedInstances(t *testing.T) {
	stub := &summaryServiceQueryStub{}
	handler := NewSummaryHandler(stub)
	requestContext := newUserContext(42)
	dingtalkh5session.SetAuth(requestContext, &model.DingTalkH5PerfUser{ID: 42}, "token")
	requestContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/summary/export?definitionId=999&instanceIds=instance-v2,instance-v7&format=xlsx")

	handler.Export(context.Background(), requestContext)

	want := workflowsummary.ExportRequest{
		InstanceIDs: []string{"instance-v2", "instance-v7"},
		Format:      workflowsummary.ExportFormatXLSX,
	}
	if !reflect.DeepEqual(stub.exportRequest, want) {
		t.Fatalf("export request = %#v, want %#v", stub.exportRequest, want)
	}
	if string(requestContext.Response.Body()) != "exported" {
		t.Fatalf("export body = %q", requestContext.Response.Body())
	}
}

func TestSummaryListInstancesParsesDefinitionNameWithoutDefinitionID(t *testing.T) {
	stub := &summaryServiceQueryStub{}
	handler := NewSummaryHandler(stub)
	requestContext := newUserContext(42)
	dingtalkh5session.SetAuth(requestContext, &model.DingTalkH5PerfUser{ID: 42}, "token")
	requestContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/summary/instances?definitionName=%E7%BB%A9%E6%95%88&instanceTitle=2026%E5%B9%B48%E6%9C%88%E7%BB%A9%E6%95%88&businessPeriodKey=2026-08&page=2&pageSize=50")

	handler.ListInstances(context.Background(), requestContext)

	if stub.query.DefinitionID != 0 || stub.query.DefinitionName != "绩效" ||
		stub.query.InstanceTitle != "2026年8月绩效" || stub.query.BusinessPeriodKey != "2026-08" ||
		stub.query.Page != 2 || stub.query.PageSize != 50 {
		t.Fatalf("summary query = %+v", stub.query)
	}
}
