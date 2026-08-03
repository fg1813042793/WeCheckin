-- 补齐统一授权表高频读取索引。
--
-- 适用接口：
-- - GET /api/v2/admin/roles?page=1&pageSize=20
-- - 权限校验、应用菜单/接口权限批量读取
--
-- 说明：
-- - 历史环境可能已经存在该索引，本迁移保持幂等。

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'permission_grants'
    AND INDEX_NAME = 'idx_permission_grants_subject_effect_status_key'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `permission_grants` ADD INDEX `idx_permission_grants_subject_effect_status_key` (`grant_subject_type`, `grant_effect`, `grant_status`, `grant_subject_id`, `grant_permission_key`)',
  'SELECT ''idx_permission_grants_subject_effect_status_key exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
