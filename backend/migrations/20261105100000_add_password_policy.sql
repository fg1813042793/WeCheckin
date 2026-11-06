-- Add password lifecycle state without changing existing passwords or forcing existing accounts.

SET @password_changed_at_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_password_changed_at'
);
SET @password_changed_at_ddl := IF(
  @password_changed_at_exists = 0,
  'ALTER TABLE `users` ADD COLUMN `user_password_changed_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''密码最后修改时间（毫秒）'' AFTER `user_password`',
  'SELECT ''users.user_password_changed_at exists'''
);
PREPARE password_changed_at_stmt FROM @password_changed_at_ddl;
EXECUTE password_changed_at_stmt;
DEALLOCATE PREPARE password_changed_at_stmt;

SET @password_must_change_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'user_password_must_change'
);
SET @password_must_change_ddl := IF(
  @password_must_change_exists = 0,
  'ALTER TABLE `users` ADD COLUMN `user_password_must_change` TINYINT NOT NULL DEFAULT 0 COMMENT ''下次登录必须修改密码'' AFTER `user_password_changed_at`',
  'SELECT ''users.user_password_must_change exists'''
);
PREPARE password_must_change_stmt FROM @password_must_change_ddl;
EXECUTE password_must_change_stmt;
DEALLOCATE PREPARE password_must_change_stmt;

CREATE TABLE IF NOT EXISTS `user_password_histories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `changed_at` BIGINT NOT NULL DEFAULT 0,
  `change_source` VARCHAR(32) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_user_password_history_user_time` (`user_id`,`changed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户密码哈希历史';
