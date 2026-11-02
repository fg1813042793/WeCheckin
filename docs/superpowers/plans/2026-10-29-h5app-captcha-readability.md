# H5App Captcha Readability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the DingTalk H5 password-login captcha consistently as readable as the approved A reference without weakening its existing interference settings.

**Architecture:** Increase the shared Base64 captcha canvas while preserving answer length, noise count, and both line options. Update the H5 login panel to render the full image at a larger stable size, and encode both decisions in focused backend and frontend contract tests.

**Tech Stack:** Go 1.24, base64Captcha, Vue 3, uni-app, TypeScript, uView Pro, SCSS.

---

### Task 1: Lock and update captcha generator dimensions

**Files:**
- Modify: `backend/internal/support/logincaptcha/captcha_test.go`
- Modify: `backend/internal/support/logincaptcha/captcha.go`

- [ ] **Step 1: Write the failing generator configuration test**

Add a test that type-asserts `NewImageGenerator()` to `*imageGenerator` and requires `Width == 180`, `Height == 56`, `NoiseCount == 2`, the existing sine plus slime line flags, and the readable `RitaSmith.ttf` font.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/support/logincaptcha -run TestImageGeneratorUsesReadableCanvasWithoutReducingInterference -count=1`

Expected: FAIL because the current driver is `150×48`.

- [ ] **Step 3: Increase only the image canvas**

Change `NewDriverString` to use height `56`, width `180`, and `RitaSmith.ttf`. Keep noise `2`, both existing line flags, length `5`, source characters, background, storage, TTL, and verification behavior unchanged.

- [ ] **Step 4: Run the package tests and verify GREEN**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/support/logincaptcha -count=1`

Expected: PASS.

### Task 2: Render the complete larger image in H5App

**Files:**
- Modify: `h5app/scripts/check-dingtalk-module.mjs`
- Modify: `h5app/src/components/dingtalk-login-panel/dingtalk-login-panel.vue`

- [ ] **Step 1: Add failing H5 module contract markers**

Require the login component to contain `mode="aspectFit"`, mobile width/flex-basis `256rpx`, desktop width/flex-basis `128px`, and desktop height `44px`.

- [ ] **Step 2: Run the contract check and verify RED**

Run: `cd h5app && pnpm check:dingtalk-module`

Expected: FAIL because the component still uses `aspectFill`, `220rpx`, and `110×40px`.

- [ ] **Step 3: Update the H5 image presentation**

Switch the image to `aspectFit`, use `256rpx × 88rpx` on mobile and `128px × 44px` on desktop, and keep the existing image-click and keyboard refresh behavior unchanged without adding a separate refresh button.

- [ ] **Step 4: Run focused and full H5 verification**

Run:

```bash
cd h5app
pnpm check:dingtalk-module
pnpm lint
pnpm type-check
pnpm check:ui-style
pnpm build:h5
```

Expected: PASS, with any unrelated pre-existing baseline failure reported separately.

### Task 3: Visual regression check

**Files:**
- No repository changes.

- [ ] **Step 1: Generate multiple captcha challenges**

Exercise the local captcha endpoint or shared generator several times after rebuilding/restarting the backend.

- [ ] **Step 2: Check desktop and mobile login layouts**

Verify that characters are not cropped or stretched, the input remains usable, refresh remains clickable, and several generated samples are at least as readable as the approved A reference.

- [ ] **Step 3: Review the focused diff**

Run: `git diff --check -- backend/internal/support/logincaptcha/captcha.go backend/internal/support/logincaptcha/captcha_test.go h5app/scripts/check-dingtalk-module.mjs h5app/src/components/dingtalk-login-panel/dingtalk-login-panel.vue`

Expected: no output.
