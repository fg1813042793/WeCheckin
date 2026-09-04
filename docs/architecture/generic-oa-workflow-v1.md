# 通用 OA 流程能力 V1

## 1. 目标与边界

在现有纯 Go 工作流引擎上补齐可独立复用的 OA 运行闭环。流程定义、表单、实例、任务和审计均属于通用 workflow 模块，业务模块只通过 `businessType + businessKey` 建立关联。

本阶段不接入绩效流程，不修改钉钉 H5 绩效模型、接口或页面。

## 2. 能力范围

### 2.1 流程定义

- 流程定义可配置独立 Logo，用于管理端流程列表和设计器标题区展示；Logo 属于定义元数据，不写入设计草稿、BPMN 或历史发布版本。
- Logo 仅接受 PNG、JPG/JPEG 和 WebP，单文件不超过 2MB；创建和修改接口兼容原 JSON 请求，并支持通过 `multipart/form-data` 随定义元数据一起提交。
- 发布版本包含表单 Schema，支持文本、多行文本、数字、单选、多选、日期、日期时间、用户、部门、附件和布尔字段。
- 表单字段支持必填、默认值、长度及数值范围约束。
- 表单字段可配置结构化高级校验规则：最小/最大长度、格式匹配、数值范围、小数位、选择数量、字段比较、条件必填和明细数字/金额列合计；规则支持自定义错误提示，列合计中的空单元格按 `0` 计算，明细列必填及类型错误优先返回。
- 发布时可将允许发起范围配置为全部用户，或在一个组织树中配置指定用户与多个部门的并集；部门只匹配所选本级，不自动包含子部门，用户部门归属在列表、详情和正式发起时从 `user_depts` 实时读取。两种范围都支持 `excludedUserIds`，排除用户优先于明确允许用户和允许部门；历史定义缺少该字段时按空列表处理。
- 设计器“流程配置”页签集中维护发布配置；允许发起时间支持长期有效、一次性时间段、每周周期和每月周期，统一按 `Asia/Shanghai` 计算并在正式创建实例时再次校验。
- 每月规则可选择具体日期或“最后一天”；具体 `31` 日在短月跳过，运行中的实例不受发起窗口关闭影响。
- 审批节点按字段配置 `hidden`、`read`、`write` 权限；未显式配置时默认只读。
- 发布前校验字段编码唯一、选项完整、校验规则参数及字段引用有效、节点权限引用有效。

### 2.2 流程运行

- 启动实例时校验并保存表单数据，表单数据与流程变量分离。
- 管理端在发起和任务提交前显示字段级校验提示，后端在发起及审批表单合并后执行相同规则并作为最终校验边界。
- 任务处理时只允许修改当前节点具有写权限的字段。
- 保留现有单人、顺序、并行、会签、排他网关和并行网关能力。
- 发起人在尚无审批任务处理时可以撤回；管理员可以取消运行中的实例。
- 所有启动、审批、拒绝、撤回、取消和完成动作写入历史审计。
- 审批节点可独立配置“任务处理结果通知”的启用状态、发送渠道、标题和正文；接收人固定为流程发起人，模板可使用 `{{result}}`，审批通过和驳回时分别渲染为“已通过”和“已驳回”。
- 单人审批在任务通过时发送结果；顺序、并行和会签在节点整体满足通过条件后发送一次结果；任一有效任务驳回时立即发送驳回结果。历史发布定义缺少 `resultNotification` 时保持不发送，不自动改变原流程行为。
- 实例详情附加返回 `nodeProgress`，按实例绑定的已发布定义版本投影全部节点，避免流程定义后续修改影响历史记录；节点状态固定为 `completed`（已完成）、`processing`（处理中）、`not_started`（未开始）、`skipped`（已跳过）和 `terminated`（已终止）。该字段只读计算，不修改历史实例或运行时表。

### 2.3 用户接口

- 查询已发布且可发起的流程定义及表单 Schema。
- 按用户和流程定义保存、恢复一份发起草稿；草稿不创建流程实例、待办、通知或历史记录。
- 发起流程。
- 查询我的申请、我的待办和我参与过的流程详情。
- 审批或拒绝自己的待办。
- 撤回自己尚未被处理的申请。

用户身份统一取客户端认证上下文，接口不接受可伪造的发起人或处理人参数。

### 2.4 管理接口

- 继续使用现有定义、实例和任务管理接口。
- 增加运行实例强制取消能力。
- 管理员可以代员工发起流程，但必须明确选择业务发起人。实例以 `starterId` 解析直属上级和组织审批身份，以登录管理员的 `operatorId` 记录实际操作审计。

### 2.5 HTTP 接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/api/v2/workflows/definitions` | 查询已发布流程 |
| `GET` | `/api/v2/workflows/definitions/:id` | 查询流程和表单 Schema |
| `GET` | `/api/v2/workflows/drafts/:definitionId` | 查询当前用户的流程发起草稿 |
| `PUT` | `/api/v2/workflows/drafts/:definitionId` | 保存当前用户的流程发起草稿 |
| `POST` | `/api/v2/workflows/instances` | 发起流程 |
| `GET` | `/api/v2/workflows/instances` | 查询我的申请 |
| `GET` | `/api/v2/workflows/instances/:id` | 查询我发起或参与的流程详情 |
| `POST` | `/api/v2/workflows/instances/:id/withdraw` | 撤回未处理的申请 |
| `GET` | `/api/v2/workflows/tasks` | 查询我的任务 |
| `POST` | `/api/v2/workflows/tasks/:id/complete` | 审批或拒绝我的任务 |
| `POST` | `/api/v2/admin/workflow-instances/:id/cancel` | 管理员取消运行实例 |

普通用户分别需要 `workflow:view`、`workflow:start`、`workflow:handle` 和 `workflow:withdraw` 权限；迁移仅注册权限，不自动授权给任何普通角色。

## 3. 数据模型

- `workflow_definitions.definition_logo_url` 保存 Logo 的内部资源路径，接口输出时按系统静态资源域名转换为可访问地址。
- `workflow_process_instances.form_data_json` 保存当前表单数据快照。
- `workflow_start_drafts` 按 `definition_id + starter_id` 保存一份未发起的表单草稿；正式发起成功时在同一事务内删除，发布版本变化时客户端不自动套用旧草稿。
- `workflow_process_instances.starter_id` 保存业务发起人，`operator_id` 保存实际发起操作人；普通用户自发时两者相同。
- 流程变量继续保存在 `workflow_process_variables`，仅用于条件、审批人解析和业务扩展。
- 发布版本 JSON 是表单 Schema、节点字段权限和流程配置的唯一运行依据。
- `user_reporting_relations` 保存直属、虚线等汇报关系及其生效区间；`manager` 审批人类型读取发起人当前生效的主直属关系，不再依赖 `users` 表字段。
- `workflow_org_approver_assignments` 通过 `subject_type + subject_id` 表达部门默认或指定人员的组织身份处理人。解析发起人组织身份时，指定人员配置优先，未配置时回退到发起人所属部门配置。

## 4. 扩展边界

应用服务在事务提交后发布通用生命周期事件：实例启动、任务完成、实例完成、实例拒绝、实例撤回和实例取消。通知、待办同步和具体业务回写通过订阅者接入，workflow 核心不直接依赖钉钉或绩效代码。

admin 与 client 的流程运行服务共享默认生命周期事件总线。业务模块按 `businessType` 注册状态回写器，回写器负责把通用流程状态转换成自己的业务状态并按 `businessKey` 更新业务记录：

```go
workflowapp.RegisterBusinessStatusUpdater("leave_request", workflowapp.BusinessStatusUpdaterFunc(
	func(ctx context.Context, update workflowapp.BusinessStatusUpdate) error {
		return leaveService.UpdateWorkflowStatus(ctx, update.BusinessKey, update.Status, update.InstanceID)
	},
))
```

状态回写器只接收实例启动、完成、拒绝、撤回和取消事件，不接收普通任务完成事件。处理器在流程事务提交后执行，因此必须按流程实例和事件类型实现幂等；回写失败会记录业务类型、业务标识、流程实例和事件类型，不会回滚已经提交的流程状态。

## 5. V1 不包含

- 加签、转交、委托、退回到任意节点和流程迁移。
- 定时器、自动服务任务、脚本任务和子流程。
- 可视化业务表单设计器与通用用户端页面。
- 绩效流程适配器及绩效数据迁移。
