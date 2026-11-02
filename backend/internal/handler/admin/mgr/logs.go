package mgr

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	adminlogservice "wecheckin/backend/internal/service/admin/adminlog"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/response"
)

// GetLogList returns operations only; log:list never grants access to login attempts.
func (h *AdminMgrHandler) GetLogList(ctx context.Context, c *app.RequestContext) {
	h.getAuditLogPage(ctx, c, adminlogservice.CategoryOperation, false)
}

func (h *AdminMgrHandler) GetLoginLogList(ctx context.Context, c *app.RequestContext) {
	h.getAuditLogPage(ctx, c, adminlogservice.CategoryLogin, false)
}

func (h *AdminMgrHandler) ExportLogs(ctx context.Context, c *app.RequestContext) {
	h.getAuditLogPage(ctx, c, adminlogservice.CategoryOperation, true)
}

func (h *AdminMgrHandler) ExportLoginLogs(ctx context.Context, c *app.RequestContext) {
	h.getAuditLogPage(ctx, c, adminlogservice.CategoryLogin, true)
}

func (h *AdminMgrHandler) GetLogDetail(ctx context.Context, c *app.RequestContext) {
	h.getAuditLogDetail(ctx, c, adminlogservice.CategoryOperation)
}

func (h *AdminMgrHandler) GetLoginLogDetail(ctx context.Context, c *app.RequestContext) {
	h.getAuditLogDetail(ctx, c, adminlogservice.CategoryLogin)
}

func (h *AdminMgrHandler) getAuditLogPage(ctx context.Context, c *app.RequestContext, category string, export bool) {
	actorID, ok := auditReaderID(c)
	if !ok {
		return
	}
	values := make(url.Values)
	c.Request.URI().QueryArgs().VisitAll(func(key, value []byte) { values.Add(string(key), string(value)) })
	query, err := adminlogservice.ParseQuery(values, export)
	if err != nil {
		auditLogError(ctx, c, err)
		return
	}
	var page adminlogservice.Page
	if export {
		page, err = adminlogservice.ExportContext(ctx, category, query, actorID)
	} else {
		page, err = adminlogservice.ListContext(ctx, category, query, actorID)
	}
	if err != nil {
		auditLogError(ctx, c, err)
		return
	}
	if export {
		auditlog.Update(ctx, func(d *auditlog.Details) {
			d.AffectedCount = len(page.List)
			d.TargetType = "audit_log"
		})
	}
	response.JSON(c, page)
}

func (h *AdminMgrHandler) getAuditLogDetail(ctx context.Context, c *app.RequestContext, category string) {
	actorID, ok := auditReaderID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, strconv.IntSize)
	if err != nil || id == 0 {
		response.Fail(c, "id: 必须为正整数")
		return
	}
	row, err := adminlogservice.DetailContext(ctx, category, uint(id), actorID)
	if err != nil {
		auditLogError(ctx, c, err)
		return
	}
	response.JSON(c, row)
}

func auditReaderID(c *app.RequestContext) (uint, bool) {
	value, _ := c.Get("admin")
	admin, ok := value.(*model.Admin)
	if !ok || admin == nil || admin.ID == 0 {
		response.Fail(c, "无权限访问审计日志")
		return 0, false
	}
	return admin.ID, true
}

func auditLogError(ctx context.Context, c *app.RequestContext, err error) {
	var invalid *adminlogservice.ValidationError
	switch {
	case errors.As(err, &invalid):
		response.Fail(c, invalid.Field+": "+invalid.Message)
	case errors.Is(err, adminlogservice.ErrForbidden):
		response.Fail(c, "无权限访问审计日志")
	case errors.Is(err, adminlogservice.ErrNotFound):
		response.Fail(c, "日志不存在或无权访问")
	default:
		response.FailInternal(ctx, c, "admin.auditlog", "获取审计日志失败，请稍后重试", err)
	}
}

// ClearLog is a read-only compatibility response for the former DELETE endpoint.
func (h *AdminMgrHandler) ClearLog(_ context.Context, c *app.RequestContext) {
	response.Fail(c, "审计日志只读，不允许删除或清空")
}
