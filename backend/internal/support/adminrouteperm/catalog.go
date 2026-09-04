package adminrouteperm

import "strings"

type Declaration struct {
	Key          string
	Name         string
	Perms        string
	CategoryKey  string
	CategoryName string
	Method       string
	Path         string
}

type Category struct {
	Key  string
	Name string
	Sort int
}

func Declarations() []Declaration {
	categories := categoryMap()
	codes := []struct {
		perms string
		name  string
	}{
		{"home", "控制台操作接口"},
		{"user:list", "用户查看接口"},
		{"user:add", "用户创建接口"},
		{"user:edit", "用户编辑接口"},
		{"user:del", "用户删除接口"},
		{"online:list", "在线用户查看接口"},
		{"online:force_offline", "在线用户强制下线接口"},
		{"enroll:list", "报名查看接口"},
		{"enroll:add", "报名创建接口"},
		{"enroll:edit", "报名编辑接口"},
		{"enroll:del", "报名删除接口"},
		{"enroll:status", "报名状态管理接口"},
		{"enroll:vouch", "报名推荐接口"},
		{"enroll:export", "报名导出接口"},
		{"enroll:users", "报名参与用户接口"},
		{"news:list", "通知查看接口"},
		{"news:add", "通知创建接口"},
		{"news:edit", "通知编辑接口"},
		{"news:del", "通知删除接口"},
		{"news:status", "通知状态管理接口"},
		{"news:vouch", "通知推荐接口"},
		{"mgr:list", "管理员查看接口"},
		{"mgr:add", "管理员创建接口"},
		{"mgr:edit", "管理员编辑接口"},
		{"mgr:del", "管理员删除接口"},
		{"setup:list", "系统设置查看接口"},
		{"setup:edit", "系统设置接口"},
		{"upload:create", "后台文件上传接口"},
		{"dingtalk:settings:list", "钉钉配置查看接口"},
		{"dingtalk:settings:edit", "钉钉配置保存接口"},
		{"dingtalk:bindings:list", "钉钉用户绑定查看接口"},
		{"dingtalk:bindings:edit", "钉钉用户绑定维护接口"},
		{"dingtalk:perf-reviews:list", "绩效考评单查看接口"},
		{"dingtalk:perf-reviews:detail", "绩效考评单详情接口"},
		{"dingtalk:perf-reviews:del", "绩效考评单删除接口"},
		{"dingtalk:perf-histories:list", "绩效流转记录查看接口"},
		{"dingtalk:perf-histories:del", "绩效流转记录删除接口"},
		{"dict:list", "字典查看接口"},
		{"dict:add", "字典创建接口"},
		{"dict:edit", "字典编辑接口"},
		{"dict:del", "字典删除接口"},
		{"log:list", "日志查看接口"},
		{"log:del", "日志删除接口"},
		{"event:list", "赛事活动查看接口"},
		{"event:add", "赛事活动创建接口"},
		{"event:edit", "赛事活动编辑接口"},
		{"event:del", "赛事活动删除接口"},
		{"event:status", "赛事活动状态管理接口"},
		{"event:vouch", "赛事活动推荐接口"},
		{"event:top", "赛事活动置顶接口"},
		{"event:users", "赛事活动参与用户接口"},
		{"dept:list", "部门查看接口"},
		{"dept:add", "部门创建接口"},
		{"dept:edit", "部门编辑接口"},
		{"dept:del", "部门删除接口"},
		{"position:list", "岗位查看接口"},
		{"position:add", "岗位创建接口"},
		{"position:edit", "岗位编辑接口"},
		{"position:del", "岗位删除接口"},
		{"role:list", "角色查看接口"},
		{"role:add", "角色创建接口"},
		{"role:edit", "角色编辑接口"},
		{"role:del", "角色删除接口"},
		{"menu:list", "菜单查看接口"},
		{"menu:add", "菜单创建接口"},
		{"menu:edit", "菜单编辑接口"},
		{"menu:del", "菜单删除接口"},
		{"survey:list", "问卷查看接口"},
		{"survey:add", "问卷创建接口"},
		{"survey:edit", "问卷编辑接口"},
		{"survey:del", "问卷删除接口"},
		{"survey:status", "问卷状态接口"},
		{"survey:copy", "问卷复制接口"},
		{"response:list", "答卷查看接口"},
		{"response:del", "答卷删除接口"},
		{"response:export", "答卷导出接口"},
		{"question-bank:list", "题库查看接口"},
		{"question-bank:add", "题库创建接口"},
		{"question-bank:edit", "题库编辑接口"},
		{"question-bank:del", "题库删除接口"},
		{"exam:list", "考试查看接口"},
		{"exam:add", "考试创建接口"},
		{"exam:edit", "考试编辑接口"},
		{"exam:del", "考试删除接口"},
		{"workflow:list", "流程定义查看接口"},
		{"workflow:add", "流程定义创建接口"},
		{"workflow:edit", "流程定义编辑接口"},
		{"workflow:publish", "流程定义发布接口"},
		{"workflow:del", "流程定义删除接口"},
		{"workflow:instance:list", "流程实例查看接口"},
		{"workflow:instance:start", "流程实例发起接口"},
		{"workflow:instance:detail", "流程实例详情接口"},
		{"workflow:instance:cancel", "流程实例取消接口"},
		{"workflow:instance:delete", "流程实例删除接口"},
		{"workflow:task:list", "流程任务查看接口"},
		{"workflow:task:complete", "流程任务处理接口"},
		{"workflow:task:delete", "流程任务删除接口"},
		{"workflow:notification:list", "流程通知查看接口"},
		{"workflow:notification:retry", "流程通知重试接口"},
		{"workflow:org-approver:list", "组织审批身份查看接口"},
		{"workflow:org-approver:edit", "组织审批身份维护接口"},
		{"scheduled-task:list", "定时任务查看接口"},
		{"scheduled-task:add", "定时任务创建接口"},
		{"scheduled-task:edit", "定时任务编辑接口"},
		{"scheduled-task:delete", "定时任务删除接口"},
		{"scheduled-task:status", "定时任务启停接口"},
		{"scheduled-task:run", "定时任务立即运行接口"},
		{"scheduled-task:run:list", "定时任务运行记录查看接口"},
		{"scheduled-task:run:retry", "定时任务运行重试接口"},
		{"scheduled-task:run:cancel", "定时任务运行取消接口"},
		{"scheduled-task:worker:list", "定时任务执行节点查看接口"},
		{"scheduled-task:http", "定时任务 HTTP 处理器配置接口"},
		{"scheduled-task:shell", "定时任务 Shell 处理器配置接口"},
		{"scheduled-task:sql:read", "定时任务 SQL 查询配置接口"},
		{"scheduled-task:sql:write", "定时任务 SQL 写入配置接口"},
		{"notification:list", "站内信查看接口"},
		{"notification:read", "站内信已读接口"},
		{"notification:send", "站内信发送接口"},
		{"notification:dingtalk:send", "钉钉通知发送接口"},
	}
	out := make([]Declaration, 0, len(codes))
	for _, item := range codes {
		category := categoryForPerms(categories, item.perms)
		route := primaryRouteForPerms(item.perms)
		out = append(out, Declaration{
			Key:          KeyForPerms(item.perms),
			Name:         item.name,
			Perms:        item.perms,
			CategoryKey:  category.Key,
			CategoryName: category.Name,
			Method:       route.method,
			Path:         route.path,
		})
	}
	return out
}

func Categories() []Category {
	return []Category{
		{Key: "admin:api-category:dashboard", Name: "控制台", Sort: 10},
		{Key: "admin:api-category:user", Name: "用户与在线", Sort: 20},
		{Key: "admin:api-category:content", Name: "内容运营", Sort: 30},
		{Key: "admin:api-category:system", Name: "系统权限", Sort: 40},
		{Key: "admin:api-category:dingtalk", Name: "钉钉应用", Sort: 50},
		{Key: "admin:api-category:survey", Name: "问卷管理", Sort: 60},
		{Key: "admin:api-category:exam", Name: "考试管理", Sort: 70},
		{Key: "admin:api-category:workflow", Name: "流程管理", Sort: 75},
		{Key: "admin:api-category:scheduled-task", Name: "定时任务", Sort: 80},
		{Key: "admin:api-category:notification", Name: "通知", Sort: 85},
	}
}

func categoryMap() map[string]Category {
	result := make(map[string]Category)
	for _, item := range Categories() {
		result[item.Key] = item
	}
	return result
}

func categoryForPerms(categories map[string]Category, perms string) Category {
	module := strings.Split(strings.TrimSpace(perms), ":")[0]
	key := "admin:api-category:system"
	switch module {
	case "home":
		key = "admin:api-category:dashboard"
	case "user", "online", "position":
		key = "admin:api-category:user"
	case "enroll", "event", "news":
		key = "admin:api-category:content"
	case "dingtalk":
		key = "admin:api-category:dingtalk"
	case "survey", "response", "question-bank":
		key = "admin:api-category:survey"
	case "exam":
		key = "admin:api-category:exam"
	case "workflow":
		key = "admin:api-category:workflow"
	case "scheduled-task":
		key = "admin:api-category:scheduled-task"
	case "notification":
		key = "admin:api-category:notification"
	}
	if item, ok := categories[key]; ok {
		return item
	}
	return Category{Key: key, Name: "系统权限", Sort: 40}
}

func KeyForPerms(perms string) string {
	perms = strings.TrimSpace(perms)
	if perms == "" {
		return ""
	}
	return "admin:api:" + perms
}

type primaryRoute struct {
	method string
	path   string
}

func primaryRouteForPerms(perms string) primaryRoute {
	if route, ok := primaryAdminAPIRoutes[strings.TrimSpace(perms)]; ok {
		return route
	}
	return primaryRoute{}
}

var primaryAdminAPIRoutes = map[string]primaryRoute{
	"home":                         {method: "DELETE", path: "/api/v2/admin/home/recommendations"},
	"user:list":                    {method: "GET", path: "/api/v2/admin/users"},
	"user:add":                     {method: "POST", path: "/api/v2/admin/users"},
	"user:edit":                    {method: "PUT", path: "/api/v2/admin/users/:id"},
	"user:del":                     {method: "DELETE", path: "/api/v2/admin/users/:id"},
	"online:list":                  {method: "GET", path: "/api/v2/admin/admin-sessions"},
	"online:force_offline":         {method: "POST", path: "/api/v2/admin/admin-sessions/:id/force-offline"},
	"enroll:list":                  {method: "GET", path: "/api/v2/admin/enrollments"},
	"enroll:add":                   {method: "POST", path: "/api/v2/admin/enrollments"},
	"enroll:edit":                  {method: "PUT", path: "/api/v2/admin/enrollments/:id"},
	"enroll:del":                   {method: "DELETE", path: "/api/v2/admin/enrollments/:id"},
	"enroll:status":                {method: "PATCH", path: "/api/v2/admin/enrollments/:id/status"},
	"enroll:vouch":                 {method: "PATCH", path: "/api/v2/admin/enrollments/:id/recommendation"},
	"enroll:export":                {method: "GET", path: "/api/v2/admin/enrollments/:id/export"},
	"enroll:users":                 {method: "GET", path: "/api/v2/admin/enrollments/:id/users"},
	"news:list":                    {method: "GET", path: "/api/v2/admin/news"},
	"news:add":                     {method: "POST", path: "/api/v2/admin/news"},
	"news:edit":                    {method: "PUT", path: "/api/v2/admin/news/:id"},
	"news:del":                     {method: "DELETE", path: "/api/v2/admin/news/:id"},
	"news:status":                  {method: "PATCH", path: "/api/v2/admin/news/:id/status"},
	"news:vouch":                   {method: "PATCH", path: "/api/v2/admin/news/:id/recommendation"},
	"mgr:list":                     {method: "GET", path: "/api/v2/admin/managers"},
	"mgr:add":                      {method: "POST", path: "/api/v2/admin/managers"},
	"mgr:edit":                     {method: "PUT", path: "/api/v2/admin/managers/:id"},
	"mgr:del":                      {method: "DELETE", path: "/api/v2/admin/managers/:id"},
	"setup:list":                   {method: "GET", path: "/api/v2/admin/settings/content"},
	"setup:edit":                   {method: "PUT", path: "/api/v2/admin/settings"},
	"upload:create":                {method: "POST", path: "/api/v2/admin/uploads"},
	"dingtalk:settings:list":       {method: "GET", path: "/api/v2/admin/dingtalk/settings"},
	"dingtalk:settings:edit":       {method: "PUT", path: "/api/v2/admin/dingtalk/settings"},
	"dingtalk:bindings:list":       {method: "GET", path: "/api/v2/admin/dingtalk/user-bindings"},
	"dingtalk:bindings:edit":       {method: "POST", path: "/api/v2/admin/dingtalk/user-bindings"},
	"dingtalk:perf-reviews:list":   {method: "GET", path: "/api/v2/admin/dingtalk/perf-reviews"},
	"dingtalk:perf-reviews:detail": {method: "GET", path: "/api/v2/admin/dingtalk/perf-reviews/:id"},
	"dingtalk:perf-reviews:del":    {method: "DELETE", path: "/api/v2/admin/dingtalk/perf-reviews/:id"},
	"dingtalk:perf-histories:list": {method: "GET", path: "/api/v2/admin/dingtalk/perf-histories"},
	"dingtalk:perf-histories:del":  {method: "DELETE", path: "/api/v2/admin/dingtalk/perf-histories/:id"},
	"dict:list":                    {method: "GET", path: "/api/v2/admin/dict/types"},
	"dict:add":                     {method: "POST", path: "/api/v2/admin/dict/items"},
	"dict:edit":                    {method: "PUT", path: "/api/v2/admin/dict/items/:id"},
	"dict:del":                     {method: "DELETE", path: "/api/v2/admin/dict/items/:id"},
	"log:list":                     {method: "GET", path: "/api/v2/admin/logs"},
	"log:del":                      {method: "DELETE", path: "/api/v2/admin/logs"},
	"event:list":                   {method: "GET", path: "/api/v2/admin/events"},
	"event:add":                    {method: "POST", path: "/api/v2/admin/events"},
	"event:edit":                   {method: "PUT", path: "/api/v2/admin/events/:id"},
	"event:del":                    {method: "DELETE", path: "/api/v2/admin/events/:id"},
	"event:status":                 {method: "PATCH", path: "/api/v2/admin/events/:id/status"},
	"event:vouch":                  {method: "PATCH", path: "/api/v2/admin/events/:id/recommendation"},
	"event:top":                    {method: "PATCH", path: "/api/v2/admin/events/:id/top"},
	"event:users":                  {method: "GET", path: "/api/v2/admin/events/:id/participants"},
	"dept:list":                    {method: "GET", path: "/api/v2/admin/departments/tree"},
	"dept:add":                     {method: "POST", path: "/api/v2/admin/departments"},
	"dept:edit":                    {method: "PUT", path: "/api/v2/admin/departments/:id"},
	"dept:del":                     {method: "DELETE", path: "/api/v2/admin/departments/:id"},
	"position:list":                {method: "GET", path: "/api/v2/admin/positions"},
	"position:add":                 {method: "POST", path: "/api/v2/admin/positions"},
	"position:edit":                {method: "PUT", path: "/api/v2/admin/positions/:id"},
	"position:del":                 {method: "DELETE", path: "/api/v2/admin/positions/:id"},
	"role:list":                    {method: "GET", path: "/api/v2/admin/roles"},
	"role:add":                     {method: "POST", path: "/api/v2/admin/roles"},
	"role:edit":                    {method: "PUT", path: "/api/v2/admin/roles/:id"},
	"role:del":                     {method: "DELETE", path: "/api/v2/admin/roles/:id"},
	"menu:list":                    {method: "GET", path: "/api/v2/admin/permissions/tree"},
	"menu:add":                     {method: "POST", path: "/api/v2/admin/permissions"},
	"menu:edit":                    {method: "PUT", path: "/api/v2/admin/permissions/:key"},
	"menu:del":                     {method: "DELETE", path: "/api/v2/admin/permissions/:key"},
	"survey:list":                  {method: "GET", path: "/api/v2/admin/surveys"},
	"survey:add":                   {method: "POST", path: "/api/v2/admin/surveys"},
	"survey:edit":                  {method: "PUT", path: "/api/v2/admin/surveys/:id"},
	"survey:del":                   {method: "DELETE", path: "/api/v2/admin/surveys/:id"},
	"survey:status":                {method: "PATCH", path: "/api/v2/admin/surveys/:id/status"},
	"survey:copy":                  {method: "POST", path: "/api/v2/admin/surveys/:id/copy"},
	"response:list":                {method: "GET", path: "/api/v2/admin/surveys/:id/responses"},
	"response:del":                 {method: "DELETE", path: "/api/v2/admin/survey-responses/:id"},
	"response:export":              {method: "GET", path: "/api/v2/admin/surveys/:id/responses/export"},
	"question-bank:list":           {method: "GET", path: "/api/v2/admin/survey-question-bank"},
	"question-bank:add":            {method: "POST", path: "/api/v2/admin/survey-question-bank"},
	"question-bank:edit":           {method: "PUT", path: "/api/v2/admin/survey-question-bank/:id"},
	"question-bank:del":            {method: "DELETE", path: "/api/v2/admin/survey-question-bank/:id"},
	"exam:list":                    {method: "GET", path: "/api/v2/admin/exams"},
	"exam:add":                     {method: "POST", path: "/api/v2/admin/exams"},
	"exam:edit":                    {method: "PUT", path: "/api/v2/admin/exams/:id"},
	"exam:del":                     {method: "DELETE", path: "/api/v2/admin/exams/:id"},
	"workflow:list":                {method: "GET", path: "/api/v2/admin/workflow-definitions"},
	"workflow:add":                 {method: "POST", path: "/api/v2/admin/workflow-definitions"},
	"workflow:edit":                {method: "PUT", path: "/api/v2/admin/workflow-definitions/:id"},
	"workflow:publish":             {method: "POST", path: "/api/v2/admin/workflow-definitions/:id/publish"},
	"workflow:del":                 {method: "DELETE", path: "/api/v2/admin/workflow-definitions/:id"},
	"workflow:instance:list":       {method: "GET", path: "/api/v2/admin/workflow-instances"},
	"workflow:instance:start":      {method: "POST", path: "/api/v2/admin/workflow-instances"},
	"workflow:instance:detail":     {method: "GET", path: "/api/v2/admin/workflow-instances/:id"},
	"workflow:instance:cancel":     {method: "POST", path: "/api/v2/admin/workflow-instances/:id/cancel"},
	"workflow:instance:delete":     {method: "DELETE", path: "/api/v2/admin/workflow-instances/:id"},
	"workflow:task:list":           {method: "GET", path: "/api/v2/admin/workflow-tasks"},
	"workflow:task:complete":       {method: "POST", path: "/api/v2/admin/workflow-tasks/:id/complete"},
	"workflow:task:delete":         {method: "DELETE", path: "/api/v2/admin/workflow-tasks/:id"},
	"workflow:notification:list":   {method: "GET", path: "/api/v2/admin/workflow-notifications"},
	"workflow:notification:retry":  {method: "POST", path: "/api/v2/admin/workflow-notifications/:id/retry"},
	"workflow:org-approver:list":   {method: "GET", path: "/api/v2/admin/workflow-org-approver-identities"},
	"workflow:org-approver:edit":   {method: "PUT", path: "/api/v2/admin/workflow-org-approver-assignments"},
	"scheduled-task:list":          {method: "GET", path: "/api/v2/admin/scheduled-tasks"},
	"scheduled-task:add":           {method: "POST", path: "/api/v2/admin/scheduled-tasks"},
	"scheduled-task:edit":          {method: "PUT", path: "/api/v2/admin/scheduled-tasks/:id"},
	"scheduled-task:delete":        {method: "DELETE", path: "/api/v2/admin/scheduled-tasks/:id"},
	"scheduled-task:status":        {method: "PATCH", path: "/api/v2/admin/scheduled-tasks/:id/status"},
	"scheduled-task:run":           {method: "POST", path: "/api/v2/admin/scheduled-tasks/:id/run"},
	"scheduled-task:run:list":      {method: "GET", path: "/api/v2/admin/scheduled-task-runs"},
	"scheduled-task:run:retry":     {method: "POST", path: "/api/v2/admin/scheduled-task-runs/:id/retry"},
	"scheduled-task:run:cancel":    {method: "POST", path: "/api/v2/admin/scheduled-task-runs/:id/cancel"},
	"scheduled-task:worker:list":   {method: "GET", path: "/api/v2/admin/scheduled-task-workers"},
	"scheduled-task:http":          {method: "POST", path: "/api/v2/admin/scheduled-tasks"},
	"scheduled-task:shell":         {method: "POST", path: "/api/v2/admin/scheduled-tasks"},
	"scheduled-task:sql:read":      {method: "POST", path: "/api/v2/admin/scheduled-tasks"},
	"scheduled-task:sql:write":     {method: "POST", path: "/api/v2/admin/scheduled-tasks"},
	"notification:list":            {method: "GET", path: "/api/v2/admin/in-app-notifications"},
	"notification:read":            {method: "PATCH", path: "/api/v2/admin/in-app-notifications/:id/read"},
	"notification:send":            {method: "POST", path: "/api/v2/admin/in-app-notifications"},
	"notification:dingtalk:send":   {method: "POST", path: "/api/v2/admin/dingtalk-notifications"},
}
