# 明细列表列合计校验规则实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为工作流表单的明细列表增加可配置、可发布校验、前后端一致执行的“列合计”规则，并修复明细列表误出现普通字段比较及字段比较默认选中目标字段的问题。

**Architecture:** 在现有结构化表单规则模型中新增 `column_sum`，由流程定义版本持久化。设计器仅允许选择当前明细列表中的数字或金额列；前端负责即时反馈，后端负责发布时结构校验和提交时权威校验。明细行的必填与类型校验先于合计规则，空单元格在非必填场景按 `0` 计算。

**Tech Stack:** Go 1.24.5、Vue 3、TypeScript、Element Plus、Node.js 回归脚本

---

## 任务 1：用后端测试固定列合计语义

**文件：**
- 修改：`backend/internal/workflow/form_rules_test.go`

- [ ] 增加 `TestValidateFormDataSupportsDetailColumnSumRule`，构造包含 `number`、`amount` 和 `text` 列的明细列表。
- [ ] 覆盖合计通过、六种比较关系、空单元格按零、自定义错误提示和浮点金额求和。
- [ ] 覆盖必填数字列为空时优先返回列必填错误，而不是列合计错误。
- [ ] 增加发布结构校验用例：规则挂在非明细字段、列不存在、文本列、非法操作符、目标值缺失。
- [ ] 运行定向测试并确认因为 `column_sum` 模型尚不存在而失败：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -run 'TestValidateFormDataSupportsDetailColumnSumRule|TestValidateDefinitionRejectsInvalidStructuredRules' -count=1
```

## 任务 2：实现后端模型、发布校验和运行时校验

**文件：**
- 修改：`backend/internal/workflow/types.go`
- 修改：`backend/internal/workflow/validation.go`
- 修改：`backend/internal/workflow/form.go`
- 验证：`backend/internal/workflow/form_rules_test.go`

- [ ] 在规则常量中增加 `FormRuleColumnSum = "column_sum"`。
- [ ] 为 `FormValidationRule` 增加 `Column string` 和 `Value *float64` JSON 字段。
- [ ] 发布校验要求规则所属字段为 `detail_list`、引用列存在且为 `number`/`amount`、操作符属于 `eq/ne/gt/gte/lt/lte`、目标值存在且不是 NaN/Inf。
- [ ] 运行时遍历全部明细行，将空单元格按 `0` 计入，对有效数字求和后执行比较。
- [ ] 对合计比较使用前后端一致的浮点容差，避免 `0.1 + 0.2` 与 `0.3` 的等值误判。
- [ ] 默认错误信息包含明细字段名称、列名称、比较关系和目标值；自定义 `message` 优先。
- [ ] 运行任务 1 的定向测试并确认通过。

## 任务 3：用前端回归脚本固定运行时与设计器契约

**文件：**
- 修改：`admin/scripts/check-workflow-runtime-form.mjs`
- 修改：`admin/scripts/check-workflow-form-designer.mjs`

- [ ] 增加列合计通过和失败用例，包含空单元格按零、浮点金额、自定义提示。
- [ ] 增加必填列为空时优先显示列必填错误的断言。
- [ ] 增加结构断言：规则模型包含 `column`/`value`，设计器包含“列合计”，只枚举数字和金额列。
- [ ] 增加结构断言：明细列表不允许普通字段比较，新增字段比较不自动选择第一个目标字段。
- [ ] 运行两个检查脚本并确认在实现前失败：

```bash
cd admin
npm run check:workflow-runtime-form
npm run check:workflow-form-designer
```

## 任务 4：实现前端类型和运行时校验

**文件：**
- 修改：`admin/src/views/workflow/types.ts`
- 修改：`admin/src/views/workflow/runtimeForm.ts`

- [ ] 在 `WorkflowValidationRuleType` 中增加 `column_sum`，并为规则增加可选 `column`、`value`。
- [ ] 将明细列表校验顺序调整为：列表本身约束、逐行列约束、列表高级规则。
- [ ] 实现列合计规则：读取数组行、空值按零、数值求和、按操作符比较并生成默认提示。
- [ ] 使用与后端相同的浮点容差和比较边界。
- [ ] 运行 `npm run check:workflow-runtime-form` 并确认通过。

## 任务 5：实现设计器列合计配置并修正字段比较

**文件：**
- 修改：`admin/src/views/workflow/designer/components/WorkflowValidationRulesEditor.vue`
- 验证：`admin/scripts/check-workflow-form-designer.mjs`

- [ ] 计算当前明细列表中可合计的 `number`/`amount` 列。
- [ ] 仅当存在可合计列时，为明细列表显示“列合计”规则类型。
- [ ] 为规则显示“合计列”“比较关系”“目标值”和现有“错误提示”控件。
- [ ] 新建列合计规则时默认选择第一项可合计列、`eq` 和目标值 `100`。
- [ ] 将普通字段比较限制为后端支持的标量类型，排除 `detail_list`。
- [ ] 新建普通字段比较时将目标字段初始化为空，要求管理员明确选择。
- [ ] 运行 `npm run check:workflow-form-designer` 并确认通过。

## 任务 6：文档与完整回归

**文件：**
- 修改：`docs/architecture/generic-oa-workflow-v1.md`
- 参考：`docs/superpowers/specs/2026-09-01-workflow-detail-column-sum-rule-design.md`

- [ ] 在通用工作流架构文档中补充明细列合计能力及空值语义。
- [ ] 格式化 Go 文件：

```bash
cd backend
gofmt -w internal/workflow/types.go internal/workflow/validation.go internal/workflow/form.go internal/workflow/form_rules_test.go
```

- [ ] 运行后端回归：

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -count=1
GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/... ./internal/service/admin/workflow -count=1
```

- [ ] 运行前端回归和生产构建：

```bash
cd admin
npm run check:workflow-runtime-form
npm run check:workflow-form-designer
npm run check:workflow-runtime-pages
npm run check:workflow-tree
npm run build
```

- [ ] 对本次涉及文件执行 `git diff --check`，并逐项确认未覆盖工作区其他未提交修改。
- [ ] 本工作区存在大量用户未提交改动，本次不创建提交；完成后只报告实际修改文件和验证结果。
