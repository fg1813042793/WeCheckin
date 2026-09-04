# 工作流发起范围统一组织树与排除用户 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将指定发起部门和用户合并为一个组织树，并让全部用户、指定范围两种模式都支持最高优先级的排除用户规则。

**Architecture:** 在 `workflowcore.InitiatorConfig` 中增加可选 `excludedUserIds`，由工作流应用服务在允许用户和部门判断前统一拒绝排除用户。Admin 复用并扩展组织用户树组件，使动态部门规则与明确用户分别输出，同时用第二个用户选择器配置排除列表。

**Tech Stack:** Go 1.24、Gorm、Vue 3、TypeScript strict、Element Plus、Node.js 结构回归、Codex in-app Browser

---

### Task 1: 扩展核心定义和发布复制

**Files:**
- Modify: `backend/internal/workflowcore/types.go`
- Modify: `backend/internal/workflowcore/validation.go`
- Test: `backend/internal/workflowcore/definition_test.go`
- Modify: `backend/internal/service/admin/workflow/service.go`
- Test: `backend/internal/service/admin/workflow/service_test.go`

- [ ] **Step 1: 写失败测试**

增加 `all` 携带排除用户合法、`specified` 携带排除用户合法、排除 ID 为零或重复非法，以及发布复制不受请求切片后续修改影响的用例。

- [ ] **Step 2: 运行失败测试**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/workflowcore ./internal/service/admin/workflow -run 'TestValidateDefinition.*Initiator|TestApplyPublishInitiator' -count=1
```

预期：因 `InitiatorConfig.ExcludedUserIDs` 尚不存在而编译失败。

- [ ] **Step 3: 实现最小契约**

```go
type InitiatorConfig struct {
    Scope           string `json:"scope"`
    UserIDs         []uint `json:"userIds,omitempty"`
    DepartmentIDs   []uint `json:"departmentIds,omitempty"`
    ExcludedUserIDs []uint `json:"excludedUserIds,omitempty"`
}
```

`all` 只要求允许用户和部门为空；两种模式都使用 `validInitiatorIDs` 校验排除用户。发布覆盖和旧顶层配置迁移均对三个切片执行防御性复制。

- [ ] **Step 4: 重跑定向测试并确认通过**

### Task 2: 运行时优先执行排除规则

**Files:**
- Modify: `backend/internal/modules/workflow/application/service.go`
- Test: `backend/internal/modules/workflow/application/service_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Test: `backend/internal/modules/workflow/infrastructure/mapping_test.go`
- Test: `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- Test: `backend/internal/modules/workflow/transport/httpclient/handler_test.go`

- [ ] **Step 1: 写失败测试**

扩展 `publishedInitiatorAllows` 用例，覆盖全部用户排除、明确允许用户同时被排除、允许部门用户同时被排除和普通允许用户。扩展映射和 HTTP JSON 断言，要求保留 `excludedUserIds`。

- [ ] **Step 2: 运行失败测试**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure ./internal/modules/workflow/transport/httpadmin ./internal/modules/workflow/transport/httpclient -run 'Test.*Initiator|TestListDefinitions' -count=1
```

- [ ] **Step 3: 实现统一判定**

```go
func publishedInitiatorAllows(config workflowcore.InitiatorConfig, starterID string, departmentIDs []uint) bool {
    if initiatorContainsUser(config.ExcludedUserIDs, starterID) { return false }
    if config.Scope != workflowcore.InitiatorScopeSpecified { return true }
    if initiatorContainsUser(config.UserIDs, starterID) { return true }
    return initiatorDepartmentsIntersect(config.DepartmentIDs, departmentIDs)
}
```

`initiatorNeedsDepartments` 对排除用户和明确允许用户都不查询部门；定义映射对 `ExcludedUserIDs` 防御性复制。

- [ ] **Step 4: 重跑定向测试并确认通过**

### Task 3: 设计器统一组织树和排除用户

**Files:**
- Modify: `admin/src/views/workflow/types.ts`
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/views/workflow/components/WorkflowUserTreePicker.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowStartConfig.vue`
- Modify: `admin/src/views/workflow/instances/index.vue`
- Modify: `admin/src/views/scheduled-task/components/TaskEditorDialog.vue`
- Test: `admin/scripts/check-workflow-runtime-pages.mjs`
- Test: `admin/scripts/check-scheduled-task.mjs`

- [ ] **Step 1: 写失败结构检查**

要求类型包含 `excludedUserIds`；流程配置不再出现“允许发起部门”“额外允许用户”和独立 `el-tree-select`；统一选择器同时绑定 `userIds`、`departmentIds`；排除用户在两种 scope 下存在；管理代发和定时任务候选人过滤排除列表。

- [ ] **Step 2: 运行失败检查**

```bash
cd admin
npm run check:workflow-runtime-pages
npm run check:scheduled-task
```

- [ ] **Step 3: 扩展共享选择器**

增加 `departmentModelValue?: number[]` 和 `selectDepartmentRules?: boolean`。规则模式下部门节点启用 `check-strictly`，勾选部门通过 `update:departmentModelValue` 输出部门 ID；用户仍通过 `update:modelValue` 输出用户 ID；输入框汇总显示部门和用户名称。

- [ ] **Step 4: 改造流程配置和候选用户过滤**

开始节点始终保留 `excludedUserIds`；指定范围使用一个共享树绑定允许用户、允许部门；排除用户使用第二个共享树。候选用户过滤先判断 `excludedUserIds`，再执行全部或指定范围判断。

- [ ] **Step 5: 重跑两个结构检查并确认通过**

### Task 4: 同步跨端契约和文档

**Files:**
- Modify: `h5app/src/types/workflow.ts`
- Modify: `docs/architecture/generic-oa-workflow-v1.md`

- [ ] **Step 1: 类型同步**

为 H5 的 `WorkflowInitiatorConfig` 增加 `departmentIds?: number[]` 和 `excludedUserIds?: number[]`，与后端发布定义响应一致。

- [ ] **Step 2: 文档同步**

记录允许范围并集、部门实时匹配且不包含子部门、排除最高优先级、历史缺省兼容性。

### Task 5: 全量验证和浏览器回归

- [ ] **Step 1: 格式化并运行后端全量测试**

```bash
cd backend
gofmt -w <本次修改的 Go 文件>
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
```

- [ ] **Step 2: 运行 Admin 全量门禁**

```bash
cd admin
npm run check:all
```

- [ ] **Step 3: 运行 H5App 类型检查**

```bash
cd h5app
pnpm type-check
```

- [ ] **Step 4: 浏览器验证**

验证路径：流程设计器流程配置 -> 切换全部用户/指定范围 -> 统一组织树独立勾选部门和用户 -> 选择排除用户 -> 检查输入框折叠、回显和控制台。若认证阻塞，使用同一正式组件建立临时本地验证入口，验证后删除。

- [ ] **Step 5: 最终检查**

```bash
git diff --check
git status --short
```
