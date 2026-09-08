-- 流程通知投递记录采用管理端软删除，保留删除人和删除时间用于审计。
-- 删除权限为独立的破坏性能力，仅登记权限，不扩大已有角色授权。

SET @workflow_notification_admin_deleted_at_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND COLUMN_NAME = 'admin_deleted_at'
);
SET @workflow_notification_admin_deleted_at_sql := IF(
  @workflow_notification_admin_deleted_at_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD COLUMN `admin_deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''管理员删除时间，0表示未删除'' AFTER `edit_time`',
  'SELECT 1'
);
PREPARE workflow_notification_admin_deleted_at_stmt FROM @workflow_notification_admin_deleted_at_sql;
EXECUTE workflow_notification_admin_deleted_at_stmt;
DEALLOCATE PREPARE workflow_notification_admin_deleted_at_stmt;

SET @workflow_notification_admin_deleted_by_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND COLUMN_NAME = 'admin_deleted_by'
);
SET @workflow_notification_admin_deleted_by_sql := IF(
  @workflow_notification_admin_deleted_by_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD COLUMN `admin_deleted_by` VARCHAR(64) NOT NULL DEFAULT '''' COMMENT ''删除操作管理员ID'' AFTER `admin_deleted_at`',
  'SELECT 1'
);
PREPARE workflow_notification_admin_deleted_by_stmt FROM @workflow_notification_admin_deleted_by_sql;
EXECUTE workflow_notification_admin_deleted_by_stmt;
DEALLOCATE PREPARE workflow_notification_admin_deleted_by_stmt;

SET @workflow_notification_admin_deleted_index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND INDEX_NAME = 'idx_workflow_notification_admin_deleted_time'
);
SET @workflow_notification_admin_deleted_index_sql := IF(
  @workflow_notification_admin_deleted_index_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD INDEX `idx_workflow_notification_admin_deleted_time` (`admin_deleted_at`, `add_time`, `id`)',
  'SELECT 1'
);
PREPARE workflow_notification_admin_deleted_index_stmt FROM @workflow_notification_admin_deleted_index_sql;
EXECUTE workflow_notification_admin_deleted_index_stmt;
DEALLOCATE PREPARE workflow_notification_admin_deleted_index_stmt;

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:workflow:notification:delete', '删除流程通知投递记录', 'admin', 'button',
   'admin:menu:workflow:instances', '', '', 'workflow:notification:delete',
   8, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:notification:delete', '流程通知投递记录删除接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-notifications/:id', '', 'workflow:notification:delete',
   440, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
