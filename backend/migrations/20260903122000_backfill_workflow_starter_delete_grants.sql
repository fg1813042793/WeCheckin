-- 已有流程查看权限的角色和用户应具备删除本人终态申请的接口权限。
-- 删除服务仍会校验发起人身份和实例终态，不会扩大可删除的数据范围。

SET @now_ms = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000;

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  source_grant.`grant_subject_type`,
  source_grant.`grant_subject_id`,
  delete_perm.`permission_key`,
  delete_perm.`id`,
  'allow',
  '',
  'workflow-view-delete-backfill',
  1,
  @now_ms,
  @now_ms,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` delete_perm
  ON delete_perm.`permission_key` = 'dingtalk_h5:api:workflow:delete'
  AND delete_perm.`permission_platform` = 'dingtalk_h5'
  AND delete_perm.`permission_type` = 'api'
WHERE source_grant.`grant_permission_key` = 'dingtalk_h5:api:workflow:view'
  AND source_grant.`grant_subject_type` IN ('role', 'user')
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
