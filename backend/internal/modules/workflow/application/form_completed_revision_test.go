package application

import (
	"context"
	"errors"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestCreateCompletedFormRevisionDoesNotGrantStarterIdentity(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	_, err := service.CreateCompletedFormRevision(context.Background(), CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: state.Instance.StarterID,
		ExpectedRevision: 3, FormData: map[string]interface{}{"remark": "越权"}, Reason: "尝试修改",
	})
	if !errors.Is(err, ErrCompletedRevisionNotAllowed) {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateCompletedFormRevisionDoesNotMergeHandledNodePermissions(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	state.History = append(state.History,
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "manager", ActorID: "21"},
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "hr", ActorID: "21"},
	)
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	_, err := service.CreateCompletedFormRevision(context.Background(), CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: "21",
		ExpectedRevision: 3, FormData: map[string]interface{}{"hrComment": "不能跨节点"}, Reason: "权限测试",
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateDirectCompletedFormRevisionAppliesAtomically(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	state.History = append(state.History, workflowdomain.HistoryEvent{
		Type: workflowdomain.HistoryTaskApproved, NodeID: "manager", ActorID: "21",
	})
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	result, err := service.CreateCompletedFormRevision(context.Background(), CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: "21",
		ExpectedRevision: 3, FormData: map[string]interface{}{"remark": "已更正"}, Reason: "修正说明",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "applied" || result.AppliedFormRevision != 4 {
		t.Fatalf("result=%#v", result)
	}
	if store.savedState == nil || store.savedState.FormData["remark"] != "已更正" || store.createdFormRevision == nil {
		t.Fatalf("state=%#v revision=%#v", store.savedState, store.createdFormRevision)
	}
	if store.savedState.Instance.Status != workflowdomain.InstanceStatusCompleted || store.savedState.Instance.FormRevision != 4 {
		t.Fatalf("instance=%#v", store.savedState.Instance)
	}
}

func TestPreviewCompletedFormRevisionDoesNotWrite(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	state.History = append(state.History, workflowdomain.HistoryEvent{
		Type: workflowdomain.HistoryTaskApproved, NodeID: "manager", ActorID: "21",
	})
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	preview, err := service.PreviewCompletedFormRevision(context.Background(), PreviewCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: "21",
		ExpectedRevision: 3, FormData: map[string]interface{}{"remark": "已更正"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Mode != workflowcore.CompletedRevisionModeDirect || len(preview.ChangedFields) != 1 || preview.ChangedFields[0] != "说明" {
		t.Fatalf("preview=%#v", preview)
	}
	if store.createdFormRevision != nil || store.savedState != nil {
		t.Fatalf("preview wrote state: revision=%#v state=%#v", store.createdFormRevision, store.savedState)
	}
}

func TestCompletedInstanceFormRevisionCapabilityUsesHandledNodesByLatestTime(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	detail := &InstanceDetail{
		Instance: InstanceSummary{ID: state.Instance.ID, Status: string(state.Instance.Status), FormRevision: state.Instance.FormRevision},
		Form:     definition.Form, FieldPermissions: map[string][]workflowcore.FieldPermission{},
	}
	for _, node := range definition.Nodes {
		detail.Nodes = append(detail.Nodes, PublishedNode{
			ID: node.ID, Type: node.Type, Name: node.Name, PostHandleEdit: node.PostHandleEdit,
		})
		detail.FieldPermissions[node.ID] = node.FormPermissions
	}
	for _, edge := range definition.Edges {
		detail.Edges = append(detail.Edges, PublishedEdge{Source: edge.Source, Target: edge.Target, Condition: edge.Condition})
	}
	detail.History = []HistorySummary{
		{EventType: string(workflowdomain.HistoryTaskApproved), NodeID: "manager", ActorID: "21", EventTime: 1000},
		{EventType: string(workflowdomain.HistoryTaskApproved), NodeID: "hr", ActorID: "21", EventTime: 2000},
	}

	decorateInstanceFormRevision(detail, "21")
	if len(detail.FormRevision.CompletedRevisionNodes) != 2 || detail.FormRevision.CompletedRevisionNodes[0].NodeID != "hr" {
		t.Fatalf("completed capability=%#v", detail.FormRevision)
	}
	if detail.FormRevision.Allowed || detail.FormRevision.RunningRevisionAllowed {
		t.Fatalf("running revision must stay disabled: %#v", detail.FormRevision)
	}
}

func completedRevisionApplicationFixture() (workflowcore.Definition, *workflowdomain.State) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion, Key: "completed_revision", Name: "完成后修订",
		Form: []workflowcore.FormField{
			{Key: "remark", Label: "说明", Type: workflowcore.FormFieldTypeTextarea},
			{Key: "hrComment", Label: "HR 意见", Type: workflowcore.FormFieldTypeTextarea},
		},
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{
				ID: "manager", Type: workflowcore.NodeTypeApproval, Name: "主管审批",
				ApprovalMode: workflowcore.ApprovalModeSingle,
				PostHandleEdit: &workflowcore.PostHandleEditConfig{CompletedRevision: &workflowcore.CompletedRevisionConfig{
					Enabled: true, DirectFields: []string{"remark"},
				}},
				FormPermissions: []workflowcore.FieldPermission{{Field: "remark", Access: workflowcore.FieldAccessWrite}},
			},
			{
				ID: "hr", Type: workflowcore.NodeTypeApproval, Name: "HR 审批",
				ApprovalMode:    workflowcore.ApprovalModeSingle,
				PostHandleEdit:  &workflowcore.PostHandleEditConfig{CompletedRevision: &workflowcore.CompletedRevisionConfig{Enabled: true}},
				FormPermissions: []workflowcore.FieldPermission{{Field: "hrComment", Access: workflowcore.FieldAccessWrite}},
			},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "manager"},
			{ID: "e2", Source: "manager", Target: "hr"},
			{ID: "e3", Source: "hr", Target: "end"},
		},
	}
	return definition, &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{
			ID: "instance-1", StarterID: "7", Status: workflowdomain.InstanceStatusCompleted, FormRevision: 3,
		},
		FormData: map[string]interface{}{"remark": "原说明", "hrComment": "原意见"},
	}
}
