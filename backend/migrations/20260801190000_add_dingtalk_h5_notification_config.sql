-- 钉钉 H5 绩效流程通知配置。
-- AgentId 用于调用钉钉工作通知接口；通知开关默认关闭，避免历史部署升级后自动发消息。

SET @schema_name = DATABASE();

SET @has_agent_id = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'dingtalk_h5_corp_configs'
    AND COLUMN_NAME = 'agent_id'
);

SET @ddl = IF(
  @has_agent_id = 0,
  'ALTER TABLE `dingtalk_h5_corp_configs` ADD COLUMN `agent_id` varchar(80) NOT NULL DEFAULT '''' COMMENT ''钉钉内部应用AgentId'' AFTER `app_secret`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`)
SELECT 'DINGTALK_H5_AGENT_ID', '', 'string', CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), 0, NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_AGENT_ID'
);

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`)
SELECT 'DINGTALK_H5_NOTIFY_ENABLED', '0', 'switch', CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), 0, NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_NOTIFY_ENABLED'
);

UPDATE `dingtalk_h5_corp_configs`
SET `agent_id` = TRIM(COALESCE((SELECT `setup_value` FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_AGENT_ID' LIMIT 1), '')),
    `edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `updated_at` = NOW(3)
WHERE TRIM(COALESCE(`agent_id`, '')) = ''
  AND TRIM(COALESCE((SELECT `setup_value` FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_AGENT_ID' LIMIT 1), '')) <> '';
