-- 补齐钉钉 H5 独立业务表的数据归属字段。
--
-- 说明：
-- - `users` 是共享用户表，不在本迁移中新增 H5 专属字段。
-- - `dingtalk_h5_perf_reviews` 的历史归属按被考评员工账号和用户部门做近似回填。
-- - `dingtalk_h5_perf_histories` 的历史归属优先按操作人账号回填，部门路径沿用对应考评单。
-- - `dingtalk_h5_perf_sessions` 和模板表只补齐字段，便于统一审计，不作为业务列表的数据范围主表。

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_sessions' AND COLUMN_NAME = 'edit_time'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_sessions` ADD COLUMN `edit_time` bigint DEFAULT 0 COMMENT ''修改时间'' AFTER `add_time`',
  'SELECT ''dingtalk_h5_perf_sessions.edit_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_sessions' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_sessions`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人ID'' AFTER `edit_time`,
     ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人ID'' AFTER `create_account`,
     ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人部门ID'' AFTER `update_account`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`',
  'SELECT ''dingtalk_h5_perf_sessions audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_reviews' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人ID'' AFTER `edit_time`,
     ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人ID'' AFTER `create_account`,
     ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人部门ID'' AFTER `update_account`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`',
  'SELECT ''dingtalk_h5_perf_reviews audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_histories' AND COLUMN_NAME = 'edit_time'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_histories` ADD COLUMN `edit_time` bigint DEFAULT 0 COMMENT ''修改时间'' AFTER `add_time`',
  'SELECT ''dingtalk_h5_perf_histories.edit_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_histories' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_histories`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人ID'' AFTER `edit_time`,
     ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人ID'' AFTER `create_account`,
     ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人部门ID'' AFTER `update_account`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`',
  'SELECT ''dingtalk_h5_perf_histories audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_templates' AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_templates`
     ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人ID'' AFTER `edit_time`,
     ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`,
     ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人ID'' AFTER `create_account`,
     ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`,
     ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人部门ID'' AFTER `update_account`,
     ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`',
  'SELECT ''dingtalk_h5_perf_templates audit fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_reviews' AND INDEX_NAME = 'idx_dt_h5_review_create_by'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_create_by` (`create_by`)',
  'SELECT ''idx_dt_h5_review_create_by exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_reviews' AND INDEX_NAME = 'idx_dt_h5_review_create_account'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_create_account` (`create_account`)',
  'SELECT ''idx_dt_h5_review_create_account exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'dingtalk_h5_perf_reviews' AND INDEX_NAME = 'idx_dt_h5_review_create_dept'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_create_dept` (`create_dept_id`)',
  'SELECT ''idx_dt_h5_review_create_dept exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `dingtalk_h5_perf_sessions` s
LEFT JOIN `users` u ON BINARY u.`user_mini_openid` = BINARY s.`user_account`
LEFT JOIN (
  SELECT `user_dept_user_id`, MIN(`user_dept_dept_id`) AS `dept_id`
  FROM `user_depts`
  GROUP BY `user_dept_user_id`
) ud ON ud.`user_dept_user_id` = u.`id`
SET
  s.`edit_time` = IF(COALESCE(s.`edit_time`, 0) = 0, COALESCE(s.`add_time`, 0), s.`edit_time`),
  s.`create_by` = IF(COALESCE(s.`create_by`, 0) = 0, COALESCE(u.`id`, 0), s.`create_by`),
  s.`create_account` = IF(COALESCE(s.`create_account`, '') = '', COALESCE(NULLIF(s.`user_account`, ''), 'system'), s.`create_account`),
  s.`update_by` = IF(COALESCE(s.`update_by`, 0) = 0, COALESCE(u.`id`, 0), s.`update_by`),
  s.`update_account` = IF(COALESCE(s.`update_account`, '') = '', COALESCE(NULLIF(s.`user_account`, ''), 'system'), s.`update_account`),
  s.`create_dept_id` = IF(COALESCE(s.`create_dept_id`, 0) = 0, COALESCE(ud.`dept_id`, 0), s.`create_dept_id`);

UPDATE `dingtalk_h5_perf_reviews` r
LEFT JOIN `users` u ON BINARY u.`user_mini_openid` = BINARY r.`employee_account`
LEFT JOIN (
  SELECT `user_dept_user_id`, MIN(`user_dept_dept_id`) AS `dept_id`
  FROM `user_depts`
  GROUP BY `user_dept_user_id`
) ud ON ud.`user_dept_user_id` = u.`id`
SET
  r.`create_by` = IF(COALESCE(r.`create_by`, 0) = 0, COALESCE(u.`id`, 0), r.`create_by`),
  r.`create_account` = IF(COALESCE(r.`create_account`, '') = '', COALESCE(NULLIF(r.`employee_account`, ''), 'system'), r.`create_account`),
  r.`update_by` = IF(COALESCE(r.`update_by`, 0) = 0, COALESCE(u.`id`, 0), r.`update_by`),
  r.`update_account` = IF(COALESCE(r.`update_account`, '') = '', COALESCE(NULLIF(r.`employee_account`, ''), 'system'), r.`update_account`),
  r.`create_dept_id` = IF(COALESCE(r.`create_dept_id`, 0) = 0, COALESCE(ud.`dept_id`, 0), r.`create_dept_id`),
  r.`create_dept_path` = IF(COALESCE(r.`create_dept_path`, '') = '', COALESCE(r.`department`, ''), r.`create_dept_path`);

UPDATE `dingtalk_h5_perf_histories` h
LEFT JOIN `dingtalk_h5_perf_reviews` r ON r.`id` = h.`review_id`
LEFT JOIN `users` u ON BINARY u.`user_mini_openid` = BINARY h.`by_account`
LEFT JOIN (
  SELECT `user_dept_user_id`, MIN(`user_dept_dept_id`) AS `dept_id`
  FROM `user_depts`
  GROUP BY `user_dept_user_id`
) ud ON ud.`user_dept_user_id` = u.`id`
SET
  h.`edit_time` = IF(COALESCE(h.`edit_time`, 0) = 0, COALESCE(h.`add_time`, 0), h.`edit_time`),
  h.`create_by` = IF(COALESCE(h.`create_by`, 0) = 0, COALESCE(u.`id`, r.`create_by`, 0), h.`create_by`),
  h.`create_account` = IF(COALESCE(h.`create_account`, '') = '', COALESCE(NULLIF(h.`by_account`, ''), r.`create_account`, 'system'), h.`create_account`),
  h.`update_by` = IF(COALESCE(h.`update_by`, 0) = 0, COALESCE(u.`id`, r.`update_by`, r.`create_by`, 0), h.`update_by`),
  h.`update_account` = IF(COALESCE(h.`update_account`, '') = '', COALESCE(NULLIF(h.`by_account`, ''), r.`update_account`, r.`create_account`, 'system'), h.`update_account`),
  h.`create_dept_id` = IF(COALESCE(h.`create_dept_id`, 0) = 0, COALESCE(ud.`dept_id`, r.`create_dept_id`, 0), h.`create_dept_id`),
  h.`create_dept_path` = IF(COALESCE(h.`create_dept_path`, '') = '', COALESCE(r.`create_dept_path`, r.`department`, ''), h.`create_dept_path`);

UPDATE `dingtalk_h5_perf_templates`
SET
  `create_account` = IF(COALESCE(`create_account`, '') = '', 'system', `create_account`),
  `update_account` = IF(COALESCE(`update_account`, '') = '', 'system', `update_account`);
