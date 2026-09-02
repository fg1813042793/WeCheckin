package application

import (
	"context"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestStartInstanceIdempotentReturnsExistingBusinessInstance(t *testing.T) {
	existing := &workflowdomain.State{Instance: workflowdomain.ProcessInstance{
		ID: "instance-existing", BusinessType: "scheduled_task", BusinessKey: "run-42",
	}}
	base := &fakeStore{definition: simpleDefinition(), publishedVersion: 1}
	store := &idempotentStore{fakeStore: base, existing: existing}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	state, err := service.StartInstance(context.Background(), StartInstanceRequest{
		DefinitionID: 9, BusinessType: "scheduled_task", BusinessKey: "run-42",
		StarterID: "7", OperatorID: "7", Idempotent: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state != existing || base.createdState != nil {
		t.Fatalf("state/created = %#v / %#v", state, base.createdState)
	}
}

type idempotentStore struct {
	*fakeStore
	existing *workflowdomain.State
}

func (store *idempotentStore) InTransaction(_ context.Context, fn func(TransactionStore) error) error {
	store.transactions++
	return fn(&idempotentTransactionStore{fakeStore: store.fakeStore, existing: store.existing})
}

func (store *idempotentStore) FindStateByBusiness(context.Context, string, string) (*workflowdomain.State, bool, error) {
	return store.existing, store.existing != nil, nil
}

type idempotentTransactionStore struct {
	*fakeStore
	existing *workflowdomain.State
}

func (store *idempotentTransactionStore) FindStateByBusiness(context.Context, string, string) (*workflowdomain.State, bool, error) {
	return store.existing, store.existing != nil, nil
}
