-- ============================================================
-- 迁移：将考试菜单从 survey 独立出来
-- 运行前请确认库名，建议先备份 menus 和 role_menu 表
-- 使用方法: mysql -h 192.168.1.109 -P 3306 -u root -p go_wecheckin < migrate_exam_menu.sql
-- ============================================================

-- 0. 清理旧考试菜单关联的角色权限
DELETE FROM `role_menu` WHERE `role_menu_menu_id` IN (
    SELECT `id` FROM `menus` WHERE `menu_path` LIKE '/survey/exam/%'
);
DELETE FROM `role_menu` WHERE `role_menu_menu_id` IN (
    SELECT `id` FROM `menus` WHERE `menu_path` = '/exam' OR `menu_path` = '/exam/list' OR `menu_path` LIKE '/exam/%'
);

-- 1. 删除旧版挂在 /survey 下的考试子菜单
DELETE FROM `menus` WHERE `menu_path` LIKE '/survey/exam/%';

-- 2. 删除旧版挂在 /survey 下的考试相关按钮权限
SET @survey_id = (SELECT `id` FROM `menus` WHERE `menu_path` = '/survey' LIMIT 1);
DELETE FROM `role_menu` WHERE `role_menu_menu_id` IN (
    SELECT `id` FROM `menus` WHERE `menu_parent_id` = @survey_id AND (
        `menu_perms` LIKE 'question:%' OR
        `menu_perms` LIKE 'paper:%' OR
        `menu_perms` LIKE 'exam:%' OR
        `menu_perms` LIKE 'record:%' OR
        `menu_perms` LIKE 'grade:%'
    )
);
DELETE FROM `menus` WHERE `menu_parent_id` = @survey_id AND (
    `menu_perms` LIKE 'question:%' OR
    `menu_perms` LIKE 'paper:%' OR
    `menu_perms` LIKE 'exam:%' OR
    `menu_perms` LIKE 'record:%' OR
    `menu_perms` LIKE 'grade:%'
);

-- 3. 更新 survey 目录的 perms 字段，去掉旧考试相关权限
UPDATE `menus` SET `menu_perms` = 'survey:list,survey:add,survey:edit,survey:del,survey:status,survey:copy,response:list,response:del,response:export'
WHERE `menu_path` = '/survey';

-- 4. 删除旧的 /exam 相关菜单（避免重复）
DELETE FROM `role_menu` WHERE `role_menu_menu_id` IN (
    SELECT `id` FROM `menus` WHERE `menu_path` = '/exam' OR `menu_path` = '/exam/list' OR `menu_path` LIKE '/exam/%'
);
DELETE FROM `menus` WHERE `menu_path` = '/exam' OR `menu_path` = '/exam/list' OR `menu_path` LIKE '/exam/%';

-- 5. 插入新考试目录
SET @now = UNIX_TIMESTAMP() * 1000;
INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`)
VALUES ('在线考试', 0, '/exam', 'exam:list,exam:add,exam:edit,exam:del', 'EditPen', 15, 1, 0, @now, @now);

-- 6. 插入考试管理子菜单
SET @exam_id = LAST_INSERT_ID();
INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_path`, `menu_perms`, `menu_icon`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`)
VALUES ('考试管理', @exam_id, '/exam/list', '', '', 1, 1, 1, @now, @now);

-- 7. 插入考试按钮权限
INSERT INTO `menus` (`menu_name`, `menu_parent_id`, `menu_perms`, `menu_sort`, `menu_status`, `menu_type`, `menu_add_time`, `menu_edit_time`)
VALUES
('考试列表', @exam_id, 'exam:list', 1, 1, 2, @now, @now),
('考试新增', @exam_id, 'exam:add', 2, 1, 2, @now, @now),
('考试编辑', @exam_id, 'exam:edit', 3, 1, 2, @now, @now),
('考试删除', @exam_id, 'exam:del', 4, 1, 2, @now, @now);
