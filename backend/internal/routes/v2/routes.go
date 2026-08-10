package v2

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	adminroutes "wecheckin/backend/internal/routes/v2/admin"
	clientroutes "wecheckin/backend/internal/routes/v2/client"
	dingtalkh5routes "wecheckin/backend/internal/routes/v2/dingtalkh5"
)

func Register(h *server.Hertz) {
	clientroutes.Register(h)
	adminroutes.Register(h)
	dingtalkh5routes.Register(h)
}
