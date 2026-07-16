package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const DefaultQueryTimeout = 5 * time.Second

func QueryContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, DefaultQueryTimeout)
}

func WithContext(parent context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := QueryContext(parent)
	if DB == nil {
		return DB, cancel
	}
	return DB.WithContext(ctx), cancel
}
