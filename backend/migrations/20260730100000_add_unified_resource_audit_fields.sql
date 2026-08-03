-- 为旧业务主表补齐统一审计/数据归属影子字段。
--
-- 说明：
-- - 本迁移不删除、不重命名历史前缀字段，避免破坏现有模型和线上查询。
-- - 新增字段用于后续把 access.ResourceAuditFields 映射逐步切到统一列。
-- - 历史数据从现有 `xxx_create_by`、`xxx_dept_id`、`xxx_add_time`、`xxx_edit_time` 回填。
-- - 无历史更新人字段的表，`update_by` 暂按创建人回填，`edit_time` 暂按已有修改时间或创建时间回填。
--
-- 回滚参考：
-- ALTER TABLE `news` DROP INDEX `idx_news_unified_audit_scope`, DROP COLUMN `create_by`, DROP COLUMN `update_by`, DROP COLUMN `create_dept_id`, DROP COLUMN `create_dept_path`, DROP COLUMN `add_time`, DROP COLUMN `edit_time`;
-- 其他表按同名字段和 `idx_*_unified_audit_scope` 索引执行对应 DROP。

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `news`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `news_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''news unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `news`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`news_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`news_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`news_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`news_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`news_edit_time`, 0), `news_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = 'idx_news_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `news` ADD INDEX `idx_news_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `id`)',
  'SELECT ''idx_news_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enrolls' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `enrolls`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `enroll_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''enrolls unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `enrolls`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`enroll_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`enroll_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`enroll_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`enroll_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`enroll_edit_time`, 0), `enroll_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enrolls' AND INDEX_NAME = 'idx_enrolls_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enrolls` ADD INDEX `idx_enrolls_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `id`)',
  'SELECT ''idx_enrolls_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `events`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `event_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''events unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `events`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`event_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`event_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`event_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`event_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`event_edit_time`, 0), `event_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND INDEX_NAME = 'idx_events_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `events` ADD INDEX `idx_events_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `id`)',
  'SELECT ''idx_events_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `survey`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `survey_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''survey unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `survey`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`survey_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`survey_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`survey_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`survey_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`survey_edit_time`, 0), `survey_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = 'idx_survey_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `survey` ADD INDEX `idx_survey_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `survey_id`)',
  'SELECT ''idx_survey_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `exam`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `exam_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''exam unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `exam`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`exam_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`exam_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`exam_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`exam_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`exam_edit_time`, 0), `exam_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = 'idx_exam_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `exam` ADD INDEX `idx_exam_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `exam_id`)',
  'SELECT ''idx_exam_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `survey_question`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `survey_q_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''survey_question unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `survey_question`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`survey_q_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`survey_q_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`survey_q_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`survey_q_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`survey_q_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = 'idx_survey_q_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `survey_question` ADD INDEX `idx_survey_q_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `survey_q_id`)',
  'SELECT ''idx_survey_q_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `exam_question`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `exam_q_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''exam_question unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `exam_question`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`exam_q_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`exam_q_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`exam_q_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`exam_q_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`exam_q_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = 'idx_exam_q_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `exam_question` ADD INDEX `idx_exam_q_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `exam_q_id`)',
  'SELECT ''idx_exam_q_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_paper' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `exam_paper`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `exam_p_create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''exam_paper unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `exam_paper`
SET
  `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`exam_p_create_by`, 0), `create_by`),
  `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`exam_p_create_by`, 0), `create_by`, 0), `update_by`),
  `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`exam_p_dept_id`, 0), `create_dept_id`),
  `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`exam_p_add_time`, 0), `add_time`),
  `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`exam_p_add_time`, 0), `edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_paper' AND INDEX_NAME = 'idx_exam_p_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `exam_paper` ADD INDEX `idx_exam_p_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `exam_p_id`)',
  'SELECT ''idx_exam_p_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `logs`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人ID'' AFTER `log_admin_id`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人ID'' AFTER `create_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一创建人部门ID'' AFTER `update_by`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''统一创建人部门路径'' AFTER `create_dept_id`,
     ADD COLUMN `add_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一创建时间'' AFTER `create_dept_path`,
     ADD COLUMN `edit_time` bigint NOT NULL DEFAULT 0 COMMENT ''统一更新时间'' AFTER `add_time`',
  'SELECT ''logs unified audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `logs` l
LEFT JOIN (
  SELECT `user_dept_user_id`, MIN(`user_dept_dept_id`) AS `dept_id`
  FROM `user_depts`
  GROUP BY `user_dept_user_id`
) ud ON ud.`user_dept_user_id` = CASE
  WHEN l.`log_admin_id` REGEXP '^[0-9]+$' THEN CAST(l.`log_admin_id` AS UNSIGNED)
  ELSE 0
END
SET
  l.`create_by` = IF(COALESCE(l.`create_by`, 0) = 0, CASE WHEN l.`log_admin_id` REGEXP '^[0-9]+$' THEN CAST(l.`log_admin_id` AS UNSIGNED) ELSE 0 END, l.`create_by`),
  l.`update_by` = IF(COALESCE(l.`update_by`, 0) = 0, CASE WHEN l.`log_admin_id` REGEXP '^[0-9]+$' THEN CAST(l.`log_admin_id` AS UNSIGNED) ELSE 0 END, l.`update_by`),
  l.`create_dept_id` = IF(COALESCE(l.`create_dept_id`, 0) = 0, COALESCE(ud.`dept_id`, 0), l.`create_dept_id`),
  l.`add_time` = IF(COALESCE(l.`add_time`, 0) = 0, COALESCE(l.`log_add_time`, 0), l.`add_time`),
  l.`edit_time` = IF(COALESCE(l.`edit_time`, 0) = 0, COALESCE(l.`log_add_time`, 0), l.`edit_time`);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_unified_audit_scope` (`create_dept_id`, `create_by`, `add_time`, `id`)',
  'SELECT ''idx_logs_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
