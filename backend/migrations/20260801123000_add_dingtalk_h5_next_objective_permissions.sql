-- 补齐钉钉 H5「下月目标」操作按钮权限。
--
-- 下月目标仍通过员工自评保存/提交接口持久化；这里新增的是按钮级权限，
-- 并给已拥有「保存员工自评」或「提交员工自评」按钮权限的主体自动回填授权。

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  (
    'dingtalk_h5:button:review:next_objective_edit', '编辑下月目标', 'dingtalk_h5', 'button',
    'dingtalk_h5:menu:performance:mine', 'review:next_objective_edit', '', '',
    34, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    NOW(3), NOW(3)
  ),
  (
    'dingtalk_h5:button:review:next_objective_add', '新增下月目标', 'dingtalk_h5', 'button',
    'dingtalk_h5:menu:performance:mine', 'review:next_objective_add', '', '',
    35, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    NOW(3), NOW(3)
  ),
  (
    'dingtalk_h5:button:review:next_objective_delete', '删除下月目标', 'dingtalk_h5', 'button',
    'dingtalk_h5:menu:performance:mine', 'review:next_objective_delete', '', '',
    36, 1, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    NOW(3), NOW(3)
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
  source_grant.`grant_subject_type`,
  source_grant.`grant_subject_id`,
  target_perm.`permission_key`,
  target_perm.`id`,
  'allow',
  '',
  'dingtalk-h5-next-objective-backfill',
  1,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm
  ON target_perm.`permission_key` IN (
    'dingtalk_h5:button:review:next_objective_edit',
    'dingtalk_h5:button:review:next_objective_add',
    'dingtalk_h5:button:review:next_objective_delete'
  )
WHERE source_grant.`grant_subject_type` IN ('role', 'user')
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
  AND source_grant.`grant_permission_key` IN (
    'dingtalk_h5:button:review:self_save',
    'dingtalk_h5:button:review:self_submit'
  )
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
