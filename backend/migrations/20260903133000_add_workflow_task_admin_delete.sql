-- 管理员删除流程任务采用软删除，保留流程实例详情和流转历史用于审计。
-- 删除权限不自动授予现有角色或用户。

SET @workflow_task_admin_deleted_at_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_tasks'
    AND COLUMN_NAME = 'admin_deleted_at'
);
SET @ddl := IF(
  @workflow_task_admin_deleted_at_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `admin_deleted_at` bigint NOT NULL DEFAULT 0 COMMENT ''管理员删除时间'' AFTER `handled_at`',
  'SELECT ''workflow_process_tasks.admin_deleted_at exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @workflow_task_admin_deleted_by_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_tasks'
    AND COLUMN_NAME = 'admin_deleted_by'
);
SET @ddl := IF(
  @workflow_task_admin_deleted_by_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `admin_deleted_by` varchar(64) NOT NULL DEFAULT '''' COMMENT ''删除操作管理员ID'' AFTER `admin_deleted_at`',
  'SELECT ''workflow_process_tasks.admin_deleted_by exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @workflow_task_admin_deleted_index_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_tasks'
    AND INDEX_NAME = 'idx_workflow_tasks_admin_deleted_time'
);
SET @ddl := IF(
  @workflow_task_admin_deleted_index_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD INDEX `idx_workflow_tasks_admin_deleted_time` (`admin_deleted_at`, `created_at`)',
  'SELECT ''idx_workflow_tasks_admin_deleted_time exists'''
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
  ('admin:menu:workflow:task:delete', '删除流程任务', 'admin', 'button',
   'admin:menu:workflow:tasks', '', '', 'workflow:task:delete',
   3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:task:delete', '流程任务删除接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-tasks/:id', '', 'workflow:task:delete',
   420, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
