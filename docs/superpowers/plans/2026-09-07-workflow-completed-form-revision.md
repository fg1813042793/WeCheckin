# 已完成流程表单修订 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在通用工作流中交付“已完成流程由实际节点处理人发起表单修订、低风险字段直接生效、其他字段由原路径下游重新确认”的端到端能力。

**Architecture:** 已完成原实例始终保持终态；修订请求和修订任务使用独立聚合、独立数据表和独立状态机。流程定义只在节点 `postHandleEdit.completedRevision` 中声明能力，运行时从实例绑定的发布版本及实际历史路径生成不可变确认计划，并在事务内通过乐观锁应用候选表单、写审计、通知 Outbox 和业务事件 Outbox。

**Tech Stack:** Go 1.24、Hertz、GORM、MySQL、Vue 3、TypeScript strict、Element Plus、uni-app、uView Pro、Vitest/Go test、Vite H5。

---

## 实施口径补充

- 修订任务的 `approval_mode` 直接保存现有工作流枚举 `single/sequential/parallel/countersign`，并保存 `completion_rate`。这比另造 `all/any/sequential` 映射更能完整保留现有串行、并行和会签阈值语义；或签使用现有 countersign 低阈值规则。
- 修订预览是只读辅助接口，创建请求仍在事务内重新校验和解析，不能凭预览绕过并发、权限或组织关系变化。
- Admin 不新增第二个实例详情授权入口；现有 `workflow:instance:detail` 响应直接包含只读修订摘要，避免同一页面出现两套权限判断。

## 文件结构与职责

- `backend/internal/workflowcore/types.go`：定义完成后修订配置、能力和确认计划纯类型。
- `backend/internal/workflowcore/form_completed_revision.go`：完成后字段权限、补丁校验、直接字段判断和实际路径确认阶段构建。
- `backend/internal/workflowcore/form_completed_revision_test.go`：纯规则、串并行路径和条件/计算字段保护测试。
- `backend/internal/workflowcore/validation.go`：发布定义时校验完成后修订配置。
- `backend/internal/workflowcore/compiler.go`、`backend/internal/workflowcore/definition_test.go`：将完成后修订配置编译为 Flowable BPMN 扩展属性。
- `backend/internal/model/workflow/form_revision.go`：修订请求、修订任务和业务事件 Outbox 的 GORM 模型。
- `backend/migrations/20260907110000_create_workflow_completed_form_revisions.sql`：新增修订请求、任务和业务事件 Outbox 表及定时派发任务。
- `backend/migrations/20260907111000_add_workflow_completed_form_revision_permissions.sql`：只注册 H5 按钮/API 权限，不自动授权。
- `backend/test/internal/bootstrap/migrations/workflow_completed_form_revision_migration_test.go`：迁移、索引、权限和默认不授权保护。
- `backend/internal/modules/workflow/domain/form_revision.go`：修订聚合状态、任务推进及幂等领域规则。
- `backend/internal/modules/workflow/domain/form_revision_test.go`：修订状态机单元测试。
- `backend/internal/modules/workflow/application/form_completed_revision.go`：创建、取消、查询、处理和最终应用修订的事务编排。
- `backend/internal/modules/workflow/application/form_completed_revision_test.go`：身份、权限、直接生效、下游确认、冲突及事务行为测试。
- `backend/internal/modules/workflow/application/form_revision_capability.go`：实例详情暴露可选来源节点，不改变运行中修订能力。
- `backend/internal/modules/workflow/application/types.go`：H5/Admin 稳定 DTO，任务增加业务类型标识。
- `backend/internal/modules/workflow/application/service.go`：增加修订存储接口和稳定错误。
- `backend/internal/modules/workflow/application/events.go`：增加 `workflow.instance.form_revised` 业务事件负载。
- `backend/internal/modules/workflow/infrastructure/form_revision_store.go`：修订聚合持久化、锁、查询、统一待办和实例详情映射。
- `backend/internal/modules/workflow/infrastructure/form_revision_store_test.go`：GORM 映射、查询语义和并发保护测试。
- `backend/internal/modules/workflow/infrastructure/query_filters.go`：已处理范围纳入用户处理过的修订任务。
- `backend/internal/modules/workflow/infrastructure/task_store.go`：H5 请求时合并普通任务和修订任务，并保持统一分页顺序。
- `backend/internal/modules/workflow/infrastructure/overview_store.go`：待处理、已处理统计纳入修订任务。
- `backend/internal/modules/workflow/infrastructure/notification_store.go`：站内信使用修订详情 `sourceType/sourceId`。
- `backend/internal/modules/workflow/infrastructure/notification_channels.go`：钉钉 OA 跳转到修订详情动态页。
- `backend/internal/modules/workflow/infrastructure/business_event_store.go`：业务事件 Outbox 领取、成功和失败重试。
- `backend/internal/modules/scheduledtask/infrastructure/workflow_business_event_job.go`：taskd 派发工作流业务事件，不监听 HTTP。
- `backend/internal/modules/scheduledtask/infrastructure/wiring.go`、`backend/cmd/taskd/main.go`：注册业务事件派发任务。
- `backend/internal/modules/workflow/transport/httpclient/form_revision_handler.go`：H5 修订接口协议适配。
- `backend/internal/modules/workflow/transport/httpclient/form_revision_handler_test.go`：身份注入、参数校验和错误映射测试。
- `backend/internal/modules/workflow/transport/httpadmin/handler.go`：Admin 实例详情复用只读修订 DTO。
- `backend/internal/routes/v2/dingtalkh5/routes.go`：注册 H5 修订路由。
- `backend/internal/support/appapiperm/catalog.go`、`backend/internal/support/appmenuperm/catalog.go`：权限目录声明。
- `backend/internal/routes/v2/swagger/h5app.go`、`backend/internal/routes/v2/swagger/request_models.go`：Swagger 注解源。
- `docs/API_V2.md`：完成后修订接口、状态、幂等和错误语义。
- `admin/src/types/workflow.ts`：流程定义和修订只读 DTO。
- `admin/src/views/workflow/designer/components/NodeInspector.vue`：完成后修订开关及直接字段配置。
- `backend/internal/service/admin/workflow/version_diff.go`、`backend/internal/service/admin/workflow/version_diff_test.go`：版本变更摘要识别完成后修订配置变化。
- `admin/src/views/workflow/instances/components/WorkflowFormRevisionRecords.vue`：Admin 修订记录、差异和确认任务只读展示。
- `admin/src/views/workflow/instances/index.vue`：实例详情挂载修订记录组件。
- `admin/scripts/check-workflow-form-revision.mjs`：Admin 结构回归契约。
- `h5app/src/types/workflow.ts`：H5 修订请求、详情、任务和状态类型。
- `h5app/src/api/workflow.ts`：H5 修订 API。
- `h5app/src/pages/workflow/workflow-route-keys.ts`、`h5app/src/pages/workflow/workflow.routes.ts`：修订发起页和详情页动态路由。
- `h5app/src/pages/workflow/components/WorkflowCompletedFormRevisionPage.vue`：来源节点选择、表单编辑、差异、原因和确认路径预览。
- `h5app/src/pages/workflow/components/WorkflowFormRevisionDetailPage.vue`：修订状态、前后差异、确认任务、取消和处理。
- `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`：已完成实例入口和修订记录入口。
- `h5app/src/pages/workflow/components/WorkflowCenter.vue`：按任务类型打开普通任务或修订任务。
- `h5app/src/components/app-notification-panel/app-notification-panel.vue`、`h5app/src/pages/notifications/components/NotificationHistoryPage.vue`：修订通知深链。
- `h5app/src/pages/notifications/workflow-notification-open.ts`：通知面板和历史页共用的工作流实例/修订跳转决策。
- `h5app/scripts/check-workflow-completed-form-revision.mjs`：H5 结构回归契约。

### Task 1: 扩展流程定义与纯规则

**Files:**
- Modify: `backend/internal/workflowcore/types.go`
- Modify: `backend/internal/workflowcore/validation.go`
- Modify: `backend/internal/workflowcore/compiler.go`
- Modify: `backend/internal/workflowcore/definition_test.go`
- Create: `backend/internal/workflowcore/form_completed_revision.go`
- Create: `backend/internal/workflowcore/form_completed_revision_test.go`

- [ ] **Step 1: 写失败测试，固定节点级权限和直接字段规则**

```go
func TestCompletedRevisionCapabilityUsesOneHandledNode(t *testing.T) {
	definition := completedRevisionDefinition()
	capability, err := CompletedRevisionCapabilityForNode(definition, "manager")
	if err != nil {
		t.Fatal(err)
	}
	if capability.NodeID != "manager" || !reflect.DeepEqual(capability.DirectFields, []string{"remark"}) {
		t.Fatalf("capability = %#v", capability)
	}
	if permissionAccess(capability.FieldPermissions, "amount") != FieldAccessWrite {
		t.Fatalf("amount must remain review-required write field")
	}
}

func TestValidateCompletedRevisionPatchRejectsRoutingAndCalculationFields(t *testing.T) {
	definition := completedRevisionDefinition()
	for _, patch := range []map[string]interface{}{{"amount": 200}, {"total": 200}} {
		if err := ValidateCompletedRevisionPatch(definition, "manager", map[string]interface{}{}, patch); err == nil {
			t.Fatalf("patch %#v must fail", patch)
		}
	}
}
```

- [ ] **Step 2: 运行纯规则测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/workflowcore -run 'TestCompletedRevision' -count=1`

Expected: FAIL，提示 `CompletedRevisionCapabilityForNode` 或相关类型尚未定义。

- [ ] **Step 3: 增加配置和纯规则类型**

```go
type PostHandleEditConfig struct {
	Enabled           bool                     `json:"enabled"`
	CompletedRevision *CompletedRevisionConfig `json:"completedRevision,omitempty"`
}

type CompletedRevisionConfig struct {
	Enabled      bool     `json:"enabled"`
	DirectFields []string `json:"directFields,omitempty"`
}

type CompletedRevisionCapability struct {
	NodeID           string            `json:"nodeId"`
	NodeName         string            `json:"nodeName"`
	FieldPermissions []FieldPermission `json:"fieldPermissions"`
	DirectFields     []string          `json:"directFields"`
}

type RevisionPlanStage struct {
	Stage int      `json:"stage"`
	Nodes []string `json:"nodes"`
}
```

在 `form_completed_revision.go` 实现：

```go
func CompletedRevisionCapabilityForNode(definition Definition, nodeID string) (CompletedRevisionCapability, error)
func ValidateCompletedRevisionPatch(definition Definition, nodeID string, current, patch map[string]interface{}) error
func CompletedRevisionMode(capability CompletedRevisionCapability, patch map[string]interface{}) string
func BuildCompletedRevisionPlan(definition Definition, actualHumanNodeIDs []string, sourceNodeID string) ([]RevisionPlanStage, error)
```

规则必须按一个来源节点计算；条件字段和 calculation 字段降为只读；`directFields` 排序去重并只允许可写字段；路径构建只保留来源节点之后实际执行过的人工节点，并用拓扑层级表达串行/并行阶段。

- [ ] **Step 4: 将定义校验接入审批/办理节点**

```go
const ValidationCompletedRevision = "completed_revision_invalid"

func validateCompletedRevision(node Node, fields map[string]FormField, routingFields map[string]struct{}) []ValidationError {
	config := node.PostHandleEdit
	if config == nil || config.CompletedRevision == nil || !config.CompletedRevision.Enabled {
		return nil
	}
	if node.Type != NodeTypeApproval && node.Type != NodeTypeHandle {
		return []ValidationError{{Code: ValidationCompletedRevision, Message: "仅审批或办理节点支持流程完成后修订", NodeID: node.ID}}
	}
	writable := writablePermissionFields(node.FormPermissions)
	result := make([]ValidationError, 0)
	for _, fieldKey := range config.CompletedRevision.DirectFields {
		field, exists := fields[strings.TrimSpace(fieldKey)]
		_, canWrite := writable[strings.TrimSpace(fieldKey)]
		_, routesFlow := routingFields[strings.TrimSpace(fieldKey)]
		if !exists || !canWrite || routesFlow || field.Type == FormFieldTypeCalculation {
			result = append(result, ValidationError{Code: ValidationCompletedRevision, Message: "直接修订字段必须是非条件、非计算的可写数据字段", NodeID: node.ID})
		}
	}
	return result
}
```

`ValidateDefinition` 在完成边和字段收集后调用此校验，旧定义缺少 `completedRevision` 时不产生错误。

- [ ] **Step 5: 运行纯规则全包测试**

在 `definition_test.go` 增加 BPMN 断言：

```go
func TestCompileBPMNIncludesCompletedRevisionAttributes(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[1].PostHandleEdit = &PostHandleEditConfig{
		Enabled: true,
		CompletedRevision: &CompletedRevisionConfig{Enabled: true, DirectFields: []string{"remark", "contactPhone"}},
	}
	bpmn, err := CompileBPMN(definition)
	if err != nil { t.Fatal(err) }
	for _, token := range []string{
		`flowable:postHandleEditEnabled="true"`,
		`flowable:completedRevisionEnabled="true"`,
		`flowable:completedRevisionDirectFields="contactPhone,remark"`,
	} {
		if !strings.Contains(string(bpmn), token) { t.Fatalf("BPMN missing %s", token) }
	}
}
```

`compiler.go` 在 approval/handle 节点增加上述三个稳定扩展属性；`directFields` 排序后输出，未配置时不输出完成后属性。

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/workflowcore -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add backend/internal/workflowcore/types.go backend/internal/workflowcore/validation.go backend/internal/workflowcore/compiler.go backend/internal/workflowcore/definition_test.go backend/internal/workflowcore/form_completed_revision.go backend/internal/workflowcore/form_completed_revision_test.go
git commit -m "feat: 增加完成后表单修订规则"
```

### Task 2: 新增修订聚合状态机

**Files:**
- Create: `backend/internal/modules/workflow/domain/form_revision.go`
- Create: `backend/internal/modules/workflow/domain/form_revision_test.go`
- Modify: `backend/internal/modules/workflow/domain/runtime.go`

- [ ] **Step 1: 写失败测试，固定直接生效、分阶段激活、驳回和幂等**

```go
func TestRevisionApproveActivatesNextStageAndAppliesAtEnd(t *testing.T) {
	revision := pendingRevisionWithStages(1, 1)
	result, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionApprove, "同意", nil, 1000)
	if err != nil || result.Applied || revision.Tasks[1].Status != RevisionTaskStatusPending {
		t.Fatalf("first stage result=%#v revision=%#v err=%v", result, revision, err)
	}
	result, err = revision.CompleteTask("task-2-1", "user-2-1", RevisionActionApprove, "同意", nil, 2000)
	if err != nil || !result.Applied || revision.Status != FormRevisionStatusApplied {
		t.Fatalf("final result=%#v revision=%#v err=%v", result, revision, err)
	}
}

func TestRevisionRejectCancelsRemainingTasks(t *testing.T) {
	revision := pendingRevisionWithStages(2, 1)
	_, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionReject, "数据不正确", nil, 1000)
	if err != nil { t.Fatal(err) }
	if revision.Status != FormRevisionStatusRejected { t.Fatalf("status=%s", revision.Status) }
	for _, task := range revision.Tasks {
		if task.ID != "task-1-1" && task.Status != RevisionTaskStatusCancelled {
			t.Fatalf("remaining task=%#v", task)
		}
	}
}

func TestRevisionCompleteTaskIsIdempotent(t *testing.T) {
	revision := pendingRevisionWithStages(1)
	first, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionApprove, "同意", nil, 1000)
	if err != nil { t.Fatal(err) }
	second, err := revision.CompleteTask("task-1-1", "user-1-1", RevisionActionApprove, "同意", nil, 1000)
	if err != nil { t.Fatal(err) }
	if !first.Applied || !second.Idempotent || revision.Status != FormRevisionStatusApplied {
		t.Fatalf("first=%#v second=%#v revision=%#v", first, second, revision)
	}
}

func pendingRevisionWithStages(taskCounts ...int) *FormRevisionRequest {
	revision := &FormRevisionRequest{ID: "revision-1", Status: FormRevisionStatusPending}
	for stageIndex, count := range taskCounts {
		for taskIndex := 0; taskIndex < count; taskIndex++ {
			status := RevisionTaskStatusWaiting
			if stageIndex == 0 { status = RevisionTaskStatusPending }
			revision.Tasks = append(revision.Tasks, FormRevisionTask{
				ID: fmt.Sprintf("task-%d-%d", stageIndex+1, taskIndex+1),
				Stage: stageIndex + 1,
				AssigneeID: fmt.Sprintf("user-%d-%d", stageIndex+1, taskIndex+1),
				ApprovalMode: workflowcore.ApprovalModeParallel,
				CompletionRate: 100,
				Status: status,
			})
		}
	}
	return revision
}
```

- [ ] **Step 2: 运行领域测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/domain -run Revision -count=1`

Expected: FAIL，提示修订聚合尚未定义。

- [ ] **Step 3: 实现独立聚合和状态常量**

```go
type FormRevisionStatus string
const (
	FormRevisionStatusPending   FormRevisionStatus = "pending"
	FormRevisionStatusApplied   FormRevisionStatus = "applied"
	FormRevisionStatusRejected  FormRevisionStatus = "rejected"
	FormRevisionStatusCancelled FormRevisionStatus = "cancelled"
	FormRevisionStatusConflict  FormRevisionStatus = "conflict"
)

type FormRevisionRequest struct {
	ID, SourceInstanceID, SourceNodeID, SourceNodeName, RequesterID string
	BaseFormRevision, AppliedFormRevision int64
	BeforeFormData, Patch, ProposedFormData map[string]interface{}
	ChangedFieldLabels []string
	Reason, Mode string
	Status FormRevisionStatus
	Tasks []FormRevisionTask
	CreatedAt, CompletedAt int64
}

type FormRevisionTask struct {
	ID, RevisionRequestID, SourceTaskID, NodeID, NodeName string
	Stage int
	AssigneeID, AssigneeName, ApprovalMode string
	CompletionRate, Sequence, Total int
	Status RevisionTaskStatus
	Action RevisionAction
	Comment string
	Images []workflowcore.FormAttachment
	HandledBy string
	HandledAt int64
}

type CompleteRevisionTaskResult struct {
	Applied bool
	Idempotent bool
}
```

任务保留现有审批语义所需的 `ApprovalMode`、`CompletionRate`、`Sequence`、`Total` 和阶段；`single`、`sequential`、`parallel`、`countersign` 的推进规则与原工作流一致，但只改变修订聚合。

- [ ] **Step 4: 增加修订审计和通知种类常量**

```go
const (
	HistoryInstanceFormRevisionRequested HistoryEventType = "instance_form_revision_requested"
	HistoryInstanceFormRevisionApplied   HistoryEventType = "instance_form_revision_applied"
	HistoryInstanceFormRevisionRejected  HistoryEventType = "instance_form_revision_rejected"
	HistoryInstanceFormRevisionCancelled HistoryEventType = "instance_form_revision_cancelled"
	HistoryInstanceFormRevisionConflict  HistoryEventType = "instance_form_revision_conflict"
)
```

- [ ] **Step 5: 运行领域测试**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/domain -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add backend/internal/modules/workflow/domain/form_revision.go backend/internal/modules/workflow/domain/form_revision_test.go backend/internal/modules/workflow/domain/runtime.go
git commit -m "feat: 增加表单修订状态机"
```

### Task 3: 新增数据库模型、迁移和权限目录

**Files:**
- Create: `backend/internal/model/workflow/form_revision.go`
- Create: `backend/migrations/20260907110000_create_workflow_completed_form_revisions.sql`
- Create: `backend/migrations/20260907111000_add_workflow_completed_form_revision_permissions.sql`
- Create: `backend/test/internal/bootstrap/migrations/workflow_completed_form_revision_migration_test.go`
- Modify: `backend/internal/support/appapiperm/catalog.go`
- Modify: `backend/internal/support/appapiperm/catalog_test.go`
- Modify: `backend/internal/support/appmenuperm/catalog.go`
- Create: `backend/internal/support/appmenuperm/workflow_form_revision_test.go`

- [ ] **Step 1: 写失败迁移和权限目录测试**

```go
func TestWorkflowCompletedFormRevisionMigration(t *testing.T) {
	source := readWorkflowRevisionMigration(t, "*_create_workflow_completed_form_revisions.sql")
	for _, token := range []string{
		"workflow_form_revision_requests", "workflow_form_revision_tasks",
		"workflow_business_event_outbox", "uk_workflow_form_revision_active",
		"workflow.business-event.dispatch_due",
	} {
		if !strings.Contains(source, token) { t.Fatalf("migration missing %s", token) }
	}
}

func TestWorkflowCompletedRevisionPermissionsAreRegisteredWithoutRoleGrant(t *testing.T) {
	source := readWorkflowRevisionMigration(t, "*_add_workflow_completed_form_revision_permissions.sql")
	for _, key := range []string{
		"dingtalk_h5:button:workflow:form-revision-create",
		"dingtalk_h5:api:workflow:form-revision-create",
		"dingtalk_h5:button:workflow:form-revision-handle",
		"dingtalk_h5:api:workflow:form-revision-handle",
	} {
		if !strings.Contains(source, key) { t.Fatalf("permission migration missing %s", key) }
	}
	if strings.Contains(strings.ToLower(source), "role_permission_grants") {
		t.Fatal("completed revision permissions must not be granted automatically")
	}
}

func readWorkflowRevisionMigration(t *testing.T, pattern string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", pattern))
	if err != nil || len(matches) != 1 { t.Fatalf("migration matches=%v err=%v", matches, err) }
	source, err := os.ReadFile(matches[0])
	if err != nil { t.Fatal(err) }
	return string(source)
}
```

- [ ] **Step 2: 运行迁移/目录测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations ./internal/support/appapiperm ./internal/support/appmenuperm -run 'CompletedFormRevision|CompletedRevision' -count=1`

Expected: FAIL，提示迁移和权限声明不存在。

- [ ] **Step 3: 创建三张表和定时任务**

迁移必须创建：

```sql
CREATE TABLE IF NOT EXISTS `workflow_form_revision_requests` (
  `id` varchar(64) NOT NULL,
  `source_instance_id` varchar(64) NOT NULL,
  `source_node_id` varchar(100) NOT NULL,
  `source_node_name` varchar(200) NOT NULL DEFAULT '',
  `requester_id` varchar(64) NOT NULL,
  `base_form_revision` bigint NOT NULL,
  `before_form_data_json` mediumtext NOT NULL,
  `patch_json` mediumtext NOT NULL,
  `proposed_form_data_json` mediumtext NOT NULL,
  `changed_field_labels_json` mediumtext NOT NULL,
  `reason` varchar(500) NOT NULL,
  `revision_mode` varchar(32) NOT NULL,
  `revision_status` varchar(24) NOT NULL,
  `active_lock_key` varchar(64) DEFAULT NULL,
  `applied_form_revision` bigint NOT NULL DEFAULT 0,
  `add_time` bigint NOT NULL,
  `edit_time` bigint NOT NULL,
  `completed_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_workflow_form_revision_active` (`active_lock_key`),
  KEY `idx_workflow_form_revision_instance_time` (`source_instance_id`,`add_time`),
  KEY `idx_workflow_form_revision_requester_time` (`requester_id`,`add_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

`workflow_form_revision_tasks` 包含请求 ID、原任务 ID、节点快照、阶段、处理人快照、审批方式、完成比例、顺序、状态、意见、图片和处理时间，并为 `(assignee_id, task_status, add_time)`、`(revision_request_id, stage, task_status)` 建索引。

`workflow_business_event_outbox` 包含 `event_type`、`aggregate_id`、`business_type`、`business_key`、`payload_json`、`event_status`、`attempts`、`next_retry_at`、`last_error`、`sent_at` 和唯一 `dedupe_key`；迁移使用现有 `scheduled_tasks` 幂等写入 `workflow.business-event.dispatch_due`。

- [ ] **Step 4: 注册四个 H5 权限但不授予历史角色**

```text
dingtalk_h5:button:workflow:form-revision-create
dingtalk_h5:api:workflow:form-revision-create
dingtalk_h5:button:workflow:form-revision-handle
dingtalk_h5:api:workflow:form-revision-handle
```

API 目录中创建权限绑定 `POST /api/v2/dingtalk/h5/workflows/instances/:id/form-revisions`，处理权限绑定 `POST /api/v2/dingtalk/h5/workflows/form-revision-tasks/:id/complete`；GET 修订记录/详情复用 `workflow:view`，取消复用创建权限。

- [ ] **Step 5: 增加 GORM 模型并运行格式化与测试**

Run: `cd backend && gofmt -w internal/model/workflow/form_revision.go test/internal/bootstrap/migrations/workflow_completed_form_revision_migration_test.go internal/support/appapiperm/catalog.go internal/support/appapiperm/catalog_test.go internal/support/appmenuperm/catalog.go internal/support/appmenuperm/workflow_form_revision_test.go`

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations ./internal/support/appapiperm ./internal/support/appmenuperm -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add backend/internal/model/workflow/form_revision.go backend/migrations/20260907110000_create_workflow_completed_form_revisions.sql backend/migrations/20260907111000_add_workflow_completed_form_revision_permissions.sql backend/test/internal/bootstrap/migrations/workflow_completed_form_revision_migration_test.go backend/internal/support/appapiperm backend/internal/support/appmenuperm
git commit -m "feat: 持久化完成后表单修订"
```

### Task 4: 实现修订存储和映射

**Files:**
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/types.go`
- Create: `backend/internal/modules/workflow/infrastructure/form_revision_store.go`
- Create: `backend/internal/modules/workflow/infrastructure/form_revision_store_test.go`

- [ ] **Step 1: 写失败测试，固定锁、活动请求唯一性和完整映射**

```go
func TestFormRevisionStoreLoadsInstanceAndActiveRequestForUpdate(t *testing.T) {
	db := dryRunWorkflowDB(t)
	instanceSQL := lockCompletedRevisionSourceQuery(db, "instance-1").Statement.SQL.String()
	requestSQL := lockActiveFormRevisionQuery(db, "instance-1").Statement.SQL.String()
	for name, sqlText := range map[string]string{"instance": instanceSQL, "request": requestSQL} {
		if !strings.Contains(strings.ToUpper(sqlText), "FOR UPDATE") {
			t.Fatalf("%s lock query=%s", name, sqlText)
		}
	}
}

func TestFormRevisionRowsRoundTrip(t *testing.T) {
	want := completedRevisionFixture()
	requestRow, taskRows, err := formRevisionModels(want)
	if err != nil { t.Fatal(err) }
	got, err := formRevisionFromModels(requestRow, taskRows)
	if err != nil || !reflect.DeepEqual(got, want) { t.Fatalf("got=%#v err=%v", got, err) }
}
```

- [ ] **Step 2: 运行存储测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure -run FormRevision -count=1`

Expected: FAIL，提示映射/存储方法不存在。

- [ ] **Step 3: 扩展应用存储接口**

```go
type FormRevisionStore interface {
	ListFormRevisions(context.Context, string, string) ([]FormRevisionSummary, error)
	GetFormRevision(context.Context, string, string) (*FormRevisionDetail, error)
}

type FormRevisionTransactionStore interface {
	LoadCompletedRevisionSourceForUpdate(context.Context, string) (workflowcore.Definition, *workflowdomain.State, error)
	LoadActiveFormRevisionForUpdate(context.Context, string) (*workflowdomain.FormRevisionRequest, error)
	LoadFormRevisionForUpdate(context.Context, string) (*workflowdomain.FormRevisionRequest, error)
	LoadFormRevisionByTaskForUpdate(context.Context, string) (*workflowdomain.FormRevisionRequest, error)
	CreateFormRevision(context.Context, *workflowdomain.FormRevisionRequest) error
	SaveFormRevision(context.Context, *workflowdomain.FormRevisionRequest) error
	PersistRevisionNotifications(context.Context, *workflowdomain.State, *workflowdomain.FormRevisionRequest, []workflowdomain.NotificationIntent) ([]string, error)
	CreateBusinessEvent(context.Context, WorkflowBusinessEvent) error
}
```

让 `Store`/`TransactionStore` 嵌入这些接口；测试 fake 必须显式实现，不能在应用层对 GORM 做类型断言。

- [ ] **Step 4: 实现 JSON 映射和查询**

`form_revision_store.go` 负责：

```go
func formRevisionModels(*workflowdomain.FormRevisionRequest) (workflowmodel.FormRevisionRequest, []workflowmodel.FormRevisionTask, error)
func formRevisionFromModels(workflowmodel.FormRevisionRequest, []workflowmodel.FormRevisionTask) (*workflowdomain.FormRevisionRequest, error)
func (store *GormStore) CreateFormRevision(context.Context, *workflowdomain.FormRevisionRequest) error
func (store *GormStore) SaveFormRevision(context.Context, *workflowdomain.FormRevisionRequest) error
func (store *GormStore) ListFormRevisions(context.Context, string, string) ([]application.FormRevisionSummary, error)
func (store *GormStore) GetFormRevision(context.Context, string, string) (*application.FormRevisionDetail, error)
```

所有请求链路使用 `contextDB`；创建依赖 `active_lock_key=source_instance_id` 唯一索引；进入终态时更新为 `NULL`；任务保存按主键更新状态/意见/图片/处理时间，不删除历史行。

- [ ] **Step 5: 运行存储测试**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure -run 'FormRevision|RevisionRows' -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add backend/internal/modules/workflow/application/service.go backend/internal/modules/workflow/application/types.go backend/internal/modules/workflow/infrastructure/form_revision_store.go backend/internal/modules/workflow/infrastructure/form_revision_store_test.go
git commit -m "feat: 增加表单修订存储"
```

### Task 5: 实现修订能力、创建和直接生效

**Files:**
- Modify: `backend/internal/modules/workflow/application/form_revision_capability.go`
- Create: `backend/internal/modules/workflow/application/form_completed_revision.go`
- Create: `backend/internal/modules/workflow/application/form_completed_revision_test.go`
- Modify: `backend/internal/modules/workflow/application/instance.go`
- Modify: `backend/internal/modules/workflow/application/service.go`

- [ ] **Step 1: 写失败测试，覆盖身份、单节点权限和直接修订**

```go
func TestCreateCompletedFormRevisionDoesNotGrantStarterIdentity(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})
	_, err := service.CreateCompletedFormRevision(context.Background(), CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: state.Instance.StarterID,
		ExpectedRevision: 3, FormData: map[string]interface{}{"remark": "越权"}, Reason: "尝试修改",
	})
	if !errors.Is(err, ErrCompletedRevisionNotAllowed) { t.Fatalf("err=%v", err) }
}

func TestCreateCompletedFormRevisionDoesNotMergeHandledNodePermissions(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	state.History = append(state.History,
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "manager", ActorID: "21"},
		workflowdomain.HistoryEvent{Type: workflowdomain.HistoryTaskApproved, NodeID: "hr", ActorID: "21"},
	)
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})
	_, err := service.CreateCompletedFormRevision(context.Background(), CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: "21",
		ExpectedRevision: 3, FormData: map[string]interface{}{"hrComment": "不能跨节点"}, Reason: "权限测试",
	})
	if !errors.Is(err, workflowcore.ErrFormDataInvalid) { t.Fatalf("err=%v", err) }
}

func TestCreateDirectCompletedFormRevisionAppliesAtomically(t *testing.T) {
	definition, state := completedRevisionApplicationFixture()
	state.History = append(state.History, workflowdomain.HistoryEvent{
		Type: workflowdomain.HistoryTaskApproved, NodeID: "manager", ActorID: "21",
	})
	store := &fakeStore{definition: definition, state: state}
	service := NewService(store, fixedResolver{"99"}, &sequenceIDs{})
	result, err := service.CreateCompletedFormRevision(context.Background(), CreateCompletedFormRevisionRequest{
		InstanceID: "instance-1", SourceNodeID: "manager", RequesterID: "21",
		ExpectedRevision: 3, FormData: map[string]interface{}{"remark": "已更正"}, Reason: "修正说明",
	})
	if err != nil { t.Fatal(err) }
	if result.Status != "applied" || result.AppliedFormRevision != 4 { t.Fatalf("result=%#v", result) }
	if store.savedState.FormData["remark"] != "已更正" || store.createdFormRevision == nil {
		t.Fatalf("state=%#v revision=%#v", store.savedState, store.createdFormRevision)
	}
}

func completedRevisionApplicationFixture() (workflowcore.Definition, *workflowdomain.State) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion, Key: "completed_revision", Name: "完成后修订",
		Form: []workflowcore.FormField{
			{Key: "remark", Label: "说明", Type: workflowcore.FormFieldTypeTextarea},
			{Key: "hrComment", Label: "HR 意见", Type: workflowcore.FormFieldTypeTextarea},
		},
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "manager", Type: workflowcore.NodeTypeApproval, Name: "主管审批",
				PostHandleEdit: &workflowcore.PostHandleEditConfig{CompletedRevision: &workflowcore.CompletedRevisionConfig{Enabled: true, DirectFields: []string{"remark"}}},
				FormPermissions: []workflowcore.FieldPermission{{Field: "remark", Access: workflowcore.FieldAccessWrite}}},
			{ID: "hr", Type: workflowcore.NodeTypeApproval, Name: "HR 审批",
				PostHandleEdit: &workflowcore.PostHandleEditConfig{CompletedRevision: &workflowcore.CompletedRevisionConfig{Enabled: true}},
				FormPermissions: []workflowcore.FieldPermission{{Field: "hrComment", Access: workflowcore.FieldAccessWrite}}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{{ID: "e1", Source: "start", Target: "manager"}, {ID: "e2", Source: "manager", Target: "hr"}, {ID: "e3", Source: "hr", Target: "end"}},
	}
	return definition, &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{ID: "instance-1", StarterID: "7", Status: workflowdomain.InstanceStatusCompleted, FormRevision: 3},
		FormData: map[string]interface{}{"remark": "原说明", "hrComment": "原意见"},
	}
}
```

- [ ] **Step 2: 运行应用测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application -run 'CompletedFormRevision|DirectCompleted' -count=1`

Expected: FAIL，提示创建方法和 DTO 尚未定义。

- [ ] **Step 3: 扩展实例详情能力 DTO**

```go
type FormRevisionCapability struct {
	Allowed                  bool                                  `json:"allowed"`
	Revision                 int64                                 `json:"revision"`
	FieldPermissions         []workflowcore.FieldPermission        `json:"fieldPermissions"`
	RunningRevisionAllowed   bool                                  `json:"runningRevisionAllowed"`
	CompletedRevisionNodes   []workflowcore.CompletedRevisionCapability `json:"completedRevisionNodes"`
}
```

运行中实例继续填充旧 `allowed/fieldPermissions`；已完成实例只根据当前用户的 `task_approved/task_submitted/task_returned` 实际历史事件返回 `completedRevisionNodes`，默认按最后处理时间倒序，不因发起人身份开放。

- [ ] **Step 4: 实现事务内创建和直接应用**

```go
func (service *Service) PreviewCompletedFormRevision(ctx context.Context, request PreviewCompletedFormRevisionRequest) (*FormRevisionPreview, error)
func (service *Service) CreateCompletedFormRevision(ctx context.Context, request CreateCompletedFormRevisionRequest) (*FormRevisionDetail, error)
```

抽出无写入的共享规划函数供预览和创建调用。实现顺序固定为：规范化 ID/原因/补丁 -> 加载实例和发布版本 -> 校验 `completed`、`expectedRevision`、无活动请求 -> 校验来源节点实际处理历史 -> `ValidateCompletedRevisionPatch` -> 合并并 `ApplyFormCalculations` -> 生成差异/标签 -> 判定 `direct` 或 `downstream_review` -> 解析确认阶段；创建入口必须在事务内重新执行并使用 `FOR UPDATE`，不能复用客户端提交的预览。

仅直接字段时，同一事务中写请求、更新原实例 `form_data_json/form_revision`、写 `instance_form_revision_requested` 和 `instance_form_revision_applied` 历史、释放活动锁并写通知/业务事件；任何一步失败都回滚。

- [ ] **Step 5: 运行应用测试**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application -run 'CompletedFormRevision|DirectCompleted|FormRevisionCapability' -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add backend/internal/modules/workflow/application/form_revision_capability.go backend/internal/modules/workflow/application/form_completed_revision.go backend/internal/modules/workflow/application/form_completed_revision_test.go backend/internal/modules/workflow/application/instance.go backend/internal/modules/workflow/application/service.go
git commit -m "feat: 支持发起完成后表单修订"
```

### Task 6: 实现下游确认、取消和版本冲突

**Files:**
- Modify: `backend/internal/modules/workflow/application/form_completed_revision.go`
- Modify: `backend/internal/modules/workflow/application/form_completed_revision_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/form_revision_store.go`

- [ ] **Step 1: 写失败测试，覆盖串行、并行、会签、或签、取消和冲突**

```go
func TestCreateReviewedRevisionRebuildsActualDownstreamStages(t *testing.T) {
	service, store := reviewedRevisionServiceFixture(t)
	result, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil { t.Fatal(err) }
	if result.Status != "pending" || len(store.createdFormRevision.Tasks) != 3 {
		t.Fatalf("result=%#v revision=%#v", result, store.createdFormRevision)
	}
	if got := revisionTaskNodeStages(store.createdFormRevision.Tasks); got != "finance:1:pending,legal:1:pending,hr:2:waiting" {
		t.Fatalf("stages=%s", got)
	}
}

func TestCompleteRevisionTaskAppliesOnlyAfterRequiredApprovals(t *testing.T) {
	service, store := reviewedRevisionServiceFixture(t)
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil { t.Fatal(err) }
	for _, taskID := range []string{"revision-task-1", "revision-task-2"} {
		result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{TaskID: taskID, ActorID: "99", Action: "approve"})
		if err != nil || result.Status != "pending" { t.Fatalf("task=%s result=%#v err=%v", taskID, result, err) }
	}
	result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{TaskID: "revision-task-3", ActorID: "99", Action: "approve"})
	if err != nil || result.Status != "applied" || store.savedState.Instance.FormRevision != 4 {
		t.Fatalf("result=%#v state=%#v err=%v", result, store.savedState, err)
	}
}

func TestRejectRevisionKeepsSourceFormUntouched(t *testing.T) {
	service, store := reviewedRevisionServiceFixture(t)
	before := cloneFormData(store.state.FormData)
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil { t.Fatal(err) }
	result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{TaskID: "revision-task-1", ActorID: "99", Action: "reject", Comment: "数据不正确"})
	if err != nil || result.Status != "rejected" || !reflect.DeepEqual(store.state.FormData, before) {
		t.Fatalf("result=%#v form=%#v err=%v", result, store.state.FormData, err)
	}
}

func TestCancelRevisionAllowedBeforeFirstHandledTask(t *testing.T) {
	service, _ := reviewedRevisionServiceFixture(t)
	created, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil { t.Fatal(err) }
	cancelled, err := service.CancelCompletedFormRevision(context.Background(), CancelCompletedFormRevisionRequest{RevisionID: created.ID, ActorID: "21"})
	if err != nil || cancelled.Status != "cancelled" { t.Fatalf("cancelled=%#v err=%v", cancelled, err) }

	service, _ = reviewedRevisionServiceFixture(t)
	created, err = service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest())
	if err != nil { t.Fatal(err) }
	if _, err = service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{TaskID: "revision-task-1", ActorID: "99", Action: "approve"}); err != nil { t.Fatal(err) }
	_, err = service.CancelCompletedFormRevision(context.Background(), CancelCompletedFormRevisionRequest{RevisionID: created.ID, ActorID: "21"})
	if !errors.Is(err, ErrCompletedRevisionCannotCancel) { t.Fatalf("err=%v", err) }
}

func TestApplyRevisionMarksConflictWhenSourceRevisionChanged(t *testing.T) {
	service, store := reviewedRevisionServiceFixture(t)
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil { t.Fatal(err) }
	store.state.Instance.FormRevision = 4
	approveAllButLastRevisionTasks(t, service)
	result, err := service.CompleteFormRevisionTask(context.Background(), CompleteFormRevisionTaskRequest{TaskID: "revision-task-3", ActorID: "99", Action: "approve"})
	if err != nil || result.Status != "conflict" || store.state.Instance.FormRevision != 4 {
		t.Fatalf("result=%#v revision=%d err=%v", result, store.state.Instance.FormRevision, err)
	}
}

func TestCompleteRevisionTaskIsIdempotent(t *testing.T) {
	service, store := reviewedRevisionServiceFixture(t)
	if _, err := service.CreateCompletedFormRevision(context.Background(), reviewedRevisionRequest()); err != nil { t.Fatal(err) }
	request := CompleteFormRevisionTaskRequest{TaskID: "revision-task-1", ActorID: "99", Action: "approve"}
	if _, err := service.CompleteFormRevisionTask(context.Background(), request); err != nil { t.Fatal(err) }
	historyCount, notificationCount := len(store.state.History), len(store.persistedRevisionNotifications)
	result, err := service.CompleteFormRevisionTask(context.Background(), request)
	if err != nil || !result.Idempotent || len(store.state.History) != historyCount || len(store.persistedRevisionNotifications) != notificationCount {
		t.Fatalf("result=%#v history=%d notifications=%d err=%v", result, len(store.state.History), len(store.persistedRevisionNotifications), err)
	}
}
```

在同一测试文件实现固定夹具：`reviewedRevisionServiceFixture` 构造 `manager -> parallel split -> finance/legal -> join -> hr -> end`，历史只包含 `manager/finance/legal/hr`，另放一个未经过的 `audit` 分支；`fixedResolver{"99"}` 为三个确认节点返回处理人；`sequenceIDs` 固定生成 `revision-task-1..3`。`reviewedRevisionRequest` 修改来源节点 manager 的非直接字段并使用 `expectedRevision=3`；`approveAllButLastRevisionTasks` 依次通过前两项；`revisionTaskNodeStages` 按 stage/node 排序生成测试中的精确字符串。

- [ ] **Step 2: 运行应用测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application -run 'ReviewedRevision|RevisionTask|CancelRevision|RevisionConflict' -count=1`

Expected: FAIL，提示处理/取消方法尚未完成。

- [ ] **Step 3: 创建时解析原实际路径和当前处理人**

对 `BuildCompletedRevisionPlan` 的每个节点调用现有 `AssigneeResolver.Resolve`，传入原实例的 starter/operator/variables 和发布节点配置；解析失败、空处理人、路径不稳定均拒绝创建。任务使用当前组织解析结果快照，首阶段按原审批模式激活，其余状态为 `waiting`。

- [ ] **Step 4: 实现处理和取消事务**

```go
func (service *Service) CompleteFormRevisionTask(ctx context.Context, request CompleteFormRevisionTaskRequest) (*FormRevisionDetail, error)
func (service *Service) CancelCompletedFormRevision(ctx context.Context, request CancelCompletedFormRevisionRequest) (*FormRevisionDetail, error)
```

处理时锁修订请求、全部任务和源实例；校验任务 `pending` 与 assignee；通过时按组/阶段推进，驳回时取消剩余任务且不改原表单。最终通过前再次比较 `base_form_revision`；一致才应用候选表单并递增版本，不一致写 `conflict` 历史并释放活动锁。

- [ ] **Step 5: 运行应用与存储测试**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -run 'Revision' -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add backend/internal/modules/workflow/application/form_completed_revision.go backend/internal/modules/workflow/application/form_completed_revision_test.go backend/internal/modules/workflow/infrastructure/form_revision_store.go
git commit -m "feat: 支持表单修订下游确认"
```

### Task 7: 接入通知深链和业务事件 Outbox

**Files:**
- Modify: `backend/internal/modules/workflow/application/events.go`
- Modify: `backend/internal/modules/workflow/application/events_test.go`
- Modify: `backend/internal/modules/workflow/application/notification.go`
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/infrastructure/notification_store.go`
- Modify: `backend/internal/modules/workflow/infrastructure/notification_channels.go`
- Modify: `backend/internal/modules/workflow/infrastructure/notification_dispatcher_test.go`
- Create: `backend/internal/modules/workflow/infrastructure/business_event_store.go`
- Create: `backend/internal/modules/workflow/infrastructure/business_event_store_test.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/workflow_business_event_job.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/workflow_business_event_job_test.go`
- Modify: `backend/internal/modules/scheduledtask/infrastructure/wiring.go`
- Modify: `backend/cmd/taskd/main.go`

- [ ] **Step 1: 写失败测试，固定通知目标和业务事件负载**

```go
func TestRevisionNotificationUsesRevisionDeepLink(t *testing.T) {
	notify := workflowInAppNotify(revisionNotificationRecord(), 1000)
	if notify.SourceType != "workflow_form_revision" || notify.SourceID != "revision-1" {
		t.Fatalf("source=%s/%s", notify.SourceType, notify.SourceID)
	}
}

func TestFormRevisedBusinessEventCarriesBusinessReferenceAndPatch(t *testing.T) {
	event := NewFormRevisedBusinessEvent(instanceFixture(), "revision-1", 4, map[string]interface{}{"remark": "x"})
	if event.Type != LifecycleInstanceFormRevised || event.BusinessType != "purchase" || event.BusinessKey != "PO-1" {
		t.Fatalf("event=%#v", event)
	}
}
```

- [ ] **Step 2: 运行通知/事件测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure ./internal/modules/scheduledtask/infrastructure -run 'RevisionNotification|FormRevisedBusiness|WorkflowBusinessEvent' -count=1`

Expected: FAIL，提示深链负载或业务事件派发器不存在。

- [ ] **Step 3: 扩展通知负载而不修改通知表结构**

```go
type NotificationPayload struct {
	MessageType string `json:"messageType"`
	Title string `json:"title"`
	Content string `json:"content"`
	StarterID string `json:"starterId"`
	SourceType string `json:"sourceType,omitempty"`
	SourceID string `json:"sourceId,omitempty"`
	View string `json:"view,omitempty"`
}
```

修订通知写 `sourceType=workflow_form_revision`、`sourceId=<revision id>`、`view=workflow:form-revision-detail:<revision id>`；普通工作流通知保持 `workflow_instance` 回退。钉钉 URL 构建读取 `View`，站内信读取 `SourceType/SourceID`。

- [ ] **Step 4: 增加五类通知及事务内 Outbox 写入**

```text
instance_form_revision_requested
instance_form_revision_approved
instance_form_revision_rejected
instance_form_revision_cancelled
instance_form_revision_conflict
```

请求通知首阶段处理人；生效通知修订人、原发起人和已确认参与人并去重；驳回/冲突通知修订人和原发起人；取消通知已收到未处理待办的人。使用修订 ID + 事件 + 接收人 + 渠道作为 dedupe 后缀。

- [ ] **Step 5: 实现可重试业务事件 Outbox**

```go
const LifecycleInstanceFormRevised LifecycleEventType = "workflow.instance.form_revised"

type LifecycleEvent struct {
	FormRevision int64                  `json:"formRevision,omitempty"`
	RevisionRequestID string            `json:"revisionRequestId,omitempty"`
	FormPatch map[string]interface{}    `json:"formPatch,omitempty"`
}
```

`business_event_store.go` 使用租约领取 `pending/failed` 事件，调用可返回错误的 `LifecycleEventBus.Dispatch`，成功标记 `sent`，失败按 30s/60s/120s/300s/600s 上限退避并在第五次后 `dead`。`workflow_business_event_job.go` 注册 key `workflow.business-event.dispatch_due`；`taskd` 只运行已有 scheduler/worker，不增加 HTTP 监听。

- [ ] **Step 6: 运行通知、事件和 taskd 测试**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure ./internal/modules/scheduledtask/infrastructure ./cmd/taskd -count=1`

Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add backend/internal/modules/workflow/application/events.go backend/internal/modules/workflow/application/events_test.go backend/internal/modules/workflow/application/notification.go backend/internal/modules/workflow/application/types.go backend/internal/modules/workflow/infrastructure/notification_store.go backend/internal/modules/workflow/infrastructure/notification_channels.go backend/internal/modules/workflow/infrastructure/notification_dispatcher_test.go backend/internal/modules/workflow/infrastructure/business_event_store.go backend/internal/modules/workflow/infrastructure/business_event_store_test.go backend/internal/modules/scheduledtask/infrastructure/workflow_business_event_job.go backend/internal/modules/scheduledtask/infrastructure/workflow_business_event_job_test.go backend/internal/modules/scheduledtask/infrastructure/wiring.go backend/cmd/taskd/main.go
git commit -m "feat: 派发表单修订通知和业务事件"
```

### Task 8: 统一待办、已处理、总览和实例详情

**Files:**
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/application/notification_service.go`
- Modify: `backend/internal/modules/workflow/infrastructure/task_store.go`
- Modify: `backend/internal/modules/workflow/infrastructure/query_filters.go`
- Modify: `backend/internal/modules/workflow/infrastructure/overview_store.go`
- Modify: `backend/internal/modules/workflow/infrastructure/overview_store_test.go`
- Create: `backend/internal/modules/workflow/infrastructure/form_revision_query_integration_test.go`

- [ ] **Step 1: 写失败查询测试**

```go
func TestMyPendingTasksIncludesFormRevisionTasks(t *testing.T) {
	store := seededRevisionQueryStore(t)
	result, err := store.ListTasks(context.Background(), application.TaskQuery{
		AssigneeID: "42", Status: workflowmodel.TaskStatusPending,
		IncludeFormRevisions: true, Page: 1, PageSize: 20,
	})
	if err != nil { t.Fatal(err) }
	if result.Total != 2 || result.List[0].TaskType != "form_revision" || result.List[0].RevisionRequestID != "revision-1" {
		t.Fatalf("result=%#v", result)
	}
}

func TestHandledInstancesIncludesHandledRevisionTasks(t *testing.T) {
	store := seededRevisionQueryStore(t)
	result, err := store.ListInstances(context.Background(), application.InstanceQuery{
		ActorID: "42", Scope: application.InstanceScopeHandled, Page: 1, PageSize: 20,
	})
	if err != nil { t.Fatal(err) }
	if result.Total != 1 || result.List[0].ID != "instance-1" { t.Fatalf("result=%#v", result) }
}

func TestOverviewIncludesRevisionPendingAndHandledWithoutDoubleCountingInstances(t *testing.T) {
	store := seededRevisionQueryStore(t)
	overview, err := store.GetWorkflowOverview(context.Background(), "42")
	if err != nil { t.Fatal(err) }
	if overview.Pending != 2 || overview.Handled != 1 {
		t.Fatalf("overview=%#v", overview)
	}
}
```

`seededRevisionQueryStore` 使用测试数据库创建一个源实例、一个普通 pending 任务、一个 revision pending 任务，并为同一实例创建一个用户已处理的普通任务和一个已处理修订任务；修订任务 `add_time` 晚于普通任务，从而同时验证 UNION 排序和 handled 按实例去重。

- [ ] **Step 2: 运行查询测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure -run 'Revision.*Pending|HandledRevision|OverviewIncludesRevision' -count=1`

Expected: FAIL。

- [ ] **Step 3: 扩展任务 DTO 和 H5 查询开关**

```go
type TaskSummary struct {
	TaskType string `json:"taskType"`
	RevisionRequestID string `json:"revisionRequestId,omitempty"`
}

type TaskQuery struct {
	IncludeFormRevisions bool `json:"-"`
}
```

H5 `ListMyTasks` 设置 `IncludeFormRevisions=true`，Admin 默认 false，避免 Admin 的普通任务处理入口误处理修订任务。

- [ ] **Step 4: 使用 UNION ALL 统一 H5 任务分页**

普通任务和修订任务投影到同一列集，先执行统一 count，再按 `sort_time DESC, id DESC` 分页；字段筛选作用于两类任务的源实例定义名、标题和发起人。修订任务只返回当前用户的 `pending` 或终态处理记录。

- [ ] **Step 5: 修改已处理和总览 SQL**

`scope=handled` 的 EXISTS 增加修订任务终态；overview 的 pending 为普通 pending + 修订 pending，handled 使用实例 ID 的 UNION 后 COUNT，避免同一实例普通任务和修订任务重复计数。

- [ ] **Step 6: 在实例详情返回修订记录**

```go
type InstanceDetail struct {
	FormRevisions []FormRevisionSummary `json:"formRevisions"`
}
```

Admin/H5 获取实例详情时按 `add_time DESC` 返回修订摘要；无历史时返回空数组而不是 `null`。

- [ ] **Step 7: 运行查询测试并提交**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure ./internal/modules/workflow/application -count=1`

Expected: PASS。

```bash
git add backend/internal/modules/workflow/application/types.go backend/internal/modules/workflow/application/notification_service.go backend/internal/modules/workflow/infrastructure/task_store.go backend/internal/modules/workflow/infrastructure/query_filters.go backend/internal/modules/workflow/infrastructure/overview_store.go backend/internal/modules/workflow/infrastructure/overview_store_test.go backend/internal/modules/workflow/infrastructure/form_revision_query_integration_test.go
git commit -m "feat: 将表单修订接入流程工作台"
```

### Task 9: 增加 H5 HTTP 接口、错误语义和 Swagger

**Files:**
- Create: `backend/internal/modules/workflow/transport/httpclient/form_revision_handler.go`
- Create: `backend/internal/modules/workflow/transport/httpclient/form_revision_handler_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/handler.go`
- Modify: `backend/internal/modules/workflow/transport/httperror/error.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/routes.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/workflow_routes_test.go`
- Modify: `backend/internal/routes/v2/swagger/h5app.go`
- Modify: `backend/internal/routes/v2/swagger/request_models.go`
- Modify: `docs/API_V2.md`

- [ ] **Step 1: 写失败 handler 与路由测试**

```go
func TestCreateCompletedFormRevisionUsesAuthenticatedActor(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})
	c.Request.SetBodyString(`{"sourceNodeId":"manager","expectedRevision":3,"formData":{"remark":"x"},"reason":"修正"}`)
	handler.CreateCompletedFormRevision(context.Background(), c)
	if stub.createRevisionRequest.RequesterID != "42" || stub.createRevisionRequest.InstanceID != "instance-1" {
		t.Fatalf("request=%#v", stub.createRevisionRequest)
	}
}

func TestCompleteFormRevisionTaskRejectsInvalidAction(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "revision-task-1"})
	c.Request.SetBodyString(`{"action":"return","comment":"invalid"}`)
	handler.CompleteFormRevisionTask(context.Background(), c)
	if c.Response.StatusCode() != 400 || stub.completeRevisionTaskCalls != 0 {
		t.Fatalf("status=%d calls=%d body=%s", c.Response.StatusCode(), stub.completeRevisionTaskCalls, c.Response.Body())
	}
}

func TestCancelCompletedFormRevisionMapsActiveTaskError(t *testing.T) {
	stub := &runtimeServiceStub{cancelRevisionErr: workflowapp.ErrCompletedRevisionCannotCancel}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "revision-1"})
	handler.CancelCompletedFormRevision(context.Background(), c)
	if c.Response.StatusCode() != 400 || !strings.Contains(string(c.Response.Body()), "已有确认任务被处理") {
		t.Fatalf("status=%d body=%s", c.Response.StatusCode(), c.Response.Body())
	}
}
```

`runtimeServiceStub` 增加 `createRevisionRequest`、`completeRevisionTaskRequest`、`cancelRevisionRequest`、`completeRevisionTaskCalls` 和 `cancelRevisionErr`，并为新增 service interface 的列表、预览、创建、详情、取消、处理方法提供确定返回值。

- [ ] **Step 2: 运行 transport/route 测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/transport/httpclient ./internal/routes/v2/dingtalkh5 -run 'FormRevision|WorkflowRoutes' -count=1`

Expected: FAIL，提示路由或 handler 不存在。

- [ ] **Step 3: 实现命名 DTO 和五个接口**

```go
type previewCompletedFormRevisionBody struct {
	SourceNodeID string `json:"sourceNodeId"`
	ExpectedRevision int64 `json:"expectedRevision"`
	FormData map[string]interface{} `json:"formData"`
}

type createCompletedFormRevisionBody struct {
	SourceNodeID string `json:"sourceNodeId"`
	ExpectedRevision int64 `json:"expectedRevision"`
	FormData map[string]interface{} `json:"formData"`
	Reason string `json:"reason"`
}

type completeFormRevisionTaskBody struct {
	Action string `json:"action"`
	Comment string `json:"comment"`
	Images []workflowcore.FormAttachment `json:"images"`
}
```

注册六个接口：

```text
GET  /workflows/instances/:id/form-revisions
POST /workflows/instances/:id/form-revisions/preview
POST /workflows/instances/:id/form-revisions
GET  /workflows/form-revisions/:id
POST /workflows/form-revisions/:id/cancel
POST /workflows/form-revision-tasks/:id/complete
```

预览接口执行与创建一致的状态、身份、字段、路径和处理人解析，但不写数据库，并返回 `mode`、`changedFields` 和 `confirmationStages`；创建接口在事务内完整重算，不信任预览结果。所有 path 参数使用 `routeparam`；身份只从 DingTalk H5 认证上下文读取；列表/详情同时校验实例参与权或修订任务归属。

- [ ] **Step 4: 补错误映射、权限路由声明和 API 文档**

400：参数、非法字段、无确认路径、不可取消；403：不是实际处理人/任务处理人；404：实例、请求或任务不存在；409：活动请求、版本冲突或重复终态冲突。预览路由在 `DingTalkH5RouteDeclarations` 中复用 `form-revision-create` 权限。`docs/API_V2.md` 给出完整请求/响应 JSON 和状态枚举。

- [ ] **Step 5: 生成 Swagger，不手改生成文件**

Run: `cd backend && swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger`

Expected: 更新 `backend/docs/swagger/docs.go`、`swagger.json`、`swagger.yaml`。

- [ ] **Step 6: 运行路由、权限和 Swagger 测试并提交**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/transport/httpclient ./internal/routes/v2/dingtalkh5 ./internal/support/appapiperm ./internal/routes/v2/swagger -count=1`

Expected: PASS。

```bash
git add backend/internal/modules/workflow/transport/httpclient backend/internal/modules/workflow/transport/httperror/error.go backend/internal/routes/v2/dingtalkh5 backend/internal/routes/v2/swagger backend/internal/support/appapiperm docs/API_V2.md backend/docs/swagger
git commit -m "feat: 提供完成后表单修订接口"
```

### Task 10: Admin 配置完成后修订

**Files:**
- Modify: `admin/src/types/workflow.ts`
- Modify: `admin/src/views/workflow/designer/components/NodeInspector.vue`
- Modify: `admin/src/views/workflow/designer/workflowFieldCatalog.ts`
- Modify: `admin/scripts/check-workflow-form-revision.mjs`
- Modify: `backend/internal/service/admin/workflow/version_diff.go`
- Modify: `backend/internal/service/admin/workflow/version_diff_test.go`

- [ ] **Step 1: 先扩展 Admin 结构契约检查并确认失败**

```js
for (const snippet of [
  'WorkflowCompletedRevisionConfig',
  'completedRevision?: WorkflowCompletedRevisionConfig',
  '流程完成后允许修订',
  '允许直接修订的低风险字段',
  'updateCompletedRevisionEnabled',
  'updateCompletedRevisionDirectFields',
]) {
  if (!typeSource.includes(snippet) && !inspectorSource.includes(snippet))
    throw new Error(`completed revision designer contract missing: ${snippet}`)
}
```

Run: `cd admin && node scripts/check-workflow-form-revision.mjs`

Expected: FAIL。

- [ ] **Step 2: 增加严格类型并保留旧定义兼容**

```ts
export interface WorkflowCompletedRevisionConfig {
  enabled: boolean
  directFields: string[]
}

export interface WorkflowPostHandleEditConfig {
  enabled: boolean
  completedRevision?: WorkflowCompletedRevisionConfig
}
```

- [ ] **Step 3: 在 NodeInspector 增加配置控件**

只有 approval/handle 节点显示。开启后列出该节点 `write` 字段；条件字段和 calculation 字段保持可见但 checkbox disabled，并通过 tooltip 说明“用于流程路由”或“计算字段不可直接修改”。关闭完成后修订只删除 `completedRevision`，不得改变运行中 `enabled`。

```ts
function updateCompletedRevisionEnabled(value: boolean) {
  const postHandleEdit = ensurePostHandleEdit()
  if (value) postHandleEdit.completedRevision = { enabled: true, directFields: [] }
  else delete postHandleEdit.completedRevision
  emit('change')
}
```

- [ ] **Step 4: 运行 Admin 定向检查和构建**

先在 `version_diff_test.go` 增加定义前后仅 `completedRevision.enabled/directFields` 变化的用例，断言变更项包含节点名、“完成后修订”和直接字段标签；再让 `version_diff.go` 按字段 key 排序生成稳定摘要，旧定义缺少配置时按 disabled/空列表处理。

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/service/admin/workflow -run VersionDiff -count=1`

Expected: PASS。

Run: `cd admin && node scripts/check-workflow-form-revision.mjs && npm run build`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add admin/src/types/workflow.ts admin/src/views/workflow/designer/components/NodeInspector.vue admin/src/views/workflow/designer/workflowFieldCatalog.ts admin/scripts/check-workflow-form-revision.mjs backend/internal/service/admin/workflow/version_diff.go backend/internal/service/admin/workflow/version_diff_test.go
git commit -m "feat: 配置流程完成后表单修订"
```

### Task 11: Admin 只读查看修订记录

**Files:**
- Modify: `admin/src/types/workflow.ts`
- Create: `admin/src/views/workflow/instances/components/WorkflowFormRevisionRecords.vue`
- Modify: `admin/src/views/workflow/instances/index.vue`
- Modify: `admin/scripts/check-workflow-form-revision.mjs`

- [ ] **Step 1: 扩展结构契约检查并确认失败**

```js
const recordsSource = read('src/views/workflow/instances/components/WorkflowFormRevisionRecords.vue')
for (const snippet of ['字段差异', '修改原因', '确认记录', '通知投递']) {
  if (!recordsSource.includes(snippet)) throw new Error(`revision records missing ${snippet}`)
}
```

Run: `cd admin && node scripts/check-workflow-form-revision.mjs`

Expected: FAIL，组件不存在。

- [ ] **Step 2: 增加只读 DTO 和展示组件**

```ts
export interface WorkflowFormRevisionSummary {
  id: string
  sourceNodeId: string
  sourceNodeName: string
  requesterId: string
  requesterName: string
  reason: string
  mode: 'direct' | 'downstream_review'
  status: 'pending' | 'applied' | 'rejected' | 'cancelled' | 'conflict'
  changedFieldLabels: string[]
  addTime: number
  completedAt: number
}
```

组件使用折叠列表显示状态、修订人、来源节点、原因、字段 before/after 和任务时间线；所有表单值使用现有只读样式，不渲染可编辑按钮。

- [ ] **Step 3: 挂载到实例详情并关联通知记录**

`instances/index.vue` 在流程表单后增加“表单修订”折叠区，传入 `detail.formRevisions`、`detail.form` 和当前实例通知记录；无数据时显示“暂无修订记录”。

- [ ] **Step 4: 运行 Admin 检查**

Run: `cd admin && node scripts/check-workflow-form-revision.mjs && npm run check:all`

Expected: PASS；若仓库既有基线失败，记录与本次文件无关的具体失败，且本任务定向检查和 build 必须通过。

- [ ] **Step 5: 提交**

```bash
git add admin/src/types/workflow.ts admin/src/views/workflow/instances/components/WorkflowFormRevisionRecords.vue admin/src/views/workflow/instances/index.vue admin/scripts/check-workflow-form-revision.mjs
git commit -m "feat: 查看流程表单修订记录"
```

### Task 12: H5App 增加类型、API 和动态路由

**Files:**
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/src/api/workflow.ts`
- Modify: `h5app/src/pages/workflow/workflow-route-keys.ts`
- Modify: `h5app/src/pages/workflow/workflow.routes.ts`
- Create: `h5app/scripts/check-workflow-completed-form-revision.mjs`
- Modify: `h5app/package.json`

- [ ] **Step 1: 写失败结构契约检查**

```js
const expectations = [
  ['src/types/workflow.ts', ['WorkflowCompletedRevisionNode', 'WorkflowFormRevisionDetail', "taskType: 'workflow' | 'form_revision'"]],
  ['src/api/workflow.ts', ['previewWorkflowFormRevision', 'createWorkflowFormRevision', 'getWorkflowFormRevision', 'completeWorkflowFormRevisionTask']],
  ['src/pages/workflow/workflow-route-keys.ts', ['workflowCompletedFormRevisionContentKey', 'workflowFormRevisionDetailContentKey']],
]
```

Run: `cd h5app && node scripts/check-workflow-completed-form-revision.mjs`

Expected: FAIL。

- [ ] **Step 2: 增加严格类型和 API**

```ts
export interface WorkflowCompletedRevisionNode {
  nodeId: string
  nodeName: string
  fieldPermissions: WorkflowFieldPermission[]
  directFields: string[]
}

export interface WorkflowCreateFormRevisionRequest {
  sourceNodeId: string
  expectedRevision: number
  formData: WorkflowFormData
  reason: string
}
```

API 方法覆盖列表、预览、创建、详情、取消和任务完成六个后端接口；所有动态 ID 使用 `encodeURIComponent`。

- [ ] **Step 3: 增加两个动态路由键**

```ts
workflow:completed-form-revision:<instanceId>
workflow:form-revision-detail:<revisionId>
```

`normalizeWorkflowDynamicContentKey`、resolver 和 tab 恢复都必须识别；修订详情仅以 revisionId 路由，不在 URL 泄露表单数据。

- [ ] **Step 4: 运行类型和路由契约检查**

Run: `cd h5app && pnpm check:workflow-completed-form-revision && pnpm type-check`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add h5app/src/types/workflow.ts h5app/src/api/workflow.ts h5app/src/pages/workflow/workflow-route-keys.ts h5app/src/pages/workflow/workflow.routes.ts h5app/scripts/check-workflow-completed-form-revision.mjs h5app/package.json
git commit -m "feat: 增加 H5 表单修订契约"
```

### Task 13: H5App 实现修订发起和详情处理

**Files:**
- Create: `h5app/src/pages/workflow/components/WorkflowCompletedFormRevisionPage.vue`
- Create: `h5app/src/pages/workflow/components/WorkflowFormRevisionDetailPage.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`
- Modify: `h5app/scripts/check-workflow-completed-form-revision.mjs`

- [ ] **Step 1: 扩展失败结构检查**

```js
for (const snippet of ['选择来源节点', '修改字段', '修改原因', '提交后立即生效', '需要重新确认']) {
  requireSnippet('src/pages/workflow/components/WorkflowCompletedFormRevisionPage.vue', snippet)
}
for (const snippet of ['修改前', '修改后', '确认记录', '取消修订', '通过', '驳回']) {
  requireSnippet('src/pages/workflow/components/WorkflowFormRevisionDetailPage.vue', snippet)
}
```

Run: `cd h5app && pnpm check:workflow-completed-form-revision`

Expected: FAIL。

- [ ] **Step 2: 实现修订发起页**

加载实例详情后默认选择 `completedRevisionNodes[0]`，允许切换来源节点；切换时重置可编辑快照。复用 `WorkflowRuntimeForm` 和现有字段权限工具，只展示来源节点可见字段，只提交实际变化字段。必须填写 1-500 字原因。

用户点击提交前调用 `previewWorkflowFormRevision`，由后端返回最终 `mode`、变更字段和确认阶段：全部直接字段显示“提交后立即生效”；含其他字段显示确认节点和处理人，并在确认按钮文案中明确“提交修订申请”。注册 tab 关闭保护，真正提交仍携带相同 patch/expectedRevision，成功后打开修订详情。

- [ ] **Step 3: 实现修订详情和任务处理**

详情页展示状态、原流程入口、来源节点、原因、字段前后值、确认时间线。当前用户存在 pending 修订任务且有 handle 按钮/API 权限时显示固定底部“通过/驳回”；驳回必须填写原因，通过可填写意见。修订发起人拥有 create 按钮/API 权限且首任务尚未处理时看到“取消修订”。所有提交按钮带 loading 并阻止重复点击。

- [ ] **Step 4: 在原流程详情增加入口**

已完成实例且 `completedRevisionNodes.length > 0` 且拥有创建按钮/API 权限时显示“申请修订”。存在修订历史时显示“修订记录”，点击最近记录进入详情；运行中的“修改表单”入口保持原行为。

- [ ] **Step 5: 运行定向检查、lint 和类型检查**

Run: `cd h5app && pnpm check:workflow-completed-form-revision && pnpm check:workflow-module && pnpm lint && pnpm type-check`

Expected: 本次定向检查、workflow 检查和 type-check PASS；lint 若仍有既有基线错误，记录数量和首个与本次无关的错误，本次新增文件不得新增 lint 错误。

- [ ] **Step 6: 提交**

```bash
git add h5app/src/pages/workflow/components/WorkflowCompletedFormRevisionPage.vue h5app/src/pages/workflow/components/WorkflowFormRevisionDetailPage.vue h5app/src/pages/workflow/components/WorkflowDetailPanel.vue h5app/scripts/check-workflow-completed-form-revision.mjs
git commit -m "feat: 实现 H5 表单修订交互"
```

### Task 14: H5App 接入待办、已处理、总览和通知深链

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowCenter.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowTaskPage.vue`
- Modify: `h5app/src/pages/workflow/workflow-task.ts`
- Modify: `h5app/src/components/app-notification-panel/app-notification-panel.vue`
- Modify: `h5app/src/pages/notifications/components/NotificationHistoryPage.vue`
- Create: `h5app/src/pages/notifications/workflow-notification-open.ts`
- Modify: `h5app/scripts/check-workflow-completed-form-revision.mjs`

- [ ] **Step 1: 写失败结构检查**

```js
requireSnippet('src/pages/workflow/components/WorkflowCenter.vue', "task.taskType === 'form_revision'")
requireSnippet('src/components/app-notification-panel/app-notification-panel.vue', "sourceType === 'workflow_form_revision'")
requireSnippet('src/pages/notifications/components/NotificationHistoryPage.vue', 'workflowFormRevisionDetailContentKey')
```

Run: `cd h5app && pnpm check:workflow-completed-form-revision`

Expected: FAIL。

- [ ] **Step 2: 按任务类型打开正确详情**

`WorkflowCenter` 的 pending 行遇到 `taskType=form_revision` 时使用 `workflowFormRevisionDetailContentKey(task.revisionRequestId)`；普通任务继续使用 `workflowTaskContentKey`。修订任务标题显示“表单修订 · <来源节点/确认节点>”，状态颜色复用统一 workflow status 元数据。

- [ ] **Step 3: 保持已处理和总览刷新一致**

处理、驳回、取消或直接修订成功后调用 `requestRefresh()` 并刷新 `/workflows/overview`；切回浏览器前台和点击卡片仍沿用现有刷新机制。已处理页由后端扩展后的实例范围返回，不在前端拼接第二套分页。

- [ ] **Step 4: 接入五类通知和深链**

`workflow-notification-open.ts` 暴露纯函数，把 `workflow_form_revision` 的 `sourceId` 转换为 `workflowFormRevisionDetailContentKey`，把 `workflow_instance` 转换为原实例 key；通知面板和历史页共同调用，避免两处路由判断漂移。标签分别映射“修订待确认、修订已生效、修订已驳回、修订已取消、修订冲突”；无权限或记录不存在时保留通知详情并显示可理解错误。

- [ ] **Step 5: 运行 H5 完整静态验证和构建**

Run: `cd h5app && pnpm check:workflow-completed-form-revision && pnpm check:workflow-module && pnpm check:ui-style && pnpm type-check && pnpm build:h5`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add h5app/src/pages/workflow/components/WorkflowCenter.vue h5app/src/pages/workflow/components/WorkflowTaskPage.vue h5app/src/pages/workflow/workflow-task.ts h5app/src/components/app-notification-panel/app-notification-panel.vue h5app/src/pages/notifications/components/NotificationHistoryPage.vue h5app/src/pages/notifications/workflow-notification-open.ts h5app/scripts/check-workflow-completed-form-revision.mjs
git commit -m "feat: 接入表单修订待办和通知"
```

### Task 15: 全链路回归与浏览器验收

**Files:**
- Modify only if a regression is found: files already listed in Tasks 1-14
- Test: `backend/internal/modules/workflow/...`
- Test: `admin/scripts/check-workflow-form-revision.mjs`
- Test: `h5app/scripts/check-workflow-completed-form-revision.mjs`

- [ ] **Step 1: 运行后端定向和全量测试**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/workflowcore ./internal/modules/workflow/... ./internal/modules/scheduledtask/infrastructure ./internal/routes/v2/dingtalkh5 ./internal/support/appapiperm ./internal/support/appmenuperm ./test/internal/bootstrap/migrations -count=1`

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./... -count=1`

Expected: PASS；若全量存在任务开始前已确认的基线失败，保存命令、包名和错误，定向包必须 PASS。

- [ ] **Step 2: 运行 Admin 和 H5 门禁**

Run: `cd admin && npm run check:all`

Run: `cd h5app && pnpm lint && pnpm type-check && pnpm check:workflow-module && pnpm check:ui-style && pnpm check:workflow-completed-form-revision && pnpm build:h5`

Expected: 本次定向检查、构建和类型检查 PASS；任何基线失败与本次新增失败分开记录。

- [ ] **Step 3: 运行根级复核**

Run: `bash scripts/verify-local.sh`

Expected: Backend、Admin、Frontend 均通过，或只保留已确认且与本次无关的基线失败。

- [ ] **Step 4: 使用浏览器验证桌面、紧凑桌面和手机**

启动 H5 dev server 后验证至少以下场景：

```text
1. 旧定义/未授权用户看不到申请修订按钮。
2. 实际处理人可选择来源节点，字段不会跨节点合并。
3. 直接字段提交后原表单版本递增，实例仍为 completed。
4. 需确认字段生成待办；通过、驳回、取消、冲突状态正确。
5. 待办/已处理/总览数字在操作后刷新。
6. 站内信和钉钉跳转参数打开修订详情。
7. Admin 可配置并回显 directFields，可只读查看差异和任务。
8. 1366x768、1024x768、390x844 无横向溢出，弹层层级正确。
```

对三个视口保存截图，并检查 `document.documentElement.scrollWidth <= document.documentElement.clientWidth`。

- [ ] **Step 5: 检查改动边界和提交状态**

Run: `git status --short`

Run: `git diff --check`

Expected: 不包含绩效业务文件和无关格式化；所有本功能改动已按任务提交，用户原有未提交文件保持原状。

- [ ] **Step 6: 交付人工验收说明**

列出迁移执行顺序、新权限默认未授权、taskd 需要运行、已验证命令、浏览器截图和未覆盖的真实钉钉 OA/生产组织关系风险；最终页面与真实钉钉设备验收由用户执行。
