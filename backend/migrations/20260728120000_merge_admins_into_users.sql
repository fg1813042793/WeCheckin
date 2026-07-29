-- 将后台管理员账号合并进 users 表。
--
-- 设计说明：
-- - users.user_admin_enabled/user_admin_type/user_account 为历史兼容字段，不再作为后台登录准入条件。
-- - 后台登录准入以 users.user_status、users.user_role_id、roles.role_status 和统一权限授权为准。
-- - 数据权限走 roles.role_data_scope、permission_grants、user_depts；旧 admin_depts 会合并进 user_depts。
-- - 旧 admins 表仅作为迁移来源保留，确认稳定后可单独清理。

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_account'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_account` varchar(100) DEFAULT NULL COMMENT ''后台登录账号'' AFTER `user_mini_openid`', 'SELECT ''user_account exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_admin_enabled'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_admin_enabled` tinyint DEFAULT 0 COMMENT ''是否后台账号'' AFTER `user_password`', 'SELECT ''user_admin_enabled exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_admin_type'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_admin_type` bigint DEFAULT 0 COMMENT ''后台账号类型:1超级管理员'' AFTER `user_admin_enabled`', 'SELECT ''user_admin_type exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_role_id'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_role_id` bigint unsigned DEFAULT 0 COMMENT ''后台角色ID'' AFTER `user_admin_type`', 'SELECT ''user_role_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_admin_desc'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_admin_desc` varchar(200) DEFAULT NULL COMMENT ''管理员描述'' AFTER `user_role_id`', 'SELECT ''user_admin_desc exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_admin_token'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_admin_token` varchar(100) DEFAULT NULL COMMENT ''后台登录token'' AFTER `user_admin_desc`', 'SELECT ''user_admin_token exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_admin_token_time'
);
SET @ddl := IF(@col_exists = 0, 'ALTER TABLE `users` ADD COLUMN `user_admin_token_time` bigint DEFAULT 0 COMMENT ''后台token生成时间'' AFTER `user_admin_token`', 'SELECT ''user_admin_token_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'idx_users_admin_account'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `users` ADD INDEX `idx_users_admin_account` (`user_admin_enabled`, `user_account`)', 'SELECT ''idx_users_admin_account exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'idx_users_admin_role'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `users` ADD INDEX `idx_users_admin_role` (`user_admin_enabled`, `user_role_id`)', 'SELECT ''idx_users_admin_role exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `admin_user_merge_maps` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '映射ID',
  `legacy_admin_id` bigint unsigned NOT NULL COMMENT '旧admins.id',
  `user_id` bigint unsigned NOT NULL COMMENT '新users.id',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_admin_user_merge_legacy` (`legacy_admin_id`),
  UNIQUE KEY `idx_admin_user_merge_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员合并用户映射表';

SET @admins_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admins'
);
SET @ddl := IF(
  @admins_table_exists > 0,
  CONCAT(
    'INSERT INTO `users` (`user_mini_openid`, `user_account`, `user_name`, `user_password`, `user_admin_desc`, `user_pic`, `user_mobile`, `user_status`, `user_admin_enabled`, `user_admin_type`, `user_role_id`, `user_admin_token`, `user_admin_token_time`, `user_login_cnt`, `user_login_time`, `user_add_time`, `user_edit_time`, `user_add_ip`, `user_edit_ip`, `created_at`, `updated_at`) ',
    'SELECT CONCAT(''admin:'', `id`), `admin_name`, COALESCE(NULLIF(`admin_desc`, ''''), `admin_name`), `admin_password`, `admin_desc`, `admin_pic`, `admin_phone`, `admin_status`, 1, `admin_type`, `admin_role_id`, `admin_token`, `admin_token_time`, `admin_login_cnt`, `admin_login_time`, `admin_add_time`, `admin_edit_time`, `admin_add_ip`, `admin_edit_ip`, ',
    'CASE WHEN `created_at` IS NULL OR CAST(`created_at` AS CHAR) IN (''0000-00-00'', ''0000-00-00 00:00:00'', ''0000-00-00 00:00:00.000'') THEN NOW(3) ELSE `created_at` END, ',
    'CASE WHEN `updated_at` IS NULL OR CAST(`updated_at` AS CHAR) IN (''0000-00-00'', ''0000-00-00 00:00:00'', ''0000-00-00 00:00:00.000'') THEN NOW(3) ELSE `updated_at` END ',
    'FROM `admins` ON DUPLICATE KEY UPDATE `user_account` = VALUES(`user_account`), `user_password` = VALUES(`user_password`), `user_admin_desc` = VALUES(`user_admin_desc`), `user_pic` = VALUES(`user_pic`), `user_mobile` = VALUES(`user_mobile`), `user_status` = VALUES(`user_status`), `user_admin_enabled` = 1, `user_admin_type` = VALUES(`user_admin_type`), `user_role_id` = VALUES(`user_role_id`), `user_admin_token` = VALUES(`user_admin_token`), `user_admin_token_time` = VALUES(`user_admin_token_time`), `user_login_cnt` = VALUES(`user_login_cnt`), `user_login_time` = VALUES(`user_login_time`), `user_edit_time` = VALUES(`user_edit_time`), `user_edit_ip` = VALUES(`user_edit_ip`)'
  ),
  'SELECT ''admins table missing'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF(
  @admins_table_exists > 0,
  'INSERT INTO `admin_user_merge_maps` (`legacy_admin_id`, `user_id`, `created_at`, `updated_at`) SELECT a.`id`, u.`id`, NOW(3), NOW(3) FROM `admins` a INNER JOIN `users` u ON CAST(u.`user_mini_openid` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci = CAST(CONCAT(''admin:'', a.`id`) AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci ON DUPLICATE KEY UPDATE `user_id` = VALUES(`user_id`), `updated_at` = NOW(3)',
  'SELECT ''admins table missing'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @admin_depts_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_depts'
);
SET @ddl := IF(
  @admin_depts_table_exists > 0,
  'INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`) SELECT DISTINCT m.`user_id`, ad.`admin_dept_dept_id`, NOW(3), NOW(3) FROM `admin_depts` ad INNER JOIN `admin_user_merge_maps` m ON ad.`admin_dept_admin_id` IN (m.`legacy_admin_id`, m.`user_id`) LEFT JOIN `user_depts` ud ON ud.`user_dept_user_id` = m.`user_id` AND ud.`user_dept_dept_id` = ad.`admin_dept_dept_id` WHERE ad.`admin_dept_dept_id` > 0 AND ud.`id` IS NULL',
  'SELECT ''admin_depts table missing'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @logs_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs'
);
SET @ddl := IF(
  @logs_table_exists > 0,
  'UPDATE `logs` l INNER JOIN `admin_user_merge_maps` m ON CAST(l.`log_admin_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci = CAST(m.`legacy_admin_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci LEFT JOIN `admin_user_merge_maps` existing ON CAST(l.`log_admin_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci = CAST(existing.`user_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci SET l.`log_admin_id` = CAST(m.`user_id` AS CHAR) WHERE existing.`user_id` IS NULL',
  'SELECT ''logs table missing'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
