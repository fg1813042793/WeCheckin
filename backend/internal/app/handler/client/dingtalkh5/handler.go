package dingtalkh5

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	dingtalkh5service "wecheckin-backend/backend/internal/app/service/dingtalkh5"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := strings.TrimSpace(string(c.Request.Header.Peek("Authorization")))
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		user, err := dingtalkh5service.AuthenticateContext(ctx, token)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{"code": 1, "msg": err.Error()})
			c.Abort()
			return
		}
		c.Set("dingtalk_h5_user", user)
		c.Set("dingtalk_h5_token", token)
		c.Next(ctx)
	}
}

func currentUser(c *app.RequestContext) (*model.DingTalkH5PerfUser, bool) {
	value, ok := c.Get("dingtalk_h5_user")
	if !ok {
		return nil, false
	}
	user, ok := value.(*model.DingTalkH5PerfUser)
	return user, ok
}

func currentToken(c *app.RequestContext) string {
	value, _ := c.Get("dingtalk_h5_token")
	token, _ := value.(string)
	return token
}

func (h *Handler) Login(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		AuthCode string `json:"authCode"`
	}
	_ = c.BindAndValidate(&req)
	if req.Name == "" {
		req.Name = c.PostForm("name")
	}
	if req.Password == "" {
		req.Password = c.PostForm("password")
	}
	data, err := dingtalkh5service.LoginContext(ctx, req.Name, req.Password, c.ClientIP(), string(c.UserAgent()))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) Logout(ctx context.Context, c *app.RequestContext) {
	if err := dingtalkh5service.LogoutContext(ctx, currentToken(c)); err != nil {
		response.Fail(c, "退出失败")
		return
	}
	response.JSON(c, nil)
}

func (h *Handler) Bootstrap(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.BootstrapContext(ctx, user)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) Template(ctx context.Context, c *app.RequestContext) {
	data, err := dingtalkh5service.TemplateContext(ctx)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) Workbench(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.WorkbenchStatsContext(ctx, user)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if err := dingtalkh5service.ChangePasswordContext(ctx, user, req.CurrentPassword, req.NewPassword, req.ConfirmPassword); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *Handler) ListReviews(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.ListReviewsContext(ctx, user, filtersFromQuery(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) ReviewDetail(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.GetReviewContext(ctx, user, c.Param("id"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) CreateReview(ctx context.Context, c *app.RequestContext) {
	user, payload, ok := reviewRequest(c)
	if !ok {
		return
	}
	data, err := dingtalkh5service.CreateReviewContext(ctx, user, payload)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) SaveSelf(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.SaveSelfContext)
}

func (h *Handler) SubmitSelf(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.SubmitSelfContext)
}

func (h *Handler) SubmitManager(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.SubmitManagerContext)
}

func (h *Handler) SubmitHRBP(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.SubmitHRBPContext)
}

func (h *Handler) ConfirmResult(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.ConfirmResultContext)
}

func (h *Handler) DisputeResult(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.DisputeResultContext)
}

func (h *Handler) Finalize(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.FinalizeContext)
}

func (h *Handler) ReturnEmployee(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.ReturnEmployeeContext)
}

func (h *Handler) ReturnManager(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.ReturnManagerContext)
}

func (h *Handler) ReturnHRBP(ctx context.Context, c *app.RequestContext) {
	h.reviewAction(ctx, c, dingtalkh5service.ReturnHRBPContext)
}

func (h *Handler) Withdraw(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.WithdrawContext(ctx, user, c.Param("id"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) DeleteReview(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	if err := dingtalkh5service.DeleteReviewContext(ctx, user, c.Param("id")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *Handler) reviewAction(ctx context.Context, c *app.RequestContext, fn func(context.Context, *model.DingTalkH5PerfUser, string, dingtalkh5service.ReviewPayload) (*dingtalkh5service.ReviewDTO, error)) {
	user, payload, ok := reviewRequest(c)
	if !ok {
		return
	}
	data, err := fn(ctx, user, c.Param("id"), payload)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func reviewRequest(c *app.RequestContext) (*model.DingTalkH5PerfUser, dingtalkh5service.ReviewPayload, bool) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return nil, dingtalkh5service.ReviewPayload{}, false
	}
	var payload dingtalkh5service.ReviewPayload
	_ = c.BindAndValidate(&payload)
	return user, payload, true
}

func (h *Handler) ExportReviews(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.ExportReviewsContext(ctx, user, filtersFromQuery(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	c.Header("Content-Type", data.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+data.Filename)
	c.Write([]byte(data.Body))
}

func (h *Handler) ListUsers(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.ListUsersContext(ctx, user)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) CreateUser(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var payload dingtalkh5service.UserPayload
	_ = c.BindAndValidate(&payload)
	created, users, err := dingtalkh5service.CreateUserContext(ctx, user, payload)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, utils.H{"user": created, "users": users})
}

func (h *Handler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var payload dingtalkh5service.UserPayload
	_ = c.BindAndValidate(&payload)
	updated, users, err := dingtalkh5service.UpdateUserContext(ctx, user, c.Param("id"), payload)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, utils.H{"user": updated, "users": users})
}

func (h *Handler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	user, ok := currentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	users, err := dingtalkh5service.DeleteUserContext(ctx, user, c.Param("id"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, utils.H{"users": users})
}

func filtersFromQuery(c *app.RequestContext) dingtalkh5service.ReviewFilters {
	return dingtalkh5service.ReviewFilters{
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		Scope:      strings.TrimSpace(c.Query("scope")),
		Department: strings.TrimSpace(c.Query("department")),
		Period:     strings.TrimSpace(c.Query("period")),
		NextPeriod: strings.TrimSpace(c.Query("nextPeriod")),
		Status:     strings.TrimSpace(c.Query("status")),
		ManagerID:  strings.TrimSpace(c.Query("managerId")),
		HRBPID:     strings.TrimSpace(c.Query("hrbpId")),
		Grade:      strings.TrimSpace(c.Query("grade")),
		Page:       parsePositiveQueryInt(c, "page", 1),
		PageSize:   parsePositiveQueryInt(c, "pageSize", 20),
	}
}

func parsePositiveQueryInt(c *app.RequestContext, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(c.Query(key)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
