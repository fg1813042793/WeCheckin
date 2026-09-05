package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestReviseInstanceFormPersistsAuditAndSelectedNotificationsInOneTransaction(t *testing.T) {
	definition, state := revisionTestState()
	store := &fakeStore{
		definition:      definition,
		state:           state,
		effectOutboxIDs: []string{"notification-1-in_app", "notification-2-in_app"},
	}
	dispatcher := &recordingNotificationDispatcher{store: store}
	service := NewServiceWithNotifications(store, fixedResolver{"99"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher)

	updated, err := service.ReviseInstanceForm(context.Background(), ReviseInstanceFormRequest{
		InstanceID:       "instance-1",
		ActorID:          "42",
		ExpectedRevision: 3,
		FormData:         map[string]interface{}{"summary": "修订后的说明"},
		Reason:           "补充核实结果",
		Notification: &FormRevisionNotificationRequest{
			UserIDs:  []string{"7", "99"},
			Channels: []string{workflowcore.NotificationChannelInApp},
		},
	})
	if err != nil {
		t.Fatalf("ReviseInstanceForm() error = %v", err)
	}
	if updated.Instance.FormRevision != 4 || updated.FormData["summary"] != "修订后的说明" {
		t.Fatalf("updated state = revision %d, form %#v", updated.Instance.FormRevision, updated.FormData)
	}
	if store.savedState != updated || !store.persistedEffectsInTransaction {
		t.Fatal("form revision and notification outbox must be persisted in the transaction")
	}
	lastHistory := updated.History[len(updated.History)-1]
	if lastHistory.Type != workflowdomain.HistoryInstanceFormRevised || lastHistory.ActorID != "42" {
		t.Fatalf("revision history = %#v", lastHistory)
	}
	if !strings.Contains(lastHistory.Message, "修改字段：说明；") || strings.Contains(lastHistory.Message, "summary") {
		t.Fatalf("revision history message = %q", lastHistory.Message)
	}
	if len(updated.NotificationIntents) != 2 {
		t.Fatalf("notification intents = %#v", updated.NotificationIntents)
	}
	if len(dispatcher.dispatches) != 1 || dispatcher.dispatchedInTransaction {
		t.Fatalf("notification dispatches = %#v, in transaction = %v", dispatcher.dispatches, dispatcher.dispatchedInTransaction)
	}
}

func TestReviseInstanceFormRejectsStaleRevisionBeforeSaving(t *testing.T) {
	definition, state := revisionTestState()
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	_, err := service.ReviseInstanceForm(context.Background(), ReviseInstanceFormRequest{
		InstanceID:       "instance-1",
		ActorID:          "42",
		ExpectedRevision: 2,
		FormData:         map[string]interface{}{"summary": "过期修改"},
		Reason:           "补充说明",
	})
	if !errors.Is(err, ErrInstanceFormRevisionChanged) {
		t.Fatalf("ReviseInstanceForm() error = %v, want ErrInstanceFormRevisionChanged", err)
	}
	if store.savedState != nil {
		t.Fatal("stale form data must not be saved")
	}
}

func TestReviseInstanceFormRequiresAnEnabledPreviouslyHandledNode(t *testing.T) {
	definition, state := revisionTestState()
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	_, err := service.ReviseInstanceForm(context.Background(), ReviseInstanceFormRequest{
		InstanceID:       "instance-1",
		ActorID:          "88",
		ExpectedRevision: 3,
		FormData:         map[string]interface{}{"summary": "越权修改"},
		Reason:           "补充说明",
	})
	if !errors.Is(err, ErrInstanceFormRevisionNotAllowed) {
		t.Fatalf("ReviseInstanceForm() error = %v, want ErrInstanceFormRevisionNotAllowed", err)
	}
}

func TestReviseInstanceFormAllowsCopiedParticipantAsNotificationRecipient(t *testing.T) {
	definition, state := revisionTestState()
	store := &fakeStore{definition: definition, state: state, hasParticipant: true}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})

	_, err := service.ReviseInstanceForm(context.Background(), ReviseInstanceFormRequest{
		InstanceID:       "instance-1",
		ActorID:          "42",
		ExpectedRevision: 3,
		FormData:         map[string]interface{}{"summary": "通知抄送人"},
		Reason:           "补充说明",
		Notification: &FormRevisionNotificationRequest{
			UserIDs: []string{"77"}, Channels: []string{workflowcore.NotificationChannelInApp},
		},
	})
	if err != nil {
		t.Fatalf("ReviseInstanceForm() copied participant error = %v", err)
	}
	if store.participantUserID != "77" || store.participantRole != string(workflowdomain.ParticipantRoleCC) {
		t.Fatalf("participant lookup = user %q role %q", store.participantUserID, store.participantRole)
	}
}

func TestGetMyInstanceExposesPostHandleFormRevisionCapability(t *testing.T) {
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{ID: "instance-1", StarterID: "7", Status: string(workflowdomain.InstanceStatusRunning), FormRevision: 5},
		Form: []workflowcore.FormField{
			{Key: "application", Label: "申请内容", Type: workflowcore.FormFieldTypeTextarea},
			{Key: "summary", Label: "说明", Type: workflowcore.FormFieldTypeTextarea},
			{Key: "internal", Label: "内部信息", Type: workflowcore.FormFieldTypeText},
		},
		Nodes: []PublishedNode{{
			ID: "approval", Type: workflowcore.NodeTypeApproval,
			PostHandleEdit: &workflowcore.PostHandleEditConfig{Enabled: true},
		}},
		FieldPermissions: map[string][]workflowcore.FieldPermission{
			"approval": {
				{Field: "application", Access: workflowcore.FieldAccessRead},
				{Field: "summary", Access: workflowcore.FieldAccessWrite},
				{Field: "internal", Access: workflowcore.FieldAccessHidden},
			},
		},
		Tasks:   []TaskSummary{{ID: "task-1", NodeID: "approval", AssigneeID: "42", HandledBy: "42", Status: string(workflowdomain.TaskStatusApproved)}},
		History: []HistorySummary{{EventType: string(workflowdomain.HistoryTaskApproved), NodeID: "approval", ActorID: "42"}},
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	detail, err := service.GetMyInstance(context.Background(), "42", "instance-1")
	if err != nil {
		t.Fatalf("GetMyInstance() error = %v", err)
	}
	if !detail.FormRevision.Allowed || detail.FormRevision.Revision != 5 {
		t.Fatalf("form revision capability = %#v", detail.FormRevision)
	}
	if len(detail.FormRevision.FieldPermissions) != 3 {
		t.Fatalf("form revision field permissions = %#v", detail.FormRevision.FieldPermissions)
	}
	accessByField := make(map[string]string, len(detail.FormRevision.FieldPermissions))
	for _, permission := range detail.FormRevision.FieldPermissions {
		accessByField[permission.Field] = permission.Access
	}
	if accessByField["application"] != workflowcore.FieldAccessRead ||
		accessByField["summary"] != workflowcore.FieldAccessWrite ||
		accessByField["internal"] != workflowcore.FieldAccessHidden {
		t.Fatalf("form revision field permissions = %#v", detail.FormRevision.FieldPermissions)
	}
}

func revisionTestState() (workflowcore.Definition, *workflowdomain.State) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "revision_test",
		Name:          "修订测试流程",
		Form: []workflowcore.FormField{{
			Key: "summary", Label: "说明", Type: workflowcore.FormFieldTypeTextarea, Required: true,
		}},
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{
				ID: "approval", Type: workflowcore.NodeTypeApproval, Name: "主管审批",
				PostHandleEdit:  &workflowcore.PostHandleEditConfig{Enabled: true},
				FormPermissions: []workflowcore.FieldPermission{{Field: "summary", Access: workflowcore.FieldAccessWrite}},
			},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
	}
	state := &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{
			ID: "instance-1", DefinitionID: 9, DefinitionVersion: 2, DefinitionKey: definition.Key,
			StarterID: "7", OperatorID: "7", Status: workflowdomain.InstanceStatusRunning,
			FormRevision: 3,
		},
		FormData: map[string]interface{}{"summary": "原说明"},
		Tasks: []workflowdomain.Task{
			{ID: "task-1", NodeID: "approval", NodeName: "主管审批", AssigneeID: "42", Status: workflowdomain.TaskStatusApproved},
			{ID: "task-2", NodeID: "approval", NodeName: "主管审批", AssigneeID: "99", Status: workflowdomain.TaskStatusPending},
		},
		History: []workflowdomain.HistoryEvent{{
			ID: "history-1", Type: workflowdomain.HistoryTaskApproved, NodeID: "approval", TaskID: "task-1", ActorID: "42",
		}},
	}
	return definition, state
}
