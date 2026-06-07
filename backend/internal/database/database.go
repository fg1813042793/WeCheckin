package database

import (
	"fmt"
	"log"
	"time"

	"wecheckin-backend/backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(host string, port int, user, password, dbname string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := autoMigrate(); err != nil {
		log.Printf("Migration warning: %v (continuing)", err)
	}

	DB.Exec("DROP TABLE IF EXISTS `user_form_fields`")

	if err := DB.AutoMigrate(&model.Role{}); err != nil {
		log.Printf("Migration warning (role): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.SysDict{}); err != nil {
		log.Printf("Migration warning (sys_dict): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.Department{}); err != nil {
		log.Printf("Migration warning (department): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.UserDept{}); err != nil {
		log.Printf("Migration warning (user_dept): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.Menu{}); err != nil {
		log.Printf("Migration warning (menu): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.RoleMenu{}); err != nil {
		log.Printf("Migration warning (role_menu): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.AdminDept{}); err != nil {
		log.Printf("Migration warning (admin_dept): %v (continuing)", err)
	}

	if err := DB.AutoMigrate(&model.RoleDept{}); err != nil {
		log.Printf("Migration warning (role_dept): %v (continuing)", err)
	}

	log.Println("Database initialized successfully")
	migrateExamMenus()
	seedMenus()
	seedSetups()
}

// migrateExamMenus 将旧版 exam 菜单从 survey 迁移到独立子系统
func migrateExamMenus() {
	// 1. 删除旧版挂在 /survey 下的考试子菜单
	DB.Where("`menu_path` LIKE ?", "/survey/exam/%").Delete(&model.Menu{})

	// 2. 删除旧版挂在 /survey 下的考试相关按钮权限
	var surveyMenu model.Menu
	if err := DB.Where("`menu_path` = ?", "/survey").First(&surveyMenu).Error; err == nil {
		oldPatterns := []string{"question:%", "paper:%", "exam:%", "record:%", "grade:%"}
		for _, p := range oldPatterns {
			DB.Where("`menu_parent_id` = ? AND `menu_perms` LIKE ?", surveyMenu.ID, p).Delete(&model.Menu{})
		}
		// 3. 更新 survey 目录的 perms 字段，去掉旧考试相关权限
		DB.Model(&model.Menu{}).Where("`menu_path` = ?", "/survey").
			Update("menu_perms", "survey:list,survey:add,survey:edit,survey:del,survey:status,survey:copy,response:list,response:del,response:export")
	}

	// 4. 删除旧的 /exam 相关菜单（如果有，避免与 seed 冲突）
	DB.Where("`menu_path` = ?", "/exam").Delete(&model.Menu{})
	DB.Where("`menu_path` = ?", "/exam/list").Delete(&model.Menu{})
	DB.Where("`menu_path` LIKE ?", "/exam/%").Delete(&model.Menu{})
}

func seedSetups() {
	type setupDef struct {
		Key   string
		Value string
		Type  string
	}
	defs := []setupDef{
		{Key: "ADMIN_SINGLE_LOGIN", Value: "0", Type: "switch"},
		{Key: "USER_SINGLE_LOGIN", Value: "0", Type: "switch"},
		{Key: "TOKEN_ADMIN_EXPIRE", Value: "168h", Type: "string"},
		{Key: "TOKEN_ADMIN_REDIS_PREFIX", Value: "admin_token:", Type: "string"},
		{Key: "TOKEN_USER_EXPIRE", Value: "999d", Type: "string"},
		{Key: "TOKEN_USER_REDIS_PREFIX", Value: "user_token:", Type: "string"},
	}
	for _, d := range defs {
		var existing model.Setup
		if err := DB.Where("setup_key = ?", d.Key).First(&existing).Error; err != nil {
			setup := model.Setup{
				Key:     d.Key,
				Value:   d.Value,
				Type:    d.Type,
				AddTime: time.Now().UnixMilli(),
			}
			if err := DB.Create(&setup).Error; err != nil {
				log.Printf("seed setup %s error: %v", d.Key, err)
			}
		}
	}
}

func seedMenus() {
	type menuDef struct {
		Name   string
		Path   string
		Perms  string
		Icon   string
		Sort   int
		Type   int
		Parent string // path of parent menu (empty for root)
	}
	defs := []menuDef{
		// Main pages
		{Name: "控制台", Path: "/dashboard", Perms: "", Icon: "Odometer", Sort: 1, Type: 1},
		{Name: "用户管理", Path: "/user", Perms: "user:list,user:add,user:edit,user:del,user:status", Icon: "User", Sort: 2, Type: 1},
		{Name: "在线用户", Path: "/online", Perms: "online:list,online:force_offline", Icon: "Monitor", Sort: 3, Type: 1},
		{Name: "打卡管理", Path: "/enroll", Perms: "enroll:list,enroll:add,enroll:edit,enroll:del,enroll:status,enroll:vouch,enroll:export,enroll:users", Icon: "List", Sort: 4, Type: 1},
		{Name: "内容管理", Path: "/news", Perms: "news:list,news:add,news:edit,news:del,news:status,news:vouch", Icon: "Document", Sort: 5, Type: 1},
		{Name: "管理员管理", Path: "/mgr", Perms: "mgr:list,mgr:add,mgr:edit,mgr:del", Icon: "Setting", Sort: 6, Type: 1},
		{Name: "操作日志", Path: "/log", Perms: "log:list,log:del", Icon: "Clock", Sort: 7, Type: 1},
		{Name: "字典管理", Path: "/dict", Perms: "dict:list,dict:add,dict:edit,dict:del", Icon: "Notebook", Sort: 8, Type: 1},
		{Name: "部门管理", Path: "/department", Perms: "dept:list,dept:add,dept:edit,dept:del", Icon: "FolderOpened", Sort: 9, Type: 1},
		{Name: "角色管理", Path: "/role", Perms: "role:list,role:add,role:edit,role:del", Icon: "UserFilled", Sort: 10, Type: 1},
		{Name: "菜单权限", Path: "/menu", Perms: "menu:list,menu:add,menu:edit,menu:del", Icon: "Grid", Sort: 11, Type: 1},
		{Name: "系统配置", Path: "/setup", Perms: "setup:list,setup:edit", Icon: "Setting", Sort: 12, Type: 1},
		{Name: "赛事活动", Path: "/event", Perms: "event:list,event:add,event:edit,event:del,event:status,event:vouch,event:top,event:users", Icon: "TrophyBase", Sort: 13, Type: 1},
		{Name: "问卷调查", Path: "/survey", Perms: "survey:list,survey:add,survey:edit,survey:del,survey:status,survey:copy,response:list,response:del,response:export", Icon: "List", Sort: 14, Type: 0},
		{Name: "问卷管理", Path: "/survey", Parent: "/survey", Sort: 1, Type: 1},
		{Name: "答卷管理", Path: "/survey/responses", Parent: "/survey", Sort: 2, Type: 1},
		{Name: "问卷统计", Path: "/survey/statistic", Parent: "/survey", Sort: 3, Type: 1},
		// Online exam (independent subsystem, separated from survey)
		{Name: "在线考试", Path: "/exam", Perms: "exam:list,exam:add,exam:edit,exam:del", Icon: "EditPen", Sort: 15, Type: 0},
		{Name: "考试管理", Path: "/exam/list", Parent: "/exam", Sort: 1, Type: 1},
		// Button permissions for each module (children of the corresponding parent menu)
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
		// Survey buttons
		{Name: "问卷列表", Perms: "survey:list", Parent: "/survey", Sort: 1, Type: 2},
		{Name: "问卷新增", Perms: "survey:add", Parent: "/survey", Sort: 2, Type: 2},
		{Name: "问卷编辑", Perms: "survey:edit", Parent: "/survey", Sort: 3, Type: 2},
		{Name: "问卷删除", Perms: "survey:del", Parent: "/survey", Sort: 4, Type: 2},
		{Name: "问卷状态管理", Perms: "survey:status", Parent: "/survey", Sort: 5, Type: 2},
		{Name: "复制问卷", Perms: "survey:copy", Parent: "/survey", Sort: 6, Type: 2},
		{Name: "答卷列表", Perms: "response:list", Parent: "/survey", Sort: 7, Type: 2},
		{Name: "答卷删除", Perms: "response:del", Parent: "/survey", Sort: 8, Type: 2},
		{Name: "导出答卷", Perms: "response:export", Parent: "/survey", Sort: 9, Type: 2},
		// Exam buttons
		{Name: "考试列表", Perms: "exam:list", Parent: "/exam", Sort: 1, Type: 2},
		{Name: "考试新增", Perms: "exam:add", Parent: "/exam", Sort: 2, Type: 2},
		{Name: "考试编辑", Perms: "exam:edit", Parent: "/exam", Sort: 3, Type: 2},
		{Name: "考试删除", Perms: "exam:del", Parent: "/exam", Sort: 4, Type: 2},
	}
	for _, d := range defs {
		if d.Type == 2 {
			// Button type: look up parentId by parent path, skip if perms already exists
			var parent model.Menu
			if err := DB.Where("`menu_path` = ?", d.Parent).First(&parent).Error; err != nil {
				continue
			}
			var cnt int64
			DB.Model(&model.Menu{}).Where("`menu_perms` = ?", d.Perms).Count(&cnt)
			if cnt > 0 {
				continue
			}
			DB.Create(&model.Menu{
				Name:     d.Name,
				ParentID: parent.ID,
				Perms:    d.Perms,
				Sort:     d.Sort,
				Status:   1,
				Type:     2,
				AddTime:  Now(),
				EditTime: Now(),
			})
		} else {
			// Menu type: create if not exists by path
			var cnt int64
			DB.Model(&model.Menu{}).Where("`menu_path` = ?", d.Path).Count(&cnt)
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
				AddTime:  Now(),
				EditTime: Now(),
			}
			if d.Parent != "" {
				var parent model.Menu
				if err := DB.Where("`menu_path` = ?", d.Parent).First(&parent).Error; err == nil {
					m.ParentID = parent.ID
				}
			}
			DB.Create(&m)
		}
	}
}

func autoMigrate() error {
	err := DB.AutoMigrate(
		&model.User{},
		&model.News{},
		&model.Enroll{},
		&model.EnrollJoin{},
		&model.EnrollUser{},
		&model.Favorite{},
		&model.Admin{},
		&model.Log{},
		&model.Setup{},
		&model.Role{},
		&model.SysDict{},
		&model.Department{},
		&model.UserDept{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.AdminDept{},
		&model.RoleDept{},
		&model.Event{},
		&model.EventRole{},
		&model.EventParticipant{},
		&model.EventDynamic{},
		&model.EventScore{},
		&model.ExamQuestion{},
		&model.ExamPaper{},
		&model.Exam{},
		&model.ExamRecord{},
		&model.Survey{},
		&model.SurveyResponse{},
		&model.SurveyChannel{},
		&model.SurveyAILog{},
		&model.SurveyResource{},
		&model.SurveyQuestion{},
	)
	if err != nil {
		return err
	}
	DB.Exec("ALTER TABLE `event_scores` MODIFY COLUMN `event_score_score` TEXT COMMENT '成绩'")
	return nil
}

func Now() int64 {
	return time.Now().UnixMilli()
}

func GetDB() *gorm.DB {
	return DB
}
