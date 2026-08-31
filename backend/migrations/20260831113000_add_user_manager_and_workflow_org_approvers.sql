ALTER TABLE `users`
  ADD COLUMN `manager_user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '直属上级用户ID' AFTER `user_position_id`,
  ADD KEY `idx_users_manager_user_id` (`manager_user_id`);

CREATE TABLE IF NOT EXISTS `workflow_org_approver_identities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '组织审批身份ID',
  `identity_code` varchar(80) NOT NULL COMMENT '身份编码',
  `identity_name` varchar(100) NOT NULL DEFAULT '' COMMENT '身份名称',
  `identity_sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `identity_status` tinyint NOT NULL DEFAULT 1 COMMENT '状态:1启用 0停用',
  `identity_add_time` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `identity_edit_time` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_org_approver_identity_code` (`identity_code`),
  KEY `idx_workflow_org_approver_identities_status_sort` (`identity_status`,`identity_sort`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='流程组织审批身份';

CREATE TABLE IF NOT EXISTS `workflow_org_approver_assignments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '组织审批身份人员ID',
  `department_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '部门ID',
  `identity_code` varchar(80) NOT NULL COMMENT '身份编码',
  `user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
  `assignment_sort` int NOT NULL DEFAULT 0 COMMENT '审批顺序',
  `assignment_status` tinyint NOT NULL DEFAULT 1 COMMENT '状态:1启用 0停用',
  `assignment_add_time` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `assignment_edit_time` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_org_approver_assignment` (`department_id`,`identity_code`,`user_id`),
  KEY `idx_workflow_org_approver_assignment_lookup` (`department_id`,`identity_code`,`assignment_status`,`assignment_sort`,`id`),
  KEY `idx_workflow_org_approver_assignment_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='流程组织审批身份人员配置';

INSERT INTO `workflow_org_approver_identities` (
  `identity_code`, `identity_name`, `identity_sort`, `identity_status`,
  `identity_add_time`, `identity_edit_time`, `created_at`, `updated_at`
) VALUES
  ('department_leader', '部门负责人', 10, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('group_leader', '组长', 20, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('hrbp', 'HRBP', 30, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('finance_contact', '财务对接人', 40, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `identity_name` = VALUES(`identity_name`),
  `identity_sort` = VALUES(`identity_sort`),
  `identity_status` = 1,
  `identity_edit_time` = VALUES(`identity_edit_time`),
  `updated_at` = NOW(3);

-- 为已完成权限初始化的数据库补齐组织审批身份设置菜单与接口权限。
-- 本迁移只注册权限定义，不自动给普通角色授权。
INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:workflow:org-approvers', '组织审批身份设置', 'admin', 'menu',
   'admin:menu:workflow', '/workflow/org-approvers', '', 'workflow:org-approver:list',
   4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:workflow:org-approver:list', '组织审批身份查看', 'admin', 'button',
   'admin:menu:workflow:org-approvers', '', '', 'workflow:org-approver:list',
   1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:menu:workflow:org-approver:edit', '组织审批身份维护', 'admin', 'button',
   'admin:menu:workflow:org-approvers', '', '', 'workflow:org-approver:edit',
   2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:org-approver:list', '组织审批身份查看接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-org-approver-identities', '', 'workflow:org-approver:list',
   400, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:workflow:org-approver:edit', '组织审批身份维护接口', 'admin', 'api',
   'admin:api-category:workflow', '/api/v2/admin/workflow-org-approver-assignments', '', 'workflow:org-approver:edit',
   410, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
