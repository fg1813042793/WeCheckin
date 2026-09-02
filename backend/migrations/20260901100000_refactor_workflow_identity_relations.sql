-- 将直属上级从用户字段迁移为可扩展、可保留历史的汇报关系；
-- 将组织审批身份配置扩展为部门默认或指定人员两种适用对象。
CREATE TABLE IF NOT EXISTS `user_reporting_relations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '汇报关系ID',
  `employee_user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '员工用户ID',
  `manager_user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '上级用户ID',
  `relation_type` varchar(40) NOT NULL DEFAULT 'direct' COMMENT '关系类型:direct直属 dotted虚线',
  `is_primary` tinyint NOT NULL DEFAULT 1 COMMENT '是否主关系',
  `relation_sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `relation_status` tinyint NOT NULL DEFAULT 1 COMMENT '状态:1启用 0停用',
  `effective_from` bigint NOT NULL DEFAULT 0 COMMENT '生效时间',
  `effective_to` bigint NOT NULL DEFAULT 0 COMMENT '失效时间,0长期有效',
  `relation_add_time` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `relation_edit_time` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_reporting_relation_history` (`employee_user_id`,`relation_type`,`manager_user_id`,`effective_from`),
  KEY `idx_user_reporting_relation_current` (`employee_user_id`,`relation_type`,`relation_status`,`is_primary`,`relation_sort`,`id`),
  KEY `idx_user_reporting_relation_manager` (`manager_user_id`,`relation_status`),
  KEY `idx_user_reporting_relation_effective` (`effective_from`,`effective_to`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户汇报关系';

INSERT INTO `user_reporting_relations` (
  `employee_user_id`, `manager_user_id`, `relation_type`, `is_primary`, `relation_sort`,
  `relation_status`, `effective_from`, `effective_to`, `relation_add_time`, `relation_edit_time`,
  `created_at`, `updated_at`
)
SELECT `id`, `manager_user_id`, 'direct', 1, 0, 1, 0, 0,
       UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
       UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
       NOW(3), NOW(3)
FROM `users`
WHERE `manager_user_id` > 0
ON DUPLICATE KEY UPDATE
  `relation_status` = 1,
  `effective_to` = 0,
  `relation_edit_time` = VALUES(`relation_edit_time`),
  `updated_at` = NOW(3);

ALTER TABLE `users`
  DROP INDEX `idx_users_manager_user_id`,
  DROP COLUMN `manager_user_id`;

ALTER TABLE `workflow_org_approver_assignments`
  ADD COLUMN `subject_type` varchar(20) NOT NULL DEFAULT 'department' COMMENT '适用对象类型:department部门 user人员' AFTER `id`,
  ADD COLUMN `subject_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '适用对象ID' AFTER `subject_type`;

UPDATE `workflow_org_approver_assignments`
SET `subject_type` = 'department',
    `subject_id` = `department_id`
WHERE `subject_id` = 0;

ALTER TABLE `workflow_org_approver_assignments`
  DROP INDEX `uk_workflow_org_approver_assignment`,
  ADD UNIQUE KEY `uk_workflow_org_approver_assignment_subject` (`subject_type`,`subject_id`,`identity_code`,`user_id`),
  ADD KEY `idx_workflow_org_approver_assignment_subject_lookup` (`subject_type`,`subject_id`,`identity_code`,`assignment_status`,`assignment_sort`,`id`);
