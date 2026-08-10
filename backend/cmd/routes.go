package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"wecheckin/backend/internal/routes"
)

func registerRoutes(h *server.Hertz) {
	routes.Register(h)
}
