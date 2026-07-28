-- 为后台管理员列表和部门数据范围过滤增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/admin/managers?page=1&pageSize=20

SET @admins_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admins'
);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admins' AND INDEX_NAME = 'idx_admins_add_time_id'
);
SET @ddl := IF(@admins_table_exists > 0 AND @idx_exists = 0, 'ALTER TABLE `admins` ADD INDEX `idx_admins_add_time_id` (`admin_add_time`, `id`)', 'SELECT ''admins table missing or idx_admins_add_time_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admins' AND INDEX_NAME = 'idx_admins_phone'
);
SET @ddl := IF(@admins_table_exists > 0 AND @idx_exists = 0, 'ALTER TABLE `admins` ADD INDEX `idx_admins_phone` (`admin_phone`)', 'SELECT ''admins table missing or idx_admins_phone exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @admin_depts_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_depts'
);

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_depts' AND INDEX_NAME = 'idx_admin_depts_dept_admin'
);
SET @ddl := IF(@admin_depts_table_exists > 0 AND @idx_exists = 0, 'ALTER TABLE `admin_depts` ADD INDEX `idx_admin_depts_dept_admin` (`admin_dept_dept_id`, `admin_dept_admin_id`)', 'SELECT ''admin_depts table missing or idx_admin_depts_dept_admin exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_depts' AND INDEX_NAME = 'idx_admin_depts_admin_dept'
);
SET @ddl := IF(@admin_depts_table_exists > 0 AND @idx_exists = 0, 'ALTER TABLE `admin_depts` ADD INDEX `idx_admin_depts_admin_dept` (`admin_dept_admin_id`, `admin_dept_dept_id`)', 'SELECT ''admin_depts table missing or idx_admin_depts_admin_dept exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
