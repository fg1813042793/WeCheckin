-- Preserve complete browser and embedded WebView User-Agent values on survey responses.

SET @survey_response_device_type := (
  SELECT LOWER(DATA_TYPE)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'survey_response'
    AND COLUMN_NAME = 'survey_resp_device'
  LIMIT 1
);

SET @ddl := CASE
  WHEN @survey_response_device_type IS NULL THEN
    'SELECT ''survey_response.survey_resp_device missing'''
  WHEN @survey_response_device_type IN ('text', 'mediumtext', 'longtext') THEN
    'SELECT ''survey_response.survey_resp_device already supports full User-Agent values'''
  ELSE
    'ALTER TABLE `survey_response` MODIFY COLUMN `survey_resp_device` TEXT NULL COMMENT ''Complete client User-Agent'''
END;

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
