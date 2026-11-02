# 流程历史表单导入规则

## 目标

后台“流程设计器 → 流程配置”增加“导入规则设置”。管理员提前设置来源和字段映射；启用后 H5 用户选择历史记录即可导入，跳过手工配置与预览步骤。未配置或关闭预设时完整保留现有三步向导。

## 已确定的范围和交互

- 历史来源沿用当前流程中本人发起、当前用户可见的历史申请，不新增跨流程或跨用户读取。
- 来源方式：`select_history` 用户选一条记录后直接应用；`latest_history` 点击导入后自动取满足条件的最近一条记录。最近按既有接口 `start_time DESC, id DESC` 确定。
- 来源状态：`completed` 仅已完成（新配置默认）；`all` 所有现有可见状态。预设为 completed 时用户不能通过筛选扩大范围。
- 字段方式：`same_key` 按同名、类型兼容的字段匹配；`mapping` 使用管理员配置的一对一源/目标映射。
- 明细表字段可配置一层 `columns` 映射；不配置列映射时沿用原同名列规则。普通字段不能携带列映射。计算及展示字段不导入。
- 冲突策略：`overwrite=false` 默认只填空字段；true 可覆盖已有内容。选择记录前清楚显示覆盖策略，不再要求下一步选择设置。
- 映射基于当前流程表单配置。历史记录按它自己的已发布版本解析；字段缺失、不兼容或无可导入值时按已有安全归一化规则跳过并给出结果反馈，不静默切回手工向导。零个可导入字段不得假报成功。
- 读取失败、没有历史记录、规则不合法均显示可理解反馈及重试/关闭入口。关闭弹窗、切换流程或规则后，旧异步响应不得继续导入。

## 跨端契约

Definition / WorkflowDraft / WorkflowPublishedDefinition 新增可选 `historyImport`，并加入既有 `RuntimeStartConfig`。配置示例：

```json
{
  "historyImport": {
    "enabled": true,
    "sourceMode": "select_history",
    "sourceStatus": "completed",
    "mappingMode": "mapping",
    "overwrite": false,
    "mappings": [
      {"sourceKey":"next_month_goals","targetKey":"current_month_goals","columns":[{"sourceKey":"goal","targetKey":"goal"}]}
    ]
  }
}
```

TypeScript names: `WorkflowHistoryImportConfig`, `WorkflowHistoryImportMapping`, `WorkflowHistoryImportColumnMapping`. Go names: `HistoryImportConfig`, `HistoryImportMapping`, `HistoryImportColumnMapping`. Both mapping types contain `sourceKey` and `targetKey`; top-level mapping optionally contains `columns`.

No configuration (`nil` / omitted) or `enabled=false` means legacy wizard. New editor defaults are sourceMode=select_history, sourceStatus=completed, mappingMode=same_key, overwrite=false, mappings=[]. Enabled configs require valid enum values; omitted optional modes normalize consistently to these defaults. Mapping mode requires 1–100 mappings, with unique source and target keys; column maps have 1–100 unique source/target column keys when explicitly configured. Field-key length <=100 and fields must exist in the relevant current form schema. Known type-family compatibility follows the existing H5 importer (text/textarea/phone/email; number/amount; select/radio; multi_select/checkbox; otherwise exact type); display/calculation fields excluded. Detail-list custom mapping is limited to importable scalar columns, not recursive nested details.

## 保存、版本和权限

- 复用已有流程保存 API、workflow:edit 权限和 definition_start_config_json；不新增权限或接口、不修改历史发布版本 JSON。
- `RuntimeStartConfigFromDefinition` / clone / apply / validation / definition DTO 全链路携带 historyImport；旧 JSON 无字段时保持原行为。
- 复用当前流程配置的即时生效约定，并纳入 startConfigRevision。对已发布流程继续校验活动版本字段；引用尚未发布的字段时给出明确错误。
- 源历史数据仍由后端鉴权和 scope=started 查询保护。H5 保持源字段可访问约束、目标当前可写约束与现有值归一化；配置不能提升字段权限或覆盖计算字段。
- 关闭预设后保留管理员的配置草稿，用户恢复手工向导。配置未启用时不因后来删除的映射字段阻塞流程的其他编辑；仍限制配置体大小。

## 验证

后端：JSON 兼容、克隆隔离、运行配置覆盖、非法枚举/重复映射/不兼容字段/非法明细列、已发布字段校验、来源查询排序和归属既有约束。

Admin：创建/启停预设、普通与明细字段映射、只读权限、保存往返、旧配置兼容。

H5：未配置保留三步；已配置选择后无下一步；最近一次模式；空字段/覆盖策略；明细列映射；隐藏/只读/计算字段；空历史、读取失败、无有效字段及取消/竞态。桌面/紧凑 PC/手机验证使用模拟 API，禁止提交真实流程。
