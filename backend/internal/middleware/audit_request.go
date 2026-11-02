package middleware

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/support/auditlog"
)

const maxAuditRequestBytes = 64 * 1024

func auditTargetIDs(input map[string]interface{}) (json.RawMessage, int) {
	var values []string
	rawIDs := input["ids"]
	if rawIDs == nil {
		rawIDs = input["userIds"]
	}
	switch raw := rawIDs.(type) {
	case string:
		if strings.TrimSpace(raw) != "" {
			values = strings.Split(raw, ",")
		}
	case []interface{}:
		for _, v := range raw {
			if id := auditString(map[string]interface{}{"id": v}, "id"); id != "" {
				values = append(values, id)
			}
		}
	}
	count := len(values)
	if len(values) > 1000 {
		values = values[:1000]
	}
	for i := range values {
		values[i] = auditlog.CleanText(values[i], 128)
	}
	if values == nil {
		values = []string{}
	}
	data, _ := json.Marshal(values)
	return data, count
}

func auditInput(c *app.RequestContext) map[string]interface{} {
	result := make(map[string]interface{})
	query, _ := url.ParseQuery(string(c.Request.URI().QueryString()))
	for key, values := range query {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	if len(c.Request.Body()) == 0 || len(c.Request.Body()) > maxAuditRequestBytes || strings.Contains(strings.ToLower(string(c.Request.Header.ContentType())), "multipart/") {
		return result
	}
	var body map[string]interface{}
	if strings.Contains(strings.ToLower(string(c.Request.Header.ContentType())), "json") {
		if json.Unmarshal(c.Request.Body(), &body) != nil {
			var rows []map[string]interface{}
			if json.Unmarshal(c.Request.Body(), &rows) == nil {
				ids := make([]interface{}, 0, len(rows))
				for _, row := range rows {
					if id := auditString(row, "idStr", "id"); id != "" {
						ids = append(ids, id)
					}
				}
				result["ids"] = ids
			}
		}
	} else {
		values, _ := url.ParseQuery(string(c.Request.Body()))
		body = make(map[string]interface{})
		for key, value := range values {
			if len(value) > 0 {
				body[key] = value[0]
			}
		}
	}
	for key, value := range body {
		result[key] = value
	}
	return result
}

func auditString(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		switch v := values[key].(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return v
			}
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
	}
	return ""
}

type auditResponse struct {
	Code *int            `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func auditResult(c *app.RequestContext, op auditlog.Operation) (outcome, code, reason string, data map[string]interface{}) {
	var envelope auditResponse
	status := c.Response.StatusCode()
	if status == 401 || status == 403 {
		return "blocked", "permission_denied", "请求未通过身份或权限检查", nil
	}
	if status == 429 {
		return "blocked", "rate_limited", "请求被限流", nil
	}
	if status >= 500 {
		return "failure", "system_error", "服务处理异常", nil
	}
	if status >= 400 {
		return "failure", "invalid_request", "请求参数或资源无效", nil
	}
	if c.Response.IsBodyStream() {
		if op.Action == "export" && strings.Contains(string(c.Response.Header.Peek("Content-Disposition")), "attachment") {
			return "success", "", "", nil
		}
		return "unknown", "result_unavailable", "流式响应，未解析业务结果", nil
	}
	if len(c.Response.Body()) > 4*1024*1024 {
		return "unknown", "result_unavailable", "业务响应超过审计结果解析上限", nil
	}
	if json.Unmarshal(c.Response.Body(), &envelope) != nil || envelope.Code == nil {
		if op.Action == "export" && status >= 200 && status < 300 && strings.Contains(string(c.Response.Header.Peek("Content-Disposition")), "attachment") {
			return "success", "", "", nil
		}
		return "unknown", "result_unavailable", "未返回可识别的业务结果", nil
	}
	if *envelope.Code != 0 {
		msg := envelope.Msg
		switch {
		case strings.Contains(msg, "审计日志只读"):
			return "blocked", "audit_read_only", "审计日志不允许删除或清空", nil
		case strings.Contains(msg, "验证码"):
			return "failure", "captcha_invalid", "验证码校验未通过", nil
		case strings.Contains(msg, "频繁") || strings.Contains(msg, "限流") || strings.Contains(msg, "冷却"):
			return "blocked", "rate_limited", "请求被限流", nil
		case strings.Contains(msg, "无权限") || strings.Contains(msg, "权限不足") || strings.Contains(msg, "不允许") || strings.HasPrefix(msg, "缺少权限-") || msg == "接口未配置权限规则，请联系管理员":
			return "blocked", "permission_denied", "请求未通过权限检查", nil
		case strings.Contains(msg, "未登录") || strings.Contains(msg, "登录已过期") || strings.Contains(msg, "请先登录"):
			return "blocked", "unauthenticated", "登录状态无效", nil
		case strings.Contains(msg, "停用") || strings.Contains(msg, "禁用"):
			return "blocked", "account_disabled", "账号或资源已停用", nil
		case op.Category == "login":
			return "failure", "authentication_failed", "身份验证失败", nil
		default:
			return "failure", "business_error", "操作未成功，可通过请求 ID 查询服务日志", nil
		}
	}
	_ = json.Unmarshal(envelope.Data, &data)
	if succeeded, ok := data["success"].(bool); op.Action == "test_send" && ok && !succeeded {
		return "failure", "notification_test_failed", "测试通知未发送成功", data
	}
	if op.Module == "scheduler" && op.Action == "cancel" && auditString(data, "status") == "running" {
		return "accepted", "", "取消请求已受理，等待执行节点确认", data
	}
	if failed, _ := data["failedCount"].(float64); failed > 0 {
		if sent, _ := data["sentCount"].(float64); sent > 0 {
			return "partial", "partial_delivery", "部分发送失败", data
		}
		return "failure", "delivery_failed", "发送失败", data
	}
	if op.Async {
		return "accepted", "", "请求已受理，执行结果见关联运行或投递记录", data
	}
	return "success", "", "", data
}

func auditOutcomeForReason(code, fallback string) string {
	switch code {
	case "permission_denied", "admin_access_denied", "account_disabled", "unauthenticated", "rate_limited":
		return "blocked"
	default:
		return fallback
	}
}
