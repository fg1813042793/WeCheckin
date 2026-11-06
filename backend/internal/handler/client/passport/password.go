package passport

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/pkg/response"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

// ChangePassword changes the current client's password and unlocks a restricted session.
// @Tags API v2-客户端账号
// @Summary 修改当前用户密码
// @Security ApiKeyAuth
// @Param body body ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Resp{data=accountsecurity.PasswordState}
// @Router /api/v2/me/password [patch]
func (h *PassportHandler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	userIDValue, ok := c.Get("user_id")
	userID, valid := userIDValue.(uint)
	if !ok || !valid || userID == 0 {
		response.Fail(c, "未登录")
		return
	}
	var request ChangePasswordRequest
	if err := c.BindAndValidate(&request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	if request.NewPassword != request.ConfirmPassword {
		response.Fail(c, "两次输入的新密码不一致")
		return
	}
	if err := accountsecurity.ChangeOwnPasswordContext(ctx, userID, request.CurrentPassword, request.NewPassword); err != nil {
		response.Fail(c, accountsecurity.ClientErrorMessage(err, "密码修改失败，请稍后重试"))
		return
	}
	response.JSON(c, accountsecurity.PasswordState{})
}
