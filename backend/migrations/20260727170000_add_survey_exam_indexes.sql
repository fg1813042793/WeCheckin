-- 为问卷、答卷、考试、考试记录和题库相关接口增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/admin/surveys、GET /api/v2/surveys
-- - GET /api/v2/admin/survey-responses、问卷统计、提交限制检查
-- - GET /api/v2/admin/exams、GET /api/v2/exams、考试记录列表、考试统计、提交限制检查
-- - 问卷/考试题库列表与分类

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = 'idx_surveys_status_order_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey` ADD INDEX `idx_surveys_status_order_id` (`survey_status`, `survey_order`, `survey_id`)', 'SELECT ''idx_surveys_status_order_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = 'idx_surveys_category_status_order'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey` ADD INDEX `idx_surveys_category_status_order` (`survey_category`, `survey_status`, `survey_order`, `survey_id`)', 'SELECT ''idx_surveys_category_status_order exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = 'idx_surveys_title'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey` ADD INDEX `idx_surveys_title` (`survey_title`)', 'SELECT ''idx_surveys_title exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = 'idx_surveys_dept_create_order'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey` ADD INDEX `idx_surveys_dept_create_order` (`survey_dept_id`, `survey_create_by`, `survey_order`, `survey_id`)', 'SELECT ''idx_surveys_dept_create_order exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_survey_id_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_survey_id_id` (`survey_resp_survey_id`, `survey_resp_id`)', 'SELECT ''idx_survey_resp_survey_id_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_survey_add_time'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_survey_add_time` (`survey_resp_survey_id`, `survey_resp_add_time`)', 'SELECT ''idx_survey_resp_survey_add_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_survey_status_add'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_survey_status_add` (`survey_resp_survey_id`, `survey_resp_status`, `survey_resp_add_time`)', 'SELECT ''idx_survey_resp_survey_status_add exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_survey_user'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_survey_user` (`survey_resp_survey_id`, `survey_resp_user_id`)', 'SELECT ''idx_survey_resp_survey_user exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_survey_device'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_survey_device` (`survey_resp_survey_id`, `survey_resp_device_id`, `survey_resp_status`)', 'SELECT ''idx_survey_resp_survey_device exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_survey_ip'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_survey_ip` (`survey_resp_survey_id`, `survey_resp_ip`, `survey_resp_status`)', 'SELECT ''idx_survey_resp_survey_ip exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_response' AND INDEX_NAME = 'idx_survey_resp_user_status_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_response` ADD INDEX `idx_survey_resp_user_status_id` (`survey_resp_user_id`, `survey_resp_status`, `survey_resp_id`)', 'SELECT ''idx_survey_resp_user_status_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_channel' AND INDEX_NAME = 'idx_survey_channels_survey_id_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_channel` ADD INDEX `idx_survey_channels_survey_id_id` (`survey_ch_survey_id`, `survey_ch_id`)', 'SELECT ''idx_survey_channels_survey_id_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = 'idx_survey_questions_category_type_time'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_question` ADD INDEX `idx_survey_questions_category_type_time` (`survey_q_category`, `survey_q_type`, `survey_q_add_time`, `survey_q_id`)', 'SELECT ''idx_survey_questions_category_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = 'idx_survey_questions_type_time'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_question` ADD INDEX `idx_survey_questions_type_time` (`survey_q_type`, `survey_q_add_time`, `survey_q_id`)', 'SELECT ''idx_survey_questions_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = 'idx_survey_questions_title'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_question` ADD INDEX `idx_survey_questions_title` (`survey_q_title`)', 'SELECT ''idx_survey_questions_title exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = 'idx_exams_status_order_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam` ADD INDEX `idx_exams_status_order_id` (`exam_status`, `exam_order`, `exam_id`)', 'SELECT ''idx_exams_status_order_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = 'idx_exams_category_status_order'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam` ADD INDEX `idx_exams_category_status_order` (`exam_category`, `exam_status`, `exam_order`, `exam_id`)', 'SELECT ''idx_exams_category_status_order exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = 'idx_exams_title'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam` ADD INDEX `idx_exams_title` (`exam_title`)', 'SELECT ''idx_exams_title exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = 'idx_exams_dept_create_order'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam` ADD INDEX `idx_exams_dept_create_order` (`exam_dept_id`, `exam_create_by`, `exam_order`, `exam_id`)', 'SELECT ''idx_exams_dept_create_order exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_exam_id_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_exam_id_id` (`exam_r_exam_id`, `exam_r_id`)', 'SELECT ''idx_exam_records_exam_id_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_exam_status_submit'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_exam_status_submit` (`exam_r_exam_id`, `exam_r_status`, `exam_r_submit_time`)', 'SELECT ''idx_exam_records_exam_status_submit exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_exam_pass'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_exam_pass` (`exam_r_exam_id`, `exam_r_pass`)', 'SELECT ''idx_exam_records_exam_pass exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_exam_user_status_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_exam_user_status_id` (`exam_r_exam_id`, `exam_r_user_id`, `exam_r_status`, `exam_r_id`)', 'SELECT ''idx_exam_records_exam_user_status_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_exam_device_status'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_exam_device_status` (`exam_r_exam_id`, `exam_r_device_id`, `exam_r_status`)', 'SELECT ''idx_exam_records_exam_device_status exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_exam_ip_status'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_exam_ip_status` (`exam_r_exam_id`, `exam_r_add_ip`, `exam_r_status`)', 'SELECT ''idx_exam_records_exam_ip_status exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_user_id'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_user_id` (`exam_r_user_id`, `exam_r_id`)', 'SELECT ''idx_exam_records_user_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_record' AND INDEX_NAME = 'idx_exam_records_session'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_record` ADD INDEX `idx_exam_records_session` (`exam_r_session`)', 'SELECT ''idx_exam_records_session exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = 'idx_exam_questions_category_type_time'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_question` ADD INDEX `idx_exam_questions_category_type_time` (`exam_q_category`, `exam_q_type`, `exam_q_add_time`, `exam_q_id`)', 'SELECT ''idx_exam_questions_category_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = 'idx_exam_questions_type_time'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_question` ADD INDEX `idx_exam_questions_type_time` (`exam_q_type`, `exam_q_add_time`, `exam_q_id`)', 'SELECT ''idx_exam_questions_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = 'idx_exam_questions_title'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_question` ADD INDEX `idx_exam_questions_title` (`exam_q_title`)', 'SELECT ''idx_exam_questions_title exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
