package routes

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"wecheckin/backend/internal/routes/common"
	clientv1 "wecheckin/backend/internal/routes/v1/client"
	routesv2 "wecheckin/backend/internal/routes/v2"
)

func Register(h *server.Hertz) {
	common.RegisterHealth(h)
	common.RegisterSwagger(h)
	clientv1.Register(h)
	routesv2.Register(h)
	common.RegisterUploadAndStatic(h)
}
