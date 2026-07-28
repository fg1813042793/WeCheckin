-- 为打卡/报名记录相关接口增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/admin/enrollments/:id/users
-- - GET /api/v2/admin/enrollment-records
-- - GET /api/v2/admin/enrollments/:id/stats
-- - 客户端打卡详情、我的打卡、日历统计、提交去重等接口
--
-- 回滚：
-- ALTER TABLE `enroll_users` DROP INDEX `idx_enroll_users_enroll_time`;
-- ALTER TABLE `enroll_users` DROP INDEX `idx_enroll_users_enroll_openid`;
-- ALTER TABLE `enroll_users` DROP INDEX `idx_enroll_users_openid_time`;
-- ALTER TABLE `enroll_users` DROP INDEX `idx_enroll_users_enroll_rank`;
-- ALTER TABLE `enroll_joins` DROP INDEX `idx_enroll_joins_enroll_time`;
-- ALTER TABLE `enroll_joins` DROP INDEX `idx_enroll_joins_enroll_user_day`;
-- ALTER TABLE `enroll_joins` DROP INDEX `idx_enroll_joins_enroll_day_user`;
-- ALTER TABLE `enroll_joins` DROP INDEX `idx_enroll_joins_user_day_time`;
-- ALTER TABLE `enroll_joins` DROP INDEX `idx_enroll_joins_user_time`;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_users'
    AND INDEX_NAME = 'idx_enroll_users_enroll_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_users` ADD INDEX `idx_enroll_users_enroll_time` (`enroll_user_enroll_id`, `enroll_user_add_time`, `id`)',
  'SELECT ''idx_enroll_users_enroll_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_users'
    AND INDEX_NAME = 'idx_enroll_users_enroll_openid'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_users` ADD INDEX `idx_enroll_users_enroll_openid` (`enroll_user_enroll_id`, `enroll_user_mini_openid`)',
  'SELECT ''idx_enroll_users_enroll_openid exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_users'
    AND INDEX_NAME = 'idx_enroll_users_openid_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_users` ADD INDEX `idx_enroll_users_openid_time` (`enroll_user_mini_openid`, `enroll_user_add_time`, `id`)',
  'SELECT ''idx_enroll_users_openid_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_users'
    AND INDEX_NAME = 'idx_enroll_users_enroll_rank'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_users` ADD INDEX `idx_enroll_users_enroll_rank` (`enroll_user_enroll_id`, `enroll_user_join_cnt`, `enroll_user_day_cnt`)',
  'SELECT ''idx_enroll_users_enroll_rank exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_joins'
    AND INDEX_NAME = 'idx_enroll_joins_enroll_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_joins` ADD INDEX `idx_enroll_joins_enroll_time` (`enroll_join_enroll_id`, `enroll_join_add_time`, `id`)',
  'SELECT ''idx_enroll_joins_enroll_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_joins'
    AND INDEX_NAME = 'idx_enroll_joins_enroll_user_day'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_joins` ADD INDEX `idx_enroll_joins_enroll_user_day` (`enroll_join_enroll_id`, `enroll_join_user_id`, `enroll_join_day`)',
  'SELECT ''idx_enroll_joins_enroll_user_day exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_joins'
    AND INDEX_NAME = 'idx_enroll_joins_enroll_day_user'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_joins` ADD INDEX `idx_enroll_joins_enroll_day_user` (`enroll_join_enroll_id`, `enroll_join_day`, `enroll_join_user_id`)',
  'SELECT ''idx_enroll_joins_enroll_day_user exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_joins'
    AND INDEX_NAME = 'idx_enroll_joins_user_day_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_joins` ADD INDEX `idx_enroll_joins_user_day_time` (`enroll_join_user_id`, `enroll_join_day`, `enroll_join_add_time`)',
  'SELECT ''idx_enroll_joins_user_day_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'enroll_joins'
    AND INDEX_NAME = 'idx_enroll_joins_user_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `enroll_joins` ADD INDEX `idx_enroll_joins_user_time` (`enroll_join_user_id`, `enroll_join_add_time`, `id`)',
  'SELECT ''idx_enroll_joins_user_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
