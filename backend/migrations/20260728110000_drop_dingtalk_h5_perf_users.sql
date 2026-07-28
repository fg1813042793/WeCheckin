-- 清理旧的钉钉 H5 绩效人员独立表。
--
-- 当前钉钉 H5 人员已复用 `users` 表：
-- - account -> users.user_mini_openid
-- - name -> users.user_name
-- - password_hash -> users.user_password
-- - role/position/department/manager/hrbp/responsibleDepartments -> users.user_obj.dingtalkH5Performance
--
-- 本迁移可重复执行：旧表存在时先合并数据再删除；旧表不存在时跳过。

SET @dt_h5_perf_user_table_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_users'
);

SET @dt_h5_perf_user_position_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'dingtalk_h5_perf_users'
    AND COLUMN_NAME = 'position'
);

SET @dt_h5_add_position_sql := IF(
  @dt_h5_perf_user_table_exists > 0 AND @dt_h5_perf_user_position_exists = 0,
  'ALTER TABLE `dingtalk_h5_perf_users` ADD COLUMN `position` varchar(100) DEFAULT '''' COMMENT ''岗位'' AFTER `role`',
  'SELECT ''dingtalk_h5_perf_users.position exists or table skipped'''
);

PREPARE stmt FROM @dt_h5_add_position_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @dt_h5_cleanup_sql := IF(
  @dt_h5_perf_user_table_exists > 0,
  'INSERT INTO `users` (
    `user_mini_openid`,
    `user_status`,
    `user_name`,
    `user_mobile`,
    `user_pic`,
    `user_forms`,
    `user_obj`,
    `user_password`,
    `user_login_cnt`,
    `user_login_time`,
    `user_add_time`,
    `user_add_ip`,
    `user_edit_time`,
    `user_edit_ip`,
    `created_at`,
    `updated_at`
  )
  SELECT
    p.`account`,
    COALESCE(p.`status`, 1),
    COALESCE(NULLIF(p.`name`, ''''), p.`account`),
    '''',
    ''/static/default-avatar.png'',
    ''[]'',
    JSON_OBJECT(
      ''dingtalkH5Performance'',
      JSON_OBJECT(
        ''role'', COALESCE(NULLIF(p.`role`, ''''), ''employee''),
        ''position'', COALESCE(p.`position`, ''''),
        ''department'', COALESCE(p.`department`, ''''),
        ''departmentLevel1'', COALESCE(p.`department_level1`, ''''),
        ''departmentLevel2'', COALESCE(p.`department_level2`, ''''),
        ''departmentLevel3'', COALESCE(p.`department_level3`, ''''),
        ''managerId'', COALESCE(p.`manager_account`, ''''),
        ''hrbpId'', COALESCE(p.`hrbp_account`, ''''),
        ''responsibleDepartments'', IF(
          JSON_VALID(COALESCE(p.`responsible_departments`, '''')),
          JSON_EXTRACT(p.`responsible_departments`, ''$''),
          JSON_ARRAY()
        )
      )
    ),
    COALESCE(p.`password_hash`, ''''),
    0,
    0,
    COALESCE(p.`add_time`, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)),
    ''127.0.0.1'',
    COALESCE(p.`edit_time`, CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)),
    ''127.0.0.1'',
    CASE WHEN p.`created_at` IS NULL OR CAST(p.`created_at` AS CHAR) IN (''0000-00-00'', ''0000-00-00 00:00:00'', ''0000-00-00 00:00:00.000'') THEN NOW(3) ELSE p.`created_at` END,
    CASE WHEN p.`updated_at` IS NULL OR CAST(p.`updated_at` AS CHAR) IN (''0000-00-00'', ''0000-00-00 00:00:00'', ''0000-00-00 00:00:00.000'') THEN NOW(3) ELSE p.`updated_at` END
  FROM `dingtalk_h5_perf_users` p
  WHERE p.`account` <> ''''
  ON DUPLICATE KEY UPDATE
    `user_status` = VALUES(`user_status`),
    `user_name` = VALUES(`user_name`),
    `user_password` = IF(COALESCE(`users`.`user_password`, '''') = '''', VALUES(`user_password`), `users`.`user_password`),
    `user_pic` = IF(COALESCE(`users`.`user_pic`, '''') = '''', VALUES(`user_pic`), `users`.`user_pic`),
    `user_forms` = IF(COALESCE(`users`.`user_forms`, '''') = '''', VALUES(`user_forms`), `users`.`user_forms`),
    `user_obj` = JSON_SET(
      IF(JSON_VALID(COALESCE(`users`.`user_obj`, '''')), `users`.`user_obj`, JSON_OBJECT()),
      ''$.dingtalkH5Performance'',
      JSON_EXTRACT(VALUES(`user_obj`), ''$.dingtalkH5Performance'')
    ),
    `user_edit_time` = VALUES(`user_edit_time`),
    `updated_at` = NOW()',
  'SELECT ''dingtalk_h5_perf_users table does not exist, skip data migration'''
);

PREPARE stmt FROM @dt_h5_cleanup_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TABLE IF EXISTS `dingtalk_h5_perf_users`;
