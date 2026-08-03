# DingTalk H5 SSO Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add DingTalk H5 password-free login so users who open the app inside DingTalk are authenticated through DingTalk identity and then continue through the existing WeCheckin token, menu, API permission, and data-scope system.

**Architecture:** The frontend obtains a temporary DingTalk auth code through `dd.runtime.permission.requestAuthCode(corpId)` and posts it to a new unauthenticated backend endpoint. The backend exchanges the code for a DingTalk user identity, maps that identity to `users.user_mini_openid`, creates the same Redis-backed `dingtalk_h5` session token used by password login, and returns the existing bootstrap payload. The implementation must not auto-create privileged users and must not bypass current DingTalk H5 route permission checks.

**Tech Stack:** Go/Hertz/GORM/Redis backend, DingTalk Open API over `net/http`, uni-app/Vue 3 H5 frontend, Element Plus admin frontend.

---

## File Structure

- Modify `backend/internal/bootstrap/seed_setup.go`
  - Add backend-managed DingTalk H5 app configuration keys.
- Modify `backend/internal/app/handler/admin/dingtalk/handler.go`
  - Extend the existing "钉钉应用管理 / 配置选项" endpoint to save `corpId`, `appKey`, and `appSecret`.
- Modify `admin/src/views/dingtalk-setup/index.vue`
  - Add fields for CorpId, AppKey, and AppSecret; never require appSecret to be re-entered unless it is being changed.
- Create `backend/internal/app/service/dingtalkh5/dingtalk_oapi.go`
  - Encapsulate DingTalk access token and auth-code exchange calls behind a small interface.
- Create `backend/internal/app/service/dingtalkh5/sso.go`
  - Add `LoginByAuthCodeContext` and share token issuance with password login.
- Modify `backend/internal/app/service/dingtalkh5/auth.go`
  - Extract shared login response/session creation into a helper used by both password login and SSO login.
- Modify `backend/internal/app/handler/client/dingtalkh5/handler.go`
  - Add `SSOLogin` handler.
- Modify `backend/cmd/routes_v2_dingtalk.go`
  - Register `POST /api/v2/dingtalk/h5/sso-login` outside the authenticated route groups.
- Modify `dingtalk-h5/services/dingtalkH5Api.js`
  - Add `ssoLogin({ authCode })`.
- Modify `dingtalk-h5/pages/index/index.vue`
  - Attempt DingTalk SSO before showing the password login form.
- Modify `dingtalk-h5/components/performance/LoginView.vue`
  - Update copy so password login is described as a fallback.
- Modify `dingtalk-h5/README.md` and `docs/DINGTALK_H5_PERFORMANCE.md`
  - Document the SSO flow, config keys, and user binding rule.

---

### Task 1: Extend DingTalk App Configuration

**Files:**
- Modify: `backend/internal/bootstrap/seed_setup.go`
- Modify: `backend/internal/app/handler/admin/dingtalk/handler.go`
- Modify: `admin/src/views/dingtalk-setup/index.vue`

- [ ] **Step 1: Add setup keys**

Add these keys next to the existing DingTalk H5 session settings:

```go
{Key: "DINGTALK_H5_CORP_ID", Value: "", Type: "string"},
{Key: "DINGTALK_H5_APP_KEY", Value: "", Type: "string"},
{Key: "DINGTALK_H5_APP_SECRET", Value: "", Type: "password"},
```

- [ ] **Step 2: Extend admin settings response**

Update `GetSettings` so it returns:

```go
"corpId":       setupValue("DINGTALK_H5_CORP_ID"),
"appKey":       setupValue("DINGTALK_H5_APP_KEY"),
"appSecretSet": setupValue("DINGTALK_H5_APP_SECRET") != "",
```

Do not return the raw `appSecret`.

- [ ] **Step 3: Extend admin settings save**

Update `SaveSettings` to persist:

```go
{key: "DINGTALK_H5_CORP_ID", value: strings.TrimSpace(c.PostForm("corpId")), typ: "string"},
{key: "DINGTALK_H5_APP_KEY", value: strings.TrimSpace(c.PostForm("appKey")), typ: "string"},
```

Only write `DINGTALK_H5_APP_SECRET` when `strings.TrimSpace(c.PostForm("appSecret")) != ""`, so editing token settings does not erase the stored secret.

- [ ] **Step 4: Update admin UI**

Add form fields above token settings:

```vue
<el-form-item label="CorpId">
  <el-input v-model="form.corpId" placeholder="钉钉企业 CorpId" />
</el-form-item>
<el-form-item label="AppKey">
  <el-input v-model="form.appKey" placeholder="钉钉内部应用 AppKey" />
</el-form-item>
<el-form-item label="AppSecret">
  <el-input v-model="form.appSecret" type="password" show-password :placeholder="form.appSecretSet ? '已保存，留空表示不修改' : '请输入 AppSecret'" />
</el-form-item>
```

Include these fields in the save payload:

```ts
corpId: form.corpId.trim(),
appKey: form.appKey.trim(),
appSecret: form.appSecret.trim(),
```

- [ ] **Step 5: Verify admin config behavior**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/handler/admin/dingtalk ./backend/internal/bootstrap -count=1
```

Expected: tests pass, and no endpoint returns raw `DINGTALK_H5_APP_SECRET`.

---

### Task 2: Add DingTalk Open API Client

**Files:**
- Create: `backend/internal/app/service/dingtalkh5/dingtalk_oapi.go`
- Test: `backend/internal/app/service/dingtalkh5/sso_test.go`

- [ ] **Step 1: Define client interface and identity DTO**

Create:

```go
type DingTalkUserIdentity struct {
	UserID  string
	UnionID string
	Name    string
}

type DingTalkIdentityClient interface {
	ExchangeAuthCodeContext(ctx context.Context, authCode string) (DingTalkUserIdentity, error)
}
```

- [ ] **Step 2: Implement HTTP client**

Implement a default client that:

1. Loads `DINGTALK_H5_APP_KEY` and `DINGTALK_H5_APP_SECRET` from setup.
2. Calls DingTalk access-token API.
3. Calls DingTalk user-info-by-auth-code API.
4. Returns `DingTalkUserIdentity{UserID: result.UserID, UnionID: result.UnionID, Name: result.Name}`.

All DingTalk HTTP code must live in this one file so future API version changes are isolated.

- [ ] **Step 3: Add deterministic tests with a fake client**

In `sso_test.go`, define:

```go
type fakeDingTalkIdentityClient struct {
	identity DingTalkUserIdentity
	err      error
}

func (f fakeDingTalkIdentityClient) ExchangeAuthCodeContext(ctx context.Context, authCode string) (DingTalkUserIdentity, error) {
	return f.identity, f.err
}
```

Use this fake for SSO service tests so CI never calls DingTalk.

---

### Task 3: Add Backend SSO Login Service

**Files:**
- Modify: `backend/internal/app/service/dingtalkh5/auth.go`
- Create: `backend/internal/app/service/dingtalkh5/sso.go`
- Test: `backend/internal/app/service/dingtalkh5/sso_test.go`

- [ ] **Step 1: Extract shared session issuance**

In `auth.go`, extract the token/session/bootstrap portion of `LoginContext` into:

```go
func issueDingTalkH5LoginResponseContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, addIP, device string) (*LoginResponse, error)
```

It should perform the same work currently done at the end of password login:

```go
token := randutil.HexString(32)
safeDevice := sessionDevice(device)
if err := onlineservice.StoreDingTalkH5SessionContext(ctx, onlineUserFromDingTalkH5User(user), token, addIP, safeDevice); err != nil {
	return nil, err
}
bootstrap := bootstrapForUserDB(ctx, db, user)
return &LoginResponse{
	Token:                 token,
	UserInfo:              bootstrap.User,
	Menus:                 bootstrap.Menus,
	ButtonPermissionKeys:  bootstrap.ButtonPermissionKeys,
	ButtonPermissionReady: bootstrap.ButtonPermissionReady,
	APIPermissionKeys:     bootstrap.APIPermissionKeys,
	APIPermissionReady:    bootstrap.APIPermissionReady,
	PermissionVersion:     bootstrap.PermissionVersion,
}, nil
```

- [ ] **Step 2: Implement SSO login**

Create:

```go
func LoginByAuthCodeContext(ctx context.Context, authCode, addIP, device string) (*LoginResponse, error) {
	return loginByAuthCodeContext(ctx, defaultDingTalkIdentityClient{}, authCode, addIP, device)
}
```

The internal function must:

1. Trim and validate `authCode`.
2. Exchange it for DingTalk identity.
3. Reject empty `identity.UserID`.
4. Map `NormalizeUserID(identity.UserID)` to `users.user_mini_openid`.
5. Require `user_status = 1`.
6. Hydrate department/position metadata through `hydratePerfUserWithUserDeptDB`.
7. Call `issueDingTalkH5LoginResponseContext`.

Use this error text for an unbound DingTalk user:

```go
return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
```

- [ ] **Step 3: Keep SSO permission-safe**

Do not create users automatically. Do not assign roles automatically. Do not synthesize menu/API permissions. SSO only replaces password verification; authorization still comes from current role/user permission grants.

- [ ] **Step 4: Add service tests**

Add tests for:

```text
empty authCode -> "免登授权码不能为空"
fake client returns error -> that error is surfaced
fake client returns empty UserID -> "钉钉身份异常"
unknown DingTalk UserID -> "钉钉账号未开通绩效系统，请联系管理员"
known active user -> returns token and bootstrap permissions
disabled user -> login rejected
```

- [ ] **Step 5: Verify service package**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/dingtalkh5 -count=1
```

Expected: all DingTalk H5 service tests pass without network access.

---

### Task 4: Register Backend SSO Endpoint

**Files:**
- Modify: `backend/internal/app/handler/client/dingtalkh5/handler.go`
- Modify: `backend/cmd/routes_v2_dingtalk.go`
- Test: `backend/cmd/routes_v2_dingtalk_test.go`

- [ ] **Step 1: Add handler method**

Add:

```go
func (h *Handler) SSOLogin(ctx context.Context, c *app.RequestContext) {
	var req struct {
		AuthCode string `json:"authCode"`
	}
	_ = c.BindAndValidate(&req)
	if req.AuthCode == "" {
		req.AuthCode = c.PostForm("authCode")
	}
	data, err := dingtalkh5service.LoginByAuthCodeContext(ctx, req.AuthCode, c.ClientIP(), string(c.UserAgent()))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}
```

- [ ] **Step 2: Register route**

In `registerV2DingTalkH5Routes`, add next to password login:

```go
group.POST("/sso-login", handler.SSOLogin)
```

This route must stay outside `handler.Auth()` and `handler.ApiPerm()`.

- [ ] **Step 3: Add route structure assertion**

In `routes_v2_dingtalk_test.go`, assert the source contains:

```go
`group.POST("/sso-login", handler.SSOLogin)`
```

and that this line appears before:

```go
`authed := group.Group("", handler.Auth())`
```

- [ ] **Step 4: Verify route package**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/cmd -run TestV2DingTalk -count=1
```

Expected: route structure tests pass.

---

### Task 5: Add Frontend Auto Login

**Files:**
- Modify: `dingtalk-h5/services/dingtalkH5Api.js`
- Modify: `dingtalk-h5/pages/index/index.vue`
- Modify: `dingtalk-h5/components/performance/LoginView.vue`

- [ ] **Step 1: Add SSO API wrapper**

Add to `dingTalkAuthApi`:

```js
ssoLogin(data) {
  return post(`${DINGTALK_H5_API}/sso-login`, data)
}
```

- [ ] **Step 2: Import DingTalk runtime helpers**

In `index.vue`, import:

```js
import { isDingTalkRuntime, requestAuthCode } from '../../utils/dingtalk'
```

- [ ] **Step 3: Implement `tryDingTalkAutoLogin`**

Add:

```js
const autoLoginTried = ref(false)

async function tryDingTalkAutoLogin() {
  if (autoLoginTried.value || authToken()) return false
  autoLoginTried.value = true
  if (!isDingTalkRuntime()) return false
  loading.value = true
  try {
    const authCode = await requestAuthCode()
    if (!authCode) return false
    const res = await dingTalkAuthApi.ssoLogin({ authCode })
    const payload = res.data || {}
    sessionAccessDenied.value = false
    setAuthToken(payload.token)
    applySessionAuthPayload(payload)
    if (!payloadHasSessionPermissions(payload)) {
      const bootstrapped = await loadBootstrapSafely()
      if (!bootstrapped) return false
    }
    await refreshDataSafely({ contentLoading: true })
    return true
  } catch (error) {
    if (isPermissionDeniedError(error)) {
      sessionAccessDenied.value = true
      sessionAccessDeniedMessage.value = error?.msg || '无权限访问，请联系管理员配置钉钉 H5 权限'
    }
    return false
  } finally {
    loading.value = false
  }
}
```

- [ ] **Step 4: Call SSO before showing login**

Update `onMounted`:

```js
onMounted(async () => {
  uni.$on(AUTH_EXPIRED_EVENT, resetSessionState)
  if (!authToken()) {
    await tryDingTalkAutoLogin()
    ready.value = true
    return
  }
  const bootstrapped = await loadBootstrapSafely()
  if (bootstrapped) {
    await refreshDataSafely({ contentLoading: true })
  }
  ready.value = true
})
```

- [ ] **Step 5: Update login copy**

Change the login description to:

```vue
<text class="login-desc">钉钉内打开会自动登录；无法免登时可使用绩效系统账号进入。</text>
```

- [ ] **Step 6: Verify H5 scaffold**

Run:

```bash
cd dingtalk-h5 && npm run check:scaffold
```

Expected: scaffold checks pass.

---

### Task 6: Update Documentation

**Files:**
- Modify: `dingtalk-h5/README.md`
- Modify: `docs/DINGTALK_H5_PERFORMANCE.md`

- [ ] **Step 1: Document login flow**

Add:

```markdown
## 钉钉免登流程

钉钉内打开 H5 时，前端通过 `requestAuthCode(corpId)` 获取一次性免登授权码，然后调用 `/api/v2/dingtalk/h5/sso-login`。后端使用钉钉 AppKey/AppSecret 换取钉钉用户身份，并将 DingTalk UserID 映射到 `users.user_mini_openid`。映射成功后后端签发系统自己的 `dingtalk_h5` Redis token，并返回用户、菜单、按钮权限、接口权限和权限版本。

免登只替代密码校验，不替代权限校验。用户仍必须在管理后台存在、状态启用，并配置角色权限或用户额外授权。
```

- [ ] **Step 2: Document config keys**

Add:

```markdown
管理后台“钉钉应用管理 / 配置选项”维护以下键：

- `DINGTALK_H5_CORP_ID`
- `DINGTALK_H5_APP_KEY`
- `DINGTALK_H5_APP_SECRET`
- `TOKEN_DINGTALK_H5_EXPIRE`
- `TOKEN_DINGTALK_H5_REDIS_PREFIX`
- `DINGTALK_H5_SINGLE_LOGIN`
```

- [ ] **Step 3: Document binding rule**

Add:

```markdown
当前最小实现使用 `users.user_mini_openid` 绑定 DingTalk UserID。若历史账号不是 DingTalk UserID，需要先在用户表中补齐绑定关系，否则免登会返回“钉钉账号未开通绩效系统，请联系管理员”。
```

---

### Task 7: End-to-End Verification

**Files:**
- No code changes.

- [ ] **Step 1: Backend verification**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/dingtalkh5 ./backend/internal/app/handler/client/dingtalkh5 ./backend/internal/app/handler/admin/dingtalk ./backend/cmd -count=1
```

Expected: all selected backend tests pass.

- [ ] **Step 2: Frontend verification**

Run:

```bash
cd dingtalk-h5 && npm run check:scaffold
```

Expected: scaffold check passes.

- [ ] **Step 3: Manual DingTalk verification**

In a DingTalk client:

1. Configure H5 app trusted domain and app home URL.
2. Configure `VITE_DINGTALK_CORP_ID` for frontend build.
3. Configure backend `DINGTALK_H5_CORP_ID`, `DINGTALK_H5_APP_KEY`, and `DINGTALK_H5_APP_SECRET`.
4. Bind a local user by setting `users.user_mini_openid` to the DingTalk UserID.
5. Grant that user DingTalk H5 menu and API permissions.
6. Open the H5 app from DingTalk workbench.

Expected: password page is skipped, app loads menu from bootstrap, and subsequent API requests carry the returned `DT_H5_TOKEN`.

---

## Residual Risks

- DingTalk UserID binding must be migrated carefully. If existing `user_mini_openid` values are demo names such as `nick`, SSO will not map until real DingTalk UserIDs are written.
- `VITE_DINGTALK_CORP_ID` and backend `DINGTALK_H5_CORP_ID` should match. Keeping both allows static frontend deployment, but mismatch will make JSAPI auth-code acquisition fail.
- DingTalk trusted domain and app homepage configuration are external deployment prerequisites.
- If DingTalk changes Open API versions, only `dingtalk_oapi.go` should need adjustment.
