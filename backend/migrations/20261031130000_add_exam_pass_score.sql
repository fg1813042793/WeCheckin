-- Add an explicit pass score for schema-based exams. Zero means no pass threshold.

SET @exam_pass_score_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'exam' AND COLUMN_NAME = 'exam_pass_score'
);
SET @exam_pass_score_ddl := IF(
  @exam_pass_score_exists = 0,
  'ALTER TABLE `exam` ADD COLUMN `exam_pass_score` int NOT NULL DEFAULT 0 COMMENT ''及格分，0表示无门槛'' AFTER `exam_show_score`',
  'SELECT ''exam.exam_pass_score exists'''
);
PREPARE exam_pass_score_stmt FROM @exam_pass_score_ddl;
EXECUTE exam_pass_score_stmt;
DEALLOCATE PREPARE exam_pass_score_stmt;
