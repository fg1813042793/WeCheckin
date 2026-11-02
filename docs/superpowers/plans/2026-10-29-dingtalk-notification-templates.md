# DingTalk Notification Templates Implementation Plan

> Execution uses superpowers:subagent-driven-development with disjoint Backend/Admin write scopes and parent integration review. User authorized implementation; no commits, external sends, production migrations or deployment are part of this task.

**Goal:** Manage reusable send templates in Message Styles and select one when sending a DingTalk notification to open a chosen workflow start page.

**Architecture:** Add a versioned template table and a bounded CRUD/option service in the existing inappnotification module. Resolve and validate template ID/version before delivery, then build links for each recipient corporation using the existing openapp wrapper. Preserve untemplated and automatic notification behavior.

**Tech Stack:** Go/Hertz/GORM/MySQL, Vue 3/Element Plus/TypeScript; no new dependencies.

## Contract and tasks

See `docs/superpowers/specs/2026-10-29-dingtalk-notification-templates-design.md` for exact request/response fields, bounds, routes and behavior.

- [x] Backend model/migration and template CRUD: model, versioned migration, application/store/handler boundaries and package-local behavior tests implemented. Validation includes names, lengths, types, URLs and update/delete version conflicts.
- [x] Send integration: server validates template ID/version and snapshots its configuration before any external delivery. Inactive, missing, stale or unpublished targets fail; each corporation gets its own openapp URL. Existing sends remain compatible.
- [x] Admin: typed API, template editor/manager, searchable workflow target and template selection implemented. Sending and business-style components extracted from the oversized page. Selected sends carry templateId and templateVersion.
- [x] Routes/permissions: six routes use existing privileges, template writes are audited with request body redaction, and Swagger is regenerated from source.
- [x] Verification: focused suites, race checks, isolated MySQL smoke, Admin runtime checks and build passed. Full suites were run with existing failures recorded below.
- [x] Browser: desktop and 390px views validated against isolated synthetic APIs with no external sends. CRUD, sender-only permission, switching cancellation, version refresh recovery and default-send failure compatibility passed.
- [x] Review/docs: focused spec and correctness review completed; two UI recovery findings fixed and independently rechecked. API/use/deployment documents updated; pre-existing dirty changes preserved.

## Verification record — 2026-10-29

- `GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/inappnotification/... ./internal/service/dingtalkh5/config ./internal/support/notificationstyle ./migrations -count=1` passed.
- `GOCACHE=$PWD/../.cache/go-build go test -race ./internal/modules/inappnotification/... ./internal/middleware/admin ./internal/support/auditlog ./migrations` passed.
- `GOCACHE=$PWD/../.cache/go-build go test ./cmd ./internal/middleware ./internal/middleware/admin ./internal/support/auditlog -count=1` passed after Swagger regeneration.
- The new SQL migration ran twice successfully on isolated MySQL 8 `template_verify`. A temporary GORM smoke program verified create, duplicate-name rejection, search, published workflow options, update version conflicts, disabling a closed-workflow template, stale-delete rejection and deletion. The task-owned container and temporary program were removed afterward.
- Admin `npm run check:notification-templates`, `npm run check:in-app-notification`, API/UI contract checks and `npm run build` passed. Runtime tests include complete pagination, request races, CRUD permissions, default send during option failures, template version refresh and dirty-content preservation.
- Backend `go test ./...` still fails the existing DingTalk settings UI assertions and tracked production configuration check. Admin `npm run check:all` stops at the existing `props.depth > 0` navigation assertion; the separate existing user-page component budget also remains outside scope. These checks are not reported as passing.
- Browser mock verification: create/edit/disable/delete template; select workflow; chosen template metadata and temporary title delivered; cancel switching preserves edits; stale version rejects then refresh/cancel/confirm submits new version; failed candidate load permits explicit untemplated send; sender-only access hides editing; mobile controls fit; no observed page errors or console errors/warnings in final flows. Management table width and scrollWidth both 998px, with version and actions visible.
- No production migration, deployment, real-recipient send or DingTalk client deep-link validation performed. Actual OA rendering requires the configured Agent work-notification channel; robot mode uses the existing supported message mapping.

## Baseline

Current checkout is shared and contains unrelated audit, captcha, user-filter, survey-template and workflow-summary work. Known earlier full-check failures include Admin navigation assertions and Backend DingTalk settings/tracked production configuration checks; final runs must report current outcomes, not assume they remain unchanged.
