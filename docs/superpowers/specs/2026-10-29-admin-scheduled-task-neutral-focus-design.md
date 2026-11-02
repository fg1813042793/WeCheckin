# Admin 定时任务表单单层焦点边框设计

## 目标

将 Admin 定时任务创建和编辑弹窗内表单控件的双层蓝色焦点框改为单层细蓝色边框，同时不影响后台其他页面。

## 根因

Admin 全局样式对 Element Plus 的输入框、下拉框和文本域焦点状态增加了主色内边框及两像素 focus shadow。`TaskEditorDialog.vue` 使用 `append-to-body`，弹窗继承该全局规则，因此聚焦时出现截图中的双层蓝框。

## 实现

- 在 `TaskEditorDialog.vue` 的 scoped 样式中使用 `:global(.task-editor-dialog ...)` 精确覆盖弹窗内容。
- 覆盖普通输入框、文本域、下拉框、数字输入、日期输入和范围输入的聚焦状态。
- 聚焦时只保留一像素主色内边框，不显示额外的外层 focus shadow。
- 内部原生 input/textarea 的 `outline` 和 `box-shadow` 清零，避免浏览器额外绘制框线。
- 不修改 `admin/src/styles/admin.css`，不影响其他 Admin 页面。

## 验证

- 扩展定时任务静态检查，要求存在局部焦点覆盖并禁止在全局样式中处理该弹窗。
- 运行 `npm run check:scheduled-task`、目标文件 TypeScript/构建检查。
- 在浏览器打开创建定时任务弹窗，聚焦输入框、文本域和下拉框，确认没有蓝框且布局不变化。

