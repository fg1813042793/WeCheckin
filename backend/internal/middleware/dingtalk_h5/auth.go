package dingtalkh5

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"wecheckin/backend/internal/service/accountsecurity"
	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/internal/support/passwordsession"
)

func DingTalkH5Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := authorizationToken(c)
		user, err := dingtalkh5service.AuthenticateContext(ctx, token)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": err.Error()})
			c.Abort()
			return
		}
		passwordState, err := accountsecurity.EvaluatePasswordStateContext(ctx, user.PasswordChangedAt, user.PasswordMustChange == 1)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "服务异常"})
			c.Abort()
			return
		}
		if passwordState.PasswordChangeRequired && !passwordsession.Allowed(passwordsession.AudienceDingTalkH5, string(c.Method()), string(c.Path())) {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": passwordsession.ErrorMessage, "errorCode": passwordsession.ErrorCode})
			c.Abort()
			return
		}
		dingtalkh5session.SetAuth(c, user, token)
		c.Set("passwordState", passwordState)
		c.Next(ctx)
	}
}

func authorizationToken(c *app.RequestContext) string {
	token := strings.TrimSpace(string(c.Request.Header.Peek("Authorization")))
	if len(token) >= len("Bearer ") && strings.EqualFold(token[:len("Bearer ")], "Bearer ") {
		token = strings.TrimSpace(token[len("Bearer "):])
	}
	return token
}
