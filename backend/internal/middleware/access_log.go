package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/pkg/logger"
)

func safeParam(v []byte) string {
	if len(v) > 500 {
		return string(v[:500]) + "..."
	}
	return string(v)
}

func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Path())
		method := string(c.Method())

		rawQuery := string(c.Request.URI().QueryString())
		body, _ := c.Body()

		c.Next(ctx)

		latency := time.Since(start)
		statusCode := c.Response.StatusCode()
		clientIP := c.ClientIP()

		detail := ""
		if method == "get" && rawQuery != "" {
			detail = " ?" + rawQuery
		} else if method == "post" && len(body) > 0 {
			detail = " | body:" + safeParam(body)
		}

		logger.Logger.Printf("[ACCESS] %s | %3d | %13v | %15s | %s %s%s",
			time.Now().Format("2006/01/02 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
			detail,
		)
	}
}
