package swagger

import (
	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/pkg/response"
)

// AdminPasswordPolicyGet returns the validated password policy.
// @Tags API v2-后台系统设置
// @Summary 获取密码策略
// @Security ApiKeyAuth
// @Success 200 {object} response.Resp{data=accountsecurity.Policy}
// @Router /api/v2/admin/password-policy [get]
func AdminPasswordPolicyGet() {}

// AdminPasswordPolicyPut validates and saves the complete password policy.
// @Tags API v2-后台系统设置
// @Summary 保存密码策略
// @Security ApiKeyAuth
// @Param body body accountsecurity.Policy true "密码策略"
// @Success 200 {object} response.Resp{data=accountsecurity.Policy}
// @Router /api/v2/admin/password-policy [put]
func AdminPasswordPolicyPut() {}

var (
	_ = response.Resp{}
	_ = accountsecurity.Policy{}
)

type ClientPasswordChangeRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

// ClientPasswordChange changes the current client's password.
// @Tags API v2-客户端账号
// @Summary 修改当前用户密码
// @Security ApiKeyAuth
// @Param body body ClientPasswordChangeRequest true "密码信息"
// @Success 200 {object} response.Resp{data=accountsecurity.PasswordState}
// @Router /api/v2/me/password [patch]
func ClientPasswordChange() {}
