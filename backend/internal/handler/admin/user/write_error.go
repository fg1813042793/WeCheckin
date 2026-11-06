package user

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	adminuserservice "wecheckin/backend/internal/service/admin/adminuser"
	"wecheckin/backend/pkg/response"
)

func respondUserWriteError(ctx context.Context, c *app.RequestContext, operation string, err error) {
	if message := adminuserservice.UserWriteErrorMessage(err); message != "" {
		response.Fail(c, message)
		return
	}
	response.FailInternal(ctx, c, operation, "操作失败，请稍后重试", err)
}
