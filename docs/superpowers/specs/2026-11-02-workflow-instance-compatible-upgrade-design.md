# 流程实例兼容版本升级设计

## 目标与范围

后台流程实例列表为运行中的旧版本实例提供“升级版本”操作。第一期只支持同一流程定义内、节点和表单结构兼容的原地升级；不重建 Token、待办、参与人、通知或审批历史，不处理节点删除、节点类型变化和字段删除等结构迁移。

## 兼容性规则

预检目标固定为流程定义当前发布版本。实例必须处于 `running`，目标版本必须高于实例版本，并满足：

- 所有 `active`/`waiting` Token 的节点 ID 在目标版本存在且节点类型一致。
- 所有 `pending`/`waiting` 待办的节点 ID 在目标版本存在且节点类型一致。
- 实例 `formData` 中已有的字段 key 在目标版本表单中仍存在。
- 实例没有活动中的表单修订请求。

检查失败时返回稳定的阻断原因，不修改实例。兼容性检查只决定能否原地升级，不自动修复表单数据、Token 或待办。

## 事务与审计

确认升级时，应用层在事务内锁定实例，重新加载源版本和最新发布版本并再次执行兼容性检查。检查通过后使用 `WHERE id = ? AND definition_version = ? AND instance_status = 'running'` 更新实例版本，随后追加 `instance_version_upgraded` 历史事件，记录旧版本、新版本和管理员 ID。任何一步失败均回滚。

升级不修改实例标题、业务标识、表单数据、FormRevision、Token、任务及已有历史。旧发布版本继续保留，版本删除保护继续生效。

## API、权限与界面

- `GET /api/v2/admin/workflow-instances/:id/version-upgrade-preview`
- `POST /api/v2/admin/workflow-instances/:id/version-upgrade`
- 业务权限：`workflow:instance:upgrade`
- 菜单按钮：`admin:menu:workflow:instance:upgrade`
- API 权限：`admin:api:workflow:instance:upgrade-preview`、`admin:api:workflow:instance:upgrade`

列表中仅对运行中的旧版本实例显示“升级版本”。点击后先展示当前版本、目标版本、兼容状态、阻断原因和提醒；兼容时管理员二次确认后执行，成功后刷新列表和详情。

## 验证

覆盖兼容升级、最新版本、终态实例、缺失节点、节点类型变化、字段缺失、活动表单修订、并发版本变化、事务失败回滚、权限目录/路由、Admin 按钮及生产构建。结构不兼容实例必须保持原版本和运行状态不变。
