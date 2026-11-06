package mgr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	adminlogservice "wecheckin/backend/internal/service/admin/adminlog"
	"wecheckin/backend/pkg/response"
)

func (h *AdminMgrHandler) DeleteLogs(ctx context.Context, c *app.RequestContext) {
	h.deleteAuditLogs(ctx, c, adminlogservice.CategoryOperation)
}

func (h *AdminMgrHandler) DeleteLoginLogs(ctx context.Context, c *app.RequestContext) {
	h.deleteAuditLogs(ctx, c, adminlogservice.CategoryLogin)
}

func (h *AdminMgrHandler) deleteAuditLogs(ctx context.Context, c *app.RequestContext, category string) {
	actorID, ok := auditReaderID(c)
	if !ok {
		return
	}
	var request adminlogservice.DeleteRequest
	if raw := c.Param("id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, strconv.IntSize)
		if err != nil || id == 0 {
			response.Fail(c, "id: 必须为正整数")
			return
		}
		if len(c.Request.Body()) != 0 {
			response.Fail(c, "单条删除仅接受路径中的日志 ID")
			return
		}
		request.IDs = []uint{uint(id)}
	} else {
		body := c.Request.Body()
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if len(body) > 4096 || !json.Valid(body) || decoder.Decode(&request) != nil {
			response.Fail(c, "ids: 请提交包含日志 ID 数组的 JSON 对象")
			return
		}
	}
	result, err := adminlogservice.DeleteForAdminContext(ctx, category, request.IDs, actorID)
	if err != nil {
		var invalid *adminlogservice.ValidationError
		if errors.As(err, &invalid) || errors.Is(err, adminlogservice.ErrForbidden) || errors.Is(err, adminlogservice.ErrNotFound) {
			auditLogError(ctx, c, err)
		} else {
			response.FailInternal(ctx, c, "admin.auditlog.delete", "删除日志失败，请稍后重试", err)
		}
		return
	}
	response.JSON(c, result)
}
