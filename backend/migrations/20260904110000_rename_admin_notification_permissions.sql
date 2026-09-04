-- 站内信与钉钉通知共用通知管理父级，名称调整单独迁移，保持已执行迁移不可变。

UPDATE `permissions`
SET `permission_name` = '通知管理',
    `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    `updated_at` = NOW(3)
WHERE `permission_key` = 'admin:menu:notification';

UPDATE `permissions`
SET `permission_name` = '通知',
    `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    `updated_at` = NOW(3)
WHERE `permission_key` = 'admin:api-category:notification';
