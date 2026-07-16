package database

import (
	"context"
	"testing"
	"time"
)

func TestQueryContextAddsDefaultTimeout(t *testing.T) {
	ctx, cancel := QueryContext(context.Background())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("QueryContext should add a default deadline")
	}
	if time.Until(deadline) <= 0 || time.Until(deadline) > DefaultQueryTimeout {
		t.Fatalf("deadline should be within DefaultQueryTimeout, got %v", time.Until(deadline))
	}
}

func TestQueryContextKeepsParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	ctx, cancel := QueryContext(parent)
	defer cancel()
	cancelParent()
	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("child query context should be canceled with parent")
	}
}
