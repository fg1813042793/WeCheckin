-- 校正钉钉 H5 用户反馈菜单元数据；已执行旧迁移的数据库也统一显示为“我的反馈”。

UPDATE `permissions`
SET
  `permission_name` = '我的反馈',
  `permission_icon` = 'chat',
  `permission_edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  `updated_at` = NOW(3)
WHERE `permission_key` = 'dingtalk_h5:menu:feedback'
  AND `permission_platform` = 'dingtalk_h5'
  AND `permission_type` = 'menu';
