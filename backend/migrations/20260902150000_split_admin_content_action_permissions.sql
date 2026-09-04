-- 将内容管理按钮权限同步为独立 API 权限，并补充后台配置读取、文件上传权限。
-- 历史主体从原有粗粒度 API 权限继承，避免迁移后已授权角色突然失去功能。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:api:enroll:status', '打卡状态管理接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/enrollments/:id/status', '', 'enroll:status', 311, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:vouch', '打卡推荐接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/enrollments/:id/recommendation', '', 'enroll:vouch', 312, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:export', '打卡导出接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/enrollments/:id/export', '', 'enroll:export', 313, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:users', '打卡参与用户接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/enrollments/:id/users', '', 'enroll:users', 314, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:news:status', '内容状态管理接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/news/:id/status', '', 'news:status', 321, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:news:vouch', '内容推荐接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/news/:id/recommendation', '', 'news:vouch', 322, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:status', '赛事活动状态管理接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/events/:id/status', '', 'event:status', 331, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:vouch', '赛事活动推荐接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/events/:id/recommendation', '', 'event:vouch', 332, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:top', '赛事活动置顶接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/events/:id/top', '', 'event:top', 333, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:users', '赛事活动参与用户接口', 'admin', 'api', 'admin:api-category:content', '/api/v2/admin/events/:id/participants', '', 'event:users', 334, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:setup:list', '系统配置查看接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/settings/content', '', 'setup:list', 401, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:upload:create', '后台文件上传接口', 'admin', 'api', 'admin:api-category:system', '/api/v2/admin/uploads', '', 'upload:create', 402, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
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
  source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '',
  'split-content-action-permissions', 1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'admin:api:enroll:edit' AS source_key, 'admin:api:enroll:status' AS target_key
  UNION ALL SELECT 'admin:api:enroll:edit', 'admin:api:enroll:vouch'
  UNION ALL SELECT 'admin:api:enroll:list', 'admin:api:enroll:export'
  UNION ALL SELECT 'admin:api:enroll:list', 'admin:api:enroll:users'
  UNION ALL SELECT 'admin:api:news:edit', 'admin:api:news:status'
  UNION ALL SELECT 'admin:api:news:edit', 'admin:api:news:vouch'
  UNION ALL SELECT 'admin:api:event:edit', 'admin:api:event:status'
  UNION ALL SELECT 'admin:api:event:edit', 'admin:api:event:vouch'
  UNION ALL SELECT 'admin:api:event:edit', 'admin:api:event:top'
  UNION ALL SELECT 'admin:api:event:list', 'admin:api:event:users'
  UNION ALL SELECT 'admin:api:setup:edit', 'admin:api:setup:list'
) mapping ON mapping.source_key = source_grant.`grant_permission_key`
JOIN `permissions` target_perm ON target_perm.`permission_key` = mapping.target_key
WHERE source_grant.`grant_effect` = 'allow' AND source_grant.`grant_status` = 1
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
  source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '',
  'admin-upload-permission-backfill', 1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm ON target_perm.`permission_key` = 'admin:api:upload:create'
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
    'admin:api:user:add', 'admin:api:user:edit',
    'admin:api:mgr:add', 'admin:api:mgr:edit',
    'admin:api:enroll:add', 'admin:api:enroll:edit',
    'admin:api:news:add', 'admin:api:news:edit',
    'admin:api:event:add', 'admin:api:event:edit',
    'admin:api:survey:add', 'admin:api:survey:edit',
    'admin:api:exam:add', 'admin:api:exam:edit',
    'admin:api:question-bank:add', 'admin:api:question-bank:edit'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
