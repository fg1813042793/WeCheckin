# Skip Empty Approval Nodes Implementation Plan

**Goal:** 审批节点支持可选 `skipIfNoAssignee` 开关，默认关闭；无匹配审批人时可自动继续下一节点，并记录跳过原因。

**Contract:** 仅 approval 节点可启用。仍须配置合法审批人来源；仅成功解析但候选为空（含逐级审批全部层级为空）可以跳过。查询异常、无效配置及逐级审批自身配置的阻断错误不能跳过。未启用时保留原错误。以发布版本为准，不回写历史实例。

**Implementation:** 后端定义/BPMN/运行时/节点进度由独立 worker 处理；主任务负责 Admin 开关、前后端类型、Admin/H5 历史标签、文档、Swagger 和浏览器验证。沿用既有事务与流程推进路径，不创建虚假审批任务或审批通过记录，也不发送跳过节点的待办通知。

- [x] 定义、审批引擎及节点进度：开关开关对照、空候选、解析失败、全部为空、多模式、并行、后续审批/结束、历史记录与发布版本回归。
- [x] 配置页面：默认关闭、可切换/保存重载、只读禁止修改、只在审批节点展示；说明逐级审批的层级缺失策略仍有效。
- [x] 历史与跨端契约：`node_skipped` 显示自动跳过；进度显示 skipped；同步 Swagger/使用文档。
- [x] 验证 Backend / Admin / H5 的受影响测试、静态检查与构建及浏览器；报告基线失败，不提交、部署或修改生产数据。

## Verification

- workflowcore / workflow domain / application tests passed, including all four approval modes, empty/nonempty candidates, resolver error propagation, layered approval, parallel joins, automatic-transition limit, persistence/error propagation and skip-progress revisits.
- Admin check:workflow-tree and check:workflow-org-approver passed; real NodeInspector browser fixture verified default-off, toggle, JSON clone/save/reload, readonly, approval-only and chain hint.
- Admin build, H5 type-check/build:h5 and changed H5 file lint passed.
- Full backend tests retain five existing failing packages (Swagger payload, DingTalk settings, audit classification, config secret guard, HTTP raw-error guard). Admin check:all remains blocked by user/index.vue line budget; H5 full lint retains existing215 errors/2warnings.
- No production writes, real notifications, commits or deployment. Temporary browser fixture removed and test service closed.
