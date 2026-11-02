-- Admin 图片验证码默认关闭。保存配置后由 Setup 服务立即刷新缓存。
INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`)
SELECT 'ADMIN_LOGIN_CAPTCHA_ENABLED', '0', 'switch', CAST(UNIX_TIMESTAMP(NOW(3)) * 1000 AS UNSIGNED), 0
WHERE NOT EXISTS (
  SELECT 1 FROM `setups` WHERE `setup_key` = 'ADMIN_LOGIN_CAPTCHA_ENABLED'
);
