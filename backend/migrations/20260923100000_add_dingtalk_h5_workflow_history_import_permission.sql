-- 注册 H5App 历史表单导入按钮，并继承已有流程发起权限的有效 allow 授权。

SET @workflow_history_import_permission_now = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES (
  'dingtalk_h5:button:workflow:history-import', '导入历史表单', 'dingtalk_h5', 'button',
  'dingtalk_h5:menu:workflow', 'workflow:history-import', '', 'workflow:history-import',
  74, 1, @workflow_history_import_permission_now, @workflow_history_import_permission_now, NOW(3), NOW(3)
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
  target_permission.`permission_key`,
  target_permission.`id`,
  'allow',
  source_grant.`grant_scope_value`,
  'h5-workflow-history-import',
  1,
  @workflow_history_import_permission_now,
  @workflow_history_import_permission_now,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_permission
  ON target_permission.`permission_key` = 'dingtalk_h5:button:workflow:history-import'
  AND target_permission.`permission_platform` = 'dingtalk_h5'
  AND target_permission.`permission_type` = 'button'
WHERE source_grant.`grant_permission_key` = 'dingtalk_h5:api:workflow:start'
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = VALUES(`grant_scope_value`),
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
