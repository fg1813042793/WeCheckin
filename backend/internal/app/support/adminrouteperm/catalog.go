package adminrouteperm

import "strings"

type Declaration struct {
	Key          string
	Name         string
	Perms        string
	CategoryKey  string
	CategoryName string
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
		{"news:list", "通知查看接口"},
		{"news:add", "通知创建接口"},
		{"news:edit", "通知编辑接口"},
		{"news:del", "通知删除接口"},
		{"mgr:list", "管理员查看接口"},
		{"mgr:add", "管理员创建接口"},
		{"mgr:edit", "管理员编辑接口"},
		{"mgr:del", "管理员删除接口"},
		{"setup:edit", "系统设置接口"},
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
		{"dept:list", "部门查看接口"},
		{"dept:add", "部门创建接口"},
		{"dept:edit", "部门编辑接口"},
		{"dept:del", "部门删除接口"},
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
	}
	out := make([]Declaration, 0, len(codes))
	for _, item := range codes {
		category := categoryForPerms(categories, item.perms)
		out = append(out, Declaration{
			Key:          KeyForPerms(item.perms),
			Name:         item.name,
			Perms:        item.perms,
			CategoryKey:  category.Key,
			CategoryName: category.Name,
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
		{Key: "admin:api-category:survey", Name: "问卷管理", Sort: 50},
		{Key: "admin:api-category:exam", Name: "考试管理", Sort: 60},
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
	case "user", "online":
		key = "admin:api-category:user"
	case "enroll", "event", "news":
		key = "admin:api-category:content"
	case "survey", "response", "question-bank":
		key = "admin:api-category:survey"
	case "exam":
		key = "admin:api-category:exam"
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
