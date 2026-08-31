CREATE TABLE IF NOT EXISTS `workflow_process_instances` (
  `id` varchar(64) NOT NULL COMMENT '流程实例ID',
  `definition_id` bigint unsigned NOT NULL COMMENT '流程定义ID',
  `definition_version` int NOT NULL COMMENT '流程定义版本',
  `definition_key` varchar(100) NOT NULL COMMENT '流程定义编码',
  `business_type` varchar(100) NOT NULL DEFAULT '' COMMENT '业务类型',
  `business_key` varchar(160) NOT NULL DEFAULT '' COMMENT '业务唯一标识',
  `starter_id` varchar(64) NOT NULL DEFAULT '' COMMENT '发起人ID',
  `instance_status` varchar(24) NOT NULL DEFAULT 'running' COMMENT '实例状态',
  `start_time` bigint NOT NULL DEFAULT 0 COMMENT '开始时间',
  `end_time` bigint NOT NULL DEFAULT 0 COMMENT '结束时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_workflow_instance_business` (`definition_id`,`business_type`,`business_key`),
  KEY `idx_workflow_instances_definition_status` (`definition_id`,`instance_status`),
  KEY `idx_workflow_instances_starter_status` (`starter_id`,`instance_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程实例';

CREATE TABLE IF NOT EXISTS `workflow_process_tokens` (
  `id` varchar(64) NOT NULL COMMENT '流程令牌ID',
  `instance_id` varchar(64) NOT NULL COMMENT '流程实例ID',
  `node_id` varchar(100) NOT NULL COMMENT '当前节点ID',
  `token_status` varchar(24) NOT NULL DEFAULT 'active' COMMENT '令牌状态',
  `branch_group` varchar(64) NOT NULL DEFAULT '' COMMENT '并行分支组',
  `branch_total` int NOT NULL DEFAULT 0 COMMENT '并行分支总数',
  `arrived_at` bigint NOT NULL DEFAULT 0 COMMENT '到达时间',
  `completed_at` bigint NOT NULL DEFAULT 0 COMMENT '完成时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_tokens_instance_status` (`instance_id`,`token_status`),
  KEY `idx_workflow_tokens_branch_node_status` (`branch_group`,`node_id`,`token_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程执行令牌';

CREATE TABLE IF NOT EXISTS `workflow_process_tasks` (
  `id` varchar(64) NOT NULL COMMENT '流程任务ID',
  `instance_id` varchar(64) NOT NULL COMMENT '流程实例ID',
  `token_id` varchar(64) NOT NULL COMMENT '流程令牌ID',
  `node_id` varchar(100) NOT NULL COMMENT '节点ID',
  `node_name` varchar(200) NOT NULL DEFAULT '' COMMENT '节点名称',
  `task_group_key` varchar(64) NOT NULL DEFAULT '' COMMENT '多人审批任务组',
  `task_assignee_id` varchar(64) NOT NULL DEFAULT '' COMMENT '处理人ID',
  `approval_mode` varchar(24) NOT NULL DEFAULT 'single' COMMENT '审批模式',
  `completion_rate` int NOT NULL DEFAULT 100 COMMENT '会签通过比例百分数',
  `task_sequence` int NOT NULL DEFAULT 1 COMMENT '顺序审批序号',
  `task_total` int NOT NULL DEFAULT 1 COMMENT '任务组任务数',
  `task_status` varchar(24) NOT NULL DEFAULT 'pending' COMMENT '任务状态',
  `task_action` varchar(24) NOT NULL DEFAULT '' COMMENT '处理动作',
  `task_comment` varchar(1000) NOT NULL DEFAULT '' COMMENT '处理意见',
  `handled_by` varchar(64) NOT NULL DEFAULT '' COMMENT '实际处理人ID',
  `handled_at` bigint NOT NULL DEFAULT 0 COMMENT '处理时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_tasks_assignee_status` (`task_assignee_id`,`task_status`),
  KEY `idx_workflow_tasks_instance_status` (`instance_id`,`task_status`),
  KEY `idx_workflow_tasks_group_status` (`task_group_key`,`task_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程待办任务';

CREATE TABLE IF NOT EXISTS `workflow_process_variables` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '流程变量ID',
  `instance_id` varchar(64) NOT NULL COMMENT '流程实例ID',
  `variable_key` varchar(100) NOT NULL COMMENT '变量名',
  `variable_value_json` mediumtext COMMENT '变量JSON值',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_workflow_variable_instance_key` (`instance_id`,`variable_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程变量';

CREATE TABLE IF NOT EXISTS `workflow_process_history` (
  `id` varchar(64) NOT NULL COMMENT '流程历史ID',
  `instance_id` varchar(64) NOT NULL COMMENT '流程实例ID',
  `event_type` varchar(40) NOT NULL COMMENT '事件类型',
  `node_id` varchar(100) NOT NULL DEFAULT '' COMMENT '节点ID',
  `task_id` varchar(64) NOT NULL DEFAULT '' COMMENT '任务ID',
  `actor_id` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人ID',
  `event_message` varchar(1000) NOT NULL DEFAULT '' COMMENT '事件说明',
  `event_time` bigint NOT NULL DEFAULT 0 COMMENT '事件时间',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_history_instance_time` (`instance_id`,`event_time`),
  KEY `idx_workflow_history_task` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程历史事件';
