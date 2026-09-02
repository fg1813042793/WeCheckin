-- 区分流程的业务发起人和实际发起操作人。
-- 旧实例没有代发语义，迁移时将操作人回填为原发起人。

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_instances'
    AND COLUMN_NAME = 'operator_id'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `operator_id` varchar(64) NOT NULL DEFAULT '''' COMMENT ''实际发起操作人ID'' AFTER `starter_id`',
  'SELECT ''operator_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `workflow_process_instances`
SET `operator_id` = `starter_id`
WHERE `operator_id` = '';

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_instances'
    AND INDEX_NAME = 'idx_workflow_instances_operator_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD INDEX `idx_workflow_instances_operator_status` (`operator_id`, `instance_status`)',
  'SELECT ''idx_workflow_instances_operator_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
