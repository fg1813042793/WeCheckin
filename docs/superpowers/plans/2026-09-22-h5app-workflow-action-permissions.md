# H5App OA Workflow Action Permissions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 H5App 通用 OA 流程的通过、驳回、退回和提交办理拆为独立按钮/API 权限，并由后端按请求 action 强制鉴权。

**Architecture:** 保留统一的任务完成 URL，在钉钉 H5 权限中间件中对该 URL 读取 JSON `action` 并映射到四个新 API 权限。H5App 同时检查动作按钮权限和动作 API 权限；版本化 SQL 将旧 `workflow:handle` 的历史授权复制到八个新权限后停用旧权限。

**Tech Stack:** Go 1.24、CloudWeGo Hertz、GORM、MySQL migrations、Vue 3、uni-app、TypeScript、Pinia、uView Pro、Node 结构检查脚本。

---

## File Map

- Modify: `backend/internal/support/appmenuperm/catalog.go` — 声明四个 H5App 工作流按钮权限。
- Create: `backend/internal/support/appmenuperm/workflow_action_permissions_test.go` — 校验按钮权限键、名称、父级和顺序。
- Modify: `backend/internal/support/appapiperm/catalog.go` — 用四个动作 API 权限替换旧 handle 声明，并阻止统一 complete 路由进入静态路由声明。
- Modify: `backend/internal/support/appapiperm/catalog_test.go` — 校验动作 API 目录和统一路由动态鉴权边界。
- Modify: `backend/internal/middleware/dingtalk_h5/permission.go` — 解析任务完成请求 action 并返回动作权限键。
- Modify: `backend/internal/middleware/dingtalk_h5/permission_test.go` — 覆盖四个动作、非法 action 和普通静态路由。
- Create: `backend/migrations/20260922120000_split_dingtalk_h5_workflow_action_permissions.sql` — 新增八项权限、复制历史授权、停用旧 handle。
- Create: `backend/test/internal/bootstrap/migrations/dingtalk_h5_workflow_action_permission_migration_test.go` — 校验迁移契约。
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue` — 按动作权限控制按钮、提交守卫和表单只读状态。
- Create: `h5app/scripts/check-workflow-action-permissions.mjs` — 校验按钮/API 双重权限和旧 handle 清理。
- Modify: `h5app/package.json` — 增加动作权限结构检查脚本。
- Modify: `docs/PERMISSION_CODE_FRONTEND_SYNC.md` — 记录按钮/API 权限对应关系和旧权限停用规则。

### Task 1: Add split button and API permission catalogs

**Files:**
- Create: `backend/internal/support/appmenuperm/workflow_action_permissions_test.go`
- Modify: `backend/internal/support/appmenuperm/catalog.go`
- Modify: `backend/internal/support/appapiperm/catalog_test.go`
- Modify: `backend/internal/support/appapiperm/catalog.go`

- [ ] **Step 1: Write failing button catalog test**

Create `workflow_action_permissions_test.go` with a table asserting these declarations:

```go
func TestDingTalkH5WorkflowActionButtonsAreDeclared(t *testing.T) {
    want := map[string]struct {
        name string
        path string
    }{
        "dingtalk_h5:button:workflow:approve": {name: "通过审批", path: "workflow:approve"},
        "dingtalk_h5:button:workflow:reject":  {name: "驳回流程", path: "workflow:reject"},
        "dingtalk_h5:button:workflow:return":  {name: "退回流程", path: "workflow:return"},
        "dingtalk_h5:button:workflow:submit":  {name: "提交办理", path: "workflow:submit"},
    }
    for _, item := range DingTalkH5ButtonDeclarations() {
        expected, ok := want[item.Key]
        if !ok {
            continue
        }
        if item.Name != expected.name || item.Path != expected.path || item.ParentKey != "dingtalk_h5:menu:workflow" {
            t.Fatalf("workflow action button %#v", item)
        }
        delete(want, item.Key)
    }
    if len(want) != 0 {
        t.Fatalf("missing workflow action buttons: %#v", want)
    }
}
```

- [ ] **Step 2: Extend API catalog test and verify RED**

Replace the required old handle key with:

```go
"dingtalk_h5:api:workflow:approve": false,
"dingtalk_h5:api:workflow:reject":  false,
"dingtalk_h5:api:workflow:return":  false,
"dingtalk_h5:api:workflow:submit":  false,
```

Add assertions that all four declarations use `POST /api/v2/dingtalk/h5/workflows/tasks/:id/complete`, and that `DingTalkH5RouteDeclarations()` contains no static declaration for that method/path.

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/appmenuperm ./internal/support/appapiperm -count=1
```

Expected: FAIL because the eight declarations do not exist and the old handle route is still static.

- [ ] **Step 3: Implement catalog declarations**

Add four button declarations under `dingtalk_h5:menu:workflow` with stable sort values before the form-revision buttons.

Replace the old API declaration with:

```go
dingtalkAPI("dingtalk_h5:api:workflow:approve", "OA 流程通过接口", "workflow:approve", "dingtalk_h5:api-category:workflow", workflowTaskCompletePath, 30),
dingtalkAPI("dingtalk_h5:api:workflow:reject", "OA 流程驳回接口", "workflow:reject", "dingtalk_h5:api-category:workflow", workflowTaskCompletePath, 31),
dingtalkAPI("dingtalk_h5:api:workflow:return", "OA 流程退回接口", "workflow:return", "dingtalk_h5:api-category:workflow", workflowTaskCompletePath, 32),
dingtalkAPI("dingtalk_h5:api:workflow:submit", "OA 流程提交办理接口", "workflow:submit", "dingtalk_h5:api-category:workflow", workflowTaskCompletePath, 33),
```

Define shared constants for the complete path and action permission prefix. While building `DingTalkH5RouteDeclarations`, skip the four dynamic action declarations so no declaration-order fallback remains.

- [ ] **Step 4: Run catalog tests and verify GREEN**

Run the focused command from Step 2.

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

```bash
git add backend/internal/support/appmenuperm/catalog.go \
  backend/internal/support/appmenuperm/workflow_action_permissions_test.go \
  backend/internal/support/appapiperm/catalog.go \
  backend/internal/support/appapiperm/catalog_test.go
git commit -m "feat: 拆分 H5 流程动作权限目录"
```

### Task 2: Enforce action-specific API permissions in middleware

**Files:**
- Modify: `backend/internal/middleware/dingtalk_h5/permission_test.go`
- Modify: `backend/internal/middleware/dingtalk_h5/permission.go`

- [ ] **Step 1: Write failing action mapping tests**

Add table tests for a pure helper:

```go
func TestDingTalkH5WorkflowTaskActionPermission(t *testing.T) {
    tests := map[string]string{
        "approve": "dingtalk_h5:api:workflow:approve",
        "reject":  "dingtalk_h5:api:workflow:reject",
        "return":  "dingtalk_h5:api:workflow:return",
        "submit":  "dingtalk_h5:api:workflow:submit",
    }
    for action, want := range tests {
        if got, ok := dingTalkH5WorkflowTaskActionPermission(action); !ok || got != want {
            t.Fatalf("action %s permission=(%q,%v), want %q", action, got, ok, want)
        }
    }
    for _, action := range []string{"", "cancel", "APPROVE"} {
        if _, ok := dingTalkH5WorkflowTaskActionPermission(action); ok {
            t.Fatalf("invalid action %q must not map to a permission", action)
        }
    }
}
```

Add request-aware tests using `app.NewContext(1)` with method, URI and JSON body. Assert the complete route resolves from body action, while normal routes continue through `dingTalkH5RoutePermission`.

- [ ] **Step 2: Run middleware test and verify RED**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/middleware/dingtalk_h5 -count=1
```

Expected: FAIL because action-aware helpers do not exist.

- [ ] **Step 3: Implement request-aware permission resolution**

Add:

```go
const dingTalkH5WorkflowTaskCompletePath = "/api/v2/dingtalk/h5/workflows/tasks/:id/complete"

var errDingTalkH5WorkflowActionInvalid = errors.New("流程任务操作无效")

func dingTalkH5WorkflowTaskActionPermission(action string) (string, bool) {
    switch strings.TrimSpace(action) {
    case "approve":
        return "dingtalk_h5:api:workflow:approve", true
    case "reject":
        return "dingtalk_h5:api:workflow:reject", true
    case "return":
        return "dingtalk_h5:api:workflow:return", true
    case "submit":
        return "dingtalk_h5:api:workflow:submit", true
    default:
        return "", false
    }
}
```

Implement `dingTalkH5RequestPermission(c)` that recognizes the concrete complete URL with route-pattern matching, unmarshals only `{ "action": ... }` from `c.Request.Body()`, and returns `errDingTalkH5WorkflowActionInvalid` for malformed or unsupported values. Update `DingTalkH5Perm` to return `请求参数格式无效` for malformed JSON, `流程任务操作无效` for invalid action, and keep `无权限访问` for permission denial or unmapped routes.

- [ ] **Step 4: Verify middleware and workflow handler behavior**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/middleware/dingtalk_h5 ./internal/modules/workflow/transport/httpclient -count=1
```

Expected: PASS. Existing handler tests still decode the same request body successfully.

- [ ] **Step 5: Commit Task 2**

```bash
git add backend/internal/middleware/dingtalk_h5/permission.go \
  backend/internal/middleware/dingtalk_h5/permission_test.go
git commit -m "feat: 按动作校验 H5 流程处理权限"
```

### Task 3: Migrate historical workflow handle grants

**Files:**
- Create: `backend/migrations/20260922120000_split_dingtalk_h5_workflow_action_permissions.sql`
- Create: `backend/test/internal/bootstrap/migrations/dingtalk_h5_workflow_action_permission_migration_test.go`

- [ ] **Step 1: Write failing migration contract test**

The test must locate exactly one `*_split_dingtalk_h5_workflow_action_permissions.sql` file and assert all eight keys, the source key `dingtalk_h5:api:workflow:handle`, `grant_effect`, `grant_scope_value`, `ON DUPLICATE KEY UPDATE`, and an update setting the old permission status to zero.

```go
for _, snippet := range []string{
    "dingtalk_h5:button:workflow:approve",
    "dingtalk_h5:button:workflow:reject",
    "dingtalk_h5:button:workflow:return",
    "dingtalk_h5:button:workflow:submit",
    "dingtalk_h5:api:workflow:approve",
    "dingtalk_h5:api:workflow:reject",
    "dingtalk_h5:api:workflow:return",
    "dingtalk_h5:api:workflow:submit",
    "dingtalk_h5:api:workflow:handle",
    "source_grant.`grant_effect`",
    "source_grant.`grant_scope_value`",
    "`permission_status` = 0",
} {
    if !strings.Contains(text, snippet) {
        t.Fatalf("migration missing %q", snippet)
    }
}
```

- [ ] **Step 2: Run migration test and verify RED**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations -run TestDingTalkH5WorkflowActionPermissionMigration -count=1
```

Expected: FAIL because the migration does not exist.

- [ ] **Step 3: Create idempotent migration**

Insert the eight permission rows using the same permission columns and millisecond timestamps as neighboring migrations. Build a derived mapping table from the old handle key to all eight targets. Copy active grants with:

```sql
source_grant.`grant_effect`,
source_grant.`grant_scope_value`,
'split-dingtalk-h5-workflow-action-permissions',
source_grant.`grant_status`
```

Use `ON DUPLICATE KEY UPDATE` to update permission ID, effect, scope, source and status. Finally disable the old permission:

```sql
UPDATE `permissions`
SET `permission_status` = 0,
    `permission_edit_time` = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
    `updated_at` = NOW(3)
WHERE `permission_key` = 'dingtalk_h5:api:workflow:handle';
```

- [ ] **Step 4: Run migration tests and verify GREEN**

Run the focused test and then:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 3**

```bash
git add backend/migrations/20260922120000_split_dingtalk_h5_workflow_action_permissions.sql \
  backend/test/internal/bootstrap/migrations/dingtalk_h5_workflow_action_permission_migration_test.go
git commit -m "feat: 迁移 H5 流程动作权限"
```

### Task 4: Wire H5App buttons to split permissions

**Files:**
- Create: `h5app/scripts/check-workflow-action-permissions.mjs`
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`
- Modify: `h5app/package.json`

- [ ] **Step 1: Write failing H5 structure check**

The script must read `WorkflowDetailPanel.vue` and assert:

```js
const required = [
  'dingtalk_h5:button:workflow:approve',
  'dingtalk_h5:api:workflow:approve',
  'dingtalk_h5:button:workflow:reject',
  'dingtalk_h5:api:workflow:reject',
  'dingtalk_h5:button:workflow:return',
  'dingtalk_h5:api:workflow:return',
  'dingtalk_h5:button:workflow:submit',
  'dingtalk_h5:api:workflow:submit',
  'const canApprove = computed',
  'const canReject = computed',
  'const canReturn = computed',
  'const canSubmit = computed',
  'const canEditTaskForm = computed',
  'const hasTaskAction = computed',
]
```

It must reject use of `dingtalk_h5:api:workflow:handle` in the component. Add `check:workflow-action-permissions` to `package.json`.

- [ ] **Step 2: Run check and verify RED**

```bash
cd h5app
pnpm check:workflow-action-permissions
```

Expected: FAIL because the component still uses the aggregate handle permission.

- [ ] **Step 3: Split structural task state from action capabilities**

Replace `canHandle` with these responsibilities:

```ts
const hasPendingAssignedTask = computed(() => Boolean(
  activeTask.value?.status === 'pending'
  && isWorkflowTaskAssignedToUser(activeTask.value, currentUserId.value),
))

function hasWorkflowActionPermission(action: 'approve' | 'reject' | 'return' | 'submit') {
  return auth.hasButtonPermission(`dingtalk_h5:button:workflow:${action}`)
    && auth.hasApiPermission(`dingtalk_h5:api:workflow:${action}`)
}

const canApprove = computed(() => hasPendingAssignedTask.value
  && activeNodeType.value === 'approval'
  && hasWorkflowActionPermission('approve'))
const canReject = computed(() => hasPendingAssignedTask.value
  && activeNodeType.value === 'approval'
  && hasWorkflowActionPermission('reject'))
const canSubmit = computed(() => hasPendingAssignedTask.value
  && activeNodeType.value === 'handle'
  && hasWorkflowActionPermission('submit'))
```

Retain existing return-path and parallel checks, then additionally require `hasWorkflowActionPermission('return')`. Define:

```ts
const canEditTaskForm = computed(() => canApprove.value || canSubmit.value)
const hasTaskAction = computed(() => canApprove.value || canReject.value || canReturn.value || canSubmit.value)
```

- [ ] **Step 4: Rewire guards and template**

- Use `canEditTaskForm` for form `readonly`, field actions, validation and writable form submission.
- Use `canApprove` only for approval confirmation.
- Use `canSubmit` only for handle submission.
- Use `canReject` in `openReject` and `submitReject`.
- Use `canReturn` in return dialog guards.
- Render each button independently instead of wrapping all actions in `<template v-if="canHandle">`.
- Use `hasTaskAction` in action-bar visibility logic.
- Keep task status display based on the assigned pending task, not on whether the user has a particular action permission.

- [ ] **Step 5: Run H5 checks and type validation**

```bash
cd h5app
pnpm check:workflow-action-permissions
pnpm check:workflow-module
pnpm lint
pnpm type-check
```

Expected: PASS.

- [ ] **Step 6: Commit Task 4**

```bash
git add h5app/src/pages/workflow/components/WorkflowDetailPanel.vue \
  h5app/scripts/check-workflow-action-permissions.mjs \
  h5app/package.json
git commit -m "feat: 细分 H5 流程操作按钮权限"
```

### Task 5: Document contract and run integrated verification

**Files:**
- Modify: `docs/PERMISSION_CODE_FRONTEND_SYNC.md`

- [ ] **Step 1: Update permission documentation**

Add a table documenting the four button/API pairs, the shared complete endpoint, action values, and the separate withdraw permission. State that both button and API permissions are required for H5App display, while the API permission is the backend security boundary.

- [ ] **Step 2: Run backend focused tests**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test \
  ./internal/support/appmenuperm \
  ./internal/support/appapiperm \
  ./internal/middleware/dingtalk_h5 \
  ./internal/modules/workflow/transport/httpclient \
  ./test/internal/bootstrap/migrations \
  -count=1
```

Expected: PASS.

- [ ] **Step 3: Run backend full verification**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./...
```

Expected: PASS, or report unrelated pre-existing failures separately without modifying unrelated modules.

- [ ] **Step 4: Run H5App verification and build**

```bash
cd h5app
pnpm check:workflow-action-permissions
pnpm check:workflow-module
pnpm lint
pnpm type-check
pnpm build:h5
```

Expected: PASS. Build output must not contain missing permission key or TypeScript errors.

- [ ] **Step 5: Run repository diff checks**

```bash
git diff --check
git status --short
```

Expected: no whitespace errors; unrelated pre-existing worktree changes remain untouched.

- [ ] **Step 6: Commit documentation**

```bash
git add docs/PERMISSION_CODE_FRONTEND_SYNC.md
git commit -m "docs: 说明 H5 流程动作权限"
```

### Task 6: Manual acceptance matrix

**Files:** None.

- [ ] **Step 1: Apply migration in the test environment**

Run the repository migration command against the test database and verify the old handle permission is disabled while all eight new permissions are active.

- [ ] **Step 2: Verify role combinations**

Use six test roles:

1. Only approve button/API: approval task shows only confirmation.
2. Only reject button/API: approval task shows only rejection; form stays read-only.
3. Only return button/API: eligible approval task shows only return; parallel or first-node tasks still hide return.
4. Only submit button/API: handle task shows only submit and allows its writable form fields.
5. All eight permissions: behavior matches the pre-split experience.
6. No action permissions: task remains visible in pending lists but exposes no processing action.

- [ ] **Step 3: Verify direct-call denial**

For each role, manually call the unified complete endpoint with a different action and confirm the backend returns `无权限访问` without changing the task or instance state.

- [ ] **Step 4: Verify migration compatibility**

Use a role that had the old handle permission before migration. After migration, confirm it received all eight new grants and retains full workflow processing capability.
