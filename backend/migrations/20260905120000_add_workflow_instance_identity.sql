-- 流程实例单据标题与业务期间快照。
-- 快照在发起时生成，保证历史单据不受流程定义后续修改影响。

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND COLUMN_NAME = 'instance_title');
SET @ddl := IF(@column_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `instance_title` VARCHAR(200) NOT NULL DEFAULT '''' COMMENT ''单据标题快照'' AFTER `definition_key`',
  'SELECT ''workflow_process_instances.instance_title exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND COLUMN_NAME = 'business_period_type');
SET @ddl := IF(@column_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `business_period_type` VARCHAR(16) NOT NULL DEFAULT '''' COMMENT ''业务期间类型'' AFTER `instance_title`',
  'SELECT ''workflow_process_instances.business_period_type exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND COLUMN_NAME = 'business_period_key');
SET @ddl := IF(@column_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `business_period_key` VARCHAR(40) NOT NULL DEFAULT '''' COMMENT ''业务期间唯一标识'' AFTER `business_period_type`',
  'SELECT ''workflow_process_instances.business_period_key exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND COLUMN_NAME = 'business_period_label');
SET @ddl := IF(@column_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `business_period_label` VARCHAR(80) NOT NULL DEFAULT '''' COMMENT ''业务期间显示文本'' AFTER `business_period_key`',
  'SELECT ''workflow_process_instances.business_period_label exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND INDEX_NAME = 'idx_workflow_instances_title');
SET @ddl := IF(@idx_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD INDEX `idx_workflow_instances_title` (`instance_title`)',
  'SELECT ''idx_workflow_instances_title exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND INDEX_NAME = 'idx_workflow_instances_definition_period_time');
SET @ddl := IF(@idx_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD INDEX `idx_workflow_instances_definition_period_time` (`definition_id`,`business_period_key`,`start_time`)',
  'SELECT ''idx_workflow_instances_definition_period_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
