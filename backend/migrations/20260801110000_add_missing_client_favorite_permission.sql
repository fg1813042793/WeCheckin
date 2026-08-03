INSERT INTO `permissions` (
  `permission_key`,
  `permission_name`,
  `permission_platform`,
  `permission_type`,
  `permission_resource_path`,
  `permission_sort`,
  `permission_status`,
  `permission_add_time`,
  `permission_edit_time`,
  `created_at`,
  `updated_at`
) VALUES (
  'client:menu:favorite',
  '我的收藏',
  'client',
  'menu',
  '/pages/my/my_fav',
  105,
  1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3),
  NOW(3)
) ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = VALUES(`permission_status`),
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);
