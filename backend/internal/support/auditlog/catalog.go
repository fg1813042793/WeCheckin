package auditlog

import "strings"

type Operation struct {
	Category, Source, Module, Action, TargetType, RelatedType string
	Async                                                     bool
}

var resources = map[string][2]string{
	"users": {"user", "user"}, "managers": {"user", "user"},
	"user-sessions": {"auth", "session"}, "admin-sessions": {"auth", "session"},
	"roles": {"role", "role"}, "permissions": {"permission", "permission"},
	"departments": {"organization", "department"}, "positions": {"organization", "position"},
	"dict": {"dictionary", "dictionary"}, "settings": {"settings", "setting"},
	"surveys": {"survey", "survey"}, "survey-responses": {"survey", "survey_response"},
	"survey-question-bank": {"survey", "question"}, "survey-resources": {"survey", "resource"},
	"survey-channels": {"survey", "channel"}, "survey-template-presets": {"survey", "preset"},
	"survey-notification-templates": {"survey", "notification_template"},
	"exam-notification-templates":   {"exam", "notification_template"},
	"survey-report":                 {"survey", "report"}, "survey-notifications": {"notification", "notification"},
	"exams": {"exam", "exam"}, "exam-records": {"exam", "exam_record"},
	"exam-question-bank": {"exam", "question"}, "exam-resources": {"exam", "resource"},
	"workflow-definitions": {"workflow", "workflow_definition"}, "workflow-instances": {"workflow", "workflow_instance"},
	"workflow-tasks": {"workflow", "workflow_task"}, "workflow-form-revisions": {"workflow", "form_revision"},
	"workflow-org-approver-assignments": {"permission", "org_approver_assignment"},
	"workflow-org-approver-identities":  {"permission", "org_approver_identity"},
	"workflow-notifications":            {"notification", "workflow_notification"},
	"scheduled-tasks":                   {"scheduler", "scheduled_task"}, "scheduled-task-runs": {"scheduler", "scheduled_task_run"},
	"in-app-notifications": {"notification", "notification"}, "webhook-notifications": {"notification", "webhook_notification"},
	"notifications":          {"notification", "notification"},
	"dingtalk-notifications": {"notification", "dingtalk_notification"}, "notification-styles": {"notification", "notification_style"},
	"dingtalk-notification-templates": {"notification", "dingtalk_notification_template"},
	"user-feedbacks":                  {"feedback", "feedback"}, "reviews": {"performance", "review"},
	"template": {"performance", "template"}, "news": {"content", "news"},
	"events": {"content", "event"}, "enrollments": {"content", "enrollment"},
	"uploads": {"file", "file"}, "home": {"settings", "recommendation"},
	"public-portal":   {"content", "public_portal"},
	"password-policy": {"settings", "password_policy"},
	"logs":            {"audit", "audit_log"}, "login-logs": {"audit", "audit_log"},
	"account": {"user", "user"}, "me": {"user", "user"},
	"auth": {"auth", "account"}, "passport": {"auth", "account"}, "user": {"user", "user"},
	"survey": {"survey", "survey"}, "exam": {"exam", "exam"}, "enroll": {"content", "enrollment"}, "event": {"content", "event"},
	"workflow-drafts": {"workflow", "workflow_draft"}, "workflow-attachments": {"workflow", "attachment"},
	"workflow-form-revision-tasks": {"workflow", "form_revision_task"}, "workflow-summary": {"workflow", "summary"},
}

// Classify keeps ordinary reads and form evaluation traffic out of the audit stream.
func Classify(method, path string) (Operation, bool) {
	method = strings.ToUpper(method)
	path = strings.SplitN(path, "?", 2)[0]
	source, relative := sourcePath(path)
	if source == "" {
		return Operation{}, false
	}
	parts := strings.Split(strings.Trim(relative, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return Operation{}, false
	}
	last := parts[len(parts)-1]
	op := Operation{Source: source, Category: "operation"}
	if method == "POST" {
		switch relative {
		case "/auth/login", "/auth/password-login", "/login", "/passport/login_pwd":
			return Operation{Source: source, Category: "login", Module: "auth", Action: "password_login", TargetType: "account"}, true
		case "/sso-login", "/auth/dingtalk-login":
			return Operation{Source: source, Category: "login", Module: "auth", Action: "sso_login", TargetType: "account"}, true
		case "/bind-self":
			return Operation{Source: source, Category: "login", Module: "auth", Action: "bind", TargetType: "account"}, true
		case "/auth/logout", "/me/logout", "/logout", "/user/logout":
			return Operation{Source: source, Category: "login", Module: "auth", Action: "logout", TargetType: "account"}, true
		}
	}
	if method == "GET" && last != "export" {
		return Operation{}, false
	}
	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
		return Operation{}, false
	}
	if nonMutatingAuditRoute(method, path) {
		return Operation{}, false
	}
	resource := parts[0]
	if resource == "workflows" && len(parts) > 1 {
		resource = "workflow-" + parts[1]
	}
	if resource == "dingtalk" && len(parts) > 1 {
		switch parts[1] {
		case "settings":
			op.Module, op.TargetType = "settings", "dingtalk_setting"
		case "user-bindings":
			op.Module, op.TargetType = "auth", "user_binding"
		default:
			op.Module, op.TargetType = "performance", parts[1]
		}
	} else if entry, ok := resources[resource]; ok {
		op.Module, op.TargetType = entry[0], entry[1]
	} else {
		// Preserve coverage for future authenticated mutation routes, without inventing semantics.
		op.Module, op.TargetType = "other", CleanText(resource, 64)
	}
	op.Action = map[string]string{"POST": "create", "PUT": "update", "PATCH": "update", "DELETE": "delete", "GET": "export"}[method]
	if method == "POST" && resource == "user-feedbacks" && last == "notify" {
		op.Action, op.Async = "send", true
	}
	if source == "admin" && method == "POST" && resource == "exams" && len(parts) == 5 && parts[2] == "records" && last == "grading" {
		op.Action, op.TargetType, op.RelatedType = "grade", "exam_record", "exam"
	}
	if method == "POST" || method == "PATCH" && resource != "dict" {
		switch last {
		case "status", "publish", "rollback", "copy", "retry", "cancel", "withdraw", "complete", "resume", "export", "import", "clear", "finalize", "start", "register":
			op.Action = last
		case "run":
			op.Action = "run"
		case "password":
			op.Action = "password_change"
		case "force-offline", "batch-force-offline":
			op.Action = "force_offline"
		case "send", "dispatch-due":
			op.Action = "send"
		case "notification-test", "in-app", "dingtalk":
			if op.Module == "notification" || strings.Contains(relative, "/notification-test") {
				op.Action = "test_send"
			}
		case "responses", "submissions", "submit", "submit-self", "submit-manager", "submit-hrbp", "confirm-result", "dispute-result", "return-employee", "return-manager", "return-hrbp":
			op.Action = last
		}
	}
	if resource == "in-app-notifications" || resource == "dingtalk-notifications" {
		if method == "POST" && len(parts) == 1 {
			op.Action = "send"
		}
	}
	if resource == "scheduled-tasks" && op.Action == "run" || resource == "scheduled-task-runs" && op.Action == "retry" ||
		(resource == "webhook-notifications" || resource == "workflow-notifications") && (op.Action == "retry" || op.Action == "send") {
		op.Async = true
	}
	if op.TargetType == "workflow_instance" || op.TargetType == "workflow_task" || op.TargetType == "scheduled_task_run" || op.Module == "notification" || op.Module == "feedback" || op.Module == "performance" {
		op.RelatedType = op.TargetType
	}
	return op, true
}

func sourcePath(path string) (string, string) {
	for _, item := range [][2]string{{"/api/v2/admin", "admin"}, {"/api/v2/dingtalk/h5", "h5app"}, {"/api/v2", "client"}} {
		if strings.HasPrefix(path, item[0]+"/") {
			return item[1], strings.TrimPrefix(path, item[0])
		}
	}
	for _, prefix := range []string{"/passport/", "/user/", "/survey/", "/exam/", "/enroll/", "/event/", "/fav/"} {
		if strings.HasPrefix(path, prefix) {
			return "client", path
		}
	}
	return "", ""
}

func Source(path string) string {
	source, _ := sourcePath(path)
	return source
}

func nonMutatingAuditRoute(method, path string) bool {
	routes := [][2]string{
		{"POST", "/api/v2/admin/survey-schema/parse"}, {"POST", "/api/v2/admin/survey-expressions/evaluate"},
		{"POST", "/api/v2/admin/workflow-definitions/:id/validate"}, {"POST", "/api/v2/admin/scheduled-tasks/cron-preview"},
		{"POST", "/api/v2/auth/dingtalk-authorization"}, {"POST", "/api/v2/survey/validate"}, {"POST", "/api/v2/survey/apply"},
		{"POST", "/api/v2/exams/:id/validation"}, {"POST", "/survey/validate"}, {"POST", "/survey/apply"}, {"POST", "/exam/validate"},
		{"POST", "/api/v2/dingtalk/h5/workflows/instances/:id/form-revisions/preview"},
		{"POST", "/api/v2/workflows/instances/:id/form-revisions/preview"},
		{"PATCH", "/api/v2/admin/in-app-notifications/read-all"}, {"PATCH", "/api/v2/admin/in-app-notifications/:id/read"},
		{"PATCH", "/api/v2/admin/survey-notifications/:id/read"},
		{"PATCH", "/api/v2/dingtalk/h5/notifications/read-all"}, {"PATCH", "/api/v2/dingtalk/h5/notifications/:id/read"},
	}
	for _, route := range routes {
		if route[0] != method {
			continue
		}
		wanted, actual := strings.Split(route[1], "/"), strings.Split(path, "/")
		if len(wanted) != len(actual) {
			continue
		}
		match := true
		for i, part := range wanted {
			if part != actual[i] && !(strings.HasPrefix(part, ":") && actual[i] != "") {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
