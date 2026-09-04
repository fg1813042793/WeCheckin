package infrastructure

import (
	"encoding/json"
	"fmt"
	"testing"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestInstanceToModelUsesWorkflowStartTimestamp(t *testing.T) {
	row, err := instanceToModel(workflowdomain.ProcessInstance{
		ID: "instance-1", DefinitionID: 9, DefinitionVersion: 3, DefinitionKey: "performance",
		BusinessType: "performance", BusinessKey: "performance-7-2026-09", StarterID: "7",
		OperatorID: "66", Status: workflowdomain.InstanceStatusRunning, StartTime: 1788364800000,
	}, nil, 1788364800123)
	if err != nil {
		t.Fatal(err)
	}
	if row.StartTime != 1788364800000 {
		t.Fatalf("start time = %d", row.StartTime)
	}
}

func TestPopulateInstanceGraphUsesBoundDefinitionVersion(t *testing.T) {
	detail := &workflowapp.InstanceDetail{}
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{
				ID: "manager", Type: workflowcore.NodeTypeApproval, Name: "主管审批",
				Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeManager, Value: "direct_manager"},
			},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "start-manager", Source: "start", Target: "manager"},
			{ID: "manager-end", Source: "manager", Target: "end"},
		},
	}

	populateInstanceGraph(detail, definition, publishedAssigneeLabels{})

	if len(detail.Nodes) != 3 || detail.Nodes[1].ID != "manager" {
		t.Fatalf("instance graph nodes = %#v", detail.Nodes)
	}
	if len(detail.Edges) != 2 || detail.Edges[1].Target != "end" {
		t.Fatalf("instance graph edges = %#v", detail.Edges)
	}
	if detail.Nodes[1].Assignee == nil || detail.Nodes[1].Assignee.Type != workflowcore.AssigneeTypeManager {
		t.Fatalf("instance graph assignee = %#v", detail.Nodes[1].Assignee)
	}
}

func TestInstanceSummariesUseStarterNamesWithoutIDFallback(t *testing.T) {
	rows := []workflowmodel.ProcessInstance{
		{ID: "named", StarterID: "7", OperatorID: "10"},
		{ID: "account", StarterID: "8", OperatorID: "404"},
		{ID: "empty", StarterID: "9"},
		{ID: "missing", StarterID: "404"},
		{ID: "invalid", StarterID: "external-user"},
	}
	users := []model.User{
		{ID: 7, Name: " 张三 ", Account: "zhangsan"},
		{ID: 8, Account: " lisi "},
		{ID: 9},
		{ID: 10, Name: "王五"},
	}

	summaries := instanceSummaries(rows, users)
	wants := []string{"张三", "lisi", "", "", ""}
	for index, want := range wants {
		if summaries[index].StarterName != want {
			t.Fatalf("summary %q starter name = %q, want %q", summaries[index].ID, summaries[index].StarterName, want)
		}
	}
	if summaries[0].OperatorName != "王五" || summaries[1].OperatorName != "" {
		t.Fatalf("operator names = %#v", summaries)
	}
}

func TestInstanceStarterUserIDsDeduplicateNumericIDs(t *testing.T) {
	rows := []workflowmodel.ProcessInstance{
		{StarterID: "7"},
		{StarterID: "8"},
		{StarterID: "7"},
		{StarterID: "0"},
		{StarterID: "external-user"},
	}

	ids := instanceStarterUserIDs(rows)
	if len(ids) != 2 || ids[0] != 7 || ids[1] != 8 {
		t.Fatalf("starter user ids = %#v, want [7 8]", ids)
	}
}

func TestTaskSummariesUseAssigneeAndHandlerNamesWithoutIDFallback(t *testing.T) {
	rows := []workflowmodel.ProcessTask{
		{ID: "pending", InstanceID: "instance-1", AssigneeID: "7"},
		{ID: "handled", AssigneeID: "8", HandledBy: "9"},
		{ID: "missing", AssigneeID: "404"},
		{ID: "invalid", AssigneeID: "external-user"},
	}
	users := []model.User{
		{ID: 7, Name: " 张三 ", Account: "zhangsan"},
		{ID: 8, Account: " lisi "},
		{ID: 9, Name: "王五"},
		{ID: 10, Name: "赵六"},
	}
	instances := []workflowmodel.ProcessInstance{
		{ID: "instance-1", DefinitionID: 3, StarterID: "10"},
	}
	definitionNames := map[uint]string{3: "请假流程"}

	summaries, err := taskSummaries(rows, users, instances, definitionNames)
	if err != nil {
		t.Fatalf("taskSummaries() error = %v", err)
	}
	if summaries[0].AssigneeName != "张三" || summaries[0].HandledByName != "" {
		t.Fatalf("pending task names = %#v", summaries[0])
	}
	if summaries[0].DefinitionName != "请假流程" || summaries[0].StarterID != "10" || summaries[0].StarterName != "赵六" {
		t.Fatalf("pending task workflow context = %#v", summaries[0])
	}
	if summaries[1].AssigneeName != "lisi" || summaries[1].HandledByName != "王五" {
		t.Fatalf("handled task names = %#v", summaries[1])
	}
	for _, index := range []int{2, 3} {
		if summaries[index].AssigneeName != "" || summaries[index].HandledByName != "" {
			t.Fatalf("unknown task %q must not fall back to user id: %#v", summaries[index].ID, summaries[index])
		}
	}
}

func TestTaskUserIDsDeduplicateAssigneeAndHandlerIDs(t *testing.T) {
	rows := []workflowmodel.ProcessTask{
		{AssigneeID: "7"},
		{AssigneeID: "8", HandledBy: "9"},
		{AssigneeID: "7", HandledBy: "8"},
		{AssigneeID: "external-user", HandledBy: "0"},
	}

	ids := taskUserIDs(rows)
	if len(ids) != 3 || ids[0] != 7 || ids[1] != 8 || ids[2] != 9 {
		t.Fatalf("task user ids = %#v, want [7 8 9]", ids)
	}
}

func TestHistorySummariesUseActorNamesWithoutIDFallback(t *testing.T) {
	rows := []workflowmodel.ProcessHistory{
		{ID: "named", ActorID: "7"},
		{ID: "missing", ActorID: "404"},
		{ID: "system", ActorID: "system"},
	}
	users := []model.User{{ID: 7, Name: " 张三 "}}

	summaries, err := historySummaries(rows, users)
	if err != nil {
		t.Fatalf("historySummaries() error = %v", err)
	}
	if summaries[0].ActorName != "张三" || summaries[1].ActorName != "" || summaries[2].ActorName != "" {
		t.Fatalf("history actor names = %#v", summaries)
	}
}

func TestStateRoundTripPreservesHistoryEventTime(t *testing.T) {
	state, err := stateFromModels(
		workflowmodel.ProcessInstance{ID: "instance-1"}, nil, nil, nil,
		[]workflowmodel.ProcessHistory{{ID: "history-1", EventType: "instance_reminded", EventTime: 1788429600000}},
	)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if len(state.History) != 1 || state.History[0].EventTime != 1788429600000 {
		t.Fatalf("history = %#v", state.History)
	}
}

func TestNotificationDedupeKeyAllowsSeparateReminderActions(t *testing.T) {
	base := workflowdomain.NotificationIntent{
		Kind: workflowdomain.NotificationKindTaskReminder, NodeID: "approve", TaskID: "task-1", RecipientUserID: "42",
	}
	first := base
	first.DedupeKeySuffix = "reminder-1"
	second := base
	second.DedupeKeySuffix = "reminder-2"
	firstKey := notificationDedupeKey("instance-1", first, workflowcore.NotificationChannelDingTalkOA)
	secondKey := notificationDedupeKey("instance-1", second, workflowcore.NotificationChannelDingTalkOA)
	if firstKey == secondKey {
		t.Fatalf("separate reminder actions share dedupe key %q", firstKey)
	}
	if firstKey != notificationDedupeKey("instance-1", first, workflowcore.NotificationChannelDingTalkOA) {
		t.Fatal("the same reminder action must keep a stable dedupe key")
	}
}

func TestRenderNotificationPayloadUsesActionCardForCommentsOnly(t *testing.T) {
	state := &workflowdomain.State{Instance: workflowdomain.ProcessInstance{ID: "instance-1", StarterID: "7"}}
	comment := renderNotificationPayload(state, workflowdomain.NotificationIntent{
		Kind:   workflowdomain.NotificationKindInstanceCommented,
		Config: workflowcore.NotificationConfig{Title: "新评论", Content: "请查看评论"},
	}, "Foster")
	if comment.MessageType != "action_card" {
		t.Fatalf("comment message type = %q, want action_card", comment.MessageType)
	}

	regular := renderNotificationPayload(state, workflowdomain.NotificationIntent{
		Kind:   workflowdomain.NotificationKindTaskArrived,
		Config: workflowcore.NotificationConfig{Title: "待办", Content: "请处理"},
	}, "Foster")
	if regular.MessageType != "" {
		t.Fatalf("regular message type = %q, want empty", regular.MessageType)
	}
}

func TestRenderNotificationPayloadRendersApprovalResult(t *testing.T) {
	state := &workflowdomain.State{Instance: workflowdomain.ProcessInstance{ID: "instance-1", StarterID: "7"}}
	base := workflowdomain.NotificationIntent{
		NodeName: "主管审批",
		Config: workflowcore.NotificationConfig{
			Title:   "{{workflowName}}审批结果",
			Content: "你发起的流程在“{{nodeName}}”节点{{result}}",
		},
		WorkflowName: "请假申请",
	}

	approved := base
	approved.Kind = workflowdomain.NotificationKindApprovalResultApproved
	approvedPayload := renderNotificationPayload(state, approved, "Foster")
	if approvedPayload.Title != "请假申请审批结果" || approvedPayload.Content != "你发起的流程在“主管审批”节点已通过" {
		t.Fatalf("approved result payload = %#v", approvedPayload)
	}

	rejected := base
	rejected.Kind = workflowdomain.NotificationKindApprovalResultRejected
	rejectedPayload := renderNotificationPayload(state, rejected, "Foster")
	if rejectedPayload.Content != "你发起的流程在“主管审批”节点已驳回" {
		t.Fatalf("rejected result payload = %#v", rejectedPayload)
	}
}

func TestCollectInstanceUserIDsIncludesFormUserFields(t *testing.T) {
	fields := []workflowcore.FormField{
		{Key: "owner", Type: workflowcore.FormFieldTypeUser},
		{Key: "watchers", Type: workflowcore.FormFieldTypeUserMulti},
		{Key: "lines", Type: workflowcore.FormFieldTypeDetailList, Columns: []workflowcore.FormField{
			{Key: "reviewer", Type: workflowcore.FormFieldTypeUser},
		}},
	}
	formData := map[string]interface{}{
		"owner": "7", "watchers": []interface{}{"8", "7"},
		"lines": []interface{}{map[string]interface{}{"reviewer": "9"}},
	}

	ids := collectInstanceUserIDs(
		workflowmodel.ProcessInstance{StarterID: "1", OperatorID: "2"},
		[]workflowmodel.ProcessTask{{AssigneeID: "3", HandledBy: "4"}},
		[]workflowmodel.ProcessHistory{{ActorID: "5"}}, fields, formData,
	)
	if fmt.Sprint(ids) != fmt.Sprint([]uint{1, 2, 3, 4, 5, 7, 8, 9}) {
		t.Fatalf("instance user ids = %#v", ids)
	}
}

func TestTaskModelRoundTripPreservesRuntimeStatus(t *testing.T) {
	statuses := []workflowdomain.TaskStatus{
		workflowdomain.TaskStatusWaiting,
		workflowdomain.TaskStatusPending,
		workflowdomain.TaskStatusCompleted,
		workflowdomain.TaskStatusApproved,
		workflowdomain.TaskStatusRejected,
		workflowdomain.TaskStatusReturned,
		workflowdomain.TaskStatusCancelled,
	}
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			task := workflowdomain.Task{
				ID: "task-1", TokenID: "token-1", NodeID: "approve", NodeName: "审批",
				GroupKey: "group-1", AssigneeID: "42", ApprovalMode: "countersign",
				CompletionRate: 67, Sequence: 2, Total: 3, Status: status,
				Action: workflowdomain.TaskActionApprove, Comment: "同意",
				Images: []workflowcore.FormAttachment{{
					ID: "uploads/workflow/2026/09/04/approval.png", Name: "approval.png",
					URL: "/uploads/workflow/2026/09/04/approval.png", MimeType: "image/png", Size: 1024,
				}},
			}
			model := taskToModel("instance-1", task, 1234, nil)
			actual, err := taskFromModel(model)
			if err != nil {
				t.Fatalf("taskFromModel() error = %v", err)
			}

			if actual.Status != status {
				t.Fatalf("status = %q, want %q", actual.Status, status)
			}
			if actual.CompletionRate != 67 {
				t.Fatalf("completion rate = %d, want 67", actual.CompletionRate)
			}
			if len(actual.Images) != 1 || actual.Images[0].Name != "approval.png" {
				t.Fatalf("images = %#v", actual.Images)
			}
		})
	}
}

func TestStateFromModelsRestoresHistoryImages(t *testing.T) {
	state, err := stateFromModels(
		workflowmodel.ProcessInstance{ID: "instance-1"}, nil, nil, nil,
		[]workflowmodel.ProcessHistory{{
			ID: "history-1", EventType: "instance_commented",
			ImagesJSON: `[{"id":"uploads/workflow/2026/09/04/comment.webp","name":"comment.webp","url":"/uploads/workflow/2026/09/04/comment.webp","mimeType":"image/webp","size":512}]`,
		}},
	)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if len(state.History) != 1 || len(state.History[0].Images) != 1 || state.History[0].Images[0].Name != "comment.webp" {
		t.Fatalf("history images = %#v", state.History)
	}
}

func TestCompletedHandleTaskPersistsSubmitActorAndTime(t *testing.T) {
	task := workflowdomain.Task{
		ID: "task-1", TokenID: "token-1", NodeID: "handle", NodeName: "填写资料",
		AssigneeID: "42", ApprovalMode: workflowcore.ApprovalModeSingle,
		Status: workflowdomain.TaskStatusCompleted, Action: workflowdomain.TaskActionSubmit,
	}
	history := []workflowdomain.HistoryEvent{{Type: workflowdomain.HistoryTaskSubmitted, TaskID: task.ID, ActorID: "42"}}
	model := taskToModel("instance-1", task, 1234, handledActors(history))

	if model.HandledBy != "42" || model.HandledAt != 1234 || model.Action != "submit" || model.Status != "completed" {
		t.Fatalf("completed handle task model = %#v", model)
	}
}

func TestDefinitionNodeTypesIncludesExtendedNodes(t *testing.T) {
	definition := workflowcore.Definition{Nodes: []workflowcore.Node{
		{ID: "start", Type: workflowcore.NodeTypeStart},
		{ID: "handle", Type: workflowcore.NodeTypeHandle},
		{ID: "timer", Type: workflowcore.NodeTypeTimer},
	}}
	nodeTypes := definitionNodeTypes(definition)
	if nodeTypes["handle"] != workflowcore.NodeTypeHandle || nodeTypes["timer"] != workflowcore.NodeTypeTimer {
		t.Fatalf("node types = %#v", nodeTypes)
	}
}

func TestDefinitionInitiatorPreservesDepartmentScopeWithDefensiveCopies(t *testing.T) {
	definition := workflowcore.Definition{Nodes: []workflowcore.Node{{
		ID: "start", Type: workflowcore.NodeTypeStart,
		Initiator: &workflowcore.InitiatorConfig{
			Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7, 8}, DepartmentIDs: []uint{3, 5}, ExcludedUserIDs: []uint{8, 9},
		},
	}}}

	initiator := definitionInitiator(definition)
	definition.Nodes[0].Initiator.UserIDs[0] = 99
	definition.Nodes[0].Initiator.DepartmentIDs[0] = 99
	definition.Nodes[0].Initiator.ExcludedUserIDs[0] = 99

	if len(initiator.UserIDs) != 2 || initiator.UserIDs[0] != 7 || initiator.UserIDs[1] != 8 {
		t.Fatalf("initiator users = %#v", initiator.UserIDs)
	}
	if len(initiator.DepartmentIDs) != 2 || initiator.DepartmentIDs[0] != 3 || initiator.DepartmentIDs[1] != 5 {
		t.Fatalf("initiator departments = %#v", initiator.DepartmentIDs)
	}
	if len(initiator.ExcludedUserIDs) != 2 || initiator.ExcludedUserIDs[0] != 8 || initiator.ExcludedUserIDs[1] != 9 {
		t.Fatalf("initiator exclusions = %#v", initiator.ExcludedUserIDs)
	}
}

func TestDefinitionStartAvailabilityPreservesRecurringListsWithDefensiveCopies(t *testing.T) {
	definition := workflowcore.Definition{Nodes: []workflowcore.Node{{
		ID: "start", Type: workflowcore.NodeTypeStart,
		Availability: &workflowcore.StartAvailabilityConfig{
			Mode: workflowcore.StartAvailabilityMonthly, Timezone: "Asia/Shanghai",
			MonthDays: []int{1, 15}, LastDayOfMonth: true, DailyStartTime: "09:00", DailyEndTime: "18:00",
		},
	}}}

	availability := definitionStartAvailability(definition)
	definition.Nodes[0].Availability.MonthDays[0] = 30

	if availability.Mode != workflowcore.StartAvailabilityMonthly || len(availability.MonthDays) != 2 || availability.MonthDays[0] != 1 || !availability.LastDayOfMonth {
		t.Fatalf("start availability = %#v", availability)
	}
}

func TestDecodeVariableUsesJSONNumber(t *testing.T) {
	value, err := decodeVariable(`{"count":9007199254740993,"ratio":1.5}`)
	if err != nil {
		t.Fatalf("decodeVariable() error = %v", err)
	}
	object, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("decoded type = %T", value)
	}
	if object["count"] != json.Number("9007199254740993") {
		t.Fatalf("count = %#v, want json.Number", object["count"])
	}
}

func TestStateFromModelsInitializesEmptyVariables(t *testing.T) {
	state, err := stateFromModels(
		workflowmodel.ProcessInstance{ID: "instance-1", StarterID: "7", OperatorID: "42", Status: string(workflowdomain.InstanceStatusRunning)},
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if state.Variables == nil {
		t.Fatal("variables must be initialized")
	}
	if state.FormData == nil {
		t.Fatal("form data must be initialized")
	}
	if state.Instance.StarterID != "7" || state.Instance.OperatorID != "42" {
		t.Fatalf("instance identities = starter %q operator %q", state.Instance.StarterID, state.Instance.OperatorID)
	}
}

func TestStateFromModelsRestoresFormData(t *testing.T) {
	state, err := stateFromModels(
		workflowmodel.ProcessInstance{
			ID: "instance-1", Status: string(workflowdomain.InstanceStatusRunning),
			FormDataJSON: `{"reason":"出差","days":2}`,
		},
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if state.FormData["reason"] != "出差" || state.FormData["days"] != json.Number("2") {
		t.Fatalf("form data = %#v", state.FormData)
	}
}
