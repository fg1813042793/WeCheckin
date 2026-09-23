-- 将钉钉 H5 通用 OA 流程处理权限拆分为通过、驳回、退回和提交办理。
-- 历史主体继承原 workflow:handle 的有效授权，避免升级后已有角色失去处理能力。

SET @workflow_action_permission_now = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('dingtalk_h5:button:workflow:approve', '通过审批', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', 'workflow:approve', '', 'workflow:approve',
   70, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:button:workflow:reject', '驳回流程', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', 'workflow:reject', '', 'workflow:reject',
   71, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:button:workflow:return', '退回流程', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', 'workflow:return', '', 'workflow:return',
   72, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:button:workflow:submit', '提交办理', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', 'workflow:submit', '', 'workflow:submit',
   73, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:approve', 'OA 流程通过接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/tasks/:id/complete', '', 'workflow:approve',
   30, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:reject', 'OA 流程驳回接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/tasks/:id/complete', '', 'workflow:reject',
   31, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:return', 'OA 流程退回接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/tasks/:id/complete', '', 'workflow:return',
   32, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:submit', 'OA 流程提交办理接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/tasks/:id/complete', '', 'workflow:submit',
   33, 1, @workflow_action_permission_now, @workflow_action_permission_now, NOW(3), NOW(3))
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
  source_grant.`grant_effect`,
  source_grant.`grant_scope_value`,
  'h5-workflow-action-split',
  source_grant.`grant_status`,
  @workflow_action_permission_now,
  @workflow_action_permission_now,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'dingtalk_h5:button:workflow:approve' AS target_key
  UNION ALL SELECT 'dingtalk_h5:button:workflow:reject'
  UNION ALL SELECT 'dingtalk_h5:button:workflow:return'
  UNION ALL SELECT 'dingtalk_h5:button:workflow:submit'
  UNION ALL SELECT 'dingtalk_h5:api:workflow:approve'
  UNION ALL SELECT 'dingtalk_h5:api:workflow:reject'
  UNION ALL SELECT 'dingtalk_h5:api:workflow:return'
  UNION ALL SELECT 'dingtalk_h5:api:workflow:submit'
) mapping
JOIN `permissions` target_perm ON target_perm.`permission_key` = mapping.target_key
WHERE source_grant.`grant_permission_key` = 'dingtalk_h5:api:workflow:handle'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = VALUES(`grant_effect`),
  `grant_scope_value` = VALUES(`grant_scope_value`),
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = VALUES(`grant_status`),
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);

UPDATE `permissions`
SET `permission_status` = 0,
    `permission_edit_time` = @workflow_action_permission_now,
    `updated_at` = NOW(3)
WHERE `permission_key` = 'dingtalk_h5:api:workflow:handle';
