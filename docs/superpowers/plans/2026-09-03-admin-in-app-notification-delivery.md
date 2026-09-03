# Admin In-App Notification Delivery Implementation Plan

> **For Codex:** Execute this plan inline with test-driven development. Work only in `/Users/fanggang/项目文件/编程学习/WeCheckin` and preserve all unrelated dirty changes.

**Goal:** Let authorized administrators send user-scoped in-app notifications immediately or through the registered scheduled-task runtime.

**Architecture:** Reuse the existing `notify` inbox table and H5App read APIs. Add a dedicated `inappnotification` application module shared by an Admin HTTP transport and a registered Go scheduled job. Recipient rules are resolved against current active users at delivery time, and a nullable delivery key provides idempotency without changing legacy workflow notification semantics.

**Tech Stack:** Go 1.24, Hertz, GORM, MySQL, Vue 3, TypeScript, Element Plus, npm.

---

## Task 1: Persist delivery idempotency and register permissions through SQL

**Files:**
- Modify: `backend/internal/model/content/content.go`
- Create: `backend/migrations/20260903160000_add_in_app_notification_delivery.sql`
- Create: `backend/test/internal/bootstrap/migrations/in_app_notification_delivery_migration_test.go`
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`
- Create: `backend/internal/support/adminmenuperm/notification_structure_test.go`
- Create: `backend/internal/support/adminrouteperm/notification_structure_test.go`

1. Write migration and structure tests that require the nullable `notify_delivery_key` column, its unique index, the `notification:list/read/send` menu and API permissions, canonical routes, and explicit grant backfill SQL.
2. Run the focused migration and permission tests and confirm they fail because the migration and declarations do not exist.
3. Add the migration using `INFORMATION_SCHEMA` guards and `INSERT ... ON DUPLICATE KEY UPDATE`; do not add any AutoMigrate or startup seed behavior.
4. Add `DeliveryKey *string` to `model.Notify` so legacy writers persist `NULL` and only the new delivery service supplies a key.
5. Add menu/API declarations and route permission mappings for all canonical notification routes.
6. Run the focused tests and confirm they pass.

## Task 2: Build the shared notification application service

**Files:**
- Create: `backend/internal/modules/inappnotification/application/types.go`
- Create: `backend/internal/modules/inappnotification/application/service.go`
- Create: `backend/internal/modules/inappnotification/application/service_test.go`

1. Define named request/result DTOs, scope constants, stable validation errors, a recipient/delivery store interface, and sources for manual and scheduled delivery.
2. Write failing table-driven tests for title/content limits, invalid scope, missing IDs, duplicate IDs, all/users/departments resolution, inactive-user filtering, descendant department behavior, empty recipients, successful delivery, replay, and persistence failure.
3. Implement normalization and validation without GORM or Hertz dependencies.
4. Generate deterministic per-user delivery keys from source type, source ID, and local user ID.
5. Make the service return planned, sent, skipped, and replayed data without exposing full recipient lists.
6. Run `GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/inappnotification/application -count=1` from `backend` and confirm it passes.

## Task 3: Implement the GORM recipient and inbox store

**Files:**
- Create: `backend/internal/modules/inappnotification/infrastructure/gorm_store.go`
- Create: `backend/internal/modules/inappnotification/infrastructure/gorm_store_test.go`

1. Write failing tests around pure query/input normalization helpers and delivery record mapping; use fake store coverage in Task 2 for transaction behavior because the repository has no SQLite/sqlmock test dependency.
2. Resolve active users from `users.user_status=1`; for department scope, walk current `departments.dept_parent_id` descendants and query `user_depts`.
3. Insert one `notify` row per deduplicated local user with `CreateInBatches` inside a transaction.
4. Before insert, detect an existing source batch. On delivery-key duplicate, roll back and re-read the completed batch as a replay; propagate unrelated database errors.
5. Add user-scoped list, unread count, mark-one-read, and mark-all-read methods for the canonical Admin inbox routes. Every read/update must include the current Admin user ID.
6. Run focused application/infrastructure tests.

## Task 4: Add canonical Admin notification APIs

**Files:**
- Create: `backend/internal/modules/inappnotification/transport/httpadmin/handler.go`
- Create: `backend/internal/modules/inappnotification/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Modify by generator: `backend/docs/swagger/docs.go`
- Modify by generator: `backend/docs/swagger/swagger.json`
- Modify by generator: `backend/docs/swagger/swagger.yaml`
- Create: `backend/cmd/in_app_notification_routes_test.go`

1. Write failing handler tests for list, unread count, mark one, mark all, manual send, invalid JSON, missing current Admin, validation errors, and stable response fields.
2. Write a route structure test for `GET/POST /in-app-notifications`, `GET /in-app-notifications/unread-count`, `PATCH /in-app-notifications/read-all`, and `PATCH /in-app-notifications/:id/read`.
3. Wire one GORM store/application service into the Admin route suite and register all routes under existing Admin auth and permission middleware.
4. Keep legacy `/survey-notifications` routes unchanged for compatibility, but make the new Admin page use canonical routes.
5. Add Swagger DTOs/declarations and regenerate Swagger only with the repository command.
6. Run focused handler, route, middleware, and Swagger tests.

## Task 5: Register and authorize the scheduled in-app notification job

**Files:**
- Create: `backend/internal/modules/scheduledtask/infrastructure/in_app_notification_job.go`
- Modify: `backend/internal/modules/scheduledtask/infrastructure/builtin_jobs_test.go`
- Modify: `backend/internal/modules/scheduledtask/transport/httpadmin/handler.go`
- Modify: `backend/internal/modules/scheduledtask/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/cmd/taskd/main.go`
- Modify: `backend/cmd/main_safety_test.go`
- Modify: `docs/SCHEDULED_TASKS.md`

1. Write failing job tests for key/name/schema, valid execution, validation failure, temporary store failure, replay result, and sanitized run logging.
2. Implement `notification.in_app.send` by adapting job params to the shared application service and source `scheduled_task_run/runID`.
3. Extend `riskPermission` so Go handler key `notification.in_app.send` requires `notification:send`; write handler tests proving create/update authorization cannot be bypassed.
4. Register the same job in Admin metadata/validation wiring and `taskd` execution wiring.
5. Add structure tests proving both process entry points register the job.
6. Update the scheduled-task operations document and run focused scheduled-task and command tests.

## Task 6: Build the Admin in-app notification page

**Files:**
- Create: `admin/src/types/notification.ts`
- Modify: `admin/src/api/index.ts`
- Create: `admin/src/views/notification/NotificationSendDialog.vue`
- Create: `admin/src/views/notification/index.vue`
- Modify: `admin/src/router/adminRoutes.ts`
- Modify: `admin/src/views/survey/SurveyNotify.vue`
- Create: `admin/scripts/check-in-app-notifications.mjs`
- Modify: `admin/package.json`

1. Add a failing static check requiring typed canonical API methods, the `/notifications` route, legacy route compatibility, permission checks, recipient scope controls, user-tree selection, request ID reuse, loading/empty/error states, and submit locking.
2. Add strict TypeScript DTOs and API methods for list, unread count, mark read/all, and send.
3. Build the top-level page with a compact header, inbox table, pagination, read controls, and a permission-gated “发送站内信” action.
4. Build a dedicated send dialog using a segmented scope selector and `WorkflowUserTreePicker`; load organization options lazily and preserve form data after recoverable errors.
5. Generate one request ID per dialog submission session, disable duplicate submission, show the actual sent/skipped count, then refresh the inbox.
6. Point both `/notifications` and legacy `/survey/notify` at the canonical page without duplicating business logic.
7. Add the new static check to `check:all`, then run it and the Admin build.

## Task 7: Add structured scheduled-task configuration

**Files:**
- Modify: `admin/src/views/scheduled-task/components/TaskEditorDialog.vue`
- Modify: `admin/scripts/check-scheduled-task.mjs`

1. Extend the static scheduled-task check so it fails until the new Go job is permission-filtered and uses structured title/content/scope/user-tree controls.
2. Filter `notification.in_app.send` from Go job options unless `admin:menu:notification:send` is present.
3. When selected, replace raw params JSON with structured fields bound to `handlerConfig.params`; keep raw JSON for every other Go job.
4. Reuse the organization option loader without requiring a workflow definition and preserve existing workflow starter behavior.
5. Normalize edited task params, validate required recipients, and prevent saving the job if permission was lost.
6. Run `npm run check:scheduled-task`, the new notification check, and `npm run build`.

## Task 8: Verify the complete behavior and preserve unrelated work

**Files:**
- Verify only; modify only when a failure is caused by this feature.

1. Run `gofmt` on only changed Go files.
2. Run focused backend package tests, then `GOCACHE=$PWD/../.cache/go-build go test ./... -count=1` from `backend`.
3. Run `npm run check:all` from `admin`.
4. Start the existing Admin dev server on an available port and use the in-app browser to verify menu visibility, send dialog layout, all three scopes, validation, success state, permission-hidden state, and the structured scheduled-task editor.
5. Where a local authenticated H5 session is available, send to that user and verify the existing header unread badge and inbox list update. If authentication or taskd infrastructure is unavailable, report that as an explicit remaining integration risk.
6. Inspect `git diff --check`, targeted diffs, and final status. Do not stage, commit, clean, or revert unrelated dirty changes.
