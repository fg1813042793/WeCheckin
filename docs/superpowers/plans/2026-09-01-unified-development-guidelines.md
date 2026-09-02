# WeCheckin Unified Development Guidelines Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立根级总则与 backend、admin、frontend、h5app 四端专项规范组成的分层开发规范体系。

**Architecture:** 根 `AGENTS.md` 为 AI 提供简短的强制规则和文档导航，根 `docs/development-guidelines.md` 管理跨端契约和共同工作流。每个子项目保留一份只描述自身技术栈和验证方式的专项规范，现有专题文档通过链接复用。

**Tech Stack:** Markdown、Go 1.24/Hertz/GORM、Vue 3/Vite/TypeScript/Element Plus、uni-app/Vue 3、uni-app/TypeScript/uView Pro。

---

### Task 1: 根级执行规则和统一规范

**Files:**
- Create: `AGENTS.md`
- Create: `docs/development-guidelines.md`

- [x] **Step 1: 创建根 `AGENTS.md`**

写入以下可执行内容：项目事实、四端职责、对应专项规范路径、规范优先级、改动范围、API v2/权限/Swagger/迁移联动要求、生成物与敏感信息规则、按端验证命令。根规则必须明确要求修改模块前阅读对应的 `docs/development-guidelines.md`。

- [x] **Step 2: 创建团队统一规范**

`docs/development-guidelines.md` 使用以下章节：

1. 目的与适用范围。
2. 项目结构与四端边界。
3. 规范层级和专项文档导航。
4. 通用开发原则。
5. API v2、DTO、权限和 Swagger 跨端契约。
6. 数据库、配置、安全、日志和兼容性。
7. UI 与跨端用户体验原则。
8. 测试与验证矩阵。
9. 文档、生成物和提交约定。
10. 功能交付检查清单。

- [x] **Step 3: 检查根级文档导航**

Run:

```bash
rg -n "backend/docs/development-guidelines.md|admin/docs/development-guidelines.md|frontend/docs/development-guidelines.md|h5app/docs/development-guidelines.md" AGENTS.md docs/development-guidelines.md
```

Expected: 四个专项规范路径在根级文档中均可找到。

### Task 2: Backend 专项规范

**Files:**
- Create: `backend/docs/development-guidelines.md`

- [x] **Step 1: 编写后端规范**

文档必须覆盖：Go 格式与命名；`cmd`、`routes`、`handler`、`service`、`modules`、`model`、`support`、`pkg`、`test` 边界；DTO/Context/事务规则；API v2 和权限目录；Swagger 同步；SQL 迁移；配置和敏感值；测试布局；以 `gofmt` 和 `go test` 为核心的验证命令。

链接引用：

- `../../docs/backend-dto-context-guidelines.md`
- `../../docs/API_V2.md`
- `../../docs/PERMISSION_CODE_FRONTEND_SYNC.md`
- `../internal/service/README.md`
- `../test/README.md`
- `../../docs/project-maintenance.md`

- [x] **Step 2: 核对后端命令**

Run:

```bash
test -f backend/go.mod
test -f backend/init.sh
test -f backend/internal/routes/v2/swagger/swagger.go
```

Expected: 全部命令退出码为 0。

### Task 3: Admin 专项规范

**Files:**
- Create: `admin/docs/development-guidelines.md`

- [x] **Step 1: 编写管理后台规范**

文档必须覆盖：Vue 3 + Vite + strict TypeScript + Element Plus 基线；`api`、`router`、`views`、`components`、`utils`、`constants`、`styles` 边界；SFC 和类型约定；请求封装和 API v2 admin 路径；菜单/路由/按钮/接口权限联动；Element Plus 与页面状态；大组件拆分；性能与 bundle 预算；`npm run check:all` 和 `npm run build`。

文档必须明确当前未配置 ESLint/Prettier，不得把它们写为已有门禁。

- [x] **Step 2: 核对 Admin 脚本**

Run:

```bash
node -e "const s=require('./admin/package.json').scripts; for (const k of ['build','check:all','check:request']) if (!s[k]) process.exit(1)"
```

Expected: 退出码为 0。

### Task 4: Frontend 专项规范

**Files:**
- Create: `frontend/docs/development-guidelines.md`

- [x] **Step 1: 编写客户端规范**

文档必须覆盖：当前根目录式 uni-app/Vue 3 基线；`pages`、`components`、`api`、`config`、`utils`、`mixins`、`static`、`uni_modules` 边界；Vue 页面与组件复用；请求封装和 API v2；登录态与本地存储；uni-app 条件编译；H5/App/微信小程序兼容；页面状态与移动端布局；`npm run check:all` 和目标平台构建。

文档必须明确当前没有 lint 或类型检查脚本，不得伪造已有门禁。

- [x] **Step 2: 核对 Frontend 脚本**

Run:

```bash
node -e "const s=require('./frontend/package.json').scripts; for (const k of ['check:all','dev:h5','build:h5','build:app','build:mp-weixin']) if (!s[k]) process.exit(1)"
```

Expected: 退出码为 0。

### Task 5: H5App 专项规范校准

**Files:**
- Modify: `h5app/docs/development-guidelines.md`

- [x] **Step 1: 保留现有 H5App 规范并补充项目契约**

在文档开头补充根规范链接和 `h5app` 权威源说明，新增“WeCheckin 集成契约”章节，覆盖 `/api/v2/dingtalk/h5`、bootstrap 菜单数据、按钮/API 权限 key、后端权限声明和迁移同步。保留原有 pnpm、strict TypeScript、ESLint、uView Pro、主题、国际化和验证规则。

- [x] **Step 2: 核对 H5App 脚本与基线**

文档中明确列出 `pnpm lint`、`pnpm type-check` 和按影响范围执行的 `pnpm build:h5`。

Run:

```bash
node -e "const p=require('./h5app/package.json'); const s=p.scripts; if (p.packageManager !== 'pnpm@10.34.5' || !s.lint || !s['type-check'] || !s['build:h5']) process.exit(1)"
```

Expected: 退出码为 0。

### Task 6: 文档一致性验证

**Files:**
- Verify: `AGENTS.md`
- Verify: `docs/development-guidelines.md`
- Verify: `backend/docs/development-guidelines.md`
- Verify: `admin/docs/development-guidelines.md`
- Verify: `frontend/docs/development-guidelines.md`
- Verify: `h5app/docs/development-guidelines.md`

- [x] **Step 1: 检查目标文件和关键命令**

Run:

```bash
test -f AGENTS.md
test -f docs/development-guidelines.md
test -f backend/docs/development-guidelines.md
test -f admin/docs/development-guidelines.md
test -f frontend/docs/development-guidelines.md
test -f h5app/docs/development-guidelines.md
node -e "for (const f of ['admin/package.json','frontend/package.json','h5app/package.json']) JSON.parse(require('fs').readFileSync(f,'utf8'))"
```

Expected: 全部命令退出码为 0。

- [x] **Step 2: 检查占位词与明显冲突**

Run:

```bash
rg -n "TBD|TODO|FIXME|待定|后续决定" AGENTS.md docs/development-guidelines.md backend/docs/development-guidelines.md admin/docs/development-guidelines.md frontend/docs/development-guidelines.md h5app/docs/development-guidelines.md
```

Expected: 无输出。

- [x] **Step 3: 检查 Markdown 差异质量**

Run:

```bash
git diff --check -- AGENTS.md docs/development-guidelines.md backend/docs/development-guidelines.md admin/docs/development-guidelines.md frontend/docs/development-guidelines.md h5app/docs/development-guidelines.md
```

Expected: 无输出，退出码为 0。

- [x] **Step 4: 审阅变更范围**

Run:

```bash
git status --short -- AGENTS.md docs/development-guidelines.md backend/docs/development-guidelines.md admin/docs/development-guidelines.md frontend/docs/development-guidelines.md h5app/docs/development-guidelines.md docs/superpowers/specs/2026-09-01-unified-development-guidelines-design.md docs/superpowers/plans/2026-09-01-unified-development-guidelines.md
```

Expected: 只列出本次规范、设计和计划文档；不包含业务代码或配置文件。
