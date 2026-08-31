-- 为已完成权限初始化的数据库补齐流程设计器权限定义。
-- 本迁移不自动给普通角色授权；超级管理员可直接访问，其他角色需在后台显式勾选。

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
    'admin:menu:workflow:definitions', '流程定义', 'admin', 'menu',
    'admin:menu:workflow', '/workflow/definitions', '', 'workflow:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:list', '流程定义查看', 'admin', 'button',
    'admin:menu:workflow:definitions', '', '', 'workflow:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:add', '流程定义创建', 'admin', 'button',
    'admin:menu:workflow:definitions', '', '', 'workflow:add',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:edit', '流程定义编辑', 'admin', 'button',
    'admin:menu:workflow:definitions', '', '', 'workflow:edit',
    3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:publish', '流程定义发布', 'admin', 'button',
    'admin:menu:workflow:definitions', '', '', 'workflow:publish',
    4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:workflow:del', '流程定义删除', 'admin', 'button',
    'admin:menu:workflow:definitions', '', '', 'workflow:del',
    5, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api-category:workflow', '流程管理', 'admin', 'api_category',
    '', '', '', '',
    75, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:list', '流程定义查看接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-definitions', '', 'workflow:list',
    300, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:add', '流程定义创建接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-definitions', '', 'workflow:add',
    310, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:edit', '流程定义编辑接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-definitions/:id', '', 'workflow:edit',
    320, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:publish', '流程定义发布接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-definitions/:id/publish', '', 'workflow:publish',
    330, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:workflow:del', '流程定义删除接口', 'admin', 'api',
    'admin:api-category:workflow', '/api/v2/admin/workflow-definitions/:id', '', 'workflow:del',
    340, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
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
