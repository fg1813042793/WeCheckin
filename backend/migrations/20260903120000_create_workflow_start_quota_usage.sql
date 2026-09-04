-- 通过独立额度行锁串行化同一流程、发起人和周期的并发提交。

CREATE TABLE IF NOT EXISTS `workflow_start_quota_usage` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '流程发起额度记录ID',
  `definition_id` bigint unsigned NOT NULL COMMENT '流程定义ID',
  `starter_id` varchar(64) NOT NULL COMMENT '业务发起人ID',
  `period_key` varchar(100) NOT NULL COMMENT '额度周期唯一标识',
  `used_count` int NOT NULL DEFAULT 0 COMMENT '已使用次数快照',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_start_quota_period` (`definition_id`,`starter_id`,`period_key`),
  KEY `idx_workflow_start_quota_starter` (`starter_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程发起次数额度';

SET @workflow_instance_quota_index_exists := (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'workflow_process_instances'
    AND index_name = 'idx_workflow_instances_definition_starter_time'
);
SET @workflow_instance_quota_index_ddl := IF(
  @workflow_instance_quota_index_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD INDEX `idx_workflow_instances_definition_starter_time` (`definition_id`,`starter_id`,`start_time`)',
  'SELECT ''idx_workflow_instances_definition_starter_time exists'''
);
PREPARE workflow_instance_quota_index_stmt FROM @workflow_instance_quota_index_ddl;
EXECUTE workflow_instance_quota_index_stmt;
DEALLOCATE PREPARE workflow_instance_quota_index_stmt;
