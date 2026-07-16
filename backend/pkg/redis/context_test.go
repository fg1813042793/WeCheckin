package redis

import (
	"context"
	"testing"
	"time"
)

func TestOperationContextAddsDefaultTimeout(t *testing.T) {
	ctx, cancel := OperationContext(context.Background())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("OperationContext should add a default deadline")
	}
	if time.Until(deadline) <= 0 || time.Until(deadline) > DefaultOperationTimeout {
		t.Fatalf("deadline should be within DefaultOperationTimeout, got %v", time.Until(deadline))
	}
}
