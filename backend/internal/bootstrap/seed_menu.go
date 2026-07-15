package bootstrap

import (
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

const menuSeedInitializedKey = "SYSTEM_MENU_SEED_INITIALIZED"

type menuDef struct {
	Name   string
	Path   string
	Perms  string
	Icon   string
	Sort   int
	Type   int
	Parent string
}

func seedMenus(enableExam bool) {
	markerValue, err := getMenuSeedInitializedValue()
	if err != nil {
		log.Printf("check menu seed marker error: %v", err)
		return
	}

	var existingMenus int64
	if err := database.DB.Model(&model.Menu{}).Count(&existingMenus).Error; err != nil {
		log.Printf("check menu count error: %v", err)
		return
	}

	if !shouldSeedMenus(existingMenus, markerValue) {
		if !isMenuSeedInitialized(markerValue) {
			markMenuSeedInitialized()
		}
		return
	}

	defs := []menuDef{
		{Name: "控制台", Path: "/dashboard", Perms: "", Icon: "Odometer", Sort: 1, Type: 1},
		{Name: "用户管理", Path: "/user", Perms: "user:list", Icon: "User", Sort: 2, Type: 1},
		{Name: "在线用户", Path: "/online", Perms: "online:list", Icon: "Monitor", Sort: 3, Type: 1},
		{Name: "打卡管理", Path: "/enroll", Perms: "enroll:list", Icon: "List", Sort: 4, Type: 1},
		{Name: "内容管理", Path: "/news", Perms: "news:list", Icon: "Document", Sort: 5, Type: 1},
		{Name: "管理员管理", Path: "/mgr", Perms: "mgr:list", Icon: "Setting", Sort: 6, Type: 1},
		{Name: "操作日志", Path: "/log", Perms: "log:list", Icon: "Clock", Sort: 7, Type: 1},
		{Name: "字典管理", Path: "/dict", Perms: "dict:list", Icon: "Notebook", Sort: 8, Type: 1},
		{Name: "部门管理", Path: "/department", Perms: "dept:list", Icon: "FolderOpened", Sort: 9, Type: 1},
		{Name: "角色管理", Path: "/role", Perms: "role:list", Icon: "UserFilled", Sort: 10, Type: 1},
		{Name: "菜单权限", Path: "/menu", Perms: "menu:list", Icon: "Grid", Sort: 11, Type: 1},
		{Name: "系统配置", Path: "/setup", Perms: "setup:list,setup:edit", Icon: "Setting", Sort: 12, Type: 1},
		{Name: "赛事活动", Path: "/event", Perms: "event:list", Icon: "TrophyBase", Sort: 13, Type: 1},
		{Name: "问卷调查", Path: "/survey", Perms: "survey:list", Icon: "List", Sort: 14, Type: 0},
		{Name: "问卷考试", Path: "/question-exam", Perms: "question-bank:list", Icon: "Collection", Sort: 16, Type: 0},
		{Name: "问卷管理", Path: "/survey", Parent: "/survey", Sort: 1, Type: 1},
		{Name: "答卷管理", Path: "/survey/responses", Parent: "/survey", Sort: 2, Type: 1},
		{Name: "问卷统计", Path: "/survey/statistic", Parent: "/survey", Sort: 3, Type: 1},
		{Name: "题库管理", Path: "/question-bank", Parent: "/question-exam", Sort: 1, Type: 1},
		{Name: "用户列表", Perms: "user:list", Parent: "/user", Sort: 1, Type: 2},
		{Name: "用户新增", Perms: "user:add", Parent: "/user", Sort: 2, Type: 2},
		{Name: "用户编辑", Perms: "user:edit", Parent: "/user", Sort: 3, Type: 2},
		{Name: "用户删除", Perms: "user:del", Parent: "/user", Sort: 4, Type: 2},
		{Name: "用户审核", Perms: "user:status", Parent: "/user", Sort: 5, Type: 2},
		{Name: "在线用户列表", Perms: "online:list", Parent: "/online", Sort: 1, Type: 2},
		{Name: "强制下线", Perms: "online:force_offline", Parent: "/online", Sort: 2, Type: 2},
		{Name: "打卡列表", Perms: "enroll:list", Parent: "/enroll", Sort: 1, Type: 2},
		{Name: "打卡新增", Perms: "enroll:add", Parent: "/enroll", Sort: 2, Type: 2},
		{Name: "打卡编辑", Perms: "enroll:edit", Parent: "/enroll", Sort: 3, Type: 2},
		{Name: "打卡删除", Perms: "enroll:del", Parent: "/enroll", Sort: 4, Type: 2},
		{Name: "打卡状态管理", Perms: "enroll:status", Parent: "/enroll", Sort: 5, Type: 2},
		{Name: "打卡推荐置顶", Perms: "enroll:vouch", Parent: "/enroll", Sort: 6, Type: 2},
		{Name: "导出Excel", Perms: "enroll:export", Parent: "/enroll", Sort: 7, Type: 2},
		{Name: "查看参与用户", Perms: "enroll:users", Parent: "/enroll", Sort: 8, Type: 2},
		{Name: "内容列表", Perms: "news:list", Parent: "/news", Sort: 1, Type: 2},
		{Name: "内容新增", Perms: "news:add", Parent: "/news", Sort: 2, Type: 2},
		{Name: "内容编辑", Perms: "news:edit", Parent: "/news", Sort: 3, Type: 2},
		{Name: "内容删除", Perms: "news:del", Parent: "/news", Sort: 4, Type: 2},
		{Name: "内容停用启用", Perms: "news:status", Parent: "/news", Sort: 5, Type: 2},
		{Name: "内容置顶", Perms: "news:vouch", Parent: "/news", Sort: 6, Type: 2},
		{Name: "管理员列表", Perms: "mgr:list", Parent: "/mgr", Sort: 1, Type: 2},
		{Name: "管理员新增", Perms: "mgr:add", Parent: "/mgr", Sort: 2, Type: 2},
		{Name: "管理员编辑", Perms: "mgr:edit", Parent: "/mgr", Sort: 3, Type: 2},
		{Name: "管理员删除", Perms: "mgr:del", Parent: "/mgr", Sort: 4, Type: 2},
		{Name: "日志列表", Perms: "log:list", Parent: "/log", Sort: 1, Type: 2},
		{Name: "日志清空", Perms: "log:del", Parent: "/log", Sort: 2, Type: 2},
		{Name: "字典列表", Perms: "dict:list", Parent: "/dict", Sort: 1, Type: 2},
		{Name: "字典新增", Perms: "dict:add", Parent: "/dict", Sort: 2, Type: 2},
		{Name: "字典编辑", Perms: "dict:edit", Parent: "/dict", Sort: 3, Type: 2},
		{Name: "字典删除", Perms: "dict:del", Parent: "/dict", Sort: 4, Type: 2},
		{Name: "部门列表", Perms: "dept:list", Parent: "/department", Sort: 1, Type: 2},
		{Name: "部门新增", Perms: "dept:add", Parent: "/department", Sort: 2, Type: 2},
		{Name: "部门编辑", Perms: "dept:edit", Parent: "/department", Sort: 3, Type: 2},
		{Name: "部门删除", Perms: "dept:del", Parent: "/department", Sort: 4, Type: 2},
		{Name: "角色列表", Perms: "role:list", Parent: "/role", Sort: 1, Type: 2},
		{Name: "角色新增", Perms: "role:add", Parent: "/role", Sort: 2, Type: 2},
		{Name: "角色编辑", Perms: "role:edit", Parent: "/role", Sort: 3, Type: 2},
		{Name: "角色删除", Perms: "role:del", Parent: "/role", Sort: 4, Type: 2},
		{Name: "菜单列表", Perms: "menu:list", Parent: "/menu", Sort: 1, Type: 2},
		{Name: "菜单新增", Perms: "menu:add", Parent: "/menu", Sort: 2, Type: 2},
		{Name: "菜单编辑", Perms: "menu:edit", Parent: "/menu", Sort: 3, Type: 2},
		{Name: "菜单删除", Perms: "menu:del", Parent: "/menu", Sort: 4, Type: 2},
		{Name: "赛事活动列表", Perms: "event:list", Parent: "/event", Sort: 1, Type: 2},
		{Name: "赛事活动新增", Perms: "event:add", Parent: "/event", Sort: 2, Type: 2},
		{Name: "赛事活动编辑", Perms: "event:edit", Parent: "/event", Sort: 3, Type: 2},
		{Name: "赛事活动删除", Perms: "event:del", Parent: "/event", Sort: 4, Type: 2},
		{Name: "开始结束", Perms: "event:status", Parent: "/event", Sort: 5, Type: 2},
		{Name: "推荐", Perms: "event:vouch", Parent: "/event", Sort: 6, Type: 2},
		{Name: "置顶", Perms: "event:top", Parent: "/event", Sort: 7, Type: 2},
		{Name: "参与用户", Perms: "event:users", Parent: "/event", Sort: 8, Type: 2},
		{Name: "问卷列表", Perms: "survey:list", Parent: "/survey", Sort: 1, Type: 2},
		{Name: "问卷新增", Perms: "survey:add", Parent: "/survey", Sort: 2, Type: 2},
		{Name: "问卷编辑", Perms: "survey:edit", Parent: "/survey", Sort: 3, Type: 2},
		{Name: "问卷删除", Perms: "survey:del", Parent: "/survey", Sort: 4, Type: 2},
		{Name: "问卷状态管理", Perms: "survey:status", Parent: "/survey", Sort: 5, Type: 2},
		{Name: "复制问卷", Perms: "survey:copy", Parent: "/survey", Sort: 6, Type: 2},
		{Name: "答卷列表", Perms: "response:list", Parent: "/survey", Sort: 7, Type: 2},
		{Name: "答卷删除", Perms: "response:del", Parent: "/survey", Sort: 8, Type: 2},
		{Name: "导出答卷", Perms: "response:export", Parent: "/survey", Sort: 9, Type: 2},
		{Name: "题库列表", Perms: "question-bank:list", Parent: "/question-exam", Sort: 1, Type: 2},
		{Name: "题库新增", Perms: "question-bank:add", Parent: "/question-exam", Sort: 2, Type: 2},
		{Name: "题库编辑", Perms: "question-bank:edit", Parent: "/question-exam", Sort: 3, Type: 2},
		{Name: "题库删除", Perms: "question-bank:del", Parent: "/question-exam", Sort: 4, Type: 2},
	}
	if enableExam {
		examDefs := []menuDef{
			{Name: "在线考试", Path: "/exam", Perms: "exam:list,exam:add,exam:edit,exam:del", Icon: "EditPen", Sort: 15, Type: 0},
			{Name: "考试管理", Path: "/exam/list", Parent: "/exam", Sort: 1, Type: 1},
			{Name: "考试列表", Perms: "exam:list", Parent: "/exam", Sort: 1, Type: 2},
			{Name: "考试新增", Perms: "exam:add", Parent: "/exam", Sort: 2, Type: 2},
			{Name: "考试编辑", Perms: "exam:edit", Parent: "/exam", Sort: 3, Type: 2},
			{Name: "考试删除", Perms: "exam:del", Parent: "/exam", Sort: 4, Type: 2},
		}
		defs = append(defs, examDefs...)
	}
	for _, d := range defs {
		if d.Type == 2 {
			var parent model.Menu
			if err := database.DB.Where("`menu_path` = ?", d.Parent).First(&parent).Error; err != nil {
				continue
			}
			var cnt int64
			database.DB.Model(&model.Menu{}).Where("`menu_perms` = ?", d.Perms).Count(&cnt)
			if cnt > 0 {
				continue
			}
			database.DB.Create(&model.Menu{
				Name:     d.Name,
				ParentID: parent.ID,
				Perms:    d.Perms,
				Sort:     d.Sort,
				Status:   1,
				Type:     2,
				AddTime:  database.Now(),
				EditTime: database.Now(),
			})
		} else {
			var cnt int64
			database.DB.Model(&model.Menu{}).Where("`menu_path` = ?", d.Path).Count(&cnt)
			if cnt > 0 {
				continue
			}
			m := model.Menu{
				Name:     d.Name,
				Path:     d.Path,
				Perms:    d.Perms,
				Icon:     d.Icon,
				Sort:     d.Sort,
				Status:   1,
				Type:     d.Type,
				AddTime:  database.Now(),
				EditTime: database.Now(),
			}
			if d.Parent != "" {
				var parent model.Menu
				if err := database.DB.Where("`menu_path` = ?", d.Parent).First(&parent).Error; err == nil {
					m.ParentID = parent.ID
				}
			}
			database.DB.Create(&m)
		}
	}

	markMenuSeedInitialized()
}

func shouldSeedMenus(existingMenus int64, markerValue string) bool {
	return existingMenus == 0 && !isMenuSeedInitialized(markerValue)
}

func isMenuSeedInitialized(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func getMenuSeedInitializedValue() (string, error) {
	var setup model.Setup
	err := database.DB.Where("setup_key = ?", menuSeedInitializedKey).First(&setup).Error
	if err == nil {
		return setup.Value, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return "", err
}

func markMenuSeedInitialized() {
	now := database.Now()
	var setup model.Setup
	err := database.DB.Where("setup_key = ?", menuSeedInitializedKey).First(&setup).Error
	if err == nil {
		if isMenuSeedInitialized(setup.Value) {
			return
		}
		if err := database.DB.Model(&setup).Updates(map[string]interface{}{
			"setup_value":     "1",
			"setup_type":      "system",
			"setup_edit_time": now,
		}).Error; err != nil {
			log.Printf("mark menu seed initialized error: %v", err)
		}
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("read menu seed marker error: %v", err)
		return
	}

	if err := database.DB.Create(&model.Setup{
		Key:      menuSeedInitializedKey,
		Value:    "1",
		Type:     "system",
		AddTime:  now,
		EditTime: now,
	}).Error; err != nil {
		log.Printf("create menu seed marker error: %v", err)
	}
}
