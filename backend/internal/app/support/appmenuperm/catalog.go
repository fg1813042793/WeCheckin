package appmenuperm

type Declaration struct {
	Key       string
	Name      string
	Platform  string
	Path      string
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
		{Key: "client:menu:profile", Name: "个人中心", Platform: "client", Path: "/pages/my/my_index", Sort: 110},
	}
}

func DingTalkH5MenuDeclarations() []Declaration {
	return []Declaration{
		{Key: "dingtalk_h5:menu:dashboard", Name: "工作台", Platform: "dingtalk_h5", Path: "dashboard", Sort: 10},
		{Key: "dingtalk_h5:menu:performance", Name: "绩效管理", Platform: "dingtalk_h5", Path: "performance", Sort: 20},
		{Key: "dingtalk_h5:menu:performance:mine", Name: "本月绩效", Platform: "dingtalk_h5", Path: "performance:mine", ParentKey: "dingtalk_h5:menu:performance", Sort: 30},
		{Key: "dingtalk_h5:menu:performance:history", Name: "历史绩效", Platform: "dingtalk_h5", Path: "performance:history", ParentKey: "dingtalk_h5:menu:performance", Sort: 40},
		{Key: "dingtalk_h5:menu:performance:hrbp", Name: "HRBP评价", Platform: "dingtalk_h5", Path: "performance:hrbp", ParentKey: "dingtalk_h5:menu:performance", Sort: 50},
		{Key: "dingtalk_h5:menu:performance:summary", Name: "HRBP汇总", Platform: "dingtalk_h5", Path: "performance:summary", ParentKey: "dingtalk_h5:menu:performance", Sort: 60},
		{Key: "dingtalk_h5:menu:performance:org", Name: "流程执行", Platform: "dingtalk_h5", Path: "performance:org", ParentKey: "dingtalk_h5:menu:performance", Sort: 70},
		{Key: "dingtalk_h5:menu:performance:template", Name: "绩效模版", Platform: "dingtalk_h5", Path: "performance:template", ParentKey: "dingtalk_h5:menu:performance", Sort: 80},
	}
}
