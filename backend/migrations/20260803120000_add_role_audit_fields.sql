-- 为角色表补统一审计字段，使角色管理列表可以走统一数据权限过滤。
-- 历史记录无法准确还原创建者，迁移会尽量按已绑定用户回填，否则保留为系统记录。

SET @has_create_by = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'roles'
    AND COLUMN_NAME = 'create_by'
);
SET @ddl := IF(
  @has_create_by = 0,
  'ALTER TABLE `roles` ADD COLUMN `create_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人ID'' AFTER `role_data_scope`',
  'SELECT ''roles.create_by exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_update_by = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'roles'
    AND COLUMN_NAME = 'update_by'
);
SET @ddl := IF(
  @has_update_by = 0,
  'ALTER TABLE `roles` ADD COLUMN `update_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人ID'' AFTER `create_by`',
  'SELECT ''roles.update_by exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_create_dept_id = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'roles'
    AND COLUMN_NAME = 'create_dept_id'
);
SET @ddl := IF(
  @has_create_dept_id = 0,
  'ALTER TABLE `roles` ADD COLUMN `create_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''创建人部门ID'' AFTER `update_by`',
  'SELECT ''roles.create_dept_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_update_dept_id = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'roles'
    AND COLUMN_NAME = 'update_dept_id'
);
SET @ddl := IF(
  @has_update_dept_id = 0,
  'ALTER TABLE `roles` ADD COLUMN `update_dept_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''更新人部门ID'' AFTER `create_dept_id`',
  'SELECT ''roles.update_dept_id exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `roles` r
LEFT JOIN (
  SELECT
    bound_roles.role_id AS role_id,
    MIN(bound_roles.user_id) AS user_id
  FROM (
    SELECT
      u.`user_role_id` AS role_id,
      u.`id` AS user_id
    FROM `users` u
    WHERE u.`user_role_id` > 0
      AND u.`user_status` = 1
    UNION ALL
    SELECT
      ur.`user_role_role_id` AS role_id,
      ur.`user_role_user_id` AS user_id
    FROM `user_roles` ur
    JOIN `users` u ON u.`id` = ur.`user_role_user_id` AND u.`user_status` = 1
    WHERE ur.`user_role_role_id` > 0
      AND ur.`user_role_status` = 1
  ) bound_roles
  GROUP BY bound_roles.role_id
) owner ON owner.role_id = r.`id`
LEFT JOIN (
  SELECT
    ud.`user_dept_user_id` AS user_id,
    MIN(ud.`user_dept_dept_id`) AS dept_id
  FROM `user_depts` ud
  GROUP BY ud.`user_dept_user_id`
) owner_dept ON owner_dept.user_id = owner.user_id
SET
  r.`create_by` = IF(r.`create_by` = 0, COALESCE(owner.user_id, 0), r.`create_by`),
  r.`update_by` = IF(r.`update_by` = 0, COALESCE(owner.user_id, r.`create_by`), r.`update_by`),
  r.`create_dept_id` = IF(r.`create_dept_id` = 0, COALESCE(owner_dept.dept_id, 0), r.`create_dept_id`),
  r.`update_dept_id` = IF(r.`update_dept_id` = 0, COALESCE(owner_dept.dept_id, r.`create_dept_id`), r.`update_dept_id`)
WHERE r.`create_by` = 0
   OR r.`update_by` = 0
   OR r.`create_dept_id` = 0
   OR r.`update_dept_id` = 0;

SET @idx_exists = (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'roles'
    AND INDEX_NAME = 'idx_roles_unified_audit_scope'
);
SET @ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `roles` ADD INDEX `idx_roles_unified_audit_scope` (`create_dept_id`, `create_by`, `role_add_time`, `id`)',
  'SELECT ''idx_roles_unified_audit_scope exists'''
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
