CREATE TABLE IF NOT EXISTS `workflow_start_drafts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '流程发起草稿ID',
  `definition_id` bigint unsigned NOT NULL COMMENT '流程定义ID',
  `definition_version` int NOT NULL COMMENT '流程定义版本',
  `starter_id` varchar(64) NOT NULL COMMENT '草稿所属发起人ID',
  `form_data_json` mediumtext COMMENT '草稿表单数据JSON',
  `edit_time` bigint NOT NULL DEFAULT 0 COMMENT '草稿更新时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_start_draft_owner` (`definition_id`,`starter_id`),
  KEY `idx_workflow_start_drafts_starter_time` (`starter_id`,`edit_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用流程发起草稿';
