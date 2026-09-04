package dingtalk

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	admindingtalkservice "wecheckin/backend/internal/service/admin/dingtalk"
	"wecheckin/backend/pkg/response"
)

func (h *AdminDingTalkHandler) GetPerfReviews(ctx context.Context, c *app.RequestContext) {
	data, err := h.service.ListPerfReviews(ctx, admindingtalkservice.PerfReviewQuery{
		Page: parsePositiveInt(c.Query("page"), 1), PageSize: parsePositiveInt(c.Query("pageSize"), 20),
		Keyword: c.Query("keyword"), Employee: c.Query("employee"), Period: c.Query("period"), Status: c.Query("status"),
	})
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminDingTalkHandler) GetPerfReviewDetail(ctx context.Context, c *app.RequestContext) {
	id := parseUint(c.Query("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	data, err := h.service.GetPerfReviewDetail(ctx, id)
	if err != nil {
		if errors.Is(err, admindingtalkservice.ErrPerfReviewNotFound) {
			response.Fail(c, "考评单不存在")
			return
		}
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminDingTalkHandler) DeletePerfReview(ctx context.Context, c *app.RequestContext) {
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	adminID, adminName := currentAdminIdentity(c)
	err := h.service.DeletePerfReview(ctx, id, admindingtalkservice.AdminIdentity{ID: adminID, Name: adminName})
	respondPerfMutation(ctx, c, err)
}

func (h *AdminDingTalkHandler) DeletePerfReviews(ctx context.Context, c *app.RequestContext) {
	ids := parseUintList(c.PostForm("ids"))
	if len(ids) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	adminID, adminName := currentAdminIdentity(c)
	err := h.service.DeletePerfReviews(ctx, ids, admindingtalkservice.AdminIdentity{ID: adminID, Name: adminName})
	respondPerfMutation(ctx, c, err)
}

func (h *AdminDingTalkHandler) GetPerfHistories(ctx context.Context, c *app.RequestContext) {
	data, err := h.service.ListPerfHistories(ctx, admindingtalkservice.PerfHistoryQuery{
		Page: parsePositiveInt(c.Query("page"), 1), PageSize: parsePositiveInt(c.Query("pageSize"), 20),
		Keyword: c.Query("keyword"), ReviewNo: c.Query("reviewNo"), ByAccount: c.Query("byAccount"), Action: c.Query("action"),
	})
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminDingTalkHandler) DeletePerfHistory(ctx context.Context, c *app.RequestContext) {
	id := parseUint(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.service.DeletePerfHistories(ctx, []uint{id}); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) DeletePerfHistories(ctx context.Context, c *app.RequestContext) {
	ids := parseUintList(c.PostForm("ids"))
	if len(ids) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.service.DeletePerfHistories(ctx, ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func respondPerfMutation(ctx context.Context, c *app.RequestContext, err error) {
	if err != nil {
		response.FailInternal(ctx, c, "admin.dingtalk.performance", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

func parseUintList(raw string) []uint {
	parts := strings.Split(raw, ",")
	ids := make([]uint, 0, len(parts))
	seen := make(map[uint]struct{}, len(parts))
	for _, part := range parts {
		id := parseUint(part)
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseUint(value string) uint {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

func currentAdminIdentity(c *app.RequestContext) (uint, string) {
	adminValue, ok := c.Get("admin")
	if !ok || adminValue == nil {
		return 0, "管理员"
	}
	admin, ok := adminValue.(*model.Admin)
	if !ok || admin == nil {
		return 0, "管理员"
	}
	name := strings.TrimSpace(admin.Name)
	if name == "" {
		name = fmt.Sprintf("管理员%d", admin.ID)
	}
	return admin.ID, name
}
