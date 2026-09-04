package workflowsummary

import (
	"context"
	"errors"
	"testing"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
)

type summaryRuntimeStub struct {
	query       workflowapp.InstanceQuery
	list        *workflowapp.InstanceList
	detail      *workflowapp.InstanceDetail
	detailCalls int
}

func (stub *summaryRuntimeStub) ListPublishedDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error) {
	return nil, nil
}

func (stub *summaryRuntimeStub) GetPublishedDefinition(context.Context, uint) (*workflowapp.PublishedDefinition, error) {
	return nil, nil
}

func (stub *summaryRuntimeStub) ListInstances(_ context.Context, query workflowapp.InstanceQuery) (*workflowapp.InstanceList, error) {
	stub.query = query
	return stub.list, nil
}

func (stub *summaryRuntimeStub) GetInstance(context.Context, string) (*workflowapp.InstanceDetail, error) {
	stub.detailCalls++
	return stub.detail, nil
}

type summaryAccessStub struct {
	visibility workflowapp.InstanceVisibility
	err        error
}

func (stub summaryAccessStub) Resolve(context.Context, *model.DingTalkH5PerfUser) (workflowapp.InstanceVisibility, error) {
	return stub.visibility, stub.err
}

func TestListInstancesSupportsCrossDefinitionNameFilterAndAppliesVisibility(t *testing.T) {
	runtime := &summaryRuntimeStub{list: &workflowapp.InstanceList{}}
	visibility := workflowapp.InstanceVisibility{Ready: true, UserIDs: []uint{7}}
	service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: visibility})

	_, err := service.ListInstances(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, InstanceQuery{
		DefinitionName: "绩效",
		Page:           3,
		PageSize:       500,
	})
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if runtime.query.DefinitionID != 0 || runtime.query.DefinitionName != "绩效" || runtime.query.Page != 3 || runtime.query.PageSize != 20 {
		t.Fatalf("query = %#v", runtime.query)
	}
	if runtime.query.Visibility == nil || len(runtime.query.Visibility.UserIDs) != 1 || runtime.query.Visibility.UserIDs[0] != 7 {
		t.Fatalf("visibility = %#v", runtime.query.Visibility)
	}
}

func TestGetInstanceRejectsInvisibleIDBeforeLoadingDetail(t *testing.T) {
	runtime := &summaryRuntimeStub{list: &workflowapp.InstanceList{}}
	service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: workflowapp.InstanceVisibility{Ready: true, UserIDs: []uint{7}}})

	_, err := service.GetInstance(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, "instance-forged")
	if !errors.Is(err, ErrSummaryAccessDenied) {
		t.Fatalf("error = %v", err)
	}
	if runtime.detailCalls != 0 {
		t.Fatalf("detail calls = %d, want 0", runtime.detailCalls)
	}
}

func TestExportRejectsPartialAuthorizedSelection(t *testing.T) {
	runtime := &summaryRuntimeStub{list: &workflowapp.InstanceList{
		List: []workflowapp.InstanceSummary{{ID: "allowed", DefinitionID: 12}},
	}}
	service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: workflowapp.InstanceVisibility{Ready: true, All: true}})

	_, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{
		DefinitionID: 12,
		InstanceIDs:  []string{"allowed", "forged"},
		Format:       ExportFormatPDF,
	})
	if !errors.Is(err, ErrSummaryAccessDenied) {
		t.Fatalf("error = %v", err)
	}
	if runtime.detailCalls != 0 {
		t.Fatalf("detail calls = %d, want 0", runtime.detailCalls)
	}
}

func TestAccessResolverErrorStopsSummaryQuery(t *testing.T) {
	runtime := &summaryRuntimeStub{}
	want := errors.New("scope unavailable")
	service := NewServiceWithAccess(runtime, summaryAccessStub{err: want})
	_, err := service.ListInstances(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, InstanceQuery{DefinitionID: 12})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
