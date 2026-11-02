# Scheduled Workflow Start Notification Implementation Plan

> **For agentic workers:** Execute inline in the shared dirty checkout. Preserve unrelated changes and do not commit, push, migrate or send real notifications.

**Goal:** Add a scheduled Go job that selects a workflow-start DingTalk template and dynamically sends it to the workflow's current eligible initiators.

**Architecture:** Resolve the latest enabled template at execution, load its published workflow definition, translate the initiator configuration into existing notification recipient rules, remove exclusions, and call the existing template-aware DingTalk sender with the scheduled run ID. Extend template options with a target-kind filter and add a dedicated template selector to the Admin task editor.

**Tech Stack:** Go, scheduled-task registry, workflow application service, in-app notification application/store, Vue 3, Element Plus, TypeScript.

---

### Task 1: Add failing backend job contracts

- [x] Add execution tests for all/specified initiator scopes, union, deduplication, exclusions, no recipients and delivery counts.
- [x] Add template resolver tests and target-kind option filtering tests.
- [x] Add taskd/HTTP registry and risk-permission structure tests.
- [x] Run focused tests and verify RED.

### Task 2: Implement backend job and wiring

- [x] Expose latest enabled workflow-start template resolution.
- [x] Add `notification.workflow_start.send` with `{templateId}` schema.
- [x] Resolve current published workflow and current active recipients at each execution.
- [x] Register the job in HTTP and taskd registries and map risk permission.

### Task 3: Implement Admin editor

- [x] Extend template option query with `targetKind`.
- [x] Add a searchable enabled workflow-start template selector.
- [x] Hide ordinary notification title/content/recipient fields for the new job.
- [x] Save only `params.templateId` and validate permission/selection.

### Task 4: Verify

- [x] Run scheduled-task, notification, workflow and Admin focused checks and builds.
- [x] Browser-test create/edit UI using synthetic handler metadata without saving or sending.
- [x] Run broad suites and record unrelated baseline failures separately.

Verification notes:

- Focused backend notification, scheduled-task, workflow, route and taskd tests passed.
- Admin scheduled-task/template checks and production build passed.
- Browser verification selected `notification.workflow_start.send`, loaded only `workflow_start` templates, displayed the dynamic-recipient explanation and did not render manual title/content/recipient fields.
- Full Backend remains blocked by pre-existing DingTalk settings text assertions, sandbox-denied `httptest` port binding and tracked production-config security failures.
- Admin `check:all` remains blocked by the pre-existing navigation contract marker `props.depth > 0`.

