# 客户端活动与报名后端优化计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 继续收敛客户端活动、报名接口的响应 DTO、数据库 context 链路和多表写入事务一致性。

**Architecture:** 本轮只处理客户端活动/报名这一条业务切片。服务层提供 `XXXContext` 入口，旧函数保留兼容；handler 只解析参数并调用 Context 版本。多表写入放入 GORM transaction，部门发布范围查询通过支持层 Context helper 传递。

**Tech Stack:** Go、Hertz、GORM、标准库 testing。

---

### Task 1: 结构约束

**Files:**
- Modify: `backend/internal/app/handler/handler_structure_test.go`
- Modify: `backend/internal/app/service/service_structure_test.go`

- [x] 客户端活动、报名核心 handler 禁止裸 `response.JSON(c, map[string]interface{})`。
- [x] 活动、报名客户端列表服务禁止返回裸 `map[string]interface{}`。
- [x] 活动、报名客户端列表服务禁止直接使用 `database.DB`。

### Task 2: 部门发布范围 Context helper

**Files:**
- Modify: `backend/internal/app/support/dept/dept.go`

- [x] 增加 `UserDeptIDContext`、`UserDeptIDsContext`、`UserDeptIDsByMiniOpenIDContext`。
- [x] 增加 `AncestorIDsContext`、`TopDeptNameContext`。
- [x] 旧函数保留并委托到 Context 版本。

### Task 3: 客户端活动服务 DTO/context/事务

**Files:**
- Modify: `backend/internal/app/service/event/client.go`
- Modify: `backend/internal/app/service/event/participation.go`
- Modify: `backend/internal/app/handler/client/event/handler.go`

- [x] `GetEventList` 返回 `ListResponse`。
- [x] 新增并使用 `GetEventListContext`。
- [x] 活动参与 `EventParticipateContext` 使用事务写参与记录和人数计数。
- [x] handler 使用 Context 版本。

### Task 4: 客户端报名服务 DTO/context/事务

**Files:**
- Modify: `backend/internal/app/service/enroll/client.go`
- Modify: `backend/internal/app/service/enroll/submission.go`
- Modify: `backend/internal/app/handler/client/enroll/handler.go`

- [x] `GetEnrollList` 返回 `ListResponse`。
- [x] 新增并使用 `GetEnrollListContext`。
- [x] 打卡提交 `EnrollJoinContext`、报名表单提交 `EnrollUserSubmitContext` 使用事务写记录和计数。
- [x] handler 使用 Context 版本和命名 DTO。

### Task 5: Verification

- [x] `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/support/dept ./backend/internal/app/service/event ./backend/internal/app/service/enroll ./backend/internal/app/handler/client/event ./backend/internal/app/handler/client/enroll ./backend/internal/app/handler ./backend/internal/app/service -count=1`
- [x] `GOCACHE=$PWD/.cache/go-build go test ./backend/...`
- [x] `git diff --check`
