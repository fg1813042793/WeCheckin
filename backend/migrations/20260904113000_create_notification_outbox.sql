-- Reliable generic notification outbox and its protected dispatch task.

CREATE TABLE IF NOT EXISTS `notification_outbox` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Outbox record ID',
  `idempotency_key` VARCHAR(191) NOT NULL COMMENT 'Source delivery idempotency key',
  `source_type` VARCHAR(64) NOT NULL COMMENT 'Source aggregate type',
  `source_id` VARCHAR(128) NOT NULL COMMENT 'Source aggregate ID',
  `notification_channel` VARCHAR(32) NOT NULL COMMENT 'Delivery channel',
  `recipient_json` MEDIUMTEXT NOT NULL COMMENT 'Channel recipient JSON',
  `payload_json` MEDIUMTEXT NOT NULL COMMENT 'Channel payload JSON',
  `notification_status` VARCHAR(24) NOT NULL DEFAULT 'pending' COMMENT 'pending, sending, sent, failed or dead',
  `attempts` INT NOT NULL DEFAULT 0 COMMENT 'Failed delivery attempts',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Next delivery time in UTC milliseconds',
  `last_error` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT 'Redacted latest delivery error',
  `sent_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Successful delivery time in UTC milliseconds',
  `add_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'Creation time in milliseconds',
  `edit_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'Update time in milliseconds',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_notification_outbox_idempotency` (`idempotency_key`),
  KEY `idx_notification_outbox_due` (`notification_status`,`next_retry_at`),
  KEY `idx_notification_outbox_status_edit` (`notification_status`,`edit_time`),
  KEY `idx_notification_outbox_source` (`source_type`,`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Reliable generic notification delivery outbox';

SET @notification_outbox_now = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS SIGNED);

INSERT INTO `scheduled_tasks` (
  `code`, `name`, `description`, `handler_type`, `handler_config_json`,
  `cron_expression`, `cron_precision`, `timezone`, `enabled`, `misfire_policy`,
  `max_catch_up`, `concurrency_policy`, `timeout_seconds`, `max_retries`, `retry_backoff_json`,
  `last_scheduled_at`, `next_run_at`, `version`, `created_by`, `updated_by`, `deleted_at`,
  `add_time`, `edit_time`
) VALUES (
  'system.notification-outbox-dispatch',
  '通用通知派发',
  '派发到期的站内通知和 webhook outbox 记录',
  'go',
  '{"handlerKey":"notification.outbox.dispatch_due","params":{"limit":100}}',
  '* * * * *',
  'minute',
  'Asia/Shanghai',
  1,
  'skip',
  1,
  'skip',
  60,
  2,
  '{"type":"fixed","seconds":30}',
  0,
  @notification_outbox_now,
  1,
  0,
  0,
  0,
  @notification_outbox_now,
  @notification_outbox_now
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
  `misfire_policy` = VALUES(`misfire_policy`),
  `max_catch_up` = VALUES(`max_catch_up`),
  `concurrency_policy` = VALUES(`concurrency_policy`),
  `timeout_seconds` = VALUES(`timeout_seconds`),
  `max_retries` = VALUES(`max_retries`),
  `retry_backoff_json` = VALUES(`retry_backoff_json`),
  `next_run_at` = IF(`next_run_at` > 0, `next_run_at`, @notification_outbox_now),
  `deleted_at` = 0,
  `edit_time` = @notification_outbox_now;
