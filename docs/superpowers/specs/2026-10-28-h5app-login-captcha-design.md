# H5App 账号密码登录图片验证码设计

## 目标

为 `h5app` 的账号密码登录增加可配置的图片验证码，降低密码暴力尝试风险。验证码仅作用于 `POST /api/v2/dingtalk/h5/login`，不得改变钉钉免登、SSO、会话恢复或首次自助绑定流程。

## 范围

- 在 Admin 的“钉钉应用配置 > 登录配置”中增加“账号密码登录图片验证码”开关。
- 新增配置键 `DINGTALK_H5_LOGIN_CAPTCHA_ENABLED`，默认关闭。
- 新增公共接口 `GET /api/v2/dingtalk/h5/captcha`。
- H5 账号密码登录请求增加 `captchaId` 和 `captchaCode`。
- H5 登录面板按开关展示验证码输入、图片和刷新操作。
- 复用 Admin 登录验证码的安全基础能力，但保持配置、Redis 数据和业务入口隔离。

以下流程不在本次改动范围内：

- `POST /api/v2/dingtalk/h5/sso-login`
- `POST /api/v2/dingtalk/h5/bind-self`
- 已登录会话恢复、退出登录和账号绑定界面

## 后端设计

### 共享验证码模块

将当前 Admin 登录验证码中的通用能力抽取到 `backend/internal/support/logincaptcha`：

- 使用安全随机数生成验证码 ID 和五位字符答案。
- 生成 Base64 PNG 图片。
- Redis 只保存答案摘要，不保存明文答案。
- 验证码有效期为两分钟。
- 使用 Redis Lua 脚本原子读取并删除，验证码无论答对或答错均只能使用一次。
- 答案忽略大小写，字符集排除容易混淆的字符。
- Redis 不可用或图片生成失败时返回“验证码服务暂不可用”，开关开启时不得绕过校验。

Admin 与 H5 分别提供轻量业务包装，传入各自的启用判断和 Redis 前缀：

- Admin：`ADMIN_LOGIN_CAPTCHA_ENABLED`、`admin:login:captcha:`
- H5：`DINGTALK_H5_LOGIN_CAPTCHA_ENABLED`、`dingtalk_h5:login:captcha:`

### H5 验证码接口

`GET /api/v2/dingtalk/h5/captcha` 返回稳定结构：

```json
{
  "enabled": true,
  "captchaId": "...",
  "imageData": "data:image/png;base64,...",
  "expiresIn": 120
}
```

开关关闭时返回 `{ "enabled": false }`。接口不依赖登录态，不进入钉钉 H5 权限中间件，并设置禁止缓存响应头。

### H5 登录校验

`POST /api/v2/dingtalk/h5/login` 接收：

```json
{
  "name": "account",
  "password": "password",
  "captchaId": "...",
  "captchaCode": "ABCDE"
}
```

处理顺序为：

1. 读取 H5 验证码开关。
2. 开关开启时验证并消费验证码。
3. 验证码通过后执行现有账号密码验证和会话签发。
4. 开关关闭时忽略验证码字段，保持现有兼容行为。

验证码错误或过期使用稳定、可展示的业务错误；内部 Redis、数据库和图片生成错误不向客户端泄露具体实现信息。

## Admin 配置设计

钉钉配置响应增加 `loginCaptchaEnabled` 数字开关字段。登录配置保存请求同步提交该字段，并以 `switch` 类型写入 Setup。

页面在“首次自助绑定”附近增加 Element Plus 开关，标签为“账号密码登录图片验证码”。该开关只说明和控制账号密码登录，不改变单点登录与自助绑定设置。

新增版本化 SQL 迁移插入默认值 `0`，重复执行时不覆盖管理员已经保存的值。

## H5App 交互设计

登录面板挂载时请求验证码接口：

- 返回 `enabled: false` 时只显示原有账号和密码字段。
- 返回 `enabled: true` 时显示验证码输入框和验证码图片，不额外显示刷新按钮。
- 点击验证码图片重新获取验证码；图片加载失败时可点击原图片占位区域重试，加载中忽略重复点击。
- 验证码开启时，账号、密码和验证码均填写后才能提交。
- 每次登录请求结束后，只要验证码仍开启，就获取新验证码，避免复用已消费的挑战。
- 验证码获取失败时禁用登录提交并展示可重试提示，不能静默跳过。

界面使用 uView Pro 组件、uni-app `image` 和主题变量；新增文案同步维护中文和英文语言包。移动端验证码行保持输入区域可用，桌面端保持现有登录卡片宽度，不增加新的页面或路由。

## 契约与兼容性

- `LoginRequest` 增加可选验证码字段，使开关关闭时仍兼容旧调用结构。
- 验证码接口仅在账号密码登录面板实际展示后调用，钉钉自动登录过程不会请求验证码。
- `sso-login` 和 `bind-self` 的请求 DTO、服务方法和前端组件不变。
- Admin 与 H5 使用不同配置键和 Redis 前缀，任何一端的开关和验证码都不能影响另一端。

## 测试与验证

后端测试覆盖：

- 共享验证码生成、大小写校验、一次性消费、过期和存储异常。
- Admin 与 H5 的配置键及 Redis 前缀隔离。
- H5 开关关闭时账号密码登录保持原行为。
- H5 开关开启时缺失、错误、重复使用和正确验证码的行为。
- `/captcha` 为公开路由，`/sso-login` 和 `/bind-self` 未接入验证码校验。
- 钉钉登录配置读写和迁移默认值。

前端验证覆盖：

- Admin 类型检查、配置载入、保存和构建。
- H5 验证码关闭、开启、刷新、错误后换图和服务不可用状态。
- `pnpm lint`、`pnpm type-check`、`pnpm check:ui-style`、`pnpm check:dingtalk-module` 和 `pnpm build:h5`。
- 浏览器验证手机宽度、PC 登录卡片，以及钉钉免登不出现验证码请求。

