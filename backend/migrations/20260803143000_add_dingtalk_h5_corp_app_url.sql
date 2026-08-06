-- 钉钉 H5 企业应用级跳转地址。
-- 通知点击时优先使用企业应用自己的 app_url；为空时回退旧全局 DINGTALK_H5_APP_URL。

SET @schema_name = DATABASE();

SET @has_app_url = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'dingtalk_h5_corp_configs'
    AND COLUMN_NAME = 'app_url'
);
SET @ddl = IF(
  @has_app_url = 0,
  'ALTER TABLE `dingtalk_h5_corp_configs` ADD COLUMN `app_url` varchar(500) NOT NULL DEFAULT '''' COMMENT ''H5应用访问地址'' AFTER `unified_app_id`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @legacy_app_url = COALESCE((
  SELECT TRIM(COALESCE(`setup_value`, ''))
  FROM `setups`
  WHERE `setup_key` = 'DINGTALK_H5_APP_URL'
  LIMIT 1
), '');

UPDATE `dingtalk_h5_corp_configs`
SET `app_url` = @legacy_app_url,
    `edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `updated_at` = NOW(3)
WHERE @legacy_app_url <> ''
  AND TRIM(COALESCE(`app_url`, '')) = '';
