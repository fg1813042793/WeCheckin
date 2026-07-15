# Debug Token 接口治理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除后端公开 `/test/debug_token` 接口，并用自动化测试防止公开 token 调试入口回归。

**Architecture:** 在 `backend/cmd` 增加轻量级源码安全测试，直接扫描 `main.go`，确认公开 debug token 路由不再存在。随后删除 `main.go` 中 public routes 下的 `/test/debug_token` handler，并把 `./backend/cmd` 加入 `scripts/check.sh`，让项目级检查覆盖这类启动入口安全问题。

**Tech Stack:** Go test、Bash、Hertz 路由源码、Markdown。

---

## File Structure

- Create: `backend/cmd/main_safety_test.go`
  - 防回归测试，禁止公开 `/test/debug_token` 路由出现在 `main.go`。
- Modify: `backend/cmd/main.go`
  - 删除 public routes 下的 `/test/debug_token` handler。
  - 删除不再使用的 `tokenutil` import。
- Modify: `scripts/check.sh`
  - 将 `./backend/cmd` 纳入项目级检查。

---

### Task 1: 增加公开 debug token 防回归测试

**Files:**
- Create: `backend/cmd/main_safety_test.go`

- [ ] **Step 1: 写入失败测试**

Create `backend/cmd/main_safety_test.go` with:

```go
package main

import (
	"os"
	"strings"
	"testing"
)

func TestPublicDebugTokenRouteIsNotRegistered(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	if strings.Contains(string(src), `"/test/debug_token"`) {
		t.Fatalf("public debug token route must not be registered")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/cmd -run TestPublicDebugTokenRouteIsNotRegistered -count=1
```

Expected: FAIL，输出包含 `public debug token route must not be registered`。

---

### Task 2: 移除公开 debug token 路由

**Files:**
- Modify: `backend/cmd/main.go`

- [ ] **Step 1: 删除 public routes 下的 debug token handler**

Remove the `h.GET("/test/debug_token", ...)` block from `backend/cmd/main.go`.

- [ ] **Step 2: 删除不再使用的 import**

Remove this import if it is no longer referenced:

```go
"wecheckin-backend/backend/pkg/tokenutil"
```

- [ ] **Step 3: 运行聚焦测试确认通过**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/cmd -run TestPublicDebugTokenRouteIsNotRegistered -count=1
```

Expected: PASS。

---

### Task 3: 扩展项目级检查脚本

**Files:**
- Modify: `scripts/check.sh`

- [ ] **Step 1: 将 backend/cmd 加入 go test 包列表**

Change the command in `scripts/check.sh` to:

```bash
GOCACHE="${GOCACHE_DIR}" go test \
  ./backend/cmd \
  ./backend/internal/app/service \
  ./backend/internal/config \
  ./backend/internal/app/formkit/...
```

- [ ] **Step 2: 运行项目级检查**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS，输出包含 `ok  	wecheckin-backend/backend/cmd` 和 `==> Checks passed`。

- [ ] **Step 3: 确认缓存目录已清理**

Run:

```bash
test ! -e .cache
```

Expected: PASS，无输出，退出码为 0。

---

### Task 4: 最终核验

**Files:**
- Test: `backend/cmd/main_safety_test.go`
- Test: `scripts/check.sh`

- [ ] **Step 1: 搜索公开 debug token 路由**

Run:

```bash
rg -n '"/test/debug_token"|debug_token' backend/cmd backend/internal/app/handler backend/internal/app/service
```

Expected: 不再出现 `"/test/debug_token"`。后台受鉴权保护的 `setup_debug_token` 如果仍存在，可以保留。

- [ ] **Step 2: 查看工作区状态**

Run:

```bash
git status --short
```

Expected: 出现本批次新增的计划文档、测试文件、`main.go` 和 `scripts/check.sh` 改动；不出现 `.cache`。
