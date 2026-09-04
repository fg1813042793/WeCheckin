package response

import (
	"context"
	"log"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/pkg/logger"
)

const defaultInternalErrorMessage = "服务异常，请稍后重试"

var internalErrorLogf = func(format string, args ...interface{}) {
	if logger.Logger != nil {
		logger.Logger.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}

func FailInternal(ctx context.Context, c *app.RequestContext, operation, publicMessage string, err error) {
	operation = strings.TrimSpace(operation)
	if operation == "" {
		operation = "unknown"
	}
	publicMessage = strings.TrimSpace(publicMessage)
	if publicMessage == "" {
		publicMessage = defaultInternalErrorMessage
	}
	var contextErr error
	if ctx != nil {
		contextErr = ctx.Err()
	}
	internalErrorLogf("[HTTP_INTERNAL_ERROR] operation=%s request_id=%s context_error=%v err=%+v", operation, requestID(c), contextErr, err)
	Fail(c, publicMessage)
}

func requestID(c *app.RequestContext) string {
	if c != nil {
		for _, header := range []string{"X-Request-ID", "X-Trace-ID"} {
			if value := strings.TrimSpace(string(c.GetHeader(header))); value != "" {
				return value
			}
		}
	}
	return "-"
}
