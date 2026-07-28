package middleware

import (
	"context"
	"log"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	menuservice "wecheckin-backend/backend/internal/app/service/menu"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
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

type routeMethodPerm struct {
	method string
	path   string
	perm   string
}

var routeMethodPerms = map[string]string{
	"GET /api/v2/admin/home":                    "",
	"DELETE /api/v2/admin/home/recommendations": "home",

	"GET /api/v2/admin/managers":                            "mgr:list",
	"POST /api/v2/admin/managers":                           "mgr:add",
	"DELETE /api/v2/admin/managers":                         "mgr:del",
	"GET /api/v2/admin/admin-sessions":                      "online:list",
	"POST /api/v2/admin/admin-sessions/batch-force-offline": "online:force_offline",
	"POST /api/v2/admin/auth/logout":                        "",
	"GET /api/v2/admin/logs":                                "log:list",
	"DELETE /api/v2/admin/logs":                             "log:del",

	"PUT /api/v2/admin/settings":                           "setup:edit",
	"PUT /api/v2/admin/settings/content":                   "setup:edit",
	"GET /api/v2/admin/settings/mini-qr":                   "setup:edit",
	"GET /api/v2/admin/settings/debug-token":               "setup:edit",
	"GET /api/v2/admin/users":                              "user:list",
	"POST /api/v2/admin/users":                             "user:add",
	"DELETE /api/v2/admin/users":                           "user:del",
	"GET /api/v2/admin/users/form-fields":                  "user:list",
	"PUT /api/v2/admin/users/form-fields":                  "user:edit",
	"GET /api/v2/admin/users/data":                         "user:list",
	"GET /api/v2/admin/users/data/export":                  "user:list",
	"GET /api/v2/admin/user-sessions":                      "online:list",
	"POST /api/v2/admin/user-sessions/batch-force-offline": "online:force_offline",

	"GET /api/v2/admin/news":    "news:list",
	"POST /api/v2/admin/news":   "news:add",
	"DELETE /api/v2/admin/news": "news:del",

	"GET /api/v2/admin/enrollments":    "enroll:list",
	"POST /api/v2/admin/enrollments":   "enroll:add",
	"DELETE /api/v2/admin/enrollments": "enroll:del",

	"GET /api/v2/admin/events":           "event:list",
	"POST /api/v2/admin/events":          "event:add",
	"DELETE /api/v2/admin/events":        "event:del",
	"GET /api/v2/admin/event-dept-users": "event:list",

	"GET /api/v2/admin/dict/types":  "dict:list",
	"GET /api/v2/admin/dict/items":  "",
	"POST /api/v2/admin/dict/items": "dict:add",

	"GET /api/v2/admin/departments/tree": "dept:list",
	"POST /api/v2/admin/departments":     "dept:add",

	"GET /api/v2/admin/roles/application-permissions": "role:list",
	"GET /api/v2/admin/roles":                         "role:list",
	"POST /api/v2/admin/roles":                        "role:add",
	"DELETE /api/v2/admin/roles":                      "role:del",

	"GET /api/v2/admin/permissions/tree": "menu:list",
	"GET /api/v2/admin/permissions":      "menu:list",
	"POST /api/v2/admin/permissions":     "menu:add",
	"GET /api/v2/admin/me/menus":         "",
	"GET /api/v2/admin/me/perms":         "",
	"PATCH /api/v2/admin/me/password":    "mgr:edit",

	"GET /api/v2/admin/survey-types":                      "survey:list",
	"POST /api/v2/admin/survey-schema/parse":              "survey:edit",
	"POST /api/v2/admin/survey-expressions/evaluate":      "survey:edit",
	"GET /api/v2/admin/survey-report/enroll":              "survey:list",
	"GET /api/v2/admin/survey-report/enroll/export":       "response:export",
	"GET /api/v2/admin/survey-report/event":               "survey:list",
	"GET /api/v2/admin/survey-report/event/export":        "response:export",
	"GET /api/v2/admin/survey-report/survey":              "survey:list",
	"GET /api/v2/admin/survey-report/survey/export":       "response:export",
	"GET /api/v2/admin/surveys":                           "survey:list",
	"POST /api/v2/admin/surveys":                          "survey:add",
	"DELETE /api/v2/admin/survey-responses":               "response:del",
	"POST /api/v2/admin/survey-resources":                 "survey:edit",
	"GET /api/v2/admin/survey-question-bank":              "question-bank:list",
	"POST /api/v2/admin/survey-question-bank":             "question-bank:add",
	"GET /api/v2/admin/survey-question-bank/categories":   "question-bank:list",
	"GET /api/v2/admin/survey-notifications":              "survey:list",
	"GET /api/v2/admin/survey-notifications/unread-count": "survey:list",
	"GET /api/v2/admin/survey-template-presets":           "survey:list,survey:edit",
	"PUT /api/v2/admin/survey-template-presets":           "survey:list,survey:edit",

	"GET /api/v2/admin/exams":                         "exam:list",
	"POST /api/v2/admin/exams":                        "exam:add,exam:edit",
	"POST /api/v2/admin/exam-resources":               "exam:edit",
	"GET /api/v2/admin/exam-question-bank":            "question-bank:list",
	"POST /api/v2/admin/exam-question-bank":           "question-bank:add",
	"GET /api/v2/admin/exam-question-bank/categories": "question-bank:list",
}

var routeMethodPermPatterns = []routeMethodPerm{
	{method: "GET", path: "/api/v2/admin/managers/:id", perm: "mgr:list"},
	{method: "PUT", path: "/api/v2/admin/managers/:id", perm: "mgr:edit"},
	{method: "DELETE", path: "/api/v2/admin/managers/:id", perm: "mgr:del"},
	{method: "PATCH", path: "/api/v2/admin/managers/:id/status", perm: "mgr:edit"},
	{method: "PATCH", path: "/api/v2/admin/managers/:id/password", perm: "mgr:edit"},
	{method: "POST", path: "/api/v2/admin/admin-sessions/:id/force-offline", perm: "online:force_offline"},

	{method: "GET", path: "/api/v2/admin/users/:id", perm: "user:list"},
	{method: "GET", path: "/api/v2/admin/users/by-openid/:openid", perm: "user:list"},
	{method: "PUT", path: "/api/v2/admin/users/:id", perm: "user:edit"},
	{method: "DELETE", path: "/api/v2/admin/users/:id", perm: "user:del"},
	{method: "PATCH", path: "/api/v2/admin/users/:id/status", perm: "user:edit"},
	{method: "PATCH", path: "/api/v2/admin/users/:id/password", perm: "user:edit"},
	{method: "DELETE", path: "/api/v2/admin/users/data/:id", perm: "user:del"},
	{method: "POST", path: "/api/v2/admin/user-sessions/:id/force-offline", perm: "online:force_offline"},

	{method: "GET", path: "/api/v2/admin/news/:id", perm: "news:list"},
	{method: "PUT", path: "/api/v2/admin/news/:id", perm: "news:edit"},
	{method: "DELETE", path: "/api/v2/admin/news/:id", perm: "news:del"},
	{method: "PATCH", path: "/api/v2/admin/news/:id/status", perm: "news:edit"},
	{method: "PATCH", path: "/api/v2/admin/news/:id/recommendation", perm: "news:edit"},
	{method: "PATCH", path: "/api/v2/admin/news/:id/sort", perm: "news:edit"},
	{method: "PATCH", path: "/api/v2/admin/news/:id/forms", perm: "news:edit"},
	{method: "PATCH", path: "/api/v2/admin/news/:id/picture", perm: "news:edit"},
	{method: "PATCH", path: "/api/v2/admin/news/:id/content", perm: "news:edit"},

	{method: "GET", path: "/api/v2/admin/enrollments/:id", perm: "enroll:list"},
	{method: "PUT", path: "/api/v2/admin/enrollments/:id", perm: "enroll:edit"},
	{method: "DELETE", path: "/api/v2/admin/enrollments/:id", perm: "enroll:del"},
	{method: "PATCH", path: "/api/v2/admin/enrollments/:id/status", perm: "enroll:edit"},
	{method: "PATCH", path: "/api/v2/admin/enrollments/:id/sort", perm: "enroll:edit"},
	{method: "PATCH", path: "/api/v2/admin/enrollments/:id/recommendation", perm: "enroll:edit"},
	{method: "PATCH", path: "/api/v2/admin/enrollments/:id/forms", perm: "enroll:edit"},
	{method: "POST", path: "/api/v2/admin/enrollments/:id/clear", perm: "enroll:del"},
	{method: "GET", path: "/api/v2/admin/enrollments/:id/joins", perm: "enroll:list"},
	{method: "DELETE", path: "/api/v2/admin/enrollments/:id/joins/:joinId", perm: "enroll:del"},
	{method: "DELETE", path: "/api/v2/admin/enrollments/:id/joins", perm: "enroll:del"},
	{method: "GET", path: "/api/v2/admin/enrollments/:id/users", perm: "enroll:list"},
	{method: "GET", path: "/api/v2/admin/enrollments/:id/stats", perm: "enroll:list"},
	{method: "DELETE", path: "/api/v2/admin/enrollments/:id/users/:userId", perm: "enroll:del"},
	{method: "DELETE", path: "/api/v2/admin/enrollments/:id/users", perm: "enroll:del"},
	{method: "PUT", path: "/api/v2/admin/enrollments/:id/users/:userId/forms", perm: "enroll:edit"},
	{method: "GET", path: "/api/v2/admin/enrollments/:id/export", perm: "enroll:list"},
	{method: "POST", path: "/api/v2/admin/enrollments/:id/export", perm: "enroll:list"},
	{method: "DELETE", path: "/api/v2/admin/enrollments/:id/export", perm: "enroll:del"},

	{method: "GET", path: "/api/v2/admin/events/:id", perm: "event:list"},
	{method: "PUT", path: "/api/v2/admin/events/:id", perm: "event:edit"},
	{method: "DELETE", path: "/api/v2/admin/events/:id", perm: "event:del"},
	{method: "PATCH", path: "/api/v2/admin/events/:id/status", perm: "event:edit"},
	{method: "PATCH", path: "/api/v2/admin/events/:id/recommendation", perm: "event:edit"},
	{method: "PATCH", path: "/api/v2/admin/events/:id/top", perm: "event:edit"},
	{method: "GET", path: "/api/v2/admin/events/:id/participants", perm: "event:list"},
	{method: "PUT", path: "/api/v2/admin/events/:id/participants/:participantId", perm: "event:edit"},
	{method: "DELETE", path: "/api/v2/admin/events/:id/participants/:participantId", perm: "event:del"},
	{method: "DELETE", path: "/api/v2/admin/events/:id/participants", perm: "event:del"},
	{method: "GET", path: "/api/v2/admin/events/:id/dynamics", perm: "event:list"},
	{method: "POST", path: "/api/v2/admin/events/:id/dynamics", perm: "event:add"},
	{method: "PUT", path: "/api/v2/admin/events/:id/dynamics/:dynamicId", perm: "event:edit"},
	{method: "DELETE", path: "/api/v2/admin/events/:id/dynamics/:dynamicId", perm: "event:del"},
	{method: "DELETE", path: "/api/v2/admin/events/:id/dynamics", perm: "event:del"},
	{method: "GET", path: "/api/v2/admin/events/:id/scores", perm: "event:list"},
	{method: "POST", path: "/api/v2/admin/events/:id/scores", perm: "event:edit"},
	{method: "PUT", path: "/api/v2/admin/events/:id/scores/:scoreId", perm: "event:edit"},

	{method: "PUT", path: "/api/v2/admin/dict/items/:id", perm: "dict:edit"},
	{method: "DELETE", path: "/api/v2/admin/dict/items/:id", perm: "dict:del"},
	{method: "DELETE", path: "/api/v2/admin/dict/types/:typeCode/items", perm: "dict:del"},
	{method: "PATCH", path: "/api/v2/admin/dict/types/:typeCode", perm: "dict:edit"},
	{method: "PUT", path: "/api/v2/admin/departments/:id", perm: "dept:edit"},
	{method: "DELETE", path: "/api/v2/admin/departments/:id", perm: "dept:del"},
	{method: "PUT", path: "/api/v2/admin/roles/:id", perm: "role:edit"},
	{method: "DELETE", path: "/api/v2/admin/roles/:id", perm: "role:del"},
	{method: "PUT", path: "/api/v2/admin/permissions/:key", perm: "menu:edit"},
	{method: "DELETE", path: "/api/v2/admin/permissions/:key", perm: "menu:del"},

	{method: "GET", path: "/api/v2/admin/surveys/:id", perm: "survey:list"},
	{method: "PUT", path: "/api/v2/admin/surveys/:id", perm: "survey:edit"},
	{method: "DELETE", path: "/api/v2/admin/surveys/:id", perm: "survey:del"},
	{method: "PATCH", path: "/api/v2/admin/surveys/:id/status", perm: "survey:status"},
	{method: "POST", path: "/api/v2/admin/surveys/:id/copy", perm: "survey:copy"},
	{method: "GET", path: "/api/v2/admin/surveys/:id/statistics", perm: "survey:list"},
	{method: "GET", path: "/api/v2/admin/surveys/:id/responses", perm: "response:list"},
	{method: "GET", path: "/api/v2/admin/surveys/:id/responses/export", perm: "response:export"},
	{method: "GET", path: "/api/v2/admin/survey-responses/:id", perm: "response:list"},
	{method: "DELETE", path: "/api/v2/admin/survey-responses/:id", perm: "response:del"},
	{method: "GET", path: "/api/v2/admin/surveys/:id/channels", perm: "survey:list"},
	{method: "POST", path: "/api/v2/admin/surveys/:id/channels", perm: "survey:edit"},
	{method: "DELETE", path: "/api/v2/admin/survey-channels/:id", perm: "survey:edit"},
	{method: "GET", path: "/api/v2/admin/surveys/:id/resources", perm: "survey:list"},
	{method: "DELETE", path: "/api/v2/admin/survey-resources/:id", perm: "survey:edit"},
	{method: "PUT", path: "/api/v2/admin/survey-question-bank/:id", perm: "question-bank:edit"},
	{method: "DELETE", path: "/api/v2/admin/survey-question-bank/:id", perm: "question-bank:del"},
	{method: "PATCH", path: "/api/v2/admin/survey-notifications/:id/read", perm: "survey:list"},

	{method: "GET", path: "/api/v2/admin/exams/:id", perm: "exam:list"},
	{method: "PUT", path: "/api/v2/admin/exams/:id", perm: "exam:edit"},
	{method: "DELETE", path: "/api/v2/admin/exams/:id", perm: "exam:del"},
	{method: "PATCH", path: "/api/v2/admin/exams/:id/status", perm: "exam:edit"},
	{method: "GET", path: "/api/v2/admin/exams/:id/records", perm: "exam:list"},
	{method: "GET", path: "/api/v2/admin/exams/:id/records/:recordId", perm: "exam:list"},
	{method: "DELETE", path: "/api/v2/admin/exams/:id/records/:recordId", perm: "exam:del"},
	{method: "DELETE", path: "/api/v2/admin/exams/:id/records", perm: "exam:del"},
	{method: "GET", path: "/api/v2/admin/exams/:id/statistics", perm: "exam:list"},
	{method: "GET", path: "/api/v2/admin/exams/:id/resources", perm: "exam:list"},
	{method: "DELETE", path: "/api/v2/admin/exam-resources/:id", perm: "exam:edit"},
	{method: "PUT", path: "/api/v2/admin/exam-question-bank/:id", perm: "question-bank:edit"},
	{method: "DELETE", path: "/api/v2/admin/exam-question-bank/:id", perm: "question-bank:del"},
}

func AdminPerm() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		adminVal, _ := c.Get("admin")
		admin := adminVal.(*model.Admin)

		// The reserved super admin role bypasses route permission checks.
		if menuservice.AdminHasReservedSuperAdminRoleContext(ctx, admin) {
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

		required, ok := adminRoutePermission(string(c.Method()), path)
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

		db, cancel := database.WithContext(ctx)
		defer cancel()
		requiredCodes := permissionCodes(required)
		for _, code := range requiredCodes {
			apiKey := permissionsupport.AdminAPIPermissionKey(code)
			if effect, ok, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, admin.ID, apiKey); err == nil && ok && effect == permissionsupport.EffectDeny {
				continue
			}
			if ok, err := permissionsupport.SubjectHasPermissionContext(ctx, db, admin.ID, admin.RoleID, apiKey); err == nil && ok {
				c.Next(ctx)
				return
			}
		}

		perms := menuservice.GetAdminPermsContext(ctx, admin)
		for _, code := range requiredCodes {
			apiKey := permissionsupport.AdminAPIPermissionKey(code)
			if effect, ok, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, admin.ID, apiKey); err == nil && ok && effect == permissionsupport.EffectDeny {
				continue
			}
			for _, p := range perms {
				if code == p {
					c.Next(ctx)
					return
				}
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

func adminRoutePermission(method, path string) (string, bool) {
	method = strings.ToUpper(method)
	if required, ok := routeMethodPerms[method+" "+path]; ok {
		return required, true
	}
	for _, route := range routeMethodPermPatterns {
		if route.method == method && routePatternMatches(route.path, path) {
			return route.perm, true
		}
	}
	required, ok := routePerms[path]
	return required, ok
}

func routePatternMatches(pattern, path string) bool {
	patternParts := routePatternParts(pattern)
	pathParts := routePatternParts(path)
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func routePatternParts(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
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

func permissionCodes(required string) []string {
	codes := make([]string, 0)
	seen := map[string]bool{}
	for _, item := range strings.Split(required, ",") {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		codes = append(codes, item)
	}
	return codes
}
