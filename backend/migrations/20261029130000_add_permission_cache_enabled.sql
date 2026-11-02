-- 为后台菜单增加页面缓存配置，历史权限默认保持缓存开启。

SET @permission_cache_enabled_exists := (
  SELECT COUNT(*)
  FROM `INFORMATION_SCHEMA`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'permissions'
    AND `COLUMN_NAME` = 'permission_cache_enabled'
);

SET @permission_cache_enabled_sql := IF(
  @permission_cache_enabled_exists = 0,
  'ALTER TABLE `permissions` ADD COLUMN `permission_cache_enabled` tinyint NOT NULL DEFAULT 1 COMMENT ''菜单页面缓存:1开启 0关闭'' AFTER `permission_status`',
  'SELECT 1'
);

PREPARE permission_cache_enabled_stmt FROM @permission_cache_enabled_sql;
EXECUTE permission_cache_enabled_stmt;
DEALLOCATE PREPARE permission_cache_enabled_stmt;
