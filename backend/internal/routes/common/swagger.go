package common

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
)

func RegisterSwagger(h *server.Hertz) {
	url := swagger.URL("/swagger/doc.json")
	swaggerHandler := swagger.WrapHandler(swaggerFiles.Handler, url)
	h.GET("/swagger", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/swagger/index.html"))
	})
	h.GET("/swagger/*any", func(ctx context.Context, c *app.RequestContext) {
		if string(c.Path()) == "/swagger/index.html" {
			serveSwaggerIndex(c)
			return
		}
		swaggerHandler(ctx, c)
	})
}
