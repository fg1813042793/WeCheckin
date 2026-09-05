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
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{
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
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 3, StarterID: "7"})
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
	definition.Form = []workflowcore.FormField{
		{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true},
		{Key: "managerComment", Label: "主管意见", Type: workflowcore.FormFieldTypeTextarea, Required: true},
	}
	definition.Nodes[0].FormPermissions = []workflowcore.FieldPermission{
		{Field: "reason", Access: workflowcore.FieldAccessWrite},
		{Field: "managerComment", Access: workflowcore.FieldAccessRead},
	}
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

func TestStartInstanceCalculatesServerOwnedFormFields(t *testing.T) {
	precision := 2
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{
		{Key: "quantity", Label: "数量", Type: workflowcore.FormFieldTypeNumber},
		{Key: "price", Label: "单价", Type: workflowcore.FormFieldTypeAmount},
		{Key: "total", Label: "合计", Type: workflowcore.FormFieldTypeCalculation, Calculation: &workflowcore.FormCalculation{
			Expression: "[quantity] * [price]", Display: workflowcore.CalculationDisplayField, Precision: &precision,
		}},
	}
	store := &fakeStore{definition: definition, publishedVersion: 1}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "expense", BusinessKey: "expense-1", StarterID: "7", OperatorID: "7",
		FormData: map[string]interface{}{"quantity": 3, "price": 12.345},
	})
	if err != nil {
		t.Fatalf("StartInstance() error = %v", err)
	}
	if state.FormData["total"] != 37.04 {
		t.Fatalf("calculated start form data = %#v", state.FormData)
	}
}

func TestSaveStartDraftAcceptsPartialFormAndPersistsCurrentVersion(t *testing.T) {
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{
		{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea, Required: true},
		{
			Key: "targets", Label: "目标", Type: workflowcore.FormFieldTypeDetailList,
			RowKey: "id", MinRows: 2,
			Columns: []workflowcore.FormField{
				{Key: "target", Label: "目标内容", Type: workflowcore.FormFieldTypeTextarea, Required: true},
				{Key: "weight", Label: "权重", Type: workflowcore.FormFieldTypeNumber, Required: true},
			},
		},
	}
	store := &fakeStore{publishedDefinition: &PublishedDefinition{
		ID: 9, Version: 3, Form: definition.Form,
		Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	draft, err := service.SaveStartDraft(context.Background(), SaveStartDraftRequest{
		DefinitionID: 9, DefinitionVersion: 3, StarterID: "7",
		FormData: map[string]interface{}{
			"targets": []interface{}{
				map[string]interface{}{"id": "row-1", "target": "进行中", "weight": ""},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveStartDraft() error = %v", err)
	}
	if draft.DefinitionVersion != 3 || store.savedDraft == nil || store.savedDraft.StarterID != "7" {
		t.Fatalf("saved draft = %#v, store draft = %#v", draft, store.savedDraft)
	}
	if _, ok := draft.FormData["targets"]; !ok {
		t.Fatalf("saved detail-list draft data = %#v", draft.FormData)
	}
}

func TestDeleteStartDraftDeletesOnlyAuthenticatedOwnerDraft(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	err := service.DeleteStartDraft(context.Background(), 9, "7")
	if err != nil {
		t.Fatalf("DeleteStartDraft() error = %v", err)
	}
	if store.deletedDraftDefinitionID != 9 || store.deletedDraftStarterID != "7" {
		t.Fatalf("deleted draft = definition %d starter %q", store.deletedDraftDefinitionID, store.deletedDraftStarterID)
	}
	if store.deletedDraftInTransaction {
		t.Fatal("manual draft deletion does not require a workflow transaction")
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
	if publishedInitiatorAllows(workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll, ExcludedUserIDs: []uint{99}}, "99", nil) {
		t.Fatal("all-user initiator scope should reject an excluded user")
	}
	configured := workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7, 8}, DepartmentIDs: []uint{3, 5}, ExcludedUserIDs: []uint{8, 99},
	}
	if !publishedInitiatorAllows(configured, "7", nil) {
		t.Fatal("specified initiator scope should allow an explicitly configured user")
	}
	if !publishedInitiatorAllows(configured, "98", []uint{9, 5}) {
		t.Fatal("specified initiator scope should allow a user in any configured department")
	}
	if publishedInitiatorAllows(configured, "8", nil) {
		t.Fatal("excluded users should override an explicit user allowance")
	}
	if publishedInitiatorAllows(configured, "99", []uint{9, 5}) {
		t.Fatal("excluded users should override a department allowance")
	}
	if publishedInitiatorAllows(configured, "99", []uint{9, 10}) || publishedInitiatorAllows(configured, "99", nil) {
		t.Fatal("specified initiator scope should reject users outside configured users and departments")
	}
}

func TestListPublishedDefinitionsForStarterRejectsExcludedUserWithoutDepartmentLookup(t *testing.T) {
	store := &fakeStore{publishedDefinitions: []PublishedDefinition{{
		ID: 1,
		Initiator: workflowcore.InitiatorConfig{
			Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3}, ExcludedUserIDs: []uint{7},
		},
	}}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	definitions, err := service.ListPublishedDefinitionsForStarter(context.Background(), "7")
	if err != nil {
		t.Fatalf("ListPublishedDefinitionsForStarter() error = %v", err)
	}
	if len(definitions) != 0 {
		t.Fatalf("excluded user definitions = %#v, want none", definitions)
	}
	if store.userDepartmentQueries != 0 {
		t.Fatalf("excluded user should not trigger department lookup, got %d", store.userDepartmentQueries)
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

func TestGetPublishedDefinitionForStarterDisplaysResolvedAssigneeNames(t *testing.T) {
	resolver := &displayNameResolver{names: []string{"主管张三", "主管李四"}}
	store := &fakeStore{publishedDefinition: &PublishedDefinition{
		ID:        1,
		Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
		Nodes: []PublishedNode{{
			ID: "manager", Type: workflowcore.NodeTypeApproval, Name: "上级审批",
			AssigneeDisplay: "发起人的直属上级",
			Assignee:        &workflowcore.Assignee{Type: workflowcore.AssigneeTypeManager, Value: "direct_manager"},
		}},
	}}
	service := NewService(store, resolver, &sequenceIDs{})

	definition, err := service.GetPublishedDefinitionForStarter(context.Background(), 1, "7")
	if err != nil {
		t.Fatalf("GetPublishedDefinitionForStarter() error = %v", err)
	}
	if definition.Nodes[0].AssigneeDisplay != "主管张三、主管李四" {
		t.Fatalf("resolved assignee display = %q", definition.Nodes[0].AssigneeDisplay)
	}
	if resolver.request.Instance.StarterID != "7" || resolver.request.Node.ID != "manager" {
		t.Fatalf("resolver request = %#v", resolver.request)
	}
	if store.publishedDefinition.Nodes[0].AssigneeDisplay != "发起人的直属上级" {
		t.Fatalf("stored published definition was mutated: %#v", store.publishedDefinition.Nodes[0])
	}
}

func TestGetInstanceDisplaysResolvedAssigneeNamesForInstanceStarter(t *testing.T) {
	resolver := &displayNameResolver{names: []string{"主管张三", "主管李四"}}
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{ID: "instance-1", StarterID: "7", OperatorID: "66"},
		Nodes: []PublishedNode{{
			ID: "manager", Type: workflowcore.NodeTypeApproval, Name: "上级审批",
			AssigneeDisplay: "发起人的直属上级",
			Assignee:        &workflowcore.Assignee{Type: workflowcore.AssigneeTypeManager, Value: "direct_manager"},
		}},
	}}
	service := NewService(store, resolver, &sequenceIDs{})

	detail, err := service.GetInstance(context.Background(), "instance-1")
	if err != nil {
		t.Fatalf("GetInstance() error = %v", err)
	}
	if detail.Nodes[0].AssigneeDisplay != "主管张三、主管李四" {
		t.Fatalf("resolved instance assignee display = %q", detail.Nodes[0].AssigneeDisplay)
	}
	if resolver.request.Instance.StarterID != "7" || resolver.request.Instance.OperatorID != "66" {
		t.Fatalf("resolver instance = %#v", resolver.request.Instance)
	}
}

func TestGetInstanceKeepsStarterNameForInitiatorAssigneeNode(t *testing.T) {
	resolver := &displayNameResolver{names: []string{"Foster"}}
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{ID: "instance-1", StarterID: "7", OperatorID: "66"},
		Nodes: []PublishedNode{{
			ID: "employee-confirm", Type: workflowcore.NodeTypeApproval, Name: "员工确认",
			AssigneeDisplay: "发起人",
			Assignee:        &workflowcore.Assignee{Type: workflowcore.AssigneeTypeInitiator},
		}},
		Tasks: []TaskSummary{{
			ID: "task-1", NodeID: "employee-confirm", AssigneeID: "66", AssigneeName: "David",
		}},
	}}
	service := NewService(store, resolver, &sequenceIDs{})

	detail, err := service.GetInstance(context.Background(), "instance-1")
	if err != nil {
		t.Fatalf("GetInstance() error = %v", err)
	}
	if detail.Nodes[0].AssigneeDisplay != "Foster" {
		t.Fatalf("initiator assignee display = %q, want Foster", detail.Nodes[0].AssigneeDisplay)
	}
	if resolver.request.Instance.StarterID != "7" || resolver.request.Instance.OperatorID != "66" {
		t.Fatalf("resolver instance = %#v", resolver.request.Instance)
	}
}

func TestGetInstanceCorrectsHistoricalTaskCreatedActors(t *testing.T) {
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{
			ID: "instance-1", StarterID: "76", StarterName: "Foster",
			OperatorID: "76", OperatorName: "Foster",
		},
		UserNames: map[string]string{"76": "Foster", "84": "David", "99": "Nick"},
		History: []HistorySummary{
			{ID: "created-1", EventType: string(workflowdomain.HistoryTaskCreated), ActorID: "84", ActorName: "David", EventTime: 100},
			{ID: "started", EventType: string(workflowdomain.HistoryInstanceStarted), ActorID: "76", ActorName: "Foster", EventTime: 100},
			{ID: "created-2", EventType: string(workflowdomain.HistoryTaskCreated), ActorID: "99", ActorName: "Nick", EventTime: 200},
			{ID: "approved", EventType: string(workflowdomain.HistoryTaskApproved), ActorID: "84", ActorName: "David", EventTime: 200},
		},
	}}
	service := NewService(store, fixedResolver{"76"}, &sequenceIDs{})

	detail, err := service.GetInstance(context.Background(), "instance-1")
	if err != nil {
		t.Fatalf("GetInstance() error = %v", err)
	}
	if detail.History[0].ActorID != "76" || detail.History[0].ActorName != "Foster" {
		t.Fatalf("initial task-created actor = %#v", detail.History[0])
	}
	if detail.History[2].ActorID != "84" || detail.History[2].ActorName != "David" {
		t.Fatalf("next task-created actor = %#v", detail.History[2])
	}
}

func TestGetInstancePrefersCreatedTaskAssigneesInGraph(t *testing.T) {
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{ID: "instance-1", StarterID: "7"},
		Nodes: []PublishedNode{{
			ID: "manager", Type: workflowcore.NodeTypeApproval, Name: "上级审批",
			AssigneeDisplay: "发起人的直属上级",
		}},
		Tasks: []TaskSummary{
			{ID: "task-1", NodeID: "manager", AssigneeID: "88", AssigneeName: "主管张三"},
			{ID: "task-2", NodeID: "manager", AssigneeID: "99", AssigneeName: "主管李四"},
			{ID: "task-3", NodeID: "manager", AssigneeID: "88", AssigneeName: "主管张三"},
		},
	}}
	service := NewService(store, fixedResolver{"88", "99"}, &sequenceIDs{})

	detail, err := service.GetInstance(context.Background(), "instance-1")
	if err != nil {
		t.Fatalf("GetInstance() error = %v", err)
	}
	if detail.Nodes[0].AssigneeDisplay != "主管张三、主管李四" {
		t.Fatalf("task assignee display = %q", detail.Nodes[0].AssigneeDisplay)
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

func TestStartInstanceRejectsExcludedStarterInsideConfiguredDepartment(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, DepartmentIDs: []uint{3}, ExcludedUserIDs: []uint{7},
	}
	store := &fakeStore{definition: definition, publishedVersion: 1, userDepartmentIDs: []uint{3}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	_, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "leave", BusinessKey: "leave-excluded-1",
		StarterID: "7", OperatorID: "7",
	})
	if !errors.Is(err, ErrStarterNotAllowed) {
		t.Fatalf("StartInstance() error = %v, want ErrStarterNotAllowed", err)
	}
	if store.userDepartmentQueries != 0 {
		t.Fatalf("excluded user should not trigger department lookup, got %d", store.userDepartmentQueries)
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
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{
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

func TestCompleteTaskRecalculatesDerivedFormFields(t *testing.T) {
	precision := 2
	definition := simpleDefinition()
	definition.Form = []workflowcore.FormField{
		{Key: "quantity", Label: "数量", Type: workflowcore.FormFieldTypeNumber},
		{Key: "price", Label: "单价", Type: workflowcore.FormFieldTypeAmount},
		{Key: "total", Label: "合计", Type: workflowcore.FormFieldTypeCalculation, Calculation: &workflowcore.FormCalculation{
			Expression: "[quantity] * [price]", Display: workflowcore.CalculationDisplayField, Precision: &precision,
		}},
	}
	definition.Nodes[1].FormPermissions = []workflowcore.FieldPermission{
		{Field: "quantity", Access: workflowcore.FieldAccessWrite},
		{Field: "price", Access: workflowcore.FieldAccessRead},
		{Field: "total", Access: workflowcore.FieldAccessRead},
	}
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{
		DefinitionID: 9, DefinitionVersion: 1, StarterID: "7",
		FormData: map[string]interface{}{"quantity": 2, "price": 10, "total": 20},
	})
	if err != nil {
		t.Fatalf("prepare state: %v", err)
	}
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	updated, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionApprove,
		FormData: map[string]interface{}{"quantity": 3},
	})
	if err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if updated.FormData["total"] != float64(30) {
		t.Fatalf("calculated task form data = %#v", updated.FormData)
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
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{
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
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{
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
	state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 1, StarterID: "7"})
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

func TestGetMyOverviewUsesAuthenticatedActor(t *testing.T) {
	want := WorkflowOverview{Pending: 3, Handled: 5, Started: 7, Copied: 11}
	store := &fakeStore{workflowOverview: want}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	got, err := service.GetMyOverview(context.Background(), " 7 ")
	if err != nil {
		t.Fatalf("GetMyOverview() error = %v", err)
	}
	if got == nil || *got != want {
		t.Fatalf("GetMyOverview() = %#v, want %#v", got, want)
	}
	if store.workflowOverviewActorID != "7" {
		t.Fatalf("overview actor = %q, want 7", store.workflowOverviewActorID)
	}

	if _, err := service.GetMyOverview(context.Background(), " "); !errors.Is(err, ErrActorRequired) {
		t.Fatalf("empty actor error = %v, want ErrActorRequired", err)
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

func TestCommentInstancePersistsParticipantHistory(t *testing.T) {
	commentedAt := time.Date(2026, time.September, 3, 10, 30, 0, 0, time.Local)
	store := &fakeStore{
		instanceDetail: &InstanceDetail{
			Instance: InstanceSummary{ID: "instance-1", StarterID: "starter"},
			Tasks: []TaskSummary{{
				ID: "task-1", NodeID: "manager-review", AssigneeID: "approver",
				Status: string(workflowdomain.TaskStatusPending),
			}},
		},
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return commentedAt }

	err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: " instance-1 ", ActorID: " approver ", Comment: "  请补充合同附件  ",
		Images: []workflowcore.FormAttachment{{
			ID: "uploads/workflow/2026/09/04/comment.png", Name: "comment.png",
			URL: "/uploads/workflow/2026/09/04/comment.png", MimeType: "image/png", Size: 2048,
		}},
	})
	if err != nil {
		t.Fatalf("CommentInstance() error = %v", err)
	}
	if store.appendedHistoryInstanceID != "instance-1" || store.appendedHistoryAt != commentedAt.UnixMilli() {
		t.Fatalf("appended history target = instance %q at %d", store.appendedHistoryInstanceID, store.appendedHistoryAt)
	}
	event := store.appendedHistory
	if event.ID != "history-1" || event.Type != workflowdomain.HistoryInstanceCommented || event.NodeID != "manager-review" || event.ActorID != "approver" || event.Message != "请补充合同附件" {
		t.Fatalf("appended history event = %#v", event)
	}
	if len(event.Images) != 1 || event.Images[0].Name != "comment.png" {
		t.Fatalf("appended history images = %#v", event.Images)
	}
	if !store.appendedHistoryInTransaction {
		t.Fatal("comment history must be persisted in the workflow transaction")
	}
}

func TestCommentInstancePersistsSelectedNotificationsAndDispatchesAfterCommit(t *testing.T) {
	store := &fakeStore{
		instanceDetail: &InstanceDetail{
			Instance: InstanceSummary{
				ID: "instance-1", DefinitionName: "绩效考评单", DefinitionKey: "performance", StarterID: "7", StarterName: "Foster",
			},
			Tasks: []TaskSummary{
				{ID: "task-current", NodeID: "manager-review", NodeName: "上级评价", AssigneeID: "42", AssigneeName: "David", Status: string(workflowdomain.TaskStatusPending)},
				{ID: "task-hrbp", NodeID: "hrbp-review", NodeName: "HRBP评价", AssigneeID: "84", AssigneeName: "Nick", Status: string(workflowdomain.TaskStatusWaiting)},
			},
			UserNames: map[string]string{"7": "Foster", "42": "David", "84": "Nick"},
		},
		effectOutboxIDs: []string{"outbox-comment-in-app", "outbox-comment-dingtalk"},
	}
	dispatcher := &recordingNotificationDispatcher{store: store}
	service := NewServiceWithNotifications(store, fixedResolver{"42"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher)

	err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "42", Comment: "请关注本次评分",
		Notification: &CommentNotificationRequest{
			UserIDs: []string{"7", "84", "7"},
			Channels: []string{
				workflowcore.NotificationChannelInApp,
				workflowcore.NotificationChannelDingTalkOA,
			},
		},
	})
	if err != nil {
		t.Fatalf("CommentInstance() error = %v", err)
	}
	if store.persistEffectsCalls != 1 || store.persistedEffects == nil {
		t.Fatalf("comment notification effects = calls %d state %#v", store.persistEffectsCalls, store.persistedEffects)
	}
	if !store.persistedEffectsInTransaction {
		t.Fatal("comment notification outbox must be persisted in the workflow transaction")
	}
	intents := store.persistedEffects.NotificationIntents
	if len(intents) != 2 {
		t.Fatalf("comment notification intents = %#v", intents)
	}
	for index, wantRecipient := range []string{"7", "84"} {
		intent := intents[index]
		if intent.Kind != workflowdomain.NotificationKindInstanceCommented || intent.RecipientUserID != wantRecipient {
			t.Fatalf("comment notification intent %d = %#v", index, intent)
		}
		if intent.NodeID != "manager-review" || intent.NodeName != "上级评价" || intent.DedupeKeySuffix != store.appendedHistory.ID {
			t.Fatalf("comment notification context %d = %#v", index, intent)
		}
		if intent.Config.Title != "《绩效考评单》有新评论" || !strings.Contains(intent.Config.Content, "David") || !strings.Contains(intent.Config.Content, "请关注本次评分") {
			t.Fatalf("comment notification copy %d = %#v", index, intent.Config)
		}
	}
	assertNotificationDispatch(t, dispatcher, store.effectOutboxIDs)
}

func TestCommentInstanceRejectsInvalidNotificationSelection(t *testing.T) {
	newService := func() *Service {
		return NewService(&fakeStore{instanceDetail: &InstanceDetail{
			Instance: InstanceSummary{ID: "instance-1", StarterID: "7"},
			Tasks:    []TaskSummary{{NodeID: "approve", AssigneeID: "42", Status: string(workflowdomain.TaskStatusPending)}},
		}}, fixedResolver{"42"}, &sequenceIDs{})
	}

	if err := newService().CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "42", Comment: "测试",
		Notification: &CommentNotificationRequest{
			UserIDs: []string{"7"}, Channels: []string{"email"},
		},
	}); !errors.Is(err, ErrCommentNotificationChannelInvalid) {
		t.Fatalf("invalid notification channel error = %v", err)
	}

	if err := newService().CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "42", Comment: "测试",
		Notification: &CommentNotificationRequest{
			UserIDs: []string{"outsider"}, Channels: []string{workflowcore.NotificationChannelInApp},
		},
	}); !errors.Is(err, ErrCommentNotificationRecipientDenied) {
		t.Fatalf("invalid notification recipient error = %v", err)
	}
}

func TestGetInstanceAssignsHistoricalCommentsToActiveNode(t *testing.T) {
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{ID: "instance-1", StarterID: "76", StarterName: "Foster"},
		History: []HistorySummary{
			{ID: "started", EventType: string(workflowdomain.HistoryInstanceStarted), ActorID: "76", EventTime: 100},
			{ID: "created", EventType: string(workflowdomain.HistoryTaskCreated), NodeID: "manager-review", TaskID: "task-1", EventTime: 100},
			{ID: "comment", EventType: string(workflowdomain.HistoryInstanceCommented), ActorID: "76", Message: "请关注本次评分", EventTime: 150},
			{ID: "approved", EventType: string(workflowdomain.HistoryTaskApproved), NodeID: "manager-review", TaskID: "task-1", ActorID: "84", EventTime: 200},
		},
	}}
	service := NewService(store, fixedResolver{"76"}, &sequenceIDs{})

	detail, err := service.GetInstance(context.Background(), "instance-1")
	if err != nil {
		t.Fatalf("GetInstance() error = %v", err)
	}
	if detail.History[2].NodeID != "manager-review" {
		t.Fatalf("historical comment node = %q, want manager-review", detail.History[2].NodeID)
	}
	if store.instanceDetail.History[2].NodeID != "" {
		t.Fatalf("GetInstance() must not mutate stored history: %#v", store.instanceDetail.History[2])
	}
}

func TestCommentInstanceAllowsImageOnlyComment(t *testing.T) {
	store := &fakeStore{instanceDetail: &InstanceDetail{Instance: InstanceSummary{ID: "instance-1", StarterID: "starter"}}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "starter",
		Images: []workflowcore.FormAttachment{{
			ID: "uploads/workflow/2026/09/04/evidence.jpg", Name: "evidence.jpg",
			URL: "https://static.example.com/uploads/workflow/2026/09/04/evidence.jpg", MimeType: "image/jpeg", Size: 4096,
		}},
	})
	if err != nil {
		t.Fatalf("image-only CommentInstance() error = %v", err)
	}
	if store.appendedHistory.Message != "" || len(store.appendedHistory.Images) != 1 {
		t.Fatalf("image-only history = %#v", store.appendedHistory)
	}
}

func TestCompleteTaskRequiresRejectReasonAndValidatesImages(t *testing.T) {
	definition := simpleDefinition()
	engine := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{})
	newState := func() *workflowdomain.State {
		state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 3, StarterID: "7"})
		if err != nil {
			t.Fatalf("prepare state: %v", err)
		}
		return state
	}

	state := newState()
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	_, err := service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionReject,
	})
	if !errors.Is(err, ErrTaskRejectCommentRequired) || store.savedState != nil {
		t.Fatalf("blank reject reason error = %v, saved = %#v", err, store.savedState)
	}

	state = newState()
	store = &fakeStore{definition: definition, state: state}
	service = NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	_, err = service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionReturn,
	})
	if !errors.Is(err, ErrTaskReturnCommentRequired) || store.savedState != nil {
		t.Fatalf("blank return reason error = %v, saved = %#v", err, store.savedState)
	}

	state = newState()
	store = &fakeStore{definition: definition, state: state}
	service = NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	_, err = service.CompleteTask(context.Background(), CompleteTaskRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionReject,
		Comment: "附件不符合要求",
		Images: []workflowcore.FormAttachment{{
			ID: "external/image.png", Name: "image.png", URL: "https://evil.example.com/image.png", MimeType: "image/png", Size: 10,
		}},
	})
	if !errors.Is(err, ErrWorkflowImageInvalid) || store.savedState != nil {
		t.Fatalf("invalid reject image error = %v, saved = %#v", err, store.savedState)
	}
}

func TestRemindInstanceNotifiesOnlyCurrentPendingAssigneesAfterCommit(t *testing.T) {
	now := time.Date(2026, time.September, 3, 14, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	definition := simpleDefinition()
	definition.Nodes[1].Notification = &workflowcore.NotificationConfig{
		Enabled: true, Channels: []string{workflowcore.NotificationChannelInApp},
	}
	state := &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{
			ID: "instance-1", StarterID: "7", Status: workflowdomain.InstanceStatusRunning,
		},
		Tasks: []workflowdomain.Task{
			{ID: "task-1", NodeID: "approve", NodeName: "审批", AssigneeID: "42", Status: workflowdomain.TaskStatusPending},
			{ID: "task-duplicate", NodeID: "approve", NodeName: "审批", AssigneeID: "42", Status: workflowdomain.TaskStatusPending},
			{ID: "task-waiting", NodeID: "approve", NodeName: "审批", AssigneeID: "43", Status: workflowdomain.TaskStatusWaiting},
			{ID: "task-self", NodeID: "approve", NodeName: "审批", AssigneeID: "7", Status: workflowdomain.TaskStatusPending},
			{ID: "task-other", NodeID: "other", NodeName: "其他节点", AssigneeID: "44", Status: workflowdomain.TaskStatusPending},
		},
	}
	store := &fakeStore{definition: definition, state: state, effectOutboxIDs: []string{"outbox-reminder"}}
	dispatcher := &recordingNotificationDispatcher{store: store}
	service := NewServiceWithNotifications(store, fixedResolver{"42"}, &sequenceIDs{}, noopEventPublisher{}, dispatcher)
	service.now = func() time.Time { return now }

	result, err := service.RemindInstance(context.Background(), RemindInstanceRequest{
		InstanceID: "instance-1", ActorID: "7", NodeID: "approve",
	})
	if err != nil {
		t.Fatalf("RemindInstance() error = %v", err)
	}
	if result.RemindedCount != 1 || result.NodeID != "approve" || result.RemindedAt != now.UnixMilli() {
		t.Fatalf("reminder result = %#v", result)
	}
	if result.NextAllowedAt != now.Add(30*time.Minute).UnixMilli() {
		t.Fatalf("next allowed at = %d", result.NextAllowedAt)
	}
	if len(state.NotificationIntents) != 1 {
		t.Fatalf("notification intents = %#v", state.NotificationIntents)
	}
	intent := state.NotificationIntents[0]
	if intent.Kind != workflowdomain.NotificationKindTaskReminder || intent.RecipientUserID != "42" || intent.TaskID != "task-1" {
		t.Fatalf("unexpected reminder intent = %#v", intent)
	}
	if intent.DedupeKeySuffix == "" {
		t.Fatal("reminder intent must have a per-request dedupe suffix")
	}
	lastHistory := state.History[len(state.History)-1]
	if lastHistory.Type != workflowdomain.HistoryInstanceReminded || lastHistory.ActorID != "7" || lastHistory.NodeID != "approve" || lastHistory.EventTime != now.UnixMilli() {
		t.Fatalf("reminder history = %#v", lastHistory)
	}
	assertNotificationDispatch(t, dispatcher, []string{"outbox-reminder"})
}

func TestRemindInstanceRejectsNonStarterAndRateLimits(t *testing.T) {
	now := time.Date(2026, time.September, 3, 14, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	newService := func(history []workflowdomain.HistoryEvent) (*Service, *fakeStore) {
		store := &fakeStore{
			definition: simpleDefinition(),
			state: &workflowdomain.State{
				Instance: workflowdomain.ProcessInstance{ID: "instance-1", StarterID: "7", Status: workflowdomain.InstanceStatusRunning},
				Tasks:    []workflowdomain.Task{{ID: "task-1", NodeID: "approve", NodeName: "审批", AssigneeID: "42", Status: workflowdomain.TaskStatusPending}},
				History:  history,
			},
		}
		service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
		service.now = func() time.Time { return now }
		return service, store
	}

	service, store := newService(nil)
	_, err := service.RemindInstance(context.Background(), RemindInstanceRequest{InstanceID: "instance-1", ActorID: "99", NodeID: "approve"})
	if !errors.Is(err, ErrReminderStarterOnly) || store.savedState != nil {
		t.Fatalf("non-starter error = %v, saved = %#v", err, store.savedState)
	}

	service, store = newService([]workflowdomain.HistoryEvent{{
		ID: "reminder-1", Type: workflowdomain.HistoryInstanceReminded, NodeID: "approve", ActorID: "7", EventTime: now.Add(-10 * time.Minute).UnixMilli(),
	}})
	_, err = service.RemindInstance(context.Background(), RemindInstanceRequest{InstanceID: "instance-1", ActorID: "7", NodeID: "approve"})
	if !errors.Is(err, ErrReminderCooldown) || store.savedState != nil {
		t.Fatalf("cooldown error = %v, saved = %#v", err, store.savedState)
	}

	history := []workflowdomain.HistoryEvent{
		{ID: "reminder-1", Type: workflowdomain.HistoryInstanceReminded, NodeID: "approve", ActorID: "7", EventTime: now.Add(-5 * time.Hour).UnixMilli()},
		{ID: "reminder-2", Type: workflowdomain.HistoryInstanceReminded, NodeID: "approve", ActorID: "7", EventTime: now.Add(-4 * time.Hour).UnixMilli()},
		{ID: "reminder-3", Type: workflowdomain.HistoryInstanceReminded, NodeID: "approve", ActorID: "7", EventTime: now.Add(-3 * time.Hour).UnixMilli()},
	}
	service, store = newService(history)
	_, err = service.RemindInstance(context.Background(), RemindInstanceRequest{InstanceID: "instance-1", ActorID: "7", NodeID: "approve"})
	if !errors.Is(err, ErrReminderDailyLimit) || store.savedState != nil {
		t.Fatalf("daily limit error = %v, saved = %#v", err, store.savedState)
	}
}

func TestGetMyInstanceDecoratesCurrentReminderNodes(t *testing.T) {
	now := time.Date(2026, time.September, 3, 14, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	store := &fakeStore{instanceDetail: &InstanceDetail{
		Instance: InstanceSummary{ID: "instance-1", StarterID: "7", Status: string(workflowdomain.InstanceStatusRunning)},
		Tasks: []TaskSummary{
			{ID: "task-1", NodeID: "approve", NodeName: "上级审批", AssigneeID: "42", AssigneeName: "张经理", Status: string(workflowdomain.TaskStatusPending)},
			{ID: "task-2", NodeID: "approve", NodeName: "上级审批", AssigneeID: "43", AssigneeName: "李经理", Status: string(workflowdomain.TaskStatusWaiting)},
		},
		History: []HistorySummary{{EventType: string(workflowdomain.HistoryInstanceReminded), NodeID: "approve", EventTime: now.Add(-10 * time.Minute).UnixMilli()}},
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return now }

	detail, err := service.GetMyInstance(context.Background(), "7", "instance-1")
	if err != nil {
		t.Fatalf("GetMyInstance() error = %v", err)
	}
	if detail.ReminderPolicy.CooldownSeconds != 1800 || detail.ReminderPolicy.DailyLimit != 3 {
		t.Fatalf("reminder policy = %#v", detail.ReminderPolicy)
	}
	if len(detail.ReminderNodes) != 1 {
		t.Fatalf("reminder nodes = %#v", detail.ReminderNodes)
	}
	node := detail.ReminderNodes[0]
	if node.NodeID != "approve" || len(node.AssigneeNames) != 1 || node.AssigneeNames[0] != "张经理" || node.CanRemind || node.NextAllowedAt != now.Add(20*time.Minute).UnixMilli() {
		t.Fatalf("reminder node = %#v", node)
	}
}

func TestCommentInstanceValidatesContentAndParticipantAccess(t *testing.T) {
	store := &fakeStore{
		instanceDetail: &InstanceDetail{Instance: InstanceSummary{ID: "instance-1", StarterID: "starter"}},
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	if err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "starter", Comment: "   ",
	}); !errors.Is(err, ErrInstanceCommentRequired) {
		t.Fatalf("blank comment error = %v, want ErrInstanceCommentRequired", err)
	}
	if err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "starter", Comment: strings.Repeat("评", 501),
	}); !errors.Is(err, ErrInstanceCommentTooLong) {
		t.Fatalf("long comment error = %v, want ErrInstanceCommentTooLong", err)
	}
	if err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "starter",
		Images: []workflowcore.FormAttachment{{
			ID: "uploads/workflow/2026/09/04/file.pdf", Name: "file.pdf",
			URL: "/uploads/workflow/2026/09/04/file.pdf", MimeType: "application/pdf", Size: 100,
		}},
	}); !errors.Is(err, ErrWorkflowImageInvalid) {
		t.Fatalf("invalid image error = %v, want ErrWorkflowImageInvalid", err)
	}
	if err := service.CommentInstance(context.Background(), CommentInstanceRequest{
		InstanceID: "instance-1", ActorID: "outsider", Comment: "无权限评论",
	}); !errors.Is(err, ErrInstanceAccessDenied) {
		t.Fatalf("outsider comment error = %v, want ErrInstanceAccessDenied", err)
	}
	if store.appendedHistory.ID != "" {
		t.Fatalf("invalid comment must not persist history: %#v", store.appendedHistory)
	}
}

func TestDeleteMyInstanceHidesOnlyOwnedTerminalApplication(t *testing.T) {
	deletedAt := time.Date(2026, time.September, 2, 17, 30, 0, 0, time.Local)
	store := &fakeStore{
		instanceDetail: &InstanceDetail{Instance: InstanceSummary{
			ID: "instance-1", StarterID: "7", Status: string(workflowdomain.InstanceStatusCompleted),
		}},
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return deletedAt }

	if err := service.DeleteMyInstance(context.Background(), "7", "instance-1"); err != nil {
		t.Fatalf("DeleteMyInstance() error = %v", err)
	}
	if store.hiddenInstanceID != "instance-1" || store.hiddenStarterID != "7" || store.hiddenAt != deletedAt.UnixMilli() {
		t.Fatalf("hidden application = instance %q starter %q at %d", store.hiddenInstanceID, store.hiddenStarterID, store.hiddenAt)
	}

	store.instanceDetail.Instance.Status = string(workflowdomain.InstanceStatusRunning)
	if err := service.DeleteMyInstance(context.Background(), "7", "instance-1"); !errors.Is(err, ErrRunningInstanceCannotDelete) {
		t.Fatalf("running delete error = %v, want ErrRunningInstanceCannotDelete", err)
	}
	store.instanceDetail.Instance.Status = string(workflowdomain.InstanceStatusCompleted)
	if err := service.DeleteMyInstance(context.Background(), "8", "instance-1"); !errors.Is(err, ErrInstanceAccessDenied) {
		t.Fatalf("other starter delete error = %v, want ErrInstanceAccessDenied", err)
	}
}

func TestDeleteInstancesSoftDeletesTerminalInstancesInOneTransaction(t *testing.T) {
	store := &fakeStore{deleteInstances: []InstanceSummary{
		{ID: "instance-completed", Status: string(workflowdomain.InstanceStatusCompleted)},
		{ID: "instance-rejected", Status: string(workflowdomain.InstanceStatusRejected)},
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return time.UnixMilli(1788393600123) }

	deleted, err := service.DeleteInstances(context.Background(), " 9 ", []string{" instance-completed ", "instance-rejected", "instance-completed"})
	if err != nil {
		t.Fatalf("DeleteInstances() error = %v", err)
	}
	if deleted != 2 || store.transactions != 1 || !store.softDeleteInTransaction {
		t.Fatalf("delete result = %d, transactions = %d, in transaction = %v", deleted, store.transactions, store.softDeleteInTransaction)
	}
	if fmt.Sprint(store.softDeletedInstanceIDs) != fmt.Sprint([]string{"instance-completed", "instance-rejected"}) {
		t.Fatalf("soft deleted ids = %#v", store.softDeletedInstanceIDs)
	}
	if store.softDeletedBy != "9" || store.softDeletedAt != 1788393600123 {
		t.Fatalf("delete audit = actor %q time %d", store.softDeletedBy, store.softDeletedAt)
	}
}

func TestDeleteInstancesRejectsRunningOrMissingInstancesWithoutWriting(t *testing.T) {
	tests := []struct {
		name string
		rows []InstanceSummary
		want error
	}{
		{name: "running", rows: []InstanceSummary{{ID: "instance-1", Status: string(workflowdomain.InstanceStatusRunning)}}, want: ErrRunningInstanceCannotDelete},
		{name: "missing", rows: nil, want: ErrInstanceDeleteTargetNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{deleteInstances: test.rows}
			service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

			if _, err := service.DeleteInstances(context.Background(), "9", []string{"instance-1"}); !errors.Is(err, test.want) {
				t.Fatalf("DeleteInstances() error = %v, want %v", err, test.want)
			}
			if len(store.softDeletedInstanceIDs) != 0 {
				t.Fatalf("invalid delete must not write: %#v", store.softDeletedInstanceIDs)
			}
		})
	}
}

func TestDeleteTaskSoftDeletesTerminalTaskInTransaction(t *testing.T) {
	store := &fakeStore{deleteTask: &TaskSummary{
		ID: "task-1", Status: string(workflowdomain.TaskStatusCancelled),
	}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return time.UnixMilli(1788393600123) }

	if err := service.DeleteTask(context.Background(), " 9 ", " task-1 "); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
	if store.softDeletedTaskID != "task-1" || store.softDeletedTaskBy != "9" || store.softDeletedTaskAt != 1788393600123 {
		t.Fatalf("task delete audit = task %q actor %q time %d", store.softDeletedTaskID, store.softDeletedTaskBy, store.softDeletedTaskAt)
	}
	if store.transactions != 1 || !store.softDeleteTaskInTransaction {
		t.Fatalf("task delete transaction = count %d in transaction %v", store.transactions, store.softDeleteTaskInTransaction)
	}
}

func TestDeleteTaskRejectsActiveOrMissingTaskWithoutWriting(t *testing.T) {
	tests := []struct {
		name string
		task *TaskSummary
		want error
	}{
		{name: "waiting", task: &TaskSummary{ID: "task-1", Status: string(workflowdomain.TaskStatusWaiting)}, want: ErrTaskDeleteNotAllowed},
		{name: "pending", task: &TaskSummary{ID: "task-1", Status: string(workflowdomain.TaskStatusPending)}, want: ErrTaskDeleteNotAllowed},
		{name: "missing", want: ErrTaskDeleteTargetNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{deleteTask: test.task}
			service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

			if err := service.DeleteTask(context.Background(), "9", "task-1"); !errors.Is(err, test.want) {
				t.Fatalf("DeleteTask() error = %v, want %v", err, test.want)
			}
			if store.softDeletedTaskID != "" {
				t.Fatalf("invalid task delete must not write: %q", store.softDeletedTaskID)
			}
		})
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
					TaskID: state.PendingTasks()[0].ID, ActorID: "42", Action: workflowdomain.TaskActionReject, Comment: "材料不齐全",
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
			state, err := workflowdomain.NewEngine(fixedResolver{"42"}, &sequenceIDs{}).Start(context.Background(),
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
		state, err := engine.Start(context.Background(), definition, workflowdomain.StartRequest{DefinitionID: 9, DefinitionVersion: 1, StarterID: "7"})
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
	definition                    workflowcore.Definition
	publishedVersion              int
	state                         *workflowdomain.State
	loadDefinitionID              uint
	loadDefinitionVersion         int
	loadedTaskID                  string
	loadedInstanceID              string
	createdState                  *workflowdomain.State
	createErr                     error
	savedState                    *workflowdomain.State
	transactions                  int
	instanceQuery                 InstanceQuery
	taskQuery                     TaskQuery
	denyStarterAccess             bool
	operatorAccessChecks          int
	persistedEffects              *workflowdomain.State
	persistEffectsCalls           int
	persistEffectsErr             error
	effectOutboxIDs               []string
	persistedEffectsInTransaction bool
	instanceDetail                *InstanceDetail
	hasParticipant                bool
	participantUserID             string
	participantRole               string
	inTransaction                 bool
	publishedDefinitions          []PublishedDefinition
	publishedDefinition           *PublishedDefinition
	userDepartmentIDs             []uint
	userDepartmentQueries         int
	userDepartmentUserID          string
	savedDraft                    *StartDraft
	deletedDraftDefinitionID      uint
	deletedDraftStarterID         string
	deletedDraftInTransaction     bool
	hiddenInstanceID              string
	hiddenStarterID               string
	hiddenAt                      int64
	appendedHistoryInstanceID     string
	appendedHistory               workflowdomain.HistoryEvent
	appendedHistoryAt             int64
	appendedHistoryInTransaction  bool
	deleteInstances               []InstanceSummary
	softDeletedInstanceIDs        []string
	softDeletedBy                 string
	softDeletedAt                 int64
	softDeleteInTransaction       bool
	deleteTask                    *TaskSummary
	softDeletedTaskID             string
	softDeletedTaskBy             string
	softDeletedTaskAt             int64
	softDeleteTaskInTransaction   bool
	startQuotaUsedCount           int
	startQuotaCountDefinition     uint
	startQuotaCountStarter        string
	startQuotaCountWindow         workflowcore.StartLimitWindow
	startQuotaConsumeCalls        int
	workflowOverview              WorkflowOverview
	workflowOverviewActorID       string
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

func (store *fakeStore) CountStartQuotaUsage(
	_ context.Context,
	definitionID uint,
	starterID string,
	window workflowcore.StartLimitWindow,
) (int, error) {
	store.startQuotaCountDefinition = definitionID
	store.startQuotaCountStarter = starterID
	store.startQuotaCountWindow = window
	return store.startQuotaUsedCount, nil
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

func (store *fakeStore) ConsumeStartQuota(
	_ context.Context,
	_ uint,
	_ string,
	_ workflowcore.StartLimitWindow,
	_ int,
) (int, bool, error) {
	store.startQuotaConsumeCalls++
	return store.startQuotaUsedCount + 1, true, nil
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
	store.persistedEffectsInTransaction = store.inTransaction
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

func (store *fakeStore) HideStartedInstance(_ context.Context, instanceID, starterID string, deletedAt int64) error {
	store.hiddenInstanceID = instanceID
	store.hiddenStarterID = starterID
	store.hiddenAt = deletedAt
	return nil
}

func (store *fakeStore) AppendInstanceHistory(_ context.Context, instanceID string, event workflowdomain.HistoryEvent, eventTime int64) error {
	store.appendedHistoryInstanceID = instanceID
	store.appendedHistory = event
	store.appendedHistoryAt = eventTime
	store.appendedHistoryInTransaction = store.inTransaction
	return nil
}

func (store *fakeStore) LoadInstancesForDelete(_ context.Context, _ []string) ([]InstanceSummary, error) {
	return append([]InstanceSummary(nil), store.deleteInstances...), nil
}

func (store *fakeStore) SoftDeleteInstances(_ context.Context, instanceIDs []string, actorID string, deletedAt int64) (int64, error) {
	store.softDeletedInstanceIDs = append([]string(nil), instanceIDs...)
	store.softDeletedBy = actorID
	store.softDeletedAt = deletedAt
	store.softDeleteInTransaction = store.inTransaction
	return int64(len(instanceIDs)), nil
}

func (store *fakeStore) LoadTaskForDelete(_ context.Context, _ string) (*TaskSummary, error) {
	if store.deleteTask == nil {
		return nil, nil
	}
	result := *store.deleteTask
	return &result, nil
}

func (store *fakeStore) SoftDeleteTask(_ context.Context, taskID, actorID string, deletedAt int64) (int64, error) {
	store.softDeletedTaskID = taskID
	store.softDeletedTaskBy = actorID
	store.softDeletedTaskAt = deletedAt
	store.softDeleteTaskInTransaction = store.inTransaction
	return 1, nil
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

func (store *fakeStore) GetWorkflowOverview(_ context.Context, actorID string) (*WorkflowOverview, error) {
	store.workflowOverviewActorID = actorID
	result := store.workflowOverview
	return &result, nil
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

func (resolver fixedResolver) Resolve(context.Context, workflowdomain.AssigneeRequest) ([]string, error) {
	return append([]string(nil), resolver...), nil
}

type displayNameResolver struct {
	names   []string
	request workflowdomain.AssigneeRequest
}

func (resolver *displayNameResolver) Resolve(_ context.Context, request workflowdomain.AssigneeRequest) ([]string, error) {
	resolver.request = request
	return []string{"88", "99"}, nil
}

func (resolver *displayNameResolver) ResolveDisplayNames(_ context.Context, request workflowdomain.AssigneeRequest) ([]string, error) {
	resolver.request = request
	return append([]string(nil), resolver.names...), nil
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
