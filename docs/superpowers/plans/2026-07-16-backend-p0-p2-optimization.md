# 后端 P0-P2 优化执行计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标：** 完成后台权限、访问日志、响应 DTO、服务测试、DB/Redis context、权限缓存与审计的 P0-P2 优化。

**架构：** 优先修复安全风险，再补稳定性和可维护性。高风险中间件加行为测试，高频后台接口用 DTO 固定响应契约，DB/Redis 操作统一通过 context helper 获得超时与取消链路。

**技术栈：** Go、Hertz、GORM、go-redis、标准库 testing。

---

### Task 1: P0 访问日志安全

**文件：**
- 修改：`backend/internal/middleware/access_log.go`
- 测试：`backend/internal/middleware/access_log_test.go`

- [x] 增加富文本、答卷、schema、description 等字段脱敏。
- [x] 增加上传和大 body 跳过策略。
- [x] 修复请求方法大小写导致日志详情不稳定的问题。
- [x] 覆盖表单、JSON、multipart、大 body、GET query 脱敏测试。

### Task 2: P1 响应 DTO 类型化

**文件：**
- 新增：`backend/internal/app/handler/admin/survey/dto.go`
- 新增：`backend/internal/app/handler/admin/exam/dto.go`
- 新增：`backend/internal/app/handler/admin/user/dto.go`
- 修改：`backend/internal/app/handler/handler_structure_test.go`
- 修改：`backend/internal/app/handler/admin/survey/handler.go`
- 修改：`backend/internal/app/handler/admin/survey/responses.go`
- 修改：`backend/internal/app/handler/admin/exam/handler.go`
- 修改：`backend/internal/app/handler/admin/exam/records.go`
- 修改：`backend/internal/app/handler/admin/user/handler.go`

- [x] 问卷列表、详情、答卷列表、答卷详情使用 DTO。
- [x] 考试列表、详情、保存返回、记录列表、记录详情使用 DTO。
- [x] 用户列表使用 DTO，列表行从 service map 改成结构体。
- [x] 增加结构测试，防止高频后台响应重新使用裸 `map[string]interface{}`。

### Task 3: P2 服务层测试

**文件：**
- 新增：`backend/internal/app/service/admincontent/enroll_test.go`
- 新增：`backend/internal/app/service/enroll/fields_test.go`
- 新增：`backend/internal/app/service/event/helpers_test.go`
- 新增：`backend/internal/app/service/exam/service_test.go`

- [x] admincontent/enroll 覆盖报名扩展对象解析。
- [x] enroll 覆盖客户端报名扩展对象解析。
- [x] event 覆盖活动扩展对象解析与时间状态展示。
- [x] exam 覆盖创建默认值和记录详情解析。

### Task 4: P2 DB/Redis Context

**文件：**
- 新增：`backend/pkg/database/context.go`
- 新增：`backend/pkg/database/context_test.go`
- 新增：`backend/pkg/redis/context.go`
- 新增：`backend/pkg/redis/context_test.go`

- [x] 数据库查询统一提供默认超时 context helper。
- [x] Redis 操作统一提供默认超时 context helper。
- [x] 后台问卷、考试、用户列表和权限路径迁移到 context-aware 调用。
- [x] 登录、认证、在线会话、客户端问卷/考试会话 Redis 调用移除全局 `rd.Ctx`。

### Task 5: P2 权限缓存与审计

**文件：**
- 新增：`backend/internal/app/service/menu/cache.go`
- 新增：`backend/internal/app/service/menu/cache_test.go`
- 修改：`backend/internal/app/service/menu/service.go`
- 修改：`backend/internal/app/service/role/service.go`
- 修改：`backend/internal/middleware/admin_permission.go`
- 修改：`backend/internal/middleware/admin_permission_test.go`

- [x] 角色权限码按角色缓存，并返回防御性副本。
- [x] 菜单新增、编辑、删除和角色菜单变更时失效缓存。
- [x] 权限拒绝写专门审计日志。
- [x] 审计日志兼容测试环境未初始化 logger 的情况。

### Verification

- [x] `GOCACHE=$PWD/.cache/go-build go test ./backend/...`
- [x] `git diff --check`
