-- 登记完成后表单修订权限。新权限不自动授予任何历史角色。

SET @workflow_form_revision_now = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('dingtalk_h5:button:workflow:form-revision-create', '发起完成后表单修订', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', 'workflow:form-revision-create', '', 'workflow:form-revision-create',
   81, 1, @workflow_form_revision_now, @workflow_form_revision_now, NOW(3), NOW(3)),
  ('dingtalk_h5:button:workflow:form-revision-handle', '处理完成后表单修订', 'dingtalk_h5', 'button',
   'dingtalk_h5:menu:workflow', 'workflow:form-revision-handle', '', 'workflow:form-revision-handle',
   82, 1, @workflow_form_revision_now, @workflow_form_revision_now, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:form-revision-create', '发起完成后表单修订接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/instances/:id/form-revisions', '', 'workflow:form-revision-create',
   81, 1, @workflow_form_revision_now, @workflow_form_revision_now, NOW(3), NOW(3)),
  ('dingtalk_h5:api:workflow:form-revision-handle', '处理完成后表单修订接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:workflow', '/api/v2/dingtalk/h5/workflows/form-revision-tasks/:id/complete', '', 'workflow:form-revision-handle',
   82, 1, @workflow_form_revision_now, @workflow_form_revision_now, NOW(3), NOW(3))
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
