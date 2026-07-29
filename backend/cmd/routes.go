package main

import "github.com/cloudwego/hertz/pkg/app/server"

func registerRoutes(h *server.Hertz) {
	registerHealthRoutes(h)
	registerSwaggerRoutes(h)
	registerPublicRoutes(h)
	registerClientRoutes(h)
	registerV2Routes(h)
	registerUploadAndStaticRoutes(h)
}
