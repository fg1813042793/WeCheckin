package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/database"
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

		var admin model.Admin
		err := database.GetDB().Where("`ADMIN_TOKEN` = ? AND `ADMIN_STATUS` = 1", token).First(&admin).Error
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "登录已过期",
			})
			c.Abort()
			return
		}

		c.Set("admin", &admin)
		c.Next(ctx)
	}
}
