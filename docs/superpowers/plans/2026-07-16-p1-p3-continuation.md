# P1-P3 后续优化执行计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成下一轮 P1-P3 优化，降低构建噪音、题库页维护成本，并补齐问卷/考试关键限制逻辑的自动化保护。

**Architecture:** 题库页从单文件继续拆成 focused components，主页面保留数据加载、筛选、保存、导入导出编排。后端先补不依赖真实 MySQL 的纯逻辑测试，优先覆盖问卷/考试限制配置解析和题目安全输出。构建产物通过 `.gitignore` 和删除已跟踪 build info 来避免反复污染工作区。

**Tech Stack:** Vue 3、Element Plus、Quill、Go、GORM、Node 检查脚本。

---

### Task 1: 题库页结构检查

**Files:**
- Create: `admin/scripts/check-question-bank-components.mjs`
- Modify: `admin/package.json`

- [ ] 新增脚本检查题库页必须导入拆分后的组件。
- [ ] 将脚本挂到后台项目 `package.json`。
- [ ] 先运行脚本，确认当前未拆分时失败。

### Task 2: 题库页组件拆分

**Files:**
- Create: `admin/src/views/question-bank/components/QuestionBankTable.vue`
- Create: `admin/src/views/question-bank/components/QuestionEditorDialog.vue`
- Create: `admin/src/views/question-bank/components/QuestionPreviewDrawer.vue`
- Modify: `admin/src/views/question-bank/QuestionBank.vue`

- [ ] 将表格渲染、编辑弹窗、完整富文本弹窗、预览抽屉从主页面拆出。
- [ ] 主页面只保留查询、导入导出、保存删除、schema 转换等编排逻辑。
- [ ] 运行结构检查和后台 build。

### Task 3: 问卷/考试限制逻辑测试

**Files:**
- Modify: `backend/internal/app/service/survey/service_test.go`
- Modify: `backend/internal/app/service/survey/client.go`
- Modify: `backend/internal/app/service/exam/service_test.go`
- Modify: `backend/internal/app/service/exam/client.go`

- [ ] 为问卷 `deviceLimit`、`ipLimit` 的数字/字符串容错解析写失败测试。
- [ ] 为考试限制解析抽出纯函数并写失败测试。
- [ ] 为考试题目安全输出补充隐藏答案/解析的测试。
- [ ] 实现最小逻辑并运行对应 package 测试。

### Task 4: 构建产物清理

**Files:**
- Modify: `.gitignore`
- Delete: `admin/tsconfig.tsbuildinfo`

- [ ] 忽略 `*.tsbuildinfo`。
- [ ] 删除已经被 Git 跟踪的后台 TypeScript build info。
- [ ] 运行 `git status --short` 确认后续 build 不再反复修改该文件。

### Task 5: 全量验证

- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./backend/...`。
- [ ] 运行 `npm --prefix admin run build`。
- [ ] 运行 `npm --prefix frontend run build:h5`。
- [ ] 运行 `git diff --check`。
