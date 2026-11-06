package client

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/internal/support/passwordsession"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

func ClientAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := string(c.Request.Header.Peek("Authorization"))
		if token == "" {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "未登录",
			})
			c.Abort()
			return
		}

		if rd.RDB == nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "服务异常",
			})
			c.Abort()
			return
		}

		expire, prefix := tokenutil.GetTokenConfigContext(ctx, "user")
		redisCtx, cancel := rd.OperationContext(ctx)
		defer cancel()

		jsonStr, err := rd.RDB.Get(redisCtx, prefix+"a:"+token).Result()
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录已过期或已被强制下线",
			})
			c.Abort()
			return
		}

		var info struct {
			ID                 uint   `json:"id"`
			Name               string `json:"name"`
			Mobile             string `json:"mobile"`
			MiniOpenID         string `json:"miniOpenID"`
			Role               string `json:"role"`
			RoleID             uint   `json:"roleId"`
			RoleIDs            []uint `json:"roleIds"`
			PasswordChangedAt  int64  `json:"passwordChangedAt"`
			PasswordMustChange bool   `json:"passwordMustChange"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &info); err != nil || info.ID == 0 {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录信息异常",
			})
			c.Abort()
			return
		}
		passwordState, err := accountsecurity.EvaluatePasswordStateContext(ctx, info.PasswordChangedAt, info.PasswordMustChange)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": "服务异常"})
			c.Abort()
			return
		}
		if passwordState.PasswordChangeRequired && !passwordsession.Allowed(passwordsession.AudienceClient, string(c.Method()), string(c.Path())) {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": passwordsession.ErrorMessage, "errorCode": passwordsession.ErrorCode})
			c.Abort()
			return
		}

		if tokenutil.IsTokenSlidingExpirationEnabledContext(ctx, "user") {
			rd.RDB.Expire(redisCtx, prefix+"a:"+token, expire)
		}

		user := &model.User{
			ID:         info.ID,
			Name:       info.Name,
			Mobile:     info.Mobile,
			MiniOpenID: info.MiniOpenID,
			Role:       info.Role,
			RoleID:     info.RoleID,
			RoleIDs:    info.RoleIDs,
		}
		if len(user.RoleIDs) == 0 && user.RoleID > 0 {
			user.RoleIDs = []uint{user.RoleID}
		}
		c.Set("user_openid", info.MiniOpenID)
		c.Set("user_id", info.ID)
		c.Set("user", user)
		c.Set("passwordState", passwordState)
		c.Next(ctx)
	}
}
