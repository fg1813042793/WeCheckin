-- 钉钉 H5 首次登录自助绑定开关，默认开启；后台「钉钉应用管理 / 配置选项」可关闭。
INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`)
VALUES (
  'DINGTALK_H5_SELF_BIND_ENABLED',
  '1',
  'switch',
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
)
ON DUPLICATE KEY UPDATE `setup_key` = VALUES(`setup_key`);
