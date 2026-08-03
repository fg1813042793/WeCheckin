-- 优化钉钉 H5 绩效上级/HRBP tab 列表过滤与排序。

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_employee_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_employee_period_id` (`employee_account`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_employee_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_manager_status_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_manager_status_period_id` (`manager_account`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_manager_status_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_hrbp_status_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_hrbp_status_period_id` (`hrbp_account`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_hrbp_status_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_hrbp_reviewer_status_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_hrbp_reviewer_status_period_id` (`hrbp_reviewer_account`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_hrbp_reviewer_status_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
