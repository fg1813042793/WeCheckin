# 性能优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可观测、可回归、可持续迭代的性能优化闭环，优先解决后台用户/管理员列表、客户端列表、钉钉 H5、权限与配置类接口的延迟问题。

**Architecture:** 先建立接口耗时基线和慢请求日志，再按“SQL 与索引、列表轻量化、权限/字典/配置缓存、前端按需加载、压测与文档沉淀”分层优化。所有高频接口都要避免列表加载大字段，权限体系要减少重复查询，钉钉 H5 只按当前用户和当前 tab 拉取数据。

**Tech Stack:** Go、Hertz、GORM、MySQL、Redis、Vue 3、Element Plus、uni-app、Vite、Node.js 性能检查脚本。

---

## 当前执行状态

| 优先级 | 任务 | 状态 | 证据 |
|---|---|---|---|
| P0 | Task 1-2 | 已落地 | `scripts/check-performance.mjs`、业务 `code` 校验、`[SLOW_REQUEST]` 日志、`npm run check:performance` |
| P1 | Task 3-6 | 已落地 | 用户/管理员/角色/客户端列表索引与轻量查询、角色授权聚合查询、钉钉 H5 bootstrap/reviews 拆分 |
| P2 | Task 7 | 已落地 | setup/dict/permission tree/钉钉 H5 菜单权限短 TTL 缓存 |
| P2 | Task 8 | 已落地 | 后台问卷、考试、打卡、赛事列表 DTO 与轻量列选择 |
| P2 | Task 9 | 已落地 | 管理后台动态路由、manualChunks、bundle budget，已用 `npm --prefix admin run build` 和 `npm --prefix admin run check:bundle` 复核 |
| P2 | Task 10 | 已落地 | 客户端 GET in-flight 去重、通知/打卡/赛事/问卷/考试列表分页与懒加载守护、钉钉 H5 当前 tab 刷新 |
| P3 | Task 11-12 | 已落地 | `docs/performance/*`、`CHECK_PERFORMANCE=1 bash scripts/check.sh`、`node scripts/check-quality-gates.mjs` |

> 已在本地 MySQL / Redis / 临时后端 18083 上携带有效 token 执行 `WECHECKIN_PERF_STRICT=1 npm run check:performance`，所有计划内接口 P95 均达标；结果记录在 `docs/performance/api-baseline.md`。

---

## 性能目标

| 模块 | 目标 |
|---|---|
| 后台管理 | `/api/v2/admin/users`、`/api/v2/admin/managers` 本地热数据 P95 控制在 250ms 内 |
| 权限配置 | 权限树、角色权限、应用权限读取使用短 TTL 缓存，授权变更后主动失效 |
| 客户端 | 通知、打卡、赛事、问卷、考试列表只返回卡片字段，首屏避免大 JSON 和富文本 |
| 钉钉 H5 | `/api/v2/dingtalk/h5/bootstrap` P95 控制在 150ms 内，工作台只返回统计，绩效数据按 tab 分页加载 |
| 数据库 | 高频查询拥有匹配索引，慢 SQL 能通过 route/requestId 定位 |
| 前端 | 后台设计器、富文本、图表、二维码等重模块按路由加载，移动端搜索和 tab 不重复触发请求 |
| 回归 | `npm run check:performance` 可以输出关键接口 min/avg/p95/max，严格模式可阻断发布 |

---

## P0：先补观测和回归基线

### Task 1: 关键接口性能检查脚本

**Files:**
- Create/Modify: `scripts/check-performance.mjs`
- Modify: `package.json`
- Modify: `README.md`

- [x] 脚本读取 `WECHECKIN_PERF_BASE_URL`、`WECHECKIN_ADMIN_TOKEN`、`WECHECKIN_USER_TOKEN`、`WECHECKIN_DINGTALK_TOKEN`。
- [x] 覆盖接口：后台用户、后台管理员、角色、钉钉 H5 bootstrap/workbench/reviews、通知、打卡、赛事、问卷、考试。
- [x] 每个接口请求 5 次，输出 `min / avg / p95 / max`。
- [x] 解析响应 JSON，只有业务 `code = 0` 才计入成功，未登录或权限错误会在 strict 模式下阻断。
- [x] 默认模式只告警；`WECHECKIN_PERF_STRICT=1` 时超过阈值退出非 0。
- [x] 文档说明本地运行方式和 token 获取方式。

### Task 2: 慢请求日志

**Files:**
- Modify: `backend/internal/middleware/access_log.go`
- Modify: `backend/internal/middleware/access_log_test.go`
- Modify: `backend/.env.example`
- Modify: `docs/DEPLOYMENT_TROUBLESHOOTING.md`

- [x] 增加 `WECHECKIN_SLOW_REQUEST_MS`，默认 800ms。
- [x] 超阈值请求单独输出 `[SLOW_REQUEST]`，字段包含 method、path、status、duration、requestId。
- [x] 慢日志不输出 body/query 敏感内容。
- [x] 上传请求、大 body、密码、token、手机号、富文本、答卷内容继续脱敏或跳过。

---

## P1：优先优化当前慢接口

### Task 3: 后台用户列表 SQL 收口

**Files:**
- Modify: `backend/internal/app/service/adminuser/service.go`
- Modify: `backend/internal/app/service/adminuser/performance_structure_test.go`
- Create/Modify: `backend/migrations/20260728210000_add_user_list_indexes.sql`

- [x] 用户列表只查询必要字段，不加载详情字段、富文本、扩展 JSON。
- [x] 搜索字段收敛到当前业务仍使用的 `user_name`、`user_mobile`。
- [x] 部门、岗位、角色名称批量查询后用 map 合并，禁止循环内查询。
- [x] 增加索引：`user_status + user_role_id`、`user_add_time + id`、`user_mobile`、`user_name`、`user_depts` 双向索引。
- [x] 保留搜索态实时 count，非搜索态可以按后续缓存策略优化。

### Task 4: 后台管理员列表改为 users 表轻量查询

**Files:**
- Modify: `backend/internal/app/service/adminmgr/service.go`
- Modify: `backend/internal/app/service/adminmgr/performance_structure_test.go`
- Create/Modify: `backend/migrations/20260728211000_add_admin_manager_indexes.sql`

- [x] 管理员列表只查 `users` 表中 `user_role_id > 0` 的用户。
- [x] 移除历史 admin 表 fallback 和历史后台账号字段依赖。
- [x] 角色、部门信息批量查询。
- [x] 增加索引：`user_role_id + user_status + user_add_time`、`user_role_id + user_login_time`。

### Task 5: 客户端列表轻量化

**Files:**
- Modify: `backend/internal/app/service/survey/client.go`
- Modify: `backend/internal/app/service/exam/client.go`
- Modify: `backend/internal/app/service/news/service.go`
- Modify: `backend/internal/app/service/enroll/client.go`
- Modify: `backend/internal/app/service/event/client.go`
- Create/Modify: `backend/migrations/20260728212000_add_client_list_query_indexes.sql`
- Existing: `backend/migrations/20260727165000_add_content_log_indexes.sql`

- [x] 通知、问卷、考试、打卡、赛事列表只返回卡片展示字段。
- [x] schema、题目、表单配置、富文本、答卷详情、大图集合只在详情接口返回。
- [x] 补齐 `status + sort/time + id` 类列表索引。
- [x] 增加结构测试，禁止列表重新退回 `Select("*")` 或加载大字段。

### Task 6: 钉钉 H5 bootstrap 和绩效接口拆分

**Files:**
- Modify: `backend/internal/app/service/dingtalkh5/reviews.go`
- Modify: `backend/internal/app/service/dingtalkh5/workbench.go`
- Modify: `backend/internal/app/handler/client/dingtalkh5/handler.go`
- Modify: `dingtalk-h5/pages/index/index.vue`
- Create/Modify: `backend/migrations/20260728213000_add_dingtalk_h5_review_indexes.sql`

- [x] `bootstrap` 只返回当前用户、一级菜单、二级 tab、权限版本号。
- [x] `workbench` 只返回统计卡片，不返回绩效列表、模板、组织架构。
- [x] `/reviews` 支持 `page`、`pageSize`、`status`、`period`、`scope`，默认 `pageSize=20`。
- [x] 绩效流程历史按 review IDs 批量查询，禁止单条绩效循环查历史。
- [x] 前端切换 tab 时按需请求当前 tab 数据。

---

## P2：缓存、DTO 和前端渲染优化

### Task 7: 权限、字典、系统配置短 TTL 缓存

**Files:**
- Modify: `backend/internal/app/support/permission/service.go`
- Modify: `backend/internal/app/service/adminpermission/service.go`
- Modify: `backend/internal/app/service/dict/service.go`
- Modify: `backend/internal/app/service/setup/cache.go`
- Modify: `backend/internal/app/service/role/service.go`

- [x] 钉钉 H5 菜单权限按 `userID:roleID` 短 TTL 缓存。
- [x] 权限树按 `platform + type` 短 TTL 缓存。
- [x] 角色列表一次性加载角色授权快照，避免菜单、接口、应用权限、数据范围分多次查询。
- [x] 字典类型、字典项、系统配置短 TTL 缓存。
- [x] 角色、用户、权限、字典、配置发生变更时主动失效缓存。
- [x] 缓存返回防御性副本，避免调用方修改全局缓存。

### Task 8: 列表 DTO 类型化

**Files:**
- Modify: `backend/internal/app/handler/admin/survey/dto.go`
- Modify: `backend/internal/app/handler/admin/exam/dto.go`
- Modify: `backend/internal/app/handler/admin/user/dto.go`
- Modify: `backend/internal/app/handler/admin/event/dto.go`
- Modify: `backend/internal/app/handler/admin/enroll/dto.go`

- [x] 高频列表接口定义明确 response DTO。
- [x] 列表 DTO 不包含 schema、富文本、题目详情、答卷详情、大图片集合。
- [x] 详情 DTO 保持完整字段。
- [x] 增加结构测试，防止列表 handler 继续扩大 `map[string]interface{}`。

### Task 9: 管理后台首屏拆包

**Files:**
- Modify: `admin/src/router/adminRoutes.ts`
- Modify: `admin/vite.config.ts`
- Modify: `admin/scripts/check-bundle-budget.mjs`
- Modify: `admin/src/views/layout/index.vue`

- [x] 问卷设计器、考试设计器、富文本、ECharts、二维码等重模块只在对应路由加载。
- [x] 系统管理页不引入设计器共享大组件。
- [x] 构建预算输出超过阈值的 chunk 名称和体积。
- [x] 保持角色、用户、权限页面首屏稳定。

### Task 10: 客户端和钉钉 H5 减少重复请求

**Files:**
- Modify: `frontend/utils/request.js`
- Modify: `frontend/pages/news/news_index.vue`
- Modify: `frontend/pages/enroll/enroll_index.vue`
- Modify: `frontend/pages/event/event_index.vue`
- Modify: `frontend/pages/survey/index.vue`
- Modify: `frontend/pages/exam/index.vue`
- Modify: `dingtalk-h5/pages/index/index.vue`

- [x] 搜索输入回车/确认后触发查询，不在输入过程中连续请求。
- [x] tab 切换相同参数不重复请求。
- [x] 分页追加只合并增量，不重置整页状态。
- [x] 图片使用懒加载和稳定尺寸，减少滚动重排。
- [x] 钉钉 H5 刷新当前 tab，不重新调用 bootstrap。

---

## P3：压测、EXPLAIN 和长期守护

### Task 11: MySQL EXPLAIN 与索引文档

**Files:**
- Create/Modify: `docs/performance/README.md`
- Create/Modify: `docs/performance/mysql-indexes.md`
- Create/Modify: `docs/performance/api-baseline.md`

- [x] 文档记录如何开启 MySQL slow query log。
- [x] 文档记录关键接口 EXPLAIN 检查项：`type`、`key`、`rows`、`Extra`。
- [x] 保存每轮 `npm run check:performance` 的接口基线。
- [x] 记录新增索引、命中场景、回滚方式。

### Task 12: 本地质量门禁

**Files:**
- Modify: `scripts/check.sh`
- Modify: `scripts/check-quality-gates.mjs`
- Modify: `README.md`

- [x] `CHECK_PERFORMANCE=1 bash scripts/check.sh` 时运行性能脚本。
- [x] 默认 CI 不阻断性能检查，发布前使用 strict 模式。
- [x] 增加前端 bundle budget、后台 Go 结构测试、钉钉 H5 bootstrap 轻量检查入口。

---

## 推荐执行顺序

| 顺序 | 优先级 | 任务 | 目的 |
|---|---|---|---|
| 1 | P0 | Task 1-2 | 先能看到慢在哪里，并能回归对比 |
| 2 | P1 | Task 3-4 | 直接处理后台用户/管理员列表 700ms 级延迟 |
| 3 | P1 | Task 6 | 直接处理钉钉 H5 bootstrap 3s 和页面一次拿全量数据 |
| 4 | P1 | Task 5 | 覆盖客户端首屏列表性能 |
| 5 | P2 | Task 7-8 | 减少重复查询和不稳定 payload |
| 6 | P2 | Task 9-10 | 降低前端首屏、页面切换和移动端滚动成本 |
| 7 | P3 | Task 11-12 | 沉淀成长期性能守护 |

---

## 验证命令

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/middleware -count=1
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/adminuser ./backend/internal/app/service/adminmgr -count=1
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/survey ./backend/internal/app/service/exam ./backend/internal/app/service/enroll ./backend/internal/app/service/event -count=1
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/dingtalkh5 ./backend/internal/app/support/permission -count=1
npm --prefix admin run build
npm --prefix dingtalk-h5 run check:scaffold
npm --prefix dingtalk-h5 run check:bootstrap
npm run check:performance
WECHECKIN_PERF_STRICT=1 npm run check:performance
```

---

## 完成标准

- 关键慢接口有优化前后耗时记录。
- 高频列表接口不加载大字段。
- 后台用户/管理员列表延迟明显下降。
- 钉钉 H5 bootstrap 不再返回大业务数据，绩效按用户权限和 tab 分页加载。
- 权限、字典、配置类重复读取有缓存并能主动失效。
- 管理后台和钉钉 H5 前端按需加载重模块。
- 文档记录 MySQL 索引、EXPLAIN、API baseline 和回归命令。
