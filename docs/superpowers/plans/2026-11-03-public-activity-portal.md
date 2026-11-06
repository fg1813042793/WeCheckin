# Public Activity Portal Implementation Plan

**Goal:** 实现后台配置的单一公共问卷/考试入口，支持匿名浏览、二维码下载和原填写入口跳转。

**Architecture:** 独立 `publicportal` 模块保存页面配置和展示项，通过类型与 ID 引用问卷/考试，匿名响应严格使用专用 DTO。Admin 增加独立管理页和独立公共路由，二维码由本地前端编码库生成，不调用第三方二维码服务。

**Tech Stack:** Go/Hertz/GORM/MySQL、Vue 3/TypeScript/Element Plus；复用 Admin UI 和公开填写路由。

## 固定契约

- 管理页 `/public-portal`；公共页 `/public/activities`；已保存内容预览 `/public/activities?preview=1`。
- `GET /api/v2/public/activities` 返回 `{enabled,title,description,logoUrl,items:PublicItem[]}`，不返回配置、题目、答案、密码或授权数据。页面关闭时返回停用状态和空列表。
- `GET /api/v2/admin/public-portal` 返回 `{settings,items:AdminItem[]}`。
- `PUT /api/v2/admin/public-portal/settings`：`{enabled,title,description,logoUrl,showEnded,version}`。
- `GET /api/v2/admin/public-portal/options?kind=survey|exam&keyword=&page=1&pageSize=20`：`{list,total,page,pageSize}`。
- `POST /api/v2/admin/public-portal/items`：`{kind,sourceId,title,description,enabled,sort}`。
- `PUT /api/v2/admin/public-portal/items/:id`：同创建字段，加 `version`；`DELETE .../:id?version=...` 只移除展示配置。
- `GET /api/v2/admin/public-portal/preview` 返回同公共 DTO，只绕过页面总开关，仍校验展示项与源公开状态。
- Settings：`enabled:boolean,title:string,description:string,logoUrl:string,showEnded:boolean,version:number`。
- PublicItem：`id:number,kind:'survey'|'exam',sourceId:number,title:string,managementTitle:string,description:string,startTime:number,endTime:number,state:'upcoming'|'active'|'ended',loginRequired:boolean,passwordRequired:boolean,url:string`。
- AdminItem：创建字段及 `id,version,managementTitle,displayTitle,available:boolean,unavailableReason:string,canManage:boolean`。
- SourceOption：`kind,sourceId,managementTitle,displayTitle,startTime,endTime,loginRequired,passwordRequired`；仅已发布且 visibility=0，并在当前管理员的数据范围内。
- 请求/响应沿用 `{code,msg,data}`。保存冲突返回“配置已更新，请刷新后重试”；缺迁移返回可操作提示，不启动自动建表。
- 页面配置默认关闭，不需要配置域名即可启用或预览。Logo 允许站内绝对路径或 HTTP(S) URL，不允许脚本协议。
- 每项最多标题255字、简介1000字；页面标题100字、简介1000字；最多200展示项，来源类型/ID唯一，排序0..99999。
- 名称优先展示项覆盖名，再源 `settings.displayTitle`，最后源 title。说明同样支持覆盖。仅从 settings 提取 displayTitle/fillTemplate/loginRequired/password 标志，绝不公开原 JSON。
- 问卷链接 sf/sf1，考试 ef/ef1；后端仅返回允许模板的根相对路径；前端按当前页面 origin 生成链接和二维码，不使用请求 Host 或源自定义外链。
- 停用/删除/改私有的来源自动隐藏；未开始可展示和下载二维码；结束默认隐藏，可保留已结束卡片。参与限制仍由原填写接口执行。
- 管理员候选和保存均校验源数据权限；不向无权管理该来源的管理员返回源管理名称，并标记不可管理。来源删除后，其展示记录允许创建者或超级管理员清理。
- 菜单 `admin:menu:public-portal`，按钮 `admin:menu:public-portal:edit`；API `public-portal:list`、`public-portal:edit`。新增权限定义与迁移，不自动扩大普通角色授权。

## 工作划分

- [x] 后端模块：`backend/internal/modules/publicportal/{application,infrastructure,transport/httpadmin,transport/httppublic}`；model、存储、来源范围、校验、匿名过滤、并发版本和行为测试。
- [x] 前端：`admin/src/api/publicPortal.ts`、`admin/src/types/publicPortal.ts`、`admin/src/views/public-portal/`、`admin/src/views/public/Activities.vue`；候选选择、管理操作、预览、响应式卡片、本地 PNG 编码和下载。
- [x] 主线程接入：Admin/public 路由、权限目录、中间件映射、审计归类、SQL 迁移与 Swagger；不修改源问卷/考试的权限和提交逻辑。
- [x] 回归：不公开内容无法进入匿名 DTO；改私有/停用立即失效；输入越界、无权限、跨范围、版本冲突、重复来源；匿名页无后台请求；二维码内容可解码且下载有效。
- [x] 验证：相关 Go 包测试、Backend `go test ./...`、Admin `npm run check:all` 与构建、桌面/手机浏览器完整操作。只使用隔离测试数据，不修改生产配置或发送通知。
- [x] 文档：记录迁移、权限授权、自动链接、预览/上线方式及实际验证结果。保留所有既有未提交修改，不提交、推送或部署。

## 验证结果

- 公共入口模块、路由、权限及 Swagger 定向 Go 测试通过；模块 race 测试通过；迁移保护测试通过。
- Admin `check:public-portal`、`check:admin-ui-contract`、TypeScript 和 `npm run build` 通过。
- 浏览器以隔离模拟数据验证匿名请求无后台 token、筛选/空态、实际 PNG 下载与二维码解码、手机布局/放大、后台保存刷新、新增/下架/移除、关闭页面预览、只读权限和错误重试。
- Backend 全量仍有既有考试复制 Swagger、钉钉配置检查、审计分类、生产配置检查及 HTTP 错误边界失败；Admin `check:all` 仍停在既有用户页面 1867/1859 行预算检查。
- 未执行生产迁移、未部署；真实 MySQL 迁移与运行环境联调需在部署流程中完成。

## 2026-11-03 调整

按用户要求移除“公共域名”输入及校验。Settings 不再返回域名；旧入参仅作兼容并忽略，数据库旧列保留不再参与链接生成。后台公共链接使用当前 origin，公开条目相对填写路径由浏览器解析到同一 origin；不将 API base 或旧配置作为二维码域名。

取消域名调整验证：公共入口 Go 模块、`npm run check:public-portal` 和浏览器保存/预览/二维码解码通过；全量仍有范围外失败，当前 Admin 构建被富文本组件类型错误阻断。

按用户补充要求，公开卡片与二维码预览显示来源管理名称及类型/编号，用于区分相同显示名称；管理名称不取配置快照，实时读取已上架且公开来源。
