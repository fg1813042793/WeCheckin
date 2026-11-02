-- 钉钉 H5 账号密码登录图片验证码默认关闭；不影响钉钉免登、SSO 和自助绑定。
INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`)
VALUES (
  'DINGTALK_H5_LOGIN_CAPTCHA_ENABLED',
  '0',
  'switch',
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
)
ON DUPLICATE KEY UPDATE `setup_key` = VALUES(`setup_key`);
