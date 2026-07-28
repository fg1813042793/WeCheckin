-- 新增统一权限定义表与统一授权表。
--
-- 将旧 menus / role_menus / role_depts 数据迁移到统一权限表。
-- 后续清理迁移会删除 role_menus / role_depts / menus。

CREATE TABLE IF NOT EXISTS `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `permission_key` varchar(160) NOT NULL COMMENT '权限编码',
  `permission_name` varchar(120) DEFAULT '' COMMENT '权限名称',
  `permission_platform` varchar(40) DEFAULT '' COMMENT '平台',
  `permission_type` varchar(40) DEFAULT '' COMMENT '权限类型',
  `permission_parent_key` varchar(160) DEFAULT '' COMMENT '父权限编码',
  `permission_resource_id` bigint unsigned DEFAULT 0 COMMENT '旧资源ID',
  `permission_resource_path` varchar(240) DEFAULT '' COMMENT '资源路径',
  `permission_icon` varchar(100) DEFAULT '' COMMENT '图标',
  `permission_perms` varchar(240) DEFAULT '' COMMENT '兼容权限标识',
  `permission_sort` int DEFAULT 0 COMMENT '排序',
  `permission_status` tinyint DEFAULT 1 COMMENT '状态',
  `permission_add_time` bigint DEFAULT 0 COMMENT '创建时间',
  `permission_edit_time` bigint DEFAULT 0 COMMENT '修改时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permissions_key` (`permission_key`),
  KEY `idx_permissions_platform` (`permission_platform`),
  KEY `idx_permissions_type` (`permission_type`),
  KEY `idx_permissions_resource_id` (`permission_resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统一权限定义表';

CREATE TABLE IF NOT EXISTS `permission_grants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '授权ID',
  `grant_subject_type` varchar(20) NOT NULL COMMENT '授权主体类型',
  `grant_subject_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '授权主体ID',
  `grant_permission_key` varchar(160) NOT NULL COMMENT '权限编码',
  `grant_permission_id` bigint unsigned DEFAULT 0 COMMENT '权限ID',
  `grant_effect` varchar(20) DEFAULT 'allow' COMMENT '授权效果',
  `grant_scope_value` text COMMENT '范围JSON',
  `grant_source` varchar(40) DEFAULT '' COMMENT '授权来源',
  `grant_status` tinyint DEFAULT 1 COMMENT '状态',
  `grant_add_time` bigint DEFAULT 0 COMMENT '创建时间',
  `grant_edit_time` bigint DEFAULT 0 COMMENT '修改时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permission_grants_subject_permission` (`grant_subject_type`, `grant_subject_id`, `grant_permission_key`),
  KEY `idx_permission_grants_permission_id` (`grant_permission_id`),
  KEY `idx_permission_grants_permission_key` (`grant_permission_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统一权限授权表';

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`, `created_at`, `updated_at`
) VALUES
  ('admin:login', '后台入口', 'admin', 'login', 0, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('data:all', '全部数据', 'data', 'data', 1, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('data:dept', '本部门数据', 'data', 'data', 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('data:self', '本人数据', 'data', 'data', 3, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('data:custom', '自定义部门数据', 'data', 'data', 4, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);

SET @legacy_menus_exists := (
  SELECT COUNT(*)
  FROM `INFORMATION_SCHEMA`.`TABLES`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'menus'
);
SET @legacy_menus_sql := IF(
  @legacy_menus_exists > 0,
  'INSERT INTO `permissions` (
    `permission_key`, `permission_name`, `permission_platform`, `permission_type`,
    `permission_parent_key`, `permission_resource_id`, `permission_resource_path`, `permission_icon`, `permission_perms`,
    `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`, `created_at`, `updated_at`
  )
  SELECT
    CONCAT(''admin:menu:'', m.`id`),
    m.`menu_name`,
    ''admin'',
    CASE WHEN m.`menu_type` = 0 THEN ''directory'' WHEN m.`menu_type` = 2 THEN ''button'' ELSE ''menu'' END,
    IF(m.`menu_parent_id` > 0, CONCAT(''admin:menu:'', m.`menu_parent_id`), ''''),
    m.`id`,
    COALESCE(m.`menu_path`, ''''),
    COALESCE(m.`menu_icon`, ''''),
    COALESCE(m.`menu_perms`, ''''),
    COALESCE(m.`menu_sort`, 0),
    COALESCE(m.`menu_status`, 1),
    COALESCE(m.`menu_add_time`, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
    COALESCE(m.`menu_edit_time`, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
    NOW(3),
    NOW(3)
  FROM `menus` m
  ON DUPLICATE KEY UPDATE
    `permission_name` = VALUES(`permission_name`),
    `permission_type` = VALUES(`permission_type`),
    `permission_parent_key` = VALUES(`permission_parent_key`),
    `permission_resource_id` = VALUES(`permission_resource_id`),
    `permission_resource_path` = VALUES(`permission_resource_path`),
    `permission_icon` = VALUES(`permission_icon`),
    `permission_perms` = VALUES(`permission_perms`),
    `permission_sort` = VALUES(`permission_sort`),
    `permission_status` = VALUES(`permission_status`),
    `permission_edit_time` = VALUES(`permission_edit_time`),
    `updated_at` = NOW(3)',
  'SELECT ''legacy menus table skipped'''
);
PREPARE legacy_menus_stmt FROM @legacy_menus_sql;
EXECUTE legacy_menus_stmt;
DEALLOCATE PREPARE legacy_menus_stmt;

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`, `permission_perms`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`, `created_at`, `updated_at`
) VALUES
  ('admin:api:home', '控制台操作接口', 'admin', 'api', 'home', 10, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:user:list', '用户查看接口', 'admin', 'api', 'user:list', 20, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:user:add', '用户创建接口', 'admin', 'api', 'user:add', 21, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:user:edit', '用户编辑接口', 'admin', 'api', 'user:edit', 22, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:user:del', '用户删除接口', 'admin', 'api', 'user:del', 23, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:online:list', '在线用户查看接口', 'admin', 'api', 'online:list', 30, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:online:force_offline', '在线用户强制下线接口', 'admin', 'api', 'online:force_offline', 31, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:list', '报名查看接口', 'admin', 'api', 'enroll:list', 40, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:add', '报名创建接口', 'admin', 'api', 'enroll:add', 41, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:edit', '报名编辑接口', 'admin', 'api', 'enroll:edit', 42, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:enroll:del', '报名删除接口', 'admin', 'api', 'enroll:del', 43, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:news:list', '通知查看接口', 'admin', 'api', 'news:list', 50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:news:add', '通知创建接口', 'admin', 'api', 'news:add', 51, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:news:edit', '通知编辑接口', 'admin', 'api', 'news:edit', 52, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:news:del', '通知删除接口', 'admin', 'api', 'news:del', 53, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:mgr:list', '管理员查看接口', 'admin', 'api', 'mgr:list', 60, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:mgr:add', '管理员创建接口', 'admin', 'api', 'mgr:add', 61, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:mgr:edit', '管理员编辑接口', 'admin', 'api', 'mgr:edit', 62, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:mgr:del', '管理员删除接口', 'admin', 'api', 'mgr:del', 63, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:setup:edit', '系统设置接口', 'admin', 'api', 'setup:edit', 70, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dict:list', '字典查看接口', 'admin', 'api', 'dict:list', 80, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dict:add', '字典创建接口', 'admin', 'api', 'dict:add', 81, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dict:edit', '字典编辑接口', 'admin', 'api', 'dict:edit', 82, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dict:del', '字典删除接口', 'admin', 'api', 'dict:del', 83, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:log:list', '日志查看接口', 'admin', 'api', 'log:list', 90, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:log:del', '日志删除接口', 'admin', 'api', 'log:del', 91, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:list', '赛事活动查看接口', 'admin', 'api', 'event:list', 100, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:add', '赛事活动创建接口', 'admin', 'api', 'event:add', 101, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:edit', '赛事活动编辑接口', 'admin', 'api', 'event:edit', 102, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:event:del', '赛事活动删除接口', 'admin', 'api', 'event:del', 103, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dept:list', '部门查看接口', 'admin', 'api', 'dept:list', 110, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dept:add', '部门创建接口', 'admin', 'api', 'dept:add', 111, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dept:edit', '部门编辑接口', 'admin', 'api', 'dept:edit', 112, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:dept:del', '部门删除接口', 'admin', 'api', 'dept:del', 113, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:role:list', '角色查看接口', 'admin', 'api', 'role:list', 120, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:role:add', '角色创建接口', 'admin', 'api', 'role:add', 121, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:role:edit', '角色编辑接口', 'admin', 'api', 'role:edit', 122, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:role:del', '角色删除接口', 'admin', 'api', 'role:del', 123, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:menu:list', '菜单查看接口', 'admin', 'api', 'menu:list', 130, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:menu:add', '菜单创建接口', 'admin', 'api', 'menu:add', 131, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:menu:edit', '菜单编辑接口', 'admin', 'api', 'menu:edit', 132, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:menu:del', '菜单删除接口', 'admin', 'api', 'menu:del', 133, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:survey:list', '问卷查看接口', 'admin', 'api', 'survey:list', 140, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:survey:add', '问卷创建接口', 'admin', 'api', 'survey:add', 141, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:survey:edit', '问卷编辑接口', 'admin', 'api', 'survey:edit', 142, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:survey:del', '问卷删除接口', 'admin', 'api', 'survey:del', 143, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:survey:status', '问卷状态接口', 'admin', 'api', 'survey:status', 144, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:survey:copy', '问卷复制接口', 'admin', 'api', 'survey:copy', 145, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:response:list', '答卷查看接口', 'admin', 'api', 'response:list', 150, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:response:del', '答卷删除接口', 'admin', 'api', 'response:del', 151, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:response:export', '答卷导出接口', 'admin', 'api', 'response:export', 152, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:question-bank:list', '题库查看接口', 'admin', 'api', 'question-bank:list', 160, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:question-bank:add', '题库创建接口', 'admin', 'api', 'question-bank:add', 161, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:question-bank:edit', '题库编辑接口', 'admin', 'api', 'question-bank:edit', 162, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:question-bank:del', '题库删除接口', 'admin', 'api', 'question-bank:del', 163, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:exam:list', '考试查看接口', 'admin', 'api', 'exam:list', 170, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:exam:add', '考试创建接口', 'admin', 'api', 'exam:add', 171, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:exam:edit', '考试编辑接口', 'admin', 'api', 'exam:edit', 172, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('admin:api:exam:del', '考试删除接口', 'admin', 'api', 'exam:del', 173, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_perms` = VALUES(`permission_perms`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);

INSERT INTO `permissions` (
  `permission_key`, `permission_name`, `permission_platform`, `permission_type`, `permission_resource_path`,
  `permission_sort`, `permission_status`, `permission_add_time`, `permission_edit_time`, `created_at`, `updated_at`
) VALUES
  ('client:menu:home', '首页', 'client', 'menu', '/pages/index/index', 10, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:news', '通知', 'client', 'menu', '/pages/news/news_index', 20, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:enroll', '打卡任务', 'client', 'menu', '/pages/enroll/enroll_index', 30, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:my_checkin', '我的打卡', 'client', 'menu', '/pages/enroll/my_user_list', 40, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:survey', '问卷中心', 'client', 'menu', '/pages/survey/index', 50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:exam', '考试列表', 'client', 'menu', '/pages/exam/index', 60, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:event', '赛事活动', 'client', 'menu', '/pages/event/event_index', 70, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:my_activity', '我的活动', 'client', 'menu', '/pages/my/my_activity', 80, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:my_competition', '我的赛事', 'client', 'menu', '/pages/my/my_competition', 90, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:event_manage', '赛事管理', 'client', 'menu', '/pages/event/my_event_manage', 100, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('client:menu:profile', '个人中心', 'client', 'menu', '/pages/my/my_index', 110, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:dashboard', '工作台', 'dingtalk_h5', 'menu', 'dashboard', 10, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:mine', '我的绩效', 'dingtalk_h5', 'menu', 'mine', 20, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:manager', '上级审批', 'dingtalk_h5', 'menu', 'manager', 30, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:hrbp', 'HRBP评价', 'dingtalk_h5', 'menu', 'hrbp', 40, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:summary', '汇总', 'dingtalk_h5', 'menu', 'summary', 50, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:org', '组织架构', 'dingtalk_h5', 'menu', 'org', 60, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:template', '模板', 'dingtalk_h5', 'menu', 'template', 70, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)),
  ('dingtalk_h5:menu:account', '账号设置', 'dingtalk_h5', 'menu', 'account', 80, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `permission_name` = VALUES(`permission_name`),
  `permission_platform` = VALUES(`permission_platform`),
  `permission_type` = VALUES(`permission_type`),
  `permission_resource_path` = VALUES(`permission_resource_path`),
  `permission_sort` = VALUES(`permission_sort`),
  `permission_status` = 1,
  `permission_edit_time` = VALUES(`permission_edit_time`),
  `updated_at` = NOW(3);

INSERT INTO `permission_grants` (
  `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
  `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
  `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
)
SELECT
  'role', r.`id`, 'admin:login', p.`id`, 'allow', '', 'legacy', 1,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)
FROM `roles` r
JOIN `permissions` p ON p.`permission_key` = 'admin:login'
WHERE COALESCE(r.`role_allow_admin_login`, 1) = 1
ON DUPLICATE KEY UPDATE
  `grant_effect` = 'allow',
  `grant_permission_id` = VALUES(`grant_permission_id`),
  `grant_status` = 1,
  `grant_edit_time` = VALUES(`grant_edit_time`),
  `updated_at` = NOW(3);

SET @legacy_role_menus_exists := (
  SELECT COUNT(*)
  FROM `INFORMATION_SCHEMA`.`TABLES`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'role_menus'
);
SET @legacy_role_menus_sql := IF(
  @legacy_role_menus_exists > 0,
  'INSERT INTO `permission_grants` (
    `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
    `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
    `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
  )
  SELECT
    ''role'', rm.`role_menu_role_id`, CONCAT(''admin:menu:'', rm.`role_menu_menu_id`), p.`id`,
    ''allow'', '''', ''legacy'', 1,
    UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, NOW(3), NOW(3)
  FROM `role_menus` rm
  JOIN `permissions` p ON p.`permission_key` = CONCAT(''admin:menu:'', rm.`role_menu_menu_id`)
  ON DUPLICATE KEY UPDATE
    `grant_effect` = ''allow'',
    `grant_permission_id` = VALUES(`grant_permission_id`),
    `grant_status` = 1,
    `grant_edit_time` = VALUES(`grant_edit_time`),
    `updated_at` = NOW(3)',
  'SELECT ''legacy role_menus table skipped'''
);
PREPARE legacy_role_menus_stmt FROM @legacy_role_menus_sql;
EXECUTE legacy_role_menus_stmt;
DEALLOCATE PREPARE legacy_role_menus_stmt;

SET @legacy_role_depts_exists := (
  SELECT COUNT(*)
  FROM `INFORMATION_SCHEMA`.`TABLES`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'role_depts'
);
SET @legacy_data_scope_sql := IF(
  @legacy_role_depts_exists > 0,
  'INSERT INTO `permission_grants` (
    `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
    `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
    `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
  )
  SELECT
    ''role'',
    r.`id`,
    CASE COALESCE(r.`role_data_scope`, 1)
      WHEN 2 THEN ''data:dept''
      WHEN 3 THEN ''data:self''
      WHEN 4 THEN ''data:custom''
      ELSE ''data:all''
    END,
    p.`id`,
    ''allow'',
    IF(
      COALESCE(r.`role_data_scope`, 1) = 4,
      COALESCE((
        SELECT JSON_OBJECT(''deptIds'', JSON_ARRAYAGG(rd.`role_dept_dept_id`))
        FROM `role_depts` rd
        WHERE rd.`role_dept_role_id` = r.`id`
      ), JSON_OBJECT(''deptIds'', JSON_ARRAY())),
      ''''
    ),
    ''legacy'',
    1,
    UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3),
    NOW(3)
  FROM `roles` r
  JOIN `permissions` p ON p.`permission_key` = CASE COALESCE(r.`role_data_scope`, 1)
      WHEN 2 THEN ''data:dept''
      WHEN 3 THEN ''data:self''
      WHEN 4 THEN ''data:custom''
      ELSE ''data:all''
    END
  ON DUPLICATE KEY UPDATE
    `grant_effect` = ''allow'',
    `grant_permission_id` = VALUES(`grant_permission_id`),
    `grant_scope_value` = VALUES(`grant_scope_value`),
    `grant_status` = 1,
    `grant_edit_time` = VALUES(`grant_edit_time`),
    `updated_at` = NOW(3)',
  'INSERT INTO `permission_grants` (
    `grant_subject_type`, `grant_subject_id`, `grant_permission_key`, `grant_permission_id`,
    `grant_effect`, `grant_scope_value`, `grant_source`, `grant_status`,
    `grant_add_time`, `grant_edit_time`, `created_at`, `updated_at`
  )
  SELECT
    ''role'',
    r.`id`,
    CASE COALESCE(r.`role_data_scope`, 1)
      WHEN 2 THEN ''data:dept''
      WHEN 3 THEN ''data:self''
      WHEN 4 THEN ''data:custom''
      ELSE ''data:all''
    END,
    p.`id`,
    ''allow'',
    IF(COALESCE(r.`role_data_scope`, 1) = 4, JSON_OBJECT(''deptIds'', JSON_ARRAY()), ''''),
    ''legacy'',
    1,
    UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    NOW(3),
    NOW(3)
  FROM `roles` r
  JOIN `permissions` p ON p.`permission_key` = CASE COALESCE(r.`role_data_scope`, 1)
      WHEN 2 THEN ''data:dept''
      WHEN 3 THEN ''data:self''
      WHEN 4 THEN ''data:custom''
      ELSE ''data:all''
    END
  ON DUPLICATE KEY UPDATE
    `grant_effect` = ''allow'',
    `grant_permission_id` = VALUES(`grant_permission_id`),
    `grant_scope_value` = VALUES(`grant_scope_value`),
    `grant_status` = 1,
    `grant_edit_time` = VALUES(`grant_edit_time`),
    `updated_at` = NOW(3)'
);
PREPARE legacy_data_scope_stmt FROM @legacy_data_scope_sql;
EXECUTE legacy_data_scope_stmt;
DEALLOCATE PREPARE legacy_data_scope_stmt;
