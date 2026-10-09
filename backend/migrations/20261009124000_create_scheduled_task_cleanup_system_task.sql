-- Register the built-in scheduled-task history cleanup job.

SET @scheduled_task_cleanup_now = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS SIGNED);

INSERT INTO `scheduled_tasks` (
  `code`, `name`, `description`, `handler_type`, `handler_config_json`,
  `cron_expression`, `cron_precision`, `timezone`, `enabled`, `misfire_policy`,
  `max_catch_up`, `concurrency_policy`, `timeout_seconds`, `max_retries`, `retry_backoff_json`,
  `last_scheduled_at`, `next_run_at`, `version`, `created_by`, `updated_by`, `deleted_at`,
  `add_time`, `edit_time`
) VALUES (
  'system.scheduled-task-cleanup',
  '清理定时任务历史',
  '按服务端保留策略清理过期运行记录和分段日志',
  'go',
  '{"handlerKey":"scheduled-task.cleanup","params":{"batchSize":1000}}',
  '0 3 * * *',
  'minute',
  'Asia/Shanghai',
  1,
  'fire_once',
  1,
  'skip',
  300,
  2,
  '{"type":"fixed","seconds":60}',
  0,
  @scheduled_task_cleanup_now,
  1,
  0,
  0,
  0,
  @scheduled_task_cleanup_now,
  @scheduled_task_cleanup_now
)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `handler_type` = VALUES(`handler_type`),
  `handler_config_json` = VALUES(`handler_config_json`),
  `cron_expression` = VALUES(`cron_expression`),
  `cron_precision` = VALUES(`cron_precision`),
  `timezone` = VALUES(`timezone`),
  `enabled` = 1,
  `misfire_policy` = 'fire_once',
  `deleted_at` = 0,
  `edit_time` = @scheduled_task_cleanup_now;
