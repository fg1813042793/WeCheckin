-- 清理旧后台账号独立表。
--
-- 背景：
-- - 后台管理员已合并到 `users` 表，`model.Admin` 当前映射到 `users`。
-- - 后台角色继续使用 `roles`；菜单资源已迁移到 `permissions`；角色授权由 `permission_grants` 承载。
-- - 后台数据权限部门统一复用 `user_depts`。
-- - `logs` 仍为后台操作日志表。
--
-- 执行前置：
-- - 已执行 `20260728120000_merge_admins_into_users.sql` 或启动迁移 `merge_admins_into_users`。
-- - 确认 `users.user_admin_enabled = 1` 中已包含原后台管理员账号。
--
-- 本迁移只删除不再使用的旧表：
-- - `admins`：旧后台账号表。
-- - `admin_user_merge_maps`：旧 admins.id 到 users.id 的临时映射表。
-- - `admin_depts`：旧后台账号数据权限部门表，删除前会合并到 user_depts。

SET @legacy_admin_depts_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'admin_depts'
);

SET @legacy_admin_merge_map_table_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'admin_user_merge_maps'
);

SET @legacy_admin_dept_cleanup_sql := IF(
  @legacy_admin_depts_table_exists > 0 AND @legacy_admin_merge_map_table_exists > 0,
  'INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`)
   SELECT DISTINCT COALESCE(m.`user_id`, ad.`admin_dept_admin_id`), ad.`admin_dept_dept_id`, NOW(3), NOW(3)
   FROM `admin_depts` ad
   LEFT JOIN `admin_user_merge_maps` m ON m.`legacy_admin_id` = ad.`admin_dept_admin_id`
   INNER JOIN `users` u ON u.`id` = COALESCE(m.`user_id`, ad.`admin_dept_admin_id`)
   LEFT JOIN `user_depts` ud
     ON ud.`user_dept_user_id` = COALESCE(m.`user_id`, ad.`admin_dept_admin_id`)
    AND ud.`user_dept_dept_id` = ad.`admin_dept_dept_id`
   WHERE ad.`admin_dept_dept_id` > 0 AND ud.`id` IS NULL',
  IF(
    @legacy_admin_depts_table_exists > 0,
    'INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`)
     SELECT DISTINCT ad.`admin_dept_admin_id`, ad.`admin_dept_dept_id`, NOW(3), NOW(3)
     FROM `admin_depts` ad
     INNER JOIN `users` u ON u.`id` = ad.`admin_dept_admin_id`
     LEFT JOIN `user_depts` ud
       ON ud.`user_dept_user_id` = ad.`admin_dept_admin_id`
      AND ud.`user_dept_dept_id` = ad.`admin_dept_dept_id`
     WHERE ad.`admin_dept_dept_id` > 0 AND ud.`id` IS NULL',
    'SELECT ''admin_depts table does not exist, skip data migration'''
  )
);

PREPARE stmt FROM @legacy_admin_dept_cleanup_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TABLE IF EXISTS `admin_depts`;
DROP TABLE IF EXISTS `admins`;
DROP TABLE IF EXISTS `admin_user_merge_maps`;
