-- User feedback aggregates, timeline messages, attachment metadata, and daily numbering.
-- All BIGINT time columns are UTC milliseconds; 0 means that an optional event has not occurred.

CREATE TABLE IF NOT EXISTS `user_feedbacks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Feedback ID',
  `feedback_no` VARCHAR(32) NOT NULL COMMENT 'Human-readable feedback number',
  `submitter_id` BIGINT UNSIGNED NOT NULL COMMENT 'Submitting user ID',
  `create_request_id` VARCHAR(64) NOT NULL COMMENT 'Create request idempotency key',
  `feedback_status` VARCHAR(24) NOT NULL DEFAULT 'pending' COMMENT 'pending, processing, resolved or closed',
  `handler_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Current admin handler ID',
  `version` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'Optimistic lock version',
  `last_activity_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Latest activity time in UTC milliseconds',
  `resolved_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Resolution time in UTC milliseconds',
  `closed_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Closure time in UTC milliseconds',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Creation time in UTC milliseconds',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Update time in UTC milliseconds',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_feedbacks_feedback_no` (`feedback_no`),
  UNIQUE KEY `uk_user_feedbacks_submitter_request` (`submitter_id`,`create_request_id`),
  KEY `idx_user_feedbacks_status_activity` (`feedback_status`,`last_activity_at`,`id`),
  KEY `idx_user_feedbacks_submitter_activity` (`submitter_id`,`last_activity_at`,`id`),
  KEY `idx_user_feedbacks_handler_activity` (`handler_id`,`last_activity_at`,`id`),
  KEY `idx_user_feedbacks_last_activity` (`last_activity_at`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='User feedback records';

CREATE TABLE IF NOT EXISTS `user_feedback_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Feedback message ID',
  `feedback_id` BIGINT UNSIGNED NOT NULL COMMENT 'Feedback ID',
  `message_type` VARCHAR(24) NOT NULL COMMENT 'initial, supplement or status',
  `author_type` VARCHAR(16) NOT NULL COMMENT 'user or admin',
  `author_id` BIGINT UNSIGNED NOT NULL COMMENT 'Message author ID',
  `content` TEXT NOT NULL COMMENT 'Message content',
  `from_status` VARCHAR(24) NOT NULL DEFAULT '' COMMENT 'Status before transition',
  `to_status` VARCHAR(24) NOT NULL DEFAULT '' COMMENT 'Status after transition',
  `request_id` VARCHAR(64) NOT NULL COMMENT 'Message request idempotency key',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Creation time in UTC milliseconds',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_feedback_messages_request` (`feedback_id`,`author_type`,`author_id`,`request_id`),
  KEY `idx_user_feedback_messages_timeline` (`feedback_id`,`created_at`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='User feedback timeline messages';

CREATE TABLE IF NOT EXISTS `user_feedback_attachments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Feedback attachment ID',
  `feedback_id` BIGINT UNSIGNED NOT NULL COMMENT 'Feedback ID',
  `message_id` BIGINT UNSIGNED NOT NULL COMMENT 'Feedback message ID',
  `storage_provider` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Object storage provider',
  `object_key` VARCHAR(512) NOT NULL COMMENT 'Object storage key',
  `original_name` VARCHAR(255) NOT NULL COMMENT 'Original file name',
  `content_type` VARCHAR(127) NOT NULL DEFAULT '' COMMENT 'Media type',
  `size_bytes` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'File size in bytes',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT 'Display order within the message',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Creation time in UTC milliseconds',
  PRIMARY KEY (`id`),
  KEY `idx_user_feedback_attachments_feedback_message` (`feedback_id`,`message_id`,`sort_order`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='User feedback attachment object metadata';

CREATE TABLE IF NOT EXISTS `user_feedback_daily_sequences` (
  `sequence_date` DATE NOT NULL COMMENT 'Sequence date',
  `current_value` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Latest value allocated for the date',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Update time in UTC milliseconds',
  PRIMARY KEY (`sequence_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Daily user feedback number sequences';
