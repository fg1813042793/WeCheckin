package appapiperm

type Category struct {
	Key      string
	Name     string
	Platform string
	Sort     int
}

type Declaration struct {
	Key          string
	Name         string
	Platform     string
	Perms        string
	CategoryKey  string
	CategoryName string
	Method       string
	Path         string
	Sort         int
}

type RouteDeclaration struct {
	Method        string
	Path          string
	PermissionKey string
}

func ClientAPICategories() []Category {
	return []Category{
		{Key: "client:api-category:user", Name: "用户与会话", Platform: "client", Sort: 10},
		{Key: "client:api-category:favorite", Name: "收藏", Platform: "client", Sort: 20},
		{Key: "client:api-category:news", Name: "通知", Platform: "client", Sort: 30},
		{Key: "client:api-category:enroll", Name: "打卡任务", Platform: "client", Sort: 40},
		{Key: "client:api-category:event", Name: "赛事活动", Platform: "client", Sort: 50},
		{Key: "client:api-category:survey", Name: "问卷", Platform: "client", Sort: 60},
		{Key: "client:api-category:exam", Name: "考试", Platform: "client", Sort: 70},
		{Key: "client:api-category:workflow", Name: "OA 流程", Platform: "client", Sort: 80},
	}
}

func ClientAPIDeclarations() []Declaration {
	return []Declaration{
		clientAPI("client:api:bootstrap:view", "应用启动接口", "bootstrap:view", "client:api-category:user", "GET", "/api/v2/me/bootstrap", 5),
		clientAPI("client:api:user:view", "用户资料查看接口", "user:view", "client:api-category:user", "GET", "/api/v2/me", 10),
		clientAPI("client:api:user:edit", "用户资料编辑接口", "user:edit", "client:api-category:user", "PUT", "/api/v2/me", 20),
		clientAPI("client:api:user:phone", "手机号授权接口", "user:phone", "client:api-category:user", "POST", "/api/v2/me/phone", 30),
		clientAPI("client:api:user:logout", "用户退出接口", "user:logout", "client:api-category:user", "POST", "/api/v2/me/logout", 40),
		clientAPI("client:api:favorite:view", "收藏查看接口", "favorite:view", "client:api-category:favorite", "GET", "/api/v2/me/favorites", 10),
		clientAPI("client:api:favorite:edit", "收藏维护接口", "favorite:edit", "client:api-category:favorite", "POST", "/api/v2/me/favorites", 20),
		clientAPI("client:api:news:view", "通知查看接口", "news:view", "client:api-category:news", "GET", "/api/v2/news", 10),
		clientAPI("client:api:enroll:view", "打卡查看接口", "enroll:view", "client:api-category:enroll", "GET", "/api/v2/enrollments", 10),
		clientAPI("client:api:enroll:join", "打卡参与接口", "enroll:join", "client:api-category:enroll", "POST", "/api/v2/enrollments/:id/joins", 20),
		clientAPI("client:api:enroll:submit", "打卡提交接口", "enroll:submit", "client:api-category:enroll", "POST", "/api/v2/enrollments/:id/submissions", 30),
		clientAPI("client:api:event:view", "赛事活动查看接口", "event:view", "client:api-category:event", "GET", "/api/v2/me/events", 10),
		clientAPI("client:api:event:join", "赛事活动参与接口", "event:join", "client:api-category:event", "POST", "/api/v2/events/:id/participants", 20),
		clientAPI("client:api:event:dynamic", "赛事活动动态接口", "event:dynamic", "client:api-category:event", "GET", "/api/v2/events/:id/dynamics", 30),
		clientAPI("client:api:event:score", "赛事活动成绩接口", "event:score", "client:api-category:event", "POST", "/api/v2/events/:id/scores", 40),
		clientAPI("client:api:survey:view", "问卷查看接口", "survey:view", "client:api-category:survey", "GET", "/api/v2/me/survey-responses", 10),
		clientAPI("client:api:survey:response", "问卷答卷接口", "survey:response", "client:api-category:survey", "GET", "/api/v2/me/survey-responses/:id", 20),
		clientAPI("client:api:exam:view", "考试查看接口", "exam:view", "client:api-category:exam", "GET", "/api/v2/me/exam-records", 10),
		clientAPI("client:api:exam:start", "考试开始接口", "exam:start", "client:api-category:exam", "POST", "/api/v2/exams/:id/start", 20),
		clientAPI("client:api:exam:answer", "考试答题接口", "exam:answer", "client:api-category:exam", "PUT", "/api/v2/exam-records/:id/answers", 30),
		clientAPI("client:api:workflow:view", "OA 流程查看接口", "workflow:view", "client:api-category:workflow", "GET", "/api/v2/workflows/instances", 10),
		clientAPI("client:api:workflow:start", "OA 流程发起接口", "workflow:start", "client:api-category:workflow", "POST", "/api/v2/workflows/instances", 20),
		clientAPI("client:api:workflow:handle", "OA 流程处理接口", "workflow:handle", "client:api-category:workflow", "POST", "/api/v2/workflows/tasks/:id/complete", 30),
		clientAPI("client:api:workflow:withdraw", "OA 流程撤回接口", "workflow:withdraw", "client:api-category:workflow", "POST", "/api/v2/workflows/instances/:id/withdraw", 40),
	}
}

func ClientRouteDeclarations() []RouteDeclaration {
	return []RouteDeclaration{
		{Method: "GET", Path: "/api/v2/me/bootstrap", PermissionKey: "client:api:bootstrap:view"},
		{Method: "GET", Path: "/api/v2/me", PermissionKey: "client:api:user:view"},
		{Method: "PUT", Path: "/api/v2/me", PermissionKey: "client:api:user:edit"},
		{Method: "POST", Path: "/api/v2/me/phone", PermissionKey: "client:api:user:phone"},
		{Method: "POST", Path: "/api/v2/me/logout", PermissionKey: "client:api:user:logout"},
		{Method: "GET", Path: "/api/v2/me/favorites", PermissionKey: "client:api:favorite:view"},
		{Method: "GET", Path: "/api/v2/me/favorites/check", PermissionKey: "client:api:favorite:view"},
		{Method: "POST", Path: "/api/v2/me/favorites", PermissionKey: "client:api:favorite:edit"},
		{Method: "DELETE", Path: "/api/v2/me/favorites/:oid", PermissionKey: "client:api:favorite:edit"},
		{Method: "GET", Path: "/api/v2/news", PermissionKey: "client:api:news:view"},
		{Method: "GET", Path: "/api/v2/news/categories", PermissionKey: "client:api:news:view"},
		{Method: "GET", Path: "/api/v2/news/:id", PermissionKey: "client:api:news:view"},
		{Method: "GET", Path: "/api/v2/enrollments", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/enrollments/:id", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/enrollments/:id/join-days", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/me/enrollments", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/me/enrollment-users", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/me/enrollment-records", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/me/enrollment-calendar", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/api/v2/me/enrollment-day-records", PermissionKey: "client:api:enroll:view"},
		{Method: "POST", Path: "/api/v2/enrollments/:id/joins", PermissionKey: "client:api:enroll:join"},
		{Method: "POST", Path: "/api/v2/enrollments/:id/submissions", PermissionKey: "client:api:enroll:submit"},
		{Method: "GET", Path: "/api/v2/me/events", PermissionKey: "client:api:event:view"},
		{Method: "GET", Path: "/api/v2/me/event-roles", PermissionKey: "client:api:event:view"},
		{Method: "GET", Path: "/api/v2/me/managed-events", PermissionKey: "client:api:event:view"},
		{Method: "GET", Path: "/api/v2/events/:id/participants", PermissionKey: "client:api:event:view"},
		{Method: "POST", Path: "/api/v2/events/:id/participants", PermissionKey: "client:api:event:join"},
		{Method: "GET", Path: "/api/v2/events/:id/dynamics", PermissionKey: "client:api:event:dynamic"},
		{Method: "POST", Path: "/api/v2/events/:id/dynamics", PermissionKey: "client:api:event:dynamic"},
		{Method: "GET", Path: "/api/v2/events/:id/scores", PermissionKey: "client:api:event:score"},
		{Method: "POST", Path: "/api/v2/events/:id/scores", PermissionKey: "client:api:event:score"},
		{Method: "GET", Path: "/api/v2/me/survey-responses", PermissionKey: "client:api:survey:view"},
		{Method: "GET", Path: "/api/v2/me/survey-responses/:id", PermissionKey: "client:api:survey:response"},
		{Method: "GET", Path: "/api/v2/me/exam-records", PermissionKey: "client:api:exam:view"},
		{Method: "POST", Path: "/api/v2/exams/:id/start", PermissionKey: "client:api:exam:start"},
		{Method: "GET", Path: "/api/v2/exam-records/:id", PermissionKey: "client:api:exam:view"},
		{Method: "PUT", Path: "/api/v2/exam-records/:id/answers", PermissionKey: "client:api:exam:answer"},
		{Method: "GET", Path: "/api/v2/workflows/definitions", PermissionKey: "client:api:workflow:view"},
		{Method: "GET", Path: "/api/v2/workflows/definitions/:id", PermissionKey: "client:api:workflow:view"},
		{Method: "GET", Path: "/api/v2/workflows/drafts/:definitionId", PermissionKey: "client:api:workflow:start"},
		{Method: "PUT", Path: "/api/v2/workflows/drafts/:definitionId", PermissionKey: "client:api:workflow:start"},
		{Method: "POST", Path: "/api/v2/workflows/instances", PermissionKey: "client:api:workflow:start"},
		{Method: "GET", Path: "/api/v2/workflows/instances", PermissionKey: "client:api:workflow:view"},
		{Method: "GET", Path: "/api/v2/workflows/instances/:id", PermissionKey: "client:api:workflow:view"},
		{Method: "POST", Path: "/api/v2/workflows/instances/:id/withdraw", PermissionKey: "client:api:workflow:withdraw"},
		{Method: "GET", Path: "/api/v2/workflows/tasks", PermissionKey: "client:api:workflow:view"},
		{Method: "POST", Path: "/api/v2/workflows/tasks/:id/complete", PermissionKey: "client:api:workflow:handle"},
		{Method: "GET", Path: "/passport/my_detail", PermissionKey: "client:api:user:view"},
		{Method: "POST", Path: "/passport/edit_base", PermissionKey: "client:api:user:edit"},
		{Method: "POST", Path: "/passport/phone", PermissionKey: "client:api:user:phone"},
		{Method: "POST", Path: "/passport/logout", PermissionKey: "client:api:user:logout"},
		{Method: "GET", Path: "/fav/my_list", PermissionKey: "client:api:favorite:view"},
		{Method: "GET", Path: "/fav/is_fav", PermissionKey: "client:api:favorite:view"},
		{Method: "POST", Path: "/fav/update", PermissionKey: "client:api:favorite:edit"},
		{Method: "POST", Path: "/fav/del", PermissionKey: "client:api:favorite:edit"},
		{Method: "GET", Path: "/news/list", PermissionKey: "client:api:news:view"},
		{Method: "GET", Path: "/news/view", PermissionKey: "client:api:news:view"},
		{Method: "GET", Path: "/news/cate_list", PermissionKey: "client:api:news:view"},
		{Method: "GET", Path: "/enroll/list", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/view", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/join_day", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/my_join_list", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/my_user_list", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/my_records", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/my_calendar", PermissionKey: "client:api:enroll:view"},
		{Method: "GET", Path: "/enroll/my_day_records", PermissionKey: "client:api:enroll:view"},
		{Method: "POST", Path: "/enroll/join", PermissionKey: "client:api:enroll:join"},
		{Method: "POST", Path: "/enroll/enroll_submit", PermissionKey: "client:api:enroll:submit"},
		{Method: "GET", Path: "/event/my_list", PermissionKey: "client:api:event:view"},
		{Method: "GET", Path: "/event/my_roles", PermissionKey: "client:api:event:view"},
		{Method: "GET", Path: "/event/my_managed", PermissionKey: "client:api:event:view"},
		{Method: "GET", Path: "/event/participant_list", PermissionKey: "client:api:event:view"},
		{Method: "POST", Path: "/event/participate", PermissionKey: "client:api:event:join"},
		{Method: "GET", Path: "/event/dynamics", PermissionKey: "client:api:event:dynamic"},
		{Method: "POST", Path: "/event/dynamic_post", PermissionKey: "client:api:event:dynamic"},
		{Method: "GET", Path: "/event/scores", PermissionKey: "client:api:event:score"},
		{Method: "POST", Path: "/event/score_save", PermissionKey: "client:api:event:score"},
		{Method: "GET", Path: "/survey/my_responses", PermissionKey: "client:api:survey:view"},
		{Method: "GET", Path: "/survey/my_response", PermissionKey: "client:api:survey:response"},
		{Method: "GET", Path: "/exam/my_records", PermissionKey: "client:api:exam:view"},
		{Method: "GET", Path: "/exam/record", PermissionKey: "client:api:exam:view"},
		{Method: "GET", Path: "/exam/start", PermissionKey: "client:api:exam:start"},
		{Method: "POST", Path: "/exam/save_answer", PermissionKey: "client:api:exam:answer"},
	}
}

func DingTalkH5APICategories() []Category {
	return []Category{
		{Key: "dingtalk_h5:api-category:session", Name: "登录会话", Platform: "dingtalk_h5", Sort: 10},
		{Key: "dingtalk_h5:api-category:workbench", Name: "工作台", Platform: "dingtalk_h5", Sort: 20},
		{Key: "dingtalk_h5:api-category:review", Name: "绩效考评", Platform: "dingtalk_h5", Sort: 30},
		{Key: "dingtalk_h5:api-category:flow", Name: "流程操作", Platform: "dingtalk_h5", Sort: 40},
		{Key: "dingtalk_h5:api-category:user", Name: "人员维护", Platform: "dingtalk_h5", Sort: 50},
		{Key: "dingtalk_h5:api-category:template", Name: "绩效模版", Platform: "dingtalk_h5", Sort: 60},
		{Key: "dingtalk_h5:api-category:workflow", Name: "OA 流程", Platform: "dingtalk_h5", Sort: 70},
	}
}

func DingTalkH5APIDeclarations() []Declaration {
	return []Declaration{
		dingtalkAPI("dingtalk_h5:api:session:logout", "退出登录接口", "session:logout", "dingtalk_h5:api-category:session", "POST", "/api/v2/dingtalk/h5/logout", 10),
		dingtalkAPI("dingtalk_h5:api:account:password", "修改密码接口", "account:password", "dingtalk_h5:api-category:session", "PATCH", "/api/v2/dingtalk/h5/account/password", 20),
		dingtalkAPI("dingtalk_h5:api:bootstrap:view", "应用启动接口", "bootstrap:view", "dingtalk_h5:api-category:workbench", "GET", "/api/v2/dingtalk/h5/bootstrap", 10),
		dingtalkAPI("dingtalk_h5:api:workbench:view", "工作台统计接口", "workbench:view", "dingtalk_h5:api-category:workbench", "GET", "/api/v2/dingtalk/h5/workbench", 20),
		dingtalkAPI("dingtalk_h5:api:review:list", "绩效列表接口", "review:list", "dingtalk_h5:api-category:review", "GET", "/api/v2/dingtalk/h5/reviews", 10),
		dingtalkAPI("dingtalk_h5:api:review:export", "导出绩效接口", "review:export", "dingtalk_h5:api-category:review", "GET", "/api/v2/dingtalk/h5/reviews/export", 15),
		dingtalkAPI("dingtalk_h5:api:review:detail", "绩效详情接口", "review:detail", "dingtalk_h5:api-category:review", "GET", "/api/v2/dingtalk/h5/reviews/:id", 20),
		dingtalkAPI("dingtalk_h5:api:review:create", "创建绩效接口", "review:create", "dingtalk_h5:api-category:review", "POST", "/api/v2/dingtalk/h5/reviews", 30),
		dingtalkAPI("dingtalk_h5:api:review:delete", "删除绩效接口", "review:delete", "dingtalk_h5:api-category:review", "DELETE", "/api/v2/dingtalk/h5/reviews/:id", 40),
		dingtalkAPI("dingtalk_h5:api:review:self_save", "员工暂存接口", "review:self_save", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/save-self", 10),
		dingtalkAPI("dingtalk_h5:api:review:self_submit", "员工提交接口", "review:self_submit", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/submit-self", 20),
		dingtalkAPI("dingtalk_h5:api:review:manager_submit", "上级评价接口", "review:manager_submit", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/submit-manager", 30),
		dingtalkAPI("dingtalk_h5:api:review:hrbp_submit", "HRBP评价接口", "review:hrbp_submit", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/submit-hrbp", 40),
		dingtalkAPI("dingtalk_h5:api:review:confirm", "员工确认接口", "review:confirm", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/confirm-result", 50),
		dingtalkAPI("dingtalk_h5:api:review:dispute", "员工异议接口", "review:dispute", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/dispute-result", 60),
		dingtalkAPI("dingtalk_h5:api:review:withdraw", "撤回绩效接口", "review:withdraw", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/withdraw", 70),
		dingtalkAPI("dingtalk_h5:api:review:return_employee", "退回员工接口", "review:return_employee", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/return-employee", 80),
		dingtalkAPI("dingtalk_h5:api:review:return_manager", "退回上级接口", "review:return_manager", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/return-manager", 90),
		dingtalkAPI("dingtalk_h5:api:review:return_hrbp", "退回 HRBP 接口", "review:return_hrbp", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/return-hrbp", 100),
		dingtalkAPI("dingtalk_h5:api:review:finalize", "归档绩效接口", "review:finalize", "dingtalk_h5:api-category:flow", "POST", "/api/v2/dingtalk/h5/reviews/:id/finalize", 110),
		dingtalkAPI("dingtalk_h5:api:user:list", "人员查看接口", "user:list", "dingtalk_h5:api-category:user", "GET", "/api/v2/dingtalk/h5/users", 10),
		dingtalkAPI("dingtalk_h5:api:user:add", "人员创建接口", "user:add", "dingtalk_h5:api-category:user", "POST", "/api/v2/dingtalk/h5/users", 20),
		dingtalkAPI("dingtalk_h5:api:user:edit", "人员编辑接口", "user:edit", "dingtalk_h5:api-category:user", "PUT", "/api/v2/dingtalk/h5/users/:id", 30),
		dingtalkAPI("dingtalk_h5:api:user:delete", "人员删除接口", "user:delete", "dingtalk_h5:api-category:user", "DELETE", "/api/v2/dingtalk/h5/users/:id", 40),
		dingtalkAPI("dingtalk_h5:api:template:view", "绩效模版查看接口", "template:view", "dingtalk_h5:api-category:template", "GET", "/api/v2/dingtalk/h5/template", 10),
		dingtalkAPI("dingtalk_h5:api:template:save", "绩效模版保存接口", "template:save", "dingtalk_h5:api-category:template", "PUT", "/api/v2/dingtalk/h5/template", 20),
		dingtalkAPI("dingtalk_h5:api:workflow:view", "OA 流程查看接口", "workflow:view", "dingtalk_h5:api-category:workflow", "GET", "/api/v2/dingtalk/h5/workflows/instances", 10),
		dingtalkAPI("dingtalk_h5:api:workflow:start", "OA 流程发起接口", "workflow:start", "dingtalk_h5:api-category:workflow", "POST", "/api/v2/dingtalk/h5/workflows/instances", 20),
		dingtalkAPI("dingtalk_h5:api:workflow:handle", "OA 流程处理接口", "workflow:handle", "dingtalk_h5:api-category:workflow", "POST", "/api/v2/dingtalk/h5/workflows/tasks/:id/complete", 30),
		dingtalkAPI("dingtalk_h5:api:workflow:withdraw", "OA 流程撤回接口", "workflow:withdraw", "dingtalk_h5:api-category:workflow", "POST", "/api/v2/dingtalk/h5/workflows/instances/:id/withdraw", 40),
	}
}

func DingTalkH5RouteDeclarations() []RouteDeclaration {
	routes := make([]RouteDeclaration, 0, len(DingTalkH5APIDeclarations())+6)
	for _, declaration := range DingTalkH5APIDeclarations() {
		routes = append(routes, RouteDeclaration{
			Method:        declaration.Method,
			Path:          declaration.Path,
			PermissionKey: declaration.Key,
		})
	}
	routes = append(routes,
		RouteDeclaration{Method: "GET", Path: "/api/v2/dingtalk/h5/workflows/definitions", PermissionKey: "dingtalk_h5:api:workflow:view"},
		RouteDeclaration{Method: "GET", Path: "/api/v2/dingtalk/h5/workflows/definitions/:id", PermissionKey: "dingtalk_h5:api:workflow:view"},
		RouteDeclaration{Method: "GET", Path: "/api/v2/dingtalk/h5/workflows/drafts/:definitionId", PermissionKey: "dingtalk_h5:api:workflow:start"},
		RouteDeclaration{Method: "PUT", Path: "/api/v2/dingtalk/h5/workflows/drafts/:definitionId", PermissionKey: "dingtalk_h5:api:workflow:start"},
		RouteDeclaration{Method: "GET", Path: "/api/v2/dingtalk/h5/workflows/instances/:id", PermissionKey: "dingtalk_h5:api:workflow:view"},
		RouteDeclaration{Method: "GET", Path: "/api/v2/dingtalk/h5/workflows/tasks", PermissionKey: "dingtalk_h5:api:workflow:view"},
	)
	return routes
}

func clientAPI(key, name, perms, categoryKey, method, path string, sort int) Declaration {
	return Declaration{Key: key, Name: name, Platform: "client", Perms: perms, CategoryKey: categoryKey, CategoryName: categoryName(ClientAPICategories(), categoryKey), Method: method, Path: path, Sort: sort}
}

func dingtalkAPI(key, name, perms, categoryKey, method, path string, sort int) Declaration {
	return Declaration{Key: key, Name: name, Platform: "dingtalk_h5", Perms: perms, CategoryKey: categoryKey, CategoryName: categoryName(DingTalkH5APICategories(), categoryKey), Method: method, Path: path, Sort: sort}
}

func categoryName(categories []Category, key string) string {
	for _, category := range categories {
		if category.Key == key {
			return category.Name
		}
	}
	return ""
}
