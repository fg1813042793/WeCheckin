-- 历史角色后台登录开关兼容字段。
--
-- 说明：
-- - 当前后台登录准入不再读取 `role_allow_admin_login`。
-- - 后台登录准入以角色启用、绑定后台菜单/按钮权限或保留超级管理员角色为准。
-- - 历史超级管理员如未绑定角色，会自动绑定到“超级管理员”角色。

SET @col_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'roles'
    AND COLUMN_NAME = 'role_allow_admin_login'
);

SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `roles` ADD COLUMN `role_allow_admin_login` tinyint DEFAULT 1 COMMENT ''是否允许后台登录'' AFTER `role_status`',
  'SELECT ''role_allow_admin_login exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `roles` SET `role_allow_admin_login` = 1
WHERE `role_allow_admin_login` IS NULL OR `role_allow_admin_login` NOT IN (0, 1);

INSERT INTO `roles` (
  `role_name`,
  `role_remark`,
  `role_sort`,
  `role_status`,
  `role_allow_admin_login`,
  `role_data_scope`,
  `role_add_time`,
  `role_edit_time`,
  `role_add_ip`,
  `role_edit_ip`,
  `created_at`,
  `updated_at`
)
SELECT
  '超级管理员',
  '系统内置角色',
  0,
  1,
  1,
  1,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  '127.0.0.1',
  '127.0.0.1',
  NOW(3),
  NOW(3)
WHERE NOT EXISTS (SELECT 1 FROM `roles` WHERE `role_name` = '超级管理员');

SELECT `id` INTO @super_admin_role_id
FROM `roles`
WHERE `role_name` = '超级管理员'
ORDER BY `id`
LIMIT 1;

UPDATE `roles`
SET `role_status` = 1,
    `role_allow_admin_login` = 1,
    `role_edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `updated_at` = NOW(3)
WHERE `id` = @super_admin_role_id;

UPDATE `users`
SET `user_role_id` = @super_admin_role_id,
    `user_edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `updated_at` = NOW(3)
WHERE `user_admin_type` = 1
  AND (`user_role_id` IS NULL OR `user_role_id` = 0);
