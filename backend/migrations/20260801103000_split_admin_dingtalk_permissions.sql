-- 拆分管理后台钉钉 H5 配置/绑定权限。
--
-- 旧版 dingtalk:config 同时控制配置页、企业应用、登录配置、应用展示配置和用户绑定。
-- 现在拆为 settings / bindings 的查看与维护权限，方便后台角色和用户单独授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'admin:menu:dingtalk:config', '配置选项', 'admin', 'menu',
    'admin:menu:dingtalk', '/dingtalk/config', '', 'dingtalk:settings:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:config:list', '钉钉配置查看', 'admin', 'button',
    'admin:menu:dingtalk:config', '', '', 'dingtalk:settings:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:config:edit', '钉钉配置保存', 'admin', 'button',
    'admin:menu:dingtalk:config', '', '', 'dingtalk:settings:edit',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:bindings', '用户绑定管理', 'admin', 'menu',
    'admin:menu:dingtalk', '/dingtalk/bindings', '', 'dingtalk:bindings:list',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:bindings:list', '钉钉用户绑定查看', 'admin', 'button',
    'admin:menu:dingtalk:bindings', '', '', 'dingtalk:bindings:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:bindings:edit', '钉钉用户绑定维护', 'admin', 'button',
    'admin:menu:dingtalk:bindings', '', '', 'dingtalk:bindings:edit',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api-category:dingtalk', '钉钉应用', 'admin', 'api_category',
    '', '', '', '',
    50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:settings:list', '钉钉配置查看接口', 'admin', 'api',
    'admin:api-category:dingtalk', '', '', 'dingtalk:settings:list',
    210, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:settings:edit', '钉钉配置保存接口', 'admin', 'api',
    'admin:api-category:dingtalk', '', '', 'dingtalk:settings:edit',
    220, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:bindings:list', '钉钉用户绑定查看接口', 'admin', 'api',
    'admin:api-category:dingtalk', '', '', 'dingtalk:bindings:list',
    230, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:bindings:edit', '钉钉用户绑定维护接口', 'admin', 'api',
    'admin:api-category:dingtalk', '', '', 'dingtalk:bindings:edit',
    240, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
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
  source_grant.`grant_subject_type`,
  source_grant.`grant_subject_id`,
  target_perm.`permission_key`,
  target_perm.`id`,
  'allow',
  '',
  'setup-dingtalk-split-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm
  ON target_perm.`permission_key` IN (
    'admin:menu:dingtalk:config:list',
    'admin:menu:dingtalk:config:edit',
    'admin:api:dingtalk:settings:list',
    'admin:api:dingtalk:settings:edit'
  )
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
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

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  source_grant.`grant_subject_type`,
  source_grant.`grant_subject_id`,
  target_perm.`permission_key`,
  target_perm.`id`,
  'allow',
  '',
  'setup-dingtalk-split-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm
  ON target_perm.`permission_key` IN (
    'admin:menu:dingtalk:bindings:list',
    'admin:menu:dingtalk:bindings:edit',
    'admin:api:dingtalk:bindings:list',
    'admin:api:dingtalk:bindings:edit'
  )
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
    'admin:menu:dingtalk:config',
    'admin:menu:dingtalk:config:edit',
    'admin:menu:dingtalk:bindings',
    'admin:menu:dingtalk:bindings:edit',
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
  `permission_name` = '旧版钉钉配置接口（已拆分）',
  `permission_status` = 0,
  `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  `updated_at` = NOW(3)
WHERE `permission_key` = 'admin:api:dingtalk:config';
