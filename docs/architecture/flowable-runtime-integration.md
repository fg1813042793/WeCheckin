# Flowable 运行时集成草案（已废弃）

本草案已废弃，不作为当前项目的实现或部署依据。

当前工作流采用 **Go 原生通用流程引擎**，不依赖 Java、Flowable 服务或 Flowable 数据库。最新架构、模块边界、运行时模型与接口约定见：

- [Go 通用工作流引擎 V1](./go-workflow-engine-v1.md)

当前实现保持以下边界：

- 工作流引擎位于 `backend/internal/modules/workflow`，独立维护领域状态机、运行时持久化和管理接口；
- 业务通过 `businessType`、`businessKey` 和通用变量关联流程实例；
- 不修改、不替换现有钉钉 H5 绩效流转逻辑；
- 后续业务接入需要单独实现业务适配层，不能把业务字段写入工作流核心模型。
