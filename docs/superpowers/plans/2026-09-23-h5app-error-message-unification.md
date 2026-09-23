# H5App Error Message Unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 统一 H5App 错误提示，优先显示具体业务错误信息并减少固定兜底文案。

**Architecture:** 新增 `src/common/app-error.ts` 作为纯错误解析与展示边界，先覆盖请求错误对象和嵌套响应，再由重点业务页面替换 catch 中的固定提示。请求拦截器继续负责接口级 toast，页面只处理可恢复错误和场景状态。

**Tech Stack:** uni-app、Vue 3、TypeScript、uView Pro、Node 检查脚本。

---

### Task 1: Build and test the shared error utility

**Files:**
- Create: `h5app/src/common/app-error.ts`
- Create: `h5app/scripts/check-app-error.mjs`
- Modify: `h5app/package.json`

- [ ] **Step 1: Write the failing check**

Cover direct `msg`, direct `message`, nested `data`, nested `response.data`, `errMsg`, fallback, blank values, and sensitive stack/path omission. Require `showAppError` to call `uni.showToast` with the resolved message.

- [ ] **Step 2: Run RED**

```bash
cd h5app
pnpm check:app-error
```

Expected: FAIL because the utility does not exist.

- [ ] **Step 3: Implement the utility**

Export:

```ts
export function appErrorMessage(error: unknown, fallback: string): string
export function showAppError(error: unknown, fallback: string): string
```

Use `unknown` narrowing, preserve Chinese backend messages, trim whitespace, and fall back only when no usable message exists.

- [ ] **Step 4: Add package script and run GREEN**

```bash
pnpm check:app-error
pnpm exec eslint src/common/app-error.ts scripts/check-app-error.mjs
```

- [ ] **Step 5: Commit**

```bash
git add h5app/src/common/app-error.ts h5app/scripts/check-app-error.mjs h5app/package.json
git commit -m "feat: 增加 H5App 统一错误提示工具"
```

### Task 2: Integrate request/upload error extraction

**Files:**
- Modify: `h5app/src/common/http.interceptor.ts`
- Modify: `h5app/src/api/dingtalk-h5/base.ts`
- Modify: `h5app/scripts/check-app-error.mjs`

- [ ] **Step 1: Add structure assertions**

Require both request boundaries to import/use `appErrorMessage`; preserve auth-expired handling and upload-specific fallback.

- [ ] **Step 2: Replace unknown request fallback text**

For non-2xx and request-fail paths, use `appErrorMessage(rawResponse, '网络请求失败，请稍后重试')`. For upload failures use response message first, with `上传失败` / `上传响应异常` only as fallback.

- [ ] **Step 3: Run focused checks**

```bash
pnpm check:app-error
pnpm exec eslint src/common/http.interceptor.ts src/api/dingtalk-h5/base.ts scripts/check-app-error.mjs
```

### Task 3: Integrate H5App business pages

**Files:**
- Modify: workflow, performance, feedback, notification and app-shell pages that catch request errors.
- Modify: `h5app/scripts/check-app-error.mjs`

- [ ] **Step 1: Add page wiring assertions**

Require `showAppError` in the workflow submit/detail/revision pages, feedback/notification list actions, performance loading pages, attachment/image operations, and profile actions where an error object is available.

- [ ] **Step 2: Replace fixed catch text**

Use `showAppError(error, '场景失败兜底')` in catch blocks. Keep local validation messages and device capability failures unchanged.

- [ ] **Step 3: Update list error states**

Where pages maintain `errorMessage`, assign `appErrorMessage(error, '场景加载失败')` instead of a fixed string.

- [ ] **Step 4: Run focused checks**

```bash
pnpm check:app-error
pnpm exec eslint src/common/app-error.ts src/common/http.interceptor.ts src/api/dingtalk-h5/base.ts scripts/check-app-error.mjs
pnpm type-check
```

### Task 4: Verify the full H5App scope

- [ ] **Step 1:** `pnpm check:app-error`
- [ ] **Step 2:** `pnpm lint`
- [ ] **Step 3:** `pnpm type-check`
- [ ] **Step 4:** `pnpm build:h5`
- [ ] **Step 5:** `git diff --check` and review unrelated dirty files.

Report existing baseline lint issues separately and leave user changes untouched.
