-- 补齐管理后台「钉钉应用管理 / 用户绑定管理」菜单权限。
--
-- 接口复用 dingtalk:config 权限；已有拥有钉钉配置权限的角色自动获得绑定管理入口。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'admin:menu:dingtalk:bindings', '用户绑定管理', 'admin', 'menu',
    'admin:menu:dingtalk', '/dingtalk/bindings', '', 'dingtalk:config',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:bindings:edit', '钉钉用户绑定维护', 'admin', 'button',
    'admin:menu:dingtalk:bindings', '', '', 'dingtalk:config',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
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
  config_grant.`grant_subject_type`,
  config_grant.`grant_subject_id`,
  binding_perm.`permission_key`,
  binding_perm.`id`,
  'allow',
  '',
  'setup-dingtalk-binding-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` config_grant
JOIN `permissions` binding_perm
  ON binding_perm.`permission_key` IN (
    'admin:menu:dingtalk:bindings',
    'admin:menu:dingtalk:bindings:edit'
  )
WHERE config_grant.`grant_subject_type` = 'role'
  AND config_grant.`grant_effect` = 'allow'
  AND config_grant.`grant_status` = 1
  AND config_grant.`grant_permission_key` IN (
    'admin:menu:dingtalk',
    'admin:menu:dingtalk:config',
    'admin:menu:dingtalk:config:edit',
    'admin:api:dingtalk:config'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);

UPDATE `permissions`
SET
  `permission_name` = '钉钉应用配置与绑定接口',
  `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  `updated_at` = NOW(3)
WHERE `permission_key` = 'admin:api:dingtalk:config';
