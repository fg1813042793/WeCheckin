# Workflow Core 包重命名设计

## 背景

后端同时存在 `internal/workflow` 和 `internal/modules/workflow`。前者承载流程定义、表单校验和 BPMN 编译，后者承载运行时状态机及应用服务。两者职责不同，但目录名称接近，阅读代码和讨论架构时容易混淆。

## 决策

将核心定义包完整重命名：

- 目录从 `backend/internal/workflow` 改为 `backend/internal/workflowcore`。
- Go 包声明从 `package workflow` 改为 `package workflowcore`。
- 所有直接调用方改为导入 `wecheckin/backend/internal/workflowcore`，移除不再需要的显式导入别名。
- `backend/internal/modules/workflow` 保持原目录和职责不变。

不采用以下方案：

- 只改目录但保留 `package workflow`：导入路径和包标识不一致，仍然需要解释两套名称。
- 移入 `internal/modules/workflow/core`：会改变当前独立核心包的边界，扩大依赖调整，并增加领域包与核心定义包发生循环依赖的风险。

## 改动范围

实施仅包含：

1. 移动核心包现有源码和测试文件，保留文件内容及未提交修改。
2. 更新后端 Go 源码和测试中的旧导入路径。
3. 更新现行架构文档及 Backend 开发规范中的目录说明。
4. 更新仍作为当前执行入口使用的测试命令或路径引用。

历史设计、历史实施计划和发布记录继续保留当时路径，不做批量回写。

## 兼容性

本次调整不修改流程定义 JSON、数据库表、API、DTO、权限、状态机行为或 BPMN 输出。已发布定义仍通过 `DefinitionID + DefinitionVersion` 绑定原发布快照，历史实例语义不变。

Go 的 `internal` 包没有对仓库外提供稳定导入契约，因此不保留旧路径兼容转发包。仓库内任何遗漏引用都应在编译阶段失败。

## 验证

实施后执行：

1. 对移动后的 Go 文件运行 `gofmt`。
2. 运行 `GOCACHE=$PWD/../.cache/go-build go test ./internal/workflowcore ./internal/modules/workflow/... ./internal/service/admin/workflow -count=1`。
3. 运行 `GOCACHE=$PWD/../.cache/go-build go test ./... -count=1`。
4. 搜索非历史文件，确认没有残留 `wecheckin/backend/internal/workflow` 导入或错误的 `package workflow` 声明。
5. 运行 `git diff --check`，确认重命名没有引入格式问题。

## 成功标准

- 核心定义包只存在于 `backend/internal/workflowcore`，包名为 `workflowcore`。
- `backend/internal/modules/workflow` 的职责、API 和运行时行为不变。
- 当前源码和现行架构规范统一使用新名称。
- 后端全量测试通过，或如实记录与本次改动无关的既有失败。
