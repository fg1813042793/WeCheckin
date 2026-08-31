-- 补齐管理后台「流程管理 / 组织审批身份设置」菜单与接口权限。
--
-- 该迁移用于已经执行过权限初始化或早期组织审批身份迁移的数据库。
-- 只注册权限定义，不在运行时自动补齐，也不自动给普通角色授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'admin:menu:workflow', '流程管理', 'admin', 'directory',
    '', '/workflow', 'Share', 'workflow:list',
    16, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:org-approvers', '组织审批身份设置', 'admin', 'menu',
    'admin:menu:workflow', '/workflow/org-approvers', '', 'workflow:org-approver:list',
    4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:org-approver:list', '组织审批身份查看', 'admin', 'button',
    'admin:menu:workflow:org-approvers', '', '', 'workflow:org-approver:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:org-approver:edit', '组织审批身份维护', 'admin', 'button',
    'admin:menu:workflow:org-approvers', '', '', 'workflow:org-approver:edit',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api-category:workflow', '流程管理', 'admin', 'api_category',
    '', '', '', '',
    75, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:org-approver:list', '组织审批身份查看接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-org-approver-identities', '', 'workflow:org-approver:list',
    410, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:org-approver:edit', '组织审批身份维护接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-org-approver-assignments', '', 'workflow:org-approver:edit',
    420, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  )
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
