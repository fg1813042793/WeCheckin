package redis

import (
	"context"
	"time"
)

const DefaultOperationTimeout = 5 * time.Second

func OperationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, DefaultOperationTimeout)
}
