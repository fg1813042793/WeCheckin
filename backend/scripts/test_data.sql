-- =============================================================
-- go_wecheckin 测试数据
-- 数据库: go_wecheckin
-- 注意: 所有时间戳为毫秒级Unix时间戳
-- =============================================================

-- 清空已有数据（按外键顺序）
DELETE FROM enroll_joins;
DELETE FROM enroll_users;
DELETE FROM favorites;
DELETE FROM logs;
DELETE FROM enrolls;
DELETE FROM news;
DELETE FROM setups;
DELETE FROM users;
DELETE FROM admins;

-- ==================== admins ====================
-- 密码 admin123 -> MD5: 0192023a7bbd73250516f069df18b500
INSERT INTO `admins` (`id`, `admin_name`, `admin_password`, `admin_desc`, `admin_pic`, `admin_phone`, `admin_status`, `admin_type`, `admin_token`, `admin_token_time`, `admin_login_cnt`, `admin_login_time`, `admin_add_time`, `admin_edit_time`, `admin_add_ip`, `admin_edit_ip`, `created_at`, `updated_at`) VALUES
(1, 'admin',    '0192023a7bbd73250516f069df18b500', '超级管理员',  '', '13800000001', 1, 1, '', 0, 5, 1780243200000, 1780243200000, 1780243200000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),
(2, 'manager',  '0192023a7bbd73250516f069df18b500', '普通管理员',  '', '13800000002', 1, 0, '', 0, 2, 1780243200000, 1780243200000, 1780243200000, '192.168.1.1', '192.168.1.1', NOW(), NOW());

-- ==================== setups ====================
INSERT INTO `setups` (`id`, `setup_key`, `setup_value`, `setup_type`, `setup_add_time`, `setup_edit_time`, `created_at`, `updated_at`) VALUES
(1, 'NEWS_CATE', '[{\"id\":1,\"name\":\"公司新闻\"},{\"id\":2,\"name\":\"行业动态\"},{\"id\":3,\"name\":\"通知公告\"}]', 'json', 1780243200000, 1780243200000, NOW(), NOW()),
(2, 'ENROLL_CATE', '[{\"id\":1,\"name\":\"运动打卡\"},{\"id\":2,\"name\":\"学习打卡\"},{\"id\":3,\"name\":\"健康打卡\"},{\"id\":4,\"name\":\"工作打卡\"}]', 'json', 1780243200000, 1780243200000, NOW(), NOW()),
(3, 'ABOUT', '{\"title\":\"关于我们\",\"content\":\"WeCheckin 是一款打卡签到系统\"}', 'json', 1780243200000, 1780243200000, NOW(), NOW());

-- ==================== users ====================
INSERT INTO `users` (`id`, `user_mini_openid`, `user_status`, `user_check_reason`, `user_name`, `user_mobile`, `user_pic`, `user_forms`, `user_obj`, `user_login_cnt`, `user_login_time`, `user_add_time`, `user_add_ip`, `user_edit_time`, `user_edit_ip`, `created_at`, `updated_at`) VALUES
(1, 'openid_test_001', 1, '', '张三', '13800138001', 'https://example.com/avatar/1.jpg', '', '', 10, 1780243200000, 1780243200000, '10.0.0.1', 1780243200000, '10.0.0.1', NOW(), NOW()),
(2, 'openid_test_002', 1, '', '李四', '13800138002', 'https://example.com/avatar/2.jpg', '', '', 8, 1780243200000, 1780243200000, '10.0.0.2', 1780243200000, '10.0.0.2', NOW(), NOW()),
(3, 'openid_test_003', 1, '', '王五', '13800138003', 'https://example.com/avatar/3.jpg', '', '', 5, 1780243200000, 1780243200000, '10.0.0.3', 1780243200000, '10.0.0.3', NOW(), NOW()),
(4, 'openid_test_004', 0, '未通过审核', '赵六', '13800138004', 'https://example.com/avatar/4.jpg', '', '', 0, 0, 1780243200000, '10.0.0.4', 1780243200000, '10.0.0.4', NOW(), NOW()),
(5, 'openid_test_005', 1, '', '孙七', '13800138005', 'https://example.com/avatar/5.jpg', '', '', 3, 1780243200000, 1780243200000, '10.0.0.5', 1780243200000, '10.0.0.5', NOW(), NOW());

-- ==================== news ====================
INSERT INTO `news` (`id`, `news_title`, `news_desc`, `news_status`, `news_cate_id`, `news_cate_name`, `news_order`, `news_vouch`, `news_content`, `news_qr`, `news_view_cnt`, `news_pic`, `news_forms`, `news_obj`, `news_add_time`, `news_edit_time`, `news_add_ip`, `news_edit_ip`, `created_at`, `updated_at`) VALUES
(1, 'WeCheckin 正式上线', 'WeCheckin 打卡签到系统正式发布上线', 1, '1', '公司新闻', 1, 1, '经过数月开发，WeCheckin 打卡签到系统正式上线了。\n\n本系统支持多种打卡模式，满足不同场景需求。', '', 120, '[]', '', '', 1780243200000, 1780243200000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),
(2, '关于启用健康打卡的通知', '即日起全体员工需完成每日健康打卡', 1, '3', '通知公告', 2, 1, '为保障员工健康，即日起启用健康打卡功能。\n\n请各位同事每日上班前完成健康打卡。', '', 85, '[]', '', '', 1780243200000, 1780243200000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),
(3, '行业动态：数字化办公新趋势', '数字化转型正在改变企业的办公方式', 1, '2', '行业动态', 3, 0, '随着数字化转型的深入，越来越多的企业开始采用数字化工具进行管理。\n\n打卡签到作为企业管理的基础环节，也在不断创新。', '', 210, '[]', '', '', 1780243200000, 1780243200000, '127.0.0.1', '127.0.0.1', NOW(), NOW());

-- ==================== enrolls ====================
-- 时间戳: 2026-05-01 ~ 2026-06-30
-- enroll_obj JSON: {"cover":["https://example.com/cover/1.jpg"],"desc":"每日运动打卡，健康生活"}
INSERT INTO `enrolls` (`id`, `enroll_title`, `enroll_status`, `enroll_cate_id`, `enroll_cate_name`, `enroll_start`, `enroll_end`, `enroll_day_cnt`, `enroll_order`, `enroll_vouch`, `enroll_forms`, `enroll_obj`, `enroll_join_forms`, `enroll_repeat`, `enroll_limit`, `enroll_qr`, `enroll_view_cnt`, `enroll_join_cnt`, `enroll_user_cnt`, `enroll_user_list`, `enroll_add_time`, `enroll_edit_time`, `enroll_add_ip`, `enroll_edit_ip`, `created_at`, `updated_at`) VALUES
(1, '每日运动打卡',   1, '1', '运动打卡', 1777564800000, 1782748800000, 61, 1, 1,
 '[{"label":"运动类型","type":"select","options":["跑步","散步","瑜伽","健身","其他"],"require":true},{"label":"运动时长(分钟)","type":"number","require":true},{"label":"备注","type":"textarea","require":false}]',
 '{"cover":["https://example.com/cover/1.jpg"],"desc":"每日运动打卡，健康生活每一天"}',
 '[{"label":"运动图片","type":"image","require":true},{"label":"运动位置","type":"location","require":false},{"label":"备注","type":"textarea","require":false}]',
 0, 1, '', 150, 0, 0, '', 1777564800000, 1777564800000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),

(2, '读书学习打卡',   1, '2', '学习打卡', 1777564800000, 1782748800000, 61, 2, 1,
 '[{"label":"学习内容","type":"text","require":true},{"label":"学习时长(小时)","type":"number","require":true}]',
 '{"cover":["https://example.com/cover/2.jpg"],"desc":"每日读书学习，提升自我"}',
 '[{"label":"学习笔记","type":"textarea","require":true},{"label":"学习图片","type":"image","require":false}]',
 0, 3, '', 80, 0, 0, '', 1777564800000, 1777564800000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),

(3, '早起打卡',       1, '1', '运动打卡', 1777564800000, 1779724800000, 31, 3, 1,
 '[{"label":"起床时间","type":"text","require":true}]',
 '{"cover":["https://example.com/cover/3.jpg"],"desc":"每天6点前起床打卡"}',
 '[{"label":"照片","type":"image","require":true}]',
 0, 1, '', 200, 0, 0, '', 1777564800000, 1777564800000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),

(4, '喝水打卡',       1, '3', '健康打卡', 1777564800000, 1782748800000, 61, 4, 0,
 '[{"label":"饮水量(ml)","type":"number","require":true}]',
 '{"cover":["https://example.com/cover/4.jpg"],"desc":"每天喝8杯水，保持健康"}',
 '[{"label":"记录图片","type":"image","require":false}]',
 '', 45, 0, 0, '', 1777564800000, 1777564800000, '127.0.0.1', '127.0.0.1', NOW(), NOW()),

(5, '周报提交打卡',   0, '4', '工作打卡', 1777564800000, 1782748800000, 61, 5, 0,
 '[{"label":"周报内容","type":"textarea","require":true}]',
 '{"cover":[],"desc":"每周五提交周报"}',
 '[{"label":"周报文件","type":"file","require":true}]',
 '', 30, 0, 0, '', 1777564800000, 1777564800000, '127.0.0.1', '127.0.0.1', NOW(), NOW());

-- ==================== enroll_users ====================
INSERT INTO `enroll_users` (`id`, `enroll_user_enroll_id`, `enroll_user_mini_openid`, `enroll_user_join_cnt`, `enroll_user_day_cnt`, `enroll_user_last_day`, `enroll_user_add_time`, `enroll_user_edit_time`, `enroll_user_add_ip`, `enroll_user_edit_ip`, `created_at`, `updated_at`) VALUES
(1, '1', 'openid_test_001', 15, 12, '2026-06-01', 1777564800000, 1780243200000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(2, '1', 'openid_test_002', 10, 8, '2026-05-30', 1777564800000, 1780070400000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(3, '1', 'openid_test_003', 5, 4, '2026-05-25', 1778112000000, 1779552000000, '10.0.0.3', '10.0.0.3', NOW(), NOW()),
(4, '2', 'openid_test_001', 8, 7, '2026-06-01', 1777564800000, 1780243200000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(5, '2', 'openid_test_005', 6, 5, '2026-05-28', 1777647600000, 1779811200000, '10.0.0.5', '10.0.0.5', NOW(), NOW()),
(6, '3', 'openid_test_001', 20, 18, '2026-06-01', 1777564800000, 1780243200000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(7, '3', 'openid_test_002', 18, 15, '2026-05-31', 1777564800000, 1780070400000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(8, '3', 'openid_test_003', 12, 10, '2026-05-29', 1777647600000, 1779897600000, '10.0.0.3', '10.0.0.3', NOW(), NOW()),
(9, '3', 'openid_test_005', 8, 7, '2026-05-27', 1777734000000, 1779724800000, '10.0.0.5', '10.0.0.5', NOW(), NOW()),
(10, '4', 'openid_test_002', 3, 3, '2026-05-20', 1778112000000, 1779206400000, '10.0.0.2', '10.0.0.2', NOW(), NOW());

-- ==================== enroll_joins ====================
-- 张三在"每日运动打卡"的打卡记录
INSERT INTO `enroll_joins` (`id`, `enroll_join_enroll_id`, `enroll_join_user_id`, `enroll_join_day`, `enroll_join_forms`, `enroll_join_status`, `enroll_join_add_time`, `enroll_join_edit_time`, `enroll_join_add_ip`, `enroll_join_edit_ip`, `created_at`, `updated_at`) VALUES
(1,  '1', 'openid_test_001', '2026-05-20', '[{"label":"运动类型","value":"跑步"},{"label":"运动时长(分钟)","value":"30"},{"label":"备注","value":"晨跑"}]', 1, 1779206400000, 1779206400000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(2,  '1', 'openid_test_001', '2026-05-21', '[{"label":"运动类型","value":"跑步"},{"label":"运动时长(分钟)","value":"35"},{"label":"备注","value":""}]', 1, 1779292800000, 1779292800000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(3,  '1', 'openid_test_001', '2026-05-22', '[{"label":"运动类型","value":"健身"},{"label":"运动时长(分钟)","value":"45"},{"label":"备注","value":"力量训练"}]', 1, 1779379200000, 1779379200000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(4,  '1', 'openid_test_001', '2026-05-23', '[{"label":"运动类型","value":"散步"},{"label":"运动时长(分钟)","value":"20"},{"label":"备注","value":""}]', 1, 1779465600000, 1779465600000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(5,  '1', 'openid_test_001', '2026-05-25', '[{"label":"运动类型","value":"跑步"},{"label":"运动时长(分钟)","value":"30"},{"label":"备注","value":"周末晨跑"}]', 1, 1779638400000, 1779638400000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
-- 李四在"每日运动打卡"的打卡记录
(6,  '1', 'openid_test_002', '2026-05-20', '[{"label":"运动类型","value":"瑜伽"},{"label":"运动时长(分钟)","value":"40"},{"label":"备注","value":"晚间瑜伽"}]', 1, 1779206400000, 1779206400000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(7,  '1', 'openid_test_002', '2026-05-22', '[{"label":"运动类型","value":"瑜伽"},{"label":"运动时长(分钟)","value":"30"},{"label":"备注","value":""}]', 1, 1779379200000, 1779379200000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(8,  '1', 'openid_test_002', '2026-05-24', '[{"label":"运动类型","value":"散步"},{"label":"运动时长(分钟)","value":"60"},{"label":"备注","value":"公园散步"}]', 1, 1779552000000, 1779552000000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
-- 王五在"每日运动打卡"的打卡记录
(9,  '1', 'openid_test_003', '2026-05-22', '[{"label":"运动类型","value":"跑步"},{"label":"运动时长(分钟)","value":"25"},{"label":"备注","value":""}]', 1, 1779379200000, 1779379200000, '10.0.0.3', '10.0.0.3', NOW(), NOW()),
(10, '1', 'openid_test_003', '2026-05-24', '[{"label":"运动类型","value":"其他"},{"label":"运动时长(分钟)","value":"15"},{"label":"备注","value":"拉伸"}]', 1, 1779552000000, 1779552000000, '10.0.0.3', '10.0.0.3', NOW(), NOW()),
-- 张三在"读书学习打卡"的打卡记录
(11, '2', 'openid_test_001', '2026-05-21', '[{"label":"学习内容","value":"Go语言编程"},{"label":"学习时长(小时)","value":"2"}]', 1, 1779292800000, 1779292800000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(12, '2', 'openid_test_001', '2026-05-23', '[{"label":"学习内容","value":"数据结构"},{"label":"学习时长(小时)","value":"1.5"}]', 1, 1779465600000, 1779465600000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(13, '2', 'openid_test_001', '2026-05-25', '[{"label":"学习内容","value":"系统设计"},{"label":"学习时长(小时)","value":"3"}]', 1, 1779638400000, 1779638400000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
-- 张三在"早起打卡"的打卡记录
(14, '3', 'openid_test_001', '2026-05-20', '[{"label":"起床时间","value":"05:30"}]', 1, 1779206400000, 1779206400000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(15, '3', 'openid_test_001', '2026-05-21', '[{"label":"起床时间","value":"05:45"}]', 1, 1779292800000, 1779292800000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(16, '3', 'openid_test_001', '2026-05-22', '[{"label":"起床时间","value":"05:30"}]', 1, 1779379200000, 1779379200000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(17, '3', 'openid_test_001', '2026-05-23', '[{"label":"起床时间","value":"06:00"}]', 1, 1779465600000, 1779465600000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
-- 李四在"早起打卡"的打卡记录
(18, '3', 'openid_test_002', '2026-05-20', '[{"label":"起床时间","value":"05:30"}]', 1, 1779206400000, 1779206400000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(19, '3', 'openid_test_002', '2026-05-21', '[{"label":"起床时间","value":"05:40"}]', 1, 1779292800000, 1779292800000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(20, '3', 'openid_test_002', '2026-05-22', '[{"label":"起床时间","value":"05:30"}]', 1, 1779379200000, 1779379200000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
-- 李四在"喝水打卡"的打卡记录
(21, '4', 'openid_test_002', '2026-05-18', '[{"label":"饮水量(ml)","value":"2000"}]', 1, 1779033600000, 1779033600000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(22, '4', 'openid_test_002', '2026-05-19', '[{"label":"饮水量(ml)","value":"1500"}]', 1, 1779120000000, 1779120000000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(23, '4', 'openid_test_002', '2026-05-20', '[{"label":"饮水量(ml)","value":"2500"}]', 1, 1779206400000, 1779206400000, '10.0.0.2', '10.0.0.2', NOW(), NOW());

-- ==================== favorites ====================
INSERT INTO `favorites` (`id`, `fav_user_id`, `fav_title`, `fav_type`, `fav_oid`, `fav_path`, `fav_add_time`, `fav_edit_time`, `fav_add_ip`, `fav_edit_ip`, `created_at`, `updated_at`) VALUES
(1, 'openid_test_001', '每日运动打卡',   'enroll', '1', '/pages/enroll/enroll_detail?id=1', 1779206400000, 1779206400000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(2, 'openid_test_001', '早起打卡',       'enroll', '3', '/pages/enroll/enroll_detail?id=3', 1779292800000, 1779292800000, '10.0.0.1', '10.0.0.1', NOW(), NOW()),
(3, 'openid_test_002', '每日运动打卡',   'enroll', '1', '/pages/enroll/enroll_detail?id=1', 1779206400000, 1779206400000, '10.0.0.2', '10.0.0.2', NOW(), NOW()),
(4, 'openid_test_003', '早起打卡',       'enroll', '3', '/pages/enroll/enroll_detail?id=3', 1779379200000, 1779379200000, '10.0.0.3', '10.0.0.3', NOW(), NOW());

-- ==================== logs ====================
INSERT INTO `logs` (`id`, `log_type`, `log_content`, `log_admin_id`, `log_admin_name`, `log_admin_desc`, `log_add_time`, `log_add_ip`, `created_at`, `updated_at`) VALUES
(1, 1, '管理员登录系统',       '1', 'admin',   '超级管理员', 1780243200000, '127.0.0.1', NOW(), NOW()),
(2, 1, '管理员登录系统',       '2', 'manager', '普通管理员', 1780243200000, '192.168.1.1', NOW(), NOW()),
(3, 2, '新增打卡项目: 每日运动打卡', '1', 'admin',   '超级管理员', 1777564800000, '127.0.0.1', NOW(), NOW()),
(4, 2, '新增打卡项目: 早起打卡',     '1', 'admin',   '超级管理员', 1777564800000, '127.0.0.1', NOW(), NOW()),
(5, 3, '发布新闻: WeCheckin 正式上线', '1', 'admin', '超级管理员', 1780243200000, '127.0.0.1', NOW(), NOW());
