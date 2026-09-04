-- 为流程发布版本补充不可变元数据快照、语义变更摘要、发布说明和回滚来源。

SET @column_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definition_versions'
    AND COLUMN_NAME = 'definition_metadata_json'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_definition_versions` ADD COLUMN `definition_metadata_json` mediumtext NOT NULL COMMENT ''发布时流程元数据JSON'' AFTER `definition_source_json`',
  'SELECT ''definition_metadata_json exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definition_versions'
    AND COLUMN_NAME = 'definition_change_base_version'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_definition_versions` ADD COLUMN `definition_change_base_version` int NOT NULL DEFAULT 0 COMMENT ''变更对比基准版本'' AFTER `definition_metadata_json`',
  'SELECT ''definition_change_base_version exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definition_versions'
    AND COLUMN_NAME = 'definition_change_summary_json'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_definition_versions` ADD COLUMN `definition_change_summary_json` mediumtext NOT NULL COMMENT ''结构化变更摘要JSON'' AFTER `definition_change_base_version`',
  'SELECT ''definition_change_summary_json exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definition_versions'
    AND COLUMN_NAME = 'definition_publish_note'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_definition_versions` ADD COLUMN `definition_publish_note` varchar(500) NOT NULL DEFAULT '''' COMMENT ''发布说明'' AFTER `definition_change_summary_json`',
  'SELECT ''definition_publish_note exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definition_versions'
    AND COLUMN_NAME = 'definition_content_hash'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_definition_versions` ADD COLUMN `definition_content_hash` char(64) NOT NULL DEFAULT '''' COMMENT ''版本内容SHA256'' AFTER `definition_publish_note`',
  'SELECT ''definition_content_hash exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definition_versions'
    AND COLUMN_NAME = 'definition_rollback_from_version'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_definition_versions` ADD COLUMN `definition_rollback_from_version` int NOT NULL DEFAULT 0 COMMENT ''回滚来源版本'' AFTER `definition_content_hash`',
  'SELECT ''definition_rollback_from_version exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
