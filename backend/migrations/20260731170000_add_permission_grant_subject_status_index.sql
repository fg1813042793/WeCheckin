-- 补齐统一授权表按主体读取权限快照的运行时索引。
--
-- 适用接口：
-- - /api/v2/admin/* 接口鉴权
-- - /api/v2/client/* 接口鉴权
-- - /api/v2/dingtalk/h5/* 接口鉴权
--
-- 查询模式：
-- WHERE grant_subject_type = ?
--   AND grant_subject_id = ?
--   AND grant_status = 1
--   AND grant_permission_key LIKE ?
--
-- 说明：
-- - 历史环境可能已经存在该索引，本迁移保持幂等。

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'permission_grants'
    AND INDEX_NAME = 'idx_permission_grants_subject_status_key'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `permission_grants` ADD INDEX `idx_permission_grants_subject_status_key` (`grant_subject_type`, `grant_subject_id`, `grant_status`, `grant_permission_key`, `grant_effect`)',
  'SELECT ''idx_permission_grants_subject_status_key exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
