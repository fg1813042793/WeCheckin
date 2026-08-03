-- 清理已执行统一审计迁移中的非标准字段，并补齐 update_dept_id。
--
-- 标准字段集合：
-- - create_by
-- - update_by
-- - create_dept_id
-- - update_dept_id
-- - add_time
-- - edit_time
--
-- 说明：
-- - 20260730100000 已执行环境不能再改旧迁移，本文件作为后续修正迁移。
-- - 历史业务前缀字段（如 news_create_by、exam_dept_id）仍被现有模型读写，本迁移暂不删除。
-- - create_dept_path、create_account、update_account 不再属于标准审计字段，本迁移删除这些冗余列。
--
-- 回滚参考：
-- - 如需回滚结构，可按业务表重新 ADD create_dept_path，并按 H5 表重新 ADD create_account/update_account。
-- - update_dept_id 为标准字段，不建议回滚删除。

SET @table_name := 'news';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'enrolls';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'events';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'survey';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'survey_question';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam_question';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam_paper';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'logs';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''统一更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_sessions';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @index_name := 'idx_dingtalk_h5_perf_sessions_create_account';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP INDEX `', @index_name, '`'), CONCAT('SELECT ''', @table_name, '.', @index_name, ' missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_account`'), CONCAT('SELECT ''', @table_name, '.create_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `update_account`'), CONCAT('SELECT ''', @table_name, '.update_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_reviews';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @index_name := 'idx_dt_h5_review_create_account';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP INDEX `', @index_name, '`'), CONCAT('SELECT ''', @table_name, '.', @index_name, ' missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @index_name := 'idx_dingtalk_h5_perf_reviews_create_account';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP INDEX `', @index_name, '`'), CONCAT('SELECT ''', @table_name, '.', @index_name, ' missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_account`'), CONCAT('SELECT ''', @table_name, '.create_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `update_account`'), CONCAT('SELECT ''', @table_name, '.update_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_histories';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @index_name := 'idx_dingtalk_h5_perf_histories_create_account';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP INDEX `', @index_name, '`'), CONCAT('SELECT ''', @table_name, '.', @index_name, ' missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_account`'), CONCAT('SELECT ''', @table_name, '.create_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `update_account`'), CONCAT('SELECT ''', @table_name, '.update_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'dingtalk_h5_perf_templates';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_dept_id');
SET @ddl := IF(@col_exists = 0, CONCAT('ALTER TABLE `', @table_name, '` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人部门ID'' AFTER `create_dept_id`'), CONCAT('SELECT ''', @table_name, '.update_dept_id exists'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @ddl := CONCAT('UPDATE `', @table_name, '` SET `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(`create_dept_id`, 0), `update_dept_id`)');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_dept_path');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_dept_path`'), CONCAT('SELECT ''', @table_name, '.create_dept_path missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @index_name := 'idx_dingtalk_h5_perf_templates_create_account';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP INDEX `', @index_name, '`'), CONCAT('SELECT ''', @table_name, '.', @index_name, ' missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'create_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `create_account`'), CONCAT('SELECT ''', @table_name, '.create_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = 'update_account');
SET @ddl := IF(@col_exists > 0, CONCAT('ALTER TABLE `', @table_name, '` DROP COLUMN `update_account`'), CONCAT('SELECT ''', @table_name, '.update_account missing'''));
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
