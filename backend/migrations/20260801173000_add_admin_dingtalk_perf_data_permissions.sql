-- 增加钉钉 H5 绩效考评单与流转记录的后台数据管理权限。
-- 只新增权限点并从已有钉钉应用授权做兼容回填，不修改历史迁移。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'admin:menu:dingtalk:perf-reviews', '绩效考评单', 'admin', 'menu',
    'admin:menu:dingtalk', '/dingtalk/perf-reviews', '', 'dingtalk:perf-reviews:list',
    3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:perf-histories', '绩效流转记录', 'admin', 'menu',
    'admin:menu:dingtalk', '/dingtalk/perf-histories', '', 'dingtalk:perf-histories:list',
    4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:perf-reviews:list', '绩效考评单查看', 'admin', 'button',
    'admin:menu:dingtalk:perf-reviews', '', '', 'dingtalk:perf-reviews:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:perf-reviews:detail', '绩效考评单详情', 'admin', 'button',
    'admin:menu:dingtalk:perf-reviews', '', '', 'dingtalk:perf-reviews:detail',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:perf-reviews:del', '绩效考评单删除', 'admin', 'button',
    'admin:menu:dingtalk:perf-reviews', '', '', 'dingtalk:perf-reviews:del',
    3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:menu:dingtalk:perf-histories:list', '绩效流转记录查看', 'admin', 'button',
    'admin:menu:dingtalk:perf-histories', '', '', 'dingtalk:perf-histories:list',
    1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:perf-reviews:list', '绩效考评单查看接口', 'admin', 'api',
    'admin:api-category:dingtalk', '/api/v2/admin/dingtalk/perf-reviews', '', 'dingtalk:perf-reviews:list',
    250, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:perf-reviews:detail', '绩效考评单详情接口', 'admin', 'api',
    'admin:api-category:dingtalk', '/api/v2/admin/dingtalk/perf-reviews/:id', '', 'dingtalk:perf-reviews:detail',
    260, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:perf-reviews:del', '绩效考评单删除接口', 'admin', 'api',
    'admin:api-category:dingtalk', '/api/v2/admin/dingtalk/perf-reviews/:id', '', 'dingtalk:perf-reviews:del',
    270, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:perf-histories:list', '绩效流转记录查看接口', 'admin', 'api',
    'admin:api-category:dingtalk', '/api/v2/admin/dingtalk/perf-histories', '', 'dingtalk:perf-histories:list',
    280, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
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
  'setup-dingtalk-perf-data-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm
  ON target_perm.`permission_key` IN (
    'admin:menu:dingtalk:perf-reviews',
    'admin:menu:dingtalk:perf-reviews:list',
    'admin:menu:dingtalk:perf-reviews:detail',
    'admin:menu:dingtalk:perf-reviews:del',
    'admin:api:dingtalk:perf-reviews:list',
    'admin:api:dingtalk:perf-reviews:detail',
    'admin:api:dingtalk:perf-reviews:del',
    'admin:menu:dingtalk:perf-histories',
    'admin:menu:dingtalk:perf-histories:list',
    'admin:api:dingtalk:perf-histories:list'
  )
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
    'admin:menu:dingtalk',
    'admin:menu:dingtalk:config',
    'admin:menu:dingtalk:config:edit',
    'admin:menu:dingtalk:bindings',
    'admin:menu:dingtalk:bindings:edit',
    'admin:api:dingtalk:settings:list',
    'admin:api:dingtalk:settings:edit',
    'admin:api:dingtalk:bindings:list',
    'admin:api:dingtalk:bindings:edit'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);

