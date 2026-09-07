-- 流程用户显示名称及实例名称快照。
-- 管理名称用于后台区分流程；用户显示名称随发布版本固化，实例发起时再次快照。

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_definitions' AND COLUMN_NAME = 'definition_display_name');
SET @ddl := IF(@column_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_display_name` VARCHAR(200) NOT NULL DEFAULT '''' COMMENT ''用户显示名称'' AFTER `definition_name`',
  'SELECT ''workflow_definitions.definition_display_name exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND COLUMN_NAME = 'definition_name_snapshot');
SET @ddl := IF(@column_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `definition_name_snapshot` VARCHAR(200) NOT NULL DEFAULT '''' COMMENT ''流程用户显示名称快照'' AFTER `definition_key`',
  'SELECT ''workflow_process_instances.definition_name_snapshot exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `workflow_process_instances` AS i
LEFT JOIN `workflow_definitions` AS d ON d.`id` = i.`definition_id`
SET i.`definition_name_snapshot` = COALESCE(NULLIF(d.`definition_display_name`, ''), NULLIF(d.`definition_name`, ''), i.`definition_key`)
WHERE i.`definition_name_snapshot` = '';
