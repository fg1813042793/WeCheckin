-- 注册通用 OA 流程客户端 API 权限与管理员取消实例权限。
-- 只补齐权限定义，不自动给普通角色授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('client:api-category:workflow', 'OA 流程', 'client', 'api_category',
   '', '', '', '',
   80, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:api:workflow:view', 'OA 流程查看接口', 'client', 'api',
   'client:api-category:workflow', '/api/v2/workflows/instances', '', 'workflow:view',
   10, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:api:workflow:start', 'OA 流程发起接口', 'client', 'api',
   'client:api-category:workflow', '/api/v2/workflows/instances', '', 'workflow:start',
   20, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:api:workflow:handle', 'OA 流程处理接口', 'client', 'api',
   'client:api-category:workflow', '/api/v2/workflows/tasks/:id/complete', '', 'workflow:handle',
   30, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:api:workflow:withdraw', 'OA 流程撤回接口', 'client', 'api',
   'client:api-category:workflow', '/api/v2/workflows/instances/:id/withdraw', '', 'workflow:withdraw',
   40, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:workflow:instance:cancel', '取消流程实例', 'admin', 'button',
   'admin:menu:workflow:instances', '', '', 'workflow:instance:cancel',
   4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:instance:cancel', '流程实例取消接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-instances/:id/cancel', '', 'workflow:instance:cancel',
   400, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
