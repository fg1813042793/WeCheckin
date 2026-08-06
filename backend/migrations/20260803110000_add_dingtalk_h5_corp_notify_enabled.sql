-- 钉钉 H5 企业应用级绩效流程通知开关。
-- 历史全局 DINGTALK_H5_NOTIFY_ENABLED 仅作为兼容镜像保留，新配置以企业应用为准。

SET @schema_name = DATABASE();

SET @has_notify_enabled = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'dingtalk_h5_corp_configs'
    AND COLUMN_NAME = 'notify_enabled'
);
SET @ddl = IF(
  @has_notify_enabled = 0,
  'ALTER TABLE `dingtalk_h5_corp_configs` ADD COLUMN `notify_enabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT ''是否开启绩效流程通知'' AFTER `unified_app_id`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`)
SELECT 'DINGTALK_H5_NOTIFY_ENABLED', '0', 'switch', CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), 0, NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_NOTIFY_ENABLED'
);

SET @legacy_notify_enabled = COALESCE((
  SELECT IF(CAST(TRIM(COALESCE(`setup_value`, '0')) AS UNSIGNED) = 1, 1, 0)
  FROM `setups`
  WHERE `setup_key` = 'DINGTALK_H5_NOTIFY_ENABLED'
  LIMIT 1
), 0);

UPDATE `dingtalk_h5_corp_configs`
SET `notify_enabled` = 1,
    `edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `updated_at` = NOW(3)
WHERE @legacy_notify_enabled = 1
  AND `notify_enabled` = 0;
