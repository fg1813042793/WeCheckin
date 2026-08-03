-- 为钉钉 H5 菜单权限补默认图标。
--
-- 仅回填空图标，保留后台权限管理中已经手动配置过的图标。

UPDATE `permissions`
SET
  `permission_icon` = CASE `permission_key`
    WHEN 'dingtalk_h5:menu:dashboard' THEN 'dashboard'
    WHEN 'dingtalk_h5:menu:performance' THEN 'performance'
    WHEN 'dingtalk_h5:menu:performance:mine' THEN 'mine'
    WHEN 'dingtalk_h5:menu:performance:history' THEN 'history'
    WHEN 'dingtalk_h5:menu:performance:manager' THEN 'manager'
    WHEN 'dingtalk_h5:menu:performance:hrbp' THEN 'hrbp'
    WHEN 'dingtalk_h5:menu:performance:summary' THEN 'summary'
    WHEN 'dingtalk_h5:menu:performance:org' THEN 'org'
    WHEN 'dingtalk_h5:menu:performance:template' THEN 'template'
    ELSE `permission_icon`
  END,
  `permission_edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  `updated_at` = NOW(3)
WHERE `permission_platform` = 'dingtalk_h5'
  AND `permission_type` IN ('directory', 'menu')
  AND `permission_key` IN (
    'dingtalk_h5:menu:dashboard',
    'dingtalk_h5:menu:performance',
    'dingtalk_h5:menu:performance:mine',
    'dingtalk_h5:menu:performance:history',
    'dingtalk_h5:menu:performance:manager',
    'dingtalk_h5:menu:performance:hrbp',
    'dingtalk_h5:menu:performance:summary',
    'dingtalk_h5:menu:performance:org',
    'dingtalk_h5:menu:performance:template'
  )
  AND (`permission_icon` IS NULL OR TRIM(`permission_icon`) = '');
