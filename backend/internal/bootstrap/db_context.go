package bootstrap

import (
	"context"
	"time"

	"gorm.io/gorm"
	"wecheckin-backend/backend/pkg/database"
)

const startupDatabaseTimeout = 2 * time.Minute

func startupDB(parent context.Context) (*gorm.DB, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	startupCtx, startupCancel := context.WithTimeout(parent, startupDatabaseTimeout)
	db, queryCancel := database.WithContext(startupCtx)
	return db, func() {
		queryCancel()
		startupCancel()
	}
}
