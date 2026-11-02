# 问卷通知模版独立页面设计

## 目标

将系统配置中的“问卷通知模版”迁移为问卷调查下的独立管理页面，页面路径为 `/survey/templates`，源码位于 `admin/src/views/survey/template/index.vue`。

## 页面结构

- 使用 `AdminPageShell` 和 `AdminPageHeader`。
- 页面标题“问卷通知模版”，保留变量说明。
- 使用“内置模版 / 自定义模版”两个标签页。
- 保留新增、编辑、删除、空状态和编辑弹窗。
- 没有维护权限时只读展示，隐藏新增、编辑、删除操作。
- 系统配置页删除原标签及全部模板状态和请求逻辑，不保留重复入口。

## 数据与接口

新增管理接口：

- `GET /api/v2/admin/survey-notification-templates`
- `PUT /api/v2/admin/survey-notification-templates`

GET 返回 `{ builtin, custom }`。PUT 接收 `{ scope: "builtin" | "custom", presets }`。

内置模版继续写入 `BUILTIN_TEMPLATE_PRESETS`，自定义模版继续写入管理员专属的 `template_presets_<adminId>`，因此不迁移模板数据。原 `/survey-template-presets` 接口保留兼容。

## 权限

- 菜单：`admin:menu:survey:templates`
- 维护按钮：`admin:menu:survey:templates:edit`
- 查看接口：`admin:api:survey-template:list`
- 维护接口：`admin:api:survey-template:edit`

页面菜单位于 `admin:menu:survey` 下。历史拥有问卷查看权限的主体获得页面与查看接口；拥有问卷编辑权限的主体获得维护按钮与维护接口。

## 错误处理

- 缺少内置配置或配置损坏时返回内置默认模版。
- 模版名称或内容为空时前端阻止提交。
- 删除操作需要确认。
- 加载失败显示错误状态与重试按钮，保存失败保留当前编辑内容。

## 验证

- 静态检查确保新页面、路由、菜单、权限和系统配置移除同步完成。
- 服务与 Handler 测试覆盖默认值、作用域保存和参数错误。
- 迁移测试覆盖权限注册和历史授权回填。
- 浏览器验证菜单入口、只读/可编辑状态、内置和自定义模板 CRUD，以及系统配置不再显示旧标签。
