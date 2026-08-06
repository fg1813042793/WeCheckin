package dingtalkh5

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
)

func DingTalkH5Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := strings.TrimSpace(string(c.Request.Header.Peek("Authorization")))
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		user, err := dingtalkh5service.AuthenticateContext(ctx, token)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": err.Error()})
			c.Abort()
			return
		}
		dingtalkh5session.SetAuth(c, user, token)
		c.Next(ctx)
	}
}
