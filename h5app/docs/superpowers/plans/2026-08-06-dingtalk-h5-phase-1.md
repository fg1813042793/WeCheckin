# 钉钉 H5 第一阶段迁移 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前 uView Pro Starter 中新增钉钉 H5 第一阶段核心闭环，复用 `/api/v2/dingtalk/h5` 接口且不修改旧 `dingtalk-h5` 项目。

**Architecture:** 新增独立钉钉模块，接口、类型、认证状态、绩效状态和页面组件分层。HTTP 仍走当前项目 `uview-pro` 插件，只扩展请求拦截器支持钉钉 token、业务响应和错误分支。

**Tech Stack:** uni-app, Vue 3, TypeScript, Pinia, uView Pro, SCSS, Node.js check script.

---

### Task 1: 结构检查脚本

**Files:**
- Create: `scripts/check-dingtalk-module.mjs`
- Modify: `package.json`

- [ ] **Step 1: 编写失败检查**

创建 `scripts/check-dingtalk-module.mjs`，检查 `src/pages/dingtalk/index.vue`、`src/api/dingtalk-h5/auth.ts`、`src/stores/dingtalkAuth.ts`、`src/stores/performance.ts`、`src/utils/dingtalk.ts` 存在，并检查 `src/pages.json` 包含 `pages/dingtalk/index`。

- [ ] **Step 2: 运行检查确认失败**

Run: `pnpm check:dingtalk-module`
Expected: FAIL，提示缺少钉钉模块文件或脚本。

- [ ] **Step 3: 在 `package.json` 增加脚本**

新增：

```json
{
  "check:dingtalk-module": "node scripts/check-dingtalk-module.mjs"
}
```

### Task 2: 类型、配置与 HTTP

**Files:**
- Create: `src/config/dingtalk-h5.ts`
- Create: `src/types/dingtalk-h5.ts`
- Modify: `src/common/http.interceptor.ts`

- [ ] **Step 1: 定义配置**

新增 `DINGTALK_H5_CONFIG`，包含 `BASE_URL`、`TOKEN_KEY`、`CLIENT_PLATFORM`、`DINGTALK_CORP_ID`。

- [ ] **Step 2: 定义共享类型**

新增认证响应、菜单、权限、用户、绩效单、模板和业务响应类型。

- [ ] **Step 3: 扩展拦截器**

支持 `VITE_API_BASE_URL`、`DT_H5_TOKEN`、`Authorization`、`X-Client-Platform`、`code=0` 成功、`code=10020` 绑定分支、登录过期清 token。

### Task 3: 接口迁移

**Files:**
- Create: `src/api/dingtalk-h5/base.ts`
- Create: `src/api/dingtalk-h5/auth.ts`
- Create: `src/api/dingtalk-h5/performance.ts`
- Create: `src/api/dingtalk-h5/profile.ts`
- Create: `src/api/dingtalk-h5/index.ts`

- [ ] **Step 1: 封装 API 常量**

定义 `API_V2` 和 `DINGTALK_H5_API`。

- [ ] **Step 2: 迁移认证接口**

实现 `publicConfig`、`login`、`ssoLogin`、`bindSelf`、`logout`。

- [ ] **Step 3: 迁移绩效接口**

实现 `bootstrap`、`workbench`、`listReviews`、`reviewDetail`、`reviewAction`、`deleteReview`、`exportReviewsUrl`、`createReview`、`listUsers`、`updateUser`、`deleteUser`、`getTemplate`、`saveTemplate`。

### Task 4: 状态与工具

**Files:**
- Create: `src/utils/dingtalk.ts`
- Create: `src/stores/dingtalkAuth.ts`
- Create: `src/stores/performance.ts`
- Modify: `src/stores/index.ts`

- [ ] **Step 1: 迁移钉钉工具**

实现 `isDingTalkRuntime`、`waitForDingTalkJSAPI`、`requestAuthCode`、`setNavigationTitle`，H5 DOM 相关代码用条件和运行时判断隔离。

- [ ] **Step 2: 实现认证 store**

实现公共配置、自动免登、账号登录、绑定账号、bootstrap、退出、权限判断。

- [ ] **Step 3: 实现绩效 store**

实现绩效单加载、待办计算、当前选中项、创建、删除和导出 URL。

### Task 5: 页面和组件

**Files:**
- Create: `src/pages/dingtalk/index.vue`
- Create: `src/pages/dingtalk/components/DingtalkLoginPanel.vue`
- Create: `src/pages/dingtalk/components/DingtalkBindPanel.vue`
- Create: `src/pages/dingtalk/components/DingtalkShell.vue`
- Create: `src/pages/dingtalk/components/PerformanceWorkbench.vue`
- Create: `src/pages/dingtalk/components/PerformanceList.vue`
- Create: `src/pages/dingtalk/components/PerformanceReviewDetail.vue`
- Modify: `src/pages.json`

- [ ] **Step 1: 新增入口路由**

在 `src/pages.json` 中加入 `pages/dingtalk/index`。

- [ ] **Step 2: 实现认证状态页**

入口页根据 ready、bind、denied、user 状态切换 loading、绑定、登录、应用壳。

- [ ] **Step 3: 实现核心应用壳**

使用 uView Pro tabs、button、cell、tag、modal 等组件渲染菜单、用户、刷新、退出。

- [ ] **Step 4: 实现工作台和绩效列表**

工作台展示待办，列表支持我的绩效、上级评审、HRBP 评审基础筛选。

- [ ] **Step 5: 实现详情基础动作**

详情展示员工、周期、状态、流程字段和可用动作按钮，动作调用 `reviewAction` 后刷新。

### Task 6: 验证

**Files:**
- Verify only

- [ ] **Step 1: 模块结构检查**

Run: `pnpm check:dingtalk-module`
Expected: PASS。

- [ ] **Step 2: Lint**

Run: `pnpm lint`
Expected: PASS，或只剩旧模板无关问题并记录。

- [ ] **Step 3: Type check**

Run: `pnpm type-check`
Expected: PASS，或只剩旧模板无关问题并记录。

- [ ] **Step 4: H5 build**

Run: `pnpm build:h5`
Expected: PASS。
