# FormRender Calculation Runtime Design

## 背景

管理后台 Formkit 设计器已经支持 `calcValue` 计算表达式，后端提交时也会通过 formkit calc 引擎重新计算答案。但移动端 `FormRender` 当前只接入了基础显隐/必填逻辑，填写过程中不会即时展示计算结果，自动保存的数据也缺少本地计算回填。

## 目标

- 在移动端表单渲染时支持 `calcValue.expr` 的即时本地求值。
- 与后端保持同类表达式语义：题目 ID 变量、数字/字符串/布尔/null、算术、比较、逻辑、三元表达式和常用函数。
- 计算失败时静默跳过，后端提交链路仍作为最终兜底。
- 用静态检查防止 `FormRender` 再次丢失计算字段能力。

## 非目标

- 不把后端 calc 引擎改写成共享运行时。
- 不在本轮实现分支跳转导航和提交前后端联动校验。
- 不改变答卷提交 API 和后端最终计算逻辑。

## 方案

新增 `frontend/utils/formkitCalc.js`，实现一个无 `eval`/`Function` 的轻量表达式求值器：

- Tokenizer 支持数字、字符串、标识符、操作符、括号和逗号。
- Pratt/递归下降解析按后端优先级求值：一元、乘除模、加减、比较、相等、逻辑与或、三元。
- 函数支持 `contains`、`empty`、`len`、`if`、`sum`、`avg`，并兼容后台提示里的大写 `SUM`、`AVG`、`CONCATENATE`、`AND`、`OR`、`NOT`、`ISBLANK`、`IFS`。
- 变量环境同时暴露题目原始 ID 与 `Q1/Q2` 别名，兼容旧设计器公式习惯。

`FormRender` 在 schema 初始化和每次 `setVal` 后调用 `applyFormkitCalcValues()`，把计算结果写回目标答案；对于计算目标题，输入控件置为只读/禁用，避免用户修改后马上被计算覆盖造成误解。

## 验收

- `frontend/scripts/check-formkit-logic.mjs` 覆盖计算运行时、FormRender 导入和调用点。
- `npm --prefix frontend run check:formkit-logic` 通过。
- `bash scripts/check.sh` 通过。
