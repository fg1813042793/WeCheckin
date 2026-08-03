-- 优化钉钉 H5 绩效列表在 HRBP/汇总/数据权限范围下的过滤与分页。
--
-- 高频查询示例：
-- - GET /api/v2/dingtalk/h5/reviews?scope=hrbp&skipHistory=1&statuses=employee_confirm,hr_final,completed
-- - GET /api/v2/dingtalk/h5/reviews?scope=summary&skipHistory=1&period=2026-07
--
-- 这些查询都会带 deleted_at = 0，并按 status/period/id 过滤排序；HRBP 范围还会叠加
-- hrbp_account、hrbp_reviewer_account、部门层级或审计创建人/部门范围。补齐复合索引可减少
-- COUNT 与分页列表扫描行数。

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_status_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_status_period_id` (`deleted_at`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_status_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_employee_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_employee_period_id` (`deleted_at`, `employee_account`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_employee_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_manager_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_manager_period_id` (`deleted_at`, `manager_account`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_manager_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_hrbp_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_hrbp_period_id` (`deleted_at`, `hrbp_account`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_hrbp_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_hrbp_reviewer_period_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_hrbp_reviewer_period_id` (`deleted_at`, `hrbp_reviewer_account`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_hrbp_reviewer_period_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_hrbp_status_period'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_hrbp_status_period` (`deleted_at`, `hrbp_account`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_hrbp_status_period exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_hrbp_reviewer_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_hrbp_reviewer_status` (`deleted_at`, `hrbp_reviewer_account`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_hrbp_reviewer_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_dept1_status_period'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_dept1_status_period` (`deleted_at`, `department_level1`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_dept1_status_period exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_dept2_status_period'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_dept2_status_period` (`deleted_at`, `department_level2`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_dept2_status_period exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_dept3_status_period'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_dept3_status_period` (`deleted_at`, `department_level3`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_dept3_status_period exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_create_by_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_create_by_status` (`deleted_at`, `create_by`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_create_by_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_del_create_dept_status'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_del_create_dept_status` (`deleted_at`, `create_dept_id`, `status`, `period`, `id`)',
  'SELECT ''idx_dt_h5_review_del_create_dept_status exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
