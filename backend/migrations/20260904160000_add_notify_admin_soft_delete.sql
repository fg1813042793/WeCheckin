-- 后台通知记录采用独立软删除，不影响接收人已经收到的站内信。
-- 删除属于破坏性管理能力，仅注册权限，不自动扩大现有角色或用户授权。

SET @admin_deleted_at_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND COLUMN_NAME = 'notify_admin_deleted_at'
);
SET @ddl := IF(
  @admin_deleted_at_exists = 0,
  'ALTER TABLE `notify` ADD COLUMN `notify_admin_deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''管理员删除时间，0表示未删除'' AFTER `notify_deleted_at`',
  'SELECT ''notify.notify_admin_deleted_at exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @admin_deleted_by_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND COLUMN_NAME = 'notify_admin_deleted_by'
);
SET @ddl := IF(
  @admin_deleted_by_exists = 0,
  'ALTER TABLE `notify` ADD COLUMN `notify_admin_deleted_by` VARCHAR(64) NOT NULL DEFAULT '''' COMMENT ''删除操作管理员ID'' AFTER `notify_admin_deleted_at`',
  'SELECT ''notify.notify_admin_deleted_by exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @admin_deleted_index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND INDEX_NAME = 'idx_notify_admin_deleted_id'
);
SET @ddl := IF(
  @admin_deleted_index_exists = 0,
  'ALTER TABLE `notify` ADD INDEX `idx_notify_admin_deleted_id` (`notify_admin_deleted_at`, `notify_id`)',
  'SELECT ''idx_notify_admin_deleted_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:notification:delete', '删除通知记录', 'admin', 'button',
   'admin:menu:notification', '', '', 'notification:delete',
   7, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:notification:delete', '通知记录删除接口', 'admin', 'api',
   'admin:api-category:notification', '/api/v2/admin/in-app-notifications/:id', '', 'notification:delete',
   70, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
