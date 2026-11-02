# Workflow Form-only Export Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development for the bounded backend task; integrate the H5 change and review both contract and rendered output.

**Goal:** 汇总页新增“导出表单”，保留标题和表单内容，省略流程元信息及审批记录。

**Architecture:** 复用现有 `/api/v2/dingtalk/h5/workflows/summary/export`，增加 `content=full|form`，缺省为 full。服务端在构造文档时选择内容，三种渲染器共用同一模型；H5 工具栏及单行按钮共用现有格式、选中记录和导出权限。

**Tech Stack:** Go/Hertz、Vue 3/uni-app/uView Pro、现有 XLSX/PDF/DOCX 渲染器。

## Backend

- [x] 在 `backend/internal/service/dingtalkh5/workflowsummary` 和 `backend/internal/modules/workflow/transport/httpclient/summary_handler_test.go` 添加失败测试：form 不含元数据和审批任务，保留名称、分组、明细及评分字段；full/缺省保持原行为；非法 content 返回参数错误；两种模式均执行访问控制。
- [x] 在 ExportRequest 增加 Content 枚举并校验，handler 读取 content；文档构造按模式控制 Metadata 和审批记录 section。纯表单文件/工作表采用标题及重名序号，不使用流程内部编号。
- [x] 验证 XLSX 单/多工作表、PDF/Word 单文档及批量 ZIP；移除元信息后没有空表格或多余占位。保留实例所属表单版本。
- [x] 更新 Swagger 声明并通过规定 swag 命令生成，执行受影响包测试及 Go 全量检查。

## H5

- [x] 给 `h5app/src/api/workflow.ts` 的导出 URL 和按钮行为增加验证：默认 full、新按钮 form、选中 ID 与格式不变、无权限/未选中不能发起请求。
- [x] 在 `h5app/src/types/workflow.ts` 增加内容类型，API 新增可选参数；`WorkflowSummarySection.vue` 中共用 exportInstances(ids, content)。
- [x] 在批量导出后增加“导出表单”，行操作“导出”后同样增加入口。使用 u-button、已选为空时禁用、复用 canExport、更新中英文文案；调整操作列宽及窄屏换行。
- [x] 使用本机演示数据验证桌面/窄屏、单条/批量、格式切换、无权限和新旧 URL，检查导出文件内容。
- [x] 执行聚焦脚本、`pnpm lint`、`pnpm type-check`、`pnpm build:h5`，报告已有失败与实测边界。

## Delivery

- [x] 更新功能/API 文档，审查全部改动，不提交或覆盖工作区无关修改。


## 验证结果（2026-10-30）

- 后端导出、HTTP handler、错误映射聚焦测试通过；`./cmd ./internal/middleware ./internal/routes/v2/dingtalkh5 ./internal/support/appapiperm` 契约测试通过。
- `swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger --templateDelims '{%,%}'` 已生成；新参数不变更路由权限。
- `pnpm check:workflow-summary-export`、列及列设置检查通过；`pnpm type-check`、`pnpm build:h5`、`pnpm check:ui-style` 和本次修改文件的 ESLint 通过。运行 pnpm 时关闭了启动器自动切换包管理器版本及自动依赖安装检查，未改依赖。
- 本机浏览器演示数据：验证 1440/900/390px、初始禁用、单条及批量三种格式、新旧模式参数、无导出权限隐藏；修复移动端被桌面内联列宽覆盖的问题，390px 下列表卡片宽 346px、操作区宽 320px。
- 通过真实 Service.Export 生成 XLSX/PDF/DOCX 及批量 ZIP/工作簿样例，检查历史字段、实际表单内 HRBP/评分内容、文件命名及元数据/审批记录排除；原生预览核对中文内容。浏览器使用本地演示接口，未连接正式数据或实际钉钉客户端。
- `GOCACHE=$PWD/../.cache/go-build go test ./...` 完成但未全绿：既有钉钉通知页面文案断言、通知记录删除的审计分类、受跟踪生产配置检查失败。初次沙箱临时端口错误经授权重跑排除。
- H5 全量 `pnpm lint` 仍有工作区既有报错；`pnpm check:workflow-module` 仍被 WorkflowCenter 中已有“发起审批”文字断言拦截。独立旧 summary-layout 脚本仍检查已弃用的 12 列筛选网格。未修改这些无关实现。
- 浏览器初始化可见已有 uni scroll-view 空引用 scrollLeft 异常，未阻止导出及布局检查；未在本次扩展修改框架组件。
