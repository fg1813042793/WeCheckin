# Admin Login Captcha Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add configurable, one-time image captcha verification to Admin password login.

**Architecture:** Implement captcha generation and verification inside the Admin auth service with injectable generator/store dependencies. The production generator uses `base64Captcha`, the store uses Redis with atomic consume semantics, and the setup service supplies the feature switch.

**Tech Stack:** Go 1.24, Hertz, Redis, base64Captcha, Vue 3, Element Plus.

---

### Task 1: Captcha Domain and Redis Store

**Files:**
- Create: `backend/internal/service/admin/adminauth/captcha.go`
- Test: `backend/internal/service/admin/adminauth/captcha_test.go`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

- [x] Write failing tests for disabled mode, generation, case-insensitive one-time verification, expiry/missing challenge, and store failure.
- [x] Add `base64Captcha` dependency and implement generator, setup-backed switch, hashed Redis store, and atomic consume.
- [x] Run Admin auth package tests.

### Task 2: Public Captcha and Login Contract

**Files:**
- Modify: `backend/internal/handler/admin/mgr/handler.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Test: `backend/internal/handler/admin/mgr/captcha_structure_test.go`
- Test: `backend/internal/routes/v2/admin/captcha_route_test.go`

- [x] Add failing route/handler tests.
- [x] Register no-store captcha route and pass captcha fields to `LoginWithCaptchaContext`.
- [x] Add Swagger annotations and verify route catalogs.

### Task 3: Admin Login and Setup UI

**Files:**
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/api/types.ts`
- Modify: `admin/src/views/login/index.vue`
- Modify: `admin/src/views/setup/index.vue`
- Modify: `admin/scripts/check-login-input-focus.mjs`

- [x] Add captcha API types and requests.
- [x] Add conditional captcha input, image refresh, validation, and post-failure refresh.
- [x] Add `ADMIN_LOGIN_CAPTCHA_ENABLED` switch to Login Settings.
- [x] Extend static UI contract checks.

### Task 4: Verification

- [x] Run focused Admin auth, handler, route, setup, and Swagger tests.
- [x] Run Admin login checks and production build.
- [x] Run Backend full tests and report unrelated baseline failures separately.
- [x] Verify `h5app/` remains untouched and run scoped `git diff --check`.
