-- 增加客户端鉴权上传权限，并从已有资料、打卡、活动、问卷和考试写权限继承授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES (
  'client:api:upload:create', '文件上传接口', 'client', 'api',
  'client:api-category:user', '/api/v2/uploads', '', 'upload:create',
  50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)
)
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
  'client-upload-permission-backfill', 1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm ON target_perm.`permission_key` = 'client:api:upload:create'
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
    'client:api:user:edit',
    'client:api:enroll:submit',
    'client:api:event:dynamic',
    'client:api:survey:response',
    'client:api:exam:answer'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
