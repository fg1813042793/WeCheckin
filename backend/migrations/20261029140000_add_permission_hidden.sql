-- 为后台目录和菜单增加左侧导航隐藏配置，历史权限默认保持显示。

SET @permission_hidden_exists := (
  SELECT COUNT(*)
  FROM `INFORMATION_SCHEMA`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'permissions'
    AND `COLUMN_NAME` = 'permission_hidden'
);

SET @permission_hidden_sql := IF(
  @permission_hidden_exists = 0,
  'ALTER TABLE `permissions` ADD COLUMN `permission_hidden` tinyint NOT NULL DEFAULT 0 COMMENT ''左侧菜单隐藏:1隐藏 0显示'' AFTER `permission_cache_enabled`',
  'SELECT 1'
);

PREPARE permission_hidden_stmt FROM @permission_hidden_sql;
EXECUTE permission_hidden_stmt;
DEALLOCATE PREPARE permission_hidden_stmt;
