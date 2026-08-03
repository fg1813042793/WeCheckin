-- 补齐系统配置按 key 读取索引。
--
-- 适用接口：
-- - GET /api/v2/user-form-fields
-- - GET /api/v2/home/setup
--
-- 说明：
-- - 新库通常会通过模型唯一索引创建 setup_key 索引。
-- - 历史库可能缺少索引或索引名不一致，本迁移按列检查，避免重复创建。

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'setups'
    AND COLUMN_NAME = 'setup_key'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `setups` ADD INDEX `idx_setups_setup_key` (`setup_key`)',
  'SELECT ''setup_key index exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
