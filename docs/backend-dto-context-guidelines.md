# 后端 DTO 与 Context 规范

## 响应 DTO

- 高频业务接口不要直接 `response.JSON(c, map[string]interface{}{})`。
- 列表接口统一使用命名结构，例如 `ListResponse`、`pagedListResponse`。
- 动态表单、FormKit 工具、报表接口可以保留动态字段，但顶层响应应尽量使用命名 DTO。
- 结构测试已覆盖后台高频接口、客户端问卷/考试/新闻/通行证核心接口，后续新增接口应同步补测试。

## 数据库 Context

- handler 触发的查询应优先调用 service 的 `XXXContext(ctx, ...)`。
- service 内部访问数据库时使用 `database.WithContext(ctx)`，并 `defer cancel()`。
- 旧函数可以保留兼容，但内部应调用 `context.Background()` 版本的 Context 函数。

## 事务

- 一次操作写多张表时必须进入事务。
- 角色新增/编辑需要同时写角色、菜单关联、部门关联，必须使用服务层事务入口。
- 管理员新增/编辑/删除需要同时写管理员、部门关联，必须使用服务层事务入口。

## 允许的动态例外

- FormKit 工具和报表需要承载任意 schema、answers、states，可以保留 `map[string]interface{}`。
- 动态例外必须集中在专用文件，例如 `formkit_tools.go`、`formkit_report_*.go`。
- 不要把动态响应扩散到常规列表、详情、提交、登录接口。
