# H5App In-App Notifications Implementation Plan

> **For Codex:** Execute this plan inline with test-driven development. Preserve all unrelated dirty changes in the shared checkout.

**Goal:** Add a functional station inbox entry to the H5App header, backed by user-scoped notification APIs and a responsive notification panel.

**Architecture:** Reuse the existing `notify` table and workflow notification writes. Add an authenticated H5 notification service whose repository queries always include the current local user ID, expose four `/api/v2/dingtalk/h5/notifications` endpoints, and consume them from a standalone H5App notification panel. Workflow notifications open the existing workflow instance detail tab.

**Tech Stack:** Go, Gin, GORM, Vue 3, Pinia, uni-app, uView Pro, TypeScript, pnpm.

---

## Task 1: User-scoped notification service

**Files:**
- Create: `backend/internal/service/dingtalkh5/notification/service.go`
- Create: `backend/internal/service/dingtalkh5/notification/service_test.go`

1. Write failing tests for pagination normalization, current-user scoping, mark-read ownership, mark-all-read, and missing notifications.
2. Run `GOCACHE=$PWD/.cache/go-build go test ./internal/service/dingtalkh5/notification -count=1` and confirm the tests fail.
3. Implement the repository contract, GORM repository, DTO mapping, and service methods against the existing `notify` table.
4. Run the focused service tests and confirm they pass.

## Task 2: Authenticated H5 API handlers and routes

**Files:**
- Create: `backend/internal/handler/dingtalkh5/notification/handler.go`
- Create: `backend/internal/handler/dingtalkh5/notification/handler_test.go`
- Modify: `backend/internal/handler/dingtalkh5/handler.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/routes.go`
- Create: `backend/internal/routes/v2/dingtalkh5/notification_routes_test.go`
- Modify: `backend/internal/routes/v2/swagger/h5app.go`

1. Write failing handler and route-structure tests for list, unread count, mark read, mark all read, authentication placement, invalid IDs, and user ownership errors.
2. Run focused handler and route tests and confirm they fail.
3. Implement handlers using `dingtalkh5session.CurrentUser`; register routes under the authenticated base group, outside app permission middleware.
4. Add Swagger declarations for all four endpoints.
5. Run focused tests and regenerate Swagger with the repository's documented command.

## Task 3: H5App API client and notification panel

**Files:**
- Create: `h5app/src/api/notifications.ts`
- Create: `h5app/src/components/app-notification-panel/app-notification-panel.vue`
- Modify: `h5app/src/components/app-shell/app-shell.vue`
- Modify: `h5app/src/locale/lang/zh-CN.json`
- Modify: `h5app/src/locale/lang/en-US.json`
- Modify: `h5app/scripts/check-dingtalk-module.mjs`

1. Extend the static module check with expected notification API paths, header entry, unread badge, and workflow-instance navigation; run it and confirm failure.
2. Add typed API calls for list, unread count, mark read, and mark all read.
3. Add the icon button immediately left of the menu-layout button; hide the badge at zero and cap it at `99+`.
4. Implement a 420px right panel on desktop and full-screen panel on mobile, including loading, empty, error, unread, pagination, mark-read, and mark-all-read states.
5. Open workflow notifications through `workflowInstanceContentKey`; display non-workflow notification content in the panel.
6. Add Chinese and English locale strings and responsive styles.
7. Run static checks, type-check, lint, and H5 build.

## Task 4: End-to-end verification

**Files:**
- Verify only; do not alter unrelated modules.

1. Run focused and full backend tests with a workspace-local Go cache.
2. Run H5App module checks, type-check, lint, and production build with pnpm.
3. Start the H5App development server on an available local port.
4. Use the in-app browser to verify the header placement, unread badge, desktop drawer, mobile full-screen panel, mark-read actions, and workflow-detail navigation at desktop and mobile widths.
5. Inspect final diffs and Git status to ensure only intended feature files were changed.
