-- 支持发起人从“我的申请”中软删除已结束流程，保留后台审计和业务关联。

ALTER TABLE `workflow_process_instances`
  ADD COLUMN `starter_deleted_at` bigint NOT NULL DEFAULT 0 COMMENT '发起人从我的申请删除时间' AFTER `end_time`,
  ADD INDEX `idx_workflow_instances_starter_deleted_time` (`starter_id`, `starter_deleted_at`, `start_time`);

-- 仅注册权限，不自动扩大现有角色或用户授权。
INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('dingtalk_h5:api:workflow:delete', 'OA 流程申请删除接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/instances/:id', '', 'workflow:delete',
   50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_parent_key` = VALUES(`permission_parent_key`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_perms` = VALUES(`permission_perms`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);
