-- 为 Admin 手动发送和定时发送站内信增加幂等键与权限。
-- 结构和权限仅通过迁移创建，不依赖服务启动时自动补齐。

SET @has_delivery_key = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND COLUMN_NAME = 'notify_delivery_key'
);
SET @ddl = IF(
  @has_delivery_key = 0,
  'ALTER TABLE `notify` ADD COLUMN `notify_delivery_key` varchar(64) NULL DEFAULT NULL COMMENT ''新站内信投递幂等键'' AFTER `notify_user_id`',
  'SELECT ''notify.notify_delivery_key exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_delivery_key_index = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notify'
    AND INDEX_NAME = 'uk_notify_delivery_key'
);
SET @ddl = IF(
  @has_delivery_key_index = 0,
  'ALTER TABLE `notify` ADD UNIQUE INDEX `uk_notify_delivery_key` (`notify_delivery_key`)',
  'SELECT ''notify.uk_notify_delivery_key exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:notification', '站内信管理', 'admin', 'menu', '', '/notifications', 'Bell', 'notification:list', 19, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:notification:list', '站内信查看', 'admin', 'button', 'admin:menu:notification', '', '', 'notification:list', 1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:notification:read', '站内信已读', 'admin', 'button', 'admin:menu:notification', '', '', 'notification:read', 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:notification:send', '发送站内信', 'admin', 'button', 'admin:menu:notification', '', '', 'notification:send', 3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api-category:notification', '站内信', 'admin', 'api_category', '', '', '', '', 85, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:notification:list', '站内信查看接口', 'admin', 'api', 'admin:api-category:notification', '/api/v2/admin/in-app-notifications', '', 'notification:list', 10, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:notification:read', '站内信已读接口', 'admin', 'api', 'admin:api-category:notification', '/api/v2/admin/in-app-notifications/:id/read', '', 'notification:read', 20, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:notification:send', '站内信发送接口', 'admin', 'api', 'admin:api-category:notification', '/api/v2/admin/in-app-notifications', '', 'notification:send', 30, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_icon` = VALUES(`permission_icon`),
  `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '',
  'in-app-notification-permission-backfill', 1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'admin:menu:survey' AS source_key, 'admin:menu:notification' AS target_key
  UNION ALL SELECT 'admin:menu:survey:button:list', 'admin:menu:notification:list'
  UNION ALL SELECT 'admin:menu:survey:button:list', 'admin:menu:notification:read'
  UNION ALL SELECT 'admin:api:survey:list', 'admin:api:notification:list'
  UNION ALL SELECT 'admin:api:survey:list', 'admin:api:notification:read'
  UNION ALL SELECT 'admin:menu:news:add', 'admin:menu:notification:send'
  UNION ALL SELECT 'admin:api:news:add', 'admin:api:notification:send'
) mapping ON mapping.source_key = source_grant.`grant_permission_key`
JOIN `permissions` target_perm ON target_perm.`permission_key` = mapping.target_key
WHERE source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
