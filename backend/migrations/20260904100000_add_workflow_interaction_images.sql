-- 为流程任务处理和评论历史增加图片元数据，并让已有评论授权主体可以上传图片。

SET @workflow_task_images_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_tasks'
    AND COLUMN_NAME = 'task_images_json'
);
SET @workflow_task_images_ddl := IF(
  @workflow_task_images_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `task_images_json` mediumtext NULL COMMENT ''处理图片JSON'' AFTER `task_comment`',
  'SELECT ''workflow_process_tasks.task_images_json exists'''
);
PREPARE workflow_task_images_stmt FROM @workflow_task_images_ddl;
EXECUTE workflow_task_images_stmt;
DEALLOCATE PREPARE workflow_task_images_stmt;

SET @workflow_history_images_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_history'
    AND COLUMN_NAME = 'event_images_json'
);
SET @workflow_history_images_ddl := IF(
  @workflow_history_images_exists = 0,
  'ALTER TABLE `workflow_process_history` ADD COLUMN `event_images_json` mediumtext NULL COMMENT ''事件图片JSON'' AFTER `event_message`',
  'SELECT ''workflow_process_history.event_images_json exists'''
);
PREPARE workflow_history_images_stmt FROM @workflow_history_images_ddl;
EXECUTE workflow_history_images_stmt;
DEALLOCATE PREPARE workflow_history_images_stmt;

SET @now_ms = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000;

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  comment_grant.`grant_subject_type`,
  comment_grant.`grant_subject_id`,
  attachment_perm.`permission_key`,
  attachment_perm.`id`,
  'allow',
  '',
  'workflow-comment-image-backfill',
  1,
  @now_ms,
  @now_ms,
  NOW(3),
  NOW(3)
FROM `permission_grants` comment_grant
JOIN `permissions` attachment_perm
  ON attachment_perm.`permission_key` = 'dingtalk_h5:api:workflow:attachment'
  AND attachment_perm.`permission_platform` = 'dingtalk_h5'
  AND attachment_perm.`permission_type` = 'api'
WHERE comment_grant.`grant_permission_key` = 'dingtalk_h5:api:workflow:comment'
  AND comment_grant.`grant_subject_type` IN ('role', 'user')
  AND comment_grant.`grant_effect` = 'allow'
  AND comment_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
