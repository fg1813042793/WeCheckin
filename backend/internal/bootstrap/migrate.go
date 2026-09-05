package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

type migrationStep struct {
	Name string
	Run  func() error
}

type migrationLogFunc func(format string, args ...interface{})

func runMigrationSteps(steps []migrationStep, logf migrationLogFunc) error {
	for _, step := range steps {
		if err := step.Run(); err != nil {
			return fmt.Errorf("migration step %s: %w", step.Name, err)
		}
		if logf != nil {
			logf("migration step completed: " + step.Name)
		}
	}
	return nil
}

func autoMigrate(enableExam bool) error {
	db, cancel := startupDB(context.Background())
	defer cancel()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	err := db.AutoMigrate(
		&model.User{},
		&model.News{},
		&model.Enroll{},
		&model.EnrollJoin{},
		&model.EnrollUser{},
		&model.Favorite{},
		&model.Log{},
		&model.Setup{},
		&model.Role{},
		&model.SysDictType{},
		&model.SysDict{},
		&model.Department{},
		&model.Position{},
		&model.UserDept{},
		&model.UserRole{},
		&model.Permission{},
		&model.PermissionGrant{},
		&model.Event{},
		&model.EventRole{},
		&model.EventParticipant{},
		&model.EventDynamic{},
		&model.EventScore{},
		&model.ExamQuestion{},
		&model.ExamPaper{},
		&model.Exam{},
		&model.ExamRecord{},
		&model.ExamResource{},
		&model.Survey{},
		&model.SurveyResponse{},
		&model.SurveyChannel{},
		&model.SurveyAILog{},
		&model.SurveyResource{},
		&model.SurveyQuestion{},
		&model.Notify{},
		&model.DingTalkH5CorpConfig{},
		&model.DingTalkH5UserBinding{},
		&model.DingTalkH5PerfReview{},
		&model.DingTalkH5PerfHistory{},
		&model.DingTalkH5PerfTemplate{},
		&model.WorkflowDefinition{},
		&model.WorkflowDefinitionVersion{},
		// workflow_process_instances is changed only by versioned SQL because new columns require explicit backfills.
		&model.WorkflowProcessToken{},
		&model.WorkflowProcessTask{},
		&model.WorkflowProcessVariable{},
		&model.WorkflowProcessHistory{},
	)
	if err != nil {
		return err
	}
	return runMigrationSteps([]migrationStep{
		{
			Name: "merge_admins_into_users",
			Run: func() error {
				return mergeAdminsIntoUsers(db)
			},
		},
		{
			Name: "ensure_role_admin_login_control",
			Run: func() error {
				return ensureRoleAdminLoginControl(db)
			},
		},
		{
			Name: "ensure_unified_permissions",
			Run: func() error {
				return ensureUnifiedPermissions(db, enableExam)
			},
		},
		{
			Name: "cleanup_legacy_role_authorization_tables",
			Run: func() error {
				return cleanupLegacyRoleAuthorizationTables(db)
			},
		},
		{
			Name: "cleanup_legacy_menu_table",
			Run: func() error {
				return cleanupLegacyMenuTable(db)
			},
		},
		{
			Name: "cleanup_legacy_admin_tables",
			Run: func() error {
				return cleanupLegacyAdminTables(db)
			},
		},
		{
			Name: "cleanup_dingtalk_h5_perf_users",
			Run: func() error {
				return cleanupDingTalkH5PerfUsers(db)
			},
		},
		{
			Name: "adjust_event_scores_score_text",
			Run: func() error {
				return db.Exec("ALTER TABLE `event_scores` MODIFY COLUMN `event_score_score` TEXT COMMENT '成绩'").Error
			},
		},
		{
			Name: "adjust_survey_schema_mediumtext",
			Run: func() error {
				return db.Exec("ALTER TABLE `survey` MODIFY COLUMN `survey_schema` MEDIUMTEXT COMMENT 'formkit schema (JSON)'").Error
			},
		},
	}, log.Printf)
}

type legacyAdminAccount struct {
	ID        uint
	Name      string
	Password  string
	Desc      string
	Pic       string
	Phone     string
	Status    int
	Type      int
	RoleID    uint
	Token     string
	TokenTime int64
	LoginCnt  int
	LoginTime int64
	AddTime   int64
	EditTime  int64
	AddIP     string
	EditIP    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type adminUserMergeMap struct {
	ID            uint      `gorm:"primaryKey"`
	LegacyAdminID uint      `gorm:"column:legacy_admin_id"`
	UserID        uint      `gorm:"column:user_id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (adminUserMergeMap) TableName() string { return "admin_user_merge_maps" }

func mergeAdminsIntoUsers(db *gorm.DB) error {
	exists, err := databaseTableExists(db, "admins")
	if err != nil || !exists {
		return err
	}
	var legacyAdmins []legacyAdminAccount
	if err := db.Table("admins").Select(
		"`id`, `admin_name` AS `name`, `admin_password` AS `password`, `admin_desc` AS `desc`, " +
			"`admin_pic` AS `pic`, `admin_phone` AS `phone`, `admin_status` AS `status`, " +
			"`admin_type` AS `type`, `admin_role_id` AS `role_id`, `admin_token` AS `token`, " +
			"`admin_token_time` AS `token_time`, `admin_login_cnt` AS `login_cnt`, " +
			"`admin_login_time` AS `login_time`, `admin_add_time` AS `add_time`, " +
			"`admin_edit_time` AS `edit_time`, `admin_add_ip` AS `add_ip`, `admin_edit_ip` AS `edit_ip`, " +
			"`created_at`, `updated_at`",
	).Scan(&legacyAdmins).Error; err != nil {
		return err
	}
	if len(legacyAdmins) == 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := ensureAdminUserMergeMapTable(tx); err != nil {
			return err
		}
		legacyAdminDeptsExists, err := databaseTableExists(tx, "admin_depts")
		if err != nil {
			return err
		}
		logsExists, err := databaseTableExists(tx, "logs")
		if err != nil {
			return err
		}
		logCreateByExists := false
		logAdminIDExists := false
		if logsExists {
			logCreateByExists, err = databaseColumnExists(tx, "logs", "create_by")
			if err != nil {
				return err
			}
			logAdminIDExists, err = databaseColumnExists(tx, "logs", "log_admin_id")
			if err != nil {
				return err
			}
		}
		for _, item := range legacyAdmins {
			merged, err := upsertMergedAdminUser(tx, item)
			if err != nil {
				return err
			}
			var mapCount int64
			if err := tx.Model(&adminUserMergeMap{}).Where("`legacy_admin_id` = ?", item.ID).Count(&mapCount).Error; err != nil {
				return err
			}
			if legacyAdminDeptsExists {
				if err := tx.Exec(
					"INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`) "+
						"SELECT DISTINCT ?, ad.`admin_dept_dept_id`, NOW(3), NOW(3) FROM `admin_depts` ad "+
						"LEFT JOIN `user_depts` ud ON ud.`user_dept_user_id` = ? AND ud.`user_dept_dept_id` = ad.`admin_dept_dept_id` "+
						"WHERE ad.`admin_dept_admin_id` IN (?, ?) AND ad.`admin_dept_dept_id` > 0 AND ud.`id` IS NULL",
					merged.ID, merged.ID, item.ID, merged.ID,
				).Error; err != nil {
					return err
				}
			}
			if mapCount == 0 {
				if logsExists {
					if logCreateByExists {
						if logAdminIDExists {
							if err := tx.Exec(
								"UPDATE `logs` SET `create_by` = ? WHERE `create_by` = ? OR `log_admin_id` = ?",
								merged.ID, item.ID, fmt.Sprint(item.ID),
							).Error; err != nil {
								return err
							}
						} else if err := tx.Exec("UPDATE `logs` SET `create_by` = ? WHERE `create_by` = ?", merged.ID, item.ID).Error; err != nil {
							return err
						}
					}
				}
			}
			if err := tx.Where("`legacy_admin_id` = ?", item.ID).Assign(adminUserMergeMap{UserID: merged.ID}).FirstOrCreate(&adminUserMergeMap{
				LegacyAdminID: item.ID,
				UserID:        merged.ID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertMergedAdminUser(tx *gorm.DB, item legacyAdminAccount) (model.User, error) {
	miniOpenID := fmt.Sprintf("admin:%d", item.ID)
	displayName := item.Desc
	if strings.TrimSpace(displayName) == "" {
		displayName = item.Name
	}
	now := time.Now()
	createdAt := migrationTimeOrNow(item.CreatedAt, now)
	updatedAt := migrationTimeOrNow(item.UpdatedAt, now)
	updates := map[string]interface{}{
		"user_account":          item.Name,
		"user_name":             displayName,
		"user_password":         item.Password,
		"user_admin_desc":       item.Desc,
		"user_pic":              item.Pic,
		"user_mobile":           item.Phone,
		"user_status":           item.Status,
		"user_admin_enabled":    1,
		"user_admin_type":       item.Type,
		"user_role_id":          item.RoleID,
		"user_admin_token":      item.Token,
		"user_admin_token_time": item.TokenTime,
		"user_login_cnt":        item.LoginCnt,
		"user_login_time":       item.LoginTime,
		"user_add_time":         item.AddTime,
		"user_edit_time":        item.EditTime,
		"user_add_ip":           item.AddIP,
		"user_edit_ip":          item.EditIP,
		"updated_at":            updatedAt,
	}
	var existing model.User
	err := tx.Where("`user_mini_openid` = ?", miniOpenID).First(&existing).Error
	if err == nil {
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return model.User{}, err
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, err
	}
	user := model.User{
		MiniOpenID:     miniOpenID,
		Account:        item.Name,
		Name:           displayName,
		Password:       item.Password,
		AdminDesc:      item.Desc,
		Pic:            item.Pic,
		Mobile:         item.Phone,
		Status:         item.Status,
		AdminEnabled:   1,
		AdminType:      item.Type,
		RoleID:         item.RoleID,
		AdminToken:     item.Token,
		AdminTokenTime: item.TokenTime,
		LoginCnt:       item.LoginCnt,
		LoginTime:      item.LoginTime,
		AddTime:        item.AddTime,
		EditTime:       item.EditTime,
		AddIP:          item.AddIP,
		EditIP:         item.EditIP,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	var idCount int64
	if err := tx.Model(&model.User{}).Where("`id` = ?", item.ID).Count(&idCount).Error; err != nil {
		return model.User{}, err
	}
	if idCount == 0 {
		user.ID = item.ID
	}
	if err := tx.Create(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func migrationTimeOrNow(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() || value.Year() <= 1 {
		return fallback
	}
	return value
}

func ensureRoleAdminLoginControl(db *gorm.DB) error {
	rolesExists, err := databaseTableExists(db, "roles")
	if err != nil || !rolesExists {
		return err
	}
	usersExists, err := databaseTableExists(db, "users")
	if err != nil || !usersExists {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		now := database.Now()
		var role model.Role
		err := tx.Where("`role_name` = ?", "超级管理员").Order("`id` ASC").First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = model.Role{
				Name:            "超级管理员",
				Remark:          "系统内置角色",
				Sort:            0,
				Status:          1,
				AllowAdminLogin: 1,
				DataScope:       1,
				AddTime:         now,
				EditTime:        now,
				AddIP:           "127.0.0.1",
				EditIP:          "127.0.0.1",
			}
			if err := tx.Create(&role).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if role.Status != 1 || role.AllowAdminLogin != 1 {
			if err := tx.Model(&role).Updates(map[string]interface{}{
				"role_status":            1,
				"role_allow_admin_login": 1,
				"role_edit_time":         now,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.User{}).
			Where("`user_admin_type` = 1 AND (`user_role_id` IS NULL OR `user_role_id` = 0)").
			Updates(map[string]interface{}{
				"user_role_id":   role.ID,
				"user_edit_time": now,
			}).Error
	})
}

func ensureAdminUserMergeMapTable(db *gorm.DB) error {
	return db.Exec("CREATE TABLE IF NOT EXISTS `admin_user_merge_maps` (`id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '映射ID', `legacy_admin_id` bigint unsigned NOT NULL COMMENT '旧admins.id', `user_id` bigint unsigned NOT NULL COMMENT '新users.id', `created_at` datetime(3) DEFAULT NULL, `updated_at` datetime(3) DEFAULT NULL, PRIMARY KEY (`id`), UNIQUE KEY `idx_admin_user_merge_legacy` (`legacy_admin_id`), UNIQUE KEY `idx_admin_user_merge_user` (`user_id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员合并用户映射表'").Error
}

func ensureUnifiedPermissions(db *gorm.DB, enableExam bool) error {
	return permissionsupport.EnsureUnifiedPermissionsContext(context.Background(), db, enableExam)
}

func cleanupLegacyRoleAuthorizationTables(db *gorm.DB) error {
	return removeObsoleteTables(db, []string{"role_menus", "role_depts"})
}

func cleanupLegacyMenuTable(db *gorm.DB) error {
	return removeObsoleteTables(db, []string{"menus"})
}

func cleanupLegacyAdminTables(db *gorm.DB) error {
	adminDeptsExists, err := databaseTableExists(db, "admin_depts")
	if err != nil {
		return err
	}
	if adminDeptsExists {
		mergeMapExists, err := databaseTableExists(db, "admin_user_merge_maps")
		if err != nil {
			return err
		}
		if mergeMapExists {
			if err := db.Exec(
				"INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`) " +
					"SELECT DISTINCT COALESCE(m.`user_id`, ad.`admin_dept_admin_id`), ad.`admin_dept_dept_id`, NOW(3), NOW(3) " +
					"FROM `admin_depts` ad " +
					"LEFT JOIN `admin_user_merge_maps` m ON m.`legacy_admin_id` = ad.`admin_dept_admin_id` " +
					"INNER JOIN `users` u ON u.`id` = COALESCE(m.`user_id`, ad.`admin_dept_admin_id`) " +
					"LEFT JOIN `user_depts` ud ON ud.`user_dept_user_id` = COALESCE(m.`user_id`, ad.`admin_dept_admin_id`) AND ud.`user_dept_dept_id` = ad.`admin_dept_dept_id` " +
					"WHERE ad.`admin_dept_dept_id` > 0 AND ud.`id` IS NULL",
			).Error; err != nil {
				return err
			}
		} else {
			if err := db.Exec(
				"INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`) " +
					"SELECT DISTINCT ad.`admin_dept_admin_id`, ad.`admin_dept_dept_id`, NOW(3), NOW(3) " +
					"FROM `admin_depts` ad " +
					"INNER JOIN `users` u ON u.`id` = ad.`admin_dept_admin_id` " +
					"LEFT JOIN `user_depts` ud ON ud.`user_dept_user_id` = ad.`admin_dept_admin_id` AND ud.`user_dept_dept_id` = ad.`admin_dept_dept_id` " +
					"WHERE ad.`admin_dept_dept_id` > 0 AND ud.`id` IS NULL",
			).Error; err != nil {
				return err
			}
		}
	}
	return removeObsoleteTables(db, []string{"admin_depts", "admins", "admin_user_merge_maps"})
}

func cleanupDingTalkH5PerfUsers(db *gorm.DB) error {
	exists, err := databaseTableExists(db, "dingtalk_h5_perf_users")
	if err != nil || !exists {
		return err
	}
	hasPosition, err := databaseColumnExists(db, "dingtalk_h5_perf_users", "position")
	if err != nil {
		return err
	}
	if !hasPosition {
		if err := db.Exec("ALTER TABLE `dingtalk_h5_perf_users` ADD COLUMN `position` varchar(100) DEFAULT '' COMMENT '岗位' AFTER `role`").Error; err != nil {
			return err
		}
	}
	if err := db.Exec(
		"INSERT INTO `users` (" +
			"`user_mini_openid`, `user_status`, `user_name`, `user_mobile`, `user_pic`, `user_forms`, `user_obj`, `user_password`, " +
			"`user_login_cnt`, `user_login_time`, `user_add_time`, `user_add_ip`, `user_edit_time`, `user_edit_ip`, `created_at`, `updated_at`) " +
			"SELECT p.`account`, COALESCE(p.`status`, 1), COALESCE(NULLIF(p.`name`, ''), p.`account`), '', '/static/default-avatar.png', '[]', " +
			"JSON_OBJECT('dingtalkH5Performance', JSON_OBJECT(" +
			"'role', COALESCE(NULLIF(p.`role`, ''), 'employee'), " +
			"'position', COALESCE(p.`position`, ''), " +
			"'department', COALESCE(p.`department`, ''), " +
			"'departmentLevel1', COALESCE(p.`department_level1`, ''), " +
			"'departmentLevel2', COALESCE(p.`department_level2`, ''), " +
			"'departmentLevel3', COALESCE(p.`department_level3`, ''), " +
			"'managerId', COALESCE(p.`manager_account`, ''), " +
			"'hrbpId', COALESCE(p.`hrbp_account`, ''), " +
			"'responsibleDepartments', IF(JSON_VALID(COALESCE(p.`responsible_departments`, '')), JSON_EXTRACT(p.`responsible_departments`, '$'), JSON_ARRAY()))) " +
			"AS `user_obj`, COALESCE(p.`password_hash`, ''), 0, 0, " +
			"COALESCE(p.`add_time`, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)), '127.0.0.1', " +
			"COALESCE(p.`edit_time`, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)), '127.0.0.1', " +
			"CASE WHEN p.`created_at` IS NULL OR CAST(p.`created_at` AS CHAR) IN ('0000-00-00', '0000-00-00 00:00:00', '0000-00-00 00:00:00.000') THEN NOW(3) ELSE p.`created_at` END, " +
			"CASE WHEN p.`updated_at` IS NULL OR CAST(p.`updated_at` AS CHAR) IN ('0000-00-00', '0000-00-00 00:00:00', '0000-00-00 00:00:00.000') THEN NOW(3) ELSE p.`updated_at` END " +
			"FROM `dingtalk_h5_perf_users` p WHERE p.`account` <> '' " +
			"ON DUPLICATE KEY UPDATE " +
			"`user_status` = VALUES(`user_status`), " +
			"`user_name` = VALUES(`user_name`), " +
			"`user_password` = IF(COALESCE(`users`.`user_password`, '') = '', VALUES(`user_password`), `users`.`user_password`), " +
			"`user_pic` = IF(COALESCE(`users`.`user_pic`, '') = '', VALUES(`user_pic`), `users`.`user_pic`), " +
			"`user_forms` = IF(COALESCE(`users`.`user_forms`, '') = '', VALUES(`user_forms`), `users`.`user_forms`), " +
			"`user_obj` = JSON_SET(IF(JSON_VALID(COALESCE(`users`.`user_obj`, '')), `users`.`user_obj`, JSON_OBJECT()), '$.dingtalkH5Performance', JSON_EXTRACT(VALUES(`user_obj`), '$.dingtalkH5Performance')), " +
			"`user_edit_time` = VALUES(`user_edit_time`), `updated_at` = NOW(3)",
	).Error; err != nil {
		return err
	}
	return removeObsoleteTables(db, []string{"dingtalk_h5_perf_users"})
}

func removeObsoleteTables(db *gorm.DB, names []string) error {
	for _, name := range names {
		quoted, err := quoteTableName(name)
		if err != nil {
			return err
		}
		sql := strings.Join([]string{"DROP", "TABLE", "IF", "EXISTS", quoted}, " ")
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func quoteTableName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty table name")
	}
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return "", fmt.Errorf("invalid table name %q", name)
		}
	}
	return "`" + name + "`", nil
}

func databaseTableExists(db *gorm.DB, tableName string) (bool, error) {
	var count int64
	if err := db.Raw("SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", tableName).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func databaseColumnExists(db *gorm.DB, tableName, columnName string) (bool, error) {
	var count int64
	if err := db.Raw("SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", tableName, columnName).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
