package admin

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

func AdminAuth() app.HandlerFunc {
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

		expire, prefix := tokenutil.GetTokenConfigContext(ctx, "admin")
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
			ID                 uint     `json:"id"`
			Name               string   `json:"name"`
			Type               int      `json:"type"`
			RoleID             uint     `json:"roleId"`
			RoleName           string   `json:"roleName"`
			RoleIDs            []uint   `json:"roleIds"`
			RoleNames          []string `json:"roleNames"`
			Desc               string   `json:"desc"`
			PasswordChangedAt  int64    `json:"passwordChangedAt"`
			PasswordMustChange bool     `json:"passwordMustChange"`
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
		if passwordState.PasswordChangeRequired && !passwordsession.Allowed(passwordsession.AudienceAdmin, string(c.Method()), string(c.Path())) {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": passwordsession.ErrorMessage, "errorCode": passwordsession.ErrorCode})
			c.Abort()
			return
		}

		// Only refresh the authentication key when sliding expiration is enabled.
		// The token index set is refreshed when tokens are added or removed.
		if tokenutil.IsTokenSlidingExpirationEnabledContext(ctx, "admin") {
			rd.RDB.Expire(redisCtx, prefix+"a:"+token, expire)
		}

		admin := &model.Admin{
			ID:        info.ID,
			Name:      info.Name,
			Type:      info.Type,
			RoleID:    info.RoleID,
			RoleIDs:   info.RoleIDs,
			RoleNames: info.RoleNames,
			Desc:      info.Desc,
		}
		if len(admin.RoleIDs) == 0 && admin.RoleID > 0 {
			admin.RoleIDs = []uint{admin.RoleID}
		}
		if len(admin.RoleNames) == 0 && info.RoleName != "" {
			admin.RoleNames = []string{info.RoleName}
		}
		c.Set("admin", admin)
		c.Set("passwordState", passwordState)
		c.Next(ctx)
	}
}
