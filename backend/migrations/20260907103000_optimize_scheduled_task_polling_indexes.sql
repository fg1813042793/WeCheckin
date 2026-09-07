-- 优化定时任务调度器的高频轮询查询。
--
-- 到期任务查询同时固定 enabled、deleted_at，并按 next_run_at、id 取一批记录；
-- 原索引缺少 deleted_at，数据增长后会产生额外扫描。另两项索引在历史迁移中已经
-- 声明，此处做幂等兜底，避免部署库曾跳过索引迁移时持续全表扫描。

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduled_tasks'
    AND INDEX_NAME = 'idx_scheduled_tasks_poll_due'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_tasks` ADD INDEX `idx_scheduled_tasks_poll_due` (`enabled`,`deleted_at`,`next_run_at`,`id`)',
  'SELECT ''idx_scheduled_tasks_poll_due exists'''
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
    AND INDEX_NAME = 'idx_scheduled_task_runs_recovery'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `scheduled_task_runs` ADD INDEX `idx_scheduled_task_runs_recovery` (`run_status`,`heartbeat_at`)',
  'SELECT ''idx_scheduled_task_runs_recovery exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
