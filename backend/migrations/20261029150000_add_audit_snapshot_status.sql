-- Record snapshot availability independently of the business outcome.
-- Historical rows stay unknown; do not infer old values from current settings.
SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_snapshot_status');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_snapshot_status` varchar(32) NOT NULL DEFAULT ''legacy_unknown''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_snapshot_reason');
SET @ddl := IF(@column_exists = 0, 'ALTER TABLE `logs` ADD COLUMN `log_snapshot_reason` varchar(80) NOT NULL DEFAULT ''''', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
