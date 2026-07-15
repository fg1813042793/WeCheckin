# 管理后台类型质量收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复管理后台 Vue/TypeScript 构建错误，让 `CHECK_ADMIN_BUILD=1 bash scripts/check.sh` 可以通过。

**Architecture:** 不调整业务交互，只修复类型系统已经暴露的真实不一致：模板引用脚本中不存在的属性或函数、`v-for` 与 `v-if` 同节点导致模板作用域推断失败、Axios 拦截器返回数据与 TypeScript 类型不一致、缺少第三方库声明，以及局部空值推断问题。以当前失败的 `npm --prefix admin run build` 作为红灯测试，逐组修复后重新运行构建验证。

**Tech Stack:** Vue 3、TypeScript、vue-tsc、Vite、Element Plus。

---

## File Structure

- Modify: `admin/src/views/exam/ExamDesigner.vue`
  - 修复 `selected.value` 空值推断。
- Modify: `admin/src/views/exam/ExamFillPC1.vue`
  - 模板跳转改为脚本函数。
  - 补齐扫码弹窗打开/关闭回调。
- Modify: `admin/src/views/survey/SurveyDesigner.vue`
  - 移除不存在的 `genQR` 调用。
  - 按拦截器返回值实际类型处理接口响应。
- Modify: `admin/src/views/survey/SurveyFillPC.vue`
  - 拆开同节点 `v-for` / `v-if`。
- Modify: `admin/src/views/survey/SurveyFillPC1.vue`
  - 模板跳转改为脚本函数。
  - 拆开同节点 `v-for` / `v-if`。
  - 补齐扫码弹窗打开/关闭回调。
- Modify: `admin/src/views/survey/SurveyStatistic.vue`
  - 通过本地声明解决 `leaflet` 类型缺失。
- Modify: `admin/src/views/survey/SurveyStatReport.vue`
  - 明确 `Object.entries` 的值类型。
- Modify: `admin/env.d.ts`
  - 增加 `leaflet` 模块声明。

---

## Tasks

- [x] 运行 `CHECK_ADMIN_BUILD=1 bash scripts/check.sh`，确认管理后台构建失败并记录错误族。
- [x] 修复模板直接引用 `window` 和不存在函数的问题。
- [x] 修复扫码弹窗缺失回调。
- [x] 修复 `v-for` / `v-if` 同节点作用域推断问题。
- [x] 修复 Axios 拦截器返回值类型不一致问题。
- [x] 修复 `leaflet` 声明和 `Object.entries` unknown 类型。
- [x] 运行 `npm --prefix admin run build`。
- [x] 运行 `CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。

## 验证记录

- `npm --prefix admin run build`：通过，Vite 仅提示产物 chunk 体积 warning。
- `CHECK_ADMIN_BUILD=1 bash scripts/check.sh`：通过，覆盖后端测试、uni-app 静态检查、管理后台 request/navigation 检查和管理后台构建。
