-- Allow authorized administrators to remove terminal scheduled-task run history.

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:scheduled-task:run:delete', '运行记录批量删除', 'admin', 'button',
   'admin:menu:scheduled-task:runs', '', '', 'scheduled-task:run:delete',
   2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:scheduled-task:run:delete', '定时任务运行记录删除接口', 'admin', 'api',
   'admin:api-category:scheduled-task', '/api/v2/admin/scheduled-task-runs', '', 'scheduled-task:run:delete',
   75, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = VALUES(`updated_at`);
