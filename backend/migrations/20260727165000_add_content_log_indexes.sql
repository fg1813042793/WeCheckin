-- 为新闻内容和后台日志相关接口增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/news、GET /api/v2/admin/news、GET /api/v2/home/list
-- - GET /api/v2/admin/logs
--
-- 回滚：
-- ALTER TABLE `news` DROP INDEX `idx_news_status_order_time`;
-- ALTER TABLE `news` DROP INDEX `idx_news_status_vouch_order_time`;
-- ALTER TABLE `news` DROP INDEX `idx_news_add_time_id`;
-- ALTER TABLE `news` DROP INDEX `idx_news_title`;
-- ALTER TABLE `news` DROP INDEX `idx_news_dept_create_time`;
-- ALTER TABLE `logs` DROP INDEX `idx_logs_add_time_id`;
-- ALTER TABLE `logs` DROP INDEX `idx_logs_admin_time`;
-- ALTER TABLE `logs` DROP INDEX `idx_logs_admin_name`;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'news'
    AND INDEX_NAME = 'idx_news_status_order_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `news` ADD INDEX `idx_news_status_order_time` (`news_status`, `news_order`, `news_add_time`, `id`)',
  'SELECT ''idx_news_status_order_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'news'
    AND INDEX_NAME = 'idx_news_status_vouch_order_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `news` ADD INDEX `idx_news_status_vouch_order_time` (`news_status`, `news_vouch`, `news_order`, `news_add_time`, `id`)',
  'SELECT ''idx_news_status_vouch_order_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'news'
    AND INDEX_NAME = 'idx_news_add_time_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `news` ADD INDEX `idx_news_add_time_id` (`news_add_time`, `id`)',
  'SELECT ''idx_news_add_time_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'news'
    AND INDEX_NAME = 'idx_news_title'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `news` ADD INDEX `idx_news_title` (`news_title`)',
  'SELECT ''idx_news_title exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'news'
    AND INDEX_NAME = 'idx_news_dept_create_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `news` ADD INDEX `idx_news_dept_create_time` (`news_dept_id`, `news_create_by`, `news_add_time`, `id`)',
  'SELECT ''idx_news_dept_create_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'logs'
    AND INDEX_NAME = 'idx_logs_add_time_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_add_time_id` (`log_add_time`, `id`)',
  'SELECT ''idx_logs_add_time_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'logs'
    AND INDEX_NAME = 'idx_logs_admin_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_admin_time` (`log_admin_id`, `log_add_time`, `id`)',
  'SELECT ''idx_logs_admin_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'logs'
    AND INDEX_NAME = 'idx_logs_admin_name'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `logs` ADD INDEX `idx_logs_admin_name` (`log_admin_name`)',
  'SELECT ''idx_logs_admin_name exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
