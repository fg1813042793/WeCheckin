# 钉钉H5微应用

该目录是独立的钉钉 H5 微应用前端工程，采用与 `frontend` 一致的 `uni-app + Vue 3 + Vite` 技术栈。目前已经接入绩效收集工作台，后端接口前缀为 `/api/v2/dingtalk/h5`。

## 目录结构

```text
dingtalk-h5/
  config/           环境配置读取
  pages/            页面
  scripts/          工程检查脚本
  services/         `/api/v2/dingtalk/h5` 接口封装
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
- `VITE_DINGTALK_CORP_ID`：钉钉企业 CorpId，用于免登授权码。

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
- `/api/v2/dingtalk/h5` 接口常量：`services/dingtalkH5Api.js`
- 脚手架完整性检查：`npm run check:scaffold`

## 后端接口

接口统一使用当前项目后端新增的独立模块：

```text
/api/v2/dingtalk/h5
```

登录接口开放，其余接口使用该模块签发的 token，通过 `Authorization` 请求头传递。导出接口为了浏览器直接下载，也支持 `token` 查询参数。

后续接入钉钉开放平台免登时，可以复用当前登录返回结构，在 `/api/v2/dingtalk/h5/login` 中增加 `authCode` 换 token 逻辑。
