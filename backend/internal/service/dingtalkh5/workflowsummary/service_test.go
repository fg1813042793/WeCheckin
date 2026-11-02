package workflowsummary

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
)

type summaryRuntimeStub struct {
	query       workflowapp.InstanceQuery
	list        *workflowapp.InstanceList
	detail      *workflowapp.InstanceDetail
	details     map[string]*workflowapp.InstanceDetail
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

func (stub *summaryRuntimeStub) GetInstance(_ context.Context, instanceID string) (*workflowapp.InstanceDetail, error) {
	stub.detailCalls++
	if stub.details != nil {
		return stub.details[instanceID], nil
	}
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
	if !runtime.query.IncludeFormData {
		t.Fatal("workflow summary list must include form data for configurable columns")
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

func TestGetSummaryInstanceAllowsVisibleNonParticipant(t *testing.T) {
	instance := workflowapp.InstanceSummary{ID: "visible", StarterID: "42"}
	runtime := &summaryRuntimeStub{
		list:   &workflowapp.InstanceList{List: []workflowapp.InstanceSummary{instance}},
		detail: &workflowapp.InstanceDetail{Instance: instance, FormData: map[string]interface{}{"score": 88}},
	}
	service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: workflowapp.InstanceVisibility{Ready: true, UserIDs: []uint{42}}})
	detail, err := service.GetInstance(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, "visible")
	if err != nil || detail == nil || detail.FormData["score"] != 88 {
		t.Fatalf("summary scope should allow an authorized nonparticipant: detail=%v err=%v", detail, err)
	}
	if runtime.detailCalls != 1 || runtime.query.Visibility == nil || len(runtime.query.InstanceIDs) != 1 || runtime.query.InstanceIDs[0] != "visible" {
		t.Fatalf("summary detail must check visibility before loading: query=%+v calls=%d", runtime.query, runtime.detailCalls)
	}
}

func TestExportRejectsPartialAuthorizedSelection(t *testing.T) {
	runtime := &summaryRuntimeStub{list: &workflowapp.InstanceList{
		List: []workflowapp.InstanceSummary{{ID: "allowed", DefinitionID: 12}},
	}}
	service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: workflowapp.InstanceVisibility{Ready: true, All: true}})

	_, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{
		InstanceIDs: []string{"allowed", "forged"},
		Format:      ExportFormatPDF,
	})
	if !errors.Is(err, ErrSummaryAccessDenied) {
		t.Fatalf("error = %v", err)
	}
	if runtime.detailCalls != 0 {
		t.Fatalf("detail calls = %d, want 0", runtime.detailCalls)
	}
}

func TestExportSupportsCrossDefinitionVersions(t *testing.T) {
	runtime := &summaryRuntimeStub{
		list: &workflowapp.InstanceList{List: []workflowapp.InstanceSummary{
			{ID: "instance-v2", DefinitionID: 12, DefinitionVersion: 2},
			{ID: "instance-v7", DefinitionID: 34, DefinitionVersion: 7},
		}},
		details: map[string]*workflowapp.InstanceDetail{
			"instance-v2": {
				Instance: workflowapp.InstanceSummary{
					ID: "instance-v2", DefinitionID: 12, DefinitionVersion: 2,
					DefinitionName: "绩效考评单", BusinessKey: "PERF-V2",
				},
			},
			"instance-v7": {
				Instance: workflowapp.InstanceSummary{
					ID: "instance-v7", DefinitionID: 34, DefinitionVersion: 7,
					DefinitionName: "绩效考评单", BusinessKey: "PERF-V7",
				},
			},
		},
	}
	service := NewServiceWithAccess(runtime, summaryAccessStub{visibility: workflowapp.InstanceVisibility{Ready: true, All: true}})

	result, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{
		InstanceIDs: []string{"instance-v2", "instance-v7"},
		Format:      ExportFormatXLSX,
	})
	if err != nil {
		t.Fatalf("export cross-definition versions: %v", err)
	}
	if runtime.query.DefinitionID != 0 {
		t.Fatalf("authorization definition ID = %d, want 0", runtime.query.DefinitionID)
	}
	if runtime.detailCalls != 2 {
		t.Fatalf("detail calls = %d, want 2", runtime.detailCalls)
	}
	entries := readZipEntries(t, result.Body)
	if !strings.Contains(entries["xl/workbook.xml"], `sheetId="2"`) {
		t.Fatal("cross-definition XLSX should contain one sheet per selected instance")
	}
	if !strings.Contains(entries["xl/worksheets/sheet1.xml"], "v2") || !strings.Contains(entries["xl/worksheets/sheet2.xml"], "v7") {
		t.Fatal("each sheet should retain its instance-bound definition version")
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
