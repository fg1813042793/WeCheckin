-- 补齐钉钉 H5「流程制定」页面所需的人员接口授权。
--
-- 背景：
-- - 历史角色/用户可能只授予了 dingtalk_h5:menu:performance:org 菜单权限。
-- - 前端按统一接口权限控制，不具备 dingtalk_h5:api:user:list 时不会加载人员列表。
-- - 因此即使用户配置了 data:extra 额外数据权限，流程制定页也看不到扩展范围内人员。
-- - 本迁移只从已有钉钉 H5 菜单/按钮授权反推必要接口授权，不扩大数据权限范围。

SET @now_ms = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000;

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  source_grant.`grant_subject_type`,
  source_grant.`grant_subject_id`,
  api_perm.`permission_key`,
  api_perm.`id`,
  'allow',
  '',
  'h5-menu-api-backfill',
  1,
  @now_ms,
  @now_ms,
  NOW(3),
  NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'dingtalk_h5:menu:performance:org' AS source_key, 'dingtalk_h5:api:user:list' AS api_key
  UNION ALL
  SELECT 'dingtalk_h5:button:user:config' AS source_key, 'dingtalk_h5:api:user:list' AS api_key
  UNION ALL
  SELECT 'dingtalk_h5:button:user:config' AS source_key, 'dingtalk_h5:api:user:edit' AS api_key
) required_api
  ON required_api.source_key = source_grant.`grant_permission_key`
JOIN `permissions` api_perm
  ON api_perm.`permission_key` = required_api.api_key
  AND api_perm.`permission_platform` = 'dingtalk_h5'
  AND api_perm.`permission_type` = 'api'
WHERE source_grant.`grant_subject_type` IN ('role', 'user')
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
