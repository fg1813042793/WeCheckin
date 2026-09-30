-- 字段权限保存后立即生效，不依赖重新发布流程版本。
SET @field_permissions_json_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_definitions' AND COLUMN_NAME = 'definition_field_permissions_json');
SET @field_permissions_json_sql := IF(@field_permissions_json_exists = 0, 'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_field_permissions_json` MEDIUMTEXT NULL COMMENT ''当前生效的字段权限JSON'' AFTER `definition_start_config_revision`', 'SELECT 1');
PREPARE field_permissions_json_stmt FROM @field_permissions_json_sql;
EXECUTE field_permissions_json_stmt;
DEALLOCATE PREPARE field_permissions_json_stmt;

SET @field_permissions_revision_exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_definitions' AND COLUMN_NAME = 'definition_field_permissions_revision');
SET @field_permissions_revision_sql := IF(@field_permissions_revision_exists = 0, 'ALTER TABLE `workflow_definitions` ADD COLUMN `definition_field_permissions_revision` INT NOT NULL DEFAULT 0 COMMENT ''当前生效的字段权限修订号'' AFTER `definition_field_permissions_json`', 'SELECT 1');
PREPARE field_permissions_revision_stmt FROM @field_permissions_revision_sql;
EXECUTE field_permissions_revision_stmt;
DEALLOCATE PREPARE field_permissions_revision_stmt;
