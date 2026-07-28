-- 将附件中的用户加入“米视科技”组织架构。
--
-- 字段映射：
-- - id -> users.user_mini_openid
-- - name -> users.user_name
-- - password -> users.user_password，此处使用兼容登录逻辑的 MD5('123456')
-- - departmentLevel1/2/3 -> departments + user_depts
--
-- 当前用户表没有 role、managerId、hrbpId、responsibleDepartments 的持久化字段，
-- 按需求不写入这些字段。

SET @seed_now_ms := CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);

DROP TEMPORARY TABLE IF EXISTS `tmp_mishi_user_seed`;
CREATE TEMPORARY TABLE `tmp_mishi_user_seed` (
  `mini_openid` VARCHAR(200) NOT NULL PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL,
  `dept_l1` VARCHAR(100) NOT NULL,
  `dept_l2` VARCHAR(100) NOT NULL DEFAULT '',
  `dept_l3` VARCHAR(100) NOT NULL DEFAULT ''
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `tmp_mishi_user_seed` (`mini_openid`, `name`, `dept_l1`, `dept_l2`, `dept_l3`) VALUES
('lip', 'Lip', 'M/H业务', '研发部', 'Java开发一组'),
('arthur', 'Arthur', 'M/H业务', '研发部', '安卓开发一组'),
('foster', 'Foster', 'M/H业务', '研发部', '运维组'),
('cube', 'Cube', 'M/H业务', '产品部', '国内组'),
('hrbp', 'HRBP', 'M/H业务', '', ''),
('nick', 'Nick', 'M/H业务', '', ''),
('rock', 'Rock', 'M/H业务', '研发部', 'Java开发一组'),
('sherif', 'Sherif', 'M/H业务', '研发部', '安卓开发二组'),
('paul', 'Paul', 'M/H业务', '研发部', 'Java开发一组'),
('neil', 'Neil', 'M/H业务', '研发部', '安卓开发一组'),
('david', 'David', 'M/H业务', '研发部', ''),
('monica', 'Monica', 'M/H业务', '综合部', ''),
('lucky', 'Lucky', 'M/H业务', '综合部', ''),
('betty', 'Betty', 'M/H业务线', '综合部', ''),
('cherry', 'Cherry', 'M/H业务', '综合部', ''),
('amy', 'Amy', 'M/H业务', '综合部', '');

INSERT INTO `departments` (`dept_name`, `dept_parent_id`, `dept_sort`, `dept_status`, `dept_add_time`, `dept_edit_time`, `dept_add_ip`, `dept_edit_ip`, `created_at`, `updated_at`)
SELECT '米视科技', 0, 1, 1, @seed_now_ms, @seed_now_ms, '127.0.0.1', '127.0.0.1', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM `departments`
  WHERE `dept_name` = '米视科技' AND `dept_parent_id` = 0
);

SET @mishi_dept_id := (
  SELECT `id` FROM `departments`
  WHERE `dept_name` = '米视科技' AND `dept_parent_id` = 0
  ORDER BY `id` LIMIT 1
);

INSERT INTO `departments` (`dept_name`, `dept_parent_id`, `dept_sort`, `dept_status`, `dept_add_time`, `dept_edit_time`, `dept_add_ip`, `dept_edit_ip`, `created_at`, `updated_at`)
SELECT DISTINCT t.`dept_l1`, @mishi_dept_id, 10, 1, @seed_now_ms, @seed_now_ms, '127.0.0.1', '127.0.0.1', NOW(), NOW()
FROM `tmp_mishi_user_seed` t
LEFT JOIN `departments` existing
  ON existing.`dept_name` = t.`dept_l1`
  AND existing.`dept_parent_id` = @mishi_dept_id
WHERE t.`dept_l1` <> ''
  AND existing.`id` IS NULL;

INSERT INTO `departments` (`dept_name`, `dept_parent_id`, `dept_sort`, `dept_status`, `dept_add_time`, `dept_edit_time`, `dept_add_ip`, `dept_edit_ip`, `created_at`, `updated_at`)
SELECT DISTINCT t.`dept_l2`, p1.`id`, 20, 1, @seed_now_ms, @seed_now_ms, '127.0.0.1', '127.0.0.1', NOW(), NOW()
FROM `tmp_mishi_user_seed` t
JOIN `departments` p1
  ON p1.`dept_name` = t.`dept_l1`
  AND p1.`dept_parent_id` = @mishi_dept_id
LEFT JOIN `departments` existing
  ON existing.`dept_name` = t.`dept_l2`
  AND existing.`dept_parent_id` = p1.`id`
WHERE t.`dept_l2` <> ''
  AND existing.`id` IS NULL;

INSERT INTO `departments` (`dept_name`, `dept_parent_id`, `dept_sort`, `dept_status`, `dept_add_time`, `dept_edit_time`, `dept_add_ip`, `dept_edit_ip`, `created_at`, `updated_at`)
SELECT DISTINCT t.`dept_l3`, p2.`id`, 30, 1, @seed_now_ms, @seed_now_ms, '127.0.0.1', '127.0.0.1', NOW(), NOW()
FROM `tmp_mishi_user_seed` t
JOIN `departments` p1
  ON p1.`dept_name` = t.`dept_l1`
  AND p1.`dept_parent_id` = @mishi_dept_id
JOIN `departments` p2
  ON p2.`dept_name` = t.`dept_l2`
  AND p2.`dept_parent_id` = p1.`id`
LEFT JOIN `departments` existing
  ON existing.`dept_name` = t.`dept_l3`
  AND existing.`dept_parent_id` = p2.`id`
WHERE t.`dept_l3` <> ''
  AND existing.`id` IS NULL;

INSERT INTO `users` (
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
  t.`mini_openid`,
  1,
  t.`name`,
  '',
  '/static/default-avatar.png',
  '[]',
  '{}',
  MD5('123456'),
  0,
  0,
  @seed_now_ms,
  '127.0.0.1',
  @seed_now_ms,
  '127.0.0.1',
  NOW(),
  NOW()
FROM `tmp_mishi_user_seed` t
ON DUPLICATE KEY UPDATE
  `user_status` = VALUES(`user_status`),
  `user_name` = VALUES(`user_name`),
  `user_password` = VALUES(`user_password`),
  `user_pic` = VALUES(`user_pic`),
  `user_edit_time` = VALUES(`user_edit_time`),
  `user_edit_ip` = VALUES(`user_edit_ip`),
  `updated_at` = NOW();

INSERT INTO `user_depts` (`user_dept_user_id`, `user_dept_dept_id`, `created_at`, `updated_at`)
SELECT
  u.`id`,
  COALESCE(d3.`id`, d2.`id`, d1.`id`) AS `dept_id`,
  NOW(),
  NOW()
FROM `tmp_mishi_user_seed` t
JOIN `users` u
  ON u.`user_mini_openid` = t.`mini_openid`
JOIN `departments` d1
  ON d1.`dept_name` = t.`dept_l1`
  AND d1.`dept_parent_id` = @mishi_dept_id
LEFT JOIN `departments` d2
  ON t.`dept_l2` <> ''
  AND d2.`dept_name` = t.`dept_l2`
  AND d2.`dept_parent_id` = d1.`id`
LEFT JOIN `departments` d3
  ON t.`dept_l3` <> ''
  AND d3.`dept_name` = t.`dept_l3`
  AND d3.`dept_parent_id` = d2.`id`
LEFT JOIN `user_depts` existing
  ON existing.`user_dept_user_id` = u.`id`
  AND existing.`user_dept_dept_id` = COALESCE(d3.`id`, d2.`id`, d1.`id`)
WHERE COALESCE(d3.`id`, d2.`id`, d1.`id`) IS NOT NULL
  AND existing.`id` IS NULL;

DROP TEMPORARY TABLE IF EXISTS `tmp_mishi_user_seed`;
