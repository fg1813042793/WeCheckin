# Password Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a configurable password policy for administrators and users, including strength validation, password expiry, expiry warnings, password history, temporary passwords, first-login change, and server-enforced restricted sessions across Admin, Frontend, and H5App.

**Architecture:** Add credential-state columns and a password-history table, then centralize policy parsing and password mutation in `internal/service/accountsecurity`. Authentication services evaluate one shared password state and persist it in Redis sessions; middleware blocks non-password routes while a password change is required. Admin exposes a typed policy API and all three clients render the same backend-owned state.

**Tech Stack:** Go 1.24, Hertz, GORM, MySQL migrations, Redis sessions, Vue 3/Element Plus, uni-app Vue, TypeScript, uView Pro.

---

### Task 1: Persist password state and history

**Files:**
- Create: `backend/migrations/20261105100000_add_password_policy.sql`
- Create: `backend/migrations/password_policy_test.go`
- Create: `backend/internal/model/account/password_history.go`
- Modify: `backend/internal/model/account/user.go`
- Modify: `backend/internal/model/admin/admin.go`
- Modify: `backend/internal/model/aliases.go`

- [x] **Step 1: Write failing migration and model tests**

Assert the migration adds `user_password_changed_at`, `user_password_must_change`, creates `user_password_histories`, and declares an index beginning with `user_id, changed_at`. Assert both `model.User` and `model.Admin` expose `PasswordChangedAt` and `PasswordMustChange` mapped to the same columns.

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./migrations ./test/internal/model -run PasswordPolicy -count=1
```

Expected: FAIL because the migration and fields do not exist.

- [x] **Step 2: Add the migration and models**

Add these columns without modifying existing passwords:

```sql
ALTER TABLE `users`
  ADD COLUMN `user_password_changed_at` BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN `user_password_must_change` TINYINT NOT NULL DEFAULT 0;
```

Create `user_password_histories` with `user_id`, a 255-character hash, millisecond timestamps, `change_source`, and a composite index. Do not store plaintext or backfill existing hashes into history.

- [x] **Step 3: Run the migration/model tests**

Expected: PASS.

### Task 2: Build the central password policy service

**Files:**
- Create: `backend/internal/service/accountsecurity/policy.go`
- Create: `backend/internal/service/accountsecurity/policy_test.go`
- Create: `backend/internal/service/accountsecurity/password.go`
- Create: `backend/internal/service/accountsecurity/password_test.go`
- Create: `backend/internal/service/accountsecurity/state.go`
- Create: `backend/internal/service/accountsecurity/state_test.go`
- Modify: `backend/internal/service/admin/setup/service.go`

- [x] **Step 1: Write failing policy tests**

Cover defaults, disabled policy, field bounds, letter/number/uppercase/lowercase/special rules, `effectiveAt`, expiry, warning windows, and malformed JSON fallback.

Use this stable config contract:

```go
type Policy struct {
    Version                    int   `json:"version"`
    Enabled                    bool  `json:"enabled"`
    MinLength                  int   `json:"minLength"`
    RequireLetter              bool  `json:"requireLetter"`
    RequireNumber              bool  `json:"requireNumber"`
    RequireUppercase           bool  `json:"requireUppercase"`
    RequireLowercase           bool  `json:"requireLowercase"`
    RequireSpecial             bool  `json:"requireSpecial"`
    MaxAgeDays                 int   `json:"maxAgeDays"`
    ExpiryWarningDays          int   `json:"expiryWarningDays"`
    HistoryCount               int   `json:"historyCount"`
    ForceChangeOnCreate        bool  `json:"forceChangeOnCreate"`
    ForceChangeOnReset         bool  `json:"forceChangeOnReset"`
    AllowWeakTemporaryPassword bool  `json:"allowWeakTemporaryPassword"`
    EffectiveAt                int64 `json:"effectiveAt"`
}
```

- [x] **Step 2: Implement policy load/save and validation**

Use setup key `PASSWORD_POLICY_CONFIG`. Missing or invalid configuration returns the disabled standard defaults, never an enabled permissive policy. `SavePolicyContext` validates the typed structure and writes `effectiveAt` only when transitioning from disabled to enabled.

- [x] **Step 3: Write failing password/history tests**

Test permanent passwords, weak temporary passwords, current-password reuse, recent-history reuse through `passwordutil.Verify`, history pruning to 20 rows, and transactional rollback.

- [x] **Step 4: Implement password mutation APIs**

Expose clear operations instead of boolean bypass flags:

```go
func ChangeOwnPasswordContext(ctx context.Context, userID uint, currentPassword, newPassword string) error
func SetInitialPasswordContext(ctx context.Context, userID uint, password string) error
func ResetTemporaryPasswordContext(ctx context.Context, userID uint, password string) error
func EvaluatePasswordState(policy Policy, changedAt, effectiveAt int64, mustChange bool, now time.Time) PasswordState
```

`ChangeOwnPasswordContext` enforces strength and history. Initial/reset operations may bypass strength only when configuration permits, never bypass history, and set the mandatory-change flag according to policy.

- [x] **Step 5: Run the accountsecurity tests**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/service/accountsecurity -count=1
```

Expected: PASS.

### Task 3: Integrate every password write path

**Files:**
- Modify: `backend/internal/service/admin/adminmgr/password.go`
- Modify: `backend/internal/service/admin/adminmgr/service.go`
- Modify: `backend/internal/service/admin/adminuser/service.go`
- Modify: `backend/internal/service/client/passport/register.go`
- Modify: `backend/internal/service/dingtalkh5/account/service.go`
- Modify: `backend/internal/service/dingtalkh5/performance/user/service.go`
- Modify: `backend/internal/service/dingtalkh5/performance/template/service.go`
- Create: `backend/internal/service/accountsecurity/integration_structure_test.go`

- [x] **Step 1: Write a failing write-path coverage test**

Scan production packages and reject direct password hashing or `Update("user_password", ...)` outside `accountsecurity`, except the legacy-login rehash path which must update `PasswordChangedAt` without creating a user-visible password change.

- [x] **Step 2: Route password writes through accountsecurity**

Use `SetInitialPasswordContext` for newly created accounts, `ResetTemporaryPasswordContext` for administrator resets, and `ChangeOwnPasswordContext` for self-service changes. Preserve existing authorization and old-password verification at the transport boundary only where required by the shared service contract.

- [x] **Step 3: Verify creation/reset behavior**

Test that weak temporary passwords are accepted only for initial/reset flows, both flows set mandatory change when configured, and self-service weak passwords remain rejected.

### Task 4: Add password state to login and restricted sessions

**Files:**
- Create: `backend/internal/support/passwordsession/restriction.go`
- Create: `backend/internal/support/passwordsession/restriction_test.go`
- Modify: `backend/internal/service/admin/adminauth/service.go`
- Modify: `backend/internal/service/client/passport/login.go`
- Modify: `backend/internal/service/dingtalkh5/auth/service.go`
- Modify: `backend/internal/middleware/admin/auth.go`
- Modify: `backend/internal/middleware/client/auth.go`
- Modify: `backend/internal/middleware/dingtalk_h5/auth.go`
- Modify: `backend/internal/routes/v2/client/routes.go`
- Create: `backend/internal/handler/client/passport/password.go`
- Modify: `backend/internal/handler/dingtalkh5/account/handler.go`
- Modify: `backend/internal/handler/admin/mgr/handler.go`

- [x] **Step 1: Write failing login-state tests**

For Admin, password login, DingTalk password login, and DingTalk SSO, assert login responses include:

```json
{
  "passwordChangeRequired": true,
  "passwordExpired": false,
  "passwordExpiresAt": 0,
  "passwordExpiryWarning": false
}
```

- [x] **Step 2: Persist password state in Redis sessions**

Add `passwordChangeRequired` to Admin and client session JSON. For H5App, include the state in the authenticated session DTO used by `AuthenticateContext`. Do not trust a client-provided restriction flag.

- [x] **Step 3: Write failing middleware tests**

Restricted sessions must allow only current-account password change, logout, password-policy summary, and minimum current-user/bootstrap data. Business routes return code `PASSWORD_CHANGE_REQUIRED` with a stable Chinese message.

- [x] **Step 4: Implement middleware enforcement**

Use explicit method/path allowlists per audience. Avoid substring checks. Expired active sessions become restricted when policy state is evaluated, even when the token predates expiration.

- [x] **Step 5: Restore the current session after password change**

Password change endpoints update or replace the current session, invalidate other tokens for the account, and return the refreshed password state. Administrator reset invalidates every token belonging to the target account.

- [x] **Step 6: Run authentication and middleware tests**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/service/admin/adminauth ./internal/service/client/passport ./internal/service/dingtalkh5/auth ./internal/service/dingtalkh5/account ./internal/middleware/admin ./internal/middleware/client ./internal/middleware/dingtalk_h5 -count=1
```

### Task 5: Add the typed Admin password-policy API

**Files:**
- Create: `backend/internal/handler/admin/setup/password_policy.go`
- Create: `backend/internal/handler/admin/setup/password_policy_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Create: `backend/internal/routes/v2/swagger/password_policy.go`
- Modify generated Swagger through the project generation command.

- [x] **Step 1: Write failing route and DTO tests**

Require `GET /api/v2/admin/password-policy` and `PUT /api/v2/admin/password-policy`, using existing setup view/edit permissions. Verify invalid combinations return a stable validation message.

- [x] **Step 2: Implement the typed API**

The handler reads/writes `accountsecurity.Policy`; it must not expose the generic setup record or accept arbitrary JSON strings.

- [x] **Step 3: Update Swagger and permission contracts**

Run the documented `swag init` command, then run route, permission, and Swagger tests.

### Task 6: Build Admin policy and forced-password UI

**Files:**
- Create: `admin/src/types/passwordPolicy.ts`
- Create: `admin/src/api/passwordPolicy.ts`
- Create: `admin/src/views/setup/components/PasswordPolicySettings.vue`
- Create: `admin/src/views/change-password/index.vue`
- Modify: `admin/src/views/setup/index.vue`
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/api/types.ts`
- Modify: `admin/src/router/adminRoutes.ts`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/src/utils/request.ts`
- Modify: `admin/src/utils/adminSession.ts`
- Create: `admin/scripts/check-password-policy.mjs`
- Modify: `admin/package.json`

- [x] **Step 1: Write failing Admin contracts**

Require the typed policy API, Login Config password-policy section, forced-change route, route guard, `PASSWORD_CHANGE_REQUIRED` interceptor, and login response password-state fields.

- [x] **Step 2: Implement the settings component**

Use switches for binary rules, checkboxes for character requirements, and input-number controls for lengths/days/history. Disable dependent fields when policy or expiry is disabled. Save one complete typed payload.

- [x] **Step 3: Implement forced password change**

The login page stores password state. The router redirects restricted users to `/change-password`; that page cannot be closed or bypassed. On success replace session state, invalidate access snapshots, and redirect to the original target.

- [x] **Step 4: Add non-blocking expiry warnings**

When `passwordExpiryWarning` is true but change is not required, display one dismissible warning per login session.

- [x] **Step 5: Run Admin checks and browser validation**

```bash
cd admin
npm run check:password-policy
npm run check:all
```

Browser flows: disabled policy, policy save, restricted login redirect, blocked deep link, successful change, and expiry warning.

### Task 7: Add Frontend password-change flow

**Files:**
- Create: `frontend/pages/my/change_password.vue`
- Modify: `frontend/pages.json`
- Modify: `frontend/api/index.js`
- Modify: `frontend/utils/auth.js`
- Modify: `frontend/utils/request.js`
- Modify: `frontend/pages/login/login_pwd.vue`
- Modify: `frontend/pages/login/login.vue`
- Create: `frontend/scripts/check-password-policy.mjs`
- Modify: `frontend/package.json`

- [x] **Step 1: Write failing Frontend contracts**

Require password-state persistence, forced redirect, `PATCH /api/v2/me/password`, request-layer handling of `PASSWORD_CHANGE_REQUIRED`, and rule-summary rendering.

- [x] **Step 2: Implement the page and auth guard**

Restricted users may only access the change-password page. Use backend validation messages as authority and local checks only for empty fields and confirmation mismatch.

- [x] **Step 3: Verify Frontend**

```bash
cd frontend
npm run check:all
```

Also validate H5 and one mini-program target if available.

### Task 8: Add H5App forced-password flow

**Files:**
- Modify: `h5app/src/types/dingtalk-h5.ts`
- Modify: `h5app/src/stores/dingtalkAuth.ts`
- Modify: `h5app/src/components/app-auth-guard/app-auth-guard.vue`
- Modify: `h5app/src/components/app-shell/app-shell.vue`
- Modify: `h5app/src/api/dingtalk-h5/profile.ts`
- Create: `h5app/scripts/check-password-policy.mjs`
- Modify: `h5app/package.json`

- [x] **Step 1: Write failing H5App contracts**

Require password state for password login, SSO, and bind responses; an unclosable forced-change view; and re-bootstrap after successful change.

- [x] **Step 2: Implement forced change for every authentication channel**

Reuse the existing current/new/confirm password controls. When restricted, hide workbench navigation and prohibit closing the view. DingTalk SSO must not bypass the state.

- [x] **Step 3: Verify H5App**

```bash
cd h5app
pnpm lint
pnpm type-check
```

Run browser H5 validation plus a real DingTalk/custom-base validation when available.

### Task 9: Final security and regression verification

**Files:**
- Modify documentation only when behavior or API descriptions need synchronization.

- [x] **Step 1: Run focused security checks**

Verify logs, audit snapshots, API responses, Swagger and frontend state never expose plaintext passwords or password hashes.

- [ ] **Step 2: Run repository checks**

```bash
cd backend && GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
cd admin && npm run check:all
cd frontend && npm run check:all
cd h5app && pnpm lint && pnpm type-check
bash scripts/verify-local.sh
```

- [ ] **Step 3: Run end-to-end scenarios**

Cover administrator and ordinary user creation, weak temporary password, password login, DingTalk SSO, restricted API access, password change, password history rejection, expiry warning, expiration, administrator reset, token invalidation, policy disable/re-enable, and existing-account compatibility.

No commit or push is performed unless the user explicitly requests it.

