-- 为收藏列表和收藏查重接口增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/favorites
-- - POST/DELETE 收藏接口按用户与对象查重、删除

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'favorites' AND INDEX_NAME = 'idx_favorites_user_time'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `favorites` ADD INDEX `idx_favorites_user_time` (`fav_user_id`, `fav_add_time`, `id`)', 'SELECT ''idx_favorites_user_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'favorites' AND INDEX_NAME = 'idx_favorites_user_oid'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `favorites` ADD INDEX `idx_favorites_user_oid` (`fav_user_id`, `fav_oid`)', 'SELECT ''idx_favorites_user_oid exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
