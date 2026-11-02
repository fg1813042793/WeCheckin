package middleware

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/pkg/response"
)

// AuditExportPage retains the existing list data scope while bounding explicit export reads.
func AuditExportPage() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		for _, field := range []struct {
			key string
			max int
		}{{"pageSize", 1000}, {"page", 1000000}} {
			if raw := c.Query(field.key); raw != "" {
				n, err := strconv.Atoi(raw)
				if err != nil || n < 1 || n > field.max {
					response.Fail(c, "导出分页参数无效")
					c.Abort()
					return
				}
			}
		}
		c.Next(ctx)
	}
}
