# Admin UI 开发规范

## 1. 目标与适用范围

本规范用于 `admin` 端后续新增的业务页面，目标是统一页面结构、间距、容器尺寸与状态反馈。现有业务页面属于旧路由基线，不要为了通过新规范而批量改造。

公共层的责任：

- `src/styles/admin-ui-tokens.css` 管理颜色、间距、控件高度、表格高度、圆角、页面边距和断点变量。
- `src/components/admin-ui/` 管理页面壳、标题、搜索区、表格容器、弹窗和抽屉的结构与尺寸。
- 业务页面只管理字段、列、权限、请求、业务校验和业务状态。

## 2. 新路由 UI 契约

新增业务路由时，必须在路由 `meta` 中声明 `adminUi`：

```ts
{
  path: 'example',
  name: 'ExampleList',
  component: () => import('@/views/example/ExampleListView.vue'),
  meta: {
    title: '示例列表',
    menuPath: '/example',
    adminUi: {
      version: 1,
      pattern: 'filter-list',
    },
  },
}
```

`pattern` 只允许以下值：

| 值 | 适用页面 |
| --- | --- |
| `list` | 普通列表 |
| `filter-list` | 包含多项筛选的列表 |
| `form` | 新增、编辑或配置表单 |
| `detail` | 详情与审阅页 |
| `workspace` | 设计器、编排器等工作台 |

新路由的本地 Vue 页面必须从 `@/components/admin-ui` 导入并渲染 `AdminPageShell`。`npm run check:admin-ui-contract` 会对比 `scripts/admin-ui-legacy-routes.json`，只检查基线之外的路由。

不得把新路由追加到旧路由基线以绕过检查。修改旧路由的 `name` 或 `path` 会被视为新路由；如确属路由迁移，需在代码评审中同时说明基线调整原因。

## 3. 页面布局

页面根节点使用 `AdminPageShell`：

- `width="fluid"`：默认，适合数据表格和工作台。
- `width="wide"`：最大宽度 `1600px`，适合常规列表。
- `width="form"`：最大宽度 `960px`，适合单列表单。
- `density="compact"`：只用于明确需要紧凑信息密度的页面。

页面内容默认纵向排列，区块间距为 `16px`。禁止业务页面重新定义页面级 `padding`、最大宽度或区块间距；确有特殊布局时，应先扩展公共组件的可选模式。

## 4. 断点与响应式

| 范围 | 用途 |
| --- | --- |
| `>= 1200px` | 标准 PC，搜索项默认四列 |
| `768px - 1199px` | 窄屏 PC 或平板，搜索项两列 |
| `< 768px` | 移动端，搜索项单列，页面边距 `12px` |

CSS 自定义属性不能直接用于媒体查询，因此断点值在 token 文件中作为契约记录，组件媒体查询使用对应的固定值。不要在业务页面创建新的全局断点。

## 5. 页面标题

使用 `AdminPageHeader` 呈现标题、说明、状态、元数据和页面级操作。

- 主标题应与路由和菜单名称一致。
- 状态使用 `status` 插槽，操作使用 `actions` 插槽。
- 返回使用 `back` 与 `@back`，不在业务页面重新制作返回按钮。
- 移动端操作自动换行，业务文案不应使用固定宽度强行压缩。

## 6. 搜索与筛选

使用 `AdminSearchBar`，搜索项直接放入默认插槽。

- 普通列表优先使用 1-3 个条件。
- 复杂筛选可使用 `collapsible`，PC 默认展开，移动端是否默认折叠由业务场景决定。
- 查询通过 `@search`触发，重置通过 `@reset` 触发。
- 搜索区不放列表新增、导出等页面级操作，这些操作放在页面标题或表格工具栏。
- 日期区间、部门树、状态等使用结构化组件，不用自由文本代替。

## 7. 表格与列表

使用 `AdminTablePanel` 包裹表格，由容器统一工具栏、计数、加载、空状态、错误状态和分页区域。

- 列宽根据内容设定，操作列固定在右侧并保持按钮对齐。
- 长文本使用省略与 tooltip，不让单行数据无限撑高。
- 表头高度、数据行最小高度由 token 管理。
- 移动端如不适合横向表格，可在业务组件中改为紧凑列表，但页面外层仍使用 `AdminTablePanel`。
- 分页放在 `footer` 插槽，不与表格内容重叠。

## 8. 表单

- 字段标签、必填标识和校验信息使用 Element Plus 表单能力。
- 同一表单的标签位置保持一致；移动端优先使用顶部标签。
- 不可编辑使用 `disabled`，仅供查看使用只读呈现，隐藏字段不参与客户端校验。
- 提交期间锁定重复操作并显示加载状态，服务端仍需完成最终校验。
- 控件高度使用 token，不在页面中逐个覆盖 Element Plus 尺寸。

## 9. 弹窗与抽屉

### AdminDialog

用于短表单、确认与小型详情。`width` 优先使用 `sm`、`md`、`lg`，特殊场景才传入自定义尺寸。默认挂载到 `body`、禁止点击遮罩关闭，内容超高时只滚动弹窗正文。

### AdminDrawer

用于较长详情、审核和需保留列表上下文的操作。`size` 优先使用 `sm`、`md`、`lg`。正文独立滚动，启用底部操作后底部区域固定。移动端复杂交互优先新开路由页，不强行使用窄抽屉。

## 10. 状态反馈

- 列表加载使用 `AdminTablePanel` 的 `loading`。
- 无数据使用 `empty` 和 `empty-text`，请求失败使用 `error`，不把错误误报为空数据。
- 权限不足应展示明确的禁用或无权状态，不仅依赖隐藏按钮。
- 成功使用简短的 message，需要用户决策的风险操作使用确认弹窗。
- 错误信息面向用户描述可执行的下一步，不直接显示服务端堆栈或数据库错误。

## 11. 参考样板

- `src/examples/admin-ui/SimpleListExample.vue`：普通列表、简单查询和标准弹窗。
- `src/examples/admin-ui/AdvancedFilterListExample.vue`：复杂筛选、表格状态和标准抽屉。

两个样板只作为源码参考，不注册路由、不挂载菜单、不发起网络请求。

## 12. 检查命令

```bash
npm run check:admin-ui-contract
npm run check:all
```

`check:admin-ui-contract` 检查公共基础文件、新路由契约、`AdminPageShell` 使用情况以及样板页的静态性。`check:all` 已包含该检查。
