# 统一权限体系设计

## 背景

当前权限分散在 `menus`、`role_menus`、`roles.role_data_scope`、`role_depts`、`user_depts` 以及少量历史后台字段中。这个模型可以支撑后台菜单和按钮，但继续扩展客户端菜单、钉钉 H5 菜单、用户单独授权时会不断新增关系表，授权判断也会变得分散。

本设计引入统一权限定义表 `permissions` 和统一授权表 `permission_grants`。第一阶段会先迁移旧授权数据，再清理旧角色授权关系表；业务读取统一表，避免后续继续扩散 `role_menus`、`role_depts` 这类分散授权表。

## 目标

- 用 `permissions` 统一保存后台菜单、后台按钮/API、后台 API 权限点、客户端菜单、钉钉 H5 菜单、数据权限类型。
- 用 `permission_grants` 统一保存授权关系，支持 `subject_type = role/user`。
- 后台登录准入改为检查 `admin:login` 权限，不再依赖 `user_admin_enabled`、`role_allow_admin_login` 等旧字段。
- 后台菜单、按钮权限读取走统一授权表；旧 `role_menus` 只作为一次性迁移来源。
- 数据权限用 `data:all`、`data:dept`、`data:self`、`data:custom` 表达；自定义部门范围保存在授权记录的 `scope_value` JSON 中。
- 用户仍绑定角色，同时允许用户独立补充或拒绝权限。

## 表设计

### permissions

权限定义表，每一行代表一个权限点。

- `permission_key`：全局唯一权限编码，例如 `admin:login`、`admin:menu:12`、`admin:api:user:list`、`data:self`、`client:menu:home`、`dingtalk_h5:menu:dashboard`。
- `permission_name`：中文名称。
- `permission_platform`：平台，取值建议为 `admin`、`client`、`dingtalk_h5`、`data`。
- `permission_type`：类型，取值建议为 `login`、`menu`、`button`、`api`、`data`。
- `permission_parent_key`：树形父节点 key，后台菜单从 `menus.menu_parent_id` 迁移得到。
- `permission_resource_id`：旧资源 ID，后台菜单第一阶段保存 `menus.id`。
- `permission_resource_path`：菜单路径或资源路径。
- `permission_perms`：兼容现有前端按钮权限码，例如 `user:add`。
- `permission_status`、`permission_sort`：状态和排序。

### permission_grants

统一授权表，每一行代表一个主体对一个权限点的授权。

- `grant_subject_type`：授权主体类型，`role` 或 `user`。
- `grant_subject_id`：角色 ID 或用户 ID。
- `grant_permission_key`：权限 key。
- `grant_permission_id`：权限 ID，用于查询优化。
- `grant_effect`：授权效果，`allow` 或 `deny`。用户级 `deny` 可以覆盖角色授权。
- `grant_scope_value`：范围配置 JSON。`data:custom` 使用 `{"deptIds":[1,2,3]}`。
- `grant_source`：来源，`legacy`、`form`、`system` 等。
- `grant_status`：状态。

## 迁移规则

第一阶段自动迁移旧数据：

1. 创建内置权限：`admin:login`、`data:all`、`data:dept`、`data:self`、`data:custom`。
2. 将 `menus` 迁移为 `permissions`，权限 key 使用 `admin:menu:<menus.id>`。
3. 将后台 API 动作码同步为 `admin:api:*` 权限点。
4. 将客户端和钉钉 H5 菜单同步为 `client:menu:*`、`dingtalk_h5:menu:*` 权限点。
5. 将 `role_menus` 迁移为角色的 `permission_grants`。
6. 将 `roles.role_allow_admin_login = 1` 迁移为角色的 `admin:login` 授权。
7. 将 `roles.role_data_scope` 迁移为对应数据权限授权。
8. 将 `role_depts` 合并进 `data:custom` 授权的 `scope_value`。

迁移完成后清理旧授权关系表：`menus` 仍用于后台菜单管理和菜单树展示，`role_menus`、`role_depts` 在回填到 `permission_grants` 后删除。

## 读取优先级

后台登录：

1. 用户状态 `user_status = 1`。
2. 用户绑定角色 `user_role_id > 0`。
3. 角色启用 `roles.role_status = 1`。
4. 用户或角色拥有 `admin:login` 授权；保留的“超级管理员”角色允许兜底通过。

后台菜单：

1. 读取角色和用户的统一授权。
2. 用户级 `deny` 移除角色已授权菜单。
3. 用户级 `allow` 补充角色未授权菜单。
4. 统一权限表不可用时返回空权限，启动迁移失败时不进入后台业务请求。

按钮/API 权限：

1. 后台路由必须有明确权限声明；未声明的 `/admin`、`/api/v2/admin` 路由默认拒绝。
2. API 权限点写入 `permissions`，编码形如 `admin:api:user:list`、`admin:api:survey:edit`。
3. 用户级 `deny` 某个 `admin:api:*` 会优先于角色菜单按钮权限。
4. 用户或角色被单独授予 `admin:api:*` 时，可以直接放行对应动作码的后台接口。
5. 根据已授权的 `permissions.permission_perms` 聚合按钮权限码，兼容前端使用的 `user:add`、`role:edit` 等权限码。

数据权限：

1. 优先读取统一授权中的 `data:*`。
2. `data:all` 不加过滤。
3. `data:dept` 使用用户所在部门及子部门。
4. `data:self` 使用本人创建/本人所属范围。
5. `data:custom` 使用授权记录中的 `scope_value.deptIds`。
6. 如统一授权不存在，回退到 `roles.role_data_scope`；自定义部门范围只从 `permission_grants.grant_scope_value` 读取。
7. 后台列表、详情、编辑、删除、批量删除、导出等高风险接口应叠加数据范围过滤，不能只依赖菜单/API 入口权限。
8. 部门或自定义部门范围为空时按“无可见数据”处理，不退化成全部数据。

客户端与钉钉 H5 菜单：

1. 客户端菜单权限编码形如 `client:menu:home`、`client:menu:survey`、`client:menu:event_manage`。
2. 钉钉 H5 菜单权限编码形如 `dingtalk_h5:menu:dashboard`、`dingtalk_h5:menu:summary`、`dingtalk_h5:menu:template`。
3. 钉钉 H5 `bootstrap` 返回当前用户可见菜单；没有配置统一授权时，继续按历史角色菜单兜底，保证单点部署兼容。
4. 客户端菜单已经进入统一权限目录，后续客户端首页/个人中心可逐步改为读取后端授权菜单。

## 表单策略

角色表单第一阶段仍保留菜单树选择，但“后台登录”展示改为“后台入口权限”，保存时写入 `admin:login` 授权，同时兼容写旧字段。

用户表单保留角色、密码、部门、岗位，并新增“独立权限覆盖”的后端能力。UI 可以先轻量展示为可选面板：允许补充后台入口权限、拒绝后台入口权限、数据范围覆盖。后续再扩展完整菜单级用户授权面板。

## 风险与回滚

- 迁移顺序必须保证先回填 `permission_grants`，再删除 `role_menus`、`role_depts`。
- 写入只写统一授权表，避免旧授权表被重新创建或继续漂移。
- 如果统一权限迁移异常，启动迁移返回错误，不静默忽略。
- 回滚到旧版本前需要先从 `permission_grants` 导出角色菜单和自定义部门范围，重建旧授权表。

## 验收标准

- 启动迁移包含 `permissions`、`permission_grants`。
- 旧菜单和角色授权可以自动迁移到统一表。
- 后台登录读取 `admin:login` 权限。
- 后台菜单、按钮和 API 权限优先走统一表，用户独立授权可以补充或拒绝权限。
- 后台所有 `/api/v2/admin` 路由必须有权限声明；未声明路由默认拒绝。
- 问卷、考试、题库、报名、通知、赛事、用户等后台资源操作必须叠加数据范围过滤。
- 客户端菜单和钉钉 H5 菜单在统一权限目录中可查询；钉钉 H5 已从 `bootstrap` 返回菜单。
- 角色表单不再表达为旧“后台登录开关”，而是“后台入口权限”。
