-- Add an explicit permission for copying an exam definition without its records.

SET @exam_copy_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:exam:copy', '复制考试', 'admin', 'button',
   'admin:menu:exam', '', '', 'exam:copy',
   5, 1, @exam_copy_now_ms, @exam_copy_now_ms, NOW(3), NOW(3)),
  ('admin:api:exam:copy', '考试复制接口', 'admin', 'api',
   'admin:api-category:exam', '/api/v2/admin/exams/:id/copy', '', 'exam:copy',
   178, 1, @exam_copy_now_ms, @exam_copy_now_ms, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`), `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`), `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`), `permission_icon` = VALUES(`permission_icon`),
  `permission_perms` = VALUES(`permission_perms`), `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1, `permission_edit_time` = VALUES(`permission_edit_time`), `updated_at` = NOW(3);

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '', 'exam-copy-backfill', 1,
  @exam_copy_now_ms, @exam_copy_now_ms, NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'admin:menu:exam:add' AS source_key, 'admin:menu:exam:copy' AS target_key
  UNION ALL SELECT 'admin:menu:exam:edit', 'admin:menu:exam:copy'
  UNION ALL SELECT 'admin:api:exam:add', 'admin:api:exam:copy'
  UNION ALL SELECT 'admin:api:exam:edit', 'admin:api:exam:copy'
) mapping ON mapping.source_key = source_grant.`grant_permission_key`
JOIN `permissions` target_perm ON target_perm.`permission_key` = mapping.target_key
WHERE source_grant.`grant_effect` = 'allow' AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`), `grant_effect` = 'allow',
  `grant_scope_value` = '', `grant_source` = VALUES(`grant_source`), `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`), `updated_at` = NOW(3);
