-- 站内信用户侧软删除字段与列表查询索引。
-- 删除后消息仍保留在 notify 表中，用户侧列表、未读数和已读操作统一忽略该记录。

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND COLUMN_NAME = 'notify_deleted_at'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `notify` ADD COLUMN `notify_deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''用户删除时间，0表示未删除'' AFTER `notify_add_time`',
  'SELECT ''notify.notify_deleted_at exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND INDEX_NAME = 'idx_notify_user_deleted_read_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `notify` ADD INDEX `idx_notify_user_deleted_read_id` (`notify_user_id`,`notify_deleted_at`,`notify_is_read`,`notify_id`)',
  'SELECT ''idx_notify_user_deleted_read_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
