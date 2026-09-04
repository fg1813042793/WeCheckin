-- 补充后端持续增长表的查询索引。
--
-- 审计依据：
-- 1. 站内信按用户/未读状态查询，以及发送批次幂等回查；
-- 2. 工作流按业务标识、处理人、待办人查询，以及通知列表和超时回收；
-- 3. 定时任务运行队列、并发判断、投递恢复和历史清理；
-- 4. 已发布流程列表及钉钉用户绑定解析。
--
-- 均为非唯一二级索引，不改变历史数据约束。大表执行 ALTER TABLE 前应在生产环境
-- 确认磁盘空间，并安排低峰窗口。回滚时按索引名执行 ALTER TABLE ... DROP INDEX。

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND INDEX_NAME = 'idx_notify_user_read_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `notify` ADD INDEX `idx_notify_user_read_id` (`notify_user_id`,`notify_is_read`,`notify_id`)',
  'SELECT ''idx_notify_user_read_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND INDEX_NAME = 'idx_notify_source_delivery'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `notify` ADD INDEX `idx_notify_source_delivery` (`notify_source_type`,`notify_source_id`,`notify_delivery_key`)',
  'SELECT ''idx_notify_source_delivery exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_instances'
    AND INDEX_NAME = 'idx_workflow_instances_business_lookup'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD INDEX `idx_workflow_instances_business_lookup` (`business_type`,`business_key`)',
  'SELECT ''idx_workflow_instances_business_lookup exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_tasks'
    AND INDEX_NAME = 'idx_workflow_tasks_handled_status_instance'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD INDEX `idx_workflow_tasks_handled_status_instance` (`handled_by`,`task_status`,`instance_id`)',
  'SELECT ''idx_workflow_tasks_handled_status_instance exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_process_tasks'
    AND INDEX_NAME = 'idx_workflow_tasks_assignee_status_created'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_process_tasks` ADD INDEX `idx_workflow_tasks_assignee_status_created` (`task_assignee_id`,`task_status`,`admin_deleted_at`,`created_at`,`id`)',
  'SELECT ''idx_workflow_tasks_assignee_status_created exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND INDEX_NAME = 'idx_workflow_notification_status_edit'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD INDEX `idx_workflow_notification_status_edit` (`notification_status`,`edit_time`)',
  'SELECT ''idx_workflow_notification_status_edit exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_notification_outbox'
    AND INDEX_NAME = 'idx_workflow_notification_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_notification_outbox` ADD INDEX `idx_workflow_notification_time` (`add_time`,`id`)',
  'SELECT ''idx_workflow_notification_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduled_task_runs'
    AND INDEX_NAME = 'idx_scheduled_runs_task_status_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_task_runs` ADD INDEX `idx_scheduled_runs_task_status_time` (`task_id`,`run_status`,`add_time`,`id`)',
  'SELECT ''idx_scheduled_runs_task_status_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduled_task_runs'
    AND INDEX_NAME = 'idx_scheduled_runs_delivery_due'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_task_runs` ADD INDEX `idx_scheduled_runs_delivery_due` (`run_status`,`redis_message_id`,`queued_at`,`id`)',
  'SELECT ''idx_scheduled_runs_delivery_due exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduled_task_runs'
    AND INDEX_NAME = 'idx_scheduled_runs_status_schedule'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_task_runs` ADD INDEX `idx_scheduled_runs_status_schedule` (`run_status`,`scheduled_at`,`id`)',
  'SELECT ''idx_scheduled_runs_status_schedule exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduled_task_runs'
    AND INDEX_NAME = 'idx_scheduled_runs_cleanup'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_task_runs` ADD INDEX `idx_scheduled_runs_cleanup` (`run_status`,`finished_at`,`id`)',
  'SELECT ''idx_scheduled_runs_cleanup exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduled_task_run_logs'
    AND INDEX_NAME = 'idx_scheduled_run_logs_cleanup'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_task_run_logs` ADD INDEX `idx_scheduled_run_logs_cleanup` (`log_time`,`id`)',
  'SELECT ''idx_scheduled_run_logs_cleanup exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'workflow_definitions'
    AND INDEX_NAME = 'idx_workflow_definitions_published_sort'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `workflow_definitions` ADD INDEX `idx_workflow_definitions_published_sort` (`definition_status`,`definition_category`,`definition_name`,`id`)',
  'SELECT ''idx_workflow_definitions_published_sort exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_user_bindings'
    AND INDEX_NAME = 'idx_dt_h5_bindings_user_enabled'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_user_bindings` ADD INDEX `idx_dt_h5_bindings_user_enabled` (`user_id`,`enabled`,`corp_id`,`dingtalk_user_id`)',
  'SELECT ''idx_dt_h5_bindings_user_enabled exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
