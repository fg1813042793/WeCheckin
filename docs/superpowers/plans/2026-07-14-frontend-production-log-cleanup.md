# 前端生产调试日志清理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 清理 uni-app 源码中的无意义生产调试输出，并增加静态检查防止 `console.log/debug/info` 回流。

**Architecture:** 本轮只扫描 `frontend/App.vue`、`frontend/pages`、`frontend/components`、`frontend/utils` 和 `frontend/config`，不扫描 `frontend/miniprogram`、`frontend/uni_modules` 等生成或第三方目录。保留 `console.error` 和 `console.warn`，因为它们仍承担错误诊断职责。新增 `frontend/scripts/check-production-logs.mjs` 并接入项目级检查。

**Tech Stack:** uni-app、Node.js 静态检查、Bash 检查脚本。

---

## File Structure

- Create: `frontend/scripts/check-production-logs.mjs`
  - 扫描源码目录中的 `console.log/debug/info`。
- Modify: `frontend/package.json`
  - 增加 `check:logs`。
- Modify: `frontend/App.vue`
  - 移除生命周期和平台判断调试输出。
- Modify: `frontend/pages/enroll/enroll_detail.vue`
  - 移除位置字段调试输出。
- Modify: `scripts/check.sh`
  - 接入 `npm --prefix frontend run check:logs`。

---

## Tasks

- [x] 增加失败检查：`check:logs` 扫描源码目录并在发现 `console.log/debug/info` 时失败。
- [x] 运行 `npm --prefix frontend run check:logs`，确认当前调试输出会触发失败。
- [x] 移除 `frontend/App.vue` 中的启动/显示/隐藏/平台调试输出。
- [x] 移除 `frontend/pages/enroll/enroll_detail.vue` 中的位置字段调试输出。
- [x] 修改 `scripts/check.sh`，纳入前端日志检查。
- [x] 运行 `npm --prefix frontend run check:logs`。
- [x] 运行 `bash scripts/check.sh`。
