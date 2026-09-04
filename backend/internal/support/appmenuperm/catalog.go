package appmenuperm

const (
	TypeDirectory = "directory"
	TypeMenu      = "menu"
)

type Declaration struct {
	Key       string
	Name      string
	Platform  string
	Type      string
	Path      string
	Icon      string
	ParentKey string
	Sort      int
}

func ClientMenuDeclarations() []Declaration {
	return []Declaration{
		{Key: "client:menu:home", Name: "首页", Platform: "client", Path: "/pages/index/index", Sort: 10},
		{Key: "client:menu:news", Name: "通知", Platform: "client", Path: "/pages/news/news_index", Sort: 20},
		{Key: "client:menu:enroll", Name: "打卡任务", Platform: "client", Path: "/pages/enroll/enroll_index", Sort: 30},
		{Key: "client:menu:my_checkin", Name: "我的打卡", Platform: "client", Path: "/pages/enroll/my_user_list", Sort: 40},
		{Key: "client:menu:survey", Name: "问卷中心", Platform: "client", Path: "/pages/survey/index", Sort: 50},
		{Key: "client:menu:exam", Name: "考试列表", Platform: "client", Path: "/pages/exam/index", Sort: 60},
		{Key: "client:menu:event", Name: "赛事活动", Platform: "client", Path: "/pages/event/event_index", Sort: 70},
		{Key: "client:menu:my_activity", Name: "我的活动", Platform: "client", Path: "/pages/my/my_activity", Sort: 80},
		{Key: "client:menu:my_competition", Name: "我的赛事", Platform: "client", Path: "/pages/my/my_competition", Sort: 90},
		{Key: "client:menu:event_manage", Name: "赛事管理", Platform: "client", Path: "/pages/event/my_event_manage", Sort: 100},
		{Key: "client:menu:favorite", Name: "我的收藏", Platform: "client", Path: "/pages/my/my_fav", Sort: 105},
		{Key: "client:menu:profile", Name: "个人中心", Platform: "client", Path: "/pages/my/my_index", Sort: 110},
	}
}

func DingTalkH5MenuDeclarations() []Declaration {
	return []Declaration{
		{Key: "dingtalk_h5:menu:dashboard", Name: "工作台", Platform: "dingtalk_h5", Path: "dashboard", Icon: "dashboard", Sort: 10},
		{Key: "dingtalk_h5:menu:performance", Name: "绩效管理", Platform: "dingtalk_h5", Type: TypeDirectory, Path: "performance", Icon: "performance", Sort: 20},
		{Key: "dingtalk_h5:menu:performance:mine", Name: "我的绩效", Platform: "dingtalk_h5", Path: "performance:mine", Icon: "mine", ParentKey: "dingtalk_h5:menu:performance", Sort: 30},
		{Key: "dingtalk_h5:menu:performance:history", Name: "历史绩效", Platform: "dingtalk_h5", Path: "performance:history", Icon: "history", ParentKey: "dingtalk_h5:menu:performance", Sort: 40},
		{Key: "dingtalk_h5:menu:performance:manager", Name: "上级评价", Platform: "dingtalk_h5", Path: "performance:manager", Icon: "manager", ParentKey: "dingtalk_h5:menu:performance", Sort: 50},
		{Key: "dingtalk_h5:menu:performance:hrbp", Name: "HRBP评价", Platform: "dingtalk_h5", Path: "performance:hrbp", Icon: "hrbp", ParentKey: "dingtalk_h5:menu:performance", Sort: 60},
		{Key: "dingtalk_h5:menu:performance:summary", Name: "HRBP汇总", Platform: "dingtalk_h5", Path: "performance:summary", Icon: "summary", ParentKey: "dingtalk_h5:menu:performance", Sort: 70},
		{Key: "dingtalk_h5:menu:performance:org", Name: "流程执行", Platform: "dingtalk_h5", Path: "performance:org", Icon: "org", ParentKey: "dingtalk_h5:menu:performance", Sort: 80},
		{Key: "dingtalk_h5:menu:performance:template", Name: "绩效模版", Platform: "dingtalk_h5", Path: "performance:template", Icon: "template", ParentKey: "dingtalk_h5:menu:performance", Sort: 90},
		{Key: "dingtalk_h5:menu:workflow", Name: "流程审批", Platform: "dingtalk_h5", Path: "workflow", Icon: "workflow", Sort: 100},
	}
}

func DingTalkH5ButtonDeclarations() []Declaration {
	return []Declaration{
		{Key: "dingtalk_h5:button:workflow:summary", Name: "流程汇总", Platform: "dingtalk_h5", Path: "workflow:summary", ParentKey: "dingtalk_h5:menu:workflow", Sort: 101},
		{Key: "dingtalk_h5:button:review:create", Name: "创建考评单", Platform: "dingtalk_h5", Path: "review:create", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 31},
		{Key: "dingtalk_h5:button:review:self_save", Name: "保存员工自评", Platform: "dingtalk_h5", Path: "review:self_save", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 32},
		{Key: "dingtalk_h5:button:review:self_submit", Name: "提交员工自评", Platform: "dingtalk_h5", Path: "review:self_submit", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 33},
		{Key: "dingtalk_h5:button:review:next_objective_edit", Name: "编辑下月目标", Platform: "dingtalk_h5", Path: "review:next_objective_edit", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 34},
		{Key: "dingtalk_h5:button:review:next_objective_add", Name: "新增下月目标", Platform: "dingtalk_h5", Path: "review:next_objective_add", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 35},
		{Key: "dingtalk_h5:button:review:next_objective_delete", Name: "删除下月目标", Platform: "dingtalk_h5", Path: "review:next_objective_delete", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 36},
		{Key: "dingtalk_h5:button:review:confirm", Name: "确认绩效结果", Platform: "dingtalk_h5", Path: "review:confirm", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 41},
		{Key: "dingtalk_h5:button:review:dispute", Name: "提交绩效异议", Platform: "dingtalk_h5", Path: "review:dispute", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 42},
		{Key: "dingtalk_h5:button:review:withdraw", Name: "撤销提交", Platform: "dingtalk_h5", Path: "review:withdraw", ParentKey: "dingtalk_h5:menu:performance:mine", Sort: 43},
		{Key: "dingtalk_h5:button:review:manager_submit", Name: "提交上级评价", Platform: "dingtalk_h5", Path: "review:manager_submit", ParentKey: "dingtalk_h5:menu:performance:manager", Sort: 51},
		{Key: "dingtalk_h5:button:review:return_employee", Name: "退回员工修改", Platform: "dingtalk_h5", Path: "review:return_employee", ParentKey: "dingtalk_h5:menu:performance:manager", Sort: 52},
		{Key: "dingtalk_h5:button:review:hrbp_submit", Name: "提交 HRBP 评价", Platform: "dingtalk_h5", Path: "review:hrbp_submit", ParentKey: "dingtalk_h5:menu:performance:hrbp", Sort: 61},
		{Key: "dingtalk_h5:button:review:return_manager", Name: "退回上级修改", Platform: "dingtalk_h5", Path: "review:return_manager", ParentKey: "dingtalk_h5:menu:performance:hrbp", Sort: 62},
		{Key: "dingtalk_h5:button:review:return_hrbp", Name: "退回 HRBP 修改", Platform: "dingtalk_h5", Path: "review:return_hrbp", ParentKey: "dingtalk_h5:menu:performance:hrbp", Sort: 63},
		{Key: "dingtalk_h5:button:review:finalize", Name: "归档绩效", Platform: "dingtalk_h5", Path: "review:finalize", ParentKey: "dingtalk_h5:menu:performance:hrbp", Sort: 64},
		{Key: "dingtalk_h5:button:review:export", Name: "导出绩效", Platform: "dingtalk_h5", Path: "review:export", ParentKey: "dingtalk_h5:menu:performance:summary", Sort: 71},
		{Key: "dingtalk_h5:button:review:delete", Name: "删除考评单", Platform: "dingtalk_h5", Path: "review:delete", ParentKey: "dingtalk_h5:menu:performance:summary", Sort: 72},
		{Key: "dingtalk_h5:button:user:config", Name: "配置流程审批人", Platform: "dingtalk_h5", Path: "user:config", ParentKey: "dingtalk_h5:menu:performance:org", Sort: 81},
		{Key: "dingtalk_h5:button:template:edit", Name: "编辑绩效模版", Platform: "dingtalk_h5", Path: "template:edit", ParentKey: "dingtalk_h5:menu:performance:template", Sort: 91},
	}
}

func DingTalkH5PermissionDeclarations() []Declaration {
	menus := DingTalkH5MenuDeclarations()
	buttons := DingTalkH5ButtonDeclarations()
	result := make([]Declaration, 0, len(menus)+len(buttons))
	result = append(result, menus...)
	result = append(result, buttons...)
	return result
}
