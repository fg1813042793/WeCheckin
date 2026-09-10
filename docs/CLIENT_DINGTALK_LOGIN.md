# 客户端钉钉验证登录

本文说明 `frontend/` 客户端的钉钉登录。Android App 使用钉钉 Native Auth SDK，H5 使用网页 OAuth；两者最终都把一次性 authCode 交给同一后端接口。该能力与 `h5app/` 钉钉工作台免登相互独立。

## Android Native 登录流程

1. 客户端调用 `GET /api/v2/auth/dingtalk-config` 获取已启用企业；多个企业时由用户选择。
2. 客户端向 `POST /api/v2/auth/dingtalk-authorization` 提交 `corpId` 和 Native SDK `redirectUri`，后端返回 `clientId`、`redirectUri` 和随机 `state`，AppSecret 不返回客户端。
3. `wecheckin-dingtalk-auth` UTS 插件使用 `com.alibaba.android:ddopenauth:1.5.0.8` 调用 `IDDAuthApi.authLogin()`，由 SDK 拉起钉钉。
4. 用户授权后，钉钉按实际 APK 包名启动 `<applicationId>.ddauth.DDAuthActivity`，例如当前基座的 `uni.app.UNIF691FB0.ddauth.DDAuthActivity`，通过 Intent extras 返回 `authCode`、`state` 和 `error`。
5. 回调 Activity 将结果写入应用私有存储并立即结束；登录页回到前台后一次性读取结果，校验 10 分钟内的 `state`。
6. 客户端调用 `POST /api/v2/auth/dingtalk-login`。后端使用对应企业的 AppKey、AppSecret 换取钉钉身份，以 `corpId + unionId` 查找绑定并签发普通客户端 Token。

系统不会根据钉钉身份自动创建用户，也不允许客户端自助绑定。未绑定、绑定停用、本地用户停用或同一企业存在重复 UnionId 绑定时均拒绝登录。

## H5 登录流程

H5 继续使用 `https://login.dingtalk.com/oauth2/auth` 网页 OAuth。授权完成后回到 `VITE_DINGTALK_OAUTH_REDIRECT_URI` 或当前 H5 页面，再调用相同的 `/api/v2/auth/dingtalk-login`。H5 不加载 Android UTS 插件。

## 后台与开放平台配置

管理后台：

1. 在“钉钉应用管理 / 企业应用”中维护 CorpId、AppKey 和 AppSecret 并启用应用。
2. 在“钉钉用户绑定管理”中选择同一企业和本地用户，填写钉钉 UserId 与 UnionId。客户端登录必须有 UnionId。

钉钉开放平台当前应用：

1. 在“钉钉登录与分享 / 接入登录”中配置与 `VITE_DINGTALK_NATIVE_REDIRECT_URI` 一致的回调域名。
2. 启用 Android 分享/登录配置，包名填写最终 APK 的实际包名。当前已制作基座的包名是 `uni.app.UNIF691FB0`；如果重新打包后使用 `com.wecheckin.app`，则填写 `com.wecheckin.app`。
3. 填写制作自定义基座所使用证书的应用签名。包名或签名不一致时，SDK 无法把授权结果回调到 App。
4. 申请客户端登录所需的用户基本信息权限。

## Frontend 配置

```dotenv
VITE_API_BASE_URL=http://192.168.50.6:8083

# 可选；留空时 Native SDK 使用 VITE_API_BASE_URL 进行回调域名校验。
VITE_DINGTALK_NATIVE_REDIRECT_URI=http://192.168.50.6:8083

# 仅 H5 网页 OAuth 使用。
VITE_DINGTALK_OAUTH_REDIRECT_URI=
```

开发环境可以使用钉钉后台已登记且手机可访问的 HTTP 地址，生产环境应使用 HTTPS。

## 自定义基座

HBuilder 标准基座包名为 `io.dcloud.HBuilder`，不包含项目的 SDK、`DDAuthActivity`、包名和签名配置，不能用于测试 Native Auth。

在 HBuilderX 中：

1. 选择“运行 / 运行到手机或模拟器 / 制作自定义调试基座”。
2. Android 包名使用最终基座的实际包名，选择固定调试证书后制作基座。
3. 将该证书签名填写到钉钉开放平台，并重新发布钉钉登录配置。
4. 运行项目时勾选“使用自定义基座”，不要选择标准基座。
5. 包名、证书、SDK、Manifest 或原生插件变化后需要重新制作基座；普通 Vue/JS 改动可以继续热更新。

正式包如果开启 Android 代码混淆，离线打包工程需保留钉钉授权 SDK 类：

```proguard
-keep class com.android.dingtalk.openauth.**{*;}
```

## HBuilderX 本地编译环境

如果运行时出现 `存在三方依赖或资源引用` 、`找不到名称“android”` 或 `DDAuthConstant` 错误，说明 HBuilderX 还没有加载 Android UTS 扩展和 Maven 依赖，不是业务代码问题。请在“设置 / 运行配置”中配置：

- 安装并启用“UTS 开发扩展 - Android”；
- Gradle `7.5` 以上且低于 `9.0`；
- Gradle JDK `17`；
- Android SDK 的 `build-tools` 不低于 `30.0.0`，并存在 `platforms;android-30` 或更高版本。

也可以直接在 HBuilderX 中制作云端自定义调试基座；标准基座不包含钉钉 SDK 和 `DDAuthActivity`。

如果已经重新打过自定义基座但仍然提示 `DDAuthConstant` 或 `com.android.dingtalk` 找不到：

1. 在 HBuilderX 执行“运行 / 清理缓存”；
2. 关闭 HBuilderX，删除项目下的 `unpackage` 缓存目录后重新打开项目；
3. 重新制作自定义调试基座，并在运行时重新选择该基座。

## 安全边界

- AppSecret 和钉钉用户 access token 只在 Backend 使用。
- Native SDK authCode 只能使用一次；客户端在提交前校验 state，并在读取后删除原生回调结果。
- `/api/v2/auth/*` 返回 `Cache-Control: private, no-store`。
- 账号密码登录保留为备用入口；公开用户 ID 登录接口已移除。
- Android Native 与 H5 OAuth 共用用户绑定和服务端 Token 逻辑，但不共用前端授权入口。
