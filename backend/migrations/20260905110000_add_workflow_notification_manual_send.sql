-- 为流程通知 Outbox 保存流程单号快照，并登记单条手动发送接口。

SET @business_key_column_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND COLUMN_NAME = 'business_key'
);
SET @business_key_column_sql := IF(
  @business_key_column_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD COLUMN `business_key` VARCHAR(160) NOT NULL DEFAULT '''' COMMENT ''流程业务标识快照'' AFTER `instance_id`',
  'SELECT 1'
);
PREPARE business_key_column_stmt FROM @business_key_column_sql;
EXECUTE business_key_column_stmt;
DEALLOCATE PREPARE business_key_column_stmt;

UPDATE `workflow_notification_outbox` AS n
JOIN `workflow_process_instances` AS i
  ON n.`instance_id` COLLATE utf8mb4_general_ci = i.`id` COLLATE utf8mb4_general_ci
SET n.`business_key` = i.`business_key`
WHERE n.`business_key` = '';

SET @business_key_index_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND INDEX_NAME = 'idx_workflow_notification_business_key'
);
SET @business_key_index_sql := IF(
  @business_key_index_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD INDEX `idx_workflow_notification_business_key` (`business_key`)',
  'SELECT 1'
);
PREPARE business_key_index_stmt FROM @business_key_index_sql;
EXECUTE business_key_index_stmt;
DEALLOCATE PREPARE business_key_index_stmt;

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:api:workflow:notification:send', '流程通知单条手动发送接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-notifications/:id/send', '', 'workflow:notification:retry',
   425, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_icon` = VALUES(`permission_icon`),
  `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);

UPDATE `permissions`
SET `permission_name` = '流程通知发送与重试',
    `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    `updated_at` = NOW(3)
WHERE `permission_key` = 'admin:menu:workflow:notification:retry';
