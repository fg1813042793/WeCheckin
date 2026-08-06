# 钉钉H5微应用

该目录是独立的钉钉 H5 微应用前端工程，采用与 `frontend` 一致的 `uni-app + Vue 3 + Vite` 技术栈。目前已经接入绩效收集工作台，后端接口前缀为 `/api/v2/dingtalk/h5`。

## 目录结构

```text
dingtalk-h5/
  api/              按页面目录拆分的接口封装
  components/       应用级、鉴权级公共组件
  config/           环境配置读取
  pages/            uni-app 入口页面装配层
  router/           菜单页面映射和通知深链解析单入口
  scripts/          工程检查脚本
  views/            按一级菜单/二级菜单组织的业务页面
  utils/            请求与钉钉 JSAPI 工具
  App.vue
  main.js
  manifest.json
  pages.json
  vite.config.js
```

## 本地开发

```bash
npm install
npm run dev:h5
```

默认 H5 端口为 `8086`。本地开发默认使用同源 `/api/v2` 请求，并通过 Vite 代理转发到后端 `8083`。如需覆盖环境变量，可先复制 `.env.example`。

```bash
cp .env.example .env
```

## 环境变量

```text
VITE_API_BASE_URL=
VITE_DEV_PROXY_TARGET=http://localhost:8083
VITE_DINGTALK_CORP_ID=
```

- `VITE_API_BASE_URL`：后端 API 地址。留空时保持 `/api/v2/...` 同源路径，推荐用于 H5 本地开发和单点部署；独立 API 域名部署时填写完整地址。
- `VITE_DEV_PROXY_TARGET`：Vite 本地代理目标，默认后端地址为 `http://localhost:8083`。
- `VITE_DINGTALK_CORP_ID`：默认钉钉企业 CorpId，用于免登授权码。多企业部署时，URL 查询参数 `corpId` 优先级更高。

## Docker 单独部署

钉钉 H5 可以作为独立静态站点部署，后端由外部服务提供：

```bash
cd dingtalk-h5
cp .env.docker.example .env
docker compose -f docker-compose.h5.yml up -d --build
```

默认会把 H5 映射到宿主机 `8086`，容器内 Nginx 会把同源 `/api/v2/`、`/uploads/` 和 `/upload/` 代理到 `NGINX_API_PROXY_TARGET`，默认是 `http://host.docker.internal:8083`。生产环境通常需要修改：

```text
DINGTALK_H5_HTTP_PORT=8086
NGINX_API_PROXY_TARGET=http://your-backend:8083
VITE_DINGTALK_CORP_ID=dingxxxx
```

如果使用独立 API 域名，也可以在构建前设置 `VITE_API_BASE_URL=https://api.example.com`，此时前端会直接请求该地址；否则推荐保持为空，由 Nginx 同源代理避免跨域。

## 钉钉免登流程

钉钉内打开 H5 时，前端先通过 `/api/v2/dingtalk/h5/public-config` 读取后台配置的默认企业 `corpId` 和页面品牌信息，再确定当前 `corpId`：URL 查询参数 `?corpId=xxx` 优先，其次使用后台默认企业，最后使用 `VITE_DINGTALK_CORP_ID`。随后前端通过 `requestAuthCode(corpId)` 获取一次性免登授权码，然后调用 `/api/v2/dingtalk/h5/sso-login`，请求体包含 `corpId` 和 `authCode`。

后端按 `corpId` 查询 `dingtalk_h5_corp_configs`，使用该企业对应的 AppKey/AppSecret 换取钉钉用户身份。拿到 DingTalk UserID 后，后端通过 `dingtalk_h5_user_bindings.corp_id + dingtalk_user_id` 找到本地 `users.id`。映射成功后后端签发系统自己的 `dingtalk_h5` Redis token，并返回用户、菜单、按钮权限、接口权限和权限版本。

如果钉钉身份未绑定本地用户，且管理后台开启 `DINGTALK_H5_SELF_BIND_ENABLED`，`/sso-login` 会返回 `code=10020` 和一次性 `bindTicket`。前端展示“绑定系统账号”页，用户输入本地系统账号和密码后调用 `/api/v2/dingtalk/h5/bind-self`。后端校验钉钉票据、账号密码、用户状态和绑定唯一性后写入 `dingtalk_h5_user_bindings`，再签发正常登录 token。绑定成功后，下次免登会直接使用已有绑定。

免登只替代密码校验，不替代权限校验。用户仍必须在管理后台存在、状态启用，并配置角色权限或用户额外授权。

管理后台“钉钉应用管理 / 配置选项”维护企业应用列表、登录态配置和 H5 展示配置。企业应用列表写入 `dingtalk_h5_corp_configs`；每个企业应用可单独配置 CorpId、AppKey、AppSecret、旧版 AgentId、新版 App ID、H5 应用地址、通知方式和绩效流程通知开关。第一条配置会同步到以下旧 setup 键，供单企业旧部署和兼容逻辑读取：

- `DINGTALK_H5_CORP_ID`
- `DINGTALK_H5_APP_KEY`
- `DINGTALK_H5_APP_SECRET`
- `DINGTALK_H5_AGENT_ID`
- `DINGTALK_H5_UNIFIED_APP_ID`
- `DINGTALK_H5_NOTIFY_MODE`
- `DINGTALK_H5_ROBOT_CODE`
- `DINGTALK_H5_APP_URL`
- `TOKEN_DINGTALK_H5_EXPIRE`
- `TOKEN_DINGTALK_H5_REDIS_PREFIX`
- `DINGTALK_H5_SINGLE_LOGIN`
- `DINGTALK_H5_SELF_BIND_ENABLED`

迁移脚本 `backend/migrations/20260731123000_add_dingtalk_h5_multi_corp.sql` 会把旧 setup 单企业配置回填到 `dingtalk_h5_corp_configs`，并将旧 `users.user_mini_openid` 回填为该企业的 `dingtalk_h5_user_bindings.dingtalk_user_id`。若历史账号不是 DingTalk UserID，可以由管理员在绑定表中补齐关系，也可以开启首次自助绑定后由用户在钉钉内输入系统账号密码完成绑定。

多 CorpId 部署时，每个企业在钉钉工作台配置自己的 H5 地址，例如：

```text
https://oa.example.com/dingtalk-h5/?corpId=ding123
https://oa.example.com/dingtalk-h5/?corpId=ding456
```

## 已内置能力

- 独立 H5 页面入口：`pages/index/index`
- 账号密码登录和退出
- 工作台、我的绩效、上级审批、HRBP 评价
- 绩效单自评、上级评价、HRBP 评价、员工确认、归档
- 退回员工、退回上级、退回 HRBP、撤销提交
- 汇总筛选和导出
- 组织架构人员维护
- 模板展示和账号密码修改
- PC 端三栏布局，手机端单栏卡片布局
- 钉钉 JSAPI 检测与初始化：`utils/dingtalk.js`
- 钉钉免登授权码获取封装：`requestAuthCode`
- 客户端请求封装：`utils/request.js`
- `/api/v2/dingtalk/h5` 接口常量：`api/common/base.js`
- 页面接口封装按 `api/一级菜单/二级菜单` 拆分，与 `views` 目录保持同层级语义
- 菜单页面映射和通知深链解析集中在 `router/index.js`
- 绩效表单 tab 白名单集中在 `views/performance/common/reviewTabs.js`
- 应用内容出口：`components/app/AppContentOutlet.js`
- 脚手架完整性检查：`npm run check:scaffold`

## 菜单图标

钉钉 H5 菜单由登录或 bootstrap 返回的 `menus` 渲染，菜单名称和图标分别来自后端统一权限表的 `permissions.permission_name`、`permissions.permission_icon`。管理后台可在“权限管理 / 钉钉 H5”里编辑目录和菜单图标。

当前 H5 内置图标键为：

```text
dashboard, performance, mine, history, manager, hrbp, summary, org, template, account
```

实际 SVG 路径维护在 `components/app/AppShell.vue` 的 `navIconMap` 中。新增图标键时，需要同步后台可选项、后端菜单默认声明和 `navIconMap`。

## 开发规范

- 所有删除按钮都必须增加二次确认弹窗。用户确认后才可以执行删除接口或修改本地删除状态；用户取消时不能发起删除请求，也不能改变页面数据。
- 所有钉钉 H5 删除接口都必须走软删除，禁止物理删除业务数据。考评单使用 `deleted_at/delete_by/delete_dept_id` 标记删除并在列表、详情、统计、导出中统一过滤；共享 `users` 表的人员删除只允许停用 `user_status=0`。
- 钉钉工作通知点击跳转必须走后端生成的内部应用链接。后端会优先使用企业应用配置中的 H5 应用地址和新版 App ID，打开对应钉钉应用，例如“钉米-OKR”；前端只负责识别 `view`、`reviewNo`、`period`、`status` 深链参数，不在页面内手写通知跳转协议。
- 钉钉通知卡片的来源名称由后端读取“钉钉应用管理 / 配置 / 应用名称”后写入机器人链接消息；如果钉钉客户端仍按链接协议显示 `dingtalkclient`，以前端深链参数为准，业务页面仍会进入对应考评单。

## 后端接口

接口统一使用当前项目后端新增的独立模块：

```text
/api/v2/dingtalk/h5
```

登录接口开放，其余接口使用该模块签发的 token，通过 `Authorization` 请求头传递。导出接口为了浏览器直接下载，也支持 `token` 查询参数。

钉钉免登使用独立开放接口 `/api/v2/dingtalk/h5/sso-login`，首次自助绑定使用 `/api/v2/dingtalk/h5/bind-self`，账号密码登录仍使用 `/api/v2/dingtalk/h5/login` 作为非钉钉环境或免登失败时的兜底入口。
