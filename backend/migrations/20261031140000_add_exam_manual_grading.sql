-- 人工批改私有快照与结果元数据。每列独立检查，允许重跑或从部分完成状态恢复。
SET @exam_grading_column_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND COLUMN_NAME = 'exam_r_review_snapshot'
);
SET @exam_grading_ddl := IF(@exam_grading_column_exists = 0,
  'ALTER TABLE `exam_record` ADD COLUMN `exam_r_review_snapshot` MEDIUMTEXT NULL',
  'SELECT ''exam_record.exam_r_review_snapshot exists''');
PREPARE exam_grading_stmt FROM @exam_grading_ddl;
EXECUTE exam_grading_stmt;
DEALLOCATE PREPARE exam_grading_stmt;

SET @exam_grading_column_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND COLUMN_NAME = 'exam_r_graded_by'
);
SET @exam_grading_ddl := IF(@exam_grading_column_exists = 0,
  'ALTER TABLE `exam_record` ADD COLUMN `exam_r_graded_by` BIGINT UNSIGNED NOT NULL DEFAULT 0',
  'SELECT ''exam_record.exam_r_graded_by exists''');
PREPARE exam_grading_stmt FROM @exam_grading_ddl;
EXECUTE exam_grading_stmt;
DEALLOCATE PREPARE exam_grading_stmt;

SET @exam_grading_column_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND COLUMN_NAME = 'exam_r_graded_at'
);
SET @exam_grading_ddl := IF(@exam_grading_column_exists = 0,
  'ALTER TABLE `exam_record` ADD COLUMN `exam_r_graded_at` BIGINT NOT NULL DEFAULT 0',
  'SELECT ''exam_record.exam_r_graded_at exists''');
PREPARE exam_grading_stmt FROM @exam_grading_ddl;
EXECUTE exam_grading_stmt;
DEALLOCATE PREPARE exam_grading_stmt;

SET @exam_grading_column_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND COLUMN_NAME = 'exam_r_grading_comment'
);
SET @exam_grading_ddl := IF(@exam_grading_column_exists = 0,
  'ALTER TABLE `exam_record` ADD COLUMN `exam_r_grading_comment` VARCHAR(1000) NOT NULL DEFAULT ''''',
  'SELECT ''exam_record.exam_r_grading_comment exists''');
PREPARE exam_grading_stmt FROM @exam_grading_ddl;
EXECUTE exam_grading_stmt;
DEALLOCATE PREPARE exam_grading_stmt;

SET @exam_manual_grading_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:exam:record-grade', '考试答卷批改', 'admin', 'button',
   'admin:menu:exam', '', '', 'exam-record:grade',
   7, 1, @exam_manual_grading_now_ms, @exam_manual_grading_now_ms, NOW(3), NOW(3)),
  ('admin:api:exam-record:grade', '考试答卷批改接口', 'admin', 'api',
   'admin:api-category:exam', '/api/v2/admin/exams/:id/records/:recordId/grading', '', 'exam-record:grade',
   179, 1, @exam_manual_grading_now_ms, @exam_manual_grading_now_ms, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`), `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`), `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`), `permission_icon` = VALUES(`permission_icon`),
  `permission_perms` = VALUES(`permission_perms`), `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1, `permission_edit_time` = VALUES(`permission_edit_time`), `updated_at` = NOW(3);

-- 只为对应的考试编辑 allow 主体补齐；目标已有记录（含 deny/停用）一律保留。
INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', source_grant.`grant_scope_value`, 'exam-manual-grading-backfill', 1,
  @exam_manual_grading_now_ms, @exam_manual_grading_now_ms, NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'admin:menu:exam:edit' AS source_key, 'admin:menu:exam:record-grade' AS target_key
  UNION ALL SELECT 'admin:api:exam:edit', 'admin:api:exam-record:grade'
) mapping ON mapping.source_key = source_grant.`grant_permission_key`
JOIN `permissions` target_perm ON target_perm.`permission_key` = mapping.target_key
WHERE source_grant.`grant_effect` = 'allow' AND source_grant.`grant_status` = 1
  AND source_grant.`grant_subject_type` IN ('role', 'user')
  AND NOT EXISTS (
    SELECT 1 FROM `permission_grants` existing_grant
    WHERE existing_grant.`grant_subject_type` = source_grant.`grant_subject_type`
      AND existing_grant.`grant_subject_id` = source_grant.`grant_subject_id`
      AND existing_grant.`grant_permission_key` = target_perm.`permission_key`
  );
