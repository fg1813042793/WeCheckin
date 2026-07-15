package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
)

func registerSwaggerRoutes(h *server.Hertz) {
	url := swagger.URL("/swagger/doc.json")
	h.GET("/swagger", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/swagger/index.html"))
	})
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))
}
