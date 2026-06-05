-- ========================================
-- WeCheckin 测试数据
-- 数据库: wecheckin
-- ========================================

-- 清空现有数据
TRUNCATE TABLE `logs`;
TRUNCATE TABLE `favorites`;
TRUNCATE TABLE `enroll_joins`;
TRUNCATE TABLE `enroll_users`;
TRUNCATE TABLE `enrolls`;
TRUNCATE TABLE `news`;
TRUNCATE TABLE `setups`;
TRUNCATE TABLE `admins`;
TRUNCATE TABLE `users`;
-- ========================================
-- 1. 管理员 (admins)
-- ========================================
INSERT INTO `admins` (`ADMIN_NAME`, `ADMIN_PASSWORD`, `ADMIN_DESC`, `ADMIN_PHONE`, `ADMIN_STATUS`, `ADMIN_TYPE`, `ADMIN_TOKEN`, `ADMIN_LOGIN_CNT`, `ADMIN_LOGIN_TIME`, `ADMIN_ADD_TIME`, `ADMIN_EDIT_TIME`, `ADMIN_ADD_IP`, `ADMIN_EDIT_IP`, `created_at`, `updated_at`) VALUES
('admin', '0192023a7bbd73250516f069df18b500', '超级管理员', '13800000000', 1, 1, '', 0, 0, 1769472000000, 1769472000000, '127.0.0.1', '127.0.0.1', NOW(), NOW());

INSERT INTO `admins` (`ADMIN_NAME`, `ADMIN_PASSWORD`, `ADMIN_DESC`, `ADMIN_PHONE`, `ADMIN_STATUS`, `ADMIN_TYPE`, `ADMIN_TOKEN`, `ADMIN_LOGIN_CNT`, `ADMIN_LOGIN_TIME`, `ADMIN_ADD_TIME`, `ADMIN_EDIT_TIME`, `ADMIN_ADD_IP`, `ADMIN_EDIT_IP`, `created_at`, `updated_at`) VALUES
('editor', '0192023a7bbd73250516f069df18b500', '编辑员', '13900000000', 1, 0, '', 0, 0, 1769472000000, 1769472000000, '127.0.0.1', '127.0.0.1', NOW(), NOW());

-- ========================================
-- 2. 用户 (users)
-- ========================================
INSERT INTO `users` (`USER_MINI_OPENID`, `USER_STATUS`, `USER_NAME`, `USER_MOBILE`, `USER_PIC`, `USER_FORMS`, `USER_OBJ`, `USER_LOGIN_CNT`, `USER_LOGIN_TIME`, `USER_ADD_TIME`, `USER_ADD_IP`, `USER_EDIT_TIME`, `USER_EDIT_IP`, `created_at`, `updated_at`) VALUES
('o1_test_openid_001', 1, '张三', '13800138001', '/static/default-avatar.png', '[]', '{}', 5, 1769500000000, 1769417600000, '192.168.1.1', 1769500000000, '', NOW(), NOW()),
('o1_test_openid_002', 1, '李四', '13800138002', '/static/default-avatar.png', '[]', '{}', 3, 1769501000000, 1769427600000, '192.168.1.2', 1769501000000, '', NOW(), NOW()),
('o1_test_openid_003', 1, '王五', '13800138003', '/static/default-avatar.png', '[]', '{}', 8, 1769502000000, 1769437600000, '192.168.1.3', 1769502000000, '', NOW(), NOW()),
('o1_test_openid_004', 1, '赵六', '13800138004', '/static/default-avatar.png', '[]', '{}', 2, 1769503000000, 1769447600000, '192.168.1.4', 1769503000000, '', NOW(), NOW()),
('o1_test_openid_005', 1, '刘七', '13800138005', '/static/default-avatar.png', '[]', '{}', 0, 0, 1769457600000, '192.168.1.5', 0, '', NOW(), NOW());

-- ========================================
-- 3. 通知 (news)
-- ========================================
INSERT INTO `news` (`NEWS_TITLE`, `NEWS_DESC`, `NEWS_STATUS`, `NEWS_CATE_ID`, `NEWS_CATE_NAME`, `NEWS_ORDER`, `NEWS_VOUCH`, `NEWS_CONTENT`, `NEWS_QR`, `NEWS_VIEW_CNT`, `NEWS_PIC`, `NEWS_FORMS`, `NEWS_OBJ`, `NEWS_ADD_TIME`, `NEWS_EDIT_TIME`, `NEWS_ADD_IP`, `NEWS_EDIT_IP`, `created_at`, `updated_at`) VALUES
('打卡活动规则更新', '2026年6月起打卡规则有调整，请仔细阅读', 1, 'notice', '公告', 1, 1, '[{"type":"text","val":"打卡活动规则更新公告"},{"type":"text","val":"自2026年6月1日起，每日打卡时间段调整为06:00~23:00，请各位用户注意调整打卡习惯。"},{"type":"text","val":"连续打卡30天可获得额外积分奖励！"}]', '', 156, '["https://picsum.photos/seed/news1/800/400"]', '[]', '{}', 1769417600000, 1769417600000, '127.0.0.1', '', NOW(), NOW()),
('五月打卡排行榜', '五月打卡活动结束，恭喜获奖用户', 1, 'notice', '公告', 2, 1, '[{"type":"text","val":"五月打卡排行榜"},{"type":"text","val":"经过一个月的激烈角逐，五月打卡活动圆满结束！"},{"type":"text","val":"一等奖：张三 - 连续打卡30天"},{"type":"text","val":"二等奖：李四 - 连续打卡28天"},{"type":"text","val":"三等奖：王五 - 连续打卡25天"}]', '', 89, '["https://picsum.photos/seed/news2/800/400"]', '[]', '{}', 1769331200000, 1769331200000, '127.0.0.1', '', NOW(), NOW()),
('端午节特别活动', '端午节期间打卡双倍积分', 1, 'notice', '公告', 3, 1, '[{"type":"text","val":"端午节特别活动"},{"type":"text","val":"端午节期间（5月30日~6月1日）打卡可获得双倍积分！"},{"type":"text","val":"祝大家端午安康！"}]', '', 210, '["https://picsum.photos/seed/news3/800/400"]', '[]', '{}', 1769244800000, 1769244800000, '127.0.0.1', '', NOW(), NOW());

-- ========================================
-- 4. 打卡项目 (enrolls)
-- ========================================
INSERT INTO `enrolls` (`ENROLL_TITLE`, `ENROLL_STATUS`, `ENROLL_CATE_ID`, `ENROLL_CATE_NAME`, `ENROLL_START`, `ENROLL_END`, `ENROLL_DAY_CNT`, `ENROLL_ORDER`, `ENROLL_VOUCH`, `ENROLL_FORMS`, `ENROLL_OBJ`, `ENROLL_JOIN_FORMS`, `ENROLL_QR`, `ENROLL_VIEW_CNT`, `ENROLL_JOIN_CNT`, `ENROLL_USER_CNT`, `ENROLL_USER_LIST`, `ENROLL_ADD_TIME`, `ENROLL_EDIT_TIME`, `ENROLL_ADD_IP`, `ENROLL_EDIT_IP`, `created_at`, `updated_at`) VALUES
('每日晨跑打卡', 1, 'sport', '运动', 1746000000000, 1770883200000, 365, 1, 1, '[]', '{"cover":["https://picsum.photos/seed/enroll1/800/400"],"desc":"坚持每天晨跑，保持健康体魄！","content":[{"type":"text","val":"每天早上6:00~8:00进行晨跑打卡"},{"type":"img","val":"https://picsum.photos/seed/run1/800/400"},{"type":"text","val":"完成后请上传跑步截图作为凭证"}]}', '[{"label":"跑步距离(公里)","type":"text","required":true},{"label":"跑步截图","type":"image","required":false}]', '', 1024, 48, 4, '[{"pic":"/static/default-avatar.png","name":"张三"},{"pic":"/static/default-avatar.png","name":"李四"},{"pic":"/static/default-avatar.png","name":"王五"}]', 1768896000000, 1768896000000, '127.0.0.1', '', NOW(), NOW()),

('每日英语学习', 1, 'study', '学习', 1746000000000, 1770883200000, 365, 2, 1, '[]', '{"cover":["https://picsum.photos/seed/enroll2/800/400"],"desc":"每天学习英语30分钟，提升语言能力","content":[{"type":"text","val":"每天至少学习英语30分钟"},{"type":"img","val":"https://picsum.photos/seed/study1/800/400"},{"type":"text","val":"可以背单词、读文章、看视频等多种形式"}]}', '[{"label":"学习内容","type":"text","required":true},{"label":"学习时长(分钟)","type":"number","required":true}]', '', 856, 62, 4, '[{"pic":"/static/default-avatar.png","name":"张三"},{"pic":"/static/default-avatar.png","name":"赵六"}]', 1768896000000, 1768896000000, '127.0.0.1', '', NOW(), NOW()),

('每日喝水打卡', 1, 'life', '生活', 1746000000000, 1770883200000, 365, 3, 0, '[]', '{"cover":["https://picsum.photos/seed/enroll3/800/400"],"desc":"每天喝足8杯水，健康生活从喝水开始","content":[{"type":"text","val":"每天喝足8杯水（约2L）"},{"type":"img","val":"https://picsum.photos/seed/water1/800/400"},{"type":"text","val":"完成一天的目标后打卡记录"}]}', '[{"label":"喝水量(ml)","type":"number","required":true}]', '', 432, 15, 2, '[]', 1768896000000, 1768896000000, '127.0.0.1', '', NOW(), NOW()),

('21天习惯养成', 1, 'study', '学习', 1746000000000, 1769990400000, 21, 99, 1, '[]', '{"cover":["https://picsum.photos/seed/enroll4/800/400"],"desc":"21天养成一个好习惯，改变从今天开始","content":[{"type":"text","val":"21天法则，坚持21天养成一个好习惯"},{"type":"text","val":"可以自己设定想要养成的习惯，每天坚持执行"}]}', '[{"label":"今日习惯","type":"text","required":true}]', '', 298, 0, 0, '[]', 1768896000000, 1768896000000, '127.0.0.1', '', NOW(), NOW());

-- ========================================
-- 5. 用户打卡关系 (enroll_users)
-- ========================================
INSERT INTO `enroll_users` (`ENROLL_USER_ENROLL_ID`, `ENROLL_USER_MINI_OPENID`, `ENROLL_USER_JOIN_CNT`, `ENROLL_USER_DAY_CNT`, `ENROLL_USER_LAST_DAY`, `ENROLL_USER_ADD_TIME`, `ENROLL_USER_EDIT_TIME`, `ENROLL_USER_ADD_IP`, `ENROLL_USER_EDIT_IP`, `created_at`, `updated_at`) VALUES
('1', 'o1_test_openid_001', 20, 20, '2026-05-28', 1769417600000, 1769500000000, '192.168.1.1', '', NOW(), NOW()),
('1', 'o1_test_openid_002', 15, 15, '2026-05-27', 1769427600000, 1769501000000, '192.168.1.2', '', NOW(), NOW()),
('1', 'o1_test_openid_003', 18, 18, '2026-05-28', 1769437600000, 1769502000000, '192.168.1.3', '', NOW(), NOW()),
('1', 'o1_test_openid_004', 10, 10, '2026-05-26', 1769447600000, 1769503000000, '192.168.1.4', '', NOW(), NOW()),
('2', 'o1_test_openid_001', 25, 25, '2026-05-28', 1769417600000, 1769500000000, '192.168.1.1', '', NOW(), NOW()),
('2', 'o1_test_openid_004', 12, 12, '2026-05-27', 1769457600000, 1769503000000, '192.168.1.4', '', NOW(), NOW()),
('3', 'o1_test_openid_001', 8, 8, '2026-05-25', 1769417600000, 1769500000000, '192.168.1.1', '', NOW(), NOW()),
('3', 'o1_test_openid_002', 6, 6, '2026-05-24', 1769427600000, 1769501000000, '192.168.1.2', '', NOW(), NOW());

-- ========================================
-- 6. 打卡记录 (enroll_joins)
-- ========================================
INSERT INTO `enroll_joins` (`ENROLL_JOIN_ENROLL_ID`, `ENROLL_JOIN_USER_ID`, `ENROLL_JOIN_DAY`, `ENROLL_JOIN_FORMS`, `ENROLL_JOIN_STATUS`, `ENROLL_JOIN_ADD_TIME`, `ENROLL_JOIN_EDIT_TIME`, `ENROLL_JOIN_ADD_IP`, `ENROLL_JOIN_EDIT_IP`, `created_at`, `updated_at`) VALUES
('1', 'o1_test_openid_001', '2026-05-28', '[{"label":"跑步距离(公里)","value":"5"}]', 1, 1769500000000, 0, '192.168.1.1', '', NOW(), NOW()),
('1', 'o1_test_openid_001', '2026-05-27', '[{"label":"跑步距离(公里)","value":"3"}]', 1, 1769413600000, 0, '192.168.1.1', '', NOW(), NOW()),
('1', 'o1_test_openid_002', '2026-05-27', '[{"label":"跑步距离(公里)","value":"4"}]', 1, 1769413600000, 0, '192.168.1.2', '', NOW(), NOW()),
('1', 'o1_test_openid_003', '2026-05-28', '[{"label":"跑步距离(公里)","value":"6"}]', 1, 1769500000000, 0, '192.168.1.3', '', NOW(), NOW()),
('1', 'o1_test_openid_004', '2026-05-26', '[{"label":"跑步距离(公里)","value":"3.5"}]', 1, 1769327200000, 0, '192.168.1.4', '', NOW(), NOW()),
('2', 'o1_test_openid_001', '2026-05-28', '[{"label":"学习内容","value":"托福单词50个"},{"label":"学习时长(分钟)","value":"45"}]', 1, 1769500000000, 0, '192.168.1.1', '', NOW(), NOW()),
('2', 'o1_test_openid_004', '2026-05-27', '[{"label":"学习内容","value":"英语阅读文章2篇"},{"label":"学习时长(分钟)","value":"35"}]', 1, 1769413600000, 0, '192.168.1.4', '', NOW(), NOW()),
('3', 'o1_test_openid_001', '2026-05-25', '[{"label":"喝水量(ml)","value":"2000"}]', 1, 1769240800000, 0, '192.168.1.1', '', NOW(), NOW()),
('3', 'o1_test_openid_002', '2026-05-24', '[{"label":"喝水量(ml)","value":"1500"}]', 1, 1769154400000, 0, '192.168.1.2', '', NOW(), NOW());

-- ========================================
-- 7. 收藏 (favorites)
-- ========================================
INSERT INTO `favorites` (`FAV_USER_ID`, `FAV_TITLE`, `FAV_TYPE`, `FAV_OID`, `FAV_PATH`, `FAV_ADD_TIME`, `FAV_EDIT_TIME`, `FAV_ADD_IP`, `FAV_EDIT_IP`, `created_at`, `updated_at`) VALUES
('o1_test_openid_001', '每日晨跑打卡', 'enroll', '1', 'pages/enroll/detail/enroll_detail?id=1', 1769500000000, 0, '192.168.1.1', '', NOW(), NOW()),
('o1_test_openid_001', '每日英语学习', 'enroll', '2', 'pages/enroll/detail/enroll_detail?id=2', 1769500000000, 0, '192.168.1.1', '', NOW(), NOW()),
('o1_test_openid_002', '每日晨跑打卡', 'enroll', '1', 'pages/enroll/detail/enroll_detail?id=1', 1769501000000, 0, '192.168.1.2', '', NOW(), NOW());

-- ========================================
-- 8. 系统设置 (setups)
-- ========================================
INSERT INTO `setups` (`SETUP_KEY`, `SETUP_VALUE`, `SETUP_TYPE`, `SETUP_ADD_TIME`, `SETUP_EDIT_TIME`, `created_at`, `updated_at`) VALUES
('SETUP_CONTENT_ABOUT', 'WeCheckin 是一款专注于习惯打卡的小程序。我们致力于帮助用户养成良好习惯，通过每日打卡记录成长。如有任何问题，请联系客服。', '', 1768896000000, 1768896000000, NOW(), NOW()),
('about_name', '关于我们', 'text', 1768896000000, 1768896000000, NOW(), NOW());

-- ========================================
-- 9. 操作日志 (logs)
-- ========================================
INSERT INTO `logs` (`LOG_TYPE`, `LOG_CONTENT`, `LOG_ADMIN_ID`, `LOG_ADMIN_NAME`, `LOG_ADMIN_DESC`, `LOG_ADD_TIME`, `LOG_ADD_IP`, `created_at`, `updated_at`) VALUES
(1, '系统初始化', '1', 'admin', '超级管理员', 1769472000000, '127.0.0.1', NOW(), NOW()),
(1, '添加测试数据', '1', 'admin', '超级管理员', 1769472600000, '127.0.0.1', NOW(), NOW());

-- ========================================
-- 10. 菜单 (menus) - 系统配置
-- ========================================
INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`, `created_at`, `updated_at`)
SELECT '系统配置', 0, '/setup', 'setup:list,setup:edit', 'Setting', 10, 1, 1, 1769472000000, 1769472000000, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `menus` WHERE `menu_path` = '/setup');

-- ========================================
-- 11. 部门数据 (departments)
-- ========================================
INSERT INTO `departments` (`dept_name`, `dept_parent_id`, `dept_sort`, `dept_status`, `dept_add_time`, `dept_edit_time`, `dept_add_ip`, `dept_edit_ip`, `created_at`, `updated_at`)
SELECT '技术部', 0, 1, 1, 1769472000000, 1769472000000, '127.0.0.1', '127.0.0.1', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `departments` WHERE `dept_name` = '技术部');
INSERT INTO `departments` (`dept_name`, `dept_parent_id`, `dept_sort`, `dept_status`, `dept_add_time`, `dept_edit_time`, `dept_add_ip`, `dept_edit_ip`, `created_at`, `updated_at`)
SELECT '市场部', 0, 2, 1, 1769472000000, 1769472000000, '127.0.0.1', '127.0.0.1', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `departments` WHERE `dept_name` = '市场部');

-- ========================================
-- 12. 菜单 - 赛事活动
-- ========================================
INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`, `created_at`, `updated_at`)
SELECT '赛事活动', 0, '/event', 'event:list,event:add,event:edit,event:del', 'TrophyBase', 4, 1, 0, 1769472000000, 1769472000000, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `menus` WHERE `menu_path` = '/event');

-- 赛事活动子菜单（按钮权限）
INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`, `created_at`, `updated_at`)
SELECT '赛事活动列表', (SELECT `id` FROM `menus` WHERE `menu_path` = '/event' LIMIT 1), '', 'event:list', '', 1, 1, 2, 1769472000000, 1769472000000, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `menus` WHERE `menu_perms` = 'event:list');

INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`, `created_at`, `updated_at`)
SELECT '赛事活动新增', (SELECT `id` FROM `menus` WHERE `menu_path` = '/event' LIMIT 1), '', 'event:add', '', 2, 1, 2, 1769472000000, 1769472000000, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `menus` WHERE `menu_perms` = 'event:add');

INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`, `created_at`, `updated_at`)
SELECT '赛事活动编辑', (SELECT `id` FROM `menus` WHERE `menu_path` = '/event' LIMIT 1), '', 'event:edit', '', 3, 1, 2, 1769472000000, 1769472000000, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `menus` WHERE `menu_perms` = 'event:edit');

INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`, `created_at`, `updated_at`)
SELECT '赛事活动删除', (SELECT `id` FROM `menus` WHERE `menu_path` = '/event' LIMIT 1), '', 'event:del', '', 4, 1, 2, 1769472000000, 1769472000000, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `menus` WHERE `menu_perms` = 'event:del');

-- ========================================
-- 13. 重置自增ID
-- ========================================
-- 自动由MySQL处理
