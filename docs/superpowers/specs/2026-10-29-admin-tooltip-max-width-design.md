# Admin Tooltip 最大宽度设计

## 目标

统一管理后台所有 Element Plus Tooltip 的文本宽度和换行行为，避免 User-Agent、URL、错误信息和长标识符把提示层撑到接近整屏宽度。

## 适用范围

- 表格 `show-overflow-tooltip` 生成的字段提示。
- 页面中显式使用的 `el-tooltip`，包括图标、按钮、权限说明和字段说明。
- 不影响 `el-popover`、Select 下拉层、ECharts Tooltip 及其他业务浮层。

## 样式规则

- 短内容保持自然宽度，不设置固定宽度或最小宽度。
- 最大宽度为 `min(420px, calc(100vw - 32px))`。
- 超过最大宽度时允许正常换行。
- User-Agent、URL、ID 等连续长字符串使用 `overflow-wrap: anywhere`，不得产生横向溢出。
- 行高使用 `1.5`，保持多行内容可读。
- 不使用全局 `word-break: break-all`，避免普通中英文文本被不合理拆分。

## 实现边界

规则放入 Admin 全局样式，只针对 Element Plus Tooltip Popper 的稳定类名，不逐页修改现有 21 个使用 `show-overflow-tooltip` 的页面，也不要求业务页面增加 `popper-class`。

## 验证

- 增加静态契约检查，要求全局 Tooltip 最大宽度、换行及长字符串处理规则存在。
- 构建管理后台。
- 浏览器检查问卷回收的设备信息 Tooltip：宽度不超过 420px，User-Agent 自动换行。
- 浏览器检查一个短图标 Tooltip：仍按内容自然宽度展示。
- 在窄窗口验证 Tooltip 不超过视口宽度减 32px。
