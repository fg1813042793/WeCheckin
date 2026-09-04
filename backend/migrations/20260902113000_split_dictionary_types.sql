-- Split dictionary type metadata from historical dictionary item rows.
-- Existing sys_dicts rows, including legacy empty-value type placeholders, are preserved.

CREATE TABLE IF NOT EXISTS `sys_dict_types` (
  `dict_type_code` VARCHAR(50) NOT NULL COMMENT 'Stable dictionary type code',
  `dict_type_name` VARCHAR(100) NOT NULL COMMENT 'Dictionary type name',
  `dict_type_status` TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 0 disabled',
  `dict_type_remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'Dictionary type remark',
  `dict_add_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'Creation time in milliseconds',
  `dict_edit_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'Update time in milliseconds',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`dict_type_code`),
  KEY `idx_sys_dict_types_status` (`dict_type_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Dictionary type definitions';

INSERT INTO `sys_dict_types` (
  `dict_type_code`,
  `dict_type_name`,
  `dict_type_status`,
  `dict_type_remark`,
  `dict_add_time`,
  `dict_edit_time`,
  `created_at`,
  `updated_at`
)
SELECT
  `dict_type_code`,
  COALESCE(NULLIF(MAX(TRIM(`dict_type_name`)), ''), `dict_type_code`),
  1,
  '',
  MIN(COALESCE(`dict_add_time`, 0)),
  MAX(COALESCE(`dict_edit_time`, 0)),
  MIN(`created_at`),
  MAX(`updated_at`)
FROM `sys_dicts`
WHERE TRIM(COALESCE(`dict_type_code`, '')) <> ''
GROUP BY `dict_type_code`
ON DUPLICATE KEY UPDATE
  `dict_type_code` = VALUES(`dict_type_code`);
