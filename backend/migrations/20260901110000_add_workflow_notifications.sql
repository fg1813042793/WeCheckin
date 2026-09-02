-- 为通用工作流增加抄送参与人和事务性通知 Outbox。
-- 本迁移只注册通知管理权限，不自动给任何角色授权。

CREATE TABLE IF NOT EXISTS `workflow_instance_participants` (
  `id` VARCHAR(64) NOT NULL COMMENT '参与人记录ID',
  `instance_id` VARCHAR(64) NOT NULL COMMENT '流程实例ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '本地用户ID',
  `participant_role` VARCHAR(24) NOT NULL COMMENT '参与角色:cc',
  `node_id` VARCHAR(100) NOT NULL COMMENT '来源节点ID',
  `add_time` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间毫秒时间戳',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_participant_source` (`instance_id`,`user_id`,`participant_role`,`node_id`),
  KEY `idx_workflow_participant_user_role` (`user_id`,`participant_role`,`instance_id`),
  KEY `idx_workflow_participant_instance` (`instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通用工作流实例参与人';

CREATE TABLE IF NOT EXISTS `workflow_notification_outbox` (
  `id` VARCHAR(64) NOT NULL COMMENT '通知Outbox ID',
  `instance_id` VARCHAR(64) NOT NULL COMMENT '流程实例ID',
  `node_id` VARCHAR(100) NOT NULL COMMENT '来源节点ID',
  `task_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源任务ID',
  `recipient_user_id` VARCHAR(64) NOT NULL COMMENT '本地接收人ID',
  `notification_kind` VARCHAR(32) NOT NULL COMMENT '通知类型',
  `notification_channel` VARCHAR(32) NOT NULL COMMENT '通知渠道',
  `notification_status` VARCHAR(24) NOT NULL DEFAULT 'pending' COMMENT '投递状态',
  `dedupe_key` VARCHAR(255) NOT NULL COMMENT '通知幂等键',
  `payload_json` MEDIUMTEXT NOT NULL COMMENT '已渲染通知负载JSON',
  `corp_id` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '固定的钉钉企业ID',
  `provider_message_id` VARCHAR(160) NOT NULL DEFAULT '' COMMENT '渠道消息标识',
  `attempts` INT NOT NULL DEFAULT 0 COMMENT '投递尝试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试毫秒时间戳',
  `last_error` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近失败摘要',
  `sent_at` BIGINT NOT NULL DEFAULT 0 COMMENT '发送成功毫秒时间戳',
  `add_time` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间毫秒时间戳',
  `edit_time` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间毫秒时间戳',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_notification_dedupe` (`dedupe_key`),
  KEY `idx_workflow_notification_due` (`notification_status`,`next_retry_at`),
  KEY `idx_workflow_notification_instance` (`instance_id`,`add_time`),
  KEY `idx_workflow_notification_recipient` (`recipient_user_id`,`notification_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通用工作流通知Outbox';

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:workflow:notification:list', '流程通知查看', 'admin', 'button',
   'admin:menu:workflow:instances', '', '', 'workflow:notification:list',
   5, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:workflow:notification:retry', '流程通知重试', 'admin', 'button',
   'admin:menu:workflow:instances', '', '', 'workflow:notification:retry',
   6, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:notification:list', '流程通知查看接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-notifications', '', 'workflow:notification:list',
   410, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:notification:retry', '流程通知单条重试接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-notifications/:id/retry', '', 'workflow:notification:retry',
   420, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:notification:dispatch', '流程通知到期投递接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-notifications/dispatch-due', '', 'workflow:notification:retry',
   430, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_icon` = VALUES(`permission_icon`),
  `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);
