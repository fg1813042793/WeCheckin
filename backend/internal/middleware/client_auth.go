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

		expire, prefix := tokenutil.GetTokenConfig("user")

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
			ID         uint   `json:"id"`
			Name       string `json:"name"`
			Mobile     string `json:"mobile"`
			MiniOpenID string `json:"miniOpenID"`
			Role       string `json:"role"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &info); err != nil || info.ID == 0 {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录信息异常",
			})
			c.Abort()
			return
		}

		rd.RDB.Expire(rd.Ctx, prefix+"a:"+token, expire)

		user := &model.User{
			ID:         info.ID,
			Name:       info.Name,
			Mobile:     info.Mobile,
			MiniOpenID: info.MiniOpenID,
			Role:       info.Role,
		}
		c.Set("user_openid", info.MiniOpenID)
		c.Set("user_id", info.ID)
		c.Set("user", user)
		c.Next(ctx)
	}
}
