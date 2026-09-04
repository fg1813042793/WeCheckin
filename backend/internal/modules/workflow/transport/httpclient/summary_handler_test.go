package httpclient

import (
	"context"
	"testing"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowsummary "wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
	"wecheckin/backend/internal/support/dingtalkh5session"
)

type summaryServiceQueryStub struct {
	query workflowsummary.InstanceQuery
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

func (stub *summaryServiceQueryStub) Export(context.Context, *model.DingTalkH5PerfUser, workflowsummary.ExportRequest) (*workflowsummary.ExportResult, error) {
	return nil, nil
}

func TestSummaryListInstancesParsesDefinitionNameWithoutDefinitionID(t *testing.T) {
	stub := &summaryServiceQueryStub{}
	handler := NewSummaryHandler(stub)
	requestContext := newUserContext(42)
	dingtalkh5session.SetAuth(requestContext, &model.DingTalkH5PerfUser{ID: 42}, "token")
	requestContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/summary/instances?definitionName=%E7%BB%A9%E6%95%88&page=2&pageSize=50")

	handler.ListInstances(context.Background(), requestContext)

	if stub.query.DefinitionID != 0 || stub.query.DefinitionName != "绩效" || stub.query.Page != 2 || stub.query.PageSize != 50 {
		t.Fatalf("summary query = %+v", stub.query)
	}
}
