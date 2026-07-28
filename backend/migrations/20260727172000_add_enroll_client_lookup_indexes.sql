-- 为客户端打卡列表“是否已参与/已打卡”状态查询增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/enrolls

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enroll_joins' AND INDEX_NAME = 'idx_enroll_joins_user_enroll'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `enroll_joins` ADD INDEX `idx_enroll_joins_user_enroll` (`enroll_join_user_id`, `enroll_join_enroll_id`)', 'SELECT ''idx_enroll_joins_user_enroll exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enroll_users' AND INDEX_NAME = 'idx_enroll_users_openid_enroll'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `enroll_users` ADD INDEX `idx_enroll_users_openid_enroll` (`enroll_user_mini_openid`, `enroll_user_enroll_id`)', 'SELECT ''idx_enroll_users_openid_enroll exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
