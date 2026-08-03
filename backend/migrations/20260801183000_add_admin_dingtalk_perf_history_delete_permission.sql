-- 增加管理后台「绩效流转记录」删除按钮与接口权限。
-- 仅对已经拥有绩效考评单删除权限的主体做兼容回填，避免把查看权限提升为删除权限。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'admin:menu:dingtalk:perf-histories:del', '绩效流转记录删除', 'admin', 'button',
    'admin:menu:dingtalk:perf-histories', '', '', 'dingtalk:perf-histories:del',
    2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3), NOW(3)
  ),
  (
    'admin:api:dingtalk:perf-histories:del', '绩效流转记录删除接口', 'admin', 'api',
    'admin:api-category:dingtalk', '/api/v2/admin/dingtalk/perf-histories/:id', '', 'dingtalk:perf-histories:del',
    290, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
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
  'setup-dingtalk-perf-history-delete-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm
  ON target_perm.`permission_key` IN (
    'admin:menu:dingtalk:perf-histories:del',
    'admin:api:dingtalk:perf-histories:del'
  )
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
    'admin:menu:dingtalk:perf-reviews:del',
    'admin:api:dingtalk:perf-reviews:del'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
