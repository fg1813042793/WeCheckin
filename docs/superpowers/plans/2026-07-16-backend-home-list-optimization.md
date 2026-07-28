# 后端首页聚合服务优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将客户端首页聚合服务从动态 `map` 和全局 DB 查询迁移到类型化响应与请求上下文。

**Architecture:** 保持 `/home/list` 返回字段兼容，新增 `ListResponse` 和 `ListItem` DTO。服务层提供 `GetHomeListContext`，旧 `GetHomeList` 委托到 Context 版本；公开首页 handler 改用请求上下文。

**Tech Stack:** Go、Hertz、GORM、项目现有 `database.WithContext`、结构测试。

---

### Task 1: 结构测试先行

**Files:**
- Modify: `backend/internal/app/service/service_structure_test.go`

- [x] **Step 1: 写失败测试**

```go
func TestHomeListServiceUsesContextAndTypedResponses(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("home", "list.go"))
	if err != nil {
		t.Fatalf("read home/list.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		") (map[string]interface{}, error)",
		"[]map[string]interface{}",
		"map[string]interface{}{",
		"database.DB.",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("home/list.go must use context-aware queries and typed responses instead of %q", snippet)
		}
	}
}
```

- [x] **Step 2: 运行失败测试**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service -run TestHomeListServiceUsesContextAndTypedResponses -count=1`

Expected: FAIL，指出 `home/list.go` 仍存在动态响应或直接 `database.DB`。

### Task 2: 首页聚合类型化

**Files:**
- Modify: `backend/internal/app/service/home/list.go`

- [x] **Step 1: 新增 `ListResponse` 和 `ListItem` DTO**
- [x] **Step 2: 基础查询和配置加载改为接收 `*gorm.DB`**
- [x] **Step 3: `GetHomeListContext` 使用请求上下文、检查查询错误**
- [x] **Step 4: 保留 `GetHomeList` 兼容旧调用**

### Task 3: Handler 上下文化

**Files:**
- Modify: `backend/internal/app/handler/public/home/handler.go`

- [x] **Step 1: `GetSetup` 改用 `setupservice.GetSetupContext`**
- [x] **Step 2: `GetHomeList` 改用 `homeservice.GetHomeListContext`**

### Task 4: 验证

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/home ./backend/internal/app/handler/public/home ./backend/internal/app/service -count=1
GOCACHE=$PWD/.cache/go-build go test ./backend/...
git diff --check
git diff --cached --check
```

实际执行结果：

- `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service -run TestHomeListServiceUsesContextAndTypedResponses -count=1`：先失败后通过。
- `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/home ./backend/internal/app/handler/public/home ./backend/internal/app/service -count=1`：通过。
- `GOCACHE=$PWD/.cache/go-build go test ./backend/...`：通过。
- `git diff --check`：通过。
- `git diff --cached --check`：通过。
