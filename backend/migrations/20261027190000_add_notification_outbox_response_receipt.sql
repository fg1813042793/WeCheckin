-- Persist bounded, redacted webhook response receipts for delivery diagnostics.

SET @response_status_code_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notification_outbox'
    AND COLUMN_NAME = 'response_status_code'
);
SET @ddl := IF(
  @response_status_code_exists = 0,
  'ALTER TABLE `notification_outbox` ADD COLUMN `response_status_code` INT NOT NULL DEFAULT 0 COMMENT ''Latest HTTP response status, 0 when unavailable'' AFTER `sent_at`',
  'SELECT ''notification_outbox.response_status_code exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @response_body_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notification_outbox'
    AND COLUMN_NAME = 'response_body'
);
SET @ddl := IF(
  @response_body_exists = 0,
  'ALTER TABLE `notification_outbox` ADD COLUMN `response_body` TEXT NOT NULL COMMENT ''Bounded and redacted latest HTTP response body'' AFTER `response_status_code`',
  'SELECT ''notification_outbox.response_body exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
