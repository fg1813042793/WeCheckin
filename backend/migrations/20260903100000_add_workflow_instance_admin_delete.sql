-- 管理员删除流程实例采用软删除，保留任务、通知和流转历史用于审计。
-- 删除权限不自动授予现有角色或用户。

ALTER TABLE `workflow_process_instances`
  ADD COLUMN `admin_deleted_at` bigint NOT NULL DEFAULT 0 COMMENT '管理员删除时间' AFTER `starter_deleted_at`,
  ADD COLUMN `admin_deleted_by` varchar(64) NOT NULL DEFAULT '' COMMENT '删除操作管理员ID' AFTER `admin_deleted_at`,
  ADD INDEX `idx_workflow_instances_admin_deleted_time` (`admin_deleted_at`, `start_time`);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:workflow:instance:delete', '删除流程实例', 'admin', 'button',
   'admin:menu:workflow:instances', '', '', 'workflow:instance:delete',
   5, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:instance:delete', '流程实例删除接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-instances/:id', '', 'workflow:instance:delete',
   410, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
