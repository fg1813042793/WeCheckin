# 简化流程设计器实施计划

## 1. 目标

在现有 WeCheckin 管理后台增加一套参考 FlyFlow 交互思路、但由本项目自行维护的简化流程设计器。

第一阶段解决以下问题：

- 用可视化方式配置审批流程，不要求业务人员直接编辑 BPMN。
- 支持开始、审批、条件分支、并行分支、结束五类核心节点。
- 支持单人审批、依次审批、并行审批、会签四种审批方式。
- 保存设计草稿，发布时生成不可变版本和 BPMN XML。
- 在进入运行阶段前完成结构校验，避免无审批人、断路或无结束节点的流程发布。
- 接入现有后台菜单权限、按钮权限、接口权限和操作审计体系。

本阶段不直接替换钉钉 H5 绩效模块当前的固定状态流转。流程设计器完成并稳定后，再单独迁移绩效流程。

## 2. 系统边界

### 2.1 WeCheckin 负责

- 流程定义、设计草稿、版本和发布状态。
- 用户、部门、岗位、角色及统一权限。
- 业务表单、评价内容、分档结果和钉钉通知。
- 将设计态 JSON 编译为标准 BPMN XML。
- 通过适配器调用 Flowable REST API。

### 2.2 Flowable 负责

- 流程实例、任务、网关和会签运行。
- 任务认领、完成、退回及运行时变量。
- 流程历史和运行状态。

设计器不实现第二套流程运行引擎，避免设计规则和执行规则长期分叉。

## 3. 设计态 DSL

流程草稿保存为 JSON，顶层结构如下：

```json
{
  "schemaVersion": 1,
  "name": "月度绩效审批",
  "nodes": [
    { "id": "start", "type": "start", "name": "发起" },
    {
      "id": "manager_review",
      "type": "approval",
      "name": "上级审批",
      "approvalMode": "single",
      "assignee": { "type": "manager", "value": "direct" }
    },
    { "id": "end", "type": "end", "name": "完成" }
  ],
  "edges": [
    { "id": "e1", "source": "start", "target": "manager_review" },
    { "id": "e2", "source": "manager_review", "target": "end" }
  ]
}
```

审批人类型第一阶段支持：

- `user`：指定用户。
- `role`：指定角色。
- `department_leader`：发起人部门负责人。
- `manager`：发起人直属上级。
- `variable`：由业务在启动流程时传入。

条件边使用结构化表达式，不允许业务人员直接录入任意脚本：

```json
{
  "field": "score",
  "operator": "gte",
  "value": 80
}
```

后端只将受控运算符编译为 Flowable 表达式，避免把任意代码带入流程引擎。

## 4. 发布与版本规则

- 流程定义拥有稳定的 `definition_key`。
- 编辑操作只更新当前草稿。
- 每次发布创建新的不可变版本，版本号按定义递增。
- 已发布版本不允许原地修改，只允许复制为新草稿后再次发布。
- 发布时执行 DSL 校验和 BPMN 编译；校验失败不写入发布版本。
- Flowable 部署标识独立保存，便于部署失败后重试，不回滚设计版本。

## 5. 发布校验

发布前必须满足：

1. 有且仅有一个开始节点。
2. 至少有一个结束节点。
3. 除结束节点外，每个节点至少有一条出边。
4. 除开始节点外，每个节点至少有一条入边。
5. 所有节点都能从开始节点到达。
6. 所有可执行路径最终都能到达结束节点。
7. 审批节点必须配置审批方式和审批人。
8. 条件分支至少有两条分支，且每条分支条件完整；最多允许一条默认分支。
9. 并行分支必须成对产生和汇聚，第一阶段不支持跨层连接。
10. 节点和连线 ID 在一个定义内唯一。

## 6. 后台界面

### 6.1 流程定义列表

- 搜索流程名称或编码。
- 展示草稿状态、当前发布版本、更新时间和更新人。
- 提供创建、编辑、复制、版本、发布、停用和删除操作。
- 删除仅允许无已发布版本且无运行实例的草稿。

### 6.2 简化设计器

采用三栏工作区：

- 左侧节点工具：审批、条件分支、并行分支。
- 中间流程画布：纵向流程主干和分支结构，节点之间可插入新节点。
- 右侧属性面板：编辑流程属性、节点名称、审批方式、审批人和条件。

顶部工具栏提供返回、保存草稿、校验和发布。画布风格参考 FlyFlow 的低门槛审批编排体验，但组件、数据结构和交互代码由本项目独立实现。

## 7. 后端模块

```text
backend/internal/model/workflow/
backend/internal/service/admin/workflow/
backend/internal/handler/admin/workflow/
```

核心表：

- `workflow_definitions`：稳定定义、草稿、当前版本和状态。
- `workflow_definition_versions`：不可变版本、DSL 快照、BPMN XML、校验结果和部署标识。

管理接口：

```text
GET    /api/v2/admin/workflow-definitions
POST   /api/v2/admin/workflow-definitions
GET    /api/v2/admin/workflow-definitions/:id
PUT    /api/v2/admin/workflow-definitions/:id
DELETE /api/v2/admin/workflow-definitions/:id
POST   /api/v2/admin/workflow-definitions/:id/validate
POST   /api/v2/admin/workflow-definitions/:id/publish
GET    /api/v2/admin/workflow-definitions/:id/versions
```

## 8. 权限编码

菜单和按钮：

```text
admin:menu:workflow
admin:menu:workflow:definitions
admin:menu:workflow:add
admin:menu:workflow:edit
admin:menu:workflow:publish
admin:menu:workflow:del
```

接口权限：

```text
admin:api:workflow:list
admin:api:workflow:add
admin:api:workflow:edit
admin:api:workflow:publish
admin:api:workflow:del
```

## 9. 分阶段交付

### P0

- DSL 类型、校验器、BPMN 编译器及测试。
- 定义与版本模型。
- 列表、详情、保存、校验、发布接口。
- 后台列表和基础设计器。
- 菜单、按钮、接口权限接入。

### P1

- Flowable REST 部署适配器和失败重试。
- 审批人预览与流程模拟。
- 版本差异对比和复制版本。

### P2

- 将绩效考评流程迁移为流程定义驱动。
- 增加业务节点表单权限、抄送、超时和通知策略。
- 运行实例和任务管理页面。

## 10. 验收标准

- 管理员可创建流程、插入节点、编辑属性并保存草稿。
- 无效流程无法发布，页面能定位并说明错误节点。
- 有效流程发布后生成递增版本和可解析 BPMN XML。
- 已发布版本不可被更新覆盖。
- 不具备菜单、按钮或接口权限时，前端不展示操作且后端拒绝请求。
- 后端核心校验与编译测试通过，管理后台类型检查和生产构建通过。
