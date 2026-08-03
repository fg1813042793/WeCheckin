-- 补齐管理后台「钉钉应用管理 / 配置选项」菜单与接口权限。
--
-- 代码声明会在新库初始化时写入这些权限；已有库的 bootstrap:seed_permissions 已经执行过，
-- 需要版本化迁移显式补齐权限定义与已有系统配置角色的授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'admin:menu:dingtalk', '钉钉应用管理', 'admin', 'directory',
    '', '/dingtalk', 'Connection', '',
    13, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:config', '配置选项', 'admin', 'menu',
    'admin:menu:dingtalk', '/dingtalk/config', '', 'dingtalk:config',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:config:edit', '钉钉配置保存', 'admin', 'button',
    'admin:menu:dingtalk:config', '', '', 'dingtalk:config',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api-category:dingtalk', '钉钉应用', 'admin', 'api_category',
    '', '', '', '',
    50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:config', '钉钉应用配置接口', 'admin', 'api',
    'admin:api-category:dingtalk', '', '', 'dingtalk:config',
    210, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
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

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  setup_grant.`grant_subject_type`,
  setup_grant.`grant_subject_id`,
  dingtalk_perm.`permission_key`,
  dingtalk_perm.`id`,
  'allow',
  '',
  'setup-dingtalk-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` setup_grant
JOIN `permissions` dingtalk_perm
  ON dingtalk_perm.`permission_key` IN (
    'admin:menu:dingtalk',
    'admin:menu:dingtalk:config',
    'admin:menu:dingtalk:config:edit',
    'admin:api:dingtalk:config'
  )
WHERE setup_grant.`grant_subject_type` = 'role'
  AND setup_grant.`grant_effect` = 'allow'
  AND setup_grant.`grant_status` = 1
  AND setup_grant.`grant_permission_key` IN (
    'admin:menu:setup',
    'admin:api:setup:edit'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
