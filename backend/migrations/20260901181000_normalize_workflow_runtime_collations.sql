-- 统一工作流参与人、通知表与流程实例表的排序规则，避免字符串关联时报 1267。
-- 前置条件：20260901110000_add_workflow_notifications.sql 已执行。
-- CONVERT 会重建字符列及相关索引，生产环境应在维护窗口执行。
-- 回滚：如确需恢复，先记录原排序规则，再对两张表执行同结构的 CONVERT。

ALTER TABLE `workflow_instance_participants` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

ALTER TABLE `workflow_notification_outbox` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
