-- Register Admin and DingTalk H5 user feedback permissions and backfill existing applicable roles.
-- The API catalogs added by later tasks must use these exact route bindings:
-- GET /api/v2/admin/user-feedbacks/overview -> admin:api:user-feedback:list (user-feedback:list)
-- GET /api/v2/admin/user-feedbacks -> admin:api:user-feedback:list (user-feedback:list)
-- GET /api/v2/admin/user-feedbacks/:id -> admin:api:user-feedback:list (user-feedback:list)
-- PATCH /api/v2/admin/user-feedbacks/:id/status -> admin:api:user-feedback:handle (user-feedback:handle)
-- GET /api/v2/dingtalk/h5/user-feedbacks -> dingtalk_h5:api:feedback:list
-- POST /api/v2/dingtalk/h5/user-feedbacks -> dingtalk_h5:api:feedback:create
-- GET /api/v2/dingtalk/h5/user-feedbacks/:id -> dingtalk_h5:api:feedback:detail
-- POST /api/v2/dingtalk/h5/user-feedbacks/:id/supplements -> dingtalk_h5:api:feedback:supplement

SET @user_feedback_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_parent_key`, `permission_resource_path`, `permission_icon`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`,
  `created_at`, `updated_at`
) VALUES
  ('admin:menu:user-feedback', '用户反馈', 'admin', 'menu',
   '', '/user-feedbacks', 'ChatDotRound', 'user-feedback:list',
   20, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('admin:menu:user-feedback:list', '用户反馈查看', 'admin', 'button',
   'admin:menu:user-feedback', '', '', 'user-feedback:list',
   1, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('admin:menu:user-feedback:handle', '用户反馈处理', 'admin', 'button',
   'admin:menu:user-feedback', '', '', 'user-feedback:handle',
   2, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('admin:api-category:user-feedback', '用户反馈', 'admin', 'api_category',
   '', '', '', '',
   90, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('admin:api:user-feedback:list', '用户反馈查看接口', 'admin', 'api',
   'admin:api-category:user-feedback', '/api/v2/admin/user-feedbacks', '', 'user-feedback:list',
   10, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('admin:api:user-feedback:handle', '用户反馈处理接口', 'admin', 'api',
   'admin:api-category:user-feedback', '/api/v2/admin/user-feedbacks/:id/status', '', 'user-feedback:handle',
   20, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:feedback', '用户反馈', 'dingtalk_h5', 'menu',
   '', 'feedback', 'mine', '',
   110, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('dingtalk_h5:api-category:feedback', '用户反馈', 'dingtalk_h5', 'api_category',
   '', '', '', '',
   80, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('dingtalk_h5:api:feedback:list', '用户反馈列表接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:feedback', '/api/v2/dingtalk/h5/user-feedbacks', '', 'feedback:list',
   10, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('dingtalk_h5:api:feedback:create', '用户反馈创建接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:feedback', '/api/v2/dingtalk/h5/user-feedbacks', '', 'feedback:create',
   20, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('dingtalk_h5:api:feedback:detail', '用户反馈详情接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:feedback', '/api/v2/dingtalk/h5/user-feedbacks/:id', '', 'feedback:detail',
   30, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)),
  ('dingtalk_h5:api:feedback:supplement', '用户反馈补充接口', 'dingtalk_h5', 'api',
   'dingtalk_h5:api-category:feedback', '/api/v2/dingtalk/h5/user-feedbacks/:id/supplements', '', 'feedback:supplement',
   40, 1, @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3))
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

-- Roles already allowed into Admin keep access to the newly introduced feedback workflow.
INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '',
  'user-feedback-admin-role-backfill', 1,
  @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN `permissions` target_perm
  ON target_perm.`permission_key` IN (
    'admin:menu:user-feedback',
    'admin:menu:user-feedback:list',
    'admin:menu:user-feedback:handle',
    'admin:api:user-feedback:list',
    'admin:api:user-feedback:handle'
  )
WHERE source_grant.`grant_subject_type` = 'role'
  AND source_grant.`grant_permission_key` = 'admin:login'
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);

-- A baseline H5 menu grant adds the menu. Explicit API roles inherit the feedback APIs
-- from bootstrap permission so the migration does not switch legacy menu-only roles into strict API mode.
INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT DISTINCT
  source_grant.`grant_subject_type`, source_grant.`grant_subject_id`,
  target_perm.`permission_key`, target_perm.`id`, 'allow', '',
  'user-feedback-h5-role-backfill', 1,
  @user_feedback_now_ms, @user_feedback_now_ms, NOW(3), NOW(3)
FROM `permission_grants` source_grant
JOIN (
  SELECT 'dingtalk_h5:menu:dashboard' AS source_key, 'dingtalk_h5:menu:feedback' AS target_key
  UNION ALL SELECT 'dingtalk_h5:api:bootstrap:view', 'dingtalk_h5:api:feedback:list'
  UNION ALL SELECT 'dingtalk_h5:api:bootstrap:view', 'dingtalk_h5:api:feedback:create'
  UNION ALL SELECT 'dingtalk_h5:api:bootstrap:view', 'dingtalk_h5:api:feedback:detail'
  UNION ALL SELECT 'dingtalk_h5:api:bootstrap:view', 'dingtalk_h5:api:feedback:supplement'
) mapping ON mapping.source_key = source_grant.`grant_permission_key`
JOIN `permissions` target_perm ON target_perm.`permission_key` = mapping.target_key
WHERE source_grant.`grant_subject_type` = 'role'
  AND source_grant.`grant_permission_key` IN (
    'dingtalk_h5:menu:dashboard',
    'dingtalk_h5:api:bootstrap:view'
  )
  AND source_grant.`grant_effect` = 'allow'
  AND source_grant.`grant_status` = 1
ON DUPLICATE KEY UPDATE
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_effect` = 'allow',
  `grant_scope_value` = '',
  `grant_source` = VALUES(`grant_source`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);
