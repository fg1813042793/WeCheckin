-- Reusable manual-send templates; no startup DDL or changes to existing messages.
CREATE TABLE IF NOT EXISTS `dingtalk_notification_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(80) NOT NULL,
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `title` VARCHAR(255) NOT NULL,
  `content` TEXT NOT NULL,
  `message_type` VARCHAR(20) NOT NULL,
  `target_kind` VARCHAR(20) NOT NULL,
  `definition_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `definition_name` VARCHAR(200) NOT NULL DEFAULT '',
  `target_url` VARCHAR(2048) NOT NULL DEFAULT '',
  `button_title` VARCHAR(20) NOT NULL DEFAULT '打开页面',
  `head_color` VARCHAR(8) NOT NULL DEFAULT 'FF1677FF',
  `version` INT NOT NULL DEFAULT 1,
  `created_by` BIGINT UNSIGNED NOT NULL,
  `updated_by` BIGINT UNSIGNED NOT NULL,
  `created_at` BIGINT NOT NULL COMMENT 'UTC milliseconds',
  `updated_at` BIGINT NOT NULL COMMENT 'UTC milliseconds',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dingtalk_notification_template_name` (`name`),
  KEY `idx_dingtalk_notification_template_enabled_id` (`enabled`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='DingTalk reusable manual-send templates';
