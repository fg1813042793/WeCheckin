# Backend 开发规范

## 1. 适用范围

本文档适用于 `backend/` 下的 Go 服务、维护命令、SQL 迁移、Swagger 声明和测试。同时遵守 [WeCheckin 统一开发规范](../../docs/development-guidelines.md)。

必读专题文档：

- [后端 DTO 与 Context 规范](../../docs/backend-dto-context-guidelines.md)
- [API v2 接口说明](../../docs/API_V2.md)
- [权限编码与前端同步](../../docs/PERMISSION_CODE_FRONTEND_SYNC.md)
- [项目维护说明](../../docs/project-maintenance.md)
- [Service 目录说明](../internal/service/README.md)
- [Backend 测试布局](../test/README.md)

## 2. 技术基线

- Go module 位于 `backend/go.mod`，使用 Go 1.24 系列。
- HTTP 框架使用 Hertz，持久化使用 GORM + MySQL，缓存/会话使用 Redis。
- 主服务入口位于 `cmd/`，维护任务位于 `cmd/maintenance/`，独立运行时进程使用独立 `cmd/<name>/`。
- 格式化以 `gofmt` 为准，不在没有项目配置时声称已启用 golangci-lint。

## 3. 目录与分层

### 3.1 基础目录

- `cmd/`：进程入口、依赖装配和启动。不承载可复用业务逻辑。
- `config/`：YAML 默认配置和环境配置。真实密钥通过环境变量注入。
- `migrations/`：版本化 SQL 迁移和内置数据演进。
- `pkg/`：不依赖具体业务域的共享基础能力，例如数据库、日志、响应、token 和密码工具。
- `test/`：跨包架构、路由、权限、迁移和仓库级保护测试。包内单元测试留在实现文件附近。

### 3.2 `internal/` 边界

- `routes/`：路由分组、中间件装配、handler 绑定和 Swagger 路由目录。
- `handler/`：HTTP 边界，负责参数解析、请求校验、调用 service/application 和输出响应。不直接编排多表写入。
- `service/`：按 `admin`、`client`、`dingtalkh5` 拆分的业务编排。新文件使用明确业务前缀，文件过大时按职责拆分。
- `modules/`：边界清晰的独立业务模块，按 `domain`、`application`、`infrastructure`、`transport` 或模块实际需要分层。
- `model/`：GORM 持久化模型，按业务域拆分。不在模型中隐藏复杂业务流转。
- `support/`：跨业务包的稳定支撑能力，例如权限目录、数据范围、部门、媒体地址、查询和发布范围。
- `middleware/`：认证、权限、请求上下文和通用 HTTP 拦截。
- `bootstrap/`：启动装配、迁移执行和安全护栏。

### 3.3 通用模块规则

- 纯领域状态和流转不依赖 Hertz、GORM 或全局数据库。
- application/service 负责权限、事务、幂等、事件和外部协作编排。
- infrastructure 实现持久化、第三方系统、通知和调度边界。
- transport/handler 只做协议适配。模块有多个入口时，按 `httpadmin`、`httpclient` 等边界拆分。
- `internal/workflowcore` 承载流程定义核心、表单校验和 BPMN 编译；`internal/modules/workflow` 承载运行时状态机与应用服务，不应将两者合并为同一个包。

## 4. Go 编码约定

- 新增和修改的 Go 文件必须执行 `gofmt`。
- package 名使用简短小写单词，不使用下划线；导出名称能从所属 package 理解时不重复包名。
- 错误使用 `errors.Is`/`errors.As` 可判断的稳定类型或 sentinel error，在边界层统一转换为 HTTP 响应。
- 不用 panic 表示业务错误。启动前置条件失败应返回或记录可诊断错误。
- 复杂分支、领域状态或协议映射优先使用类型、枚举常量和明确函数，不依赖分散字符串判断。
- 注释解释“为什么”和特殊约束，不复述代码本身。

## 5. HTTP、DTO 与 Context

- handler 对 path/query/body 来源分别解析并验证，同一业务值同时出现在多个来源时必须检查冲突。
- 稳定接口使用命名请求/响应 DTO。FormKit 或报表可在业务层使用动态字段，顶层响应仍应尽量稳定。
- handler 向 service/application 传递请求 Context。数据库查询通过 `database.WithContext(ctx)` 执行，不在请求链路中丢失取消和超时信号。
- 兼容旧无 Context API 时，由旧函数调用 Context 版本，新请求链路不继续扩散旧入口。
- 一次操作写入多张表、更改状态同时记录历史、或同时发布事件时，使用服务层/application 事务边界。

## 6. API v2、权限与 Swagger

- v2 路由集中在 `internal/routes/v2`，按 `admin`、`client`、`dingtalkh5` 分类。
- 后台 v2 路由必须在管理端路由权限目录中声明，未声明路由应默认拒绝。
- 客户端和钉钉 H5 API 必须与对应应用权限目录保持 method + path + permission key 一致。
- 修改 v2 路由或 DTO 后同步 `internal/routes/v2/swagger/swagger.go`，并生成 `docs/swagger/docs.go`、`swagger.json`、`swagger.yaml`。
- Swagger 生成命令：

  ```bash
  swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger
  ```
- 权限 key 改名必须通过 SQL 迁移同步已有权限和授权数据，不只修改代码常量。

## 7. GORM 与 SQL 迁移

- GORM model 只表达持久化结构和必要的表名/索引元数据，业务流转留在 service/module。
- 主服务启动不执行 AutoMigrate、版本化迁移或种子数据。
- SQL 文件使用 `YYYYMMDDHHMMSS_description.sql` 命名，按文件名顺序执行。
- 迁移必须考虑已存在表/列/索引、历史数据量、默认值、NULL 语义和重复执行。
- 修改表结构时同步 GORM model、查询/映射、Swagger/DTO、前端类型和迁移保护测试。
- 不在没有历史数据分析和迁移计划时直接添加唯一索引或非空约束。
- 创建表时要对应数据中的字符集和排训规则

## 8. 配置、日志与外部边界

- 配置结构在 `internal/config` 集中管理，YAML 与 `WECHECKIN_` 环境变量覆盖语义必须一致。
- 新配置项需要合理默认值、示例配置和非法值校验，不把生产凭证写入默认文件。
- 外部 HTTP、钉钉 OpenAPI、Redis、通知和文件存储使用可替换边界，设置超时并转换外部错误。
- 日志保留请求 ID、业务键和内部错误链，不记录密码、token、完整证件号或其他敏感数据。

## 9. 测试规则

- 包内行为、纯函数、领域状态机和映射测试放在实现附近的 `*_test.go`。
- 路由布局、权限声明、SQL 迁移、启动安全和跨包架构护栏放在 `backend/test/internal/...`。
- 修复缺陷时先添加能稳定复现的回归测试。状态机、幂等和事务变更覆盖成功、重复请求、非法状态和中途失败。
- 测试不依赖公网、生产凭证或不可控时间。需要外部边界时使用局部 fake/store/interface。

## 10. 验证命令

修改 Go 文件后：

```bash
gofmt -w <changed-go-files>
```

先运行受影响包：

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./path/to/package -count=1
```

涉及共享契约、路由、权限、数据库、启动或模块间交互时：

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
```

只修改 Swagger 后至少运行：

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./cmd ./internal/middleware -count=1
```

## 11. 交付检查

1. handler、service/application、domain 和 infrastructure 边界没有被穿透。
2. Context、事务、权限和幂等要求与风险匹配。
3. 路由、DTO、Swagger、权限和调用端契约一致。
4. 表结构变更包含迁移和历史数据考量。
5. Go 文件已格式化，所有实际运行的测试和未验证风险已如实报告。
