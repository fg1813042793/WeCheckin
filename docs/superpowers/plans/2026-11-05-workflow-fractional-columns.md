# Workflow Fractional Columns Implementation Plan

> Execute in this session; use parallel workers only for disjoint backend and frontend files. Preserve all unrelated changes and do not commit or deploy.

**Goal:** 普通字段、分组字段和明细列支持 1/5、1/6、1/8，并支持 6/8 + 1/8 + 1/8 同行。

**Architecture:** 保留 `span / 24` 的现有语义及旧值 0/6/8/12/24，增加 3、4、4.8、18，分别对应 1/8、1/6、1/5、6/8。Go span 接受这些明确的数值，拒绝其他值。Admin 用精确宽度布局，H5 继续百分比布局；导出按需要使用更细网格表达 1/5，不回写历史定义或实例。

**Tech Stack:** Vue 3 / TypeScript / Element Plus / uni-app / Go。

## Tasks

- [x] 后端：修改 `backend/internal/workflowcore/types.go`、`validation.go` 与用例，验证新比例能保存、旧比例有效、非法 7 被拒绝；同步 `workflowsummary` 的 PDF/Word/Excel 列宽计算。
- [x] Admin：共享列宽选项与宽度计算；更新设计器、嵌套预览和运行时普通字段/明细列；新增选项使用四列两行排列以保持可读。
- [x] H5：接受 3/4/4.8/18，仍以 `span / 24 * 100%` 计算宽度，保持现有移动布局策略。
- [x] 回归：精确五等分、六等分、八等分、75%/12.5%/12.5%，保存重载、旧表单及非法配置回退；检查手机与 PC、明细列与普通字段。
- [x] 交付：更新契约文档/Swagger（需要时由命令生成），执行 Backend / Admin / H5 相应检查及浏览器验证，说明未通过的基线检查和未部署状态。

**Commands:** Backend `GOCACHE=$PWD/../.cache/go-build go test ./...`；Admin `npm run check:workflow-form-designer`、`npm run check:all`、`npm run build`；H5 使用项目指定 pnpm 10.34.5 执行 `lint`、`type-check`、`build:h5` 和相关流程检查。

## Verification result

- Backend: workflowcore 与 workflowsummary 测试通过；全量测试仍有既有 5 个失败包（cmd Swagger、钉钉设置、审计分类、配置安全、HTTP 错误边界）。
- Admin: 列宽保存重载测试、设计器检查与构建通过；check:all 仍被 user/index.vue 1867 行超过1859预算阻断。
- H5: 列宽/画布检查、改动文件 lint、type-check 与 build:h5 通过；全量 lint 仍为既有215错误、2警告。
- 浏览器: 设计器选项、明细列保存重载、后台运行表单及 H5 的五/六/八等分和 6/8+1/8+1/8 同行通过；H5 800/1023/1366/1920 和手机360/390验证通过，设计器1366窄列无内容溢出。
- 导出: Word/Excel 网格与合并、PDF矩形宽度回归通过；未进行真实 Office 客户端验证。
- 未提交、部署或修改生产数据；已移除临时浏览器页面，关闭本任务测试服务。
