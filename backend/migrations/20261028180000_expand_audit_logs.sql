-- 扩展既有 logs 表，保留旧字段及操作者快照。仅在维护窗口由迁移工具执行。
-- DDL 在 MySQL 中隐式提交；按列/索引检查支持失败后重跑，索引创建应评估历史表大小。
-- event_id 允许 NULL 兼容滚动发布的旧写入方；新存储入口总是生成唯一非空值。
-- 再次执行仅回填未带事件 ID 的旧记录，绝不将未知结果推断为成功。

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_browser');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_browser` varchar(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_device_type');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_device_type` varchar(32) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_platform_type');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_platform_type` varchar(32) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_target_ids');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_target_ids` json NULL', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_requested_count');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_requested_count` int NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_event_id'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_event_id` varchar(64) NULL DEFAULT NULL', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_category'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_category` varchar(16) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_source'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_source` varchar(16) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_module'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_module` varchar(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_action'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_action` varchar(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_outcome'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_outcome` varchar(16) NOT NULL DEFAULT ''unknown''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_reason_code'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_reason_code` varchar(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_reason'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_reason` varchar(1000) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_actor_type'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_actor_type` varchar(16) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_attempted_account'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_attempted_account` varchar(200) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_target_type'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_target_type` varchar(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_target_id'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_target_id` varchar(128) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_target_name'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_target_name` varchar(200) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_request_id'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_request_id` varchar(128) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_method'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_method` varchar(16) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_path'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_path` varchar(1000) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_user_agent'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_user_agent` varchar(1000) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_related_type'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_related_type` varchar(64) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_related_id'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_related_id` varchar(128) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_affected_count'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_affected_count` bigint NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_duration_ms'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_duration_ms` bigint NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_changes'
);
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_changes` json NULL', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `logs`
SET `log_event_id` = CONCAT('legacy-', `id`),
    `log_category` = CASE WHEN `log_type` = 1 THEN 'login' ELSE 'operation' END,
    `log_source` = 'admin',
    `log_actor_type` = 'admin',
    `log_outcome` = 'unknown',
    `log_changes` = COALESCE(`log_changes`, JSON_ARRAY())
WHERE `log_event_id` IS NULL OR `log_event_id` = '';

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'uidx_logs_event_id'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD UNIQUE INDEX `uidx_logs_event_id` (`log_event_id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_category_time'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_category_time` (`log_category`, `add_time`, `id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_actor_category_time'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_actor_category_time` (`create_by`, `log_category`, `add_time`, `id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_source_outcome_time'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_source_outcome_time` (`log_category`, `log_source`, `log_outcome`, `add_time`, `id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_module_action_time'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_module_action_time` (`log_category`, `log_module`, `log_action`, `add_time`, `id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_request_id'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_request_id` (`log_request_id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_target'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_target` (`log_target_type`, `log_target_id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = 'idx_logs_related'
);
SET @ddl := IF(@index_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_related` (`log_related_type`, `log_related_id`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 仅声明新菜单动作/API；不从旧 log:list 或 log:del 权限复制任何授权。
-- 安全日志和两类导出均需要显式授权，超级管理员继续使用既有绕过机制。
START TRANSACTION;
UPDATE `permissions` SET `permission_status` = 0, `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, `updated_at` = NOW(3)
WHERE `permission_key` = 'admin:menu:log:del';
INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES

  ('admin:api:log:security', '登录安全日志接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/login-logs', '', 'log:security', 3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:log:security', '登录安全日志', 'admin', 'button', 'admin:menu:log', '', '', 'log:security', 3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:log:export', '操作日志导出接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/logs/export', '', 'log:export', 4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:log:export', '操作日志导出', 'admin', 'button', 'admin:menu:log', '', '', 'log:export', 4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:log:security-export', '登录安全日志导出接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/login-logs/export', '', 'log:security-export', 5, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:log:security-export', '登录安全日志导出', 'admin', 'button', 'admin:menu:log', '', '', 'log:security-export', 5, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))

ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);
COMMIT;
