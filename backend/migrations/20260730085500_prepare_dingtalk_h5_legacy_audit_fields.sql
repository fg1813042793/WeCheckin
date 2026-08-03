-- 兼容全新库的维护执行顺序：GORM AutoMigrate 先于版本化 SQL 迁移执行。
--
-- 现行模型只保留标准审计字段；但已发布的 20260730090000 迁移会在回填阶段读取
-- create_account、update_account、create_dept_path。全新库如果先由 AutoMigrate 建表，
-- create_by 已存在会导致 20260730090000 跳过加旧字段，从而回填时报列不存在。
--
-- 本迁移仅在 20260730090000 尚未执行、且表已提前建出标准字段时补齐临时旧字段；
-- 20260730103000 会再删除这些非标准字段。已有线上库如果已执行 20260730090000，
-- 本迁移保持 no-op，避免修改已运行结构。

SET @old_h5_audit_migration_ran := (
  SELECT COUNT(1) FROM `schema_migrations`
  WHERE `migration_version` = '20260730090000_add_dingtalk_h5_audit_fields'
);

SET @table_name := 'dingtalk_h5_perf_sessions';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`'), CONCAT('SELECT ''', @table_name, '.create_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`'), CONCAT('SELECT ''', @table_name, '.update_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_id');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.create_dept_path prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_reviews';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`'), CONCAT('SELECT ''', @table_name, '.create_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`'), CONCAT('SELECT ''', @table_name, '.update_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_id');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.create_dept_path prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_histories';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`'), CONCAT('SELECT ''', @table_name, '.create_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`'), CONCAT('SELECT ''', @table_name, '.update_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_id');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.create_dept_path prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_templates';
SET @table_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name);
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''创建人账号'' AFTER `create_by`'), CONCAT('SELECT ''', @table_name, '.create_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_by');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_account` varchar(100) NOT NULL DEFAULT '''' COMMENT ''更新人账号'' AFTER `update_by`'), CONCAT('SELECT ''', @table_name, '.update_account prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @anchor_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_id');
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@old_h5_audit_migration_ran = 0 AND @table_exists > 0 AND @anchor_exists > 0 AND @col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `create_dept_path` varchar(500) NOT NULL DEFAULT '''' COMMENT ''创建人部门路径'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.create_dept_path prepare skipped'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
