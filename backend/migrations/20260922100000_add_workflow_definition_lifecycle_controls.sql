-- 流程定义采用管理端软删除，保留发布版本和历史实例用于审计。

SET @workflow_definition_deleted_at_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND COLUMN_NAME = 'definition_deleted_at'
);
SET @workflow_definition_deleted_at_sql := IF(
  @workflow_definition_deleted_at_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''管理员删除时间，0表示未删除'' AFTER `definition_edit_time`',
  'SELECT 1'
);
PREPARE workflow_definition_deleted_at_stmt FROM @workflow_definition_deleted_at_sql;
EXECUTE workflow_definition_deleted_at_stmt;
DEALLOCATE PREPARE workflow_definition_deleted_at_stmt;

SET @workflow_definition_deleted_by_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND COLUMN_NAME = 'definition_deleted_by'
);
SET @workflow_definition_deleted_by_sql := IF(
  @workflow_definition_deleted_by_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_deleted_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''删除操作管理员ID'' AFTER `definition_deleted_at`',
  'SELECT 1'
);
PREPARE workflow_definition_deleted_by_stmt FROM @workflow_definition_deleted_by_sql;
EXECUTE workflow_definition_deleted_by_stmt;
DEALLOCATE PREPARE workflow_definition_deleted_by_stmt;

SET @workflow_definition_active_index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND INDEX_NAME = 'idx_workflow_definitions_active_status_edit'
);
SET @workflow_definition_active_index_sql := IF(
  @workflow_definition_active_index_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD INDEX `idx_workflow_definitions_active_status_edit` (`definition_deleted_at`,`definition_status`,`definition_edit_time`,`id`)',
  'SELECT 1'
);
PREPARE workflow_definition_active_index_stmt FROM @workflow_definition_active_index_sql;
EXECUTE workflow_definition_active_index_stmt;
DEALLOCATE PREPARE workflow_definition_active_index_stmt;
