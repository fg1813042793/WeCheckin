-- 增加岗位基础表，并在 users 表增加岗位字段。
--
-- 适用场景：
-- - 后台用户管理展示/维护岗位
-- - 后续钉钉 H5、组织架构等模块复用同一套用户岗位关系
--
-- 回滚：
-- ALTER TABLE `users` DROP INDEX `idx_users_position_id`;
-- ALTER TABLE `users` DROP COLUMN `user_position_id`;
-- DROP TABLE IF EXISTS `positions`;

CREATE TABLE IF NOT EXISTS `positions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '岗位ID',
  `position_name` varchar(100) DEFAULT NULL COMMENT '岗位名称',
  `position_sort` bigint DEFAULT 0 COMMENT '排序',
  `position_status` bigint DEFAULT 1 COMMENT '状态:1正常 0禁用',
  `position_add_time` bigint DEFAULT 0 COMMENT '创建时间',
  `position_edit_time` bigint DEFAULT 0 COMMENT '修改时间',
  `position_add_ip` varchar(50) DEFAULT NULL COMMENT '创建IP',
  `position_edit_ip` varchar(50) DEFAULT NULL COMMENT '修改IP',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_positions_status_sort` (`position_status`, `position_sort`, `id`),
  KEY `idx_positions_name` (`position_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位表';

SET @user_position_col_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'user_position_id'
);
SET @ddl := IF(
  @user_position_col_exists = 0,
  'ALTER TABLE `users` ADD COLUMN `user_position_id` bigint unsigned DEFAULT 0 COMMENT ''岗位ID'' AFTER `user_mobile`',
  'SELECT ''user_position_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_position_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_users_position_id` (`user_position_id`)',
  'SELECT ''idx_users_position_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
