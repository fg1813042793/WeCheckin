-- 为逐级部门审批保存任务层级快照，并增加内置“主管”审批身份。

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_tasks' AND COLUMN_NAME = 'approval_chain_key'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `approval_chain_key` varchar(64) NOT NULL DEFAULT '''' COMMENT ''分层审批链快照标识'' AFTER `task_total`',
  'SELECT ''workflow_process_tasks.approval_chain_key exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_tasks' AND COLUMN_NAME = 'approval_layer'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `approval_layer` int NOT NULL DEFAULT 0 COMMENT ''分层审批层级序号'' AFTER `approval_chain_key`',
  'SELECT ''workflow_process_tasks.approval_layer exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_tasks' AND COLUMN_NAME = 'approval_layer_total'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `approval_layer_total` int NOT NULL DEFAULT 0 COMMENT ''分层审批总层数'' AFTER `approval_layer`',
  'SELECT ''workflow_process_tasks.approval_layer_total exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_tasks' AND COLUMN_NAME = 'source_department_id'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `source_department_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''审批层来源部门ID快照'' AFTER `approval_layer_total`',
  'SELECT ''workflow_process_tasks.source_department_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_tasks' AND COLUMN_NAME = 'source_department_name'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD COLUMN `source_department_name` varchar(100) NOT NULL DEFAULT '''' COMMENT ''审批层来源部门名称快照'' AFTER `source_department_id`',
  'SELECT ''workflow_process_tasks.source_department_name exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_tasks' AND INDEX_NAME = 'idx_workflow_tasks_chain_layer'
);
SET @ddl := IF(
  @index_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD INDEX `idx_workflow_tasks_chain_layer` (`approval_chain_key`,`approval_layer`)',
  'SELECT ''idx_workflow_tasks_chain_layer exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `workflow_org_approver_identities` (
  `identity_code`, `identity_name`, `identity_sort`, `identity_status`,
  `identity_add_time`, `identity_edit_time`, `created_at`, `updated_at`
) VALUES (
  'supervisor', '主管', 15, 1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3), NOW(3)
)
ON DUPLICATE KEY UPDATE
  `identity_name` = VALUES(`identity_name`),
  `identity_sort` = VALUES(`identity_sort`),
  `identity_status` = 1,
  `identity_edit_time` = VALUES(`identity_edit_time`),
  `updated_at` = NOW(3);
