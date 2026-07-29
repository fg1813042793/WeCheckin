-- 补齐钉钉 H5 绩效列表、工作台和流程历史的组合索引。
--
-- 适用接口：
-- - GET /api/v2/dingtalk/h5/workbench
-- - GET /api/v2/dingtalk/h5/reviews?page=1&pageSize=20
-- - GET /api/v2/dingtalk/h5/reviews/:id

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_employee_period'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_employee_period` (`employee_account`, `period`)',
  'SELECT ''idx_dt_h5_review_employee_period exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_manager_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_manager_status` (`manager_account`, `status`, `period`)',
  'SELECT ''idx_dt_h5_review_manager_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_hrbp_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_hrbp_status` (`hrbp_account`, `status`, `period`)',
  'SELECT ''idx_dt_h5_review_hrbp_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_hrbp_reviewer_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_hrbp_reviewer_status` (`hrbp_reviewer_account`, `status`, `period`)',
  'SELECT ''idx_dt_h5_review_hrbp_reviewer_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_status_period'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_status_period` (`status`, `period`)',
  'SELECT ''idx_dt_h5_review_status_period exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_histories'
    AND INDEX_NAME = 'idx_dt_h5_history_review_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_histories` ADD INDEX `idx_dt_h5_history_review_time` (`review_id`, `add_time`, `id`)',
  'SELECT ''idx_dt_h5_history_review_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
