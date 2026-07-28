package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"wecheckin-backend/backend/pkg/database"
)

func registerHealthRoutes(h *server.Hertz) {
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{"status": "ok"})
	})
	h.GET("/ready", func(ctx context.Context, c *app.RequestContext) {
		db, cancel := database.WithContext(ctx)
		defer cancel()
		if db == nil {
			c.JSON(consts.StatusServiceUnavailable, utils.H{"status": "unavailable", "message": "database not initialized"})
			return
		}
		if err := db.Exec("SELECT 1").Error; err != nil {
			c.JSON(consts.StatusServiceUnavailable, utils.H{"status": "unavailable", "message": err.Error()})
			return
		}
		c.JSON(consts.StatusOK, utils.H{"status": "ready"})
	})
}
