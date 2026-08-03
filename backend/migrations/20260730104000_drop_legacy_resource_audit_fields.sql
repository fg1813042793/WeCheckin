-- 删除资源表历史审计字段，并把常用索引切到标准字段。
--
-- 标准字段：
-- - create_by、update_by、create_dept_id、update_dept_id、add_time、edit_time
--
-- 注意：
-- - 不删除业务字段：news/enroll/event/exam_publish_dept_ids、survey_dept_ids、exam_dept_ids。
-- - 不删除记录表自己的业务时间字段：enroll_join_add_time、event_part_add_time、survey_resp_add_time 等。

-- 最后一次从历史列补齐标准列，覆盖迁移执行与代码切换之间可能产生的旧写入。
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND COLUMN_NAME = 'news_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `news` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`news_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`news_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`news_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`news_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`news_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`news_edit_time`, 0), `news_add_time`, 0), `edit_time`)', 'SELECT ''news legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enrolls' AND COLUMN_NAME = 'enroll_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `enrolls` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`enroll_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`enroll_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`enroll_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`enroll_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`enroll_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`enroll_edit_time`, 0), `enroll_add_time`, 0), `edit_time`)', 'SELECT ''enrolls legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND COLUMN_NAME = 'event_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `events` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`event_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`event_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`event_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`event_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`event_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`event_edit_time`, 0), `event_add_time`, 0), `edit_time`)', 'SELECT ''events legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND COLUMN_NAME = 'survey_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `survey` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`survey_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`survey_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`survey_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`survey_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`survey_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`survey_edit_time`, 0), `survey_add_time`, 0), `edit_time`)', 'SELECT ''survey legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND COLUMN_NAME = 'exam_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `exam` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`exam_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`exam_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`exam_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`exam_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`exam_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(NULLIF(`exam_edit_time`, 0), `exam_add_time`, 0), `edit_time`)', 'SELECT ''exam legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND COLUMN_NAME = 'survey_q_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `survey_question` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`survey_q_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`survey_q_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`survey_q_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`survey_q_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`survey_q_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`survey_q_add_time`, 0), `edit_time`)', 'SELECT ''survey_question legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND COLUMN_NAME = 'exam_q_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `exam_question` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`exam_q_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`exam_q_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`exam_q_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`exam_q_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`exam_q_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`exam_q_add_time`, 0), `edit_time`)', 'SELECT ''exam_question legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_paper' AND COLUMN_NAME = 'exam_p_create_by');
SET @ddl := IF(@col_exists > 0, 'UPDATE `exam_paper` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0, COALESCE(`exam_p_create_by`, 0), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0, COALESCE(NULLIF(`exam_p_create_by`, 0), `create_by`, 0), `update_by`), `create_dept_id` = IF(COALESCE(`create_dept_id`, 0) = 0, COALESCE(`exam_p_dept_id`, 0), `create_dept_id`), `update_dept_id` = IF(COALESCE(`update_dept_id`, 0) = 0, COALESCE(NULLIF(`exam_p_dept_id`, 0), `create_dept_id`, 0), `update_dept_id`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`exam_p_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`exam_p_add_time`, 0), `edit_time`)', 'SELECT ''exam_paper legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'log_admin_id');
SET @ddl := IF(@col_exists > 0, 'UPDATE `logs` SET `create_by` = IF(COALESCE(`create_by`, 0) = 0 AND `log_admin_id` REGEXP ''^[0-9]+$'', CAST(`log_admin_id` AS UNSIGNED), `create_by`), `update_by` = IF(COALESCE(`update_by`, 0) = 0 AND `log_admin_id` REGEXP ''^[0-9]+$'', CAST(`log_admin_id` AS UNSIGNED), `update_by`), `add_time` = IF(COALESCE(`add_time`, 0) = 0, COALESCE(`log_add_time`, 0), `add_time`), `edit_time` = IF(COALESCE(`edit_time`, 0) = 0, COALESCE(`log_add_time`, 0), `edit_time`)', 'SELECT ''logs legacy audit backfill skipped''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 删除引用历史列的索引，随后用原索引名重建到标准列上。
SET @index_name := 'idx_news_status_order_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `news` DROP INDEX `idx_news_status_order_time`', 'SELECT ''idx_news_status_order_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `news` ADD INDEX `idx_news_status_order_time` (`news_status`, `news_order`, `add_time`, `id`)', 'SELECT ''idx_news_status_order_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_news_status_vouch_order_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `news` DROP INDEX `idx_news_status_vouch_order_time`', 'SELECT ''idx_news_status_vouch_order_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `news` ADD INDEX `idx_news_status_vouch_order_time` (`news_status`, `news_vouch`, `news_order`, `add_time`, `id`)', 'SELECT ''idx_news_status_vouch_order_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_news_add_time_id';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `news` DROP INDEX `idx_news_add_time_id`', 'SELECT ''idx_news_add_time_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `news` ADD INDEX `idx_news_add_time_id` (`add_time`, `id`)', 'SELECT ''idx_news_add_time_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_news_dept_create_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'news' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `news` DROP INDEX `idx_news_dept_create_time`', 'SELECT ''idx_news_dept_create_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_enrolls_status_order_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enrolls' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `enrolls` DROP INDEX `idx_enrolls_status_order_time`', 'SELECT ''idx_enrolls_status_order_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enrolls' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `enrolls` ADD INDEX `idx_enrolls_status_order_time` (`enroll_status`, `enroll_order`, `add_time`, `id`)', 'SELECT ''idx_enrolls_status_order_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_events_status_type_order_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `events` DROP INDEX `idx_events_status_type_order_time`', 'SELECT ''idx_events_status_type_order_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `events` ADD INDEX `idx_events_status_type_order_time` (`event_status`, `event_type`, `event_order`, `add_time`, `id`)', 'SELECT ''idx_events_status_type_order_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_events_add_time_id';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `events` DROP INDEX `idx_events_add_time_id`', 'SELECT ''idx_events_add_time_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `events` ADD INDEX `idx_events_add_time_id` (`add_time`, `id`)', 'SELECT ''idx_events_add_time_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_events_dept_create_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'events' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `events` DROP INDEX `idx_events_dept_create_time`', 'SELECT ''idx_events_dept_create_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_surveys_dept_create_order';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `survey` DROP INDEX `idx_surveys_dept_create_order`', 'SELECT ''idx_surveys_dept_create_order missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey` ADD INDEX `idx_surveys_dept_create_order` (`create_dept_id`, `create_by`, `survey_order`, `survey_id`)', 'SELECT ''idx_surveys_dept_create_order exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_survey_questions_category_type_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `survey_question` DROP INDEX `idx_survey_questions_category_type_time`', 'SELECT ''idx_survey_questions_category_type_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_question` ADD INDEX `idx_survey_questions_category_type_time` (`survey_q_category`, `survey_q_type`, `add_time`, `survey_q_id`)', 'SELECT ''idx_survey_questions_category_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_survey_questions_type_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `survey_question` DROP INDEX `idx_survey_questions_type_time`', 'SELECT ''idx_survey_questions_type_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'survey_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `survey_question` ADD INDEX `idx_survey_questions_type_time` (`survey_q_type`, `add_time`, `survey_q_id`)', 'SELECT ''idx_survey_questions_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_exams_dept_create_order';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `exam` DROP INDEX `idx_exams_dept_create_order`', 'SELECT ''idx_exams_dept_create_order missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam` ADD INDEX `idx_exams_dept_create_order` (`create_dept_id`, `create_by`, `exam_order`, `exam_id`)', 'SELECT ''idx_exams_dept_create_order exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_exam_questions_category_type_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `exam_question` DROP INDEX `idx_exam_questions_category_type_time`', 'SELECT ''idx_exam_questions_category_type_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_question` ADD INDEX `idx_exam_questions_category_type_time` (`exam_q_category`, `exam_q_type`, `add_time`, `exam_q_id`)', 'SELECT ''idx_exam_questions_category_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_exam_questions_type_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `exam_question` DROP INDEX `idx_exam_questions_type_time`', 'SELECT ''idx_exam_questions_type_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam_question' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `exam_question` ADD INDEX `idx_exam_questions_type_time` (`exam_q_type`, `add_time`, `exam_q_id`)', 'SELECT ''idx_exam_questions_type_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_logs_add_time_id';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `logs` DROP INDEX `idx_logs_add_time_id`', 'SELECT ''idx_logs_add_time_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `logs` ADD INDEX `idx_logs_add_time_id` (`add_time`, `id`)', 'SELECT ''idx_logs_add_time_id exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_name := 'idx_logs_admin_time';
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists > 0, 'ALTER TABLE `logs` DROP INDEX `idx_logs_admin_time`', 'SELECT ''idx_logs_admin_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @idx_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND INDEX_NAME = @index_name);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE `logs` ADD INDEX `idx_logs_admin_time` (`create_by`, `add_time`, `id`)', 'SELECT ''idx_logs_admin_time exists''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 删除历史审计列。
SET @table_name := 'news';
SET @column_name := 'news_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `news` DROP COLUMN `news_create_by`', 'SELECT ''news.news_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'news_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `news` DROP COLUMN `news_dept_id`', 'SELECT ''news.news_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'news_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `news` DROP COLUMN `news_add_time`', 'SELECT ''news.news_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'news_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `news` DROP COLUMN `news_edit_time`', 'SELECT ''news.news_edit_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'enrolls';
SET @column_name := 'enroll_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `enrolls` DROP COLUMN `enroll_create_by`', 'SELECT ''enrolls.enroll_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'enroll_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `enrolls` DROP COLUMN `enroll_dept_id`', 'SELECT ''enrolls.enroll_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'enroll_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `enrolls` DROP COLUMN `enroll_add_time`', 'SELECT ''enrolls.enroll_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'enroll_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `enrolls` DROP COLUMN `enroll_edit_time`', 'SELECT ''enrolls.enroll_edit_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'events';
SET @column_name := 'event_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `events` DROP COLUMN `event_create_by`', 'SELECT ''events.event_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'event_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `events` DROP COLUMN `event_dept_id`', 'SELECT ''events.event_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'event_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `events` DROP COLUMN `event_add_time`', 'SELECT ''events.event_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'event_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `events` DROP COLUMN `event_edit_time`', 'SELECT ''events.event_edit_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'survey';
SET @column_name := 'survey_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey` DROP COLUMN `survey_create_by`', 'SELECT ''survey.survey_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey` DROP COLUMN `survey_dept_id`', 'SELECT ''survey.survey_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey` DROP COLUMN `survey_add_time`', 'SELECT ''survey.survey_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey` DROP COLUMN `survey_edit_time`', 'SELECT ''survey.survey_edit_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam';
SET @column_name := 'exam_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam` DROP COLUMN `exam_create_by`', 'SELECT ''exam.exam_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam` DROP COLUMN `exam_dept_id`', 'SELECT ''exam.exam_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam` DROP COLUMN `exam_add_time`', 'SELECT ''exam.exam_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_edit_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam` DROP COLUMN `exam_edit_time`', 'SELECT ''exam.exam_edit_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'survey_question';
SET @column_name := 'survey_q_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey_question` DROP COLUMN `survey_q_create_by`', 'SELECT ''survey_question.survey_q_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_q_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey_question` DROP COLUMN `survey_q_dept_id`', 'SELECT ''survey_question.survey_q_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'survey_q_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `survey_question` DROP COLUMN `survey_q_add_time`', 'SELECT ''survey_question.survey_q_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam_question';
SET @column_name := 'exam_q_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam_question` DROP COLUMN `exam_q_create_by`', 'SELECT ''exam_question.exam_q_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_q_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam_question` DROP COLUMN `exam_q_dept_id`', 'SELECT ''exam_question.exam_q_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_q_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam_question` DROP COLUMN `exam_q_add_time`', 'SELECT ''exam_question.exam_q_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'exam_paper';
SET @column_name := 'exam_p_create_by';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam_paper` DROP COLUMN `exam_p_create_by`', 'SELECT ''exam_paper.exam_p_create_by missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_p_dept_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam_paper` DROP COLUMN `exam_p_dept_id`', 'SELECT ''exam_paper.exam_p_dept_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'exam_p_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `exam_paper` DROP COLUMN `exam_p_add_time`', 'SELECT ''exam_paper.exam_p_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @table_name := 'logs';
SET @column_name := 'log_admin_id';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `logs` DROP COLUMN `log_admin_id`', 'SELECT ''logs.log_admin_id missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @column_name := 'log_add_time';
SET @col_exists := (SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = @table_name AND COLUMN_NAME = @column_name);
SET @ddl := IF(@col_exists > 0, 'ALTER TABLE `logs` DROP COLUMN `log_add_time`', 'SELECT ''logs.log_add_time missing''');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
