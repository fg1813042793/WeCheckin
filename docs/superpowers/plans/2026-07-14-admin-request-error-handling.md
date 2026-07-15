# 管理后台请求层统一错误处理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化管理后台 Axios 请求层，统一登录失效处理、权限缓存清理、FormData 请求头和网络错误提示。

**Architecture:** 保持现有 `admin/src/utils/request.ts` 对外默认导出不变，只在请求层内部增加登录失效消息集合、session 清理函数、登录页跳转保护和更稳妥的表单编码逻辑。新增一个无依赖 Node 静态检查脚本，确保请求层持续导入 `clearPerms`、集中维护登录失效消息、处理 FormData 请求头，并接入 `admin/package.json`。

**Tech Stack:** Vue 3 管理后台、Axios、Element Plus、Node.js 静态检查、TypeScript build。

---

## File Structure

- Modify: `admin/src/utils/request.ts`
  - 导入 `clearPerms`。
  - 使用 `LOGIN_EXPIRED_MESSAGES` 统一登录失效判断。
  - 新增 `clearAdminSession()`，清理 token、admin_info 和权限缓存。
  - 新增 `redirectToLogin()`，避免重复跳转。
  - FormData 请求删除默认 `Content-Type`，交给浏览器设置 multipart boundary。
  - `transformRequest` 兼容空 data。
- Create: `admin/scripts/check-request.mjs`
  - 静态检查请求层是否包含上述关键结构。
- Modify: `admin/package.json`
  - 增加 `check:request` 脚本。

---

### Task 1: 新增请求层静态检查

**Files:**
- Create: `admin/scripts/check-request.mjs`
- Modify: `admin/package.json`

- [ ] **Step 1: 创建检查脚本**

Create `admin/scripts/check-request.mjs` with:

```js
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const source = readFileSync(resolve('admin/src/utils/request.ts'), 'utf8')

const requiredSnippets = [
  "import { clearPerms } from './permission'",
  'const LOGIN_EXPIRED_MESSAGES = new Set',
  'function clearAdminSession()',
  'clearPerms()',
  'function redirectToLogin()',
  'function encodeFormBody',
  'data instanceof FormData',
  'headers.delete',
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`admin request layer missing: ${snippet}`)
  }
}
```

- [ ] **Step 2: 增加 npm 脚本**

Add to `admin/package.json` scripts:

```json
"check:request": "node scripts/check-request.mjs"
```

- [ ] **Step 3: 运行检查确认失败**

Run:

```bash
npm --prefix admin run check:request
```

Expected: FAIL，输出包含 `admin request layer missing`。

---

### Task 2: 优化 request.ts

**Files:**
- Modify: `admin/src/utils/request.ts`

- [ ] **Step 1: 替换请求层实现**

Update `admin/src/utils/request.ts` so it keeps the same default export and includes:

```ts
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { clearPerms } from './permission'

const LOGIN_EXPIRED_MESSAGES = new Set([
  '未登录',
  '登录已过期',
  '登录已过期或已被强制下线',
  '账号异常'
])

let redirectingToLogin = false

function clearAdminSession() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_info')
  clearPerms()
}

function redirectToLogin() {
  if (redirectingToLogin || window.location.pathname === '/login') return
  redirectingToLogin = true
  window.location.href = '/login'
}

function encodeFormBody(data: any) {
  const params = new URLSearchParams()
  if (!data) return params.toString()
  for (const key in data) {
    if (data[key] !== undefined && data[key] !== null) {
      params.append(key, String(data[key]))
    }
  }
  return params.toString()
}

const request = axios.create({
  baseURL: '',
  timeout: 15000,
  transformRequest: [(data: any, headers: any) => {
    if (data instanceof FormData) {
      if (headers && typeof headers.delete === 'function') {
        headers.delete('Content-Type')
      } else if (headers) {
        delete headers['Content-Type']
        delete headers['content-type']
      }
      return data
    }
    return encodeFormBody(data)
  }],
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
})
```

Keep request and response interceptors, but use `clearAdminSession()` and `redirectToLogin()` in the login-expired branch.

- [ ] **Step 2: 运行检查确认通过**

Run:

```bash
npm --prefix admin run check:request
```

Expected: PASS。

- [ ] **Step 3: 运行管理后台构建**

Run:

```bash
npm --prefix admin run build
```

Expected: PASS。

---

### Task 3: 接入项目级检查

**Files:**
- Modify: `scripts/check.sh`

- [ ] **Step 1: 添加 admin 检查**

Add after frontend config checks:

```bash
echo "==> Running admin request checks"
npm --prefix admin run check:request
```

- [ ] **Step 2: 运行项目级检查**

Run:

```bash
bash scripts/check.sh
```

Expected: PASS，输出包含 `Running admin request checks` 和 `==> Checks passed`。
