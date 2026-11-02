# Unified Audit Logs Implementation Plan

**Goal:** Record login outcomes and prioritized business mutations across Admin, H5App and client entry points, with a read-only audit UI.

**Architecture:** Extend the existing `logs` table through a versioned migration. A request-level audit collector records business outcomes and denied requests; authentication and business helpers attach verified actors, reason codes, safe changes and related runtime IDs. Existing workflow histories, task runs and notification receipts remain authoritative for execution details. Legacy records have an explicit unknown outcome.

**Tech Stack:** Go/Hertz/GORM/MySQL, Vue 3/TypeScript/Element Plus.

## Contracts

- Keep existing `id`, `type`, `content`, `adminId`, `adminName`, `adminDesc`, `LOG_ADD_IP`, `_createTime` properties for compatibility.
- Add `eventId`, `category`, `source`, `module`, `action`, `outcome`, `reasonCode`, `reason`, `actorType`, `attemptedAccount`, `targetType`, `targetId`, `targetIds`, `targetName`, `requestedCount`, `affectedCount`, `requestId`, `method`, `path`, `durationMs`, `userAgent`, `browser`, `deviceType`, `platformType`, `changes`, `relatedType`, `relatedId`.
- Categories: `login`, `operation`; sources: `admin`, `client`, `h5app`, `system`.
- Outcomes: `success`, `failure`, `blocked`, `partial`, `accepted`, `unknown`. HTTP 200 alone does not imply success.
- `GET /api/v2/admin/logs` serves scoped operations; `/login-logs` serves separately authorized security logs. Both support `/:id` and `/export`. Exports are capped at 1000 records with accurate total/truncation information.
- Query filters: `search`, `source`, `module`, `action`, `outcome`, `actorId`, `ip`, `startTime`, `endTime`, `page`, `pageSize`.
- New actions: `log:security`, `log:export`, `log:security-export`; no implicit security-log grants to ordinary log viewers. Legacy `DELETE /logs` refuses deletion.
- No passwords, tokens, captcha answers, raw questionnaire answers or secret configuration values are stored in audit details. Authentication attempts remain unverified until an issuer supplies the real actor.

## Execution

- [x] Storage/API: extend model, idempotent migration, legacy unknown normalization, filtered/scoped list/detail/export, read-only deletion behavior and permission checks.
- [x] Collection: attach server-generated request/event IDs, classify supported modules and actions, record business outcomes after handlers, capture auth/permission failures before handler dispatch, and expose audit write failures through structured fallback logs.
- [x] Authentication: successful/failed passwords, captcha, SSO, bindings, logout, password changes and session invalidation, across current and compatibility routes.
- [x] Business context: whitelisted before/after changes for accounts, roles, permissions, departments, positions, dictionary/configuration, questionnaires/exams and workflows; sensitive values represented only as changes. Record counts and related IDs for batch operations, exports, task requests and notification dispatch.
- [x] UI: login/operation tabs, structured filters, result tags, details and changes, authorized server-audited CSV export, standalone pagination, no destructive controls.
- [x] Validation: focused behavior tests for failure/denial/HTTP-200 errors, legacy logs, redaction, read access, export limits and asynchronous acceptance; backend full tests, Admin checks/build, browser mock-flow validation and migration review. Full-suite exceptions are recorded below.
- [x] Documentation: API and permissions, route/module coverage, migration and rollback instructions, limitations of request-level logging and existing runtime records.

## Verification record — 2026-10-28

- Passed `GOCACHE=$PWD/../.cache/go-build go test -race` for audit support, middleware, adminlog, adminauth, client passport, H5 authentication, survey handler/service, and migrations.
- Passed Admin `npm run check:audit-logs` (including CSV export checks), `npm run check:request`, and `npm run build`.
- Applied the migration repeatedly to an isolated MySQL 8 `audit_verify` database; legacy rows and permission uniqueness were preserved. `TestAuditWriterMySQLFinalizesOneEvent` passed against this database, including finalization after request cancellation.
- Browser verification used synthetic API data at desktop and 390px widths: login outcomes, filters/reset, details, read permissions, export/error handling, and standalone pagination. No real business records were modified or exported.
- Ran backend `go test ./...`; remaining failures are existing DingTalk settings UI assertions and tracked production configuration checks. Ran Admin `npm run check:all`; its existing navigation assertion (`props.depth > 0`) blocks later checks. These failures are not reported as passing.
- Production migration, deployment, real authentication/provider integration, and real-data export have not been performed. Request snapshots are observational rather than transaction-atomic; see `docs/AUDIT_LOGS.md` for operational boundaries.

## Rollout and Boundaries

Apply the database migration before deploying application binaries. Do not run production migrations or remove existing records during development. Export and routine reads must not recursively generate duplicate audit events. Failed audit persistence must be visible; successful business mutation and eventual execution are separate outcomes. Security events for unauthenticated users cannot use the ordinary actor data scope as their authorization boundary. Existing unrelated working-tree edits, especially the ongoing login captcha work, remain intact.
