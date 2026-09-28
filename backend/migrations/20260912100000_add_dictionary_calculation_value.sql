-- 为字典项增加可选的计算值。历史字典项保持 NULL，仍按原有 label/value 语义工作。
SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'sys_dicts'
    AND COLUMN_NAME = 'dict_calculation_value'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `sys_dicts` ADD COLUMN `dict_calculation_value` DECIMAL(20,8) NULL COMMENT ''字典项计算值'' AFTER `dict_value`',
  'SELECT ''dict_calculation_value exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
