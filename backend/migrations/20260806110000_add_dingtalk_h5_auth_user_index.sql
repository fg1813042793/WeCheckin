-- 为钉钉 H5 免登/登录态用户查找补充索引。
--
-- 适用接口：
-- - GET /api/v2/dingtalk/h5/reviews
-- - 钉钉 H5 需要按 user_mini_openid + user_status 解析当前用户的接口。
--
-- 回滚：
-- ALTER TABLE `users` DROP INDEX `idx_users_mini_openid_status_id`;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_mini_openid_status_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_mini_openid_status_id` (`user_mini_openid`, `user_status`, `id`)',
  'SELECT ''idx_users_mini_openid_status_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
