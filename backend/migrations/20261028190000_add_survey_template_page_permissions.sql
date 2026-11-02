-- Register the Survey notification template page and preserve existing Survey role access.

SET @survey_template_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:survey:templates', '问卷通知模版', 'admin', 'menu',
   'admin:menu:survey', '/survey/templates', '', 'survey-template:list',
   4, 1, @survey_template_now_ms, @survey_template_now_ms, NOW(3), NOW(3)),
  ('admin:menu:survey:templates:edit', '问卷通知模版维护', 'admin', 'button',
   'admin:menu:survey:templates', '', '', 'survey-template:edit',
   1, 1, @survey_template_now_ms, @survey_template_now_ms, NOW(3), NOW(3)),
  ('admin:api:survey-template:list', '问卷通知模版查看接口', 'admin', 'api',
   'admin:api-category:survey', '/api/v2/admin/survey-notification-templates', '', 'survey-template:list',
   165, 1, @survey_template_now_ms, @survey_template_now_ms, NOW(3), NOW(3)),
  ('admin:api:survey-template:edit', '问卷通知模版维护接口', 'admin', 'api',
   'admin:api-category:survey', '/api/v2/admin/survey-notification-templates', '', 'survey-template:edit',
   166, 1, @survey_template_now_ms, @survey_template_now_ms, NOW(3), NOW(3))
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
  source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '',
  'survey-template-menu-backfill', 1,
  @survey_template_now_ms, @survey_template_now_ms, NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'admin:menu:survey' AS source_key, 'admin:menu:survey:templates' AS target_key
  UNION ALL SELECT 'admin:menu:survey:list', 'admin:menu:survey:templates'
  UNION ALL SELECT 'admin:menu:survey:button:list', 'admin:menu:survey:templates'
  UNION ALL SELECT 'admin:api:survey:list', 'admin:api:survey-template:list'
  UNION ALL SELECT 'admin:menu:survey:edit', 'admin:menu:survey:templates:edit'
  UNION ALL SELECT 'admin:api:survey:edit', 'admin:api:survey-template:edit'
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
