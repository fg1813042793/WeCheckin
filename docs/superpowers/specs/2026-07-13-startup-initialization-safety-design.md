# 启动初始化安全优化设计

## 目标

降低后端启动时的破坏性副作用，避免服务启动流程无条件删除历史表或潜在数据，同时保持现有自动建表、基础配置种子和菜单种子行为不变。

## 范围

本轮优化包含：

- 移除 `InitBusiness` 中无条件执行的 `DROP TABLE IF EXISTS user_form_fields`。
- 保留 `autoMigrate()`、`seedSetups()`、`seedMenus(enableExam)` 的现有调用顺序和行为。
- 补充一个不依赖真实数据库连接的轻量测试，确保启动初始化函数不再包含危险删表 SQL。
- 更新中文文档，说明用户扩展表单字段目前通过 `setups` 表中的 `SETUP_USER_FORM_FIELDS` 保存。

本轮优化不包含：

- 不重构整个数据库迁移系统。
- 不修改 GORM 模型列表和已有 `AutoMigrate` 行为。
- 不改变 `SETUP_USER_FORM_FIELDS` 的读取/保存格式。
- 不实现旧 `user_form_fields` 表到 `setups` 表的数据迁移。
- 不处理上一轮配置环境变量覆盖改动之外的其它配置问题。

## 当前观察

- 后端入口在 `backend/cmd/main.go` 中调用 `service.InitBusiness(*examFlag)`。
- `InitBusiness` 当前流程是：自动迁移、删除 `user_form_fields` 表、初始化系统设置、初始化菜单。
- 用户表单字段接口实际调用 `service.GetUserFormFields()` 和 `service.SaveUserFormFields()`。
- 用户表单字段当前通过 `SetContentSetup("SETUP_USER_FORM_FIELDS", ...)` 写入 `setups` 表，而不是写入 `user_form_fields` 表。
- 因此，启动时无条件删除 `user_form_fields` 表对当前主路径没有必要，但会给历史数据和未来迁移留下隐患。

## 设计

### 启动流程

将 `InitBusiness` 调整为：

1. 调用 `autoMigrate()`，失败时只记录 warning，保持当前容错行为。
2. 调用 `seedSetups()`。
3. 调用 `seedMenus(enableExam)`。

不再执行任何 `DROP TABLE` 语句。这样启动流程只做自动迁移和幂等种子初始化，不做破坏性清理。

### 测试策略

增加一个轻量测试文件，直接读取 `backend/internal/app/service/database.go` 源码文本，断言启动初始化代码不包含：

- `DROP TABLE`
- `user_form_fields`

这个测试不是业务逻辑测试，而是安全护栏测试。它的价值是防止未来再次把无条件删表 SQL 放回启动流程。它不需要真实 MySQL，也不会触发 `InitBusiness`。

### 文档

在 README 的配置或后端说明部分补充一句中文说明：

- 用户扩展表单字段保存在 `setups` 表的 `SETUP_USER_FORM_FIELDS` 配置项中。
- 后端启动不会清理旧 `user_form_fields` 表；如需迁移历史数据，应使用单独迁移脚本。

## 验证

运行：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service ./backend/internal/config ./backend/internal/app/formkit/...
```

预期结果：

- `service` 包安全护栏测试通过。
- `config` 包测试继续通过。
- formkit 相关回归测试继续通过。

如果测试生成 `.cache/`，验证后删除。

## 风险

- 如果某些环境依赖启动时删除旧 `user_form_fields` 表，这次变更会保留该旧表。但保留旧表比无条件删除数据更安全。
- 本轮不做历史数据迁移。如果确实存在旧表数据需要迁移，应单独设计一次“用户表单字段迁移”任务。
- 源码文本测试是安全护栏，不替代完整数据库集成测试；它只覆盖“不要把危险 SQL 放回启动流程”这一条约束。
