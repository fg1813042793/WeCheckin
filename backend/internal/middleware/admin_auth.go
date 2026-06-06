package middleware

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"wecheckin-backend/backend/internal/model"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
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

		expire, prefix := tokenutil.GetTokenConfig("admin")

		jsonStr, err := rd.RDB.Get(rd.Ctx, prefix+"a:"+token).Result()
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录已过期或已被强制下线",
			})
			c.Abort()
			return
		}

		var info struct {
			ID       uint   `json:"id"`
			Name     string `json:"name"`
			Type     int    `json:"type"`
			RoleID   uint   `json:"roleId"`
			RoleName string `json:"roleName"`
			Desc     string `json:"desc"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &info); err != nil || info.ID == 0 {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录信息异常",
			})
			c.Abort()
			return
		}

		// Slide TTL: only the a: key needs sliding on every request.
		// The s: Set is refreshed when tokens are added/removed.
		rd.RDB.Expire(rd.Ctx, prefix+"a:"+token, expire)

		admin := &model.Admin{
			ID:     info.ID,
			Name:   info.Name,
			Type:   info.Type,
			RoleID: info.RoleID,
			Desc:   info.Desc,
		}
		c.Set("admin", admin)
		c.Next(ctx)
	}
}
