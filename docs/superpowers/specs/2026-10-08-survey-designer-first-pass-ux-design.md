# 问卷设计器第一批体验优化设计

## 背景

问卷设计器已经完成按职责拆分，并具备撤销重做、保存状态、问题定位、大纲搜索和响应式属性面板等基础能力。当前第一批优化聚焦高频体验，不再为了减少行数继续拆分组件。

## 目标

1. 为导航和工具栏中的纯图标按钮补齐可访问名称与选中状态。
2. 在窄屏 PC 下保留问卷标题和保存状态，而不是整体隐藏状态区。
3. 未选择题目时释放右侧属性面板空间，并允许用户主动关闭属性面板。
4. 在小于 768px 的移动端明确提示使用桌面端编辑，避免展示不可操作的 900px 横向画布。

## 设计

### 图标按钮

- 左侧导航按钮增加 `type="button"`、`aria-label` 和适用的 `aria-pressed`。
- 撤销、重做、快捷键、文本导入、导出、编辑/JSON/预览模式和问卷设置按钮增加 `aria-label`。
- 编辑、JSON、预览和问卷设置按钮通过 `aria-pressed` 暴露当前状态。

### 窄屏状态

- 桌面宽度继续使用居中的完整标题、题目数、模式和保存状态。
- `<=1280px` 时隐藏完整状态区，在工具栏第二行显示紧凑标题和保存状态。
- 工具栏允许换行，操作按钮仍保持第一行可用。

### 属性面板

- 只有选中题目或打开问卷设置时才显示右侧属性面板。
- 工具栏新增“问卷设置面板”按钮，作为属性面板关闭后的明确入口。
- 属性面板标题增加关闭按钮；关闭时清空选中题目、选项索引和问卷设置状态。
- `768px–1599px` 时属性面板使用覆盖层，并在工具栏下方开始显示，避免挤压画布或遮挡保存按钮。
- 本次不实现拖动调宽，避免同时改变布局状态持久化契约。

### 移动端

- `<768px` 时隐藏设计器工作区，显示“建议在桌面端编辑问卷”的说明和返回列表按钮。
- 不改变桌面端、窄屏 PC 和问卷填报页面行为。

## 范围

- `admin/src/views/survey/SurveyDesigner.vue`
- `admin/src/views/survey/survey-designer.css`
- `admin/src/views/survey/survey-designer-responsive.css`
- `admin/src/views/survey/components/SurveyDesignerPanelTitle.vue`
- `admin/src/views/survey/components/SurveyDesignerQuestionProperties.vue`
- `admin/scripts/check-survey-designer-ux.mjs`

## 验证

1. 先增加静态契约并确认失败。
2. 实现后运行 `npm run check:survey-designer-ux`。
3. 运行 `npm run build` 和 `npm run check:bundle`。
4. 浏览器验证 1280px、1024px、768px 和 767px：按钮可访问名称、紧凑状态、属性面板关闭/重开、移动端提示均符合设计。
