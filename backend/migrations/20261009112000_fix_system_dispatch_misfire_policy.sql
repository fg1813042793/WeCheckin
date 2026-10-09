-- System outbox jobs must execute the current due occurrence. "skip" would
-- advance next_run_at without creating a run on every scheduler poll.

SET @system_dispatch_policy_now = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS SIGNED);

UPDATE `scheduled_tasks`
SET
  `misfire_policy` = 'fire_once',
  `next_run_at` = CASE
    WHEN `next_run_at` > @system_dispatch_policy_now THEN `next_run_at`
    ELSE @system_dispatch_policy_now
  END,
  `version` = `version` + 1,
  `edit_time` = @system_dispatch_policy_now
WHERE `code` IN (
  'system.notification-outbox-dispatch',
  'system.workflow-business-event-dispatch'
)
  AND `deleted_at` = 0;
