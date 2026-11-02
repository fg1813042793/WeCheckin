SET @workflow_instance_upgrade_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:workflow:instance:upgrade', '升级流程实例版本', 'admin', 'button',
   'admin:menu:workflow:instances', '', '', 'workflow:instance:upgrade',
   6, 1, @workflow_instance_upgrade_now_ms, @workflow_instance_upgrade_now_ms, NOW(3), NOW(3)),
  ('admin:api:workflow:instance:upgrade-preview', '流程实例版本升级预检接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-instances/:id/version-upgrade-preview', '', 'workflow:instance:upgrade:preview',
   161, 1, @workflow_instance_upgrade_now_ms, @workflow_instance_upgrade_now_ms, NOW(3), NOW(3)),
  ('admin:api:workflow:instance:upgrade', '流程实例版本升级接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-instances/:id/version-upgrade', '', 'workflow:instance:upgrade',
   162, 1, @workflow_instance_upgrade_now_ms, @workflow_instance_upgrade_now_ms, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`), `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`), `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`), `permission_icon` = VALUES(`permission_icon`),
  `permission_perms` = VALUES(`permission_perms`), `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1, `permission_edit_time` = VALUES(`permission_edit_time`), `updated_at` = NOW(3);
