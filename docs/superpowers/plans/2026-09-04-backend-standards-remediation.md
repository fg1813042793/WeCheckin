# Backend 规范整改 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 Backend 开发规范消除已确认的配置泄密、运行时 DDL、超时缺失、日志泄密、Context 丢失、原始错误暴露、HTTP 分层和后台通知可靠性问题，并建立可持续执行的质量门禁。

**Architecture:** 整改按“安全基线 -> 进程边界 -> 请求边界 -> 业务边界 -> 工程治理”推进。每个阶段保持现有 API、数据库和流程语义兼容；共享能力先提供 Context/Options 新入口，再由旧入口做兼容包装。数据库结构只通过版本化 SQL 迁移，HTTP 层只返回稳定业务错误，内部错误链进入结构化日志。`taskd` 继续作为独立进程，不启动 HTTP 服务，`role=all` 保持单服务同时运行调度器和执行器。

**Tech Stack:** Go 1.24、Hertz、GORM、MySQL、Redis、Viper、Go 标准库测试、仓库现有迁移与质量检查脚本。

---

## 1. 实施边界与基线

### 1.1 已确认问题

| 编号 | 优先级 | 问题 | 主要影响 |
| --- | --- | --- | --- |
| B-01 | P0 | Git 跟踪配置中存在真实数据库/Redis 地址与口令 | 凭据泄露、环境串用 |
| B-02 | P1 | 权限请求链调用 GORM Migrator 自动补列 | 请求期 DDL、锁表、权限扩大 |
| B-03 | P1 | Server/CORS 配置未完整解析和校验 | Host/超时配置无效、CORS 组合不安全 |
| B-04 | P1 | MySQL 和 OSS 外部调用缺少明确超时 | 网络异常时请求或进程长时间阻塞 |
| B-05 | P1 | GORM Info 日志输出参数，脱敏只覆盖大值 | 小型 token、口令、配置值泄露 |
| B-06 | P1 | 流程办理人、Token 配置和 Setup 写入链路丢失 Context | 客户端取消和服务端超时无法向下传递 |
| B-07 | P1 | HTTP handler 直接返回 `err.Error()` | SQL、内部路径和依赖细节泄露 |
| B-08 | P2 | PostStat 使用裸 goroutine 发送通知 | 错误丢失、进程退出时任务丢失、无法重试 |
| B-09 | P2 | 部分钉钉管理 handler 直接操作 GORM，并使用动态顶层响应 | 分层和稳定 DTO 不符合规范 |
| B-10 | P2 | Backend 存量 Go 文件未全部 gofmt，超大文件职责混合 | 审查困难、门禁无法落地 |
| B-11 | P2 | 权限同步文档仍引用旧目录和 `dingtalk-h5` | 开发指引与当前权威源 `h5app` 不一致 |

### 1.2 必须保持的行为

- 不修改绩效审批、评价、归档和通知的业务规则；钉钉绩效 handler 的整改仅做数据访问下沉与 DTO 固化。
- 不合并 `internal/workflowcore` 与 `internal/modules/workflow`。
- 不改变已发布流程定义和历史流程实例绑定版本的语义。
- 不让主 HTTP 服务执行数据库迁移、AutoMigrate 或权限种子写入。
- 不在 `taskd` 中启动 Hertz/HTTP 服务；`-role scheduler`、`-role worker`、`-role all` 均保持可用。
- 单节点部署使用 `taskd -role all` 时，调度器和执行器必须同时运行；数据库或 Redis 短暂不可用时继续指数退避重连。
- 不手工修改 Swagger 生成产物；只有路由或 DTO 发生变化时才通过规定命令重新生成。
- 当前工作区已有大量未提交改动。每项实施前先确认目标文件差异，只暂存本任务文件，不格式化、回退或覆盖其他改动。

### 1.3 执行批次

1. **第一批，立即处理：** B-01 至 B-07。安全和稳定性问题优先，并在每项结束后跑聚焦测试。
2. **第二批，可靠性与分层：** B-08、B-09。独立提交，使用契约测试保证行为不变。
3. **第三批，工程治理：** B-10、B-11。全量格式化和大文件拆分只在当前工作区形成稳定提交点后执行。

---

## Task 1：清理跟踪配置中的真实凭据

**Files:**

- Modify: `.gitignore`
- Modify: `backend/config/config.yaml`
- Modify: `backend/config/config.prod.yaml`
- Create: `backend/config/config.local.example.yaml`
- Create: `backend/test/internal/config/tracked_config_test.go`
- Modify: `docs/project-maintenance.md`
- Modify: `docs/DEPLOYMENT_TROUBLESHOOTING.md`

**设计决定：** 保留现有 `config.yaml` 和 `config.prod.yaml` 文件名，避免破坏 `LoadConfig`、`start.sh`、`taskd.sh` 和部署参数；跟踪文件只保留安全默认值。真实值统一由 `WECHECKIN_*` 环境变量或被 Git 忽略的 `config.local.yaml` 注入。

- [x] **Step 1：建立配置泄密回归测试**

  在 `backend/test/internal/config/tracked_config_test.go` 解析两个被跟踪的 YAML，至少断言：

  - 所有键名包含 `password`、`secret`、`access_key` 或 `token` 的标量值只能为空或等于 `CHANGE_ME`。
  - `database.host` 和 `redis.host` 只能是 `localhost`、`127.0.0.1`、`0.0.0.0` 或空值。
  - 测试递归遍历 YAML map/list，确保后续新增嵌套凭据也受同一规则保护。

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./test/internal/config -run TestTrackedConfigsDoNotContainSecrets -count=1
  ```

  Expected: FAIL，测试能识别当前跟踪配置中的敏感值。

- [x] **Step 2：替换跟踪配置中的环境实值**

  - 开发默认数据库和 Redis host 使用 `localhost`，密码留空。
  - 生产文件不保存真实主机和密码，使用空值并要求环境变量覆盖。
  - 不把旧值移动到示例、注释或文档。

- [x] **Step 3：增加本地覆盖示例和忽略规则**

  - `.gitignore` 增加 `backend/config/config.local.yaml` 与 `backend/config/config.*.local.yaml`。
  - `config.local.example.yaml` 只列键名和安全占位值。
  - 本地运行可使用 `-env local` 读取 `config.local.yaml`；生产仍使用 `-env prod` 和环境变量。

- [x] **Step 4：记录部署和轮换动作**

  文档明确：

  - 对已经进入 Git 的数据库/Redis/OSS 凭据按“疑似泄露”处理。
  - 先在服务端轮换，再更新部署环境变量，最后验证旧凭据失效。
  - 不在命令示例中写真实口令，使用 shell 环境变量名。

- [x] **Step 5：验证**

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./test/internal/config -count=1
  git diff --check -- .gitignore backend/config docs/project-maintenance.md docs/DEPLOYMENT_TROUBLESHOOTING.md
  ```

  Expected: PASS；`git grep` 不再能从跟踪配置中找到旧凭据。

**外部操作：** 凭据轮换无法由代码测试完成，交付时必须列为运维必做项，不能把“已清理配置文件”等同于“旧凭据已失效”。

---

## Task 2：补齐配置解析、非法值校验和 HTTP Server 配置

**Files:**

- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/config/config.yaml`
- Modify: `backend/config/config.prod.yaml`
- Modify: `backend/cmd/main.go`
- Modify: `backend/cmd/main_safety_test.go`
- Modify: `backend/cmd/taskd/main_test.go`

**设计决定：** 配置在写入全局 `config.Cfg` 之前完成校验。HTTP 主服务使用 Host、端口、读/写/空闲超时；CORS 的凭据开关由配置控制。`taskd` 不消费 HTTP Server 配置，也不新增监听端口。

- [x] **Step 1：为配置失败场景写测试**

  在 `config_test.go` 覆盖：

  - 端口非数字、超出 `1..65535` 时失败。
  - `server.timeout <= 0`，或显式 read/write/idle timeout 小于 0 时失败。
  - `cors.allow_credentials=true` 且 origins 包含 `*` 时失败。
  - 数据库 host/user/dbname 为空，Redis 端口或 DB 越界时失败。
  - 定时任务心跳、TTL、恢复超时不满足既有约束时失败。
  - 环境变量能覆盖 YAML 中的 timeout 和 allow_credentials。
  - 校验失败后 `config.Cfg` 不得指向半成品配置。

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./internal/config -count=1
  ```

  Expected: FAIL，因为当前没有完整字段和 `Validate`。

- [x] **Step 2：增加类型化字段和默认值**

  ```go
  type ServerConfig struct {
      Port            string `mapstructure:"port"`
      Host            string `mapstructure:"host"`
      Mode            string `mapstructure:"mode"`
      TimeoutSec      int    `mapstructure:"timeout"`
      ReadTimeoutSec  int    `mapstructure:"read_timeout_seconds"`
      WriteTimeoutSec int    `mapstructure:"write_timeout_seconds"`
      IdleTimeoutSec  int    `mapstructure:"idle_timeout_seconds"`
  }

  type CORSConfig struct {
      AllowOrigins     []string `mapstructure:"allow_origins"`
      AllowMethods     []string `mapstructure:"allow_methods"`
      AllowHeaders     []string `mapstructure:"allow_headers"`
      AllowCredentials bool     `mapstructure:"allow_credentials"`
  }
  ```

  `server.timeout` 保留为兼容总值，默认 30 秒；未显式配置三个细分值时，分别归一化为 30/60/120 秒。默认 host 改为 `0.0.0.0` 以保持当前 `":"+port` 的监听行为，`config.prod.yaml` 的 mode 固定为 `release`，`allow_credentials=false`。新增对应 `WECHECKIN_SERVER_*` 和 `WECHECKIN_CORS_ALLOW_CREDENTIALS` 绑定。

- [x] **Step 3：实现并调用 `Validate`**

  `LoadConfig` 顺序固定为：defaults -> base YAML -> env YAML -> environment variables -> unmarshal -> `Validate` -> 赋值 `Cfg`。

  错误消息只包含配置键和约束，不回显配置值，尤其不能回显密码、DSN 或 access key。

- [x] **Step 4：让主服务实际使用 Server/CORS 配置**

  `cmd/main.go` 使用：

  ```go
  address := net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
  h := server.Default(
      server.WithHostPorts(address),
      server.WithReadTimeout(time.Duration(cfg.Server.ReadTimeoutSec)*time.Second),
      server.WithWriteTimeout(time.Duration(cfg.Server.WriteTimeoutSec)*time.Second),
      server.WithIdleTimeout(time.Duration(cfg.Server.IdleTimeoutSec)*time.Second),
      server.WithMaxRequestBodySize(32*1024*1024),
  )
  ```

  CORS 使用 `cfg.CORS.AllowCredentials`，禁止再次硬编码 `true`。

- [x] **Step 5：保护 taskd 边界**

  扩充 `cmd/taskd/main_test.go`，继续禁止 `server.Default`、`Spin` 和任何 `WithHostPorts`，并断言 `roleAll` 同时监督 scheduler/worker。

- [x] **Step 6：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/config/config.go internal/config/config_test.go cmd/main.go cmd/main_safety_test.go cmd/taskd/main_test.go
  GOCACHE=$PWD/../.cache/go-build go test ./internal/config ./cmd ./cmd/taskd -count=1
  ```

  Expected: PASS。

---

## Task 3：统一数据库连接超时、连接池和 SQL 日志策略

**Files:**

- Modify: `backend/internal/config/config.go`
- Modify: `backend/config/config.yaml`
- Modify: `backend/config/config.prod.yaml`
- Modify: `backend/pkg/database/database.go`
- Create: `backend/pkg/database/database_test.go`
- Modify: `backend/cmd/main.go`
- Modify: `backend/cmd/taskd/main.go`
- Modify: `backend/cmd/taskd/main_test.go`
- Modify: `backend/cmd/maintenance/main.go`
- Create: `backend/cmd/maintenance/main_test.go`

**设计决定：** 使用 `github.com/go-sql-driver/mysql.Config` 构建 DSN，避免字符串拼接和特殊字符问题。`ConnectDatabase` 接收不依赖 `internal/config` 的 `database.Options`；旧签名暂时保留为兼容包装。主服务、维护进程和 taskd 使用同一组选项，只有 taskd 的连接失败策略继续由 `waitForDependency` 管理。

- [x] **Step 1：写 Options 和 DSN 测试**

  覆盖：

  - 密码含 `@`、`:`、`/` 时 DSN 可正确解析。
  - connect/read/write timeout 写入 MySQL 配置。
  - 连接池 max idle/max open/lifetime/idle time 使用配置值。
  - 生产模式日志为 Warn、`ParameterizedQueries=true`、`Colorful=false`。
  - 开发模式允许 Info，但仍必须参数化，任何参数值不进入 SQL 日志。

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./pkg/database -count=1
  ```

  Expected: FAIL。

- [x] **Step 2：定义数据库连接选项**

  ```go
  type Options struct {
      Host             string
      Port             int
      User             string
      Password         string
      DBName           string
      ConnectTimeout   time.Duration
      ReadTimeout      time.Duration
      WriteTimeout     time.Duration
      MaxIdleConns     int
      MaxOpenConns     int
      ConnMaxLifetime  time.Duration
      ConnMaxIdleTime  time.Duration
      LogLevel         logger.LogLevel
  }
  ```

  `ConnectDatabaseWithOptions` 返回 error，不调用 `log.Fatal`。兼容入口 `InitDatabase` 仅供旧命令过渡，新增入口不得调用它。

- [x] **Step 3：删除基于大小的伪脱敏策略**

  现有 `maxLoggedSQLValueBytes` 只能隐藏大参数，不能保护短 token 和密码。改为 GORM `ParameterizedQueries=true`，日志中保留 SQL 模板、耗时、rows 和错误，不输出参数值。

- [x] **Step 4：三个进程入口统一装配**

  - `cmd/main.go`：连接失败返回/退出，错误中不包含 DSN。
  - `cmd/maintenance/main.go`：使用 Options，连接失败后不继续迁移。
  - `cmd/taskd/main.go`：Options 放入 `waitForDependency` 回调；保留 1 秒起步、30 秒封顶和 jitter 的退避重试。

- [x] **Step 5：验证网络异常契约**

  在 `cmd/taskd/main_test.go` 使用 fake connect 函数验证：连续失败不会让进程提前退出；Context 取消会及时结束；恢复后只返回一次成功。不要在单元测试中访问真实 MySQL。

- [x] **Step 6：验证**

  Run:

  ```bash
  cd backend
  gofmt -w pkg/database/database.go pkg/database/database_test.go cmd/main.go cmd/taskd/main.go cmd/taskd/main_test.go cmd/maintenance/main.go cmd/maintenance/main_test.go internal/config/config.go
  GOCACHE=$PWD/../.cache/go-build go test ./pkg/database ./internal/config ./cmd ./cmd/taskd ./cmd/maintenance -count=1
  ```

  Expected: PASS；测试日志不出现测试口令。

---

## Task 4：移除请求链路中的权限 DDL

**Files:**

- Modify: `backend/internal/support/permission/service.go`
- Modify: `backend/internal/support/permission/service_structure_test.go`
- Modify: `backend/internal/service/admin/adminpermission/service.go`
- Create: `backend/internal/service/admin/adminpermission/service_test.go`
- Verify: `backend/migrations/20260728190000_add_permission_icon.sql`
- Create: `backend/test/internal/bootstrap/migrations/permission_icon_migration_test.go`
- Modify: `docs/PERMISSION_CODE_FRONTEND_SYNC.md`

**设计决定：** `EnsurePermissionSchemaContext` 只检查必要表/列是否就绪，不创建或修改结构。缺失 `permission_icon` 时返回可判断的 `ErrPermissionSchemaNotReady`；恢复方法只有在维护窗口运行 `backend/init.sh` 执行迁移。

- [x] **Step 1：把结构测试改成禁止 DDL**

  断言 `EnsurePermissionSchemaContext` 函数体不含：

  ```text
  AddColumn
  AutoMigrate
  CreateTable
  Exec("ALTER TABLE
  ```

  同时保留 ready cache 优先和 Context 检查断言。

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./internal/support/permission -run 'TestEnsurePermissionSchemaContext' -count=1
  ```

  Expected: FAIL，当前仍调用 `AddColumn`。

- [x] **Step 2：增加迁移保护测试**

  测试确认 `20260728190000_add_permission_icon.sql`：

  - 幂等检查当前 schema。
  - 新增 `permission_icon` 后回填旧菜单图标。
  - 不依赖主服务启动。

- [x] **Step 3：实现只读 schema 检查**

  定义稳定错误：

  ```go
  var ErrPermissionSchemaNotReady = errors.New("permission schema is not ready")
  ```

  错误中允许携带缺失对象名用于内部日志，但 HTTP 响应只提示“系统权限数据尚未初始化，请联系管理员”。

- [x] **Step 4：调整新增/编辑权限服务**

  `adminpermission` 在写入前调用只读检查；缺失 schema 时不做任何业务写入，保证事务原子性。

- [x] **Step 5：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/support/permission/service.go internal/support/permission/service_structure_test.go internal/service/admin/adminpermission/service.go internal/service/admin/adminpermission/service_test.go test/internal/bootstrap/migrations/permission_icon_migration_test.go
  GOCACHE=$PWD/../.cache/go-build go test ./internal/support/permission ./internal/service/admin/adminpermission ./test/internal/bootstrap/migrations -count=1
  ```

  Expected: PASS；`rg 'AddColumn|AutoMigrate' internal/support/permission internal/service/admin/adminpermission` 无运行时代码命中。

---

## Task 5：为 OSS 等外部 HTTP 边界增加 Context 和超时

**Files:**

- Modify: `backend/internal/support/storage/aliyun.go`
- Modify: `backend/internal/support/storage/storage_test.go`
- Verify unchanged: `backend/internal/support/storage/storage.go`
- Verify unchanged: `backend/internal/routes/common/upload_static.go`
- Verify unchanged: `backend/internal/handler/admin/upload/handler.go`
- Verify unchanged: `backend/internal/handler/admin/survey/resource.go`
- Verify unchanged: `backend/internal/handler/admin/exam/resource.go`
- Verify unchanged: `backend/internal/handler/admin/workflow/logo.go`
- Verify unchanged: `backend/internal/handler/dingtalkh5/account/handler.go`
- Verify unchanged: `backend/internal/handler/dingtalkh5/workflowattachment/handler.go`

**设计决定：** 公开上传入口接收请求 Context；Aliyun 上传函数接收可注入 `*http.Client`，默认客户端设置 30 秒总超时。测试使用 fake RoundTripper，不监听端口、不访问公网。

- [x] **Step 1：写取消与超时测试**

  - 请求 Context 已取消时不调用 RoundTripper。
  - RoundTripper 阻塞时，客户端超时后返回 `context deadline exceeded` 可判断错误。
  - 非 2xx 返回外部服务错误，但不回显 Authorization 签名。

- [x] **Step 2：实现 Context 版本**

  保留现有 `saveAliyun(ctx, ...)` 包内入口，在其内部调用可测试的 `saveAliyunWithClient(ctx, client, ...)`。默认 client 固定为 30 秒总超时；现有 `SaveMultipartFile(ctx, ...)` 和八个上传调用方已传递请求 Context，不改它们的对外签名。

- [x] **Step 3：统一外部错误边界**

  内部日志记录 endpoint、bucket、object key 和 wrapped error；HTTP 层只返回稳定上传失败消息，不返回签名、响应体或内部 URL。

- [x] **Step 4：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/support/storage/aliyun.go internal/support/storage/storage_test.go
  GOCACHE=$PWD/../.cache/go-build go test ./internal/support/storage ./internal/routes/common/upload_static ./internal/handler/admin/upload ./internal/handler/admin/survey ./internal/handler/admin/exam ./internal/handler/admin/workflow ./internal/handler/dingtalkh5/account ./internal/handler/dingtalkh5/workflowattachment -count=1
  ```

  Expected: PASS。

---

## Task 6：贯通请求 Context

### Task 6A：流程办理人解析

**Files:**

- Modify: `backend/internal/modules/workflow/domain/runtime.go`
- Modify: `backend/internal/modules/workflow/domain/engine.go`
- Modify: `backend/internal/modules/workflow/domain/engine_test.go`
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/assignee_resolver.go`
- Modify: `backend/internal/modules/workflow/infrastructure/assignee_resolver_test.go`

**设计决定：** 不把 Context 塞进领域数据 DTO。`AssigneeResolver.Resolve` 显式接收 `context.Context`，Engine 中所有可能解析办理人的公开方法显式接收 Context，由 application 已有请求 Context 传入。

- [x] **Step 1：新增 resolver 取消测试**

  fake resolver 收到取消 Context 时返回 `context.Canceled`；断言流程 Start/Complete 不创建任务、不追加后续历史、不提交 store 事务。

- [x] **Step 2：修改稳定接口**

  ```go
  type AssigneeResolver interface {
      Resolve(context.Context, AssigneeRequest) ([]string, error)
  }
  ```

  `Engine.Start`、`Engine.Complete` 以及内部节点推进链统一接收同一个 ctx。纯状态方法 `Cancel`、`Withdraw` 若不访问外部 resolver，可保持原签名。

- [x] **Step 3：删除 resolver 内部 `context.Background()`**

  manager、department leader、role 和 org identity 查询均使用调用方 ctx；`ResolveDisplayNames` 复用同一 Context 版本，不二次丢失。

- [x] **Step 4：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/modules/workflow/domain internal/modules/workflow/application internal/modules/workflow/infrastructure/assignee_resolver.go internal/modules/workflow/infrastructure/assignee_resolver_test.go
  GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/... -count=1
  ```

  Expected: PASS；`assignee_resolver.go` 中不再出现 `context.Background()`。

### Task 6B：Token 配置和 Setup 写入

**Files:**

- Modify: `backend/pkg/tokenutil/tokenutil.go`
- Modify: `backend/pkg/tokenutil/tokenutil_test.go`
- Modify: `backend/internal/middleware/admin/auth.go`
- Modify: `backend/internal/middleware/client/auth.go`
- Modify: `backend/internal/handler/client/passport/handler.go`
- Modify: `backend/internal/handler/client/survey/submit.go`
- Modify: `backend/internal/handler/admin/dingtalk/handler.go`
- Modify: `backend/internal/service/client/passport/login.go`
- Modify: `backend/internal/service/admin/adminauth/service.go`
- Modify: `backend/internal/service/admin/online/admin.go`
- Modify: `backend/internal/service/admin/online/user.go`
- Modify: `backend/internal/service/admin/setup/service.go`
- Create: `backend/internal/service/admin/setup/service_test.go`
- Modify: `backend/internal/handler/admin/setup/handler.go`
- Modify: `backend/internal/handler/admin/setup/content_setup_structure_test.go`

- [x] **Step 1：新增 Context API 测试**

  覆盖 `GetTokenConfigContext`、`IsAdminSingleLoginContext`、`IsUserSingleLoginContext`、`IsDingTalkH5SingleLoginContext`；取消 Context 时不查询数据库，缓存命中时不访问数据库。

- [x] **Step 2：兼容旧入口**

  新请求链只调用 Context 版本；旧入口保留，并使用 `context.Background()` 调用新入口，防止一次性修改所有非请求调用方。

- [x] **Step 3：Setup handler 使用服务层 Context 版本**

  `SetSetupContext(ctx, ...)` 和 `SetContentSetupContext(ctx, ...)` 在成功提交后再失效 token 配置缓存。失败时不能清缓存或产生部分写入。

- [x] **Step 4：验证**

  Run:

  ```bash
  cd backend
  gofmt -w pkg/tokenutil internal/service/admin/setup internal/handler/admin/setup
  GOCACHE=$PWD/../.cache/go-build go test ./pkg/tokenutil ./internal/service/admin/setup ./internal/handler/admin/setup ./internal/middleware/... -count=1
  ```

  Expected: PASS。

---

## Task 7：建立稳定 HTTP 错误映射，禁止回显内部错误

**Files:**

- Create: `backend/pkg/response/error.go`
- Create: `backend/pkg/response/error_test.go`
- Create: `backend/internal/modules/workflow/transport/httperror/error.go`
- Create: `backend/internal/modules/workflow/transport/httperror/error_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/handler.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/summary_handler.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/handler_test.go`
- Create: `backend/test/internal/httpboundary/raw_error_response_test.go`
- Modify: 其余被 `raw_error_response_test.go` 精确列出的生产 handler 文件

**设计决定：** `response` 提供“记录内部错误并返回固定消息”的通用能力；业务 sentinel 到用户消息/状态码的映射留在各 transport，避免 `pkg/response` 反向依赖业务模块。未知错误一律返回通用消息，完整 wrapped error 只进入日志。

- [x] **Step 1：写通用响应安全测试**

  构造错误：

  ```text
  Error 1064: syntax error near password='secret'
  ```

  断言 HTTP body 不包含 SQL、`secret`、本地路径和原始 error；测试 logger 能收到请求 ID、操作名和完整错误链。

- [x] **Step 2：实现响应帮助函数**

  接口固定为：

  ```go
  func FailInternal(ctx context.Context, c *app.RequestContext, operation, publicMessage string, err error)
  ```

  `operation` 使用稳定常量，如 `workflow.start`，不得拼接请求体或 token。请求 ID 从 `X-Request-ID`/`X-Trace-ID` header 读取，与现有 AccessLog 规则保持一致。

- [x] **Step 3：实现流程错误映射**

  用 `errors.Is` 映射已有 domain/application sentinel：参数/状态冲突返回可操作业务消息；无权限返回统一权限消息；store/GORM/未知错误返回“流程操作失败，请稍后重试”。不得使用字符串包含判断决定错误类型。

- [x] **Step 4：先迁移 workflow transport**

  替换 `httpadmin`、`httpclient` 和 `summary_handler` 中所有 `response.Fail(c, err.Error())`。保持当前成功响应 DTO、分页和 HTTP status 不变。

- [x] **Step 5：增加仓库级边界保护**

  `raw_error_response_test.go` 扫描生产 handler/transport，禁止以下直接回显形式：

  ```text
  response.Fail(c, err.Error())
  response.JSON(c, err.Error())
  string(err)
  ```

  对确实需要向用户展示的业务校验错误，必须先映射为稳定类型或显式公开消息。

- [x] **Step 6：分包迁移剩余命中**

  按“钉钉 H5 -> admin -> client -> survey/exam/enroll”顺序，每一包单独运行测试。不要在一个提交中批量改变成功响应或 DTO。

- [x] **Step 7：验证**

  Run:

  ```bash
  cd backend
  gofmt -w pkg/response/error.go pkg/response/error_test.go internal/modules/workflow/transport test/internal/httpboundary
  GOCACHE=$PWD/../.cache/go-build go test ./pkg/response ./internal/modules/workflow/transport/... ./test/internal/httpboundary -count=1
  GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
  ```

  Expected: PASS；仓库级保护测试无生产命中。

---

## Task 8：把 PostStat 裸 goroutine 改为可靠通知投递

**Files:**

- Modify: `backend/internal/service/client/poststat/service.go`
- Modify: `backend/internal/service/client/poststat/notify.go`
- Create: `backend/internal/service/client/poststat/service_test.go`
- Create: `backend/internal/service/client/poststat/dispatcher.go`
- Create: `backend/internal/service/client/poststat/dispatcher_test.go`
- Create: `backend/internal/model/notification/outbox.go`
- Create: `backend/internal/modules/notificationoutbox/application/service.go`
- Create: `backend/internal/modules/notificationoutbox/application/service_test.go`
- Create: `backend/internal/modules/notificationoutbox/infrastructure/gorm_store.go`
- Create: `backend/internal/modules/notificationoutbox/infrastructure/gorm_store_test.go`
- Create: `backend/internal/modules/notificationoutbox/infrastructure/channels.go`
- Create: `backend/internal/modules/notificationoutbox/infrastructure/channels_test.go`
- Create: `backend/internal/support/outboundhttp/client.go`
- Create: `backend/internal/support/outboundhttp/client_test.go`
- Modify: `backend/internal/modules/scheduledtask/infrastructure/handler_http.go`
- Modify: `backend/internal/modules/scheduledtask/infrastructure/handler_http_test.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/notification_outbox_job.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/notification_outbox_job_test.go`
- Modify: `backend/internal/modules/scheduledtask/application/service.go`
- Modify: `backend/internal/modules/scheduledtask/application/service_test.go`
- Modify: `backend/internal/modules/scheduledtask/application/management.go`
- Modify: `backend/internal/modules/scheduledtask/application/management_test.go`
- Modify: `backend/cmd/taskd/main.go`
- Modify: `backend/cmd/taskd/main_test.go`
- Create: `backend/migrations/20260904113000_create_notification_outbox.sql`
- Create: `backend/test/internal/bootstrap/migrations/notification_outbox_migration_test.go`

**设计决定：** 复用当前通用通知/定时任务投递能力，不再新建第二套进程内队列。问卷提交事务只保存“需要投递”的事实；`taskd` worker 异步投递站内通知和 webhook，使用幂等键防止重复发送。

- [x] **Step 1：先冻结现有通知行为**

  测试覆盖 channel=`webhook`、`internal`、`both`，模板占位符、收件人、无 webhook URL、重复消费和投递失败。

- [x] **Step 2：定义 Dispatcher 边界**

  ```go
  type NotificationDispatcher interface {
      Enqueue(ctx context.Context, request NotificationRequest) error
  }
  ```

  `ProcessContext` 只依赖接口，不直接启动 goroutine。Request 使用命名字段，幂等键由 survey ID、answer ID、rule ID 和 channel 组成。

- [x] **Step 3：适配通用通知 outbox**

  - 新建通用 `notification_outbox`，字段至少包含 source type/id、channel、recipient、payload、status、attempts、next retry time、last error 和唯一 idempotency key。
  - 站内通知 channel 调用现有 `inappnotification.ApplicationService.Send`，不复制收件人解析和 `notify` 表写入逻辑。
  - 先把 `handler_http.go` 的 host/CIDR/重定向/请求响应大小限制和安全 dialer 下沉为 `internal/support/outboundhttp`；现有 scheduled task HTTP handler 与 webhook channel 共用该包。抽取前后的 scheduled task HTTP 测试必须完全一致。
  - webhook channel 使用共享 outbound client，不直接调用 `http.DefaultClient`。
  - `notification.outbox.dispatch_due` 注册为 taskd 内置 Go job；单节点 `role=all` 和拆分 scheduler/worker 部署都能执行。
  - `20260904113000_create_notification_outbox.sql` 同时幂等写入 code=`system.notification-outbox-dispatch`、minute cron、enabled=1 的系统任务，首次 `next_run_at` 设为迁移执行时刻。原计划时间戳 `20260904110000` 已被权限重命名迁移占用，因此使用未占用的 `20260904113000`。
  - scheduled task application 将 `system.` 前缀任务标记为只读：允许查看运行记录，禁止通过管理 API 禁用、编辑或删除，确保 Outbox 不会因误操作停止派发。
  - 投递错误写入 outbox `last_error` 和任务运行日志，按既有退避策略重试，达到上限后进入 dead 状态。

- [x] **Step 4：删除 fire-and-forget 调用**

  `poststat` 请求链不再出现 `go sendWebhook` 或 `go sendInternalNotificationContext`。入队失败应让业务层返回可观测错误；是否回滚问卷提交由既有事务边界明确决定并写入测试，不能静默丢通知。

- [x] **Step 5：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/service/client/poststat internal/model/notification internal/modules/notificationoutbox internal/support/outboundhttp internal/modules/scheduledtask/infrastructure/handler_http.go internal/modules/scheduledtask/infrastructure/handler_http_test.go internal/modules/scheduledtask/infrastructure/notification_outbox_job.go internal/modules/scheduledtask/infrastructure/notification_outbox_job_test.go internal/modules/scheduledtask/application cmd/taskd/main.go cmd/taskd/main_test.go test/internal/bootstrap/migrations/notification_outbox_migration_test.go
  GOCACHE=$PWD/../.cache/go-build go test ./internal/service/client/poststat ./internal/modules/notificationoutbox/... ./internal/modules/inappnotification/... ./internal/support/outboundhttp ./internal/modules/scheduledtask/... ./cmd/taskd ./test/internal/bootstrap/migrations -count=1
  ```

  Expected: PASS；`rg 'go (sendWebhook|sendInternalNotification)' internal/service/client/poststat` 无命中。

---

## Task 9：钉钉后台 handler 分层与命名 DTO

**Files:**

- Modify: `backend/internal/handler/admin/dingtalk/performance.go`
- Modify: `backend/internal/handler/admin/dingtalk/bindings.go`
- Modify: `backend/internal/handler/admin/dingtalk/handler.go`
- Modify: `backend/internal/handler/admin/dingtalk/perf_admin_structure_test.go`
- Modify: `backend/internal/handler/admin/dingtalk/bindings_structure_test.go`
- Modify: `backend/internal/handler/admin/dingtalk/settings_test.go`
- Modify: `backend/internal/handler/admin/setup/handler.go`
- Modify: `backend/internal/handler/admin/setup/content_setup_structure_test.go`
- Modify: `backend/internal/handler/admin/survey/formkit_tools.go`
- Create: `backend/internal/handler/admin/survey/formkit_tools_test.go`
- Create: `backend/internal/service/admin/dingtalk/performance.go`
- Create: `backend/internal/service/admin/dingtalk/performance_test.go`
- Create: `backend/internal/service/admin/dingtalk/bindings.go`
- Create: `backend/internal/service/admin/dingtalk/bindings_test.go`
- Create: `backend/internal/service/admin/dingtalk/dto.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Generate: `backend/docs/swagger/docs.go`
- Generate: `backend/docs/swagger/swagger.json`
- Generate: `backend/docs/swagger/swagger.yaml`

**设计决定：** 这是纯分层重构。保留原查询条件、排序、分页、事务、状态枚举、权限和 JSON 字段名；不修改任何绩效流程状态转换或通知规则。

- [x] **Step 1：为现有 HTTP JSON 建立契约测试**

  对列表、详情、绑定/解绑和绩效管理操作记录当前 status、body 顶层字段、空值和分页语义。测试先通过，作为重构前基线。

- [x] **Step 2：为 handler 建立结构保护**

  禁止 handler 文件直接引用：

  ```text
  database.GetDB
  database.WithContext
  .Transaction(
  gorm.
  ```

  handler 仅解析输入、调用 service、映射错误和输出 DTO。

- [x] **Step 3：移动查询和事务到 service**

  service Context 入口负责数据访问和事务。重复查询条件提取为私有函数，不改变 SQL 的 Where/Order/Limit 语义。

- [x] **Step 4：用命名 DTO 替换顶层动态 map**

  DTO 使用明确 JSON tag；动态绩效表单内容可继续使用 `json.RawMessage` 或 `map[string]any` 作为嵌套字段，但列表/详情顶层必须命名。同时替换钉钉设置/企业列表、Setup token debug 和 FormKit 工具的顶层 map；FormKit 的 `value`、`answers`、`states` 等动态嵌套值保持原类型和 JSON 结构。

- [x] **Step 5：对比重构前后契约**

  同一 fixture 下比较 JSON 结构；任何字段增删或空值变化都视为失败，除非另立 API 变更任务。

  更新 Swagger 源声明引用命名 DTO，然后执行：

  ```bash
  cd backend
  swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger
  ```

- [x] **Step 6：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/handler/admin/dingtalk internal/service/admin/dingtalk internal/handler/admin/setup internal/handler/admin/survey/formkit_tools.go internal/handler/admin/survey/formkit_tools_test.go internal/routes/v2/swagger/swagger.go
  GOCACHE=$PWD/../.cache/go-build go test ./internal/handler/admin/dingtalk ./internal/service/admin/dingtalk ./internal/handler/admin/setup ./internal/handler/admin/survey ./internal/routes/v2/swagger -count=1
  GOCACHE=$PWD/../.cache/go-build go test ./internal/service/dingtalkh5/performance/... -count=1
  ```

  Expected: PASS；绩效领域包无行为性改动。

---

## Task 10：修正文档路径和运行手册

**Files:**

- Modify: `docs/PERMISSION_CODE_FRONTEND_SYNC.md`
- Modify: `docs/API_V2.md`
- Modify: `docs/project-maintenance.md`
- Modify: `docs/SCHEDULED_TASKS.md`
- Modify: `backend/docs/development-guidelines.md`
- Create: `backend/test/internal/documentation/current_paths_test.go`

- [x] **Step 1：增加旧路径保护测试**

  禁止规范文档引用：

  - 已删除的 `backend/internal/app/support/...`。
  - 旧钉钉目录 `dingtalk-h5/...`。

  要求权限同步文档明确 `backend/internal/support/...` 和当前权威源 `h5app/...`。

- [x] **Step 2：同步本轮新约束**

  文档增加：

  - 配置校验与环境变量注入。
  - 主服务、maintenance、taskd 的数据库连接差异。
  - `taskd -role all` 的单节点运行方式和断线恢复语义。
  - 请求边界错误映射、SQL 参数化日志和外部 HTTP 超时。
  - schema 缺失时通过 `init.sh` 迁移，服务请求不自动补结构。

- [x] **Step 3：验证**

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./test/internal/documentation -count=1
  cd ..
  node scripts/check-quality-gates.mjs
  git diff --check -- docs backend/docs
  ```

  Expected: PASS。

---

## Task 11：建立 gofmt 门禁并分批拆解超大文件

### Task 11A：gofmt 基线与门禁

**前置条件：** 当前工作区的 Backend 修改已由用户形成稳定提交点。未满足前置条件时不得执行全目录 `gofmt -w`，以免把业务修改和机械格式化混在一起。

**Files:**

- Mechanical modify: `backend/**/*.go` 中 `gofmt -l` 的存量命中文件
- Create: `scripts/check-backend-format.sh`
- Modify: `scripts/verify-local.sh`
- Modify: `scripts/check-quality-gates.mjs`
- Modify: `docs/project-maintenance.md`

- [x] **Step 1：记录格式化前基线**

  Run:

  ```bash
  cd backend
  gofmt -l cmd internal pkg test
  ```

  将输出保存到任务记录，不写入仓库。确认这些文件没有跨任务未提交改动。

- [ ] **Step 2：独立机械格式化**

  只对上一步列表执行 `gofmt -w`，不同时修改逻辑。格式化后运行 Backend 全量测试并单独审查 diff。

- [ ] **Step 3：增加门禁**

  `check-backend-format.sh` 在 `gofmt -l` 有输出时打印文件列表并退出 1。`verify-local.sh` 在 Backend 测试前运行该脚本；质量门禁测试必须断言此调用存在。

- [ ] **Step 4：验证**

  Run:

  ```bash
  bash scripts/check-backend-format.sh
  node scripts/check-quality-gates.mjs
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
  ```

  Expected: PASS；`gofmt -l cmd internal pkg test` 无输出。

### Task 11B：按职责拆解超大文件

**Files:**

- Split: `backend/internal/support/permission/service.go`
- Split: `backend/internal/modules/workflow/application/service.go`
- Split: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Create: `backend/test/internal/architecture/file_size_budget_test.go`

**设计决定：** 保持 package、导出 API 和调用方不变，只在同包内按职责移动代码。先处理前三个最影响审查的文件，不一次性重写全部 300 行以上文件。

- [x] **Step 1：建立行为和架构基线**

  先运行三个包的全量测试，记录导出符号：

  ```bash
  cd backend
  go doc ./internal/support/permission
  go doc ./internal/modules/workflow/application
  go doc ./internal/modules/workflow/infrastructure
  ```

- [ ] **Step 2：拆分 permission service**

  同包拆为 schema、query、evaluate、sync、cache 等职责文件。禁止趁机修改权限判定优先级、种子数据或缓存失效语义。

- [ ] **Step 3：拆分 workflow application service**

  同包拆为 definition、start、task、instance、notification、quota 等职责文件。保留现有 transaction 闭包和幂等边界。

- [ ] **Step 4：拆分 GORM store**

  同包拆为 definition store、runtime state、query、participant、notification、quota。SQL 条件、锁和排序必须由现有测试逐项保护。

- [ ] **Step 5：增加渐进式文件预算**

  门禁只限制本轮拆分后的三个包，单文件上限设为 700 行；其他存量大文件先记录基线，不因一次整改阻塞全仓库。

- [ ] **Step 6：验证**

  Run:

  ```bash
  cd backend
  gofmt -w internal/support/permission internal/modules/workflow/application internal/modules/workflow/infrastructure test/internal/architecture
  GOCACHE=$PWD/../.cache/go-build go test ./internal/support/permission ./internal/modules/workflow/... ./test/internal/architecture -count=1
  GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
  ```

  Expected: PASS；导出 API 清单和整改前一致。

---

## Task 12：最终全量验证与发布检查

**Files:**

- Modify only when needed: `docs/project-maintenance.md`
- No generated file changes unless Task 9 actually changes Swagger declarations

- [x] **Step 1：静态安全复核**

  Run:

  ```bash
  cd backend
  rg -n 'response\.(Fail|JSON)\([^\n]*err\.Error\(\)' internal --glob '*.go' --glob '!**/*_test.go'
  rg -n 'AddColumn|AutoMigrate|CreateTable' internal --glob '*.go' --glob '!**/*_test.go'
  rg -n 'http\.DefaultClient' internal --glob '*.go' --glob '!**/*_test.go'
  rg -n 'go (sendWebhook|sendInternalNotification)' internal/service/client/poststat --glob '*.go'
  ```

  Expected: 四条命令无违规命中。迁移/维护命令中的显式建表不属于请求期 DDL，但必须由人工确认目录边界。

- [x] **Step 2：Backend 全量测试**

  Run:

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
  ```

  Expected: PASS。

- [x] **Step 3：根级质量检查**

  Run:

  ```bash
  cd ..
  bash scripts/verify-local.sh
  git diff --check
  ```

  Expected: PASS。该脚本覆盖 Backend、Admin、Frontend；本计划没有修改 H5App，不要求额外运行 H5App 检查。

- [ ] **Step 4：受控环境集成验证**

  使用非生产 MySQL/Redis 验证：

  1. 错误密码时 HTTP 主服务在连接超时内退出，不永久挂起。
  2. taskd 在 MySQL/Redis 暂时断开时不退出，恢复后继续调度和执行。
  3. `taskd -role all` 单节点能创建、领取并完成一次定时任务。
  4. 权限列缺失时主服务不执行 DDL，并给出迁移提示；运行 `init.sh` 后恢复。
  5. CORS wildcard + credentials 的非法组合在启动前被拒绝。
  6. SQL 日志不包含登录 token、配置值或测试口令。

- [ ] **Step 5：交付记录**

  交付说明逐项列出：

  - 已运行命令及结果。
  - 凭据是否已由运维完成轮换。
  - 是否执行了真实 MySQL 迁移和 taskd 断网恢复测试。
  - 未验证的浏览器、设备或生产网络风险。
  - 本批次实际修改文件，确认未夹带当前工作区其他改动。

---

## 建议提交边界

当前工作区不是干净状态，实施时只能按路径精确暂存。以下提交只是建议边界，未获得用户明确要求时不自动提交：

1. `security: remove tracked backend credentials`
2. `fix: validate backend runtime configuration`
3. `fix: add database timeouts and parameterized logging`
4. `fix: remove request-time permission ddl`
5. `fix: propagate backend request context`
6. `fix: sanitize backend http errors`
7. `fix: enqueue poststat notifications reliably`
8. `refactor: move dingtalk admin data access to services`
9. `docs: refresh backend paths and operations`
10. `chore: enforce backend gofmt`
11. `refactor: split oversized backend services`

每次暂存前执行：

```bash
git diff -- <本任务文件列表>
git status --short
```

禁止使用 `git add -A`、`git add .` 或任何会把现有无关改动一起纳入的命令。
