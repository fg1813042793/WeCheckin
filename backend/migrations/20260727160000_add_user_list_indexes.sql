-- 为后台/客户端管理的用户列表增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/admin/users?page=1&pageSize=20
--
-- 主要优化：
-- - users(user_add_time, id)：支持默认分页排序，减少按创建时间倒序取第一页的扫描成本。
-- - users(user_mobile)、users(user_name)：支持手机号/姓名精确查询和后续前缀查询优化。
-- - user_depts(dept_id, user_id)：支持按部门数据范围过滤用户。
-- - user_depts(user_id, dept_id)：支持列表返回时批量回填用户部门 ID。
--
-- 回滚：
-- ALTER TABLE `users` DROP INDEX `idx_users_add_time_id`;
-- ALTER TABLE `users` DROP INDEX `idx_users_mobile`;
-- ALTER TABLE `users` DROP INDEX `idx_users_name`;
-- ALTER TABLE `user_depts` DROP INDEX `idx_user_depts_dept_user`;
-- ALTER TABLE `user_depts` DROP INDEX `idx_user_depts_user_dept`;

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
