# 工作流发布发起部门范围实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让管理员发布流程时可多选允许发起的部门，并按指定部门与指定用户的并集动态控制流程列表、详情和正式发起权限。

**Architecture:** 扩展开始节点 `InitiatorConfig`，将 `departmentIds` 随发布版本保存；应用服务通过存储接口一次读取发起人的实时 `user_depts`，再使用纯函数执行用户与部门并集判断。发布弹窗使用严格模式多选部门树，不级联子部门，并保留现有指定用户选择器。

**Tech Stack:** Go 1.24.5、Gorm、Vue 3、TypeScript、Element Plus、Node.js 回归脚本

---

## 任务 1：扩展核心发起范围模型与定义校验

**文件：**
- 修改：`backend/internal/workflow/types.go`
- 修改：`backend/internal/workflow/validation.go`
- 修改：`backend/internal/workflow/definition_test.go`

- [ ] 在 `definition_test.go` 增加表驱动测试：`specified` 只含用户、只含部门、同时含两者均通过；空范围、零部门 ID、重复部门 ID、`all` 携带部门 ID 均返回 `ValidationInitiator`。
- [ ] 运行测试并确认因 `DepartmentIDs` 尚不存在而编译失败：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -run 'TestValidateDefinition.*Initiator' -count=1
```

- [ ] 为 `InitiatorConfig` 增加：

```go
DepartmentIDs []uint `json:"departmentIds,omitempty"`
```

- [ ] 重构 `validateInitiator`：`all` 要求两类 ID 都为空；`specified` 分别校验正整数和列表内唯一，并要求两类至少一类非空。
- [ ] 运行定向测试并确认通过。

## 任务 2：发布服务保留并防御性复制部门范围

**文件：**
- 修改：`backend/internal/service/admin/workflow/service.go`
- 修改：`backend/internal/service/admin/workflow/service_test.go`

- [ ] 先扩展发布服务测试，要求草稿配置保留 `DepartmentIDs`，请求覆盖时两类切片均不受调用方后续修改影响。
- [ ] 增加旧顶层 `initiator` 迁移用例，确保 `departmentIds` 一并迁移到开始节点。
- [ ] 运行定向测试并确认失败：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/service/admin/workflow -run 'TestApplyPublishInitiator|TestNormalizeDraftMigratesLegacyTopLevelInitiator' -count=1
```

- [ ] 在 `applyPublishInitiator` 和旧配置迁移中分别复制 `UserIDs` 与 `DepartmentIDs`。
- [ ] 运行定向测试并确认通过。

## 任务 3：应用服务按实时部门执行并集鉴权

**文件：**
- 修改：`backend/internal/modules/workflow/application/service.go`
- 修改：`backend/internal/modules/workflow/application/service_test.go`
- 修改：`backend/internal/modules/workflow/application/idempotency_test.go`（仅在接口变化要求时补齐测试存储）

- [ ] 扩展 `fakeStore`，保存 `userDepartmentIDs []uint` 和查询次数，并实现 `UserDepartmentIDs(context.Context, string) ([]uint, error)`。
- [ ] 先增加纯判断测试，覆盖：全部用户、明确用户命中、任一部门命中、多部门未命中、无部门未命中。
- [ ] 增加列表与详情测试，断言一次读取用户部门，并只返回/允许匹配定义。
- [ ] 增加正式发起测试，断言部门命中可发起、部门未命中返回 `ErrStarterNotAllowed`；管理员代发使用业务发起人的部门。
- [ ] 运行定向测试并确认失败：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/application -run 'TestPublishedInitiator|TestListPublishedDefinitionsForStarter|TestGetPublishedDefinitionForStarter|TestStartInstance.*Department' -count=1
```

- [ ] 为 `Store` 和 `TransactionStore` 增加 `UserDepartmentIDs`。
- [ ] 将纯判断函数调整为：

```go
func publishedInitiatorAllows(initiator workflowcore.InitiatorConfig, starterID string, departmentIDs []uint) bool
```

- [ ] `ListPublishedDefinitionsForStarter` 每次请求读取一次部门，再过滤全部定义；`GetPublishedDefinitionForStarter` 读取一次后判断。
- [ ] 正式发起在事务内读取目标发起人的实时部门，对实际发布版本再次判断。
- [ ] 运行定向测试并确认通过。

## 任务 4：Gorm 存储读取部门并完整映射发布配置

**文件：**
- 修改：`backend/internal/modules/workflow/infrastructure/gorm_store.go`
- 修改：`backend/internal/modules/workflow/infrastructure/mapping_test.go`
- 新增：`backend/internal/modules/workflow/infrastructure/initiator_scope_structure_test.go`
- 修改：`backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- 修改：`backend/internal/modules/workflow/transport/httpclient/handler_test.go`

- [ ] 先增加映射测试，要求 `definitionInitiator` 返回独立复制的 `UserIDs` 和 `DepartmentIDs`。
- [ ] 增加结构测试，要求 `UserDepartmentIDs` 从 `user_depts` 按用户 ID 一次 `Pluck` 全部部门，并归一化去重。
- [ ] 扩展管理端和客户端响应断言，要求发布定义 JSON 包含 `"departmentIds":[...]`。
- [ ] 运行基础设施和传输层定向测试并确认失败：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/infrastructure ./internal/modules/workflow/transport/httpadmin ./internal/modules/workflow/transport/httpclient -run 'Test.*Initiator|TestListDefinitions' -count=1
```

- [ ] 实现 `GormStore.UserDepartmentIDs`，解析正整数用户 ID，从 `user_depts` 查询 `user_dept_dept_id`，返回排序去重后的 ID。
- [ ] 在 `definitionInitiator` 中防御性复制 `DepartmentIDs`。
- [ ] 更新测试桩的发布定义配置并运行定向测试通过。

## 任务 5：发布弹窗支持严格模式多选部门

**文件：**
- 修改：`admin/src/api/index.ts`
- 修改：`admin/src/views/workflow/types.ts`
- 修改：`admin/src/views/workflow/components/WorkflowPublishDialog.vue`
- 修改：`admin/scripts/check-workflow-runtime-pages.mjs`

- [ ] 先扩展结构检查，要求 API 和页面类型包含 `departmentIds`，发布弹窗包含多选 `el-tree-select`、`check-strictly`、部门回显、空并集提示与提交字段。
- [ ] 运行检查并确认失败：

```bash
cd admin
npm run check:workflow-runtime-pages
```

- [ ] 将前端发起配置扩展为 `departmentIds?: number[]`。
- [ ] 将发布表单改为：

```ts
const publishForm = reactive<{
  scope: WorkflowInitiatorScope
  userIds: number[]
  departmentIds: number[]
}>({ scope: 'all', userIds: [], departmentIds: [] })
```

- [ ] 将“指定用户”改为“指定范围”；增加 `multiple`、`check-strictly`、`node-key="id"` 的部门树选择器，确保父子部门不级联。
- [ ] 将用户字段改名为“额外允许用户”，允许只选用户或部门。
- [ ] 回显时分别归一化两类 ID；指定范围两类都为空时提示“请选择允许发起的部门或用户”。
- [ ] 发布时提交两类去重 ID；全部用户只提交 `{ scope: 'all' }`。
- [ ] 运行 `npm run check:workflow-runtime-pages` 并确认通过。

## 任务 6：文档与完整回归

**文件：**
- 修改：`docs/architecture/generic-oa-workflow-v1.md`
- 参考：`docs/superpowers/specs/2026-09-01-workflow-initiator-department-scope-design.md`

- [ ] 在架构文档中说明发布发起范围支持指定用户与多个部门并集、仅匹配所选部门本级、部门成员实时读取。
- [ ] 格式化本次 Go 文件。
- [ ] 运行后端完整相关回归：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -count=1
GOCACHE=$PWD/.cache/go-build go test ./internal/service/admin/workflow ./internal/modules/workflow/... -count=1
```

- [ ] 运行前端回归和生产构建：

```bash
cd admin
npm run check:workflow-runtime-pages
npm run check:workflow-form-designer
npm run check:workflow-runtime-form
npm run check:workflow-tree
npm run build
```

- [ ] 对本次文件执行 `git diff --check` 和行尾空白扫描。
- [ ] 确认本地 `http://127.0.0.1:5173/` 可访问；页面最终手工测试由用户执行。
- [ ] 当前工作区存在大量未提交改动，本次不创建提交、不回滚其他修改。
