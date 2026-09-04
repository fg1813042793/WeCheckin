package domain

import (
	"context"
	"errors"
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

type canceledContextResolver struct{}

func (canceledContextResolver) Resolve(ctx context.Context, _ AssigneeRequest) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []string{"user-1"}, nil
}

func TestEngineStartPassesCanceledContextToAssigneeResolver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	engine := NewEngine(canceledContextResolver{}, new(sequenceIDs))

	state, err := engine.Start(ctx, linearDefinition(workflowcore.ApprovalModeSingle, 0), StartRequest{StarterID: "starter"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() error = %v, want context.Canceled", err)
	}
	if state != nil {
		t.Fatalf("Start() state = %#v, want nil", state)
	}
}
