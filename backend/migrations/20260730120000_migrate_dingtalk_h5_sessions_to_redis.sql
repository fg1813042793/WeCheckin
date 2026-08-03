-- 钉钉 H5 登录态迁移到 Redis。
-- 旧 `dingtalk_h5_perf_sessions` 表不再作为运行时登录态来源；迁移后旧 H5 登录会话需要重新登录。

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`)
SELECT 'DINGTALK_H5_SINGLE_LOGIN', '0', 'switch', CAST(UNIX_TIMESTAMP(NOW(3)) * 1000 AS UNSIGNED), 0
WHERE NOT EXISTS (SELECT 1 FROM `setups` WHERE `setup_key` = 'DINGTALK_H5_SINGLE_LOGIN');

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`)
SELECT 'TOKEN_DINGTALK_H5_EXPIRE', '168h', 'string', CAST(UNIX_TIMESTAMP(NOW(3)) * 1000 AS UNSIGNED), 0
WHERE NOT EXISTS (SELECT 1 FROM `setups` WHERE `setup_key` = 'TOKEN_DINGTALK_H5_EXPIRE');

INSERT INTO `setups` (`setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`)
SELECT 'TOKEN_DINGTALK_H5_REDIS_PREFIX', 'dingtalk_h5_token:', 'string', CAST(UNIX_TIMESTAMP(NOW(3)) * 1000 AS UNSIGNED), 0
WHERE NOT EXISTS (SELECT 1 FROM `setups` WHERE `setup_key` = 'TOKEN_DINGTALK_H5_REDIS_PREFIX');

DROP TABLE IF EXISTS `dingtalk_h5_perf_sessions`;
