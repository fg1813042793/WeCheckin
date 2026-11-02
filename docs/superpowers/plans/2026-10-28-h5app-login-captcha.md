# H5App Login Captcha Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an independently configurable image captcha to H5App account/password login without changing DingTalk SSO, session restore, or self-binding.

**Architecture:** Extract the existing Admin captcha engine into `internal/support/logincaptcha`, then create thin Admin and DingTalk H5 wrappers with distinct setup keys and Redis prefixes. Expose a public H5 captcha endpoint, verify the challenge only in the password-login handler path, and render the challenge only when the password-login panel is visible.

**Tech Stack:** Go 1.24, Hertz, Redis, GORM Setup storage, base64Captcha, Vue 3, uni-app, Pinia, uView Pro, Element Plus, TypeScript.

**Execution note:** The checkout already contains unrelated and prior-task changes. Do not reset, clean, commit, or push; use focused diffs and verification instead of commit checkpoints.

---

### Task 1: Extract the shared captcha engine

**Files:**
- Create: `backend/internal/support/logincaptcha/captcha_test.go`
- Create: `backend/internal/support/logincaptcha/captcha.go`
- Modify: `backend/internal/service/admin/adminauth/captcha.go`
- Modify: `backend/internal/service/admin/adminauth/captcha_test.go`

- [x] **Step 1: Write failing shared-engine tests**

Move the behavioral expectations from the current Admin-only tests into the new support package. Cover disabled mode, challenge generation, case-insensitive verification, one-time consumption, incorrect-answer consumption, generator/store failure, and PNG data shape.

```go
func TestServiceGeneratesAndConsumesCaseInsensitiveAnswer(t *testing.T) {
    store := &captchaStoreStub{}
    service := NewService(
        func(context.Context) (bool, error) { return true, nil },
        &captchaGeneratorStub{id: "captcha-1", image: "data:image/png;base64,abc", answer: "A2BcD"},
        store,
        2*time.Minute,
    )
    challenge, err := service.Challenge(context.Background())
    // Assert the challenge and that replay fails.
}
```

- [x] **Step 2: Run the shared package test and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/logincaptcha -count=1
```

Expected: FAIL because the shared package implementation does not exist.

- [x] **Step 3: Implement the shared engine**

Create exported reusable boundaries:

```go
type Challenge struct {
    Enabled   bool   `json:"enabled"`
    ID        string `json:"captchaId,omitempty"`
    ImageData string `json:"imageData,omitempty"`
    ExpiresIn int    `json:"expiresIn,omitempty"`
}

type Generator interface {
    Generate() (id, imageData, answer string, err error)
}

type Store interface {
    Put(context.Context, string, string, time.Duration) error
    Consume(context.Context, string) (string, error)
}

func NewService(enabled func(context.Context) (bool, error), generator Generator, store Store, ttl time.Duration) *Service
func NewImageGenerator() Generator
func NewRedisStore(prefix string) Store
```

Keep the existing cryptographic ID/answer generation, digest comparison, Lua consume operation, two-minute TTL, five-character source, and fail-closed errors.

- [x] **Step 4: Replace Admin implementation with a thin wrapper**

Preserve the public Admin API while wiring the shared service with:

```go
const AdminLoginCaptchaSetupKey = "ADMIN_LOGIN_CAPTCHA_ENABLED"
const adminCaptchaRedisPrefix = "admin:login:captcha:"

type CaptchaChallenge = logincaptcha.Challenge
var ErrCaptchaInvalid = logincaptcha.ErrInvalid
var ErrCaptchaUnavailable = logincaptcha.ErrUnavailable
```

- [x] **Step 5: Run shared and Admin auth tests and verify GREEN**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/logincaptcha ./internal/service/admin/adminauth -count=1
```

Expected: PASS.

### Task 2: Add DingTalk H5 captcha service and route contracts

**Files:**
- Create: `backend/internal/service/dingtalkh5/auth/captcha.go`
- Create: `backend/internal/service/dingtalkh5/auth/captcha_test.go`
- Create: `backend/internal/routes/v2/dingtalkh5/captcha_route_test.go`
- Modify: `backend/internal/handler/dingtalkh5/auth/handler.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/routes.go`

- [x] **Step 1: Write failing H5 service and route tests**

Tests must assert:

```go
const DingTalkH5LoginCaptchaSetupKey = "DINGTALK_H5_LOGIN_CAPTCHA_ENABLED"
const dingTalkH5CaptchaRedisPrefix = "dingtalk_h5:login:captcha:"
```

The route test must require a public `group.GET("/captcha", handler.Auth.Captcha)` registration and prove `/sso-login` and `/bind-self` remain bound directly to their existing handlers.

- [x] **Step 2: Run tests and verify RED**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/service/dingtalkh5/auth ./internal/routes/v2/dingtalkh5 -count=1
```

Expected: FAIL because the H5 captcha service and route do not exist.

- [x] **Step 3: Implement H5 captcha wrapper**

Use the shared engine with the H5 setup key and Redis prefix. Export:

```go
func LoginCaptchaEnabledContext(ctx context.Context) (bool, error)
func GenerateCaptchaContext(ctx context.Context) (logincaptcha.Challenge, error)
func VerifyCaptchaContext(ctx context.Context, id, answer string) error
func LoginWithCaptchaContext(ctx context.Context, name, password, captchaID, captchaCode, addIP, device string) (*LoginResponse, error)
```

- [x] **Step 4: Add the public handler and password-login fields**

Add a no-store captcha handler. Extend only the password login request with `captchaId` and `captchaCode`, map shared sentinel errors to safe Chinese messages, and call `LoginWithCaptchaContext`. Do not edit the SSO or bind request DTOs.

- [x] **Step 5: Register the public route and verify GREEN**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/service/dingtalkh5/auth ./internal/handler/dingtalkh5/auth ./internal/routes/v2/dingtalkh5 -count=1
```

Expected: PASS.

### Task 3: Add the DingTalk login captcha setting

**Files:**
- Create: `backend/migrations/20261028170000_add_dingtalk_h5_login_captcha_setup.sql`
- Modify: `backend/internal/service/admin/dingtalk/dto.go`
- Modify: `backend/internal/service/admin/dingtalk/dto_test.go`
- Modify: `backend/internal/handler/admin/dingtalk/handler.go`
- Modify: `backend/internal/handler/admin/dingtalk/settings_test.go`
- Modify: `admin/src/api/types.ts`
- Modify: `admin/src/views/dingtalk-setup/index.vue`

- [x] **Step 1: Write failing backend settings contract tests**

Require the stable JSON field `"loginCaptchaEnabled":0`, setup key `DINGTALK_H5_LOGIN_CAPTCHA_ENABLED`, login-scope save mapping, and Admin page field/payload markers.

- [x] **Step 2: Run settings tests and verify RED**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/service/admin/dingtalk ./internal/handler/admin/dingtalk -count=1
```

Expected: FAIL because the new setting is absent.

- [x] **Step 3: Implement backend read/write and migration**

Add `LoginCaptchaEnabled int` to `SettingsResponse`, read it through `LoginCaptchaEnabledContext`, and append a switch setup item during `scope=login` saves. The migration inserts default `0` with `ON DUPLICATE KEY UPDATE setup_key = VALUES(setup_key)` so existing values are preserved.

- [x] **Step 4: Add the Admin switch**

Extend `DingTalkSettings`, reactive form loading, and login-scope payload with `loginCaptchaEnabled`. Render an Element Plus switch labelled “账号密码登录图片验证码” near self-binding.

- [x] **Step 5: Run backend settings tests and Admin build checks**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/service/admin/dingtalk ./internal/handler/admin/dingtalk -count=1
cd ../admin
npm run build
```

Expected: PASS, except any pre-existing repository baseline failure must be recorded separately with its exact output.

### Task 4: Add H5App captcha API, state, and login UI

**Files:**
- Modify: `h5app/scripts/check-dingtalk-module.mjs`
- Modify: `h5app/src/types/dingtalk-h5.ts`
- Modify: `h5app/src/api/dingtalk-h5/auth.ts`
- Modify: `h5app/src/stores/dingtalkAuth.ts`
- Modify: `h5app/src/components/dingtalk-login-panel/dingtalk-login-panel.vue`
- Modify: `h5app/src/locale/lang/zh-CN.json`
- Modify: `h5app/src/locale/lang/en-US.json`

- [x] **Step 1: Extend the DingTalk module checker and verify RED**

Require a typed `CaptchaChallenge`, a `captcha()` API call, store challenge state/refresh action, password-login captcha fields, and login-panel image/input/refresh markers. Also assert that `ssoLogin` and `bindSelf` payloads do not contain captcha fields.

Run:

```bash
cd h5app
pnpm check:dingtalk-module
```

Expected: FAIL on the missing captcha contract.

- [x] **Step 2: Add API and type contracts**

```ts
export interface LoginCaptchaChallenge {
  enabled: boolean
  captchaId?: string
  imageData?: string
  expiresIn?: number
}

export interface LoginRequest {
  name: string
  password: string
  captchaId?: string
  captchaCode?: string
}
```

Expose `captcha()` from the H5 auth API.

- [x] **Step 3: Add store challenge state**

Store `captchaEnabled`, `captchaId`, `captchaImage`, `captchaLoading`, and a recoverable error message. `refreshLoginCaptcha()` updates them from the endpoint and fails closed when the request fails. `loginWithPassword()` refreshes the challenge after unsuccessful requests while leaving SSO and binding functions unchanged.

- [x] **Step 4: Render the login challenge**

Pass challenge props and refresh event through `app-auth-guard`. Use `u-input` and a labelled, directly clickable uni-app `image` with stable dimensions and responsive flex layout; do not render a separate refresh button. Keep the empty image placeholder clickable for retry, ignore repeated clicks while loading, validate the captcha locally only when enabled, and emit uppercase trimmed code.

- [x] **Step 5: Add synchronized locale keys and verify GREEN**

```bash
cd h5app
pnpm check:dingtalk-module
pnpm lint
pnpm type-check
pnpm check:ui-style
pnpm build:h5
```

Expected: PASS.

### Task 5: Synchronize Swagger and run final regression checks

**Files:**
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Regenerate: `backend/docs/swagger/docs.go`
- Regenerate: `backend/docs/swagger/swagger.json`
- Regenerate: `backend/docs/swagger/swagger.yaml`

- [x] **Step 1: Add Swagger declarations for the H5 captcha and login fields**

Declare `GET /api/v2/dingtalk/h5/captcha` and the optional `captchaId`/`captchaCode` password-login fields without changing SSO or bind declarations.

- [x] **Step 2: Regenerate Swagger**

```bash
cd backend
swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger --templateDelims '{%,%}'
```

- [x] **Step 3: Run focused and full verification**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/logincaptcha ./internal/service/admin/adminauth ./internal/service/dingtalkh5/auth ./internal/handler/dingtalkh5/auth ./internal/routes/v2/dingtalkh5 ./internal/service/admin/dingtalk ./internal/handler/admin/dingtalk -count=1
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
cd ../admin
npm run check:all
cd ../h5app
pnpm lint && pnpm type-check && pnpm check:dingtalk-module && pnpm check:ui-style && pnpm build:h5
```

- [x] **Step 4: Review the final diff and browser behavior**

Confirm the diff changes no SSO/bind DTOs or handlers. Start the H5 dev server and verify mobile and desktop password-login layouts, refresh behavior, service-unavailable state, and absence of captcha requests during DingTalk auto-login. Record any environment-dependent behavior that cannot be verified locally.

