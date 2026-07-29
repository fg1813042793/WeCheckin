-- 补齐后台用户列表高频查询索引。
--
-- 适用接口：
-- - GET /api/v2/admin/users?page=1&pageSize=20
--
-- 说明：
-- - 历史环境可能已经存在部分索引，本迁移保持幂等。

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_status_role_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_status_role_id` (`user_status`, `user_role_id`)',
  'SELECT ''idx_users_status_role_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_add_time_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_add_time_id` (`user_add_time`, `id`)',
  'SELECT ''idx_users_add_time_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_mobile'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_mobile` (`user_mobile`)',
  'SELECT ''idx_users_mobile exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_name'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_name` (`user_name`)',
  'SELECT ''idx_users_name exists'''
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
