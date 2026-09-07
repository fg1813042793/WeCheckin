package application

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
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

func TestCreateReviewedRevisionRebuildsActualDownstreamStages(t *testing.T) {
	service, store := reviewedRevisionServiceFixture()
	result, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "pending" || store.createdFormRevision == nil || len(store.createdFormRevision.Tasks) != 3 {
		t.Fatalf("result=%#v revision=%#v", result, store.createdFormRevision)
	}
	if got := revisionTaskNodeStages(store.createdFormRevision.Tasks); got != "finance:1:pending,legal:1:pending,hr:2:waiting" {
		t.Fatalf("stages=%s", got)
	}
}

func TestCreateReviewedRevisionPersistsFirstStageNotifications(t *testing.T) {
	_, store := reviewedRevisionServiceFixture()
	store.revisionNotificationOutboxIDs = []string{"revision-notification-1-in_app"}
	dispatcher := &recordingNotificationDispatcher{store: store}
	service := NewServiceWithNotifications(
		store, fixedResolver{"99"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher,
	)

	created, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(store.persistedRevisionNotifications) != 1 {
		t.Fatalf("notifications=%#v", store.persistedRevisionNotifications)
	}
	intent := store.persistedRevisionNotifications[0]
	if intent.Kind != string(workflowdomain.NotificationKindInstanceFormRevisionRequested) ||
		intent.RecipientUserID != "99" || intent.Payload.SourceType != "workflow_form_revision" ||
		intent.Payload.SourceID != created.ID || intent.Payload.View != "workflow:form-revision-detail:"+created.ID {
		t.Fatalf("intent=%#v", intent)
	}
	if !store.revisionNotificationsInTransaction {
		t.Fatal("revision notifications must be persisted in the workflow transaction")
	}
	assertNotificationDispatch(t, dispatcher, store.revisionNotificationOutboxIDs)
}

func TestCreateDirectRevisionPersistsBusinessEventInTransaction(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	state.Instance.BusinessType = "purchase"
	state.Instance.BusinessKey = "PO-1"
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
	if len(store.createdBusinessEvents) != 1 || !store.businessEventsInTransaction {
		t.Fatalf("events=%#v inTransaction=%v", store.createdBusinessEvents, store.businessEventsInTransaction)
	}
	event := store.createdBusinessEvents[0].Event
	if event.Type != LifecycleInstanceFormRevised || event.RevisionRequestID != result.ID ||
		event.FormRevision != 4 || !reflect.DeepEqual(event.FormPatch, map[string]interface{}{"remark": "已更正"}) {
		t.Fatalf("event=%#v", event)
	}
}

func TestCompleteRevisionTaskAppliesOnlyAfterRequiredApprovals(t *testing.T) {
	service, store := reviewedRevisionServiceFixture()
	created, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil {
		t.Fatal(err)
	}
	for _, nodeID := range []string{"finance", "legal"} {
		taskID := revisionTaskIDByNode(store.createdFormRevision, nodeID)
		result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{
			TaskID: taskID, ActorID: "99", Action: workflowdomain.RevisionActionApprove,
		})
		if err != nil || result.Status != "pending" {
			t.Fatalf("node=%s result=%#v err=%v", nodeID, result, err)
		}
	}
	result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{
		TaskID: revisionTaskIDByNode(store.createdFormRevision, "hr"), ActorID: "99", Action: workflowdomain.RevisionActionApprove,
	})
	if err != nil || result.Status != "applied" || store.savedState == nil || store.savedState.Instance.FormRevision != 4 {
		t.Fatalf("created=%#v result=%#v state=%#v err=%v", created, result, store.savedState, err)
	}
	if len(store.createdBusinessEvents) != 1 || len(store.persistedRevisionNotifications) != 5 {
		t.Fatalf("events=%#v notifications=%#v", store.createdBusinessEvents, store.persistedRevisionNotifications)
	}
	if got := store.persistedRevisionNotifications[len(store.persistedRevisionNotifications)-1].Kind; got != string(workflowdomain.NotificationKindInstanceFormRevisionApproved) {
		t.Fatalf("last notification kind=%q", got)
	}
}

func TestRejectRevisionKeepsSourceFormUntouched(t *testing.T) {
	service, store := reviewedRevisionServiceFixture()
	before := workflowcore.MergeFormData(nil, store.state.FormData)
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil {
		t.Fatal(err)
	}
	result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{
		TaskID: revisionTaskIDByNode(store.createdFormRevision, "finance"), ActorID: "99",
		Action: workflowdomain.RevisionActionReject, Comment: "数据不正确",
	})
	if err != nil || result.Status != "rejected" || !reflect.DeepEqual(store.state.FormData, before) {
		t.Fatalf("result=%#v form=%#v err=%v", result, store.state.FormData, err)
	}
	if len(store.createdBusinessEvents) != 0 || len(store.persistedRevisionNotifications) != 3 {
		t.Fatalf("events=%#v notifications=%#v", store.createdBusinessEvents, store.persistedRevisionNotifications)
	}
}

func TestCancelRevisionAllowedBeforeFirstHandledTask(t *testing.T) {
	service, store := reviewedRevisionServiceFixture()
	created, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := service.CancelCompletedFormRevision(context.Background(), CancelCompletedFormRevisionRequest{
		RevisionID: created.ID, ActorID: "21",
	})
	if err != nil || cancelled.Status != "cancelled" {
		t.Fatalf("cancelled=%#v err=%v", cancelled, err)
	}
	if len(store.persistedRevisionNotifications) != 2 || store.persistedRevisionNotifications[1].Kind != string(workflowdomain.NotificationKindInstanceFormRevisionCancelled) {
		t.Fatalf("notifications=%#v", store.persistedRevisionNotifications)
	}

	service, store = reviewedRevisionServiceFixture()
	created, err = service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{
		TaskID: revisionTaskIDByNode(store.createdFormRevision, "finance"), ActorID: "99", Action: workflowdomain.RevisionActionApprove,
	}); err != nil {
		t.Fatal(err)
	}
	_, err = service.CancelCompletedFormRevision(context.Background(), CancelCompletedFormRevisionRequest{
		RevisionID: created.ID, ActorID: "21",
	})
	if !errors.Is(err, ErrCompletedRevisionCannotCancel) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyRevisionMarksConflictWhenSourceRevisionChanged(t *testing.T) {
	service, store := reviewedRevisionServiceFixture()
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil {
		t.Fatal(err)
	}
	store.state.Instance.FormRevision = 4
	for _, nodeID := range []string{"finance", "legal"} {
		if _, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{
			TaskID: revisionTaskIDByNode(store.createdFormRevision, nodeID), ActorID: "99", Action: workflowdomain.RevisionActionApprove,
		}); err != nil {
			t.Fatal(err)
		}
	}
	result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{
		TaskID: revisionTaskIDByNode(store.createdFormRevision, "hr"), ActorID: "99", Action: workflowdomain.RevisionActionApprove,
	})
	if err != nil || result.Status != "conflict" || store.state.Instance.FormRevision != 4 {
		t.Fatalf("result=%#v revision=%d err=%v", result, store.state.Instance.FormRevision, err)
	}
	if len(store.createdBusinessEvents) != 0 || store.persistedRevisionNotifications[len(store.persistedRevisionNotifications)-1].Kind != string(workflowdomain.NotificationKindInstanceFormRevisionConflict) {
		t.Fatalf("events=%#v notifications=%#v", store.createdBusinessEvents, store.persistedRevisionNotifications)
	}
}

func TestCompleteRevisionTaskIsIdempotent(t *testing.T) {
	service, store := reviewedRevisionServiceFixture()
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil {
		t.Fatal(err)
	}
	request := CompleteFormRevisionTaskRequest{
		TaskID: revisionTaskIDByNode(store.createdFormRevision, "finance"), ActorID: "99", Action: workflowdomain.RevisionActionApprove,
	}
	if _, err := service.CompleteFormRevisionTask(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	historyCount := len(store.state.History)
	notificationCount := len(store.persistedRevisionNotifications)
	businessEventCount := len(store.createdBusinessEvents)
	result, err := service.CompleteFormRevisionTask(context.Background(), request)
	if err != nil || !result.Idempotent || len(store.state.History) != historyCount ||
		len(store.persistedRevisionNotifications) != notificationCount || len(store.createdBusinessEvents) != businessEventCount {
		t.Fatalf("result=%#v history=%d notifications=%d events=%d err=%v", result, len(store.state.History), len(store.persistedRevisionNotifications), len(store.createdBusinessEvents), err)
	}
}

func reviewedRevisionServiceFixture() (*Service, *fakeStore) {
	definition, state := completedRevisionApplicationFixture()
	definition.Form = append(definition.Form, workflowcore.FormField{Key: "amount", Label: "金额", Type: workflowcore.FormFieldTypeNumber})
	definition.Nodes[1].FormPermissions = append(definition.Nodes[1].FormPermissions,
		workflowcore.FieldPermission{Field: "amount", Access: workflowcore.FieldAccessWrite})
	definition.Nodes = []workflowcore.Node{
		definition.Nodes[0], definition.Nodes[1],
		{ID: "parallel-split", Type: workflowcore.NodeTypeParallel, Name: "并行开始", GatewayMode: workflowcore.GatewayModeSplit},
		{ID: "finance", Type: workflowcore.NodeTypeApproval, Name: "财务审批", ApprovalMode: workflowcore.ApprovalModeParallel, CompletionRate: 100, Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "99"}},
		{ID: "legal", Type: workflowcore.NodeTypeApproval, Name: "法务审批", ApprovalMode: workflowcore.ApprovalModeParallel, CompletionRate: 100, Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "99"}},
		{ID: "parallel-join", Type: workflowcore.NodeTypeParallel, Name: "并行汇合", GatewayMode: workflowcore.GatewayModeJoin},
		{ID: "hr", Type: workflowcore.NodeTypeApproval, Name: "HR 审批", ApprovalMode: workflowcore.ApprovalModeSingle, CompletionRate: 100, Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "99"}},
		{ID: "audit", Type: workflowcore.NodeTypeApproval, Name: "未经过审计", ApprovalMode: workflowcore.ApprovalModeSingle, CompletionRate: 100, Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "99"}},
		definition.Nodes[len(definition.Nodes)-1],
	}
	definition.Edges = []workflowcore.Edge{
		{ID: "e1", Source: "start", Target: "manager"},
		{ID: "e2", Source: "manager", Target: "parallel-split"},
		{ID: "e3", Source: "parallel-split", Target: "finance"},
		{ID: "e4", Source: "parallel-split", Target: "legal"},
		{ID: "e5", Source: "finance", Target: "parallel-join"},
		{ID: "e6", Source: "legal", Target: "parallel-join"},
		{ID: "e7", Source: "parallel-join", Target: "hr"},
		{ID: "e8", Source: "hr", Target: "end"},
		{ID: "e9", Source: "manager", Target: "audit"},
		{ID: "e10", Source: "audit", Target: "end"},
	}
	state.FormData["amount"] = float64(100)
	state.History = append(state.History,
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "manager", ActorID: "21", EventTime: 100},
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "finance", ActorID: "31", EventTime: 200},
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "legal", ActorID: "32", EventTime: 200},
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "hr", ActorID: "33", EventTime: 300},
	)
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})
	return service, store
}

func reviewedRevisionRequest() CreateCompletedFormRevisionRequest {
	return CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: "21",
		ExpectedRevision: 3, FormData: map[string]interface{}{"amount": float64(200)}, Reason: "修正金额",
	}
}

func revisionTaskNodeStages(tasks []workflowdomain.FormRevisionTask) string {
	ordered := append([]workflowdomain.FormRevisionTask(nil), tasks...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Stage == ordered[j].Stage {
			return ordered[i].NodeID < ordered[j].NodeID
		}
		return ordered[i].Stage < ordered[j].Stage
	})
	values := make([]string, 0, len(ordered))
	for _, task := range ordered {
		values = append(values, task.NodeID+":"+fmtInt(task.Stage)+":"+string(task.Status))
	}
	return strings.Join(values, ",")
}

func revisionTaskIDByNode(revision *workflowdomain.FormRevisionRequest, nodeID string) string {
	if revision == nil {
		return ""
	}
	for _, task := range revision.Tasks {
		if task.NodeID == nodeID {
			return task.ID
		}
	}
	return ""
}

func fmtInt(value int) string {
	if value == 1 {
		return "1"
	}
	if value == 2 {
		return "2"
	}
	return ""
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
