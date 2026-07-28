-- 删除旧后台菜单表。
--
-- 前置条件：
-- - 已执行统一权限迁移，menus 数据已回填到 permissions。
-- - 后台菜单、角色授权和接口权限均已切到 permissions / permission_grants。
--
-- 说明：
-- - roles、permissions、permission_grants 仍是有效业务表，不在本迁移中删除。
-- - 旧 /api/v2/admin/me/menus 接口仅作为前端侧栏接口名保留，数据来源已经是 permissions。
--
-- 回滚：
-- - 如果需要恢复旧表，请从发布前备份恢复 menus；不建议在新版本运行时继续写 menus。

DROP TABLE IF EXISTS `menus`;
