# Survey Notification Template Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Move survey notification template management from System Settings into a dedicated Survey menu page without changing stored template data.

**Architecture:** A dedicated backend facade exposes built-in and per-admin custom templates through one typed API. The new Admin route owns the UI while the setup page drops all template responsibilities. Versioned SQL registers and backfills the menu, button, and API permissions.

**Tech Stack:** Go/Hertz/GORM, Vue 3, TypeScript, Element Plus, MySQL migrations.

---

### Task 1: Lock route and permission contracts

**Files:**
- Create: `backend/internal/support/adminmenuperm/survey_template_structure_test.go`
- Create: `backend/internal/routes/v2/admin/survey_template_routes_test.go`
- Create: `backend/test/internal/bootstrap/migrations/survey_template_permission_migration_test.go`

- [x] Write failing tests for menu keys, API routes, route permissions, API catalog declarations, and migration snippets.
- [x] Run focused Go tests and confirm they fail before production changes.

### Task 2: Add the backend template facade

**Files:**
- Create: `backend/internal/service/admin/formkitadmin/notification_templates.go`
- Create: `backend/internal/service/admin/formkitadmin/notification_templates_test.go`
- Create: `backend/internal/handler/admin/survey/notification_templates.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`

- [x] Add service tests for default built-ins, custom data aggregation, valid saves, and invalid scope.
- [x] Implement GET/PUT service and handler boundaries while preserving existing storage keys.
- [x] Register the routes and permission catalog entries.
- [x] Regenerate Swagger and run focused backend tests.

### Task 3: Register historical menu permissions

**Files:**
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Create: `backend/migrations/20261028190000_add_survey_template_page_permissions.sql`

- [x] Add the menu and edit-button declarations beneath the Survey directory.
- [x] Add idempotent permission rows for menu, button, list API, and edit API.
- [x] Backfill list grants from existing survey-list grants and edit grants from existing survey-edit grants.
- [x] Run menu and migration tests.

### Task 4: Build the standalone Admin page

**Files:**
- Create: `admin/src/views/survey/template/index.vue`
- Create: `admin/src/views/survey/template/useSurveyNotificationTemplates.ts`
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/router/adminRoutes.ts`
- Create: `admin/scripts/check-survey-template-page.mjs`
- Modify: `admin/package.json`

- [x] Write the failing Admin contract check.
- [x] Add typed API methods and template state composable.
- [x] Build the page with Admin UI shell, tabs, table, dialog, permissions, loading and error states.
- [x] Register `/survey/templates` with `menuPath`, `requiredPermission`, and `adminUi` metadata.
- [x] Run the focused Admin check and production build.

### Task 5: Remove the System Settings copy

**Files:**
- Modify: `admin/src/views/setup/index.vue`

- [x] Remove the template tab markup.
- [x] Remove all preset refs, defaults, load/save/dialog functions, mounted loads, and returned bindings.
- [x] Extend the Admin contract check to reject the old tab and API snippets.
- [x] Run the focused check and build.

### Task 6: Browser and repository verification

- [x] Apply the new migration in the local development database only.
- [x] Verify the Survey menu opens the new page and System Settings no longer contains the old tab.
- [x] Verify built-in/custom add, edit, delete and reload persistence without leaving test data.
- [x] Run focused backend tests, Admin checks, build, bundle budget, and `git diff --check`.
- [x] Report unrelated full-suite failures separately. Do not commit or push in the shared dirty worktree.
