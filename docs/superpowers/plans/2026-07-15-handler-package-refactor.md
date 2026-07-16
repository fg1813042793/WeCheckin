# Handler 分包重构执行计划

## 当前进度

- [x] P1：问卷、考试 handler 已拆到 `admin/*` 与 `client/*` 子包。
- [x] P2：活动、打卡 handler 已拆到对应领域子包。
- [x] P3：后台基础管理、系统管理、新闻、客户端账号/收藏、公开首页/地理编码入口已拆分完成。
- [x] 根 `handler` 包保留结构测试与安全测试，具体业务 handler 统一从 `admin`、`client`、`public` 子包导入。

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the largest handler files into focused admin/client domain packages while keeping route registration centralized in `backend/cmd`.

**Architecture:** Keep existing route files as the composition layer. Move survey and exam handlers into `handler/admin/*` and `handler/client/*` packages first, because they are the largest and already depend on domain service packages. Add structure tests so moved files do not drift back into the root handler package.

**Tech Stack:** Go, Hertz, existing `backend/internal/app/handler` and `backend/cmd/routes_*.go`.

---

### Task 1: Add Handler Structure Guard

**Files:**
- Create: `backend/internal/app/handler/handler_structure_test.go`

- [ ] **Step 1: Write failing structure test**

Require these files:
- `admin/survey/handler.go`
- `admin/survey/formkit.go`
- `admin/exam/handler.go`
- `client/survey/handler.go`
- `client/exam/handler.go`

Forbid these root files:
- `survey_admin.go`
- `survey_formkit.go`
- `exam_admin.go`
- `survey.go`
- `exam.go`

- [ ] **Step 2: Run red test**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/handler -run TestLargeHandlersUseDomainSubpackages -count=1`

Expected: fail because the new subpackage files do not exist yet.

### Task 2: Move Survey Handlers

**Files:**
- Move: `backend/internal/app/handler/survey_admin.go` -> `backend/internal/app/handler/admin/survey/handler.go`
- Move: `backend/internal/app/handler/survey_formkit.go` -> `backend/internal/app/handler/admin/survey/formkit.go`
- Move: `backend/internal/app/handler/survey.go` -> `backend/internal/app/handler/client/survey/handler.go`
- Modify: `backend/cmd/routes_client.go`
- Modify: `backend/cmd/routes_survey_exam.go`

- [ ] **Step 1: Move files and update package names**

Admin survey files use `package survey`; client survey file uses `package survey`.

- [ ] **Step 2: Update route imports**

Use aliases:
- `adminsurvey "wecheckin-backend/backend/internal/app/handler/admin/survey"`
- `clientsurvey "wecheckin-backend/backend/internal/app/handler/client/survey"`

- [ ] **Step 3: Run focused tests**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/handler ./backend/cmd -count=1`

Expected: pass.

### Task 3: Move Exam Handlers

**Files:**
- Move: `backend/internal/app/handler/exam_admin.go` -> `backend/internal/app/handler/admin/exam/handler.go`
- Move: `backend/internal/app/handler/exam.go` -> `backend/internal/app/handler/client/exam/handler.go`
- Modify: `backend/cmd/routes_client.go`
- Modify: `backend/cmd/routes_survey_exam.go`

- [ ] **Step 1: Move files and update package names**

Admin exam file uses `package exam`; client exam file uses `package exam`.

- [ ] **Step 2: Update route imports**

Use aliases:
- `adminexam "wecheckin-backend/backend/internal/app/handler/admin/exam"`
- `clientexam "wecheckin-backend/backend/internal/app/handler/client/exam"`

- [ ] **Step 3: Run focused tests**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/handler ./backend/cmd -count=1`

Expected: pass.

### Task 4: Final Verification

**Files:**
- Modify: `backend/internal/app/handler/handler_structure_test.go`

- [ ] **Step 1: Run backend tests**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/...`

Expected: pass.

- [ ] **Step 2: Run diff whitespace check**

Run: `git diff --check`

Expected: no output.
