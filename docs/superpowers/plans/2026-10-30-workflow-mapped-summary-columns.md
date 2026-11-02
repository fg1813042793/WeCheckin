# Workflow Mapped Summary Columns Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development for the isolated mapping engine; integrate the editor, historical schema loading and browser verification in the current task.

**Goal:** 汇总中配置一个展示列，显式映射多个流程的字段，按记录所属流程显示各自结果。

**Architecture:** H5 继续使用已授权的汇总列表数据。新增本地展示配置，来源由流程 ID、字段 key 和可选版本组成。默认来源使用该流程当前发布字段；精确版本来源优先，通过现有汇总详情接口读取当前页可见实例的历史表单。原始申请、流程定义及权限不变。

**Tech Stack:** Vue 3、TypeScript、uni-app、uView Pro，现有汇总 API 与计算字段求值函数。

- [x] 纯映射模块和测试：结构校验、序列化规范化、按流程/版本精确匹配、来源缺失提示、保存结果优先、补算、不可变输入。
- [x] 编辑弹窗：列名称、多行流程/版本/字段来源，新增/编辑/删除，取消不提交；仅展示管理名称和字段标签，不展示内部 key。不同类型给出业务口径提醒。
- [x] 版本字段读取：按当前页实例提供版本选择，仅经汇总详情接口读取表单；缓存按流程+版本区分，拒绝/失败不回退或使用旧数据。
- [x] 接入列设置：保存到现有 v2 浏览器配置的 mappedColumns 字段，保持旧单流程选择；支持显示开关、恢复默认、删除列；未知/已删除来源保留配置并显示原因。
- [x] 表格：统一列展示一列，按记录选择来源；补算继续标记，缺失通过提示解释；配置只改变展示，不修改现有表单导出。
- [x] 验证：映射与版本读取专项脚本、H5 lint/type-check/build、现有汇总回归；浏览器覆盖通用/William 不同字段 key、补算、未映射、版本覆盖、持久化、编辑取消/删除及普通设置兼容。

No commits, deployment or unrelated worktree changes are part of this task.


## 验证结果

- `pnpm check:workflow-summary-mappings`：25 项映射行为测试及配置/历史字段加载专项通过；包括权限失效、过期响应，以及定义和列表两种刷新顺序。
- `pnpm check:workflow-summary-columns`、`check:workflow-summary-column-settings`、`check:workflow-summary-export`、`check:ui-style` 通过。
- `pnpm type-check`、`pnpm build:h5`、全部本次修改文件的 ESLint 通过。运行 pnpm 时关闭了启动器自动切换版本和自动依赖安装检查，未改依赖。
- `pnpm lint` 仍有工作区既有 213 个错误及 2 个警告；`check:workflow-module` 仍被既有“发起审批”文案断言拦截。
- 本机浏览器演示数据验证：创建两流程不同 key 的同一列，分别使用各自公式，保留保存值/0；未映射、数据缺失显示原因；添加历史版本覆盖、取消编辑、刷新保留、隐藏/恢复、重复来源拦截、历史接口拒绝与重试、删除取消/确认。所有请求均为 GET。
- 1440/900/390px 下检查编辑器布局；窄屏弹窗约 359px 宽，页脚可见，内部滚动。配置入口及自定义表格字段仍沿用现有 PC 列设置和手机精简展示规则。
- 两阶段代码审查发现并修复了定义更新时历史字段没有重载的问题，复核通过。已有 uni scroll-view 初始化空引用警告仍存在，不属于本次改动；未验证真实钉钉客户端。
