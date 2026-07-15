# FormRender Calculation Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让移动端 `FormRender` 在填写过程中即时回填 Formkit `calcValue` 计算字段。

**Architecture:** 新增前端轻量表达式求值器；`FormRender` 在初始化和答案变化后应用计算结果；静态检查覆盖导入、调用和禁用态。

**Tech Stack:** UniApp/Vue Options API、JavaScript 表达式解析、Node.js 静态检查、中文文档。

---

## File Structure

- Create: `frontend/utils/formkitCalc.js`
  - 安全表达式求值和 `applyFormkitCalcValues()`。
- Modify: `frontend/components/formkit/FormRender.vue`
  - 初始化/答题变化后应用计算结果。
  - 计算目标字段只读/禁用。
- Modify: `frontend/utils/formkit.js`
  - 规范化 schema 时保留 `props`，兼容 `props.calculateFormula`。
- Modify: `frontend/scripts/check-formkit-logic.mjs`
  - 扩展计算字段静态检查。

## Tasks

- [x] 编写中文设计文档。
- [x] 扩展静态检查，并确认当前实现红灯。
- [x] 新增 `frontend/utils/formkitCalc.js`。
- [x] 改造 `FormRender.vue` 应用计算结果。
- [x] 保留 schema props 以兼容旧公式配置。
- [x] 运行 `npm --prefix frontend run check:formkit-logic`、`CHECK_FRONTEND_BUILD=1 bash scripts/check.sh`。
