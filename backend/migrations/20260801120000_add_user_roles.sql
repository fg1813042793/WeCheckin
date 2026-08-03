CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `user_role_user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `user_role_role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `user_role_is_primary` tinyint NOT NULL DEFAULT 0 COMMENT '是否主角色',
  `user_role_status` tinyint NOT NULL DEFAULT 1 COMMENT '状态:1启用 0停用',
  `user_role_source` varchar(40) NOT NULL DEFAULT '' COMMENT '来源',
  `user_role_add_time` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `user_role_edit_time` bigint NOT NULL DEFAULT 0 COMMENT '修改时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_roles_user_role` (`user_role_user_id`, `user_role_role_id`),
  KEY `idx_user_roles_user_id` (`user_role_user_id`),
  KEY `idx_user_roles_role_id` (`user_role_role_id`),
  KEY `idx_user_roles_primary` (`user_role_is_primary`),
  KEY `idx_user_roles_status` (`user_role_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户多角色绑定表';

INSERT INTO `user_roles` (
  `user_role_user_id`,
  `user_role_role_id`,
  `user_role_is_primary`,
  `user_role_status`,
  `user_role_source`,
  `user_role_add_time`,
  `user_role_edit_time`,
  `created_at`,
  `updated_at`
)
SELECT `id`, `user_role_id`, 1, 1, 'legacy', `user_add_time`, `user_edit_time`, `created_at`, `updated_at`
FROM `users`
WHERE `user_role_id` > 0
ON DUPLICATE KEY UPDATE
  `user_role_is_primary` = VALUES(`user_role_is_primary`),
  `user_role_status` = VALUES(`user_role_status`),
  `user_role_source` = VALUES(`user_role_source`),
  `user_role_edit_time` = VALUES(`user_role_edit_time`),
  `updated_at` = VALUES(`updated_at`);
