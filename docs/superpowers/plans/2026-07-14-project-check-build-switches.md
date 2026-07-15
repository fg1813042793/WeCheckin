# 项目级检查构建开关 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 扩展 `scripts/check.sh`，在保持默认快速检查的同时支持按需运行前端 H5 构建和管理后台构建。

**Architecture:** `scripts/check.sh` 默认继续运行后端 Go 测试、前端配置/请求检查和管理后台请求/导航检查。新增 `CHECK_FRONTEND_BUILD=1`、`CHECK_ADMIN_BUILD=1` 和 `CHECK_BUILDS=1` 三个环境变量开关：单独打开对应构建，或一次打开所有构建。README 的测试段落补充中文用法，并说明管理后台构建当前可能暴露既有 Vue 类型问题。

**Tech Stack:** Bash、npm scripts、Markdown。

---

## File Structure

- Modify: `scripts/check.sh`
  - 增加构建开关读取函数。
  - 根据开关执行 `npm --prefix frontend run build:h5` 和 `npm --prefix admin run build`。
- Modify: `README.md`
  - 更新项目级检查覆盖范围。
  - 说明可选构建开关命令。

---

## Tasks

- [x] 修改 `scripts/check.sh`，新增 `enabled()` 判断函数和构建开关变量。
- [x] 在默认检查之后追加可选前端 H5 构建。
- [x] 在默认检查之后追加可选管理后台构建。
- [x] 更新 README 中文测试说明，覆盖默认检查和可选构建命令。
- [x] 运行 `bash -n scripts/check.sh`。
- [x] 运行 `bash scripts/check.sh`。
- [x] 运行 `CHECK_FRONTEND_BUILD=1 bash scripts/check.sh`。
- [x] 记录 `npm --prefix admin run build` 的既有失败项，避免误报为本次脚本开关问题。
