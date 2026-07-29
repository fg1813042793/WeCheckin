-- 补齐客户端首页/列表接口的排序和状态过滤索引。
--
-- 适用接口：
-- - GET /api/v2/news?page=1&pageSize=20
-- - GET /api/v2/enrollments?page=1&pageSize=20
-- - GET /api/v2/events?page=1&pageSize=20
-- - GET /api/v2/surveys?page=1&pageSize=20
-- - GET /api/v2/exams?page=1&pageSize=20

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enrolls'
    AND INDEX_NAME = 'idx_enrolls_status_order_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enrolls` ADD INDEX `idx_enrolls_status_order_time` (`enroll_status`, `enroll_order`, `enroll_add_time`, `id`)',
  'SELECT ''idx_enrolls_status_order_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enrolls'
    AND INDEX_NAME = 'idx_enrolls_title'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enrolls` ADD INDEX `idx_enrolls_title` (`enroll_title`)',
  'SELECT ''idx_enrolls_title exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'events'
    AND INDEX_NAME = 'idx_events_status_type_order_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `events` ADD INDEX `idx_events_status_type_order_time` (`event_status`, `event_type`, `event_order`, `event_add_time`, `id`)',
  'SELECT ''idx_events_status_type_order_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'survey'
    AND INDEX_NAME = 'idx_surveys_status_order_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `survey` ADD INDEX `idx_surveys_status_order_id` (`survey_status`, `survey_order`, `survey_id`)',
  'SELECT ''idx_surveys_status_order_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'exam'
    AND INDEX_NAME = 'idx_exams_status_order_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `exam` ADD INDEX `idx_exams_status_order_id` (`exam_status`, `exam_order`, `exam_id`)',
  'SELECT ''idx_exams_status_order_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
