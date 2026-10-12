package application

import (
	"context"
	"errors"
	"testing"
	"time"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestStartInstanceRejectsExhaustedStarterQuota(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].StartLimit = &workflowcore.StartLimitConfig{
		Mode: workflowcore.StartLimitModeLimited, Period: workflowcore.StartLimitPeriodMonth, MaxCount: 1,
	}
	store := &quotaAwareStore{
		fakeStore: &fakeStore{definition: definition, publishedVersion: 3},
		usedCount: 1,
		allowed:   false,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return time.Date(2026, 9, 3, 10, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60)) }

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "performance", BusinessKey: "performance-7-2026-09",
		StarterID: "7", OperatorID: "66", AdminInitiated: true,
	})
	if !errors.Is(err, ErrStartLimitExceeded) {
		t.Fatalf("StartInstance() error = %v, want ErrStartLimitExceeded", err)
	}
	if state != nil || store.createdState != nil {
		t.Fatalf("state/created = %#v / %#v", state, store.createdState)
	}
	if store.consumedStarterID != "7" || store.consumedDefinitionID != 9 {
		t.Fatalf("quota owner = definition %d starter %q", store.consumedDefinitionID, store.consumedStarterID)
	}
}

func TestStartInstanceRejectsRunningStarterInstance(t *testing.T) {
	store := &fakeStore{
		definition:        simpleDefinition(),
		publishedVersion:  3,
		runningStartCount: 1,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "performance", BusinessKey: "performance-7-running",
		StarterID: "7", OperatorID: "7",
	})
	if !errors.Is(err, ErrStartInstanceInProgress) {
		t.Fatalf("StartInstance() error = %v, want ErrStartInstanceInProgress", err)
	}
	if state != nil || store.createdState != nil {
		t.Fatalf("state/created = %#v / %#v", state, store.createdState)
	}
	if store.startSubmissionLocks != 1 {
		t.Fatalf("start submission locks = %d, want 1", store.startSubmissionLocks)
	}
}

func TestStartInstanceConsumesStarterQuotaBeforeCreatingState(t *testing.T) {
	definition := simpleDefinition()
	definition.Nodes[0].StartLimit = &workflowcore.StartLimitConfig{
		Mode: workflowcore.StartLimitModeLimited, Period: workflowcore.StartLimitPeriodTotal, MaxCount: 2,
	}
	store := &quotaAwareStore{
		fakeStore: &fakeStore{definition: definition, publishedVersion: 3},
		usedCount: 1,
		allowed:   true,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "performance", BusinessKey: "performance-7-2",
		StarterID: "7", OperatorID: "7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if state == nil || store.createdState == nil {
		t.Fatal("workflow state was not created")
	}
	if len(store.calls) < 3 || store.calls[0] != "guard" || store.calls[1] != "quota" || store.calls[2] != "create" {
		t.Fatalf("transaction call order = %#v", store.calls)
	}
}

func TestListPublishedDefinitionsForStarterReportsRemainingQuota(t *testing.T) {
	store := &fakeStore{
		publishedDefinitions: []PublishedDefinition{{
			ID: 9, Name: "绩效单", Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
			Availability: workflowcore.StartAvailabilityConfig{Mode: workflowcore.StartAvailabilityAlways, Timezone: "Asia/Shanghai"},
			StartLimit: workflowcore.StartLimitConfig{
				Mode: workflowcore.StartLimitModeLimited, Period: workflowcore.StartLimitPeriodMonth, MaxCount: 2,
			},
		}},
		startQuotaUsedCount: 1,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return time.Date(2026, 9, 3, 10, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60)) }

	definitions, err := service.ListPublishedDefinitionsForStarter(context.Background(), "7")
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions) != 1 {
		t.Fatalf("definitions = %#v", definitions)
	}
	status := definitions[0].StartLimitStatus
	if !status.Allowed || status.UsedCount != 1 || status.RemainingCount != 1 || status.ResetsAt == 0 {
		t.Fatalf("start limit status = %#v", status)
	}
	if store.startQuotaCountDefinition != 9 || store.startQuotaCountStarter != "7" {
		t.Fatalf("quota count owner = definition %d starter %q", store.startQuotaCountDefinition, store.startQuotaCountStarter)
	}
}

func TestListPublishedDefinitionsForStarterBlocksRunningInstanceWithoutConsumingQuota(t *testing.T) {
	store := &fakeStore{
		publishedDefinitions: []PublishedDefinition{{
			ID: 9, Name: "绩效单", Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
			Availability: workflowcore.StartAvailabilityConfig{Mode: workflowcore.StartAvailabilityAlways, Timezone: "Asia/Shanghai"},
			StartLimit:   workflowcore.StartLimitConfig{Mode: workflowcore.StartLimitModeUnlimited},
		}},
		runningStartCount: 1,
	}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	definitions, err := service.ListPublishedDefinitionsForStarter(context.Background(), "7")
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions) != 1 || definitions[0].StartLimitStatus.Allowed {
		t.Fatalf("definitions = %#v", definitions)
	}
}

func TestListPublishedDefinitionsForStarterSkipsQuotaWindowOutsideAvailability(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	store := &fakeStore{publishedDefinitions: []PublishedDefinition{{
		ID: 9, Name: "绩效单", Initiator: workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll},
		Availability: workflowcore.StartAvailabilityConfig{
			Mode: workflowcore.StartAvailabilityMonthlyWindow, Timezone: "Asia/Shanghai",
			WindowStartDayFromEnd: 7, WindowStartTime: "00:00", WindowEndDay: 10, WindowEndTime: "23:59",
		},
		StartLimit: workflowcore.StartLimitConfig{Mode: workflowcore.StartLimitModeLimited, Period: workflowcore.StartLimitPeriodAvailability, MaxCount: 1},
	}}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time {
		return time.Date(2026, time.October, 12, 3, 3, 0, 0, location)
	}

	definitions, err := service.ListPublishedDefinitionsForStarter(context.Background(), "7")
	if err != nil {
		t.Fatalf("ListPublishedDefinitionsForStarter() error = %v", err)
	}
	if len(definitions) != 1 || definitions[0].AvailabilityStatus != workflowcore.StartAvailabilityStateOutsideWindow {
		t.Fatalf("definitions = %#v", definitions)
	}
	if definitions[0].StartLimitStatus.Allowed {
		t.Fatalf("outside-window definition must not be startable: %#v", definitions[0].StartLimitStatus)
	}
	if store.startQuotaCountDefinition != 0 {
		t.Fatalf("outside-window definition must skip quota lookup, got definition %d", store.startQuotaCountDefinition)
	}
}

type quotaAwareStore struct {
	*fakeStore
	usedCount            int
	allowed              bool
	consumedDefinitionID uint
	consumedStarterID    string
	calls                []string
}

func (store *quotaAwareStore) LockStartSubmission(_ context.Context, _ uint, _ string) (int, error) {
	store.calls = append(store.calls, "guard")
	return 0, nil
}

func (store *quotaAwareStore) InTransaction(_ context.Context, fn func(TransactionStore) error) error {
	store.transactions++
	store.inTransaction = true
	err := fn(store)
	store.inTransaction = false
	return err
}

func (store *quotaAwareStore) ConsumeStartQuota(
	_ context.Context,
	definitionID uint,
	starterID string,
	_ workflowcore.StartLimitWindow,
	_ int,
) (int, bool, error) {
	store.calls = append(store.calls, "quota")
	store.consumedDefinitionID = definitionID
	store.consumedStarterID = starterID
	return store.usedCount, store.allowed, nil
}

func (store *quotaAwareStore) CreateState(ctx context.Context, state *workflowdomain.State) error {
	store.calls = append(store.calls, "create")
	return store.fakeStore.CreateState(ctx, state)
}
