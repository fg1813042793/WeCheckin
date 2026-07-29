package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/pkg/logger"
)

const maxAccessLogBodyBytes = 4 * 1024
const defaultSlowRequestThreshold = 800 * time.Millisecond

func safeParam(v []byte) string {
	param := maskSensitiveParam(string(v))
	if len(param) > 500 {
		return param[:500] + "..."
	}
	return param
}

func maskSensitiveParam(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}
	if strings.HasPrefix(trimmed, "{") {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &data); err == nil {
			masked, err := json.Marshal(maskSensitiveJSON(data))
			if err == nil {
				return string(masked)
			}
		}
	}
	if values, err := url.ParseQuery(raw); err == nil && len(values) > 0 {
		return encodeMaskedQuery(values)
	}
	return raw
}

func encodeMaskedQuery(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		for _, value := range values[key] {
			if isSensitiveLogKey(key) {
				value = "***"
			} else {
				value = url.QueryEscape(value)
			}
			parts = append(parts, url.QueryEscape(key)+"="+value)
		}
	}
	return strings.Join(parts, "&")
}

func maskSensitiveJSON(data map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{}, len(data))
	for key, value := range data {
		if isSensitiveLogKey(key) {
			masked[key] = "***"
			continue
		}
		switch typed := value.(type) {
		case map[string]interface{}:
			masked[key] = maskSensitiveJSON(typed)
		case []interface{}:
			masked[key] = maskSensitiveJSONArray(typed)
		default:
			masked[key] = value
		}
	}
	return masked
}

func maskSensitiveJSONArray(values []interface{}) []interface{} {
	masked := make([]interface{}, len(values))
	for i, value := range values {
		if item, ok := value.(map[string]interface{}); ok {
			masked[i] = maskSensitiveJSON(item)
		} else {
			masked[i] = value
		}
	}
	return masked
}

func isSensitiveLogKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{
		"password", "pwd", "token", "authorization", "auth", "mobile", "phone",
		"content", "answer", "answers", "schema", "description", "richtext", "html", "remark",
	} {
		if normalized == token || strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func accessLogDetail(method, rawQuery string, body []byte, contentType string, contentLength int) string {
	switch strings.ToUpper(method) {
	case "GET":
		if rawQuery == "" {
			return ""
		}
		return " ?" + safeParam([]byte(rawQuery))
	case "POST", "PUT", "PATCH", "DELETE":
		if skipReason := skipAccessLogBodyReason(contentType, contentLength); skipReason != "" {
			return " | body:<skipped:" + skipReason + ">"
		}
		if len(body) == 0 {
			return ""
		}
		if len(body) > maxAccessLogBodyBytes {
			return " | body:<skipped:large>"
		}
		return " | body:" + safeParam(body)
	default:
		return ""
	}
}

func skipAccessLogBodyReason(contentType string, contentLength int) string {
	normalized := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if strings.HasPrefix(normalized, "multipart/") || normalized == "application/octet-stream" {
		return "multipart"
	}
	if contentLength < 0 || contentLength > maxAccessLogBodyBytes {
		return "large"
	}
	return ""
}

func slowRequestThreshold() time.Duration {
	raw := strings.TrimSpace(os.Getenv("WECHECKIN_SLOW_REQUEST_MS"))
	if raw == "" {
		return defaultSlowRequestThreshold
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		return defaultSlowRequestThreshold
	}
	return time.Duration(ms) * time.Millisecond
}

func shouldLogSlowRequest(latency, threshold time.Duration) bool {
	return threshold > 0 && latency >= threshold
}

func slowRequestLogLine(now time.Time, method, path string, statusCode int, latency time.Duration, requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		requestID = "-"
	}
	return fmt.Sprintf("[SLOW_REQUEST] %s | %3d | %13v | requestId=%s | %s %s",
		now.Format("2006/01/02 15:04:05"),
		statusCode,
		latency,
		requestID,
		method,
		path,
	)
}

func accessLogRequestID(c *app.RequestContext) string {
	for _, header := range []string{"X-Request-ID", "X-Trace-ID"} {
		if value := strings.TrimSpace(string(c.GetHeader(header))); value != "" {
			return value
		}
	}
	return "-"
}

func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Path())
		method := string(c.Method())
		requestID := accessLogRequestID(c)

		rawQuery := string(c.Request.URI().QueryString())
		contentType := string(c.Request.Header.ContentType())
		contentLength := c.Request.Header.ContentLength()
		var body []byte
		if skipAccessLogBodyReason(contentType, contentLength) == "" {
			body, _ = c.Body()
		}

		c.Next(ctx)

		latency := time.Since(start)
		statusCode := c.Response.StatusCode()
		clientIP := c.ClientIP()
		detail := accessLogDetail(method, rawQuery, body, contentType, contentLength)

		logger.Logger.Printf("[ACCESS] %s | %3d | %13v | %15s | %s %s%s",
			time.Now().Format("2006/01/02 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
			detail,
		)
		if shouldLogSlowRequest(latency, slowRequestThreshold()) {
			logger.Logger.Print(slowRequestLogLine(time.Now(), method, path, statusCode, latency, requestID))
		}
	}
}
