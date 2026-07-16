# 后端用户收藏与首页服务优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 继续收口后端高频服务中的动态响应、全局数据库调用和跨表写入一致性问题。

**Architecture:** 本轮优先优化 `favorite`、`dashboard`、`adminuser` 以及权限范围辅助函数。新增 Context 版本保留旧函数兼容，handler 改用请求上下文，列表和首页响应改为具名 DTO，用户部门写入放入事务。

**Tech Stack:** Go、Hertz、GORM、项目现有 `database.WithContext`、结构测试。

---

### Task 1: 结构测试先行

**Files:**
- Modify: `backend/internal/app/service/service_structure_test.go`

- [x] **Step 1: 写失败测试**

```go
func TestUserFavoriteDashboardServicesUseContextAndTypedResponses(t *testing.T) {
	for _, file := range []string{
		filepath.Join("favorite", "service.go"),
		filepath.Join("dashboard", "service.go"),
		filepath.Join("adminuser", "service.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			") (map[string]interface{}, error)",
			") ([]map[string]interface{}, error)",
			"return map[string]interface{}{",
			"database.DB.",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must use context-aware queries and typed responses instead of %q", file, snippet)
			}
		}
	}
}
```

- [x] **Step 2: 运行失败测试**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service -run TestUserFavoriteDashboardServicesUseContextAndTypedResponses -count=1`

Expected: FAIL，指出当前服务仍有动态 map 或直接 `database.DB` 调用。

### Task 2: 收藏服务类型化与上下文化

**Files:**
- Modify: `backend/internal/app/service/favorite/service.go`
- Modify: `backend/internal/app/handler/client/favorite/handler.go`

- [x] **Step 1: 增加 `FavoriteListItem` DTO 和 Context 版本函数**
- [x] **Step 2: `GetMyFavListContext` 返回 `[]FavoriteListItem`，保留 `GetMyFavList` 兼容旧调用**
- [x] **Step 3: handler 改用 `UpdateFavContext`、`DelFavContext`、`IsFavContext`、`GetMyFavListContext`**

### Task 3: 后台首页类型化与权限范围上下文化

**Files:**
- Modify: `backend/internal/app/support/access/access.go`
- Modify: `backend/internal/app/service/dashboard/service.go`
- Modify: `backend/internal/app/handler/admin/home/handler.go`

- [x] **Step 1: 给 `VisibleDeptIDs`、`DataScopeFilter` 增加 Context 版本**
- [x] **Step 2: `dashboard.AdminHomeContext` 返回 `AdminHomeResponse`**
- [x] **Step 3: 首页 handler 使用请求上下文，清除推荐也走 Context 版本**

### Task 4: 用户服务事务与上下文

**Files:**
- Modify: `backend/internal/app/service/adminuser/service.go`
- Modify: `backend/internal/app/handler/admin/user/handler.go`

- [x] **Step 1: 新增 `UserDetail` DTO，替换 `GetUserByID` 的动态 map**
- [x] **Step 2: 新增 `saveUserDeptsTx`，新增/编辑用户时用事务包住用户和部门关系**
- [x] **Step 3: 删除、状态、重置密码等写操作增加 Context 版本，handler 迁移到 Context 版本**

### Task 5: 验证

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/favorite ./backend/internal/app/service/dashboard ./backend/internal/app/service/adminuser ./backend/internal/app/handler/client/favorite ./backend/internal/app/handler/admin/home ./backend/internal/app/handler/admin/user ./backend/internal/app/service -count=1
GOCACHE=$PWD/.cache/go-build go test ./backend/...
git diff --check
```

实际执行结果：

- `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service -run TestUserFavoriteDashboardServicesUseContextAndTypedResponses -count=1`：先失败后通过。
- `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/favorite ./backend/internal/app/service/dashboard ./backend/internal/app/service/adminuser ./backend/internal/app/service/setup ./backend/internal/app/support/access ./backend/internal/app/handler/client/favorite ./backend/internal/app/handler/admin/home ./backend/internal/app/handler/admin/user ./backend/internal/app/service -count=1`：通过。
- `GOCACHE=$PWD/.cache/go-build go test ./backend/...`：通过。
- `git diff --check`：通过。
- `git diff --cached --check`：通过。
