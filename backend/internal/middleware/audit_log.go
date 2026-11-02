package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/internal/support/useragent"
	"wecheckin/backend/pkg/randutil"
)

func AuditLog() app.HandlerFunc {
	return auditLogWithWriter(auditlog.DatabaseWriter{}, auditlog.WriteFallback)
}

func auditLogWithWriter(writer auditlog.Writer, fallback func(*model.Log, string)) app.HandlerFunc {
	clientIP := ClientIPResolver()
	return func(ctx context.Context, c *app.RequestContext) {
		started := time.Now()
		method, path := string(c.Method()), string(c.Path())
		op, tracked := auditlog.Classify(method, path)
		requestID := "req-" + randutil.HexString(32)
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Response.Header.Set("X-Request-ID", requestID)
		auditCtx, recorder := auditlog.WithRecorder(ctx)
		event := &model.Log{EventID: "audit-" + randutil.HexString(32), RequestID: requestID,
			Method: auditlog.CleanText(method, 10), Path: auditlog.CleanText(path, 500), AddTime: started.UnixMilli(),
			AddIP: auditlog.CleanText(clientIP(c), 50), UserAgent: auditlog.CleanText(string(c.UserAgent()), 512),
			Outcome: "unknown", ActorType: "anonymous", Changes: auditlog.SafeChanges(nil), SnapshotStatus: auditlog.SnapshotPending}
		event.Browser, event.DeviceType, event.PlatformType = useragent.Parse(string(c.UserAgent()))
		auditlog.Update(auditCtx, func(d *auditlog.Details) { d.EventID = event.EventID })
		if tracked {
			event.Category, event.Source, event.Module, event.Action, event.TargetType = op.Category, op.Source, op.Module, op.Action, op.TargetType
			input := auditInput(c)
			event.TargetIDs, event.RequestedCount = auditTargetIDs(input)
			if op.Category == "login" {
				event.AttemptedAccount = auditlog.CleanText(auditString(input, "name", "account", "mobile"), 100)
			}
			event.TargetID = auditlog.CleanText(auditString(input, "id", "surveyId", "examId", "userId", "roleId", "taskId", "typeCode", "key"), 128)
			if op.Category != "login" {
				event.TargetName = auditlog.CleanText(auditString(input, "name", "title", "typeName"), 200)
			}
			setAuditActor(c, event)
			// Persist an unknown-outcome attempt before dispatch so a crash cannot invent success.
			event.EditTime = event.AddTime
			if err := writer.Write(auditCtx, event); err != nil {
				fallback(event, "begin")
			}
		}

		completed := false
		defer func() {
			outcome, code, reason, data := auditResult(c, op)
			if !tracked {
				if outcome != "blocked" || auditlog.Source(path) == "" {
					return
				}
				event.Category, event.Source, event.Module, event.Action = "login", auditlog.Source(path), "auth", "access_denied"
			}
			if !completed {
				outcome, code, reason = "failure", "system_error", "请求处理未正常完成"
			}
			event.Outcome, event.ReasonCode, event.Reason = outcome, code, reason
			event.EditTime = time.Now().UnixMilli()
			event.DurationMs = time.Since(started).Milliseconds()
			setAuditActor(c, event)
			for _, key := range []string{"recordId", "joinId", "participantId", "dynamicId", "userId", "scoreId", "id", "key", "typeCode", "definitionId"} {
				if value := c.Param(key); value != "" {
					event.TargetID = auditlog.CleanText(value, 128)
					break
				}
			}
			if event.TargetID == "" {
				event.TargetID = auditlog.CleanText(auditString(data, "id", "instanceId", "runId"), 128)
			}
			for _, key := range []string{"deletedCount", "deleted", "sentCount", "count"} {
				if value, ok := data[key].(float64); ok && value >= 0 {
					event.AffectedCount = int(value)
					break
				}
			}
			if op.Action == "export" {
				if rows, ok := data["list"].([]interface{}); ok {
					event.AffectedCount = len(rows)
					ids := make([]interface{}, 0, len(rows))
					for _, row := range rows {
						if item, ok := row.(map[string]interface{}); ok {
							if id := auditString(item, "id"); id != "" {
								ids = append(ids, id)
							}
						}
					}
					event.TargetIDs, _ = auditTargetIDs(map[string]interface{}{"ids": ids})
				}
			}
			event.RelatedType, event.RelatedID = op.RelatedType, event.TargetID
			if id := auditString(data, "instanceId"); id != "" {
				event.RelatedType, event.RelatedID = "workflow_instance", auditlog.CleanText(id, 128)
			}
			if run, ok := data["run"].(map[string]interface{}); ok {
				event.RelatedType, event.RelatedID = "scheduled_task_run", auditlog.CleanText(auditString(run, "id"), 128)
			}
			if op.Module == "notification" {
				if id := auditString(data, "requestId"); id != "" {
					event.RelatedType, event.RelatedID = "notification_request", auditlog.CleanText(id, 128)
				}
			}
			applyAuditDetails(event, recorder.Snapshot())
			if event.Outcome == "blocked" && event.AdminID == 0 && event.Category == "operation" {
				event.Category, event.Action = "login", "access_denied"
			}
			if event.RelatedType == "" {
				event.RelatedID = ""
			}
			event.Type = auditLegacyType(event.Category, event.Action)
			finalizeAuditSnapshot(event)
			event.Content = auditContent(event.Module, event.Action)
			if err := writer.Write(auditCtx, event); err != nil {
				fallback(event, "finish")
			}
		}()
		c.Next(auditCtx)
		completed = true
	}
}

func setAuditActor(c *app.RequestContext, event *model.Log) {
	if value, ok := c.Get("admin"); ok {
		if actor, ok := value.(*model.Admin); ok && actor != nil && actor.ID > 0 {
			event.AdminID, event.AdminName, event.AdminDesc, event.ActorType = actor.ID, auditlog.CleanText(actor.Name, 100), auditlog.CleanText(actor.Desc, 200), "admin"
			return
		}
	}
	if actor, ok := dingtalkh5session.CurrentUser(c); ok && actor != nil && actor.ID > 0 {
		event.AdminID, event.AdminName, event.ActorType = actor.ID, auditlog.CleanText(actor.Name, 100), "user"
		return
	}
	if value, ok := c.Get("user"); ok {
		if actor, ok := value.(*model.User); ok && actor != nil && actor.ID > 0 {
			event.AdminID, event.AdminName, event.ActorType = actor.ID, auditlog.CleanText(actor.Name, 100), "user"
			return
		}
	}
	if value, ok := c.Get("user_id"); ok {
		switch id := value.(type) {
		case uint:
			event.AdminID = id
		case int64:
			if id > 0 {
				event.AdminID = uint(id)
			}
		case float64:
			if id > 0 {
				event.AdminID = uint(id)
			}
		}
		if event.AdminID > 0 {
			event.ActorType = "user"
		}
	}
}

func applyAuditDetails(event *model.Log, details auditlog.Details) {
	if auditlog.ValidSnapshotStatus(details.SnapshotStatus) {
		event.SnapshotStatus = details.SnapshotStatus
		event.SnapshotReason = auditlog.CleanText(details.SnapshotReason, 80)
	}
	if details.ActorID > 0 {
		event.AdminID, event.AdminName, event.ActorType = details.ActorID, auditlog.CleanText(details.ActorName, 100), details.ActorType
	}
	if details.ReasonCode != "" {
		event.ReasonCode, event.Reason = auditlog.CleanText(details.ReasonCode, 64), auditlog.CleanText(details.Reason, 500)
		event.Outcome = auditOutcomeForReason(details.ReasonCode, event.Outcome)
	}
	if details.TargetType != "" {
		event.TargetType = auditlog.CleanText(details.TargetType, 64)
	}
	if details.TargetID != "" {
		event.TargetID = auditlog.CleanText(details.TargetID, 128)
	}
	if len(details.TargetIDs) > 0 {
		ids := make([]interface{}, len(details.TargetIDs))
		for i, id := range details.TargetIDs {
			ids[i] = id
		}
		event.TargetIDs, _ = auditTargetIDs(map[string]interface{}{"ids": ids})
	}
	if details.TargetName != "" {
		event.TargetName = auditlog.CleanText(details.TargetName, 200)
	}
	if details.AffectedCount > 0 {
		event.AffectedCount = details.AffectedCount
	}
	if details.RelatedType != "" {
		event.RelatedType, event.RelatedID = auditlog.CleanText(details.RelatedType, 64), auditlog.CleanText(details.RelatedID, 128)
	}
	if event.Outcome == "success" || event.Outcome == "accepted" || event.Outcome == "partial" {
		if changes, err := auditlog.EncodeChanges(details.Changes); err == nil {
			event.Changes = changes
		} else {
			event.Changes = auditlog.SafeChanges(nil)
			event.SnapshotStatus, event.SnapshotReason = auditlog.SnapshotUnavailable, "encoding_limit"
		}
	}
	if details.Outcome != "" {
		event.Outcome = details.Outcome
	}
	event.UpdateBy = event.AdminID
}

func auditLegacyType(category, action string) int {
	if category == "login" {
		return 1
	}
	switch action {
	case "create":
		return 2
	case "delete":
		return 3
	case "update", "status", "password_change":
		return 4
	default:
		return 5
	}
}

func auditContent(module, action string) string {
	modules := map[string]string{"auth": "账号安全", "user": "用户", "role": "角色", "permission": "权限", "organization": "组织", "dictionary": "字典", "survey": "问卷", "exam": "考试", "workflow": "流程", "scheduler": "定时任务", "notification": "通知", "settings": "系统配置", "feedback": "用户反馈", "content": "内容管理", "audit": "审计日志", "file": "文件", "performance": "绩效"}
	actions := map[string]string{"password_login": "密码登录", "sso_login": "钉钉免登", "bind": "账号绑定", "logout": "退出登录", "create": "创建", "update": "修改", "delete": "删除", "status": "变更状态", "publish": "发布", "rollback": "回滚", "copy": "复制", "run": "手动执行", "retry": "重试", "cancel": "取消", "withdraw": "撤回", "complete": "处理", "resume": "恢复", "export": "导出", "import": "导入", "password_change": "密码变更", "force_offline": "强制下线", "send": "发送", "test_send": "测试发送", "access_denied": "访问被拦截"}
	label := modules[module]
	if label == "" {
		label = module
	}
	verb := actions[action]
	if verb == "" {
		verb = action
	}
	return label + " / " + verb
}
