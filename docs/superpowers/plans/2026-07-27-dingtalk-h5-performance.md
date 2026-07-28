# 钉钉H5绩效系统迁移 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `.test/performance-system-2` 的绩效收集能力迁移到独立 `dingtalk-h5` 前端，并在后端新增 `/api/v2/dingtalk/h5` 隔离接口与 MySQL 表。

**Architecture:** 后端新增 `model + service/dingtalkh5 + handler/client/dingtalkh5 + routes_v2_dingtalk.go`，业务数据使用独立表前缀 `dingtalk_h5_perf_`。前端保留 uni-app H5 工程，改造成钉钉蓝风格响应式单页工作台，PC 端三栏、手机端卡片流。

**Tech Stack:** Go、Hertz、GORM、MySQL、uni-app、Vue 3、Vite。

---

### Task 1: 后端结构测试

**Files:**
- Create: `backend/internal/model/dingtalk_h5_performance_test.go`
- Create: `backend/cmd/routes_v2_dingtalk_test.go`
- Create: `backend/internal/app/service/dingtalkh5/service_test.go`

- [ ] 写失败测试，要求新增模型表名、独立路由前缀、状态流转函数。
- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/model ./backend/cmd ./backend/internal/app/service/dingtalkh5 -count=1`，确认因缺少实现失败。

### Task 2: 后端模型与迁移

**Files:**
- Create: `backend/internal/model/dingtalk_h5_performance.go`
- Modify: `backend/internal/bootstrap/migrate.go`
- Create: `backend/migrations/20260727180000_add_dingtalk_h5_performance_tables.sql`

- [ ] 新增 `DingTalkH5PerfUser`、`DingTalkH5PerfSession`、`DingTalkH5PerfReview`、`DingTalkH5PerfHistory`、`DingTalkH5PerfTemplate`。
- [ ] 将新模型加入启动期 `AutoMigrate`。
- [ ] 增加 MySQL 版本化建表 SQL，便于单点部署手动迁移。

### Task 3: 后端服务层

**Files:**
- Create: `backend/internal/app/service/dingtalkh5/types.go`
- Create: `backend/internal/app/service/dingtalkh5/defaults.go`
- Create: `backend/internal/app/service/dingtalkh5/auth.go`
- Create: `backend/internal/app/service/dingtalkh5/reviews.go`
- Create: `backend/internal/app/service/dingtalkh5/users.go`
- Create: `backend/internal/app/service/dingtalkh5/export.go`

- [ ] 实现默认种子数据、登录、token 校验、bootstrap、绩效单创建、保存、提交、退回、撤销、归档、汇总筛选、导出、人员维护。
- [ ] 所有 DB 操作使用 `database.WithContext(ctx)`。

### Task 4: 后端 Handler 与路由

**Files:**
- Create: `backend/internal/app/handler/client/dingtalkh5/handler.go`
- Create: `backend/cmd/routes_v2_dingtalk.go`
- Modify: `backend/cmd/routes_v2.go`

- [ ] 新增 `/api/v2/dingtalk/h5` 路由组，不套用现有客户端/后台鉴权中间件。
- [ ] 所有接口返回统一 `response.Resp` 格式。

### Task 5: dingtalk-h5 前端迁移

**Files:**
- Modify: `dingtalk-h5/api/index.js`
- Modify: `dingtalk-h5/pages/index/index.vue`
- Modify: `dingtalk-h5/utils/request.js`
- Modify: `dingtalk-h5/README.md`

- [ ] 接入新 API 前缀。
- [ ] 实现登录、工作台、绩效详情编辑、汇总筛选导出、组织架构、模板、账号设置。
- [ ] 样式贴近钉钉，支持 PC 三栏和手机单栏。

### Task 6: 验证

**Files:**
- Modify: `dingtalk-h5/scripts/check-scaffold.mjs`

- [ ] 补充前端检查脚本，确认关键 API 前缀、页面视图、钉钉主色存在。
- [ ] 运行 `npm --prefix dingtalk-h5 run check:scaffold`。
- [ ] 运行 `npm --prefix dingtalk-h5 run build:h5`。
- [ ] 运行后端相关 `go test`。
