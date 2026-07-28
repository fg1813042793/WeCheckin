-- 为统一权限表补充图标字段。
--
-- 适用场景：
-- - 后台权限管理直接维护菜单权限时，需要从 permissions 读取图标。
-- - 已执行过 20260728170000_add_unified_permissions.sql 的数据库需要补列。
--
-- 回滚：
-- ALTER TABLE `permissions` DROP COLUMN `permission_icon`;

SET @permission_icon_col_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'permissions'
    AND COLUMN_NAME = 'permission_icon'
);
SET @ddl := IF(
  @permission_icon_col_exists = 0,
  'ALTER TABLE `permissions` ADD COLUMN `permission_icon` varchar(100) DEFAULT '''' COMMENT ''图标'' AFTER `permission_resource_path`',
  'SELECT ''permission_icon exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @legacy_menus_exists := (
  SELECT COUNT(*)
  FROM `INFORMATION_SCHEMA`.`TABLES`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'menus'
);
SET @permission_icon_backfill_sql := IF(
  @legacy_menus_exists > 0,
  'UPDATE `permissions` p
  JOIN `menus` m ON p.`permission_resource_id` = m.`id`
  SET p.`permission_icon` = COALESCE(m.`menu_icon`, '''')
  WHERE p.`permission_platform` = ''admin''
    AND p.`permission_type` IN (''directory'', ''menu'', ''button'')
    AND (p.`permission_icon` IS NULL OR p.`permission_icon` = '''')',
  'SELECT ''legacy menus icon backfill skipped'''
);
PREPARE permission_icon_backfill_stmt FROM @permission_icon_backfill_sql;
EXECUTE permission_icon_backfill_stmt;
DEALLOCATE PREPARE permission_icon_backfill_stmt;
