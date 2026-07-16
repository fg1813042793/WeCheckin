package middleware

import (
	"context"
	"log"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	menuservice "wecheckin-backend/backend/internal/app/service/menu"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/logger"
)

// routePerms maps admin route paths to required permission keys.
// Routes listed with an empty permission are explicitly allowed after login.
var routePerms = map[string]string{
	"/admin/home":        "",
	"/admin/clear_vouch": "home",

	"/admin/user_list":                "user:list",
	"/admin/user_detail":              "user:list",
	"/admin/user_detail_by_id":        "user:list",
	"/admin/user_add":                 "user:add",
	"/admin/user_edit":                "user:edit",
	"/admin/user_del":                 "user:del",
	"/admin/user_dels":                "user:del",
	"/admin/user_status":              "user:edit",
	"/admin/user_reset_pwd":           "user:edit",
	"/admin/user_form_fields":         "user:list",
	"/admin/user_form_field_save":     "user:edit",
	"/admin/user_data_get":            "user:list",
	"/admin/user_data_export":         "user:list",
	"/admin/user_data_del":            "user:del",
	"/admin/user/online":              "online:list",
	"/admin/user/force_offline":       "online:force_offline",
	"/admin/user/batch_force_offline": "online:force_offline",

	"/admin/enroll_list":             "enroll:list",
	"/admin/enroll_detail":           "enroll:list",
	"/admin/enroll_insert":           "enroll:add",
	"/admin/enroll_edit":             "enroll:edit",
	"/admin/enroll_del":              "enroll:del",
	"/admin/enroll_dels":             "enroll:del",
	"/admin/enroll_status":           "enroll:edit",
	"/admin/enroll_sort":             "enroll:edit",
	"/admin/enroll_vouch":            "enroll:edit",
	"/admin/enroll_clear":            "enroll:del",
	"/admin/enroll_update_forms":     "enroll:edit",
	"/admin/enroll_join_list":        "enroll:list",
	"/admin/enroll_join_del":         "enroll:del",
	"/admin/enroll_join_dels":        "enroll:del",
	"/admin/enroll_user_list":        "enroll:list",
	"/admin/enroll_stats":            "enroll:list",
	"/admin/enroll_remove_user":      "enroll:del",
	"/admin/enroll_remove_users":     "enroll:del",
	"/admin/enroll_user_forms_edit":  "enroll:edit",
	"/admin/enroll_join_data_get":    "enroll:list",
	"/admin/enroll_join_data_export": "enroll:list",
	"/admin/enroll_join_data_del":    "enroll:del",

	"/admin/news_list":           "news:list",
	"/admin/news_detail":         "news:list",
	"/admin/news_insert":         "news:add",
	"/admin/news_edit":           "news:edit",
	"/admin/news_del":            "news:del",
	"/admin/news_dels":           "news:del",
	"/admin/news_status":         "news:edit",
	"/admin/news_sort":           "news:edit",
	"/admin/news_vouch":          "news:edit",
	"/admin/news_update_forms":   "news:edit",
	"/admin/news_update_pic":     "news:edit",
	"/admin/news_update_content": "news:edit",

	"/admin/mgr_list":                  "mgr:list",
	"/admin/mgr_detail":                "mgr:list",
	"/admin/mgr_insert":                "mgr:add",
	"/admin/mgr_edit":                  "mgr:edit",
	"/admin/mgr_del":                   "mgr:del",
	"/admin/mgr_dels":                  "mgr:del",
	"/admin/mgr_status":                "mgr:edit",
	"/admin/mgr_pwd":                   "mgr:edit",
	"/admin/admin/online":              "online:list",
	"/admin/admin/force_offline":       "online:force_offline",
	"/admin/admin/batch_force_offline": "online:force_offline",
	"/admin/admin/logout":              "",

	"/admin/setup_set":         "setup:edit",
	"/admin/setup_set_content": "setup:edit",
	"/admin/setup_qr":          "setup:edit",
	"/admin/setup_debug_token": "setup:edit",

	"/admin/dict/types":          "dict:list",
	"/admin/dict/items":          "",
	"/admin/dict/add":            "dict:add",
	"/admin/dict/edit":           "dict:edit",
	"/admin/dict/del":            "dict:del",
	"/admin/dict/clear":          "dict:del",
	"/admin/dict/edit_type_name": "dict:edit",

	"/admin/log_list":  "log:list",
	"/admin/log_clear": "log:del",

	"/admin/event_list":             "event:list",
	"/admin/event_detail":           "event:list",
	"/admin/event_insert":           "event:add",
	"/admin/event_edit":             "event:edit",
	"/admin/event_del":              "event:del",
	"/admin/event_dels":             "event:del",
	"/admin/event_status":           "event:edit",
	"/admin/event_participant_list": "event:list",
	"/admin/event_participant_del":  "event:del",
	"/admin/event_participant_dels": "event:del",
	"/admin/event_participant_edit": "event:edit",
	"/admin/event_dynamics":         "event:list",
	"/admin/event_dynamic_add":      "event:add",
	"/admin/event_dynamic_edit":     "event:edit",
	"/admin/event_dynamic_del":      "event:del",
	"/admin/event_dynamic_dels":     "event:del",
	"/admin/event_scores":           "event:list",
	"/admin/event_score_edit":       "event:edit",
	"/admin/dept_users":             "event:list",
	"/admin/event_vouch":            "event:edit",
	"/admin/event_top":              "event:edit",

	"/admin/dept/tree": "dept:list",
	"/admin/dept/add":  "dept:add",
	"/admin/dept/edit": "dept:edit",
	"/admin/dept/del":  "dept:del",

	"/admin/role/list": "role:list",
	"/admin/role/add":  "role:add",
	"/admin/role/edit": "role:edit",
	"/admin/role/del":  "role:del",
	"/admin/role/dels": "role:del",

	"/admin/menu/tree": "menu:list",
	"/admin/menu/list": "menu:list",
	"/admin/menu/add":  "menu:add",
	"/admin/menu/edit": "menu:edit",
	"/admin/menu/del":  "menu:del",

	"/admin/user/menus": "",
	"/admin/user/perms": "",

	"/admin/survey/types":                    "survey:list",
	"/admin/survey/schema/parse":             "survey:edit",
	"/admin/survey/eval":                     "survey:edit",
	"/admin/survey/report/enroll":            "survey:list",
	"/admin/survey/export/enroll":            "response:export",
	"/admin/survey/report/event":             "survey:list",
	"/admin/survey/export/event":             "response:export",
	"/admin/survey/report/survey":            "survey:list",
	"/admin/survey/export/survey":            "response:export",
	"/admin/survey/survey_list":              "survey:list",
	"/admin/survey/survey_detail":            "survey:list",
	"/admin/survey/survey_insert":            "survey:add",
	"/admin/survey/survey_edit":              "survey:edit",
	"/admin/survey/survey_del":               "survey:del",
	"/admin/survey/survey_status":            "survey:status",
	"/admin/survey/survey_copy":              "survey:copy",
	"/admin/survey/response_list":            "response:list",
	"/admin/survey/response_detail":          "response:list",
	"/admin/survey/response_del":             "response:del",
	"/admin/survey/response_batch_del":       "response:del",
	"/admin/survey/response_export":          "response:export",
	"/admin/survey/statistic":                "survey:list",
	"/admin/survey/channel_list":             "survey:list",
	"/admin/survey/channel_insert":           "survey:edit",
	"/admin/survey/channel_del":              "survey:edit",
	"/admin/survey/resource_upload":          "survey:edit",
	"/admin/survey/resource_list":            "survey:list",
	"/admin/survey/resource_delete":          "survey:edit",
	"/admin/survey/question_bank_list":       "question-bank:list",
	"/admin/survey/question_bank_insert":     "question-bank:add",
	"/admin/survey/question_bank_edit":       "question-bank:edit",
	"/admin/survey/question_bank_del":        "question-bank:del",
	"/admin/survey/question_bank_categories": "question-bank:list",
	"/admin/survey/notify_list":              "survey:list",
	"/admin/survey/notify_read":              "survey:list",
	"/admin/survey/notify_unread_count":      "survey:list",
	"/admin/survey/template_presets":         "survey:list,survey:edit",

	"/admin/exam/list":                     "exam:list",
	"/admin/exam/detail":                   "exam:list",
	"/admin/exam/save":                     "exam:add,exam:edit",
	"/admin/exam/status":                   "exam:edit",
	"/admin/exam/delete":                   "exam:del",
	"/admin/exam/record/list":              "exam:list",
	"/admin/exam/record/detail":            "exam:list",
	"/admin/exam/record/del":               "exam:del",
	"/admin/exam/record/batch_del":         "exam:del",
	"/admin/exam/statistics":               "exam:list",
	"/admin/exam/resource_upload":          "exam:edit",
	"/admin/exam/resource_list":            "exam:list",
	"/admin/exam/resource_delete":          "exam:edit",
	"/admin/exam/question_bank_list":       "question-bank:list",
	"/admin/exam/question_bank_insert":     "question-bank:add",
	"/admin/exam/question_bank_edit":       "question-bank:edit",
	"/admin/exam/question_bank_del":        "question-bank:del",
	"/admin/exam/question_bank_categories": "question-bank:list",
}

func AdminPerm() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		adminVal, _ := c.Get("admin")
		admin := adminVal.(*model.Admin)

		// Super admin bypasses all permission checks
		if admin.Type == 1 {
			c.Next(ctx)
			return
		}

		// No role assigned - no access
		if admin.RoleID == 0 {
			auditAdminPermissionDenied(admin, string(c.Path()), "", "role_missing")
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "未分配角色，无权限",
			})
			c.Abort()
			return
		}

		path := string(c.Path())
		// Remove query string if any
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}

		required, ok := routePerms[path]
		if !ok {
			auditAdminPermissionDenied(admin, path, "", "route_not_declared")
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "无权限访问",
			})
			c.Abort()
			return
		}
		if required == "" {
			// Explicitly allowed after login.
			c.Next(ctx)
			return
		}

		perms := menuservice.GetAdminPermsContext(ctx, admin)
		for _, p := range perms {
			if permissionMatches(required, p) {
				c.Next(ctx)
				return
			}
		}

		c.JSON(consts.StatusOK, utils.H{
			"code": 1,
			"msg":  "无权限访问",
		})
		auditAdminPermissionDenied(admin, path, required, "permission_missing")
		c.Abort()
	}
}

func auditAdminPermissionDenied(admin *model.Admin, path, required, reason string) {
	writer := logger.Logger
	if writer == nil {
		writer = log.Default()
	}
	if admin == nil {
		writer.Printf("[AdminPermDenied] admin=<nil> path=%s required=%s reason=%s", path, required, reason)
		return
	}
	writer.Printf("[AdminPermDenied] adminId=%d roleId=%d path=%s required=%s reason=%s", admin.ID, admin.RoleID, path, required, reason)
}

func permissionMatches(required, actual string) bool {
	for _, item := range strings.Split(required, ",") {
		if strings.TrimSpace(item) == actual {
			return true
		}
	}
	return false
}
