package middleware

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/jwtutil"
	"wecheckin-backend/backend/pkg/tokenutil"
)

func ClientAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tokenStr := string(c.Request.Header.Peek("Authorization"))
		if tokenStr == "" {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "未登录",
			})
			c.Abort()
			return
		}

		claims, err := jwtutil.ParseToken(tokenStr)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录已过期",
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

		// Find user by MiniOpenID from JWT claims
		var user model.User
		if err := database.GetDB().Where("`user_mini_openid` = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "用户不存在",
			})
			c.Abort()
			return
		}

		// Check Redis for online status (force-offline support)
		expire, prefix := tokenutil.GetTokenConfig("user")
		storedToken, err := rd.RDB.Get(rd.Ctx, prefix+strconv.Itoa(int(user.ID))).Result()
		if err != nil || storedToken != tokenStr {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录已过期或已被强制下线",
			})
			c.Abort()
			return
		}

		// Slide TTL
		rd.RDB.Expire(rd.Ctx, prefix+strconv.Itoa(int(user.ID)), expire)

		c.Set("user_openid", claims.UserID)
		c.Set("user_id", user.ID)
		c.Set("user", &user)
		c.Next(ctx)
	}
}
