-- 为赛事活动相关接口增加 MySQL 索引。
--
-- 适用接口：
-- - GET /api/v2/admin/events
-- - GET /api/v2/admin/events/:id/participants
-- - GET /api/v2/events、GET /api/v2/events/my、GET /api/v2/events/managed
-- - 赛事动态、成绩列表、角色判断等接口
--
-- 回滚：
-- ALTER TABLE `events` DROP INDEX `idx_events_status_type_order_time`;
-- ALTER TABLE `events` DROP INDEX `idx_events_add_time_id`;
-- ALTER TABLE `events` DROP INDEX `idx_events_title`;
-- ALTER TABLE `events` DROP INDEX `idx_events_dept_create_time`;
-- ALTER TABLE `event_participants` DROP INDEX `idx_event_parts_event_time`;
-- ALTER TABLE `event_participants` DROP INDEX `idx_event_parts_openid_event`;
-- ALTER TABLE `event_participants` DROP INDEX `idx_event_parts_event_openid`;
-- ALTER TABLE `event_roles` DROP INDEX `idx_event_roles_user_event`;
-- ALTER TABLE `event_roles` DROP INDEX `idx_event_roles_event_user`;
-- ALTER TABLE `event_dynamics` DROP INDEX `idx_event_dynamics_event_time`;
-- ALTER TABLE `event_scores` DROP INDEX `idx_event_scores_event_time`;
-- ALTER TABLE `event_scores` DROP INDEX `idx_event_scores_event_participant`;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'events'
    AND INDEX_NAME = 'idx_events_status_type_order_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `events` ADD INDEX `idx_events_status_type_order_time` (`event_status`, `event_type`, `event_order`, `event_add_time`, `id`)',
  'SELECT ''idx_events_status_type_order_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'events'
    AND INDEX_NAME = 'idx_events_add_time_id'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `events` ADD INDEX `idx_events_add_time_id` (`event_add_time`, `id`)',
  'SELECT ''idx_events_add_time_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'events'
    AND INDEX_NAME = 'idx_events_title'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `events` ADD INDEX `idx_events_title` (`event_title`)',
  'SELECT ''idx_events_title exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'events'
    AND INDEX_NAME = 'idx_events_dept_create_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `events` ADD INDEX `idx_events_dept_create_time` (`event_dept_id`, `event_create_by`, `event_add_time`, `id`)',
  'SELECT ''idx_events_dept_create_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_participants'
    AND INDEX_NAME = 'idx_event_parts_event_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_participants` ADD INDEX `idx_event_parts_event_time` (`event_part_event_id`, `event_part_add_time`, `id`)',
  'SELECT ''idx_event_parts_event_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_participants'
    AND INDEX_NAME = 'idx_event_parts_openid_event'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_participants` ADD INDEX `idx_event_parts_openid_event` (`event_part_mini_openid`, `event_part_event_id`)',
  'SELECT ''idx_event_parts_openid_event exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_participants'
    AND INDEX_NAME = 'idx_event_parts_event_openid'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_participants` ADD INDEX `idx_event_parts_event_openid` (`event_part_event_id`, `event_part_mini_openid`)',
  'SELECT ''idx_event_parts_event_openid exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_roles'
    AND INDEX_NAME = 'idx_event_roles_user_event'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_roles` ADD INDEX `idx_event_roles_user_event` (`event_role_user_id`, `event_role_event_id`)',
  'SELECT ''idx_event_roles_user_event exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_roles'
    AND INDEX_NAME = 'idx_event_roles_event_user'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_roles` ADD INDEX `idx_event_roles_event_user` (`event_role_event_id`, `event_role_user_id`)',
  'SELECT ''idx_event_roles_event_user exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_dynamics'
    AND INDEX_NAME = 'idx_event_dynamics_event_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_dynamics` ADD INDEX `idx_event_dynamics_event_time` (`event_dynamic_event_id`, `event_dynamic_add_time`, `id`)',
  'SELECT ''idx_event_dynamics_event_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_scores'
    AND INDEX_NAME = 'idx_event_scores_event_time'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_scores` ADD INDEX `idx_event_scores_event_time` (`event_score_event_id`, `event_score_add_time`, `id`)',
  'SELECT ''idx_event_scores_event_time exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'event_scores'
    AND INDEX_NAME = 'idx_event_scores_event_participant'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `event_scores` ADD INDEX `idx_event_scores_event_participant` (`event_score_event_id`, `event_score_participant_id`)',
  'SELECT ''idx_event_scores_event_participant exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
