SET @workflow_definition_logo_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND COLUMN_NAME = 'definition_logo_url'
);
SET @ddl := IF(
  @workflow_definition_logo_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_logo_url` varchar(500) NOT NULL DEFAULT '''' COMMENT ''流程Logo地址'' AFTER `definition_category`',
  'SELECT ''definition_logo_url exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
