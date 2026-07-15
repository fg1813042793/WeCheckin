# 移动端请求层增强 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化 uni-app 客户端请求封装，统一登录失效处理、请求方法规范、JSON 请求传参和错误返回行为。

**Architecture:** 保留 `frontend/utils/request.js` 的 `request/get/post/postJSON/put/del` 导出不变，在内部新增登录失效消息集合、认证状态清理、登录页跳转保护和统一错误提示。登录失效时不再让 Promise 悬空，而是清理本端登录态、跳转登录页并 reject 原始错误。新增无依赖 Node 静态检查，接入 `frontend/package.json` 和根检查脚本。

**Tech Stack:** uni-app、JavaScript、Node.js 静态检查。

---

## File Structure

- Modify: `frontend/utils/request.js`
  - 新增 `LOGIN_EXPIRED_MESSAGES`。
  - 新增 `getAuthState()`、`clearAuthState()`、`redirectToLogin()`。
  - 统一 method 大写，增加默认 timeout。
  - 登录失效后 reject，避免 Promise 悬空。
  - `postJSON` 直接传对象，不再提前 `JSON.stringify`。
- Create: `frontend/scripts/check-request.mjs`
  - 静态检查请求层是否包含关键结构。
- Modify: `frontend/package.json`
  - 增加 `check:request`。
- Modify: `scripts/check.sh`
  - 接入移动端请求层检查。

---

### Task 1: 新增移动端请求层静态检查

**Files:**
- Create: `frontend/scripts/check-request.mjs`
- Modify: `frontend/package.json`

- [ ] **Step 1: 创建检查脚本**

Create `frontend/scripts/check-request.mjs` with:

```js
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../utils/request.js'), 'utf8')

const requiredSnippets = [
  'const LOGIN_EXPIRED_MESSAGES = new Set',
  'function getAuthState',
  'function clearAuthState',
  'function redirectToLogin',
  'LOGIN_EXPIRED_MESSAGES.has',
  'reject(res.data)',
  "method: (options.method || 'GET').toUpperCase()",
  'timeout: options.timeout || 15000',
  'data,',
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`frontend request layer missing: ${snippet}`)
  }
}

if (source.includes('data: JSON.stringify(data)')) {
  throw new Error('postJSON should pass object data to uni.request instead of pre-stringifying')
}
```

- [ ] **Step 2: 增加 npm 脚本**

Add to `frontend/package.json` scripts:

```json
"check:request": "node scripts/check-request.mjs"
```

- [ ] **Step 3: 运行检查确认失败**

Run:

```bash
npm --prefix frontend run check:request
```

Expected: FAIL，输出包含 `frontend request layer missing`。

---

### Task 2: 优化 request.js

**Files:**
- Modify: `frontend/utils/request.js`

- [ ] **Step 1: 更新请求层实现**

Implement these helpers in `frontend/utils/request.js`:

```js
const LOGIN_EXPIRED_MESSAGES = new Set([
  '未登录',
  '登录已过期',
  '登录已过期或已被强制下线',
  '账号异常'
])

let redirectingToLogin = false

function getAuthState(isAdmin) {
  return {
    token: isAdmin ? uni.getStorageSync('admin_token') : uni.getStorageSync('token'),
    tokenKey: isAdmin ? 'admin_token' : 'token',
    infoKey: isAdmin ? 'admin_info' : 'userInfo',
    loginUrl: isAdmin ? '/pages/admin/admin_login' : '/pages/login/login'
  }
}

function clearAuthState(authState) {
  uni.removeStorageSync(authState.tokenKey)
  uni.removeStorageSync(authState.infoKey)
}

function redirectToLogin(authState) {
  if (redirectingToLogin) return
  redirectingToLogin = true
  uni.redirectTo({
    url: authState.loginUrl,
    complete: () => {
      redirectingToLogin = false
    }
  })
}
```

Use `LOGIN_EXPIRED_MESSAGES.has(res.data.msg)` in the response branch. When login expires, call `clearAuthState(authState)`, `redirectToLogin(authState)` and `reject(res.data)`.

- [ ] **Step 2: 更新 postJSON**

Change `postJSON` to:

```js
const postJSON = (url, data = {}) => {
  return request({
    url,
    method: 'POST',
    data,
    header: { 'Content-Type': 'application/json' }
  })
}
```

- [ ] **Step 3: 运行检查确认通过**

Run:

```bash
npm --prefix frontend run check:request
```

Expected: PASS。

---

### Task 3: 接入项目级检查

**Files:**
- Modify: `scripts/check.sh`

- [ ] **Step 1: 增加移动端请求检查**

Add after frontend config checks:

```bash
npm --prefix frontend run check:request
```

- [ ] **Step 2: 运行项目级检查**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS。
