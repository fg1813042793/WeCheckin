-- 钉钉 H5 免登录支持多 CorpId。
--
-- 新配置表保存每个企业的内部应用凭证；用户绑定表保存 corpId + 钉钉 userId 到本地 users.id 的映射。
-- 旧的单企业 setup 配置和 users.user_mini_openid 会被回填，保证已部署环境平滑迁移。
-- 旧 setup 键：`DINGTALK_H5_CORP_ID`、`DINGTALK_H5_APP_KEY`、`DINGTALK_H5_APP_SECRET`。

CREATE TABLE IF NOT EXISTS `dingtalk_h5_corp_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `corp_id` varchar(120) NOT NULL DEFAULT '' COMMENT '钉钉企业CorpId',
  `corp_name` varchar(120) NOT NULL DEFAULT '' COMMENT '企业名称',
  `app_key` varchar(160) NOT NULL DEFAULT '' COMMENT '钉钉内部应用AppKey',
  `app_secret` text COMMENT '钉钉内部应用AppSecret',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '是否启用',
  `add_time` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `edit_time` bigint NOT NULL DEFAULT 0 COMMENT '修改时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dt_h5_corp_id` (`corp_id`),
  KEY `idx_dt_h5_corp_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='钉钉H5企业配置';

CREATE TABLE IF NOT EXISTS `dingtalk_h5_user_bindings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '绑定ID',
  `corp_id` varchar(120) NOT NULL DEFAULT '' COMMENT '钉钉企业CorpId',
  `dingtalk_user_id` varchar(160) NOT NULL DEFAULT '' COMMENT '钉钉用户UserId',
  `union_id` varchar(160) NOT NULL DEFAULT '' COMMENT '钉钉UnionId',
  `user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '本地用户ID',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '是否启用',
  `add_time` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `edit_time` bigint NOT NULL DEFAULT 0 COMMENT '修改时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dt_h5_binding_corp_user` (`corp_id`,`dingtalk_user_id`),
  KEY `idx_dt_h5_binding_union_id` (`union_id`),
  KEY `idx_dt_h5_binding_user_id` (`user_id`),
  KEY `idx_dt_h5_binding_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='钉钉H5用户绑定';

ALTER TABLE `dingtalk_h5_corp_configs` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

ALTER TABLE `dingtalk_h5_user_bindings` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

INSERT INTO `dingtalk_h5_corp_configs` (
  `corp_id`, `corp_name`, `app_key`, `app_secret`, `enabled`,
  `add_time`, `edit_time`, `created_at`, `updated_at`
)
SELECT
  legacy.`corp_id`,
  legacy.`corp_id`,
  legacy.`app_key`,
  legacy.`app_secret`,
  1,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  NOW(3),
  NOW(3)
FROM (
  SELECT
    TRIM(COALESCE((SELECT `setup_value` FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_CORP_ID' LIMIT 1), '')) AS `corp_id`,
    TRIM(COALESCE((SELECT `setup_value` FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_APP_KEY' LIMIT 1), '')) AS `app_key`,
    TRIM(COALESCE((SELECT `setup_value` FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_APP_SECRET' LIMIT 1), '')) AS `app_secret`
) legacy
WHERE legacy.`corp_id` <> ''
ON DUPLICATE KEY UPDATE
  `corp_name` = IF(VALUES(`corp_name`) <> '', VALUES(`corp_name`), `dingtalk_h5_corp_configs`.`corp_name`),
  `app_key` = IF(VALUES(`app_key`) <> '', VALUES(`app_key`), `dingtalk_h5_corp_configs`.`app_key`),
  `app_secret` = IF(VALUES(`app_secret`) <> '', VALUES(`app_secret`), `dingtalk_h5_corp_configs`.`app_secret`),
  `enabled` = 1,
  `edit_time` = VALUES(`edit_time`),
  `updated_at` = NOW(3);

INSERT INTO `dingtalk_h5_user_bindings` (
  `corp_id`, `dingtalk_user_id`, `union_id`, `user_id`, `enabled`,
  `add_time`, `edit_time`, `created_at`, `updated_at`
)
SELECT
  cfg.`corp_id`,
  TRIM(u.`user_mini_openid`),
  '',
  u.`id`,
  1,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  NOW(3),
  NOW(3)
FROM `dingtalk_h5_corp_configs` cfg
JOIN `users` u ON TRIM(COALESCE(u.`user_mini_openid`, '')) <> ''
WHERE cfg.`corp_id` COLLATE utf8mb4_general_ci = TRIM(COALESCE((SELECT `setup_value` FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_CORP_ID' LIMIT 1), '')) COLLATE utf8mb4_general_ci
ON DUPLICATE KEY UPDATE
  `user_id` = VALUES(`user_id`),
  `enabled` = 1,
  `edit_time` = VALUES(`edit_time`),
  `updated_at` = NOW(3);
