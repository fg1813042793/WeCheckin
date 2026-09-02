package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestStartInstanceLoadsPublishedVersionAndPersistsState(t *testing.T) {
	store := &fakeStore{definition: simpleDefinition(), publishedVersion: 3}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID:   9,
		BusinessType:   "leave_request",
		BusinessKey:    "leave-2026-001",
		StarterID:      "7",
		OperatorID:     "99",
		AdminInitiated: true,
		Variables:      map[string]interface{}{"days": 2},
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
	if state.Instance.StarterID != "7" || state.Instance.OperatorID != "99" {
		t.Fatalf("instance identities = starter %q operator %q", state.Instance.StarterID, state.Instance.OperatorID)
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

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{DefinitionID: 9, StarterID: "7", OperatorID: "7"})
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
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1", StarterID: "7", OperatorID: "7",
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) {
		t.Fatalf("StartInstance() error = %v, want ErrFormDataInvalid", err)
	}

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-2", StarterID: "7", OperatorID: "7",
		FormData: map[string]interface{}{"reason": "出差"},
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if state.FormData["reason"] != "出差" {
		t.Fatalf("form data = %#v", state.FormData)
	}
}

func TestSaveStartDraftAcceptsPartialFormAndPersistsCurrentVersion(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true}}
	store := &fakeStore{publishedDefinition: &PublishedDefinition{
		ID: 9, Version: 3, Form: definition.Form,
		Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	draft, err := service.SaveStartDraft(context.Background(), SaveStartDraftRequest{
		DefinitionID: 9, DefinitionVersion: 3, StarterID: "7",
		FormData: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("SaveStartDraft() error = %v", err)
	}
	if draft.DefinitionVersion != 3 || store.savedDraft == nil || store.savedDraft.StarterID != "7" {
		t.Fatalf("saved draft = %#v, store draft = %#v", draft, store.savedDraft)
	}
}

func TestStartInstanceClearsOwnDraftInsideTransaction(t *testing.T) {
	store := &fakeStore{definition: simpleDefinition(), publishedVersion: 3}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-with-draft",
		StarterID: "7", OperatorID: "7", ClearStartDraft: true,
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if store.deletedDraftDefinitionID != 9 || store.deletedDraftStarterID != "7" {
		t.Fatalf("deleted draft = definition %d starter %q", store.deletedDraftDefinitionID, store.deletedDraftStarterID)
	}
	if !store.deletedDraftInTransaction {
		t.Fatal("submitted draft must be deleted in the workflow transaction")
	}
}

func TestGroupedFormFieldsFlowThroughStartAndComplete(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{{
		Key: "request_group", Label: "申请信息", Type: workflowcore.FormFieldTypeGroup,
		Fields: []workflowcore.FormField{
			{Key: "tip", Label: "填写提示", Type: workflowcore.FormFieldTypeDescription, Content: "请如实填写"},
			{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true},
			{Key: "opinion", Label: "审批意见", Type: workflowcore.FormFieldTypeTextarea},
		},
	}}
	definition.Nodes[1].FormPermissions = []workflowcore.FieldPermission{
		{Field: "reason", Access: workflowcore.FieldAccessRead},
		{Field: "opinion", Access: workflowcore.FieldAccessWrite},
	}
	store := &fakeStore{definition: definition, publishedVersion: 1}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-grouped", StarterID: "7", OperatorID: "7",
		FormData: map[string]interface{}{"reason": "出差"},
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if state.FormData["reason"] != "出差" {
		t.Fatalf("grouped start form data = %#v", state.FormData)
	}
	if _, exists := state.FormData["request_group"]; exists {
		t.Fatalf("layout group leaked into form data: %#v", state.FormData)
	}

	store.state = state
	updated, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		FormData: map[string]interface{}{"opinion": "同意"},
	})
	if err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if updated.FormData["reason"] != "出差" || updated.FormData["opinion"] != "同意" {
		t.Fatalf("grouped completed form data = %#v", updated.FormData)
	}
}

func TestStartInstanceRejectsStarterOutsideConfiguredRange(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7, 8},
	}
	service := NewService(&fakeStore{definition: definition, publishedVersion: 1}, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1",
		StarterID: "99", OperatorID: "42",
	})
	if !errors.Is(err, ErrStarterNotAllowed) {
		t.Fatalf("StartInstance() error = %v, want ErrStarterNotAllowed", err)
	}
}

func TestStartInstanceRejectsUnavailablePublishedTime(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	now := time.Date(2026, time.September, 1, 8, 0, 0, 0, location)
	definition := simpleDefinition()
	definition.Nodes[0].Availability = &workflowcore.StartAvailabilityConfig{
		Mode: workflowcore.StartAvailabilityFixed, Timezone: "Asia/Shanghai",
		StartsAt: now.Add(time.Hour).UnixMilli(), EndsAt: now.Add(2 * time.Hour).UnixMilli(),
	}
	store := &fakeStore{definition: definition, publishedVersion: 1}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return now }

	_, err = service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1", StarterID: "7", OperatorID: "7",
	})
	if !errors.Is(err, ErrStartNotYetAvailable) {
		t.Fatalf("StartInstance() error = %v, want ErrStartNotYetAvailable", err)
	}
	if store.createdState != nil {
		t.Fatal("unavailable workflow must not create runtime state")
	}
}

func TestPublishedDefinitionsExposeCurrentAvailabilityStatus(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	now := time.Date(2026, time.September, 1, 10, 0, 0, 0, location)
	store := &fakeStore{publishedDefinitions: []PublishedDefinition{
		{ID: 1, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}, Availability: workflowcore.StartAvailabilityConfig{Mode: workflowcore.StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{2}, DailyStartTime: "09:00", DailyEndTime: "18:00"}},
		{ID: 2, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}, Availability: workflowcore.StartAvailabilityConfig{Mode: workflowcore.StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{3}, DailyStartTime: "09:00", DailyEndTime: "18:00"}},
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return now }

	definitions, err := service.ListPublishedDefinitionsForStarter(context.Background(), "7")
	if err != nil {
		t.Fatalf("ListPublishedDefinitionsForStarter() error = %v", err)
	}
	if len(definitions) != 2 || definitions[0].AvailabilityStatus != workflowcore.StartAvailabilityStateAvailable || definitions[1].AvailabilityStatus != workflowcore.StartAvailabilityStateOutsideWindow {
		t.Fatalf("availability statuses = %#v", definitions)
	}
}

func TestStartInstanceRejectsStarterOutsideOperatorDataScope(t *testing.T) {
	service := NewService(&fakeStore{
		definition: simpleDefinition(), publishedVersion: 1, denyStarterAccess: true,
	}, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1",
		StarterID: "7", OperatorID: "42", AdminInitiated: true,
	})
	if !errors.Is(err, ErrStarterAccessDenied) {
		t.Fatalf("StartInstance() error = %v, want ErrStarterAccessDenied", err)
	}
}

func TestAdminStartChecksDataScopeWhenAdminAndUserIDsMatch(t *testing.T) {
	store := &fakeStore{
		definition: simpleDefinition(), publishedVersion: 1, denyStarterAccess: true,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1",
		StarterID: "7", OperatorID: "7", AdminInitiated: true,
	})
	if !errors.Is(err, ErrStarterAccessDenied) {
		t.Fatalf("StartInstance() error = %v, want ErrStarterAccessDenied", err)
	}
	if store.operatorAccessChecks != 1 {
		t.Fatalf("operator data scope checks = %d, want 1", store.operatorAccessChecks)
	}
}

func TestNonAdminStartCannotDelegateToAnotherUser(t *testing.T) {
	store := &fakeStore{definition: simpleDefinition(), publishedVersion: 1}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-1",
		StarterID: "7", OperatorID: "42",
	})
	if !errors.Is(err, ErrStarterAccessDenied) {
		t.Fatalf("StartInstance() error = %v, want ErrStarterAccessDenied", err)
	}
	if store.operatorAccessChecks != 0 {
		t.Fatalf("non-admin start must not invoke admin data scope checks, got %d", store.operatorAccessChecks)
	}
}

func TestPublishedInitiatorAllowsConfiguredUsersAndDepartments(t *testing.T) {
	if !publishedInitiatorAllows(workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}, "99", nil) {
		t.Fatal("all-user initiator scope should allow any active user")
	}
	configured := workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7, 8}, DepartmentIDs: []uint{3, 5},
	}
	if !publishedInitiatorAllows(configured, "7", nil) {
		t.Fatal("specified initiator scope should allow an explicitly configured user")
	}
	if !publishedInitiatorAllows(configured, "99", []uint{9, 5}) {
		t.Fatal("specified initiator scope should allow a user in any configured department")
	}
	if publishedInitiatorAllows(configured, "99", []uint{9, 10}) || publishedInitiatorAllows(configured, "99", nil) {
		t.Fatal("specified initiator scope should reject users outside configured users and departments")
	}
}

func TestListPublishedDefinitionsForStarterFiltersByDepartmentWithOneLookup(t *testing.T) {
	store := &fakeStore{
		publishedDefinitions: []PublishedDefinition{
			{ID: 1, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3}}},
			{ID: 2, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{8}}},
			{ID: 3, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}},
		},
		userDepartmentIDs: []uint{3, 5},
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	definitions, err := service.ListPublishedDefinitionsForStarter(context.Background(), "7")
	if err != nil {
		t.Fatalf("ListPublishedDefinitionsForStarter() error = %v", err)
	}
	if len(definitions) != 2 || definitions[0].ID != 1 || definitions[1].ID != 3 {
		t.Fatalf("filtered definitions = %#v", definitions)
	}
	if store.userDepartmentQueries != 1 || store.userDepartmentUserID != "7" {
		t.Fatalf("department lookups = %d for user %q, want one lookup for user 7", store.userDepartmentQueries, store.userDepartmentUserID)
	}
}

func TestGetPublishedDefinitionForStarterChecksDepartment(t *testing.T) {
	store := &fakeStore{
		publishedDefinition: &PublishedDefinition{
			ID: 1, Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3}},
		},
		userDepartmentIDs: []uint{3, 5},
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	definition, err := service.GetPublishedDefinitionForStarter(context.Background(), 1, "7")
	if err != nil || definition.ID != 1 {
		t.Fatalf("GetPublishedDefinitionForStarter() definition=%#v error=%v", definition, err)
	}
	store.userDepartmentIDs = []uint{5}
	if _, err := service.GetPublishedDefinitionForStarter(context.Background(), 1, "7"); !errors.Is(err, ErrStarterNotAllowed) {
		t.Fatalf("GetPublishedDefinitionForStarter() error = %v, want ErrStarterNotAllowed", err)
	}
}

func TestStartInstanceAllowsStarterInConfiguredDepartment(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3},
	}
	store := &fakeStore{definition: definition, publishedVersion: 1, userDepartmentIDs: []uint{3, 5}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	if _, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-dept-1",
		StarterID: "7", OperatorID: "7",
	}); err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if store.userDepartmentQueries != 1 || store.userDepartmentUserID != "7" {
		t.Fatalf("department lookups = %d for user %q", store.userDepartmentQueries, store.userDepartmentUserID)
	}
}

func TestStartInstanceRejectsStarterOutsideConfiguredDepartment(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3},
	}
	service := NewService(&fakeStore{
		definition: definition, publishedVersion: 1, userDepartmentIDs: []uint{5},
	}, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-dept-2",
		StarterID: "7", OperatorID: "7",
	})
	if !errors.Is(err, ErrStarterNotAllowed) {
		t.Fatalf("StartInstance() error = %v, want ErrStarterNotAllowed", err)
	}
}

func TestAdminStartChecksBusinessStartersDepartment(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3},
	}
	store := &fakeStore{definition: definition, publishedVersion: 1, userDepartmentIDs: []uint{3}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	if _, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-admin-dept",
		StarterID: "7", OperatorID: "42", AdminInitiated: true,
	}); err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if store.userDepartmentUserID != "7" {
		t.Fatalf("department range checked user %q, want business starter 7", store.userDepartmentUserID)
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

func TestCompleteTaskRejectsNodeFormPatchThatClearsRequiredField(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{
		{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true},
	}
	definition.Nodes[1].FormPermissions = []workflowcore.FieldPermission{
		{Field: "reason", Access: workflowcore.FieldAccessWrite},
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
		FormData: map[string]interface{}{"reason": ""},
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) {
		t.Fatalf("CompleteTask() error = %v, want ErrFormDataInvalid", err)
	}
	if store.savedState != nil {
		t.Fatal("invalid merged form data must not be saved")
	}
}

func TestCompleteTaskRequiresDetailListRowActions(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{
		{
			Key:     "objectives",
			Label:   "我的目标",
			Type:    workflowcore.FormFieldTypeDetailList,
			RowKey:  "id",
			MinRows: 1,
			MaxRows: 20,
			Columns: []workflowcore.FormField{
				{Key: "target", Label: "目标", Type: workflowcore.FormFieldTypeTextarea, Required: true},
				{Key: "weight", Label: "权重", Type: workflowcore.FormFieldTypeNumber, Min: numberPointer(0), Max: numberPointer(100)},
				{Key: "result", Label: "结果", Type: workflowcore.FormFieldTypeTextarea},
			},
		},
	}
	definition.Nodes[1].FormPermissions = []workflowcore.FieldPermission{
		{Field: "objectives", Access: workflowcore.FieldAccessWrite},
	}
	initialRows := []interface{}{
		map[string]interface{}{"id": "obj-1", "target": "提升续费率", "weight": 40, "result": "待跟进"},
	}
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	state, err := engine.Start(definition, workflowdomain.StartRequest{
		DefinitionID: 9, DefinitionVersion: 1, StarterID: "7", FormData: map[string]interface{}{"objectives": initialRows},
	})
	if err != nil {
		t.Fatalf("prepare state: %v", err)
	}
	addRows := []interface{}{
		map[string]interface{}{"id": "obj-1", "target": "提升续费率", "weight": 40, "result": "已拜访"},
		map[string]interface{}{"id": "obj-2", "target": "新增专项", "weight": 10, "result": ""},
	}
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err = service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		FormData: map[string]interface{}{"objectives": addRows},
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) {
		t.Fatalf("CompleteTask() error = %v, want ErrFormDataInvalid", err)
	}
	if store.savedState != nil {
		t.Fatal("detail row add without action must not be saved")
	}

	definition.Nodes[1].FormPermissions[0].Actions = []string{workflowcore.FieldActionAdd}
	store = &fakeStore{definition: definition, state: state}
	service = NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	updated, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		FormData: map[string]interface{}{"objectives": addRows},
	})
	if err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	rows, ok := updated.FormData["objectives"].([]interface{})
	if !ok || len(rows) != 2 {
		t.Fatalf("objectives form data = %#v", updated.FormData["objectives"])
	}
}

func TestResumeTimersLocksInstanceAndPersistsDueTransition(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "delayed_finish",
		Name:          "延时完成",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "timer", Type: workflowcore.NodeTypeTimer, Name: "等待", Timer: &workflowcore.TimerConfig{DelaySeconds: 30}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "timer"},
			{ID: "e2", Source: "timer", Target: "end"},
		},
	}
	state := &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{ID: "instance-1", DefinitionID: 7, DefinitionVersion: 3, Status: workflowdomain.InstanceStatusRunning},
		Tokens:   []workflowdomain.Token{{ID: "token-1", NodeID: "timer", Status: workflowdomain.TokenStatusWaiting}},
		Variables: map[string]interface{}{
			"__workflow_timer_due.token-1": int64(1),
		},
		FormData: map[string]interface{}{},
	}
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	resumer, ok := interface{}(service).(interface {
		ResumeTimers(context.Context, string, string) (*workflowdomain.State, int, error)
	})
	if !ok {
		t.Fatal("application timer resume support missing")
	}

	updated, advanced, err := resumer.ResumeTimers(context.Background(), "instance-1", "admin-9")
	if err != nil {
		t.Fatalf("ResumeTimers() error = %v", err)
	}
	if advanced != 1 || updated.Instance.Status != workflowdomain.InstanceStatusCompleted {
		t.Fatalf("ResumeTimers() advanced=%d state=%#v", advanced, updated.Instance)
	}
	if store.loadedInstanceID != "instance-1" || store.savedState != updated {
		t.Fatalf("resume persistence mismatch: loaded=%q saved=%p updated=%p", store.loadedInstanceID, store.savedState, updated)
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
	if store.instanceQuery.StarterID != "" || store.instanceQuery.Scope != InstanceScopeStarted || store.instanceQuery.ScopeUserID != "7" {
		t.Fatalf("authenticated instance scope = %#v", store.instanceQuery)
	}
	_, _ = service.ListMyTasks(context.Background(), "7", TaskQuery{AssigneeID: "99"})
	if store.taskQuery.AssigneeID != "7" {
		t.Fatalf("assignee filter = %q, want 7", store.taskQuery.AssigneeID)
	}
}

func TestStartInstancePersistsRuntimeEffectsInSameTransaction(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[1].Notification = &workflowcore.NotificationConfig{
		Enabled: true, Channels: []string{workflowcore.NotificationChannelInApp}, Title: "待办", Content: "请处理 {{nodeName}}",
	}
	store := &fakeStore{definition: definition, publishedVersion: 1, effectOutboxIDs: []string{"outbox-1"}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "effects-ok", StarterID: "7", OperatorID: "7",
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if store.persistedEffects != state || store.persistEffectsCalls != 1 {
		t.Fatalf("effects persistence = state %p calls %d", store.persistedEffects, store.persistEffectsCalls)
	}
	if len(state.NotificationIntents) != 1 {
		t.Fatalf("notification intents = %#v", state.NotificationIntents)
	}
}

func TestRuntimeEffectPersistenceFailureFailsWorkflowTransaction(t *testing.T) {
	store := &fakeStore{
		definition: simpleDefinition(), publishedVersion: 1,
		persistEffectsErr: errors.New("outbox unavailable"),
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "effects-fail", StarterID: "7", OperatorID: "7",
	})
	if err == nil || !strings.Contains(err.Error(), "outbox unavailable") {
		t.Fatalf("StartInstance() error = %v, want effect persistence failure", err)
	}
	if store.persistEffectsCalls != 1 {
		t.Fatalf("effect persistence calls = %d, want 1", store.persistEffectsCalls)
	}
}

func TestListMyInstancesAppliesAuthenticatedScope(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	for _, test := range []struct {
		scope string
		want  string
	}{
		{scope: "", want: InstanceScopeStarted},
		{scope: InstanceScopeStarted, want: InstanceScopeStarted},
		{scope: InstanceScopeHandled, want: InstanceScopeHandled},
		{scope: InstanceScopeCopied, want: InstanceScopeCopied},
	} {
		_, err := service.ListMyInstances(context.Background(), "7", InstanceQuery{Scope: test.scope, StarterID: "99"})
		if err != nil {
			t.Fatalf("ListMyInstances(%q) error = %v", test.scope, err)
		}
		if store.instanceQuery.Scope != test.want || store.instanceQuery.ScopeUserID != "7" || store.instanceQuery.StarterID != "" {
			t.Fatalf("ListMyInstances(%q) query = %#v", test.scope, store.instanceQuery)
		}
	}
	if _, err := service.ListMyInstances(context.Background(), "7", InstanceQuery{Scope: "forged"}); !errors.Is(err, ErrInstanceScopeInvalid) {
		t.Fatalf("invalid scope error = %v, want ErrInstanceScopeInvalid", err)
	}
}

func TestGetMyInstanceAllowsCCParticipantReadOnlyAccess(t *testing.T) {
	store := &fakeStore{
		instanceDetail: &InstanceDetail{Instance: InstanceSummary{ID: "instance-1", StarterID: "starter"}},
		hasParticipant: true,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	detail, err := service.GetMyInstance(context.Background(), "cc-user", "instance-1")
	if err != nil {
		t.Fatalf("GetMyInstance() error = %v", err)
	}
	if detail.Instance.ID != "instance-1" || store.participantUserID != "cc-user" || store.participantRole != string(workflowdomain.ParticipantRoleCC) {
		t.Fatalf("cc access lookup = detail %#v user %q role %q", detail.Instance, store.participantUserID, store.participantRole)
	}

	store.hasParticipant = false
	if _, err := service.GetMyInstance(context.Background(), "other", "instance-1"); !errors.Is(err, ErrInstanceAccessDenied) {
		t.Fatalf("non-participant access error = %v", err)
	}
}

func TestStartInstancePublishesLifecycleEventOnlyAfterPersistence(t *testing.T) {
	publisher := &recordingPublisher{}
	successStore := &fakeStore{definition: simpleDefinition(), publishedVersion: 1}
	successService := NewServiceWithPublisher(successStore, fixedResolver{"42"}, &sequenceIDs{}, publisher)

	state, err := successService.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-success", StarterID: "7", OperatorID: "99", AdminInitiated: true,
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Type != LifecycleInstanceStarted || publisher.events[0].InstanceID != state.Instance.ID {
		t.Fatalf("published events = %#v", publisher.events)
	}
	if publisher.events[0].ActorID != "99" {
		t.Fatalf("lifecycle actor = %q, want operator 99", publisher.events[0].ActorID)
	}

	publisher.events = nil
	failureStore := &fakeStore{definition: simpleDefinition(), publishedVersion: 1, createErr: errors.New("persist failed")}
	failureService := NewServiceWithPublisher(failureStore, fixedResolver{"42"}, &sequenceIDs{}, publisher)
	_, err = failureService.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-failure", StarterID: "7", OperatorID: "99", AdminInitiated: true,
	})
	if err == nil {
		t.Fatal("StartInstance() error = nil, want persistence failure")
	}
	if len(publisher.events) != 0 {
		t.Fatalf("failed transaction published events = %#v", publisher.events)
	}
}

func TestServiceUsesDefaultLifecyclePublisherWhenOmitted(t *testing.T) {
	store := &fakeStore{}
	fromDefaultConstructor := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	fromNilPublisher := NewServiceWithPublisher(store, fixedResolver{"42"}, &sequenceIDs{}, nil)
	want := DefaultLifecycleEventPublisher()

	if fromDefaultConstructor.publisher != want {
		t.Fatalf("NewService() publisher = %#v, want default lifecycle publisher", fromDefaultConstructor.publisher)
	}
	if fromNilPublisher.publisher != want {
		t.Fatalf("NewServiceWithPublisher(nil) publisher = %#v, want default lifecycle publisher", fromNilPublisher.publisher)
	}
}

func TestServicePublishesBusinessLifecycleEventsForEveryTerminalStatus(t *testing.T) {
	tests := []struct {
		name      string
		wantType  LifecycleEventType
		operation func(*Service, *workflowdomain.State) error
	}{
		{
			name:     "completed",
			wantType: LifecycleInstanceCompleted,
			operation: func(service *Service, state *workflowdomain.State) error {
				_, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
					TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
				})
				return err
			},
		},
		{
			name:     "rejected",
			wantType: LifecycleInstanceRejected,
			operation: func(service *Service, state *workflowdomain.State) error {
				_, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
					TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionReject,
				})
				return err
			},
		},
		{
			name:     "withdrawn",
			wantType: LifecycleInstanceWithdrawn,
			operation: func(service *Service, _ *workflowdomain.State) error {
				_, err := service.WithdrawInstance(context.Background(), WithdrawInstanceRequest{
					InstanceID: "instance-1", ActorID: "7", Reason: "撤回",
				})
				return err
			},
		},
		{
			name:     "cancelled",
			wantType: LifecycleInstanceCancelled,
			operation: func(service *Service, _ *workflowdomain.State) error {
				_, err := service.CancelInstance(context.Background(), CancelInstanceRequest{
					InstanceID: "instance-1", ActorID: "admin-1", Reason: "取消",
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, err := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{}).Start(
				simpleDefinition(),
				workflowdomain.StartRequest{
					DefinitionID: 9, DefinitionVersion: 1,
					BusinessType: "leave_request", BusinessKey: "leave-100", StarterID: "7", OperatorID: "7",
				},
			)
			if err != nil {
				t.Fatalf("prepare state: %v", err)
			}
			publisher := &recordingPublisher{}
			store := &fakeStore{definition: simpleDefinition(), publishedVersion: 1, state: state}
			service := NewServiceWithPublisher(store, fixedResolver{"42"}, &sequenceIDs{}, publisher)

			if err := test.operation(service, state); err != nil {
				t.Fatalf("operation error = %v", err)
			}
			if len(publisher.events) == 0 {
				t.Fatal("no lifecycle event published")
			}
			got := publisher.events[len(publisher.events)-1]
			if got.Type != test.wantType || got.BusinessType != "leave_request" || got.BusinessKey != "leave-100" {
				t.Fatalf("last event = %#v, want type=%s business=leave_request/leave-100", got, test.wantType)
			}
		})
	}
}

func TestWorkflowMutationsDispatchNewNotificationsAfterCommit(t *testing.T) {
	t.Run("start ignores delivery failure after commit", func(t *testing.T) {
		store := &fakeStore{definition: simpleDefinition(), publishedVersion: 1, effectOutboxIDs: []string{"outbox-start"}}
		dispatcher := &recordingNotificationDispatcher{store: store, dispatchErr: errors.New("channel unavailable")}
		service := NewServiceWithNotifications(store, fixedResolver{"42"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher)

		state, err := service.StartInstance(context.Background(), StartInstanceRequest{
			DefinitionID: 9, BusinessType: "leave", BusinessKey: "notify-start", StarterID: "7", OperatorID: "7",
		})
		if err != nil || state == nil {
			t.Fatalf("StartInstance() state=%#v error=%v", state, err)
		}
		assertNotificationDispatch(t, dispatcher, []string{"outbox-start"})
	})

	t.Run("complete dispatches only current effects", func(t *testing.T) {
		definition := simpleDefinition()
		engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
		state, err := engine.Start(definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 1, StarterID: "7"})
		if err != nil {
			t.Fatalf("prepare state: %v", err)
		}
		store := &fakeStore{definition: definition, state: state, effectOutboxIDs: []string{"outbox-complete"}}
		dispatcher := &recordingNotificationDispatcher{store: store}
		service := NewServiceWithNotifications(store, fixedResolver{"42"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher)

		if _, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
			TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		}); err != nil {
			t.Fatalf("CompleteTask() error = %v", err)
		}
		assertNotificationDispatch(t, dispatcher, []string{"outbox-complete"})
	})

	t.Run("timer resume dispatches current effects", func(t *testing.T) {
		definition := workflowcore.Definition{
			SchemaVersion: workflowcore.CurrentSchemaVersion, Key: "timer-notify", Name: "定时通知",
			Nodes: []workflowcore.Node{
				{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
				{ID: "timer", Type: workflowcore.NodeTypeTimer, Name: "等待", Timer: &workflowcore.TimerConfig{DelaySeconds: 1}},
				{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
			},
			Edges: []workflowcore.Edge{{ID: "e1", Source: "start", Target: "timer"}, {ID: "e2", Source: "timer", Target: "end"}},
		}
		state := &workflowdomain.State{
			Instance:  workflowdomain.ProcessInstance{ID: "instance-timer", Status: workflowdomain.InstanceStatusRunning},
			Tokens:    []workflowdomain.Token{{ID: "token-timer", NodeID: "timer", Status: workflowdomain.TokenStatusWaiting}},
			Variables: map[string]interface{}{"__workflow_timer_due.token-timer": int64(1)}, FormData: map[string]interface{}{},
		}
		store := &fakeStore{definition: definition, state: state, effectOutboxIDs: []string{"outbox-timer"}}
		dispatcher := &recordingNotificationDispatcher{store: store}
		service := NewServiceWithNotifications(store, fixedResolver{"42"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher)

		if _, advanced, err := service.ResumeTimers(context.Background(), "instance-timer", "admin"); err != nil || advanced != 1 {
			t.Fatalf("ResumeTimers() advanced=%d error=%v", advanced, err)
		}
		assertNotificationDispatch(t, dispatcher, []string{"outbox-timer"})
	})
}

func assertNotificationDispatch(t *testing.T, dispatcher *recordingNotificationDispatcher, want []string) {
	t.Helper()
	if len(dispatcher.dispatches) != 1 || fmt.Sprint(dispatcher.dispatches[0]) != fmt.Sprint(want) {
		t.Fatalf("notification dispatches = %#v, want %#v", dispatcher.dispatches, want)
	}
	if dispatcher.dispatchedInTransaction {
		t.Fatal("notifications must be dispatched after the workflow transaction commits")
	}
}

type fakeStore struct {
	definition                workflowcore.Definition
	publishedVersion          int
	state                     *workflowdomain.State
	loadDefinitionID          uint
	loadDefinitionVersion     int
	loadedTaskID              string
	loadedInstanceID          string
	createdState              *workflowdomain.State
	createErr                 error
	savedState                *workflowdomain.State
	transactions              int
	instanceQuery             InstanceQuery
	taskQuery                 TaskQuery
	denyStarterAccess         bool
	operatorAccessChecks      int
	persistedEffects          *workflowdomain.State
	persistEffectsCalls       int
	persistEffectsErr         error
	effectOutboxIDs           []string
	instanceDetail            *InstanceDetail
	hasParticipant            bool
	participantUserID         string
	participantRole           string
	inTransaction             bool
	publishedDefinitions      []PublishedDefinition
	publishedDefinition       *PublishedDefinition
	userDepartmentIDs         []uint
	userDepartmentQueries     int
	userDepartmentUserID      string
	savedDraft                *StartDraft
	deletedDraftDefinitionID  uint
	deletedDraftStarterID     string
	deletedDraftInTransaction bool
}

func (store *fakeStore) InTransaction(_ context.Context, fn func(TransactionStore) error) error {
	store.transactions++
	store.inTransaction = true
	err := fn(store)
	store.inTransaction = false
	return err
}

func (store *fakeStore) ListPublishedDefinitions(context.Context) ([]PublishedDefinition, error) {
	return append([]PublishedDefinition(nil), store.publishedDefinitions...), nil
}

func (store *fakeStore) GetPublishedDefinition(context.Context, uint) (*PublishedDefinition, error) {
	if store.publishedDefinition == nil {
		return &PublishedDefinition{}, nil
	}
	result := *store.publishedDefinition
	return &result, nil
}

func (store *fakeStore) UserDepartmentIDs(_ context.Context, userID string) ([]uint, error) {
	store.userDepartmentQueries++
	store.userDepartmentUserID = userID
	return append([]uint(nil), store.userDepartmentIDs...), nil
}

func (store *fakeStore) LoadPublishedDefinition(_ context.Context, definitionID uint, version int) (workflowcore.Definition, int, error) {
	store.loadDefinitionID = definitionID
	store.loadDefinitionVersion = version
	if store.definition.Key == "" {
		return workflowcore.Definition{}, 0, errors.New("definition unavailable")
	}
	return store.definition, store.publishedVersion, nil
}

func (store *fakeStore) IsActiveUser(_ context.Context, userID string) (bool, error) {
	return strings.TrimSpace(userID) != "", nil
}

func (store *fakeStore) CanOperatorStartFor(_ context.Context, _, _ string) (bool, error) {
	store.operatorAccessChecks++
	return !store.denyStarterAccess, nil
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

func (store *fakeStore) LoadDefinitionAndStateByInstanceForUpdate(_ context.Context, instanceID string) (workflowcore.Definition, *workflowdomain.State, error) {
	store.loadedInstanceID = instanceID
	return store.definition, store.state, nil
}

func (store *fakeStore) SaveState(_ context.Context, state *workflowdomain.State) error {
	store.savedState = state
	return nil
}

func (store *fakeStore) PersistEffects(_ context.Context, state *workflowdomain.State) ([]string, error) {
	store.persistedEffects = state
	store.persistEffectsCalls++
	return append([]string(nil), store.effectOutboxIDs...), store.persistEffectsErr
}

func (store *fakeStore) GetStartDraft(context.Context, uint, string) (*StartDraft, error) {
	return store.savedDraft, nil
}

func (store *fakeStore) SaveStartDraft(_ context.Context, draft StartDraft) (*StartDraft, error) {
	store.savedDraft = &draft
	return store.savedDraft, nil
}

func (store *fakeStore) DeleteStartDraft(_ context.Context, definitionID uint, starterID string) error {
	store.deletedDraftDefinitionID = definitionID
	store.deletedDraftStarterID = starterID
	store.deletedDraftInTransaction = store.inTransaction
	return nil
}

func (store *fakeStore) ListInstances(_ context.Context, query InstanceQuery) (*InstanceList, error) {
	store.instanceQuery = query
	return &InstanceList{}, nil
}

func (store *fakeStore) GetInstance(context.Context, string) (*InstanceDetail, error) {
	if store.instanceDetail != nil {
		return store.instanceDetail, nil
	}
	return &InstanceDetail{}, nil
}

func (store *fakeStore) HasParticipant(_ context.Context, _, userID, role string) (bool, error) {
	store.participantUserID = userID
	store.participantRole = role
	return store.hasParticipant, nil
}

func (store *fakeStore) ListTasks(_ context.Context, query TaskQuery) (*TaskList, error) {
	store.taskQuery = query
	return &TaskList{}, nil
}

type fixedResolver []string

type recordingPublisher struct {
	events []LifecycleEvent
}

type recordingNotificationDispatcher struct {
	store                   *fakeStore
	dispatches              [][]string
	dispatchErr             error
	dispatchedInTransaction bool
}

func (dispatcher *recordingNotificationDispatcher) List(context.Context, NotificationQuery) (*NotificationList, error) {
	return &NotificationList{}, nil
}

func (dispatcher *recordingNotificationDispatcher) Dispatch(_ context.Context, ids []string) (int, error) {
	dispatcher.dispatches = append(dispatcher.dispatches, append([]string(nil), ids...))
	if dispatcher.store != nil && dispatcher.store.inTransaction {
		dispatcher.dispatchedInTransaction = true
	}
	return len(ids), dispatcher.dispatchErr
}

func (dispatcher *recordingNotificationDispatcher) DispatchDue(context.Context, int) (int, error) {
	return 0, nil
}

func (dispatcher *recordingNotificationDispatcher) Retry(context.Context, string) error { return nil }

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

func numberPointer(value float64) *float64 { return &value }
