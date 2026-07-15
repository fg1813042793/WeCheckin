# Token 前缀统一 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 消除客户端公开 handler 中硬编码的 Redis token key 前缀，让 token 读取统一走 `tokenutil` 配置入口。

**Architecture:** 在 `backend/pkg/tokenutil` 增加 `TokenAuthKey` 和 `TokenSetKey` helper，由它们统一组合 `GetTokenConfig(role)` 返回的 Redis 前缀和 `a:`/`s:` key。先用单元测试证明 helper 会使用配置前缀，再用 handler 源码安全测试禁止 `user_token:a:`、`admin_token:a:` 等硬编码写法回归，最后替换 `survey.go` 和 `exam.go` 中绕过配置的写法。

**Tech Stack:** Go test、Go 标准库、现有 `tokenutil`、Bash 检查脚本。

---

## File Structure

- Modify: `backend/pkg/tokenutil/tokenutil.go`
  - 新增 `TokenAuthKey(role, token string)`。
  - 新增 `TokenSetKey(role, id string)`。
- Create: `backend/pkg/tokenutil/tokenutil_test.go`
  - 验证 helper 使用配置中的 Redis 前缀。
- Create: `backend/internal/app/handler/token_prefix_safety_test.go`
  - 防止 handler 中重新出现硬编码 token Redis key。
- Modify: `backend/internal/app/handler/survey.go`
  - 将 `user_token:a:` 替换为 `tokenutil.TokenAuthKey("user", token)`。
- Modify: `backend/internal/app/handler/exam.go`
  - 引入 `tokenutil`，将 `user_token:a:` 替换为 `tokenutil.TokenAuthKey("user", token)`。
- Modify: `scripts/check.sh`
  - 将 `./backend/pkg/tokenutil` 和 `./backend/internal/app/handler` 纳入项目级检查。

---

### Task 1: 为 tokenutil 增加 key helper 测试

**Files:**
- Create: `backend/pkg/tokenutil/tokenutil_test.go`
- Modify: `backend/pkg/tokenutil/tokenutil.go`

- [ ] **Step 1: 写入失败测试**

Create `backend/pkg/tokenutil/tokenutil_test.go` with:

```go
package tokenutil

import (
	"testing"

	"wecheckin-backend/backend/internal/config"
)

func TestTokenRedisKeysUseConfiguredPrefix(t *testing.T) {
	oldCfg := config.Cfg
	t.Cleanup(func() { config.Cfg = oldCfg })

	config.Cfg = &config.Config{
		Token: config.TokenConfig{
			User:  config.TokenRoleConfig{Expire: "24h", RedisPrefix: "custom_user_token:"},
			Admin: config.TokenRoleConfig{Expire: "24h", RedisPrefix: "custom_admin_token:"},
		},
	}

	if got := TokenAuthKey("user", "abc"); got != "custom_user_token:a:abc" {
		t.Fatalf("user auth key = %q", got)
	}
	if got := TokenSetKey("admin", "42"); got != "custom_admin_token:s:42" {
		t.Fatalf("admin set key = %q", got)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/pkg/tokenutil -run TestTokenRedisKeysUseConfiguredPrefix -count=1
```

Expected: FAIL，输出包含 `undefined: TokenAuthKey` 或 `undefined: TokenSetKey`。

- [ ] **Step 3: 实现 helper**

Add to `backend/pkg/tokenutil/tokenutil.go` after `GetTokenConfig`:

```go
func TokenAuthKey(role, token string) string {
	_, prefix := GetTokenConfig(role)
	return prefix + "a:" + token
}

func TokenSetKey(role, id string) string {
	_, prefix := GetTokenConfig(role)
	return prefix + "s:" + id
}
```

- [ ] **Step 4: 运行测试确认通过**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/pkg/tokenutil -run TestTokenRedisKeysUseConfiguredPrefix -count=1
```

Expected: PASS。

---

### Task 2: 禁止 handler 硬编码 token Redis key

**Files:**
- Create: `backend/internal/app/handler/token_prefix_safety_test.go`
- Modify: `backend/internal/app/handler/survey.go`
- Modify: `backend/internal/app/handler/exam.go`

- [ ] **Step 1: 写入失败测试**

Create `backend/internal/app/handler/token_prefix_safety_test.go` with:

```go
package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlersDoNotHardcodeTokenRedisKeys(t *testing.T) {
	forbidden := []string{
		`"user_token:a:"`,
		`"admin_token:a:"`,
		`"user_token:s:"`,
		`"admin_token:s:"`,
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob handler files: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s must not hardcode token redis key %s", file, needle)
			}
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/handler -run TestHandlersDoNotHardcodeTokenRedisKeys -count=1
```

Expected: FAIL，输出指向 `survey.go` 或 `exam.go` 中的硬编码 token key。

- [ ] **Step 3: 替换 survey.go 硬编码**

Change:

```go
rdKey := "user_token:a:" + token
```

to:

```go
rdKey := tokenutil.TokenAuthKey("user", token)
```

- [ ] **Step 4: 替换 exam.go 硬编码并导入 tokenutil**

Add import:

```go
"wecheckin-backend/backend/pkg/tokenutil"
```

Change each:

```go
rdKey := "user_token:a:" + token
rdKey := "user_token:a:" + auth
```

to:

```go
rdKey := tokenutil.TokenAuthKey("user", token)
rdKey := tokenutil.TokenAuthKey("user", auth)
```

- [ ] **Step 5: gofmt 并运行测试确认通过**

Run:

```bash
gofmt -w backend/pkg/tokenutil/tokenutil.go backend/pkg/tokenutil/tokenutil_test.go backend/internal/app/handler/token_prefix_safety_test.go backend/internal/app/handler/survey.go backend/internal/app/handler/exam.go
GOCACHE=$PWD/.cache/go-build go test ./backend/pkg/tokenutil ./backend/internal/app/handler -run 'TestTokenRedisKeysUseConfiguredPrefix|TestHandlersDoNotHardcodeTokenRedisKeys' -count=1
```

Expected: PASS。

---

### Task 3: 扩展项目级检查并最终验证

**Files:**
- Modify: `scripts/check.sh`

- [ ] **Step 1: 更新检查脚本包列表**

Change `scripts/check.sh` go test package list to include:

```bash
  ./backend/pkg/tokenutil \
  ./backend/internal/app/handler \
```

- [ ] **Step 2: 运行项目级检查**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS，输出包含 `backend/pkg/tokenutil`、`backend/internal/app/handler` 和 `==> Checks passed`。

- [ ] **Step 3: 确认缓存目录已清理**

Run:

```bash
test ! -e .cache
```

Expected: PASS，无输出，退出码为 0。

- [ ] **Step 4: 搜索硬编码 token key**

Run:

```bash
rg -n '"user_token:a:"|"admin_token:a:"|"user_token:s:"|"admin_token:s:"' backend/internal/app/handler backend/pkg/tokenutil
```

Expected: 只允许测试文件中出现禁止字符串；生产 handler 不应出现。
