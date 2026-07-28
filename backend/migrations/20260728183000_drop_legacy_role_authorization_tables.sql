-- 清理旧角色授权关系表。
-- 前置条件：已执行统一权限迁移，`role_menus` 和 `role_depts` 的数据已回填到 `permission_grants`。
-- 说明：`roles` 仍是有效业务表；`menus` 会由后续独立迁移删除，不在本迁移中处理。

DROP TABLE IF EXISTS `role_menus`;
DROP TABLE IF EXISTS `role_depts`;
