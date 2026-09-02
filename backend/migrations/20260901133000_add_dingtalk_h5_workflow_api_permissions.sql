-- 注册钉钉 H5 通用流程审批 API 权限。
-- 仅创建权限定义，不为任何角色或用户授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('dingtalk_h5:api-category:workflow', 'OA 流程', 'dingtalk_h5', 'api_category',
   '', '', '', '',
   70, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:view', 'OA 流程查看接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/instances', '', 'workflow:view',
   10, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:start', 'OA 流程发起接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/instances', '', 'workflow:start',
   20, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:handle', 'OA 流程处理接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/tasks/:id/complete', '', 'workflow:handle',
   30, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:withdraw', 'OA 流程撤回接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/instances/:id/withdraw', '', 'workflow:withdraw',
   40, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), NOW(3), NOW(3))
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
