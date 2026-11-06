package user

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/accountsecurity"
	adminuserservice "wecheckin/backend/internal/service/admin/adminuser"
	"wecheckin/backend/pkg/response"
)

type ResetPasswordRequest struct {
	Password string `form:"password"`
}

func (h *AdminUserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	// Older clients omit password. An explicitly empty value is never a reset
	// to the legacy default, and passwords retain their surrounding whitespace.
	request := ResetPasswordRequest{Password: c.DefaultPostForm("password", "123456")}
	switch {
	case strings.TrimSpace(request.Password) == "":
		response.Fail(c, "请输入新密码")
		return
	case utf8.RuneCountInString(request.Password) < 6:
		response.Fail(c, "密码至少需要 6 位")
		return
	case len(request.Password) > 72:
		response.Fail(c, "密码不能超过 72 字节")
		return
	}
	if err := adminuserservice.ResetUserPasswordWithPasswordForAdminContext(ctx, id, request.Password, admin.ID); err != nil {
		if message := accountsecurity.ClientErrorMessage(err, ""); message != "" {
			response.Fail(c, message)
		} else if message := adminuserservice.UserWriteErrorMessage(err); message != "" {
			response.Fail(c, message)
		} else {
			response.FailInternal(ctx, c, "admin.user.reset_password", "密码重置失败，请稍后重试", err)
		}
		return
	}
	response.JSON(c, nil)
}
