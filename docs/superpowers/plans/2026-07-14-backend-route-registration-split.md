# 后端路由注册拆分 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `backend/cmd/main.go` 中的大段路由注册拆出到独立文件，降低启动入口维护成本。

**Architecture:** 新增同包文件 `backend/cmd/routes.go`，提供 `registerRoutes(h *server.Hertz)`，承接 Swagger、public/client/admin、上传、静态文件和 Admin SPA 注册逻辑。`main.go` 只保留配置加载、数据库/Redis/日志初始化、中间件挂载和 `registerRoutes(h)` 调用。新增源码安全测试防止路由注册再次回流到 `main.go`。

**Tech Stack:** Go、Hertz、Go test。

---

## File Structure

- Create: `backend/cmd/routes.go`
  - 保存所有 HTTP 路由注册。
- Modify: `backend/cmd/main.go`
  - 删除直接路由注册逻辑。
  - 调用 `registerRoutes(h)`。
  - 清理不再需要的 import。
- Modify: `backend/cmd/main_safety_test.go`
  - 增加 main 入口只委托路由注册的检查。

---

## Tasks

- [x] 增加失败测试：`main.go` 必须包含 `registerRoutes(h)`，且不应包含 `adminGroup.GET` / `h.POST("/upload"` 等大段路由注册。
- [x] 新增 `routes.go` 并搬迁路由注册逻辑。
- [x] 精简 `main.go` imports。
- [x] 运行 `gofmt -w backend/cmd/main.go backend/cmd/routes.go backend/cmd/main_safety_test.go`。
- [x] 运行 `GOCACHE=$PWD/.cache/go-build go test ./backend/cmd -count=1`。
- [x] 运行 `bash scripts/check.sh`。
