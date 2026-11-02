# Admin Menu Hidden Implementation Plan

> **For agentic workers:** Execute inline in the current dirty workspace. Preserve unrelated changes and do not commit or push.

**Goal:** Add a persistent Admin menu hidden switch that removes configured menu subtrees from the left navigation without removing authorization or route access.

**Architecture:** Extend the unified permission model and management DTO with `hidden` and add an idempotent migration. Keep the complete menu tree for route authorization and page caching, while deriving a filtered tree only for left-navigation display. Keep permission-key resolution unchanged and add the field and switch to the Admin permission management page.

**Tech Stack:** Go, GORM, MySQL migration SQL, Hertz, Vue 3, TypeScript, Element Plus.

---

### Task 1: Add failing contracts

- [x] Extend model, service and handler structure tests for `Hidden`, `permission_hidden` and `hidden` form parsing.
- [x] Add runtime tests proving the authorization tree retains hidden nodes and permission keys while the sidebar display tree removes hidden subtrees.
- [x] Add a migration contract test and an Admin UI contract script.
- [x] Run focused tests and confirm they fail for the missing field.

### Task 2: Implement backend persistence and filtering

- [x] Add `Hidden` and `HiddenValue()` to the permission model.
- [x] Add the versioned migration and schema readiness check.
- [x] Map `hidden` through permission tree nodes, save requests, create and update operations.
- [x] Keep hidden nodes in the authorization tree and leave permission keys unchanged.
- [x] Add Swagger form parameters and regenerate Swagger output.

### Task 3: Implement Admin UI

- [x] Add `hidden` to the Admin menu contract.
- [x] Add the list status column and conditional edit switch.
- [x] Default new records to visible, restore edit values and submit `hidden`.
- [x] Derive a filtered sidebar tree while retaining the full tree for authorization and route caching.
- [x] Notify the layout to refresh the effective menu after save.

### Task 4: Verify

- [x] Run focused backend tests, Admin contract checks and build.
- [x] Run broader checks and separate unrelated baseline failures.
- [x] Browser-test the list column and edit switch without saving or modifying existing data; direct-route authorization is covered by the complete-tree contract check.

## Execution Result

- Focused model, permission schema, Admin permission handler/service, Admin menu runtime, migration and Swagger tests passed.
- Admin menu-hidden, request, route-cache, Swagger checks and production build passed.
- Browser verification confirmed the “左侧菜单” column and “左侧菜单隐藏” switch on an Admin menu record; no save was performed.
- `npm run check:all` remains blocked by the existing navigation assertion `props.depth > 0`.
- Full backend tests remain blocked by unrelated DingTalk settings text assertions, audit snapshot expectations, sandbox `httptest` binding and tracked production config safety checks.
- Migration execution and backend restart were not performed.

