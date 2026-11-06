package setup

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/pkg/response"
)

func (h *AdminSetupHandler) GetPasswordPolicy(ctx context.Context, c *app.RequestContext) {
	policy, err := accountsecurity.LoadPolicyContext(ctx)
	if err != nil {
		response.FailInternal(ctx, c, "admin.setup.password_policy", "获取密码策略失败，请稍后重试", err)
		return
	}
	response.JSON(c, policy)
}

func (h *AdminSetupHandler) SavePasswordPolicy(ctx context.Context, c *app.RequestContext) {
	var request accountsecurity.Policy
	if err := c.BindAndValidate(&request); err != nil {
		response.Fail(c, "请求参数格式无效")
		return
	}
	policy, err := accountsecurity.SavePolicyContext(ctx, request, c.ClientIP())
	if err != nil {
		response.Fail(c, accountsecurity.ClientErrorMessage(err, "保存密码策略失败，请稍后重试"))
		return
	}
	response.JSON(c, policy)
}
