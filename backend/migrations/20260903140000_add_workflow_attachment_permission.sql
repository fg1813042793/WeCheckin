-- 注册 H5App 工作流附件上传权限，并继承已有发起或处理权限的授权对象。

SET @now_ms = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000;

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('dingtalk_h5:api:workflow:attachment', 'OA 流程附件上传接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/attachments', '', 'workflow:attachment',
   35, 1, @now_ms, @now_ms, NOW(3), NOW(3))
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
  source_grant.`grant_subject_type`,
  source_grant.`grant_subject_id`,
  attachment_perm.`permission_key`,
  attachment_perm.`id`,
  'allow',
  '',
  'workflow-write-attachment-backfill',
  1,
  @now_ms,
  @now_ms,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` attachment_perm
  ON attachment_perm.`permission_key` = 'dingtalk_h5:api:workflow:attachment'
  AND attachment_perm.`permission_platform` = 'dingtalk_h5'
  AND attachment_perm.`permission_type` = 'api'
WHERE source_grant.`grant_permission_key` IN (
    'dingtalk_h5:api:workflow:start',
    'dingtalk_h5:api:workflow:handle'
  )
  AND source_grant.`grant_subject_type` IN ('role', 'user')
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
