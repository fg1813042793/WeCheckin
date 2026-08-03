-- 钉钉 H5 删除统一改为软删除。
--
-- 说明：
-- - 考评单删除只标记 deleted_at/delete_by/delete_dept_id，不物理删除主表和流转记录。
-- - review_no 由单列唯一切换为 review_no + deleted_at 组合唯一，避免软删除后无法重建同月考评单。
-- - 历史和模板表也补齐软删除审计列，保持 H5 业务表字段规范一致。

SET @col_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND COLUMN_NAME = 'deleted_at'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews`
     ADD COLUMN `delete_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''删除人ID'' AFTER `update_dept_id`,
     ADD COLUMN `delete_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''删除人部门ID'' AFTER `delete_by`,
     ADD COLUMN `deleted_at` bigint NOT NULL DEFAULT 0 COMMENT ''软删除时间'' AFTER `delete_dept_id`',
  'SELECT ''dingtalk_h5_perf_reviews soft delete fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_histories'
    AND COLUMN_NAME = 'deleted_at'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_histories`
     ADD COLUMN `delete_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''删除人ID'' AFTER `update_dept_id`,
     ADD COLUMN `delete_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''删除人部门ID'' AFTER `delete_by`,
     ADD COLUMN `deleted_at` bigint NOT NULL DEFAULT 0 COMMENT ''软删除时间'' AFTER `delete_dept_id`',
  'SELECT ''dingtalk_h5_perf_histories soft delete fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_templates'
    AND COLUMN_NAME = 'deleted_at'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_templates`
     ADD COLUMN `delete_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''删除人ID'' AFTER `update_dept_id`,
     ADD COLUMN `delete_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''删除人部门ID'' AFTER `delete_by`,
     ADD COLUMN `deleted_at` bigint NOT NULL DEFAULT 0 COMMENT ''软删除时间'' AFTER `delete_dept_id`',
  'SELECT ''dingtalk_h5_perf_templates soft delete fields exist'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dingtalk_h5_perf_reviews_review_no'
);
SET @ddl := IF(
  @idx_exists > 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` DROP INDEX `idx_dingtalk_h5_perf_reviews_review_no`',
  'SELECT ''idx_dingtalk_h5_perf_reviews_review_no absent'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_no_deleted'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD UNIQUE KEY `idx_dt_h5_review_no_deleted` (`review_no`, `deleted_at`)',
  'SELECT ''idx_dt_h5_review_no_deleted exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_reviews'
    AND INDEX_NAME = 'idx_dt_h5_review_deleted_at'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_reviews` ADD INDEX `idx_dt_h5_review_deleted_at` (`deleted_at`)',
  'SELECT ''idx_dt_h5_review_deleted_at exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_histories'
    AND INDEX_NAME = 'idx_dt_h5_history_deleted_at'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_histories` ADD INDEX `idx_dt_h5_history_deleted_at` (`deleted_at`)',
  'SELECT ''idx_dt_h5_history_deleted_at exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_templates'
    AND INDEX_NAME = 'idx_dt_h5_template_deleted_at'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_templates` ADD INDEX `idx_dt_h5_template_deleted_at` (`deleted_at`)',
  'SELECT ''idx_dt_h5_template_deleted_at exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
