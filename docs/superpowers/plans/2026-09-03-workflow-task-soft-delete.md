# Workflow Task Soft Delete Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add permission-controlled soft deletion for terminal workflow tasks in the Admin task list without removing runtime or audit data.

**Architecture:** Mirror the existing workflow-instance Admin soft-delete flow. The application service validates a locked task in a transaction, the GORM store writes task-level deletion audit fields, and only the Admin task list filters those rows; runtime state and instance detail loaders intentionally continue reading them.

**Tech Stack:** Go, Hertz, GORM, MySQL migration SQL, Vue 3, TypeScript, Element Plus, Node structural checks.

---

## File Map

- Create `backend/migrations/20260903133000_add_workflow_task_admin_delete.sql`: task audit columns and permission catalog rows.
- Create `backend/test/internal/bootstrap/migrations/workflow_task_admin_delete_migration_test.go`: migration contract.
- Modify `backend/internal/model/workflow/runtime.go`: persistence-only soft-delete fields.
- Modify `backend/internal/modules/workflow/application/types.go`: explicit Admin-list soft-delete filter flag.
- Modify `backend/internal/modules/workflow/application/service.go`: store contracts, errors, and terminal-state validation.
- Modify `backend/internal/modules/workflow/application/service_test.go`: service red/green coverage and fake store support.
- Modify `backend/internal/modules/workflow/infrastructure/gorm_store.go`: locked lookup, soft-delete update, and Admin list filter.
- Create `backend/internal/modules/workflow/infrastructure/task_admin_delete_filter_test.go`: DryRun SQL proof for list filtering.
- Modify `backend/internal/modules/workflow/transport/httpadmin/handler.go`: Admin task deletion handler.
- Modify `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`: authenticated actor and route ID contract.
- Modify `backend/internal/routes/v2/admin/routes.go`: DELETE route registration.
- Modify `backend/internal/middleware/admin/route_permissions.go`: route permission mapping.
- Modify `backend/internal/middleware/admin/permission_test.go`: concrete route lookup assertion.
- Modify `backend/internal/support/adminmenuperm/declarations.go`: button permission declaration.
- Modify `backend/internal/support/adminmenuperm/workflow_structure_test.go`: menu permission assertion.
- Modify `backend/internal/support/adminrouteperm/catalog.go`: API permission declaration.
- Modify `backend/internal/support/adminrouteperm/workflow_runtime_structure_test.go`: API permission assertion.
- Modify `admin/src/api/index.ts`: typed delete request.
- Modify `admin/src/views/workflow/tasks/index.vue`: permission-aware delete action and confirmation.
- Modify `admin/scripts/check-workflow-runtime-pages.mjs`: Admin API/UI structural checks.

### Task 1: Lock the migration and persistence contract

**Files:**
- Create: `backend/test/internal/bootstrap/migrations/workflow_task_admin_delete_migration_test.go`
- Create: `backend/migrations/20260903133000_add_workflow_task_admin_delete.sql`
- Modify: `backend/internal/model/workflow/runtime.go`

- [ ] **Step 1: Write the failing migration contract test**

```go
func TestWorkflowTaskAdminDeleteMigrationAddsAuditColumnsAndPermission(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_task_admin_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow task admin delete migration = %v, err = %v", matches, err)
	}
	text := readMigration(t, matches[0])
	for _, snippet := range []string{
		"workflow_process_tasks", "admin_deleted_at", "admin_deleted_by",
		"admin:menu:workflow:task:delete", "admin:api:workflow:task:delete",
		"workflow:task:delete", "/api/v2/admin/workflow-tasks/:id",
	} {
		if !strings.Contains(text, snippet) { t.Fatalf("migration missing %q", snippet) }
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("destructive permission must not be auto-granted")
	}
}
```

- [ ] **Step 2: Run the test and verify RED**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations -run TestWorkflowTaskAdminDeleteMigration -count=1`

Expected: FAIL because the task migration does not exist.

- [ ] **Step 3: Add the migration and model fields**

```sql
ALTER TABLE `workflow_process_tasks`
  ADD COLUMN `admin_deleted_at` bigint NOT NULL DEFAULT 0 COMMENT '管理员删除时间' AFTER `handled_at`,
  ADD COLUMN `admin_deleted_by` varchar(64) NOT NULL DEFAULT '' COMMENT '删除操作管理员ID' AFTER `admin_deleted_at`,
  ADD INDEX `idx_workflow_tasks_admin_deleted_time` (`admin_deleted_at`, `created_at`);
```

Insert `admin:menu:workflow:task:delete` and `admin:api:workflow:task:delete` with `workflow:task:delete`, using the same idempotent permission SQL pattern as the instance-delete migration and without grants.

```go
AdminDeletedAt int64  `json:"-" gorm:"column:admin_deleted_at;index:idx_workflow_tasks_admin_deleted_time,priority:1;comment:管理员删除时间"`
AdminDeletedBy string `json:"-" gorm:"size:64;column:admin_deleted_by;comment:删除操作管理员ID"`
```

- [ ] **Step 4: Run the migration test and model package tests**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations ./internal/model/workflow -count=1`

Expected: PASS.

### Task 2: Add transaction-safe application and storage behavior

**Files:**
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Create: `backend/internal/modules/workflow/infrastructure/task_admin_delete_filter_test.go`

- [ ] **Step 1: Write failing service tests**

```go
func TestDeleteTaskSoftDeletesTerminalTaskInTransaction(t *testing.T) {
	store := &fakeStore{deleteTask: &TaskSummary{ID: "task-1", Status: string(workflowdomain.TaskStatusCancelled)}}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
	service.now = func() time.Time { return time.UnixMilli(1788393600123) }

	err := service.DeleteTask(context.Background(), " 9 ", " task-1 ")
	if err != nil { t.Fatalf("DeleteTask() error = %v", err) }
	if store.softDeletedTaskID != "task-1" || store.softDeletedTaskBy != "9" || store.softDeletedTaskAt != 1788393600123 || !store.softDeleteTaskInTransaction {
		t.Fatalf("task delete audit not persisted in transaction")
	}
}

func TestDeleteTaskRejectsActiveOrMissingTaskWithoutWriting(t *testing.T) {
	tests := []struct {
		name string
		task *TaskSummary
		want error
	}{
		{name: "waiting", task: &TaskSummary{ID: "task-1", Status: string(workflowdomain.TaskStatusWaiting)}, want: ErrTaskDeleteNotAllowed},
		{name: "pending", task: &TaskSummary{ID: "task-1", Status: string(workflowdomain.TaskStatusPending)}, want: ErrTaskDeleteNotAllowed},
		{name: "missing", task: nil, want: ErrTaskDeleteTargetNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{deleteTask: test.task}
			service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})
			if err := service.DeleteTask(context.Background(), "9", "task-1"); !errors.Is(err, test.want) {
				t.Fatalf("DeleteTask() error = %v, want %v", err, test.want)
			}
			if store.softDeletedTaskID != "" { t.Fatalf("invalid delete wrote task %q", store.softDeletedTaskID) }
		})
	}
}
```

- [ ] **Step 2: Run the service tests and verify RED**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application -run TestDeleteTask -count=1`

Expected: FAIL because `DeleteTask` and task-delete store methods do not exist.

- [ ] **Step 3: Add minimal application behavior**

Add transaction-store methods:

```go
LoadTaskForDelete(ctx context.Context, taskID string) (*TaskSummary, error)
SoftDeleteTask(ctx context.Context, taskID, actorID string, deletedAt int64) (int64, error)
```

Implement:

```go
func (service *Service) DeleteTask(ctx context.Context, actorID, taskID string) error {
	actorID, taskID = strings.TrimSpace(actorID), strings.TrimSpace(taskID)
	if actorID == "" { return ErrActorRequired }
	if taskID == "" { return ErrTaskIDRequired }
	if service == nil || service.store == nil { return errors.New("工作流应用服务未初始化") }
	return service.store.InTransaction(ctx, func(store TransactionStore) error {
		task, err := store.LoadTaskForDelete(ctx, taskID)
		if err != nil { return err }
		if task == nil { return ErrTaskDeleteTargetNotFound }
		switch task.Status {
		case string(workflowdomain.TaskStatusCompleted), string(workflowdomain.TaskStatusApproved),
			string(workflowdomain.TaskStatusRejected), string(workflowdomain.TaskStatusCancelled):
		default:
			return ErrTaskDeleteNotAllowed
		}
		count, err := store.SoftDeleteTask(ctx, taskID, actorID, service.currentTime().UnixMilli())
		if err != nil { return err }
		if count != 1 { return ErrTaskDeleteTargetNotFound }
		return nil
	})
}
```

- [ ] **Step 4: Add the Admin list SQL test and verify RED**

```go
func TestApplyTaskFiltersHidesAdminDeletedTasksOnlyForAdminList(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: &sql.DB{}, SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil { t.Fatalf("open dry-run database: %v", err) }
	statement := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), application.TaskQuery{HideAdminDeleted: true}).Find(&[]workflowmodel.ProcessTask{}).Statement
	if !strings.Contains(statement.SQL.String(), "workflow_process_tasks.admin_deleted_at = ?") {
		t.Fatalf("Admin task query must hide soft-deleted rows: %s", statement.SQL.String())
	}
	if !reflect.DeepEqual(statement.Vars, []interface{}{int64(0)}) { t.Fatalf("unexpected vars: %#v", statement.Vars) }
	statement = applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), application.TaskQuery{}).Find(&[]workflowmodel.ProcessTask{}).Statement
	if strings.Contains(statement.SQL.String(), "workflow_process_tasks.admin_deleted_at") {
		t.Fatalf("runtime task query must retain audit rows: %s", statement.SQL.String())
	}
}
```

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure -run TestApplyTaskFiltersHidesAdminDeletedTasks -count=1`

Expected: FAIL because `applyTaskFilters` has no Admin delete predicate.

- [ ] **Step 5: Implement GORM methods and list filtering**

```go
func (store *GormStore) LoadTaskForDelete(ctx context.Context, taskID string) (*application.TaskSummary, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil { return nil, err }
	defer cancel()
	var row workflowmodel.ProcessTask
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "task_status").
		Where("id = ? AND admin_deleted_at = 0", strings.TrimSpace(taskID)).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
	if err != nil { return nil, err }
	return &application.TaskSummary{ID: row.ID, Status: row.Status}, nil
}

func (store *GormStore) SoftDeleteTask(ctx context.Context, taskID, actorID string, deletedAt int64) (int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil { return 0, err }
	defer cancel()
	result := db.Model(&workflowmodel.ProcessTask{}).
		Where("id = ? AND admin_deleted_at = 0 AND task_status IN ?", strings.TrimSpace(taskID), []string{
			workflowmodel.TaskStatusCompleted, workflowmodel.TaskStatusApproved,
			workflowmodel.TaskStatusRejected, workflowmodel.TaskStatusCancelled,
		}).Updates(map[string]interface{}{"admin_deleted_at": deletedAt, "admin_deleted_by": strings.TrimSpace(actorID)})
	return result.RowsAffected, result.Error
}
```

Add `HideAdminDeleted bool` to `TaskQuery`. When it is true, start `applyTaskFilters` with `db.Where("workflow_process_tasks.admin_deleted_at = ?", int64(0))`; otherwise do not apply that predicate. Set this flag only in the Admin HTTP task-list handler. Do not add the predicate to `ListMyTasks`, `loadStateRecords`, or `loadInstanceDetail`.

- [ ] **Step 6: Run application and infrastructure tests**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -count=1`

Expected: PASS.

### Task 3: Expose the Admin route and permission

**Files:**
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`
- Modify: `backend/internal/middleware/admin/permission_test.go`
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Modify: `backend/internal/support/adminmenuperm/workflow_structure_test.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/support/adminrouteperm/workflow_runtime_structure_test.go`

- [ ] **Step 1: Write failing handler and permission tests**

```go
func TestDeleteTaskUsesAuthenticatedAdminAndRouteTaskID(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "task-route"})
	handler.DeleteTask(context.Background(), c)
	if stub.deleteTaskCalls != 1 || stub.deleteTaskActorID != "42" || stub.deleteTaskID != "task-route" {
		t.Fatalf("delete task must use route ID and authenticated admin")
	}
}
```

Add `DELETE /api/v2/admin/workflow-tasks/wft_99 -> workflow:task:delete` to middleware tests, and task-delete expectations to both declaration structure tests.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/transport/httpadmin ./internal/middleware/admin ./internal/support/adminmenuperm ./internal/support/adminrouteperm -run 'Test(DeleteTask|AdminPermission|WorkflowRuntime)' -count=1`

Expected: FAIL because the handler, route mapping, and declarations are absent.

- [ ] **Step 3: Implement route, handler, and permission declarations**

```go
DeleteTask(context.Context, string, string) error
```

Register `admin.DELETE("/workflow-tasks/:id", runtimeHandler.DeleteTask)`, map it to `workflow:task:delete`, and return `{ "id": taskID }` after a successful handler call.

```go
func (handler *RuntimeHandler) DeleteTask(ctx context.Context, c *app.RequestContext) {
	actorID, ok := authenticatedActorID(c)
	if !ok { response.Fail(c, "未登录或权限失效"); return }
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" { response.Fail(c, "流程任务不能为空"); return }
	if err := handler.service.DeleteTask(ctx, actorID, taskID); err != nil { response.Fail(c, err.Error()); return }
	response.JSON(c, map[string]string{"id": taskID})
}
```

- [ ] **Step 4: Run focused permission and handler tests**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/transport/httpadmin ./internal/middleware/admin ./internal/support/adminmenuperm ./internal/support/adminrouteperm -count=1`

Expected: PASS.

### Task 4: Add the Admin delete interaction

**Files:**
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/views/workflow/tasks/index.vue`
- Modify: `admin/scripts/check-workflow-runtime-pages.mjs`

- [ ] **Step 1: Add failing structural expectations**

Require these snippets in the workflow runtime checker:

```js
'workflowTaskDelete(',
"hasPerm('admin:menu:workflow:task:delete')",
'canDeleteTask',
'deleteTask(row)',
'确认删除流程任务',
```

- [ ] **Step 2: Run the checker and verify RED**

Run: `cd admin && npm run check:workflow-runtime-pages`

Expected: FAIL on the first missing task-delete snippet.

- [ ] **Step 3: Add the API and page action**

```ts
workflowTaskDelete(id: ID) {
  return request.delete<{ id: string }>(`${ADMIN_V2}/workflow-tasks/${encodePath(id)}`)
}
```

```ts
const canDelete = computed(() => hasPerm('admin:menu:workflow:task:delete'))
const terminalTaskStatuses = new Set<WorkflowTaskStatus>(['completed', 'approved', 'rejected', 'cancelled'])
function canDeleteTask(row: WorkflowTaskSummary) {
  return canDelete.value && terminalTaskStatuses.has(row.status)
}
```

`deleteTask(row)` opens an Element Plus warning confirmation, calls the API, decrements the page when the deleted item was the last row on a non-first page, reloads the list, and shows `流程任务已删除`.

```ts
async function deleteTask(row: WorkflowTaskSummary) {
  try {
    await ElMessageBox.confirm(`删除后该任务将不再出现在管理列表中，确认删除“${row.nodeName || row.nodeId}”？`, '确认删除流程任务', {
      type: 'warning', confirmButtonText: '删除', confirmButtonClass: 'el-button--danger',
    })
  } catch { return }
  await adminApi.workflowTaskDelete(row.id)
  if (list.value.length === 1 && page.value > 1) page.value -= 1
  ElMessage.success('流程任务已删除')
  await loadList()
}
```

- [ ] **Step 4: Run Admin checks and build**

Run: `cd admin && npm run check:workflow-runtime-pages && npm run build`

Expected: PASS.

### Task 5: Verify the integrated behavior

**Files:**
- Verify only; do not add generated or temporary artifacts to the repository.

- [ ] **Step 1: Format and run focused backend tests**

Run: `cd backend && gofmt -w internal/model/workflow/runtime.go internal/modules/workflow/application/service.go internal/modules/workflow/application/service_test.go internal/modules/workflow/infrastructure/gorm_store.go internal/modules/workflow/infrastructure/task_admin_delete_filter_test.go internal/modules/workflow/transport/httpadmin/handler.go internal/modules/workflow/transport/httpadmin/handler_test.go internal/middleware/admin/route_permissions.go internal/middleware/admin/permission_test.go internal/support/adminmenuperm/declarations.go internal/support/adminmenuperm/workflow_structure_test.go internal/support/adminrouteperm/catalog.go internal/support/adminrouteperm/workflow_runtime_structure_test.go test/internal/bootstrap/migrations/workflow_task_admin_delete_migration_test.go && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/... ./internal/middleware/admin ./internal/support/adminmenuperm ./internal/support/adminrouteperm ./test/internal/bootstrap/migrations -count=1`

Expected: PASS.

- [ ] **Step 2: Run Admin checks**

Run: `cd admin && npm run check:workflow-runtime-pages && npm run build`

Expected: PASS.

- [ ] **Step 3: Browser smoke test**

The flow under test is: Admin workflow tasks -> terminal task delete -> confirmation -> success message -> row disappears; active tasks show only their existing processing action.

Verify page identity, non-blank content, no framework overlay, console health, delete-button visibility, confirmation cancellation, successful deletion, and a desktop screenshot. Do not use production-like data unless a disposable terminal task exists.

- [ ] **Step 4: Review the scoped diff**

Run: `git diff --check && git status --short`

Expected: no whitespace errors; report unrelated pre-existing changes separately.
