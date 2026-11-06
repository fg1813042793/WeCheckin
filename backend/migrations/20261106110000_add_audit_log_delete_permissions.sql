-- 日志删除为新增显式授权，不继承旧只读兼容入口 log:del 的历史授权。
START TRANSACTION;
UPDATE `permissions` SET `permission_status` = 0,
  `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, `updated_at` = NOW(3)
WHERE `permission_key` IN ('admin:menu:log:del', 'admin:api:log:del');

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`, `created_at`, `updated_at`
) VALUES
  ('admin:menu:log:delete', '删除操作日志', 'admin', 'button', 'admin:menu:log', '', '', 'log:delete', 6, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:log:delete', '操作日志删除接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/logs', '', 'log:delete', 6, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:log:security-delete', '删除登录日志', 'admin', 'button', 'admin:menu:log', '', '', 'log:security-delete', 7, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:log:security-delete', '登录日志删除接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/login-logs', '', 'log:security-delete', 7, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`), `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`), `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`), `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`), `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`), `updated_at` = NOW(3);
COMMIT;
