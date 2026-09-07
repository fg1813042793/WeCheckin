# 已完成流程表单修订设计

## 1. 目标与边界

在独立通用工作流模块中增加“已完成流程的节点处理人修订”能力，使中间审批人或办理人在流程结束后仍可更正自己负责的表单字段，同时保证后续审批结论与最终表单数据一致。

本期包含：

- 实际处理过节点的用户对已完成实例发起表单修订。
- 按节点字段权限限制可修订字段。
- 低风险字段直接修订，高风险字段由原实际路径中的后续节点重新确认。
- 修订前后快照、字段差异、原因、任务和通知的完整审计。
- H5App 发起、处理和查看修订；Admin 配置及查看修订记录。
- 修订任务计入通用工作流待办、已处理和总览统计。

本期不包含：

- 修改现有绩效业务流程或直接写入绩效业务表。
- 发起人通过本能力修改自己提交的表单；发起人继续使用撤回或重新发起。
- 将已完成原实例重新置为运行中，或恢复、覆盖原流程任务。
- 允许修改分支条件字段、计算字段、流程发起人、业务编号和流程定义版本。
- 对历史流程定义自动开放完成后修订。

## 2. 当前能力与新增语义

现有 `PATCH /api/v2/dingtalk/h5/workflows/instances/{id}/form-data` 保持不变：

- 只处理运行中的流程实例。
- 用户必须实际办理过已开启“办理完成后允许修改表单”的节点。
- 可修改字段是该用户已办理节点中配置为 `write` 的字段。
- 修改使用 `expectedRevision` 乐观锁，记录原因并可通知流程参与人。

新增能力只处理状态为 `completed` 的实例，名称统一为“完成后修订”。它不会复用现有直接修改接口，也不会放宽现有运行中修改的权限判断。

### 2.1 发起人

原流程发起人不因发起身份获得完成后修订权限。发起人需要修改时使用以下既有业务方式：

- 流程运行中：撤回后修改并重新发起。
- 流程已完成：基于原表单重新发起一条新流程。

发起人若同时是某个节点的实际处理人，只能以该节点处理人身份按本设计申请修订。

### 2.2 节点处理人

完成后修订必须绑定一个“来源节点”：

- 当前用户必须在原实例历史中实际审批、办理或退回过该节点。
- 节点类型只允许 `approval` 或 `handle`。
- 节点必须在发布版本中显式开启完成后修订。
- 用户处理过多个符合条件的节点时，H5App 默认选择最后处理的节点，并允许用户切换。
- 可修订字段只取所选来源节点的权限，不合并多个节点权限，避免下游重新确认范围不明确。

## 3. 流程定义配置

扩展现有 `postHandleEdit` 配置，旧字段保持兼容：

```json
{
  "postHandleEdit": {
    "enabled": true,
    "completedRevision": {
      "enabled": true,
      "directFields": ["contactPhone", "remark"]
    }
  }
}
```

语义如下：

- `postHandleEdit.enabled`：保持现有含义，控制流程运行期间办理后修改。
- `completedRevision.enabled`：控制流程完成后，实际节点处理人能否发起修订。
- `completedRevision.directFields`：允许不经过后续节点确认而直接生效的低风险字段。
- 来源节点中其他 `write` 字段默认属于“后续节点重新确认”字段。
- `directFields` 必须是该节点有效的 `write` 字段子集。

定义校验必须拒绝：

- 将分支条件字段放入完成后可修订范围或 `directFields`。
- 将计算字段配置为可修订或直接修订。
- `directFields` 引用了不存在、隐藏、只读或非数据组件字段。
- 非审批、非办理节点配置完成后修订。

历史定义没有 `completedRevision` 时视为禁用，不改变已发布定义和历史实例行为。

## 4. 修订类型与处理规则

### 4.1 直接修订

当一次修订只包含 `directFields` 时：

1. 锁定原实例并校验状态、来源节点、处理人身份和表单版本。
2. 校验补丁、明细行操作权限并重新执行表单计算。
3. 在同一事务内写入修订请求、前后快照、原实例新表单数据、版本号和历史事件。
4. 修订请求直接进入 `applied` 状态。
5. 事务提交后通知原发起人和原流程下游参与人。

直接修订仍必须填写原因，不能绕过审计。

### 4.2 需确认修订

一次修订包含任一非 `directFields` 的可写字段时，整次修订进入后续确认：

1. 原实例保持 `completed`，当前表单不立即变化。
2. 保存修订请求、基准版本、修改前快照、补丁和计算后的候选表单。
3. 按原实例实际执行路径生成来源节点之后的确认任务。
4. 所有确认任务通过后，事务内把候选表单应用到原实例并递增版本。
5. 任一任务驳回后修订请求进入 `rejected`，原表单不变化。
6. 修订发起人可在首个任务处理前取消请求；取消后原表单不变化。

## 5. 下游确认路径

修订不重新计算原流程的历史分支。由于分支条件字段禁止修改，原实例实际走过的路径仍然有效。

确认路径按以下规则生成：

1. 从原实例发布版本读取节点和边。
2. 从原实例任务及流转历史提取实际执行过的审批、办理节点。
3. 仅保留在来源节点之后、实际执行过且未被跳过或终止的人工节点。
4. 网关和自动节点不生成任务，只用于恢复节点之间的先后和并行关系。
5. 串行节点按拓扑顺序逐级创建任务。
6. 原流程中的并行分支按原拓扑阶段同时创建任务，全部完成后再进入汇合后的节点。
7. 会签、或签和依次审批沿用原发布节点的审批模式。
8. 任务处理人按原发布节点人员规则在修订请求创建时重新解析，使用当前有效的组织关系。
9. 修订发起人不会收到来源节点的自审批任务；确认从来源节点的下游人工节点开始。

如果来源节点之后没有任何可确认的人工节点：

- 仅修改 `directFields` 时允许直接生效。
- 包含需确认字段时拒绝发起，并提示流程设计者将字段加入低风险直接修订范围，或由发起人重新发起完整流程。

人员解析失败、后续节点没有有效处理人或路径无法稳定重建时，修订请求不得创建，不能降级为直接生效。

## 6. 领域与应用分层

新增完成后修订聚合，归属 `backend/internal/modules/workflow`：

- `FormRevisionRequest`：修订请求、候选数据、状态和版本。
- `FormRevisionTask`：后续节点确认任务。
- `FormRevisionPlan`：从原发布图和实际执行历史生成的不可变确认计划。

`backend/internal/workflowcore` 只负责纯规则：

- 解析来源节点可见、可写和直接修订字段。
- 排除条件字段和计算字段。
- 校验修订补丁及明细行增删权限。
- 从发布图及实际节点集合构建确认阶段。
- 合并补丁并执行表单计算。

应用层负责事务、身份、并发、持久化、任务推进和通知。领域层不得依赖 H5App、Admin、钉钉或绩效模块。

## 7. 数据模型

使用版本化 SQL 新增 `workflow_form_revision_requests`：

| 字段 | 说明 |
| --- | --- |
| `id` | 修订请求 ID |
| `source_instance_id` | 已完成原流程实例 ID |
| `source_node_id`、`source_node_name` | 修订人对应的原办理节点快照 |
| `requester_id` | 修订发起人本地用户 ID |
| `base_form_revision` | 创建请求时的原表单版本 |
| `before_form_data_json` | 修改前完整表单快照 |
| `patch_json` | 用户实际修改字段 |
| `proposed_form_data_json` | 校验和计算后的候选完整表单 |
| `changed_field_labels_json` | 修改字段名称快照 |
| `reason` | 修订原因，最多 500 字符 |
| `revision_mode` | `direct` 或 `downstream_review` |
| `status` | `pending`、`applied`、`rejected`、`cancelled`、`conflict` |
| `active_lock_key` | 进行中请求使用原实例 ID，终态为空 |
| `applied_form_revision` | 生效后的原表单版本 |
| `add_time`、`edit_time`、`completed_at` | 时间字段 |

`active_lock_key` 建唯一索引；MySQL 允许多个 `NULL`，从而保证同一原实例最多一条进行中的修订请求。

新增 `workflow_form_revision_tasks`：

| 字段 | 说明 |
| --- | --- |
| `id` | 修订任务 ID |
| `revision_request_id` | 修订请求 ID |
| `source_task_id` | 可空的原流程任务来源 |
| `node_id`、`node_name` | 节点快照 |
| `stage` | 串行或并行确认阶段 |
| `assignee_id`、`assignee_name` | 处理人快照 |
| `approval_mode` | `all`、`any` 或 `sequential` |
| `status` | `waiting`、`pending`、`approved`、`rejected`、`cancelled` |
| `comment`、`images_json` | 处理意见和图片 |
| `handled_by`、`handled_at` | 实际处理信息 |
| `add_time`、`edit_time` | 时间字段 |

两张表都使用软状态终止，不物理覆盖历史记录。原实例继续使用现有 `form_revision` 作为当前表单版本。

## 8. 事务与并发

创建修订请求时必须在一个事务中：

- `FOR UPDATE` 锁定原实例。
- 校验原实例仍为 `completed`。
- 校验 `expectedRevision` 等于原实例当前 `form_revision`。
- 校验没有进行中的修订请求。
- 写入修订请求、确认计划、首阶段任务、历史和通知 Outbox。

修订最终生效时再次锁定原实例并比较 `base_form_revision`：

- 版本一致：写入候选表单，递增原实例版本，标记请求 `applied`。
- 版本不一致：请求标记为 `conflict`，不覆盖原表单；用户刷新后重新提交。

任务处理、请求状态推进、原表单应用、历史和 Outbox 必须在同一事务中完成。通知发送失败不得回滚已经提交的业务状态。

## 9. 历史与审计

原实例流转记录新增事件：

- `instance_form_revision_requested`：谁从哪个节点发起、修改哪些字段。
- `instance_form_revision_applied`：修订请求生效及新表单版本。
- `instance_form_revision_rejected`：修订被谁在哪个节点驳回。
- `instance_form_revision_cancelled`：修订发起人取消。
- `instance_form_revision_conflict`：基准版本冲突，未覆盖原数据。

原有运行中直接修改继续使用 `instance_form_revised`，不追溯迁移。详情页按 `form_revision` 展示当前版本，并允许查看每次完成后修订的修改前、修改后、差异、原因和确认记录。

## 10. HTTP 与权限

新增钉钉 H5 接口：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/api/v2/dingtalk/h5/workflows/instances/{id}/form-revisions` | 查询原实例修订记录 |
| `POST` | `/api/v2/dingtalk/h5/workflows/instances/{id}/form-revisions` | 发起完成后修订 |
| `GET` | `/api/v2/dingtalk/h5/workflows/form-revisions/{id}` | 查询修订详情与差异 |
| `POST` | `/api/v2/dingtalk/h5/workflows/form-revisions/{id}/cancel` | 取消未处理修订 |
| `POST` | `/api/v2/dingtalk/h5/workflows/form-revision-tasks/{id}/complete` | 通过或驳回修订任务 |

实例详情的 `formRevision` 能力扩展为：

```json
{
  "revision": 2,
  "runningRevisionAllowed": false,
  "completedRevisionNodes": [
    {
      "nodeId": "manager_review",
      "nodeName": "上级评价",
      "fieldPermissions": [],
      "directFields": []
    }
  ]
}
```

新增权限：

- `dingtalk_h5:button:workflow:form-revision-create`
- `dingtalk_h5:api:workflow:form-revision-create`
- `dingtalk_h5:button:workflow:form-revision-handle`
- `dingtalk_h5:api:workflow:form-revision-handle`

新权限通过迁移注册，但不自动授权历史角色，避免升级后无意开放已完成数据修改。后端始终重新校验实例访问权、实际办理历史、节点配置和字段权限，前端按钮不是安全边界。

Admin 增加只读查询接口，用于实例详情查看修订请求、任务、差异和通知投递；本期不提供管理员直接覆盖表单数据的接口。

## 11. H5App 与 Admin 交互

### 11.1 H5App

- 已完成流程详情在存在可修订来源节点且用户有权限时显示“申请修订”。
- 多个来源节点时先选择“以哪个办理节点发起修订”。
- 修订页面仅展示该来源节点可见字段，仅允许编辑有效 `write` 字段。
- 页面实时展示修改字段摘要；提交时必须填写原因。
- 直接修订明确提示“提交后立即生效”。
- 需确认修订展示将重新确认的节点和处理人，不把普通确认伪装成即时保存。
- 下游处理人的“我的待办”显示“表单修订”类型，详情默认展示字段前后对比、原因和原流程入口。
- 修订发起人从原流程详情查看状态，不新增独立顶级菜单。

### 11.2 Admin

- 节点配置“办理后修改”区域增加“流程完成后允许修订”。
- 开启后展示该节点可写字段，并可从中选择“允许直接修订”的低风险字段。
- 条件字段和计算字段显示禁用原因，不能被选中。
- 流程校验、版本变更摘要和 BPMN 扩展属性包含完成后修订配置。
- 实例详情增加“表单修订”区域，展示请求状态、修订人、来源节点、字段差异、原因、任务及通知记录。

## 12. 通知

新增通知类型：

- `instance_form_revision_requested`：首阶段下游确认人收到修订待办。
- `instance_form_revision_approved`：修订最终生效，通知修订人、原发起人和已确认参与人。
- `instance_form_revision_rejected`：通知修订人和原发起人。
- `instance_form_revision_cancelled`：通知已经收到待办但尚未处理的人员。
- `instance_form_revision_conflict`：通知修订人重新提交。

通知继续使用通用 Outbox，支持站内信和钉钉 OA。站内信点击后打开修订详情，钉钉跳转使用企业应用地址并携带修订详情路由参数。

## 13. 业务同步

完成后修订首先只更新通用流程实例的 `form_data_json`。修订生效后产生通用 Outbox 事件：

```text
workflow.instance.form_revised
```

事件包含实例、业务类型、业务标识、表单版本、字段补丁和修订请求 ID。绑定 `BusinessType + BusinessKey` 的业务模块可独立订阅并决定是否同步自身数据；工作流模块不能直接依赖或修改绩效业务表。

业务订阅失败时保留事件并重试，不回滚已经通过的修订审批。业务页面在实现订阅前仍以流程表单详情为准，不能假定业务表已同步。

## 14. 错误处理

- 原实例不是已完成状态：拒绝完成后修订，提示使用当前可用流程操作。
- 用户不是来源节点实际处理人：返回无权限，不泄露字段或参与人信息。
- 字段无权限、属于条件或计算字段：整次请求失败，不部分保存。
- 后续路径为空但包含需确认字段：拒绝创建，不自动降级直接生效。
- 人员解析失败：拒绝创建，并返回具体失败节点名称。
- 同一实例已有进行中请求：返回现有修订请求 ID，前端引导查看。
- 乐观锁冲突：返回当前版本，用户刷新差异后重新提交。
- 重复任务提交：幂等返回当前任务及请求状态，不重复推进或通知。

## 15. 兼容与迁移

- 数据库结构只通过 `backend/migrations/` 的版本化 SQL 增加。
- 旧流程定义缺少 `completedRevision` 时保持禁用。
- 旧实例没有修订请求时详情返回空数组，不补写历史。
- 现有运行中表单修改接口、权限和通知语义保持不变。
- 已发布定义继续绑定原版本；修订能力读取原实例绑定的发布版本，不读取当前草稿或最新版本。
- Swagger、`docs/API_V2.md`、权限目录、迁移和 H5App 类型必须同步更新。

## 16. 测试与验收

后端至少覆盖：

- 只有实际处理人可以选择来源节点。
- 发起人仅凭发起身份不能修订。
- 旧定义和未开启节点不可修订。
- 单来源节点权限，不跨节点合并字段。
- 条件字段、计算字段和无权限字段被拒绝。
- 仅低风险字段直接生效并递增版本。
- 含需确认字段时生成下游串行、并行、会签和或签任务。
- 下游全部通过后原表单原子更新；驳回、取消时保持不变。
- 同一实例单进行中请求、重复提交幂等和版本冲突。
- 历史、任务、请求、原表单和通知 Outbox 的事务一致性。
- 业务修订事件包含 `BusinessType + BusinessKey`，且不依赖绩效模块。

Admin 至少覆盖：

- 完成后修订配置、直接字段选择和禁用原因。
- 定义校验、保存、发布版本、回显和变更摘要。
- 实例修订详情及字段差异只读展示。

H5App 至少覆盖：

- 已完成实例的按钮权限和来源节点选择。
- 修订页面字段权限、修改原因、直接生效提示和确认路径预览。
- 修订任务进入待办、处理后进入已处理，总览数字同步刷新。
- 修订通过、驳回、取消、冲突和人员解析失败状态。
- 站内信深链打开正确修订详情。
- PC、紧凑 PC 和手机布局无溢出、弹层层级正确。

验证命令按受影响范围执行：

```bash
cd backend && GOCACHE=$PWD/../.cache/go-build go test ./...
cd admin && npm run check:all
cd h5app && pnpm lint && pnpm type-check && pnpm check:workflow-module && pnpm check:ui-style && pnpm build:h5
bash scripts/verify-local.sh
```

浏览器验证覆盖流程设计、发起修订、下游确认、修订详情和通知深链；用户负责最终手工页面验收。
