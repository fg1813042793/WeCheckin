# 验证脚本收口优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增项目级健康检查脚本，并在 README 中用中文说明如何一条命令验证当前关键后端回归测试。

**Architecture:** 在仓库根目录新增 `scripts/check.sh`，脚本自行定位仓库根目录，使用项目内 `.cache/go-build` 作为 Go 构建缓存，运行当前稳定的后端测试集合，并在退出时清理 `.cache`。README 的“测试”段落改为推荐使用该脚本，同时保留手动 Go 测试命令作为说明。

**Tech Stack:** Bash、Go test、Markdown。

---

## File Structure

- Create: `scripts/check.sh`
  - 项目级健康检查入口。
  - 从脚本路径定位仓库根目录，避免调用者当前目录影响测试路径。
  - 使用 `trap` 在脚本退出时清理 `.cache`。
- Modify: `README.md`
  - 更新“测试”章节为中文验证说明。
  - 推荐执行 `bash scripts/check.sh`。
  - 明确当前覆盖范围：启动安全、配置加载、formkit 相关测试。

---

### Task 1: 新增项目级检查脚本

**Files:**
- Create: `scripts/check.sh`

- [ ] **Step 1: 先运行缺失脚本，确认当前没有统一入口**

Run:

```bash
bash scripts/check.sh
```

Expected: FAIL，输出包含 `No such file or directory` 或等价的脚本不存在错误。

- [ ] **Step 2: 创建脚本文件**

Create `scripts/check.sh` with:

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
GOCACHE_DIR="${ROOT_DIR}/.cache/go-build"

cleanup() {
  rm -rf "${ROOT_DIR}/.cache"
}

trap cleanup EXIT

cd "${ROOT_DIR}"

echo "==> Running backend checks"
GOCACHE="${GOCACHE_DIR}" go test \
  ./backend/internal/app/service \
  ./backend/internal/config \
  ./backend/internal/app/formkit/...
echo "==> Checks passed"
```

- [ ] **Step 3: 运行脚本，确认测试通过且缓存被清理**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS，输出包含 `==> Checks passed`，并且 Go 测试包返回 `ok`。

- [ ] **Step 4: 检查脚本运行后没有留下 `.cache`**

Run:

```bash
test ! -e .cache
```

Expected: PASS，无输出，退出码为 0。

---

### Task 2: 更新 README 测试说明

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 修改 README 的“测试”章节**

Replace the current “测试” section with:

````markdown
## 测试

推荐使用项目级检查脚本验证当前后端关键回归测试：

```bash
bash scripts/check.sh
```

该脚本会使用项目内 `.cache/go-build` 作为 Go 构建缓存，并在结束时自动清理 `.cache/`。当前覆盖范围包括：

- 启动初始化安全检查：`./backend/internal/app/service`
- 配置加载和环境变量覆盖：`./backend/internal/config`
- formkit 子系统：`./backend/internal/app/formkit/...`

如需单独运行 formkit 测试，也可以执行：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/formkit/...
```
````

- [ ] **Step 2: 确认 README 中存在推荐命令和覆盖范围**

Run:

```bash
rg -n "bash scripts/check.sh|启动初始化安全检查|配置加载和环境变量覆盖|formkit 子系统" README.md
```

Expected: PASS，输出匹配上述中文说明。

---

### Task 3: 最终验证

**Files:**
- Test: `scripts/check.sh`
- Test: `README.md`

- [ ] **Step 1: 运行项目级检查脚本**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS，输出包含 `==> Checks passed`。

- [ ] **Step 2: 确认缓存目录已清理**

Run:

```bash
test ! -e .cache
```

Expected: PASS，无输出，退出码为 0。

- [ ] **Step 3: 检查工作区状态**

Run:

```bash
git status --short
```

Expected: 只包含本轮新增的 `scripts/check.sh`、README 改动，以及本轮开始前已经存在的未提交优化文件；不应出现 `.cache`。

- [ ] **Step 4: 保持实现改动未提交，等待用户决定是否统一提交**

Run:

```bash
git diff -- README.md
sed -n '1,120p' scripts/check.sh
```

Expected: README diff 仅展示测试说明相关改动；脚本内容与 Task 1 Step 2 中的脚本一致。
