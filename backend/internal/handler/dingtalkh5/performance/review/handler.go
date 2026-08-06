package review

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) ListReviews(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
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
	user, ok := dingtalkh5session.CurrentUser(c)
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
	if len(payload.EmployeeIDs) > 0 {
		data, err := dingtalkh5service.CreateReviewsContext(ctx, user, payload)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		response.JSON(c, data)
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
	h.reviewAction(ctx, c, dingtalkh5service.WithdrawContext)
}

func (h *Handler) DeleteReview(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
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
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return nil, dingtalkh5service.ReviewPayload{}, false
	}
	var payload dingtalkh5service.ReviewPayload
	_ = c.BindAndValidate(&payload)
	return user, payload, true
}

func (h *Handler) ExportReviews(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
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
	c.Write(data.Body)
}

func filtersFromQuery(c *app.RequestContext) dingtalkh5service.ReviewFilters {
	objectiveScore := strings.TrimSpace(c.Query("objectiveScore"))
	if objectiveScore == "" {
		objectiveScore = strings.TrimSpace(c.Query("finalScore"))
	}
	return dingtalkh5service.ReviewFilters{
		Keyword:         strings.TrimSpace(c.Query("keyword")),
		Scope:           strings.TrimSpace(c.Query("scope")),
		EmployeeName:    strings.TrimSpace(c.Query("employeeName")),
		Department:      strings.TrimSpace(c.Query("department")),
		DepartmentName:  strings.TrimSpace(c.Query("departmentName")),
		DepartmentNames: splitQueryList(c.Query("departmentNames")),
		Period:          strings.TrimSpace(c.Query("period")),
		Year:            strings.TrimSpace(c.Query("year")),
		Month:           strings.TrimSpace(c.Query("month")),
		NotPeriod:       strings.TrimSpace(c.Query("notPeriod")),
		NextPeriod:      strings.TrimSpace(c.Query("nextPeriod")),
		Status:          strings.TrimSpace(c.Query("status")),
		Statuses:        splitQueryList(c.Query("statuses")),
		ManagerID:       strings.TrimSpace(c.Query("managerId")),
		HRBPID:          strings.TrimSpace(c.Query("hrbpId")),
		Grade:           strings.TrimSpace(c.Query("grade")),
		ObjectiveScore:  objectiveScore,
		Page:            parsePositiveQueryInt(c, "page", 1),
		PageSize:        parsePositiveQueryInt(c, "pageSize", 20),
		SkipHistory:     parseBoolQuery(c, "skipHistory", true) && !parseBoolQuery(c, "includeHistory", false),
	}
}

func splitQueryList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		items = append(items, item)
	}
	return items
}

func parsePositiveQueryInt(c *app.RequestContext, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(c.Query(key)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func parseBoolQuery(c *app.RequestContext, key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(c.Query(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
