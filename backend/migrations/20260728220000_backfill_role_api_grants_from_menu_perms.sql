-- 根据已有后台菜单授权补齐后台接口授权。
--
-- 背景：
-- - 早期角色可能只从菜单/按钮权限迁移出了 admin:menu:* 授权。
-- - 接口权限收紧后，接口访问应只依赖 admin:api:*。
-- - 本迁移不依赖已清理的 role_menus 表，直接从 permission_grants 的菜单授权反推接口授权。

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  menu_grant.`grant_subject_type`,
  menu_grant.`grant_subject_id`,
  api_perm.`permission_key`,
  api_perm.`id`,
  'allow',
  '',
  'menu-api-backfill',
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
FROM `permission_grants` menu_grant
JOIN `permissions` menu_perm
  ON menu_perm.`permission_key` = menu_grant.`grant_permission_key`
  AND menu_perm.`permission_platform` = 'admin'
  AND menu_perm.`permission_type` IN ('directory', 'menu', 'button')
  AND menu_perm.`permission_perms` <> ''
JOIN `permissions` api_perm
  ON api_perm.`permission_platform` = 'admin'
  AND api_perm.`permission_type` = 'api'
  AND api_perm.`permission_perms` <> ''
  AND FIND_IN_SET(api_perm.`permission_perms`, REPLACE(menu_perm.`permission_perms`, ' ', '')) > 0
WHERE menu_grant.`grant_subject_type` = 'role'
  AND menu_grant.`grant_effect` = 'allow'
  AND menu_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
