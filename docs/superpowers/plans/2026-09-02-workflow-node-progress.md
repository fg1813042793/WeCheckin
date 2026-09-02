# Workflow Node Progress Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在通用流程实例详情中返回实例历史版本的全部节点进度，并在钉钉 H5 流转记录中展示五态节点和节点下任务明细。

**Architecture:** 后端 application 层新增纯节点进度投影函数，输入发布定义、实例、令牌、任务和历史事件，输出稳定的 `nodeProgress` DTO；infrastructure 在加载实例绑定版本后调用该函数。H5 新增独立节点进度列表组件，历史抽屉和办理详情复用该组件，旧后端缺少 `nodeProgress` 时回退已有任务数据。

**Tech Stack:** Go 1.24、workflowcore、GORM 映射层、Vue 3、TypeScript、uni-app、uView Pro、SCSS、Node 结构回归脚本

---

## File Map

- Create `backend/internal/modules/workflow/application/node_progress.go`: 节点排序、执行证据收集、网关必经路径判断和五态投影。
- Create `backend/internal/modules/workflow/application/node_progress_test.go`: 线性、条件、并行和终止状态的纯函数测试。
- Modify `backend/internal/modules/workflow/application/types.go`: 声明节点进度状态、DTO，并扩展 `InstanceDetail`。
- Modify `backend/internal/modules/workflow/infrastructure/gorm_store.go`: 使用实例绑定的定义版本生成 `nodeProgress`。
- Modify `docs/architecture/generic-oa-workflow-v1.md`: 记录详情接口的历史版本节点进度契约。
- Modify `h5app/src/types/workflow.ts`: 增加节点进度类型和可选详情字段。
- Create `h5app/src/pages/workflow/components/WorkflowNodeProgressList.vue`: 全节点纵向进度和节点任务明细。
- Modify `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`: 历史抽屉与历史页接入统一进度组件。
- Modify `h5app/scripts/check-workflow-module.mjs`: 锁定全节点进度契约和组件接线。

> 当前工作区相关现有文件已经包含未提交修改。实现阶段不自动提交这些共享文件；每个任务以聚焦测试和 `git diff --check` 作为检查点，避免把用户改动带入提交。

### Task 1: Define the backend progress contract with a failing test

**Files:**
- Create: `backend/internal/modules/workflow/application/node_progress_test.go`
- Modify: `backend/internal/modules/workflow/application/types.go`

- [ ] **Step 1: Write the failing linear-flow test**

Add a test that describes the public projection API before implementing it:

```go
func TestBuildNodeProgressLinearFlow(t *testing.T) {
    definition := workflowcore.Definition{
        Nodes: []workflowcore.Node{
            {ID: "start", Type: workflowcore.NodeTypeStart, Name: "发起"},
            {ID: "approve", Type: workflowcore.NodeTypeApproval, Name: "主管审批"},
            {ID: "handle", Type: workflowcore.NodeTypeHandle, Name: "归档"},
            {ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
        },
        Edges: []workflowcore.Edge{
            {ID: "e1", Source: "start", Target: "approve"},
            {ID: "e2", Source: "approve", Target: "handle"},
            {ID: "e3", Source: "handle", Target: "end"},
        },
    }
    actual := BuildNodeProgress(
        definition,
        InstanceSummary{Status: "running", StartTime: 1000},
        []TokenSummary{{ID: "token-1", NodeID: "handle", Status: "waiting"}},
        []TaskSummary{
            {ID: "task-1", NodeID: "approve", Status: "approved"},
            {ID: "task-2", NodeID: "handle", Status: "pending"},
        },
        nil,
    )

    assertNodeProgress(t, actual, []nodeProgressExpectation{
        {nodeID: "start", status: NodeProgressCompleted},
        {nodeID: "approve", status: NodeProgressCompleted},
        {nodeID: "handle", status: NodeProgressProcessing},
        {nodeID: "end", status: NodeProgressNotStarted},
    })
}
```

Use this test helper to compare array length, node order and status:

```go
type nodeProgressExpectation struct {
    nodeID string
    status string
}

func assertNodeProgress(t *testing.T, actual []NodeProgressSummary, expected []nodeProgressExpectation) {
    t.Helper()
    if len(actual) != len(expected) {
        t.Fatalf("progress length = %d, want %d: %#v", len(actual), len(expected), actual)
    }
    for index := range expected {
        if actual[index].NodeID != expected[index].nodeID || actual[index].Status != expected[index].status {
            t.Fatalf("progress[%d] = (%q, %q), want (%q, %q)", index, actual[index].NodeID, actual[index].Status, expected[index].nodeID, expected[index].status)
        }
    }
}
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application -run TestBuildNodeProgressLinearFlow -count=1
```

Expected: compilation fails because `BuildNodeProgress` and progress constants do not exist.

- [ ] **Step 3: Add the response types only**

Extend `application/types.go` with:

```go
const (
    NodeProgressCompleted  = "completed"
    NodeProgressProcessing = "processing"
    NodeProgressNotStarted = "not_started"
    NodeProgressSkipped    = "skipped"
    NodeProgressTerminated = "terminated"
)

type NodeProgressSummary struct {
    NodeID      string `json:"nodeId"`
    NodeName    string `json:"nodeName"`
    NodeType    string `json:"nodeType"`
    GatewayMode string `json:"gatewayMode,omitempty"`
    Status      string `json:"status"`
}
```

Add this field to `InstanceDetail`:

```go
NodeProgress []NodeProgressSummary `json:"nodeProgress"`
```

- [ ] **Step 4: Keep the test failing for the missing behavior**

Run the same focused test.

Expected: compilation still fails only because `BuildNodeProgress` is missing.

### Task 2: Implement and test the pure progress projection

**Files:**
- Create: `backend/internal/modules/workflow/application/node_progress.go`
- Modify: `backend/internal/modules/workflow/application/node_progress_test.go`

- [ ] **Step 1: Implement the minimal linear projection**

Create this public boundary and keep all helpers private:

```go
func BuildNodeProgress(
    definition workflowcore.Definition,
    instance InstanceSummary,
    tokens []TokenSummary,
    tasks []TaskSummary,
    history []HistorySummary,
) []NodeProgressSummary
```

For each node, apply status precedence in this order:

```go
if processing[node.ID] {
    status = NodeProgressProcessing
} else if terminated[node.ID] {
    status = NodeProgressTerminated
} else if completed[node.ID] {
    status = NodeProgressCompleted
} else if instance.Status == "completed" {
    status = NodeProgressSkipped
} else if instance.Status == "running" {
    status = NodeProgressNotStarted
} else {
    status = NodeProgressTerminated
}
```

Treat the start node as completed, successful terminal task states (`approved`, `completed`, `submitted`) as completed, `pending`/`waiting` tasks and `active`/`waiting` tokens as processing, rejected tasks as terminated, and the end node as completed only when the instance completed.

- [ ] **Step 2: Run the linear test and verify GREEN**

Run the focused command from Task 1.

Expected: PASS.

- [ ] **Step 3: Add failing condition-branch and terminal tests**

Add tests with a split gateway and two branches:

```go
func TestBuildNodeProgressMarksUnselectedConditionBranchSkipped(t *testing.T) {
    definition := conditionalDefinition()
    actual := BuildNodeProgress(
        definition,
        InstanceSummary{Status: "completed", StartTime: 1000, EndTime: 2000},
        []TokenSummary{{ID: "token-1", NodeID: "end", Status: "completed"}},
        []TaskSummary{{ID: "task-finance", NodeID: "finance", Status: "approved"}},
        []HistorySummary{{ID: "history-finance", EventType: "task_approved", NodeID: "finance", EventTime: 1500}},
    )
    status := progressStatusByNode(actual)
    for _, nodeID := range []string{"start", "split", "finance", "join", "end"} {
        if status[nodeID] != NodeProgressCompleted {
            t.Fatalf("node %s status = %q, want completed", nodeID, status[nodeID])
        }
    }
    if status["hr"] != NodeProgressSkipped {
        t.Fatalf("hr status = %q, want skipped", status["hr"])
    }
}

func TestBuildNodeProgressMarksRemainingNodesTerminatedAfterReject(t *testing.T) {
    definition := workflowcore.Definition{
        Nodes: []workflowcore.Node{
            {ID: "start", Type: workflowcore.NodeTypeStart},
            {ID: "approve", Type: workflowcore.NodeTypeApproval},
            {ID: "archive", Type: workflowcore.NodeTypeHandle},
            {ID: "end", Type: workflowcore.NodeTypeEnd},
        },
        Edges: []workflowcore.Edge{
            {ID: "e1", Source: "start", Target: "approve"},
            {ID: "e2", Source: "approve", Target: "archive"},
            {ID: "e3", Source: "archive", Target: "end"},
        },
    }
    actual := BuildNodeProgress(
        definition,
        InstanceSummary{Status: "rejected", StartTime: 1000, EndTime: 2000},
        []TokenSummary{{ID: "token-1", NodeID: "approve", Status: "cancelled"}},
        []TaskSummary{{ID: "task-1", NodeID: "approve", Status: "rejected"}},
        []HistorySummary{{ID: "history-1", EventType: "task_rejected", NodeID: "approve", EventTime: 1800}},
    )
    status := progressStatusByNode(actual)
    if status["start"] != NodeProgressCompleted {
        t.Fatalf("start status = %q, want completed", status["start"])
    }
    for _, nodeID := range []string{"approve", "archive", "end"} {
        if status[nodeID] != NodeProgressTerminated {
            t.Fatalf("node %s status = %q, want terminated", nodeID, status[nodeID])
        }
    }
}

func conditionalDefinition() workflowcore.Definition {
    return workflowcore.Definition{
        Nodes: []workflowcore.Node{
            {ID: "start", Type: workflowcore.NodeTypeStart},
            {ID: "split", Type: workflowcore.NodeTypeExclusive, GatewayMode: workflowcore.GatewayModeSplit},
            {ID: "finance", Type: workflowcore.NodeTypeApproval},
            {ID: "hr", Type: workflowcore.NodeTypeApproval},
            {ID: "join", Type: workflowcore.NodeTypeExclusive, GatewayMode: workflowcore.GatewayModeJoin},
            {ID: "end", Type: workflowcore.NodeTypeEnd},
        },
        Edges: []workflowcore.Edge{
            {ID: "e1", Source: "start", Target: "split"},
            {ID: "e2", Source: "split", Target: "finance"},
            {ID: "e3", Source: "split", Target: "hr"},
            {ID: "e4", Source: "finance", Target: "join"},
            {ID: "e5", Source: "hr", Target: "join"},
            {ID: "e6", Source: "join", Target: "end"},
        },
    }
}

func progressStatusByNode(progress []NodeProgressSummary) map[string]string {
    result := make(map[string]string, len(progress))
    for _, item := range progress {
        result[item.NodeID] = item.Status
    }
    return result
}
```

Expected assertions use all five stable status constants and verify topological order.

- [ ] **Step 4: Run the new tests and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application -run TestBuildNodeProgress -count=1
```

Expected: the unselected branch or gateway assertions fail.

- [ ] **Step 5: Implement gateway dominance and stable topology ordering**

Build adjacency and indegree maps from valid edges. Return Kahn topological order; sort each ready group by position Y, then position X, then original definition index. If invalid historical data leaves nodes after Kahn traversal, append remaining nodes in original order instead of dropping them.

For each executed/current evidence node, mark a gateway completed when removing that gateway makes the evidence node unreachable from the start node. This identifies split/join gateways on the selected route without marking nodes on an unselected condition branch.

- [ ] **Step 6: Add and pass the parallel-flow test**

Add a table-driven test where both parallel branches have pending tasks, then both branches are complete and a post-join task is pending:

```go
func TestBuildNodeProgressTracksParallelBranchesAndJoin(t *testing.T) {
    definition := parallelDefinition()
    cases := []struct {
        name     string
        tokens   []TokenSummary
        tasks    []TaskSummary
        expected map[string]string
    }{
        {
            name: "branches active",
            tokens: []TokenSummary{
                {ID: "left-token", NodeID: "left", Status: "waiting"},
                {ID: "right-token", NodeID: "right", Status: "waiting"},
            },
            tasks: []TaskSummary{
                {ID: "left-task", NodeID: "left", Status: "pending"},
                {ID: "right-task", NodeID: "right", Status: "pending"},
            },
            expected: map[string]string{"split": NodeProgressCompleted, "left": NodeProgressProcessing, "right": NodeProgressProcessing, "join": NodeProgressNotStarted},
        },
        {
            name: "joined",
            tokens: []TokenSummary{{ID: "next-token", NodeID: "next", Status: "waiting"}},
            tasks: []TaskSummary{
                {ID: "left-task", NodeID: "left", Status: "approved"},
                {ID: "right-task", NodeID: "right", Status: "approved"},
                {ID: "next-task", NodeID: "next", Status: "pending"},
            },
            expected: map[string]string{"split": NodeProgressCompleted, "left": NodeProgressCompleted, "right": NodeProgressCompleted, "join": NodeProgressCompleted, "next": NodeProgressProcessing},
        },
    }
    for _, test := range cases {
        t.Run(test.name, func(t *testing.T) {
            status := progressStatusByNode(BuildNodeProgress(definition, InstanceSummary{Status: "running", StartTime: 1000}, test.tokens, test.tasks, nil))
            for nodeID, expected := range test.expected {
                if status[nodeID] != expected {
                    t.Fatalf("node %s status = %q, want %q", nodeID, status[nodeID], expected)
                }
            }
        })
    }
}
```

Use this parallel definition helper:

```go
func parallelDefinition() workflowcore.Definition {
    return workflowcore.Definition{
        Nodes: []workflowcore.Node{
            {ID: "start", Type: workflowcore.NodeTypeStart},
            {ID: "split", Type: workflowcore.NodeTypeParallel, GatewayMode: workflowcore.GatewayModeSplit},
            {ID: "left", Type: workflowcore.NodeTypeApproval},
            {ID: "right", Type: workflowcore.NodeTypeApproval},
            {ID: "join", Type: workflowcore.NodeTypeParallel, GatewayMode: workflowcore.GatewayModeJoin},
            {ID: "next", Type: workflowcore.NodeTypeHandle},
            {ID: "end", Type: workflowcore.NodeTypeEnd},
        },
        Edges: []workflowcore.Edge{
            {ID: "e1", Source: "start", Target: "split"},
            {ID: "e2", Source: "split", Target: "left"},
            {ID: "e3", Source: "split", Target: "right"},
            {ID: "e4", Source: "left", Target: "join"},
            {ID: "e5", Source: "right", Target: "join"},
            {ID: "e6", Source: "join", Target: "next"},
            {ID: "e7", Source: "next", Target: "end"},
        },
    }
}
```

Run the focused application test suite and expect PASS.

- [ ] **Step 7: Format and inspect the backend projection**

Run:

```bash
gofmt -w backend/internal/modules/workflow/application/types.go backend/internal/modules/workflow/application/node_progress.go backend/internal/modules/workflow/application/node_progress_test.go
git diff --check -- backend/internal/modules/workflow/application/types.go backend/internal/modules/workflow/application/node_progress.go backend/internal/modules/workflow/application/node_progress_test.go
```

Expected: no output from `git diff --check`.

### Task 3: Populate progress from the instance-bound definition version

**Files:**
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Modify: `docs/architecture/generic-oa-workflow-v1.md`

- [ ] **Step 1: Add a failing infrastructure mapping assertion**

Extract a small mapping helper named `populateInstanceNodeProgress` and test its intended behavior before implementing it:

```go
func TestPopulateInstanceNodeProgressUsesLoadedDefinition(t *testing.T) {
    definition := workflowcore.Definition{
        Nodes: []workflowcore.Node{
            {ID: "start-v2", Type: workflowcore.NodeTypeStart},
            {ID: "approve-v2", Type: workflowcore.NodeTypeApproval},
            {ID: "end-v2", Type: workflowcore.NodeTypeEnd},
        },
        Edges: []workflowcore.Edge{
            {ID: "e1", Source: "start-v2", Target: "approve-v2"},
            {ID: "e2", Source: "approve-v2", Target: "end-v2"},
        },
    }
    detail := &application.InstanceDetail{
        Instance: application.InstanceSummary{DefinitionVersion: 2, Status: "running", StartTime: 1000},
        Tasks: []application.TaskSummary{{ID: "task-1", NodeID: "approve-v2", Status: "pending"}},
    }

    populateInstanceNodeProgress(detail, definition)

    if len(detail.NodeProgress) != 3 || detail.NodeProgress[0].NodeID != "start-v2" || detail.NodeProgress[2].Status != application.NodeProgressNotStarted {
        t.Fatalf("node progress = %#v", detail.NodeProgress)
    }
}
```

- [ ] **Step 2: Run the infrastructure test and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure -run NodeProgress -count=1
```

Expected: FAIL because `loadInstanceDetail` does not populate `NodeProgress`.

- [ ] **Step 3: Populate the new field after all runtime rows are mapped**

In `loadInstanceDetail`, keep the existing call to:

```go
definition, _, err := loadDefinitionVersion(db, instance.DefinitionID, instance.DefinitionVersion)
```

Implement the helper and call it after mapping tokens, tasks and history:

```go
func populateInstanceNodeProgress(detail *application.InstanceDetail, definition workflowcore.Definition) {
    detail.NodeProgress = application.BuildNodeProgress(
        definition,
        detail.Instance,
        detail.Tokens,
        detail.Tasks,
        detail.History,
    )
}
```

Do not load the current definition or mutate runtime tables.

- [ ] **Step 4: Document the additive detail contract**

Add to `docs/architecture/generic-oa-workflow-v1.md` that instance detail exposes `nodeProgress`, computed from the instance-bound published version, and list the five statuses.

- [ ] **Step 5: Run backend focused tests**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -count=1
```

Expected: PASS.

### Task 4: Lock the H5 contract and component structure with RED checks

**Files:**
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: Add failing structure requirements**

Require `WorkflowDetailPanel.vue` to import and render `WorkflowNodeProgressList` in both history presentations. Add a new file rule for `WorkflowNodeProgressList.vue` requiring:

```js
patterns: [
  'WorkflowNodeProgressSummary',
  'WorkflowTaskSummary',
  'nodeProgress',
  'legacyProgress',
  'nodeProgressStatusMeta',
  'nodeTypeLabel',
  'workflow-node-progress__item--processing',
  'workflow-node-progress__item--completed',
  'workflow-node-progress__item--not-started',
  'workflow-node-progress__item--skipped',
  'workflow-node-progress__item--terminated',
]
```

- [ ] **Step 2: Run the workflow structure check and verify RED**

Run:

```bash
cd h5app
node scripts/check-workflow-module.mjs
```

Expected: FAIL because the component and types do not exist.

- [ ] **Step 3: Add the TypeScript contract**

Add:

```ts
export type WorkflowNodeProgressStatus
  = | 'completed'
    | 'processing'
    | 'not_started'
    | 'skipped'
    | 'terminated'

export interface WorkflowNodeProgressSummary {
  nodeId: string
  nodeName: string
  nodeType: string
  gatewayMode?: string
  status: WorkflowNodeProgressStatus
}
```

Extend `WorkflowInstanceDetail` with `nodeProgress?: WorkflowNodeProgressSummary[]` so H5 can roll out before or after the backend.

### Task 5: Build and integrate the reusable H5 node progress list

**Files:**
- Create: `h5app/src/pages/workflow/components/WorkflowNodeProgressList.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`

- [ ] **Step 1: Implement the component input and legacy fallback**

Use explicit props:

```ts
const props = defineProps<{
  nodeProgress: WorkflowNodeProgressSummary[]
  tasks: WorkflowTaskSummary[]
  instance: WorkflowInstanceSummary
}>()
```

When `nodeProgress` is empty, group tasks by `nodeId` and synthesize one item per task node. Map pending/waiting to processing, rejected/cancelled to terminated, and all other terminal tasks to completed. This preserves the old visible data during gray deployment but does not invent missing nodes.

- [ ] **Step 2: Group tasks and render all node categories**

Create a computed `tasksByNode` map. Render each progress item with:

- node name and node type label;
- status tag from `nodeProgressStatusMeta`;
- start initiator/time or completed end time;
- every task under that node with assignee, handled time, action and comment;
- no fabricated actor line for gateway, CC, notification, automation or timer nodes.

Use uView Pro `u-tag` and existing visual tokens. Keep a stable rail/dot column and content width so status and long node names cannot shift the drawer.

- [ ] **Step 3: Integrate in both record surfaces**

Import the component in `WorkflowDetailPanel.vue`. Replace the drawer's `v-for="task in sortedTasks"` block with:

```vue
<WorkflowNodeProgressList
  :node-progress="detail.nodeProgress || []"
  :tasks="sortedTasks"
  :instance="detail.instance"
/>
```

Render the same component at the top of the non-drawer `activeSection === 'history'` view, before raw history events. Keep raw events below it.

- [ ] **Step 4: Run the H5 RED check again and finish GREEN**

Run:

```bash
cd h5app
node scripts/check-workflow-module.mjs
```

Expected: PASS after all required component patterns and integrations exist.

- [ ] **Step 5: Run focused lint and type-check**

Run:

```bash
cd h5app
node_modules/.bin/eslint src/types/workflow.ts src/pages/workflow/components/WorkflowDetailPanel.vue src/pages/workflow/components/WorkflowNodeProgressList.vue scripts/check-workflow-module.mjs
node_modules/.bin/vue-tsc --noEmit
```

Expected: both commands exit 0.

### Task 6: Full verification and rendered QA

**Files:**
- Verify only; no planned source changes.

- [ ] **Step 1: Run the backend full suite**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
```

Expected: PASS. If the environment blocks loopback listeners, rerun outside the sandbox and record the reason.

- [ ] **Step 2: Run H5 checks and build**

```bash
cd h5app
node scripts/check-workflow-module.mjs
node_modules/.bin/eslint src/types/workflow.ts src/pages/workflow/components/WorkflowDetailPanel.vue src/pages/workflow/components/WorkflowNodeProgressList.vue scripts/check-workflow-module.mjs
node_modules/.bin/vue-tsc --noEmit
node_modules/.bin/uni build
```

Expected: structure check, focused lint, type-check and build all exit 0. Also run `pnpm lint`; if existing unrelated files still block the full lint, report the exact baseline count without modifying them.

- [ ] **Step 3: Inspect the final diff boundary**

```bash
git diff --check
git status --short
git -C h5app status --short
```

Expected: no whitespace errors; only the files listed in this plan are attributed to this task. Preserve all unrelated dirty files and staged changes.

- [ ] **Step 4: Validate the rendered workflow history**

Start H5App on an available local port and test:

1. 发起审批 -> 历史记录 -> 查看实例。
2. Confirm all historical-version nodes appear in topology order.
3. Confirm completed, processing, not-started, skipped and terminated styles are distinct.
4. Confirm multi-person approval tasks remain grouped under one node.
5. Confirm the desktop drawer remains 25% wide and the mobile breakpoint remains 94% wide.
6. Confirm page identity, nonblank content, no framework overlay, console health and one target interaction with Browser.

If authentication blocks the target interaction, do not forge login state; record the blocker and leave final manual flow verification to the user.
