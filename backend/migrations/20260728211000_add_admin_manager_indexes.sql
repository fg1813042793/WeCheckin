-- 补齐后台管理员列表在 users 表合并后的查询索引。
--
-- 适用接口：
-- - GET /api/v2/admin/managers?page=1&pageSize=20

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_admin_list'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_admin_list` (`user_role_id`, `user_status`, `user_add_time`, `id`)',
  'SELECT ''idx_users_admin_list exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_admin_login_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_admin_login_time` (`user_role_id`, `user_login_time`)',
  'SELECT ''idx_users_admin_login_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_depts'
    AND INDEX_NAME = 'idx_user_depts_dept_user'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `user_depts` ADD INDEX `idx_user_depts_dept_user` (`user_dept_dept_id`, `user_dept_user_id`)',
  'SELECT ''idx_user_depts_dept_user exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_depts'
    AND INDEX_NAME = 'idx_user_depts_user_dept'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `user_depts` ADD INDEX `idx_user_depts_user_dept` (`user_dept_user_id`, `user_dept_dept_id`)',
  'SELECT ''idx_user_depts_user_dept exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
