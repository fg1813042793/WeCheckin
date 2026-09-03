# Workflow Center Search Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为钉钉 H5 流程中心的待办、已处理和抄送列表增加服务端搜索条件，为四个记录页签保留独立筛选状态，并提供不受发起权限裁剪的已发布流程分类选项。

**Architecture:** 在现有实例与任务查询 DTO 上增加可选筛选字段，GORM 使用实例子查询在统计和分页前完成分类、提交时间及发起人用户名过滤。application 从全部已发布定义中提取去重分类，H5 通过独立、复用流程查看权限的分类接口加载记录筛选选项，现有可发起定义接口保持不变。H5 将“我的申请”现有筛选状态泛化为按页签隔离的编辑态和应用态，并根据页签展示发起人或审批状态控件。

**Tech Stack:** Go 1.24、Hertz、GORM、MySQL、Swagger、Vue 3、TypeScript、uni-app、uView Pro、SCSS

---

## File Map

- Modify `backend/internal/modules/workflow/application/types.go`: 扩展实例和任务查询 DTO。
- Modify `backend/internal/modules/workflow/application/service.go`: 统一裁剪并校验发起人用户名关键字，提取已发布流程分类。
- Create `backend/internal/modules/workflow/application/search_filters_test.go`: 锁定用户名关键字长度边界及分类去重排序。
- Modify `backend/internal/modules/workflow/infrastructure/gorm_store.go`: 增加用户名字面量模糊匹配，并让任务查询复用实例筛选。
- Create `backend/internal/modules/workflow/infrastructure/workflow_search_filters_test.go`: 锁定实例与任务筛选 SQL。
- Modify `backend/internal/modules/workflow/transport/httpclient/handler.go`: 提供分类读取端点并解析 H5 新增查询参数。
- Create `backend/internal/modules/workflow/transport/httpclient/search_filters_test.go`: 锁定分类响应、H5 参数解析和登录用户边界。
- Modify `backend/internal/routes/v2/dingtalkh5/routes.go`: 注册受保护的流程分类路由。
- Modify `backend/internal/routes/v2/dingtalkh5/workflow_routes_test.go`: 锁定分类路由使用受保护分组。
- Modify `backend/internal/support/appapiperm/catalog.go`: 将分类路由映射到现有流程查看权限。
- Modify `backend/internal/support/appapiperm/catalog_test.go`: 锁定分类路由权限映射。
- Modify `backend/internal/routes/v2/swagger/h5app.go`: 补充分类接口及实例、任务搜索参数声明。
- Create `backend/cmd/swagger_workflow_search_filters_test.go`: 校验生成 Swagger 包含分类接口和新增参数。
- Regenerate `backend/docs/swagger/docs.go`, `backend/docs/swagger/swagger.json`, `backend/docs/swagger/swagger.yaml`: 通过 `swag init` 生成，不手工编辑。
- Modify `h5app/src/types/workflow.ts`: 扩展实例和任务查询参数类型。
- Modify `h5app/src/api/workflow.ts`: 增加流程分类接口调用。
- Modify `h5app/src/pages/workflow/components/WorkflowCenter.vue`: 实现筛选矩阵、页签独立状态与查询接线。
- Modify `h5app/scripts/check-workflow-module.mjs`: 锁定筛选控件和状态隔离结构。

> 当前共享后端文件和 `h5app` 已含大量未提交修改。实施时不自动提交共享实现文件，不清理或覆盖无关改动；每个任务用聚焦测试和 `git diff --check` 作为检查点。

### Task 1: Add server-side workflow search filters and category projection

**Files:**
- Create: `backend/internal/modules/workflow/application/search_filters_test.go`
- Create: `backend/internal/modules/workflow/infrastructure/workflow_search_filters_test.go`
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`

- [ ] **Step 1: Write failing application validation and category tests**

Create `application/search_filters_test.go` using the existing same-package `fakeStore`:

```go
package application

import (
    "context"
    "errors"
    "strings"
    "testing"
)

func TestWorkflowSearchRejectsLongStarterName(t *testing.T) {
    service := &Service{store: &fakeStore{}}
    keyword := strings.Repeat("名", 51)

    if _, err := service.ListInstances(context.Background(), InstanceQuery{StarterName: keyword}); !errors.Is(err, ErrStarterNameSearchTooLong) {
        t.Fatalf("ListInstances error = %v, want %v", err, ErrStarterNameSearchTooLong)
    }
    if _, err := service.ListTasks(context.Background(), TaskQuery{StarterName: keyword}); !errors.Is(err, ErrStarterNameSearchTooLong) {
        t.Fatalf("ListTasks error = %v, want %v", err, ErrStarterNameSearchTooLong)
    }
}

func TestListPublishedDefinitionCategoriesReturnsSortedUniqueValues(t *testing.T) {
    service := &Service{store: &fakeStore{publishedDefinitions: []PublishedDefinition{
        {Category: " hr "},
        {Category: "finance"},
        {Category: "hr"},
        {Category: " "},
    }}}

    categories, err := service.ListPublishedDefinitionCategories(context.Background())
    if err != nil {
        t.Fatalf("ListPublishedDefinitionCategories() error = %v", err)
    }
    if got, want := strings.Join(categories, ","), "finance,hr"; got != want {
        t.Fatalf("categories = %q, want %q", got, want)
    }
}
```

- [ ] **Step 2: Write failing SQL projection tests**

Create a DryRun test file with a shared database helper:

```go
package infrastructure

import (
    "database/sql"
    "reflect"
    "strings"
    "testing"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"

    workflowmodel "wecheckin/backend/internal/model/workflow"
    "wecheckin/backend/internal/modules/workflow/application"
)

func openWorkflowSearchDryRunDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(mysql.New(mysql.Config{
        Conn: &sql.DB{}, SkipInitializeWithVersion: true,
    }), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
    if err != nil {
        t.Fatalf("open dry-run database: %v", err)
    }
    return db
}

func TestApplyInstanceFiltersIncludesLiteralStarterName(t *testing.T) {
    db := openWorkflowSearchDryRunDB(t)
    query := application.InstanceQuery{StarterName: " 研发%_!张 "}
    var rows []workflowmodel.ProcessInstance
    statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
    sqlText := statement.SQL.String()

    for _, fragment := range []string{"users", "user_name LIKE ?", "ESCAPE '!'"} {
        if !strings.Contains(sqlText, fragment) {
            t.Fatalf("starter name query missing %q: %s", fragment, sqlText)
        }
    }
    if want := []interface{}{"%研发!%!_!!张%"}; !reflect.DeepEqual(statement.Vars, want) {
        t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
    }
}

func TestApplyTaskFiltersIncludesInstanceSearchFilters(t *testing.T) {
    db := openWorkflowSearchDryRunDB(t)
    query := application.TaskQuery{
        Status: "pending", DefinitionCategory: "performance", StarterName: "张",
        StartTimeFrom: 1000, StartTimeTo: 1999,
    }
    var rows []workflowmodel.ProcessTask
    statement := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), query).Find(&rows).Statement
    sqlText := statement.SQL.String()

    for _, fragment := range []string{
        "task_status = ?", "instance_id IN", "workflow_process_instances",
        "workflow_definitions", "definition_category = ?", "users",
        "user_name LIKE ?", "start_time >= ?", "start_time <= ?",
    } {
        if !strings.Contains(sqlText, fragment) {
            t.Fatalf("task query missing %q: %s", fragment, sqlText)
        }
    }
    if want := []interface{}{"pending", "performance", "%张%", int64(1000), int64(1999)}; !reflect.DeepEqual(statement.Vars, want) {
        t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
    }
}
```

- [ ] **Step 3: Run the focused tests and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -run 'WorkflowSearch|StarterName|DefinitionCategories|TaskFiltersIncludes' -count=1
```

Expected: compilation fails because the query fields, `ErrStarterNameSearchTooLong` and `ListPublishedDefinitionCategories` do not exist.

- [ ] **Step 4: Add query DTO fields and application validation**

Extend `InstanceQuery`:

```go
StarterName string
```

Extend `TaskQuery`:

```go
DefinitionCategory string
StarterName        string
StartTimeFrom      int64
StartTimeTo        int64
```

Add to the application errors:

```go
ErrStarterNameSearchTooLong = errors.New("发起人用户名关键字不能超过50个字符")
```

Import `sort` and `unicode/utf8`, then add:

```go
const maxStarterNameSearchLength = 50

func normalizeStarterNameSearch(value string) (string, error) {
    value = strings.TrimSpace(value)
    if utf8.RuneCountInString(value) > maxStarterNameSearchLength {
        return "", ErrStarterNameSearchTooLong
    }
    return value, nil
}

func (service *Service) ListPublishedDefinitionCategories(ctx context.Context) ([]string, error) {
    definitions, err := service.ListPublishedDefinitions(ctx)
    if err != nil {
        return nil, err
    }
    seen := make(map[string]struct{}, len(definitions))
    categories := make([]string, 0, len(definitions))
    for _, definition := range definitions {
        category := strings.TrimSpace(definition.Category)
        if category == "" {
            continue
        }
        if _, exists := seen[category]; exists {
            continue
        }
        seen[category] = struct{}{}
        categories = append(categories, category)
    }
    sort.Strings(categories)
    return categories, nil
}
```

Call this helper from both `ListInstances` and `ListTasks` before page normalization, assign the normalized value back to `query.StarterName`, and return the validation error without invoking the store.

Run the focused test again. Expected: the application test passes; infrastructure tests compile and fail because the SQL does not contain the new filters.

- [ ] **Step 5: Implement literal LIKE matching**

Add a private helper in `gorm_store.go`:

```go
func containsLikePattern(value string) string {
    replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
    return "%" + replacer.Replace(strings.TrimSpace(value)) + "%"
}
```

In `applyInstanceFilters`, add after the direct starter ID filter and before time filters:

```go
if value := strings.TrimSpace(query.StarterName); value != "" {
    db = db.Where(`EXISTS (
        SELECT 1 FROM users starter_user
        WHERE starter_user.id = CAST(workflow_process_instances.starter_id AS UNSIGNED)
        AND starter_user.user_name LIKE ? ESCAPE '!'
    )`, containsLikePattern(value))
}
```

This matches `users.user_name` only and treats `%`, `_` and `!` as literal input.

- [ ] **Step 6: Reuse instance filters for task pagination**

Add a predicate:

```go
func hasTaskInstanceFilters(query application.TaskQuery) bool {
    return strings.TrimSpace(query.DefinitionCategory) != "" ||
        strings.TrimSpace(query.StarterName) != "" ||
        query.StartTimeFrom > 0 || query.StartTimeTo > 0
}
```

After task ID/assignee/status filters in `applyTaskFilters`, add:

```go
if hasTaskInstanceFilters(query) {
    instances := db.Session(&gorm.Session{NewDB: true}).
        Model(&workflowmodel.ProcessInstance{}).
        Select("id")
    instances = applyInstanceFilters(instances, application.InstanceQuery{
        DefinitionCategory: query.DefinitionCategory,
        StarterName:        query.StarterName,
        StartTimeFrom:      query.StartTimeFrom,
        StartTimeTo:        query.StartTimeTo,
    })
    db = db.Where("instance_id IN (?)", instances)
}
```

`ListTasks` already invokes `applyTaskFilters` for both `COUNT` and `Find`, so no second query path is added.

- [ ] **Step 7: Format and verify GREEN**

Run:

```bash
cd backend
gofmt -w internal/modules/workflow/application/types.go internal/modules/workflow/application/service.go internal/modules/workflow/application/search_filters_test.go internal/modules/workflow/infrastructure/gorm_store.go internal/modules/workflow/infrastructure/workflow_search_filters_test.go
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -run 'WorkflowSearch|StarterName|DefinitionCategories|TaskFiltersIncludes' -count=1
git diff --check -- internal/modules/workflow/application/types.go internal/modules/workflow/application/service.go internal/modules/workflow/application/search_filters_test.go internal/modules/workflow/infrastructure/gorm_store.go internal/modules/workflow/infrastructure/workflow_search_filters_test.go
```

Expected: tests pass and `git diff --check` has no output.

### Task 2: Add the H5 category endpoint, parse query parameters, and update Swagger

**Files:**
- Create: `backend/internal/modules/workflow/transport/httpclient/search_filters_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/handler.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/routes.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/workflow_routes_test.go`
- Modify: `backend/internal/support/appapiperm/catalog.go`
- Modify: `backend/internal/support/appapiperm/catalog_test.go`
- Create: `backend/cmd/swagger_workflow_search_filters_test.go`
- Modify: `backend/internal/routes/v2/swagger/h5app.go`
- Regenerate: `backend/docs/swagger/docs.go`
- Regenerate: `backend/docs/swagger/swagger.json`
- Regenerate: `backend/docs/swagger/swagger.yaml`

- [ ] **Step 1: Write a failing H5 transport test**

Use the existing `runtimeServiceStub` and `newUserContext` helpers from the same package:

```go
package httpclient

import (
    "context"
    "strings"
    "testing"
)

func (stub *runtimeServiceStub) ListPublishedDefinitionCategories(context.Context) ([]string, error) {
    return []string{"finance", "hr"}, nil
}

func TestListDefinitionCategoriesReturnsPublishedCategories(t *testing.T) {
    handler := NewRuntimeHandler(&runtimeServiceStub{})
    requestContext := newUserContext(42)

    handler.ListDefinitionCategories(context.Background(), requestContext)

    body := string(requestContext.Response.Body())
    if !strings.Contains(body, `"data":["finance","hr"]`) {
        t.Fatalf("category response = %s", body)
    }
}

func TestWorkflowListQueriesParseSearchFilters(t *testing.T) {
    stub := &runtimeServiceStub{}
    handler := NewRuntimeHandler(stub)

    instanceContext := newUserContext(42)
    instanceContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/instances?scope=handled&definitionCategory=performance&starterName=%E5%BC%A0&startTimeFrom=1000&startTimeTo=1999&page=2&pageSize=10")
    handler.ListMyInstances(context.Background(), instanceContext)
    if stub.instanceQuery.DefinitionCategory != "performance" || stub.instanceQuery.StarterName != "张" ||
        stub.instanceQuery.StartTimeFrom != 1000 || stub.instanceQuery.StartTimeTo != 1999 {
        t.Fatalf("instance search filters = %+v", stub.instanceQuery)
    }

    taskContext := newUserContext(42)
    taskContext.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/tasks?assigneeId=999&status=pending&definitionCategory=performance&starterName=%E6%9D%8E&startTimeFrom=2000&startTimeTo=2999&page=3&pageSize=15")
    handler.ListMyTasks(context.Background(), taskContext)
    if stub.taskQuery.AssigneeID != "" || stub.taskQuery.Status != "pending" ||
        stub.taskQuery.DefinitionCategory != "performance" || stub.taskQuery.StarterName != "李" ||
        stub.taskQuery.StartTimeFrom != 2000 || stub.taskQuery.StartTimeTo != 2999 {
        t.Fatalf("task search filters = %+v", stub.taskQuery)
    }
}
```

- [ ] **Step 2: Run the transport test and verify RED**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/transport/httpclient -run 'TestListDefinitionCategories|TestWorkflowListQueries' -count=1
```

Expected: compilation or assertions fail because the category service/handler and the new query fields are not wired.

- [ ] **Step 3: Add the protected category endpoint and parse the new parameters**

Extend `RuntimeService` with `ListPublishedDefinitionCategories(context.Context) ([]string, error)` and add `ListDefinitionCategories`. The handler must require an authenticated actor exactly like the existing definition list, call the application method, and return the string array through `response.JSON`.

Register:

```go
auth.GET("/workflows/categories", workflowHandler.ListDefinitionCategories)
```

Add the route snippet to `workflow_routes_test.go`. Increase the `DingTalkH5RouteDeclarations` extra-capacity allowance from `+7` to `+8`, then add the exact route and its expected map in `catalog_test.go`, reusing:

```go
RouteDeclaration{
    Method: "GET",
    Path: "/api/v2/dingtalk/h5/workflows/categories",
    PermissionKey: "dingtalk_h5:api:workflow:view",
}
```

Do not add a new API permission declaration or migration.

In `ListMyInstances`, populate:

```go
StarterName: strings.TrimSpace(c.Query("starterName")),
```

In `ListMyTasks`, populate:

```go
DefinitionCategory: strings.TrimSpace(c.Query("definitionCategory")),
StarterName:        strings.TrimSpace(c.Query("starterName")),
StartTimeFrom:      queryInt64(c, "startTimeFrom"),
StartTimeTo:        queryInt64(c, "startTimeTo"),
```

Do not parse `assigneeId`; `Service.ListMyTasks` remains responsible for setting the authenticated actor.

- [ ] **Step 4: Verify transport, route, and permission tests GREEN**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/transport/httpclient ./internal/routes/v2/dingtalkh5 ./internal/support/appapiperm -run 'DefinitionCategories|WorkflowListQueries|WorkflowRoutes|APIDeclarations' -count=1
```

Expected: PASS, including the protected route and `workflow:view` mapping.

- [ ] **Step 5: Write a failing generated Swagger test**

Create `backend/cmd/swagger_workflow_search_filters_test.go` using the existing `loadSwaggerOperations` and `hasSwaggerParameter` helpers:

```go
package main

import "testing"

func TestSwaggerDocumentsWorkflowSearchFilters(t *testing.T) {
    paths := loadSwaggerOperations(t)
    if _, ok := paths["/api/v2/dingtalk/h5/workflows/categories"]["get"]; !ok {
        t.Error("Swagger must document GET /api/v2/dingtalk/h5/workflows/categories")
    }
    tests := []struct {
        path string
        name string
    }{
        {path: "/api/v2/dingtalk/h5/workflows/instances", name: "starterName"},
        {path: "/api/v2/dingtalk/h5/workflows/tasks", name: "definitionCategory"},
        {path: "/api/v2/dingtalk/h5/workflows/tasks", name: "starterName"},
        {path: "/api/v2/dingtalk/h5/workflows/tasks", name: "startTimeFrom"},
        {path: "/api/v2/dingtalk/h5/workflows/tasks", name: "startTimeTo"},
    }
    for _, test := range tests {
        operation, ok := paths[test.path]["get"]
        if !ok || !hasSwaggerParameter(operation.Parameters, test.name, "query", false) {
            t.Errorf("Swagger GET %s must document query parameter %q", test.path, test.name)
        }
    }
}
```

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./cmd -run TestSwaggerDocumentsWorkflowSearchFilters -count=1
```

Expected: FAIL because generated Swagger lacks the category route and new parameters.

- [ ] **Step 6: Update Swagger source comments and regenerate**

Add a `swaggerV2H5AppWorkflowCategoriesGet` declaration with `@Router /api/v2/dingtalk/h5/workflows/categories [get]`. Add `starterName` to `swaggerV2H5AppWorkflowInstancesGet`. Add category, starter name and start-time parameters to `swaggerV2H5AppWorkflowTasksGet`:

```go
// @Param starterName query string false "发起人用户名关键字"
// @Param definitionCategory query string false "流程定义分类"
// @Param startTimeFrom query int false "流程提交时间起始值（毫秒）"
// @Param startTimeTo query int false "流程提交时间截止值（毫秒）"
```

Run the required generator rather than editing generated files:

```bash
cd backend
swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger
GOCACHE=$PWD/../.cache/go-build go test ./cmd -run TestSwaggerDocumentsWorkflowSearchFilters -count=1
```

Expected: generated files update and the Swagger test passes.

- [ ] **Step 7: Format and inspect the backend boundary**

```bash
cd backend
gofmt -w internal/modules/workflow/transport/httpclient/handler.go internal/modules/workflow/transport/httpclient/search_filters_test.go internal/routes/v2/dingtalkh5/routes.go internal/routes/v2/dingtalkh5/workflow_routes_test.go internal/support/appapiperm/catalog.go internal/support/appapiperm/catalog_test.go internal/routes/v2/swagger/h5app.go cmd/swagger_workflow_search_filters_test.go
git diff --check -- internal/modules/workflow/transport/httpclient/handler.go internal/modules/workflow/transport/httpclient/search_filters_test.go internal/routes/v2/dingtalkh5/routes.go internal/routes/v2/dingtalkh5/workflow_routes_test.go internal/support/appapiperm/catalog.go internal/support/appapiperm/catalog_test.go internal/routes/v2/swagger/h5app.go cmd/swagger_workflow_search_filters_test.go docs/swagger
```

Expected: no whitespace errors.

### Task 3: Add isolated per-tab filter state in H5

**Files:**
- Modify: `h5app/scripts/check-workflow-module.mjs`
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/src/api/workflow.ts`
- Modify: `h5app/src/pages/workflow/components/WorkflowCenter.vue`

- [ ] **Step 1: Add failing structural requirements**

Update the `WorkflowCenter.vue` rule to require:

```js
[
  'WorkflowRecordFilters',
  'recordFilters',
  'appliedRecordFilters',
  'activeRecordFilters',
  'activeAppliedRecordFilters',
  'showStarterNameFilter',
  'showStatusFilter',
  'historyStatusOptions',
  'listWorkflowCategories',
  'workflowCategories',
  'recordCategoryOptions',
  'queryRecords',
  'resetRecordFilters',
  'placeholder="输入发起人用户名"',
  'maxlength="50"',
  'activeTab === \'pending\'',
  'activeTab === \'handled\'',
  'activeTab === \'copied\'',
]
```

Remove the old required names `applicationFilters`, `appliedApplicationFilters`, `applicationCategoryOptions`, `queryApplications` and `resetApplicationFilters`, then forbid them so the page cannot regress to one shared filter state or derive record categories from startable definitions.

Require `starterName?: string` in the workflow query type source and task query source together with task category/start-time fields.

- [ ] **Step 2: Run the H5 structure check and verify RED**

```bash
cd h5app
node scripts/check-workflow-module.mjs
```

Expected: FAIL on the new per-tab state and controls.

- [ ] **Step 3: Extend H5 query types and API**

Add to `WorkflowInstanceQuery`:

```ts
starterName?: string
```

Add to `WorkflowTaskQuery`:

```ts
definitionCategory?: string
starterName?: string
startTimeFrom?: number
startTimeTo?: number
```

Add to `h5app/src/api/workflow.ts`:

```ts
export function listWorkflowCategories() {
  return get<string[]>(`${WORKFLOW_API}/categories`)
}
```

- [ ] **Step 4: Replace the single application filter state**

Define:

```ts
type WorkflowListTab = Exclude<WorkflowCenterTab, 'start'>

interface WorkflowRecordFilters extends WorkflowHistoryDateFilters {
  definitionCategory: string
  starterName: string
  status: string
}

const workflowListTabs: WorkflowListTab[] = ['pending', 'handled', 'started', 'copied']

function emptyRecordFilters(): WorkflowRecordFilters {
  return {
    definitionCategory: '',
    starterName: '',
    status: '',
    startDateFrom: '',
    startDateTo: '',
  }
}

function createRecordFilters() {
  return Object.fromEntries(
    workflowListTabs.map(tab => [tab, emptyRecordFilters()]),
  ) as Record<WorkflowListTab, WorkflowRecordFilters>
}
```

Replace the two application refs with reactive per-tab maps:

```ts
const recordFilters = reactive(createRecordFilters())
const appliedRecordFilters = reactive(createRecordFilters())
const activeListTab = computed<WorkflowListTab>(() => activeTab.value === 'start' ? 'pending' : activeTab.value)
const activeRecordFilters = computed(() => recordFilters[activeListTab.value])
const activeAppliedRecordFilters = computed(() => appliedRecordFilters[activeListTab.value])
const showStarterNameFilter = computed(() => activeTab.value === 'pending' || activeTab.value === 'handled')
const showStatusFilter = computed(() => activeTab.value === 'copied')
```

Reuse the existing status values:

```ts
const historyStatusOptions = [
  { label: '全部状态', value: '' },
  { label: '审批中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已撤回', value: 'withdrawn' },
  { label: '已取消', value: 'cancelled' },
]
```

Import `listWorkflowCategories`, keep start-page categories based on `definitions`, and use the dedicated response only for record filters:

```ts
const workflowCategories = ref<string[]>([])
const categoriesLoading = ref(false)
const recordCategoryOptions = computed(() => [
  { value: '', label: '全部分类' },
  ...workflowCategories.value.map(value => ({ value, label: value })),
])

async function loadWorkflowCategories() {
  if (!hasViewPermission.value || categoriesLoading.value)
    return
  categoriesLoading.value = true
  try {
    const response = await listWorkflowCategories()
    workflowCategories.value = Array.isArray(response?.data) ? response.data : []
  }
  catch {
    uni.showToast({ title: '流程分类加载失败', icon: 'none' })
  }
  finally {
    categoriesLoading.value = false
  }
}
```

Include `loadWorkflowCategories()` in `refreshAll`. Do not replace `categories`, which still groups the startable definitions on the “发起审批” tab.

- [ ] **Step 5: Send only the active tab's applied filters**

At the start of `loadCurrentList`, derive:

```ts
const filters = activeAppliedRecordFilters.value
const timeQuery = buildWorkflowHistoryTimeQuery(filters) || {}
const definitionCategory = filters.definitionCategory || undefined
```

For pending tasks pass:

```ts
{
  status: 'pending',
  definitionCategory,
  starterName: filters.starterName.trim() || undefined,
  startTimeFrom: timeQuery.startTimeFrom,
  startTimeTo: timeQuery.startTimeTo,
  page: page.value,
  pageSize,
}
```

For instance tabs pass:

```ts
{
  scope: activeTab.value,
  definitionCategory,
  starterName: activeTab.value === 'handled' ? filters.starterName.trim() || undefined : undefined,
  status: activeTab.value === 'copied' ? filters.status || undefined : undefined,
  startTimeFrom: timeQuery.startTimeFrom,
  startTimeTo: timeQuery.startTimeTo,
  page: page.value,
  pageSize,
}
```

Keep `loadCounts` unchanged so navigation totals remain unfiltered.

- [ ] **Step 6: Generalize query and reset commands**

Replace the application-only methods with:

```ts
function queryRecords() {
  if (listLoading.value)
    return
  if (buildWorkflowHistoryTimeQuery(activeRecordFilters.value) === null) {
    uni.showToast({ title: '请检查提交时间范围', icon: 'none' })
    return
  }
  Object.assign(appliedRecordFilters[activeListTab.value], activeRecordFilters.value)
  page.value = 1
  void loadCurrentList()
}

function resetRecordFilters() {
  if (listLoading.value)
    return
  Object.assign(recordFilters[activeListTab.value], emptyRecordFilters())
  Object.assign(appliedRecordFilters[activeListTab.value], emptyRecordFilters())
  page.value = 1
  void loadCurrentList()
}
```

- [ ] **Step 7: Render the page-specific controls**

Show the filter band on every list tab. Keep category and submission date always visible. Add:

```vue
<view v-if="showStarterNameFilter" class="workflow-center__filter-field">
  <text class="workflow-center__filter-label">发起人</text>
  <input
    v-model="activeRecordFilters.starterName"
    class="workflow-center__filter-input"
    type="text"
    maxlength="50"
    placeholder="输入发起人用户名"
    :disabled="listLoading"
    @keyup.enter="queryRecords"
  >
</view>

<view v-if="showStatusFilter" class="workflow-center__filter-field">
  <text class="workflow-center__filter-label">审批状态</text>
  <select
    v-model="activeRecordFilters.status"
    class="workflow-center__filter-select"
    :disabled="listLoading"
    aria-label="审批状态"
  >
    <option v-for="option in historyStatusOptions" :key="option.value" :value="option.value">
      {{ option.label }}
    </option>
  </select>
</view>
```

Bind category/date controls to `activeRecordFilters`, use `recordCategoryOptions` for the record category select, and bind buttons to `queryRecords` / `resetRecordFilters`.

- [ ] **Step 8: Keep the filter layout responsive**

Rename `.workflow-center__application-filters` to `.workflow-center__record-filters` and use:

```scss
.workflow-center__record-filters {
  grid-template-columns: minmax(150px, 180px) minmax(300px, 430px) minmax(160px, 220px) auto;
}

.workflow-center__filter-select,
.workflow-center__filter-input {
  width: 100%;
  height: 36px;
  min-width: 0;
  box-sizing: border-box;
}
```

At `max-width: 900px`, use two columns and place actions across both columns. At `max-width: 768px`, use one column and retain the two equal-width action buttons.

- [ ] **Step 9: Run H5 checks and verify GREEN**

```bash
cd h5app
node scripts/check-workflow-module.mjs
node_modules/.bin/eslint src/types/workflow.ts src/api/workflow.ts src/pages/workflow/components/WorkflowCenter.vue scripts/check-workflow-module.mjs
node_modules/.bin/vue-tsc --noEmit
node_modules/.bin/uni build
```

Expected: all four commands exit 0.

### Task 4: Full regression and rendered verification

**Files:**
- Verify all files listed above; no additional planned source files.

- [ ] **Step 1: Run backend focused packages**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure ./internal/modules/workflow/transport/httpclient ./internal/routes/v2/dingtalkh5 ./internal/support/appapiperm ./cmd -count=1
```

Expected: PASS.

- [ ] **Step 2: Run the complete backend suite**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Run H5 verification**

```bash
cd h5app
node scripts/check-workflow-module.mjs
node_modules/.bin/eslint src/types/workflow.ts src/api/workflow.ts src/pages/workflow/components/WorkflowCenter.vue scripts/check-workflow-module.mjs
node_modules/.bin/vue-tsc --noEmit
node_modules/.bin/uni build
```

Also run `pnpm lint`. If the Codex pnpm wrapper or unrelated baseline files block it, report the exact failure and retain the successful focused lint result.

- [ ] **Step 4: Inspect the final worktree boundary**

```bash
git diff --check
git status --short
git -C h5app diff --check
git -C h5app status --short
```

Confirm only the files in this plan are attributed to this feature and preserve all unrelated staged/untracked changes.

- [ ] **Step 5: Verify the rendered workflow center**

Use the existing H5 development server when available; otherwise start one on an unused port. After login, verify:

1. 我的待办 shows category, submission time and starter-name inputs; combined filters update data and total before pagination.
2. 已处理 shows the same three inputs and preserves its own values after tab switches.
3. 我的申请 remains category plus submission time only; record categories come from the category endpoint rather than the startable definition list.
4. 抄送我的 shows category, submission time and approval status, with all six status options.
5. Query and reset return to page 1, while top navigation counts remain unfiltered.
6. Desktop and mobile layouts contain no overflow or overlapping controls.
7. Browser console has no new errors and one query interaction completes.

If authentication blocks target-page verification, do not synthesize login state; record the blocker and leave authenticated data verification to the user.
