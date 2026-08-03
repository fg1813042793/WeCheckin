-- 为资源表标准审计字段切换准备临时历史列。
--
-- 背景：
-- - 旧索引迁移早于统一审计迁移，仍引用 news_add_time、event_dept_id 等历史列。
-- - 新模型已改为 create_by/update_by/create_dept_id/update_dept_id/add_time/edit_time。
-- - 对于全新库，AutoMigrate 不再创建历史列，因此这里在旧索引迁移前临时补齐。
-- - 如果统一审计迁移已经执行，本迁移跳过，避免在已升级环境重新补旧列。

SET @unified_resource_audit_migration_ran := (
  SELECT COUNT(1)
  FROM `schema_migrations`
  WHERE `migration_version` = '20260730100000_add_unified_resource_audit_fields'
    AND `migration_status` = 'success'
);

SET @table_name := 'news';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'news_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `news` ADD COLUMN `news_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''news.news_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'news_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `news` ADD COLUMN `news_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''news.news_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'news_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `news` ADD COLUMN `news_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''news.news_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'news_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `news` ADD COLUMN `news_edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史更新时间''', 'SELECT ''news.news_edit_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'enrolls';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'enroll_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `enrolls` ADD COLUMN `enroll_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''enrolls.enroll_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'enroll_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `enrolls` ADD COLUMN `enroll_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''enrolls.enroll_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'enroll_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `enrolls` ADD COLUMN `enroll_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''enrolls.enroll_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'enroll_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `enrolls` ADD COLUMN `enroll_edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史更新时间''', 'SELECT ''enrolls.enroll_edit_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'events';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'event_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `events` ADD COLUMN `event_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''events.event_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'event_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `events` ADD COLUMN `event_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''events.event_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'event_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `events` ADD COLUMN `event_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''events.event_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'event_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `events` ADD COLUMN `event_edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史更新时间''', 'SELECT ''events.event_edit_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'survey';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'survey_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey` ADD COLUMN `survey_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''survey.survey_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey` ADD COLUMN `survey_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''survey.survey_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey` ADD COLUMN `survey_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''survey.survey_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey` ADD COLUMN `survey_edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史更新时间''', 'SELECT ''survey.survey_edit_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'exam_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam` ADD COLUMN `exam_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''exam.exam_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam` ADD COLUMN `exam_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''exam.exam_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam` ADD COLUMN `exam_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''exam.exam_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam` ADD COLUMN `exam_edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史更新时间''', 'SELECT ''exam.exam_edit_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'survey_question';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'survey_q_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey_question` ADD COLUMN `survey_q_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''survey_question.survey_q_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_q_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey_question` ADD COLUMN `survey_q_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''survey_question.survey_q_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_q_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `survey_question` ADD COLUMN `survey_q_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''survey_question.survey_q_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam_question';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'exam_q_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam_question` ADD COLUMN `exam_q_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''exam_question.exam_q_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_q_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam_question` ADD COLUMN `exam_q_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''exam_question.exam_q_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_q_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam_question` ADD COLUMN `exam_q_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''exam_question.exam_q_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam_paper';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'exam_p_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam_paper` ADD COLUMN `exam_p_create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建人ID''', 'SELECT ''exam_paper.exam_p_create_by prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_p_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam_paper` ADD COLUMN `exam_p_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''历史创建部门ID''', 'SELECT ''exam_paper.exam_p_dept_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_p_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `exam_paper` ADD COLUMN `exam_p_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''exam_paper.exam_p_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'logs';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @column_name := 'log_admin_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_admin_id` varchar(50) NOT NULL DEFAULT '''' COMMENT ''历史管理员ID''', 'SELECT ''logs.log_admin_id prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'log_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@unified_resource_audit_migration_ran = 0 AND @table_exists > 0 AND @col_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_add_time` bigint NOT NULL DEFAULT 0 COMMENT ''历史创建时间''', 'SELECT ''logs.log_add_time prepare skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
