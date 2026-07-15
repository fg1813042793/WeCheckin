# 前端 API 地址环境化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 uni-app 客户端配置中的局域网 API 地址硬编码，让 H5/App/小程序构建都通过环境变量配置后端地址。

**Architecture:** 保留现有 `CONFIG.BASE_URL` 调用方式，只修改 `frontend/config/index.js` 的来源：优先读取 `import.meta.env.VITE_API_BASE_URL`，未配置时回退到 `http://localhost:8083`。新增 `frontend/.env.example` 作为中文配置示例，并新增一个无第三方依赖的 Node 检查脚本，防止局域网 IP 和注释掉的备用地址再次进入配置文件。

**Tech Stack:** uni-app、Vite env、Node.js 静态检查、Markdown。

---

## File Structure

- Modify: `frontend/config/index.js`
  - `BASE_URL` 改为读取 `import.meta.env.VITE_API_BASE_URL`。
  - 删除局域网地址和注释掉的备用地址。
- Create: `frontend/.env.example`
  - 提供 `VITE_API_BASE_URL=http://localhost:8083` 示例。
- Create: `frontend/scripts/check-config.mjs`
  - 检查前端配置不得包含 `192.168.*` 局域网地址。
  - 检查配置文件必须读取 `VITE_API_BASE_URL`。
  - 检查 `.env.example` 包含示例配置。
- Modify: `frontend/package.json`
  - 增加 `check:config` 脚本。
- Modify: `scripts/check.sh`
  - 将前端配置检查接入项目级检查。
- Modify: `README.md`
  - 补充前端 API 地址环境变量说明。

---

### Task 1: 新增前端配置防回归检查

**Files:**
- Create: `frontend/scripts/check-config.mjs`
- Modify: `frontend/package.json`

- [ ] **Step 1: 写入失败检查脚本**

Create `frontend/scripts/check-config.mjs` with:

```js
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const configPath = resolve(root, 'config/index.js')
const envExamplePath = resolve(root, '.env.example')

const configSource = readFileSync(configPath, 'utf8')
if (!existsSync(envExamplePath)) {
  throw new Error('frontend .env.example must exist')
}
const envExample = readFileSync(envExamplePath, 'utf8')

if (/192\.168\.\d+\.\d+/.test(configSource)) {
  throw new Error('frontend config must not hardcode LAN API addresses')
}

if (!configSource.includes('VITE_API_BASE_URL')) {
  throw new Error('frontend config must read VITE_API_BASE_URL')
}

if (!envExample.includes('VITE_API_BASE_URL=http://localhost:8083')) {
  throw new Error('frontend .env.example must document VITE_API_BASE_URL')
}
```

- [ ] **Step 2: 在 package.json 增加检查脚本**

Add to `frontend/package.json` scripts:

```json
"check:config": "node scripts/check-config.mjs"
```

- [ ] **Step 3: 运行检查确认失败**

Run:

```bash
npm --prefix frontend run check:config
```

Expected: FAIL。当前 `.env.example` 不存在，或 `config/index.js` 仍包含局域网地址 / 未读取 `VITE_API_BASE_URL`。

---

### Task 2: 实现 API 地址环境化

**Files:**
- Modify: `frontend/config/index.js`
- Create: `frontend/.env.example`
- Modify: `README.md`

- [ ] **Step 1: 更新配置入口**

Change `frontend/config/index.js` to:

```js
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083'

export default {
  BASE_URL: API_BASE_URL,
  VER: 'build 2026.05.28',
  COMPANY: 'MY打卡',

  IS_DEMO: false,
  MOBILE_CHECK: false,

  IMG_UPLOAD_SIZE: 20,

  CACHE_IS_LIST: true,
  CACHE_LIST_TIME: 60 * 30
}
```

- [ ] **Step 2: 新增环境变量示例**

Create `frontend/.env.example` with:

```bash
# uni-app 客户端后端 API 地址
# 本地开发通常使用 http://localhost:8083
# 真机/小程序调试请改成设备可访问的局域网或测试环境地址
VITE_API_BASE_URL=http://localhost:8083
```

- [ ] **Step 3: 更新 README 配置说明**

Add to README “配置说明” section:

```markdown
uni-app 客户端默认读取 `frontend/.env` 中的 `VITE_API_BASE_URL` 作为后端 API 地址。可以复制 `frontend/.env.example` 为 `frontend/.env` 后按环境修改：

```bash
cd frontend
cp .env.example .env
```

本地 H5 调试可使用 `http://localhost:8083`；真机或小程序调试需要填写设备可访问的局域网、测试环境或生产环境地址。
```

- [ ] **Step 4: 运行检查确认通过**

Run:

```bash
npm --prefix frontend run check:config
```

Expected: PASS。

---

### Task 3: 接入项目级检查并验证

**Files:**
- Modify: `scripts/check.sh`

- [ ] **Step 1: 将前端配置检查加入项目级脚本**

Add after backend checks in `scripts/check.sh`:

```bash
echo "==> Running frontend config checks"
npm --prefix frontend run check:config
```

- [ ] **Step 2: 运行项目级检查**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS，输出包含 `Running frontend config checks` 和 `frontend config` 检查通过信息。

- [ ] **Step 3: 确认缓存目录已清理**

Run:

```bash
test ! -e .cache
```

Expected: PASS，无输出，退出码为 0。

- [ ] **Step 4: 搜索前端局域网硬编码**

Run:

```bash
rg -n '192\.168\.[0-9]+\.[0-9]+' frontend/config frontend/.env.example README.md
```

Expected: 无输出。
