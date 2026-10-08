# H5App 流程表单详情分组查看设计

## 目标

- 独立流程表单详情页支持与发起页、审批页一致的“分组显示 / 连续显示”切换。
- PC 端复用现有表单展示偏好，用户从填写、审批或查看详情切换模式后，其他入口保持相同模式。
- 手机端继续使用连续显示，避免窄屏页签拥挤和横向操作负担。

## 当前问题

`WorkflowRuntimeForm` 已支持 `continuous` 和 `tabs` 两种展示方式，发起页与审批任务页也已接入切换按钮和 `workflow_start_form_presentation_v1` 本地偏好。

截图中的页面由 `WorkflowFormDetailPage.vue` 独立渲染。该页面当前没有传入 `presentation`，因此始终使用组件默认的连续模式，也没有展示切换按钮。

## 交互设计

在独立表单详情页顶部右侧增加操作区域，将展示切换按钮放在流程状态标签左侧：

- 当前为连续显示时，按钮文案为“分组显示”。
- 当前为分组显示时，按钮文案为“连续显示”。
- PC 端显示按钮并允许切换。
- 手机端不显示按钮，表单固定使用连续显示。

切换只改变只读表单的展示方式，不修改表单值、字段权限、流程状态或后端数据。

## 实现边界

- 修改 `h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue`，增加展示状态、桌面判断、窗口尺寸同步、切换按钮和 `WorkflowRuntimeForm` 的 `presentation` 绑定。
- 继续使用现有存储键 `workflow_start_form_presentation_v1`，不新增另一份详情页偏好。
- 修改现有未跟踪契约脚本 `h5app/scripts/check-workflow-form-tabs.mjs`，将独立详情页纳入表单页签契约。
- 不修改 `WorkflowRuntimeForm.vue` 的页签实现，不改变历史详情上下分栏中的连续模式，不修改 API、DTO 或后端。

## 验证

- 先扩充 `check-workflow-form-tabs.mjs`，确认详情页尚未接入时检查失败。
- 实现后运行 `pnpm check:workflow-form-tabs`、针对性 ESLint、`pnpm type-check`、`pnpm check:ui-style` 和 `pnpm build:h5`。
- 浏览器验证标准 PC、PC 小屏和手机宽度：PC 可切换且共享偏好，手机无切换按钮并连续显示，详情页无横向溢出或标题操作重叠。
- 工作区中已有其他流程改动导致的无关失败单独记录，不覆盖或回退用户修改。
