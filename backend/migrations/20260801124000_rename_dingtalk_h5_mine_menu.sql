-- 将钉钉 H5「本月绩效」入口改名为「我的绩效」。
--
-- 权限编码和路径保持不变，避免影响已有授权和前端路由。

UPDATE `permissions`
SET
  `permission_name` = '我的绩效',
  `permission_edit_time` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  `updated_at` = NOW(3)
WHERE `permission_key` = 'dingtalk_h5:menu:performance:mine'
  AND `permission_platform` = 'dingtalk_h5';
