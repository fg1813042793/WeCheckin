package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestStartInstanceApprovalSkipPersistsHistoryAndPropagatesStoreErrors(t *testing.T) {
	dbError := errors.New("database write failed")
	for _, stage := range []string{"success", "create", "effects"} {
		t.Run(stage, func(t *testing.T) {
			definition := simpleDefinition()
			if err := json.Unmarshal([]byte(`{"skipIfNoAssignee":true}`), &definition.Nodes[1]); err != nil {
				t.Fatal(err)
			}
			store := &fakeStore{definition: definition, publishedVersion: 3}
			if stage == "create" {
				store.createErr = dbError
			} else if stage == "effects" {
				store.persistEffectsErr = dbError
			}
			service := NewService(store, fixedResolver{}, &sequenceIDs{})
			state, err := service.StartInstance(context.Background(), StartInstanceRequest{
				DefinitionID: 9, StarterID: "7", OperatorID: "7",
				BusinessType: "approval_skip_test", BusinessKey: "skip-1",
			})
			if stage != "success" {
				if !errors.Is(err, dbError) {
					t.Fatalf("store error = %v, want %v", err, dbError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if state.Instance.Status != workflowdomain.InstanceStatusCompleted || len(state.Tasks) != 0 || len(state.NotificationIntents) != 0 || store.createdState != state || store.persistedEffects != state || !store.persistedEffectsInTransaction {
				t.Fatalf("skip not persisted atomically: %#v", state)
			}
			found := false
			for _, event := range store.createdState.History {
				if event.Type == workflowdomain.HistoryEventType("node_skipped") && event.NodeID == definition.Nodes[1].ID && event.ActorID == "" {
					found = true
				}
			}
			if !found {
				t.Fatal("skip audit not persisted")
			}
		})
	}
}
