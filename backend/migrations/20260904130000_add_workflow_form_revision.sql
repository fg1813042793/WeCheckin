-- 为运行中流程表单增加乐观并发版本，并注册办理后修改权限。

SET @column_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'workflow_process_instances' AND COLUMN_NAME = 'form_revision'
);
SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE `workflow_process_instances` ADD COLUMN `form_revision` bigint NOT NULL DEFAULT 1 COMMENT ''流程表单修订版本'' AFTER `form_data_json`',
  'SELECT ''workflow_process_instances.form_revision exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 仅注册权限，不自动扩大现有角色或用户授权。
INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('dingtalk_h5:button:workflow:form-revise', '修改已办理流程表单', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', '', '', 'workflow:form-revise',
   80, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:form-revise', 'OA 流程表单修改接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/instances/:id/form-data', '', 'workflow:form-revise',
   80, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
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
