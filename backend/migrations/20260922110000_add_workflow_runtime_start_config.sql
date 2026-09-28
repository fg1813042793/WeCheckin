-- 将流程配置拆分为独立运行时策略，保存后立即生效，不依赖重新发布流程版本。

SET @workflow_start_config_json_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND COLUMN_NAME = 'definition_start_config_json'
);
SET @workflow_start_config_json_sql := IF(
  @workflow_start_config_json_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_start_config_json` MEDIUMTEXT NULL COMMENT ''当前生效的流程发起配置JSON'' AFTER `definition_draft_json`',
  'SELECT 1'
);
PREPARE workflow_start_config_json_stmt FROM @workflow_start_config_json_sql;
EXECUTE workflow_start_config_json_stmt;
DEALLOCATE PREPARE workflow_start_config_json_stmt;

SET @workflow_start_config_revision_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND COLUMN_NAME = 'definition_start_config_revision'
);
SET @workflow_start_config_revision_sql := IF(
  @workflow_start_config_revision_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_start_config_revision` INT NOT NULL DEFAULT 0 COMMENT ''当前生效的流程发起配置修订号'' AFTER `definition_start_config_json`',
  'SELECT 1'
);
PREPARE workflow_start_config_revision_stmt FROM @workflow_start_config_revision_sql;
EXECUTE workflow_start_config_revision_stmt;
DEALLOCATE PREPARE workflow_start_config_revision_stmt;
