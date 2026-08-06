-- 钉钉 H5 新版机器人通知配置字段。
-- 旧版 AgentId 工作通知继续保留；新版机器人通知使用 RobotCode 调用 v1.0 机器人单聊接口。

SET @schema_name = DATABASE();

SET @has_unified_app_id = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'dingtalk_h5_corp_configs'
    AND COLUMN_NAME = 'unified_app_id'
);
SET @ddl = IF(
  @has_unified_app_id = 0,
  'ALTER TABLE `dingtalk_h5_corp_configs` ADD COLUMN `unified_app_id` varchar(120) NOT NULL DEFAULT '''' COMMENT ''钉钉新版应用统一ID'' AFTER `agent_id`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_notify_mode = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'dingtalk_h5_corp_configs'
    AND COLUMN_NAME = 'notify_mode'
);
SET @ddl = IF(
  @has_notify_mode = 0,
  'ALTER TABLE `dingtalk_h5_corp_configs` ADD COLUMN `notify_mode` varchar(30) NOT NULL DEFAULT ''agent'' COMMENT ''通知模式:agent/robot'' AFTER `unified_app_id`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_robot_code = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'dingtalk_h5_corp_configs'
    AND COLUMN_NAME = 'robot_code'
);
SET @ddl = IF(
  @has_robot_code = 0,
  'ALTER TABLE `dingtalk_h5_corp_configs` ADD COLUMN `robot_code` varchar(160) NOT NULL DEFAULT '''' COMMENT ''钉钉机器人编码'' AFTER `notify_mode`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`)
SELECT 'DINGTALK_H5_UNIFIED_APP_ID', '', 'string', CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), 0, NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_UNIFIED_APP_ID'
);

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`)
SELECT 'DINGTALK_H5_NOTIFY_MODE', 'agent', 'string', CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), 0, NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_NOTIFY_MODE'
);

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`)
SELECT 'DINGTALK_H5_ROBOT_CODE', '', 'string', CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), 0, NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_ROBOT_CODE'
);

UPDATE `dingtalk_h5_corp_configs`
SET `notify_mode` = 'agent',
    `edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `updated_at` = NOW(3)
WHERE TRIM(COALESCE(`notify_mode`, '')) = '';
