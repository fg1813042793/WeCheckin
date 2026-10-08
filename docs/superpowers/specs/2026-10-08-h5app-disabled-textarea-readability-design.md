# H5App 禁用长文本可读性优化设计

## 目标

- 提高表单查看场景中长文本内容的对比度，避免已填写内容显示为浅灰色。
- 所有使用 `WorkflowTextarea` 的禁用长文本统一使用正常正文色。
- 继续通过浅灰背景、边框和禁用光标表达不可编辑状态。

## 当前问题

H5 端 `WorkflowTextarea` 使用浏览器原生 `textarea`。组件在禁用状态下将 `.workflow-textarea__native-control` 设置为 `var(--u-light-color)`，浏览器计算结果约为 `rgb(192, 196, 204)`，导致浅灰文字叠加浅灰背景，阅读对比度不足。

`WorkflowRuntimeForm` 已尝试为普通 uView 输入框和 textarea 覆盖只读文字色，但该选择器没有命中新加入的原生 `.workflow-textarea__native-control`，因此截图中的长文本仍然偏灰。

## 设计

按用户选择的全局方案，直接修改 `WorkflowTextarea` 的禁用态样式：

- H5 原生 textarea 的已填写内容使用 `var(--u-main-color)`，并同步设置 `-webkit-text-fill-color`，避免浏览器禁用态默认颜色覆盖。
- 非 H5 的 uView textarea 禁用字段同样覆盖为 `var(--u-main-color)`，保持跨端一致。
- placeholder 继续使用 `var(--u-light-color)`。
- 字数统计继续使用辅助灰色，不与正文争夺视觉层级。
- 禁用背景、边框和 `not-allowed` 光标保持不变，避免深色文字被误认为可编辑。

该改动会影响所有使用 `WorkflowTextarea` 且传入 `disabled` 的场景，不仅限于流程表单详情页。这是方案 2 的明确范围。

## 实现边界

- 修改 `h5app/src/pages/workflow/components/WorkflowTextarea.vue` 的禁用态样式，不改变组件 props、事件、IME 输入逻辑或自动高度逻辑。
- 扩充现有未跟踪脚本 `h5app/scripts/check-workflow-textarea-ime.mjs`，断言 H5 和非 H5 禁用文字都使用正文色，并保留 placeholder 辅助色。
- 不修改普通 input、数字、选择器或空值展示；这些控件现有只读颜色已经正常。
- 不修改接口、表单数据、字段权限或后端。

## 验证

- 先扩充 `check-workflow-textarea-ime.mjs`，确认当前禁用文字仍为浅色时检查失败。
- 实现后运行 `pnpm check:workflow-textarea-ime`、针对性 ESLint、`pnpm type-check`、`pnpm check:ui-style` 和 `pnpm build:h5`。
- 浏览器检查标准 PC、PC 小屏和手机宽度下的有值长文本：正文清晰可读，placeholder 和辅助信息仍为灰色，背景和边框仍能表达不可编辑状态。
- 记录全量 lint 和 workflow-module 的既有无关失败，不覆盖或回退当前工作区修改。
