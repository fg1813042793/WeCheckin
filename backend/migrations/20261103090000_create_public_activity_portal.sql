-- Public activity portal. This migration does not publish any questionnaire or exam.
CREATE TABLE IF NOT EXISTS `public_portal_settings` (
  `id` BIGINT UNSIGNED NOT NULL,
  `enabled` TINYINT NOT NULL DEFAULT 0,
  `title` VARCHAR(100) NOT NULL DEFAULT '公开问卷与考试',
  `description` TEXT NOT NULL,
  `logo_url` VARCHAR(2048) NOT NULL DEFAULT '',
  `public_base_url` VARCHAR(512) NOT NULL DEFAULT '',
  `show_ended` TINYINT NOT NULL DEFAULT 0,
  `version` BIGINT UNSIGNED NOT NULL DEFAULT 1,
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Single public activity page configuration';

INSERT INTO `public_portal_settings` (`id`, `enabled`, `title`, `description`, `version`)
VALUES (1, 0, '公开问卷与考试', '', 1)
ON DUPLICATE KEY UPDATE `id` = `id`;

CREATE TABLE IF NOT EXISTS `public_portal_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `kind` VARCHAR(16) NOT NULL,
  `source_id` BIGINT UNSIGNED NOT NULL,
  `title` VARCHAR(255) NOT NULL DEFAULT '',
  `description` TEXT NOT NULL,
  `enabled` TINYINT NOT NULL DEFAULT 0,
  `sort` INT NOT NULL DEFAULT 0,
  `version` BIGINT UNSIGNED NOT NULL DEFAULT 1,
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_public_portal_source` (`kind`,`source_id`),
  KEY `idx_public_portal_order` (`enabled`,`sort`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Curated public survey and exam references';

SET @public_portal_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);
INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:public-portal', '公共页面管理', 'admin', 'menu',
   '', '/public-portal', 'Grid', 'public-portal:list', 21, 1, @public_portal_now_ms, @public_portal_now_ms, NOW(3), NOW(3)),
  ('admin:menu:public-portal:edit', '公共页面维护', 'admin', 'button',
   'admin:menu:public-portal', '', '', 'public-portal:edit', 1, 1, @public_portal_now_ms, @public_portal_now_ms, NOW(3), NOW(3)),
  ('admin:api-category:public-portal', '公共页面', 'admin', 'api_category',
   '', '', '', '', 95, 1, @public_portal_now_ms, @public_portal_now_ms, NOW(3), NOW(3)),
  ('admin:api:public-portal:list', '公共页面配置查看接口', 'admin', 'api',
   'admin:api-category:public-portal', '/api/v2/admin/public-portal', '', 'public-portal:list', 1, 1, @public_portal_now_ms, @public_portal_now_ms, NOW(3), NOW(3)),
  ('admin:api:public-portal:edit', '公共页面配置维护接口', 'admin', 'api',
   'admin:api-category:public-portal', '/api/v2/admin/public-portal/settings', '', 'public-portal:edit', 2, 1, @public_portal_now_ms, @public_portal_now_ms, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`), `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`), `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_icon` = VALUES(`permission_icon`), `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`), `permission_edit_time` = VALUES(`permission_edit_time`), `updated_at` = NOW(3);

-- Only the existing reserved super-administrator roles receive initial grants.
-- Existing grants (including explicit denies) are preserved on repeat execution.
INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT 'role', r.`id`, p.`permission_key`, p.`id`, 'allow', '', 'public-portal-initial', 1,
  @public_portal_now_ms, @public_portal_now_ms, NOW(3), NOW(3)
FROM `roles` r
JOIN `permissions` p ON p.`permission_key` IN (
  'admin:menu:public-portal', 'admin:menu:public-portal:edit',
  'admin:api:public-portal:list', 'admin:api:public-portal:edit'
)
WHERE r.`role_name` = '超级管理员' AND r.`role_status` = 1
  AND NOT EXISTS (
    SELECT 1 FROM `permission_grants` existing_grant
    WHERE existing_grant.`grant_subject_type` = 'role'
      AND existing_grant.`grant_subject_id` = r.`id`
      AND existing_grant.`grant_permission_key` = p.`permission_key`
  );
