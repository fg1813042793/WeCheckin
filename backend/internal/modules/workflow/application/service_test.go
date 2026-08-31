package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowcore "wecheckin/backend/internal/workflow"
)

func TestStartInstanceLoadsPublishedVersionAndPersistsState(t *testing.T) {
	store := &fakeStore{definition: simpleDefinition(), publishedVersion: 3}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9,
		BusinessType: "leave_request",
		BusinessKey:  "leave-2026-001",
		StarterID:    "7",
		Variables:    map[string]interface{}{"days": 2},
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if store.loadDefinitionID != 9 || store.loadDefinitionVersion != 0 {
		t.Fatalf("published definition load = (%d, %d), want (9, 0)", store.loadDefinitionID, store.loadDefinitionVersion)
	}
	if state.Instance.DefinitionVersion != 3 {
		t.Fatalf("definition version = %d, want 3", state.Instance.DefinitionVersion)
	}
	if state.Instance.BusinessType != "leave_request" || state.Instance.BusinessKey != "leave-2026-001" {
		t.Fatalf("business reference = (%q, %q)", state.Instance.BusinessType, state.Instance.BusinessKey)
	}
	if len(state.PendingTasks()) != 1 || state.PendingTasks()[0].AssigneeID != "42" {
		t.Fatalf("pending tasks = %#v", state.PendingTasks())
	}
	if store.createdState != state {
		t.Fatal("runtime state was not persisted in the transaction")
	}
	if store.transactions != 1 {
		t.Fatalf("transaction count = %d, want 1", store.transactions)
	}
}

func TestCompleteTaskLocksStateAndPersistsTransition(t *testing.T) {
	definition := simpleDefinition()
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	state, err := engine.Start(definition, workflowdomain.StartRequest{
		DefinitionID: 9, DefinitionVersion: 3, StarterID: "7",
	})
	if err != nil {
		t.Fatalf("prepare state: %v", err)
	}
	taskID := state.PendingTasks()[0].ID
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	updated, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID:  taskID,
		ActorID: "42",
		Action:  workflowdomain.TaskActionApprove,
		Comment: "同意",
	})
	if err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if store.loadedTaskID != taskID {
		t.Fatalf("locked task = %q, want %q", store.loadedTaskID, taskID)
	}
	if updated.Instance.Status != workflowdomain.InstanceStatusCompleted {
		t.Fatalf("instance status = %q, want completed", updated.Instance.Status)
	}
	if store.savedState != updated {
		t.Fatal("transitioned state was not persisted")
	}
}

func TestCompleteTaskRejectsDifferentActorWithoutSaving(t *testing.T) {
	definition := simpleDefinition()
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	state, err := engine.Start(definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 3, StarterID: "7"})
	if err != nil {
		t.Fatalf("prepare state: %v", err)
	}
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err = service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID:  state.PendingTasks()[0].ID,
		ActorID: "99",
		Action:  workflowdomain.TaskActionApprove,
	})
	if !errors.Is(err, workflowdomain.ErrTaskActorMismatch) {
		t.Fatalf("CompleteTask() error = %v, want ErrTaskActorMismatch", err)
	}
	if store.savedState != nil {
		t.Fatal("state must not be saved after actor validation failure")
	}
}

func TestStartInstanceValidatesGenericBusinessReference(t *testing.T) {
	service := NewService(&fakeStore{}, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{DefinitionID: 9, StarterID: "7"})
	if !errors.Is(err, ErrBusinessReferenceRequired) {
		t.Fatalf("StartInstance() error = %v, want ErrBusinessReferenceRequired", err)
	}
}

func TestStartInstanceValidatesAndPersistsFormData(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true}}
	store := &fakeStore{definition: definition, publishedVersion: 1}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1", StarterID: "7",
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) {
		t.Fatalf("StartInstance() error = %v, want ErrFormDataInvalid", err)
	}

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-2", StarterID: "7",
		FormData: map[string]interface{}{"reason": "出差"},
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if state.FormData["reason"] != "出差" {
		t.Fatalf("form data = %#v", state.FormData)
	}
}

func TestCompleteTaskOnlyUpdatesWritableNodeFields(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{
		{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true},
		{Key: "opinion", Label: "审批意见", Type: workflowcore.FormFieldTypeTextarea},
	}
	definition.Nodes[1].FormPermissions = []workflowcore.FieldPermission{
		{Field: "reason", Access: workflowcore.FieldAccessRead},
		{Field: "opinion", Access: workflowcore.FieldAccessWrite},
	}
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	state, err := engine.Start(definition, workflowdomain.StartRequest{
		DefinitionID: 9, DefinitionVersion: 1, StarterID: "7", FormData: map[string]interface{}{"reason": "出差"},
	})
	if err != nil {
		t.Fatalf("prepare state: %v", err)
	}
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err = service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		FormData: map[string]interface{}{"reason": "篡改"},
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) {
		t.Fatalf("CompleteTask() error = %v, want ErrFormDataInvalid", err)
	}
	if store.savedState != nil {
		t.Fatal("invalid form patch must not be saved")
	}

	updated, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		FormData: map[string]interface{}{"opinion": "同意"},
	})
	if err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if updated.FormData["opinion"] != "同意" || updated.FormData["reason"] != "出差" {
		t.Fatalf("form data = %#v", updated.FormData)
	}
}

func TestWithdrawInstanceLocksByInstanceAndPersists(t *testing.T) {
	definition := simpleDefinition()
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	state, err := engine.Start(definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 1, StarterID: "7"})
	if err != nil {
		t.Fatalf("prepare state: %v", err)
	}
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	updated, err := service.WithdrawInstance(context.Background(), WithdrawInstanceRequest{InstanceID: state.Instance.ID, ActorID: "7", Reason: "信息有误"})
	if err != nil {
		t.Fatalf("WithdrawInstance() error = %v", err)
	}
	if store.loadedInstanceID != state.Instance.ID || updated.Instance.Status != workflowdomain.InstanceStatusCancelled {
		t.Fatalf("withdraw result = %#v, loaded = %q", updated.Instance, store.loadedInstanceID)
	}
}

func TestUserScopedQueriesCannotOverrideActor(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, _ = service.ListMyInstances(context.Background(), "7", InstanceQuery{StarterID: "99", Page: 2})
	if store.instanceQuery.StarterID != "7" {
		t.Fatalf("starter filter = %q, want 7", store.instanceQuery.StarterID)
	}
	_, _ = service.ListMyTasks(context.Background(), "7", TaskQuery{AssigneeID: "99"})
	if store.taskQuery.AssigneeID != "7" {
		t.Fatalf("assignee filter = %q, want 7", store.taskQuery.AssigneeID)
	}
}

func TestStartInstancePublishesLifecycleEventOnlyAfterPersistence(t *testing.T) {
	publisher := &recordingPublisher{}
	successStore := &fakeStore{definition: simpleDefinition(), publishedVersion: 1}
	successService := NewServiceWithPublisher(successStore, fixedResolver{"42"}, &sequenceIDs{}, publisher)

	state, err := successService.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-success", StarterID: "7",
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Type != LifecycleInstanceStarted || publisher.events[0].InstanceID != state.Instance.ID {
		t.Fatalf("published events = %#v", publisher.events)
	}

	publisher.events = nil
	failureStore := &fakeStore{definition: simpleDefinition(), publishedVersion: 1, createErr: errors.New("persist failed")}
	failureService := NewServiceWithPublisher(failureStore, fixedResolver{"42"}, &sequenceIDs{}, publisher)
	_, err = failureService.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-failure", StarterID: "7",
	})
	if err == nil {
		t.Fatal("StartInstance() error = nil, want persistence failure")
	}
	if len(publisher.events) != 0 {
		t.Fatalf("failed transaction published events = %#v", publisher.events)
	}
}

type fakeStore struct {
	definition            workflowcore.Definition
	publishedVersion      int
	state                 *workflowdomain.State
	loadDefinitionID      uint
	loadDefinitionVersion int
	loadedTaskID          string
	loadedInstanceID      string
	createdState          *workflowdomain.State
	createErr             error
	savedState            *workflowdomain.State
	transactions          int
	instanceQuery         InstanceQuery
	taskQuery             TaskQuery
}

func (store *fakeStore) InTransaction(_ context.Context, fn func(TransactionStore) error) error {
	store.transactions++
	return fn(store)
}

func (store *fakeStore) ListPublishedDefinitions(context.Context) ([]PublishedDefinition, error) {
	return nil, nil
}

func (store *fakeStore) GetPublishedDefinition(context.Context, uint) (*PublishedDefinition, error) {
	return &PublishedDefinition{}, nil
}

func (store *fakeStore) LoadPublishedDefinition(_ context.Context, definitionID uint, version int) (workflowcore.Definition, int, error) {
	store.loadDefinitionID = definitionID
	store.loadDefinitionVersion = version
	if store.definition.Key == "" {
		return workflowcore.Definition{}, 0, errors.New("definition unavailable")
	}
	return store.definition, store.publishedVersion, nil
}

func (store *fakeStore) CreateState(_ context.Context, state *workflowdomain.State) error {
	store.createdState = state
	return store.createErr
}

func (store *fakeStore) LoadStateByTaskForUpdate(_ context.Context, taskID string) (workflowcore.Definition, *workflowdomain.State, error) {
	store.loadedTaskID = taskID
	return store.definition, store.state, nil
}

func (store *fakeStore) LoadStateByInstanceForUpdate(_ context.Context, instanceID string) (*workflowdomain.State, error) {
	store.loadedInstanceID = instanceID
	return store.state, nil
}

func (store *fakeStore) SaveState(_ context.Context, state *workflowdomain.State) error {
	store.savedState = state
	return nil
}

func (store *fakeStore) ListInstances(_ context.Context, query InstanceQuery) (*InstanceList, error) {
	store.instanceQuery = query
	return &InstanceList{}, nil
}

func (store *fakeStore) GetInstance(context.Context, string) (*InstanceDetail, error) {
	return &InstanceDetail{}, nil
}

func (store *fakeStore) ListTasks(_ context.Context, query TaskQuery) (*TaskList, error) {
	store.taskQuery = query
	return &TaskList{}, nil
}

type fixedResolver []string

type recordingPublisher struct {
	events []LifecycleEvent
}

func (publisher *recordingPublisher) Publish(_ context.Context, event LifecycleEvent) {
	publisher.events = append(publisher.events, event)
}

func (resolver fixedResolver) Resolve(workflowdomain.AssigneeRequest) ([]string, error) {
	return append([]string(nil), resolver...), nil
}

type sequenceIDs struct{ next int }

func (ids *sequenceIDs) NewID(prefix string) string {
	ids.next++
	return fmt.Sprintf("%s-%d", prefix, ids.next)
}

func simpleDefinition() workflowcore.Definition {
	return workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "leave",
		Name:          "请假审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "approve", Type: workflowcore.NodeTypeApproval, Name: "审批", ApprovalMode: workflowcore.ApprovalModeSingle, Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "42"}, CompletionRate: 100},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "approve"},
			{ID: "e2", Source: "approve", Target: "end"},
		},
	}
}
