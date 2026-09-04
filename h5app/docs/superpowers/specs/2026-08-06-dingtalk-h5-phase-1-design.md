# 钉钉 H5 第一阶段迁移设计

## 背景

当前仓库是 `uni-app + Vue 3 + TypeScript + uView Pro + Pinia` 模板；旧项目 `/Users/fanggang/项目文件/编程学习/WeCheckin/dingtalk-h5` 是独立钉钉 H5 微应用，已实现绩效管理业务。迁移目标是在当前仓库中重构第一阶段核心闭环，同时不修改旧项目，并复用旧项目已验证的后端接口。

## 第一阶段目标

第一阶段先交付可运行的钉钉 H5 核心闭环：

- 入口页：`/pages/dingtalk/index`。
- 认证：公共配置、账号密码登录、钉钉免登、自助绑定、退出登录。
- 会话数据：bootstrap 加载用户、菜单、按钮权限、接口权限、应用配置。
- 导航：按后端菜单权限渲染工作台和绩效菜单。
- 工作台：展示待办绩效单，并能进入对应绩效详情。
- 我的绩效：展示当前用户绩效单、创建考评单入口、刷新、详情表单基础展示。
- 评审入口：上级评审、HRBP 评审先提供列表与详情基础能力。
- 接口复用：统一使用 `/api/v2/dingtalk/h5` 及旧项目 API 语义。

不在第一阶段完整重写汇总导出、流程配置、模板配置和个人中心的全部复杂编辑体验；这些页面在第一阶段提供占位或基础列表入口，后续分阶段迁移。

## 架构

新增钉钉模块按当前项目规范组织：

- `src/config/dingtalk-h5.ts`：环境配置和 token key。
- `src/types/dingtalk-h5.ts`：认证、菜单、权限、绩效单、用户、模板等共享类型。
- `src/utils/dingtalk.ts`：钉钉 JSAPI 检测、免登授权码、导航标题封装。
- `src/api/dingtalk-h5/`：按业务域拆分接口，保留旧项目接口路径。
- `src/stores/dingtalkAuth.ts`：认证、token、用户、菜单、权限和应用配置。
- `src/stores/performance.ts`：绩效单、用户目录、模板、待办、当前选中绩效单。
- `src/pages/dingtalk/`：钉钉入口页和局部组件。

HTTP 层复用当前项目的 `uview-pro` `http` 插件，但扩展拦截器以支持钉钉 H5：

- 基础地址来自 `VITE_API_BASE_URL`，为空时保持同源 `/api/v2/...`。
- token 使用 `DT_H5_TOKEN` 存储。
- 请求头增加 `Authorization` 和 `X-Client-Platform: dingtalk-h5`。
- 业务响应 `code` 缺省或为 `0` 视为成功。
- `code=10020` 作为自助绑定分支抛出，不弹通用错误。

## 页面与交互

入口页使用 `app-page` 和 uView Pro 组件。页面状态分为加载中、绑定账号、登录、无权限、已登录应用壳五种。应用壳包含：

- 顶部应用标题、用户名称、刷新和退出。
- 横向菜单 tabs，使用后端授权菜单过滤。
- 内容区按当前菜单展示：工作台、我的绩效、上级评审、HRBP 评审。
- 详情区使用统一绩效详情组件，展示流程状态、基础字段和按权限展示动作按钮。

第一阶段不追求旧项目 PC 三栏布局 1:1 复刻，优先保持移动端 H5 可用和当前 uView Pro 视觉一致。桌面宽屏使用响应式单栏加卡片布局。

## 数据流

1. 页面加载时先请求 `public-config`，合并应用标题、名称、logo 和默认 `corpId`。
2. 如果已有 token，则调用 `bootstrap` 恢复会话。
3. 如果在钉钉环境且能取得 `corpId`，调用钉钉 JSAPI 获取授权码并请求 `sso-login`。
4. 如果免登返回 `code=10020`，展示绑定账号表单并调用 `bind-self`。
5. 非钉钉或免登失败时展示账号密码登录。
6. 登录成功后加载绩效单列表；工作台从绩效单中计算待办。
7. 点击待办或列表项更新选中绩效单并切换到对应视图。

## 权限

菜单权限来自后端 `menus`。按钮权限和接口权限保存在认证 store 中：

- `hasButtonPermission(key)` 用于创建、删除、导出等按钮。
- `hasApiPermission(key)` 为后续接口级控制保留。
- 菜单过滤只展示当前前端已注册且后端授权的页面。

第一阶段保留旧项目核心按钮 key，例如：

- `dingtalk_h5:button:review:create`
- `dingtalk_h5:button:review:delete`
- `dingtalk_h5:button:review:export`

## 错误处理

- 登录过期、未登录、账号异常会清理 token 并回到登录态。
- 无权限访问展示无权限状态。
- 绑定需要错误进入绑定态，不弹通用失败 toast。
- 删除类动作必须通过 `u-modal` 或 `uni.showModal` 二次确认。
- 导出等 H5 浏览器能力使用条件编译包裹。

## 验证

第一阶段验证包含：

- `pnpm check:dingtalk-module`：检查钉钉模块关键文件、路由、接口路径和脚本。
- `pnpm lint`
- `pnpm type-check`
- `pnpm build:h5`

如构建发现旧模板示例代码已有问题，本阶段只修与钉钉模块直接相关的问题。
