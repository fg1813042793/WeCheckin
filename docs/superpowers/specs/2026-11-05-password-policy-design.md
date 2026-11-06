# 用户密码策略设计

## 目标

在系统配置的登录配置中增加统一密码策略，覆盖后台管理员和普通用户，并在 Admin、客户端和钉钉 H5 中一致执行密码强度、有效期、历史密码及首次登录强制改密规则。

## 范围

- 后台管理员账号和普通用户账号共用 `users` 表，因此共用一套密码状态和历史记录。
- 账号密码登录、钉钉免密登录及其他已认证入口都执行强制改密状态检查。
- 升级后密码策略默认关闭，配置页面预填标准模式参数；管理员主动开启后生效。
- 不改变现有密码哈希兼容逻辑，历史 MD5 密码继续支持登录并在成功验证后升级为 bcrypt。

## 配置模型

新增系统配置键 `PASSWORD_POLICY_CONFIG`，保存带版本号的 JSON：

```json
{
  "version": 1,
  "enabled": false,
  "minLength": 8,
  "requireLetter": true,
  "requireNumber": true,
  "requireUppercase": false,
  "requireLowercase": false,
  "requireSpecial": false,
  "maxAgeDays": 90,
  "expiryWarningDays": 7,
  "historyCount": 5,
  "forceChangeOnCreate": true,
  "forceChangeOnReset": true,
  "allowWeakTemporaryPassword": true,
  "effectiveAt": 0
}
```

约束：最小长度为 6 至 64；有效期为 0 至 3650 天，0 表示不启用过期；提醒天数不得大于有效期；历史数量为 0 至 20。

首次从关闭切换到开启时由后端写入 `effectiveAt`。关闭后重新开启时重新生成 `effectiveAt`，避免现有账号因历史时间立即过期。

## 数据模型

通过版本化 SQL 迁移为 `users` 增加：

- `user_password_changed_at BIGINT NOT NULL DEFAULT 0`
- `user_password_must_change TINYINT NOT NULL DEFAULT 0`

新增 `user_password_histories`：

- `id`
- `user_id`
- `password_hash`
- `changed_at`
- `change_source`：`self_change`、`admin_reset`、`create`、`migration`
- `created_at`、`updated_at`

按 `(user_id, changed_at)` 建立索引。历史表只保存哈希，不保存明文。每次写入后最多保留 20 条，策略只读取最近 `historyCount` 条。

## 策略服务

新增独立密码策略服务，负责：

- 读取、校验和缓存 `PASSWORD_POLICY_CONFIG`。
- 校验永久密码的长度和字符组成。
- 使用 `passwordutil.Verify` 检查当前密码及最近 N 个历史密码，兼容 bcrypt 和历史 MD5。
- 判断首次登录强制改密、密码过期、到期提醒和过期时间。
- 在事务内记录旧密码哈希、更新新密码、更新时间和强制改密标记。
- 管理员创建或重置时允许临时密码绕过强度规则，但不绕过历史重复检查，并强制设置 `user_password_must_change = 1`。

所有密码写入入口必须调用该服务，不允许各模块自行哈希并直接更新 `user_password`。登录时仅执行状态判断，不在请求日志中记录密码、哈希或历史数据。

## 密码状态

登录响应统一增加：

```json
{
  "passwordChangeRequired": true,
  "passwordExpired": false,
  "passwordExpiresAt": 0,
  "passwordExpiryWarning": false
}
```

判断规则：

- `user_password_must_change = 1` 时必须修改。
- 有效期启用后，基准时间取 `user_password_changed_at`；值为 0 时取策略 `effectiveAt`。
- 当前时间超过基准时间加有效期时视为过期并强制修改。
- 进入提醒窗口但尚未过期时只提示，不限制业务操作。

## 受限会话

密码验证或免密身份验证成功后仍签发 Token，但会话保存 `passwordChangeRequired` 状态。Admin、客户端和钉钉 H5 的认证中间件在状态为真时只允许：

- 当前账号修改密码接口。
- 退出登录或客户端清理会话所需接口。
- 获取当前密码策略摘要和当前账号基础信息所需的只读接口。

访问其他接口返回稳定业务错误码 `PASSWORD_CHANGE_REQUIRED`，不返回内部实现细节。

修改密码成功后：

- 当前会话更新为正常状态或签发新 Token。
- 同账号其他历史会话失效。
- 客户端重新加载权限和账号信息。

管理员重置其他账号密码时立即使该账号所有现有会话失效。

## API

新增管理端配置接口：

- `GET /api/v2/admin/password-policy`
- `PUT /api/v2/admin/password-policy`

复用系统配置查看和编辑权限，后端执行完整字段校验，不允许前端直接写任意 JSON。

密码修改接口：

- Admin 复用 `PATCH /api/v2/admin/me/password`。
- 钉钉 H5 复用 `PATCH /api/v2/dingtalk/h5/account/password`。
- 客户端新增 `PATCH /api/v2/me/password`。
- 管理员创建、编辑和重置管理员/用户密码的现有接口改为调用统一策略服务。

接口、DTO、Swagger、权限目录和所有调用端同步更新。

## Admin 页面

系统配置的“登录配置”中增加“密码策略”分区：

- 总开关。
- 最小长度输入。
- 字母、数字、大写、小写、特殊字符要求复选框。
- 密码有效期、提前提醒天数、历史密码数量。
- 新账号首次登录强制修改。
- 管理员重置后强制修改。
- 允许弱临时密码。

保存时一次提交完整策略，并展示字段级校验错误。

Admin 增加强制修改密码独立页面。路由守卫读取登录响应状态；受限账号不能进入后台布局、菜单或业务页面。到期提醒在登录后以可关闭提醒展示。

## Frontend 与 H5App

- `frontend` 增加普通用户改密页面和受限路由逻辑。
- `h5app` 复用现有账号密码修改表单，强制改密时以不可关闭页面或遮罩展示。
- 钉钉免密登录同样执行强制改密判断。
- 三端统一展示密码规则摘要和后端错误，不在前端单独维护另一套规则。
- 修改成功后替换当前 Token 或重新初始化认证状态。

## 兼容与迁移

- 策略默认关闭，不改变升级前登录行为。
- 现有账号不设置首次登录强制改密。
- 首次启用策略时使用 `effectiveAt` 作为旧密码有效期起点。
- 密码强度只在新建、重置和自主修改时校验，不反向判定现有密码强度。
- 无密码账号保持现状；首次设置密码时按新账号规则处理。
- 自动创建和模板导入使用的临时密码可绕过强度，但必须标记强制修改。

## 审计与安全

- 策略变更记录配置前后差异，但不记录密码或哈希。
- 密码修改、管理员重置、历史密码拒绝、强制改密完成使用明确审计动作和原因码。
- 登录错误仍保持统一“账号或密码错误”，不泄露账号是否存在、密码是否过期或历史策略信息。
- 密码策略和历史数据不通过普通用户详情接口返回。

## 验证

- 策略解析、边界校验、强度组合和默认关闭行为单元测试。
- 密码历史、临时密码、首次登录、重置、过期、提醒和会话失效测试。
- Admin、客户端、钉钉 H5 登录及中间件受限访问测试。
- 迁移幂等、索引和历史数据兼容测试。
- Admin 配置页面、强制改密页面及三端浏览器/目标端验证。
- Backend、Admin、Frontend 和 H5App 按模块规范运行对应完整检查。

