# WeCheckin 接口鉴权清单

核对日期：2026-11-05（Asia/Shanghai）。依据当前工作区的实际注册路由、中间件及重点处理函数；不是仅按 Swagger 的 Security 标记推断。没有读取生产凭证、调用生产接口或修改鉴权逻辑。生产部署版本及外围网关规则需以运行环境为准。

本清单区分三件事：账号登录认证（Auth）、独立角色/API 权限（Perm）、业务条件或专用凭证。没有挂统一 Auth/Perm，不代表不用密码、授权码、答题 session 或下载 token。

## 汇总

| 范围 | 数量 | 含义 |
| --- | ---: | --- |
| V2 客户端/公开入口 | 25 | 路由没有统一 Auth/Perm，部分在处理函数内校验访问条件或凭证 |
| V2 后台登录入口 | 2 | 验证码和登录；不要求已有后台 token |
| V2 钉钉 H5 登录入口 | 5 | 公共配置、验证码、登录、免登、自助绑定 |
| 旧版兼容入口 | 21 | 仍被当前根路由注册，未挂统一 Auth/Perm |
| 后台登录后免独立 API 权限 | 4 | 仍经过 AdminAuth / AdminPerm；必须满足登录和角色前置条件 |
| H5 登录后免独立 API 权限 | 9 | 经过 DingTalkH5Auth，不经过 DingTalkH5Perm；限当前账号业务 |

以上 32 条 V2 加 21 条兼容路由，共 53 条应用路由未挂统一 Auth/Perm。另有健康检查、通用上传、文档及静态资源，单列于后文。计数按 HTTP 方法 + 路径，不按路径前缀通配；不把 OPTIONS 预检算作业务接口。

## 登录后不需要单独分配 API 权限

### 管理后台：4 条

这四条在权限目录中的 required 值明确为空字符串。AdminPerm 在放行前仍检查登录用户及角色，普通后台账号没有角色会被拒绝。

| 方法 | 完整路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v2/admin/home` | 后台首页；权限映射为空字符串。 |
| POST | `/api/v2/admin/auth/logout` | 退出当前后台登录；权限映射为空字符串。 |
| GET | `/api/v2/admin/me/menus` | 读取本人菜单；权限映射为空字符串。 |
| GET | `/api/v2/admin/me/perms` | 读取本人权限编码；权限映射为空字符串。 |

依据：[后台权限映射](../backend/internal/middleware/admin/route_permissions.go)、[后台权限中间件](../backend/internal/middleware/admin/permission.go)。

### 钉钉 H5：9 条

以下接口使用单独的 authed 路由组，只要求有效 H5 登录态，不调用独立 API 权限中间件。本人通知的记录 ID 仍需通过归属校验。

| 方法 | 完整路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v2/dingtalk/h5/logout` | 退出当前 H5 登录。 |
| POST | `/api/v2/dingtalk/h5/account/avatar` | 上传本人头像。 |
| PATCH | `/api/v2/dingtalk/h5/account/profile` | 修改本人资料。 |
| PATCH | `/api/v2/dingtalk/h5/account/password` | 修改本人密码；还要校验当前密码。 |
| GET | `/api/v2/dingtalk/h5/notifications` | 读取本人通知。 |
| GET | `/api/v2/dingtalk/h5/notifications/unread-count` | 本人未读通知数量。 |
| PATCH | `/api/v2/dingtalk/h5/notifications/read-all` | 将本人通知标记为已读。 |
| PATCH | `/api/v2/dingtalk/h5/notifications/:id/read` | 标记本人某条通知已读，服务端校验归属。 |
| DELETE | `/api/v2/dingtalk/h5/notifications/:id` | 删除本人某条通知，服务端校验归属。 |

依据：[H5 路由](../backend/internal/routes/v2/dingtalkh5/routes.go)、[账号处理函数](../backend/internal/handler/dingtalkh5/account/handler.go)、[通知处理函数](../backend/internal/handler/dingtalkh5/notification/handler.go)。

普通客户端 /api/v2 的 authenticated 路由组同时使用 ClientAuth 和 ClientPerm，没有类似的“登录后免独立 API 权限”子组；例如 /api/v2/me/logout 仍有 client:api:user:logout 权限要求。

## V2 未挂统一登录/权限中间件的入口

### 客户端/公开入口：25 条

| 方法 | 完整路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v2/public/activities` | 公共入口；检查入口开关、上架状态和来源公开状态。 |
| GET | `/api/v2/public/exam-record-exports/download` | 不需要账号登录，但必须有有效、未过期的下载 token；不属于无凭证下载。 |
| GET | `/api/v2/home` | 首页内容读取。 |
| GET | `/api/v2/home/setup` | 客户端首页配置读取。 |
| GET | `/api/v2/user-form-fields` | 用户扩展表单字段定义读取。 |
| GET | `/api/v2/auth/dingtalk-config` | 钉钉登录配置读取。 |
| POST | `/api/v2/auth/dingtalk-authorization` | 生成钉钉授权参数，不要求已有登录态。 |
| POST | `/api/v2/auth/dingtalk-login` | 凭钉钉授权码登录，仍校验授权码与身份。 |
| POST | `/api/v2/auth/password-login` | 账号密码登录，仍校验登录凭证。 |
| POST | `/api/v2/auth/register` | 注册入口，仍执行注册参数和业务校验。 |
| GET | `/api/v2/geo/reverse` | 逆地理编码服务。 |
| GET | `/api/v2/dict/types` | 公开字典类型读取。 |
| GET | `/api/v2/dict/items` | 公开字典项读取。 |
| GET | `/api/v2/events` | 活动列表读取。 |
| GET | `/api/v2/events/:id` | 活动详情读取。 |
| GET | `/api/v2/surveys` | 已发布问卷列表；列表查询未按 visibility 排除非公开项，详情另有访问检查。 |
| GET | `/api/v2/surveys/:id` | 处理函数按 visibility 校验登录及部门，并检查发布状态、开始/结束时间。 |
| POST | `/api/v2/surveys/:id/responses` | 提交答卷；settings.loginRequired 开启时要求用户 token，另有发布状态、时间和次数限制。 |
| POST | `/api/v2/survey/apply` | 对请求中的 schema / answers 应用计算与显隐逻辑。 |
| POST | `/api/v2/survey/validate` | 校验答案；按 surveyId 或请求 schema 工作，并可校验发布状态/次数。 |
| GET | `/api/v2/exams` | 已发布考试列表；列表查询未按 visibility 排除非公开项，详情另有访问检查。 |
| GET | `/api/v2/exams/:id` | 处理函数按 visibility / loginRequired 校验登录，并可检查部门、开放时间和次数。 |
| POST | `/api/v2/exams/:id/submissions` | Schema 提交分支要求有效 examId + session；recordId 分支另要求用户上下文并校验答卷归属/状态。 |
| POST | `/api/v2/exams/:id/validation` | 检查考试已发布、提交限制及答案，不走角色 API 权限。 |
| GET | `/api/v2/exam-results` | 凭 session 查询答卷结果；没有账号登录或独立角色权限校验。 |

依据：[V2 客户端路由](../backend/internal/routes/v2/client/routes.go)。这里不把问卷/考试详情和提交的业务校验概括为统一权限模型，详见后文。

### 后台登录：2 条

| 方法 | 完整路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v2/admin/auth/captcha` | 生成后台登录验证码。 |
| POST | `/api/v2/admin/auth/login` | 校验后台账号、密码及按配置启用的验证码。 |

### 钉钉 H5 登录：5 条

| 方法 | 完整路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v2/dingtalk/h5/public-config` | H5 公共登录配置。 |
| GET | `/api/v2/dingtalk/h5/captcha` | 生成 H5 登录验证码。 |
| POST | `/api/v2/dingtalk/h5/login` | 校验系统账号密码，验证码按配置启用。 |
| POST | `/api/v2/dingtalk/h5/sso-login` | 校验企业信息与钉钉免登授权码。 |
| POST | `/api/v2/dingtalk/h5/bind-self` | 必须开启首次自助绑定，并校验有效 bindTicket、系统账号及密码。 |

依据：[后台路由](../backend/internal/routes/v2/admin/routes.go)、[H5 路由](../backend/internal/routes/v2/dingtalkh5/routes.go)、[自助绑定服务](../backend/internal/service/dingtalkh5/auth/service.go)。

## 仍在注册的旧版入口：21 条

这些旧地址同样不能漏掉。多数复用 V2 的处理函数，因此也带有对应的业务条件；旧参数名可能不同。

| 方法 | 完整路径 | 对应 V2 入口或说明 |
| --- | --- | --- |
| GET | `/test/test` | 测试响应；没有对应 V2 别名。 |
| POST | `/survey/apply` | POST /api/v2/survey/apply |
| POST | `/survey/validate` | POST /api/v2/survey/validate |
| GET | `/home/setup_get` | GET /api/v2/home/setup |
| GET | `/home/list` | GET /api/v2/home |
| GET | `/user_form_fields` | GET /api/v2/user-form-fields |
| POST | `/passport/login_pwd` | POST /api/v2/auth/password-login |
| POST | `/passport/register` | POST /api/v2/auth/register |
| GET | `/geo/reverse` | GET /api/v2/geo/reverse |
| GET | `/dict/types` | GET /api/v2/dict/types |
| GET | `/dict/items` | GET /api/v2/dict/items |
| GET | `/event/list` | GET /api/v2/events |
| GET | `/event/view` | GET /api/v2/events/:id |
| GET | `/survey/list` | GET /api/v2/surveys |
| GET | `/survey/view` | GET /api/v2/surveys/:id |
| POST | `/survey/submit` | POST /api/v2/surveys/:id/responses |
| GET | `/exam/list` | GET /api/v2/exams |
| GET | `/exam/view` | GET /api/v2/exams/:id |
| POST | `/exam/submit` | POST /api/v2/exams/:id/submissions |
| POST | `/exam/validate` | POST /api/v2/exams/:id/validation |
| GET | `/exam/result` | GET /api/v2/exam-results |

依据：[根路由](../backend/internal/routes/routes.go)、[V1 公共路由](../backend/internal/routes/v1/client/public.go)、[V1 客户端路由](../backend/internal/routes/v1/client/routes.go)。`v1/admin/survey_exam.go` 虽保留旧后台注册函数，当前根注册链没有调用，不计作已启用接口。

## 健康检查、上传与静态内容

| 方法 | 路径 | 当前代码行为 |
| --- | --- | --- |
| GET | `/health` | 无登录/权限校验，存活检查 |
| GET | `/ready` | 无登录/权限校验，执行数据库就绪检查 |
| POST | `/upload` | **无登录/权限校验的写入接口**；允许指定图片/视频扩展名，单文件最大 20MB |
| GET | `/uploads/*filepath` | 无登录/权限校验的本地上传文件读取 |
| GET | `/swagger` | 公开跳转到 Swagger 页面 |
| GET | `/swagger/doc.json` | 公开读取接口文档 |
| GET | `/swagger/*any` | Swagger UI 和资源公开读取 |
| GET | `/*any` | 仅 `../admin/dist` 存在时注册，提供前端静态文件/SPA 回退；不是 API 免权限规则 |

依据：[健康检查](../backend/internal/routes/common/health.go)、[上传与静态路由](../backend/internal/routes/common/upload_static.go)、[Swagger 注册](../backend/internal/routes/common/swagger.go)。这些函数由根路由直接调用；不能只审查 /api/v2 前缀。

## 不能误认为免授权的接口

| 接口 | 实际要求 |
| --- | --- |
| `GET /api/v2/admin/permissions/tree` | 后台登录 + `admin:api:menu:list` |
| `GET /api/v2/admin/public-portal`、`/options`、`/preview` | 后台登录 + `admin:api:public-portal:list` |
| 公共页面设置、展示项增删改 | 后台登录 + `admin:api:public-portal:edit` |
| `PATCH /api/v2/admin/me/password` | 后台登录 + `admin:api:mgr:edit` |
| `GET /api/v2/dingtalk/h5/bootstrap` | H5 登录 + `dingtalk_h5:api:bootstrap:view` |
| `GET /api/v2/dingtalk/h5/workbench` | H5 登录 + `dingtalk_h5:api:workbench:view` |
| `GET /api/v2/me/bootstrap` | 客户端登录 + `client:api:bootstrap:view` |
| `POST /api/v2/uploads` | 客户端登录 + `client:api:upload:create`；与旧 `/upload` 的保护不同 |

后台、客户端和 H5 权限中间件对未声明的业务路由均采用拒绝访问，而不是“没有登记权限就自动放行”。后台保留超级管理员角色可跳过 API 权限判断，但仍须登录，这是账号级特例，不是公开接口。

依据：[后台映射](../backend/internal/middleware/admin/route_permissions.go)、[后台权限逻辑](../backend/internal/middleware/admin/permission.go)、[应用 API 权限目录](../backend/internal/support/appapiperm/catalog.go)。

## 业务鉴权需要单独区分的地方

1. `/upload` 当前的处理函数内也没有用户校验；仅有文件扩展名和大小检查。列入免登录清单不代表建议保持公开。此处没有执行上传测试或更改权限。
2. 考试答卷下载必须使用有效的下载 token；它不是账号 token。考试结果则用 session 定位答卷，不经过账号 Auth/Perm，当前查询链也没有按账号校验答卷归属。不能对外笼统描述为“无凭证接口”。
3. 问卷详情按 visibility 校验登录/部门，问卷提交主要按 settings.loginRequired 判断是否必须登录；考试详情按 visibility/loginRequired 判断，Schema 提交主要按 session 校验，recordId 提交分支还要求用户上下文。各入口的检查并不完全一致，不能假定详情页的限制自动覆盖提交接口。
4. 公开问卷/考试列表当前按发布状态查询，未按 visibility 排除非公开项；这与公共入口 /api/v2/public/activities 只展示上架且允许公开的来源不同。
5. 本次没有把前端填写密码、验证码页面或菜单隐藏当作后端角色权限。没有逐项确认的业务限制不作为放行依据。

处理链证据：[问卷详情](../backend/internal/handler/client/survey/browse.go)、[问卷提交](../backend/internal/handler/client/survey/submit.go)、[问卷提交服务](../backend/internal/service/client/survey/response_submit.go)、[考试详情](../backend/internal/handler/client/exam/browse.go)、[考试提交](../backend/internal/handler/client/exam/submit.go)、[考试结果](../backend/internal/handler/client/exam/record.go)、[考试 session 与限制](../backend/internal/service/client/exam/client.go)、[下载凭证校验](../backend/internal/modules/examrecordexportlink/transport/httppublic/handler.go)。

## 验证范围

路由清单由源码中的注册方法和显式免权限映射提取，再人工核对父路由组和处理函数。所有列出的 V2/兼容路由均来自当前根路由可达的注册链。未进行生产 HTTP 探测，反向代理、网络 ACL 及生产构建版本可能影响实际可达性。本次只新增本清单，不修改业务或鉴权代码。

验证命令（backend 目录）：

```sh
GOCACHE=$PWD/../.cache/go-build go test ./internal/middleware/admin ./internal/middleware/client ./internal/middleware/dingtalk_h5 ./internal/support/appapiperm ./internal/routes/v2/admin ./internal/routes/v2/dingtalkh5 ./internal/routes/v2/client ./internal/routes/common
```

上述 8 个包测试全部通过；同时核对了清单计数和每个方法/路径条目的完整性。未运行真实上传、提交答卷、发送通知或生产探测。
