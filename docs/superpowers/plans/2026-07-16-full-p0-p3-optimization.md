# 全量 P0-P3 优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 P0 到 P3 完整推进后端、管理端、客户端的结构、安全和可维护性优化。

**Architecture:** P0 先收后端数据一致性和 handler 职责边界；P1 继续类型化 DTO/API 与大设计器拆分；P2 补测试和题库/客户端页面拆分；P3 做构建产物治理与文档化。每一轮先加结构测试或低风险回归测试，再改生产代码，最后跑后端全量测试和前端构建/类型检查。

**Tech Stack:** Go、Hertz、GORM、Vue 3、Element Plus、uni-app、项目现有结构测试。

---

### Task 1: P0 Handler 直连 DB 下沉

**Files:**
- Create: `backend/internal/app/service/formkitadmin/resource.go`
- Create: `backend/internal/app/service/formkitadmin/question_bank.go`
- Create: `backend/internal/app/service/formkitadmin/template_presets.go`
- Modify: `backend/internal/app/handler/admin/survey/resource.go`
- Modify: `backend/internal/app/handler/admin/exam/resource.go`
- Modify: `backend/internal/app/handler/admin/survey/question_bank.go`
- Modify: `backend/internal/app/handler/admin/exam/question_bank.go`
- Modify: `backend/internal/app/handler/admin/survey/template_presets.go`
- Modify: `backend/internal/app/handler/handler_structure_test.go`

- [x] **Step 1: 写结构测试，禁止这些 handler 直接访问 database 包**
- [x] **Step 2: 新建服务层并迁移资源、题库、模板预设 DB 逻辑**
- [x] **Step 3: handler 保留参数和文件处理，只调用服务层**
- [x] **Step 4: 扩展到后台问卷/考试核心、报表、通知、渠道、统计 handler**
- [x] **Step 5: 扩展到客户端问卷/考试 handler，业务 handler 不再直接访问 database 包**

### Task 2: P0 删除/计数事务化

**Files:**
- Modify: `backend/internal/app/service/admincontent/enroll_records.go`
- Modify: `backend/internal/app/service/event/admin.go`
- Modify: `backend/internal/app/service/enroll/records.go`

- [x] **Step 1: 找出删除记录后更新计数的路径**
- [x] **Step 2: 用事务包住删除、计数和回写**
- [x] **Step 3: 增加结构测试防止关键路径回退到散写**

### Task 3: P1 后端 DTO 与 Context 继续收口

**Files:**
- Modify: `backend/internal/app/service/event/my.go`
- Modify: `backend/internal/app/service/event/dynamic.go`
- Modify: `backend/internal/app/service/event/score.go`
- Modify: `backend/internal/app/service/enroll/records.go`
- Modify: `backend/internal/app/service/exam/service.go`

- [x] **Step 1: 将 event secondary、exam admin/client、survey admin/client 的分页/统计返回改为具名 DTO**
- [x] **Step 2: 服务层补 Context 版本，旧函数委托**
- [x] **Step 3: handler 调用 Context 版本**
- [x] **Step 4: 继续收口 enroll/records 与 poststat 的 Context/DTO**

### Task 4: P1 管理端 API 类型化

**Files:**
- Modify: `admin/src/api/index.ts`
- Create: `admin/src/api/types.ts`

- [x] **Step 1: 定义分页、用户、问卷、考试、题库核心响应类型**
- [x] **Step 2: 高频 API 方法替换裸 `any`**
- [x] **Step 3: 跑管理端类型检查或构建**

### Task 5: P1 设计器共享组件和 composable

**Files:**
- Modify: `admin/src/views/survey/SurveyDesigner.vue`
- Modify: `admin/src/views/exam/ExamDesigner.vue`
- Create/Modify: `admin/src/views/shared/formkit/*`

- [ ] **Step 1: 抽共享题型/大纲/题库面板样式与逻辑**
- [ ] **Step 2: 问卷和考试设计器复用共享实现，保留各自主题色**
- [ ] **Step 3: 做视觉回归检查**

### Task 6: P2 测试补齐和题库拆分

**Files:**
- Modify: `admin/src/views/question-bank/QuestionBank.vue`
- Create: `admin/src/views/question-bank/components/*`
- Test: `backend/internal/app/service/**`

- [x] **Step 1: 题库页面拆出导入导出工具模块，保留 UI 行为兼容**
- [x] **Step 2: 补 handler/service 结构测试，覆盖 DB 下沉、事务和 Context 约束**
- [ ] **Step 3: 继续拆题库列表、编辑弹窗、富文本编辑器组件**

### Task 7: P2 客户端大页面拆分

**Files:**
- Modify: `frontend/pages/admin/enroll/admin_enroll_edit.vue`
- Modify: `frontend/pages/enroll/enroll_detail.vue`
- Modify: `frontend/pages/exam/fill.vue`
- Modify: `frontend/pages/survey/fill.vue`

- [x] **Step 1: 抽 exam/survey fill 共享表单工具函数**
- [x] **Step 2: 保持页面行为和接口字段兼容**
- [ ] **Step 3: 继续抽答题卡、提交状态、倒计时组件**

### Task 8: P3 构建产物治理

**Files:**
- Modify: `.gitignore`
- Modify/Create: `docs/project-maintenance.md`

- [x] **Step 1: 确认 `frontend/dist`、`frontend/unpackage`、`node_modules` 是否被版本管理**
- [x] **Step 2: 只更新忽略规则，不删除用户未确认的产物**
- [x] **Step 3: 文档记录构建产物治理方式**

### Verification

Run after each backend slice:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/...
git diff --check
git diff --cached --check
```

Run after frontend/admin slices:

```bash
npm --prefix admin run build
npm --prefix frontend run build:h5
```

## 进度记录

- P0 Task 1 已完成：新增 `formkitadmin` 服务层，问卷/考试资源、题库、模板预设 handler 已移除直接 DB 访问。
- P0 Task 1 扩展完成：后台问卷/考试核心管理、通知、渠道、统计、报表 handler，以及客户端问卷/考试 handler 均已移除直接 DB 访问，并新增结构测试保护。
- P0 Task 2 已完成：打卡记录删除、移除打卡用户、活动删除、活动参与者删除已事务化并补结构测试。
- P1 Task 3 已完成本轮计划项：event secondary、exam admin/client、survey admin/client、enroll/records 与 poststat 已补 Context 服务方法和结构测试保护；少量兼容旧接口仍保留 map 返回，后续可逐接口做破坏性更小的 DTO 替换。
- P1 Task 4 已完成：`admin/src/api/types.ts` 与高频 API 类型化，`npm --prefix admin run build` 已通过。
- P2 Task 6 部分完成：题库导入导出逻辑拆到 `admin/src/views/question-bank/utils/importExport.ts`，管理端构建通过。
- P2 Task 7 部分完成：考试/问卷 fill 页面共享工具拆到 `frontend/utils/formFill.js`，H5 构建通过。
- P3 Task 8 已完成：确认构建产物未被 Git 跟踪，补充 `docs/project-maintenance.md`。
- 已执行：`GOCACHE=$PWD/.cache/go-build go test ./backend/...`、`git diff --check`、`git diff --cached --check`。
