# Audit Snapshot Clarity

**Goal:** Show safe configuration changes and Chinese field/value descriptions, explain missing snapshots precisely, and preserve sensitive-value protection and historical uncertainty.

**Architecture:** Persist snapshot status/reason separately from business outcome/reason. Collect explicitly allowlisted public scalar settings as typed values; treat all other settings as private. The writer validates public scalar values again. Admin shows localized fields and capture states, preserving redaction markers in old records.

## Contract

Log JSON adds `snapshotStatus` and `snapshotReason`. SQL columns `log_snapshot_status` (varchar32, default `legacy_unknown`) and `log_snapshot_reason` (varchar80, default empty) are added by `20261029150000_add_audit_snapshot_status.sql`. No old values or outcomes are backfilled from current business data.

Statuses: `pending`, `recorded`, `unchanged`, `not_applicable`, `not_configured`, `operation_failed`, `unavailable`, `legacy_unknown`.

- pending: request begun but capture/finalization not completed.
- recorded: actual differences were recorded, some may be redacted.
- unchanged: supported collection completed and found no differences among collected fields, not a claim that every business field was unchanged.
- not_applicable: auth/export/create-only-target/runtime-trigger actions without a before/after strategy by design.
- not_configured: mutation has no snapshot strategy, or an unsupported configuration key.
- operation_failed: business failure/denial; no snapshot values attached.
- unavailable: collection/serialization failed or exceeded limits; business result remains separate.
- legacy_unknown: old record lacks capture metadata; do not guess its capture reason.

Reason codes include `no_field_snapshot`, `create_target_only`, `runtime_history`, `unsupported_route`, `unsupported_setup_key`, `before_read_failed`, `after_read_failed`, `before_limit`, `after_limit`, `diff_limit`, `encoding_limit`, `request_unavailable`, `invalid_target`, `result_unconfirmed`, `business_not_successful`, `no_collected_changes`. Other existing collector-local closed reason codes remain safely mapped by Admin.

## Public setting value policy

The backend exports `auditlog.PublicSnapshotValue(field string, value interface{}) (interface{}, bool)` for validated, normalized values. Unknown keys or invalid scalar values return false; no broad keyword exemption.

- Boolean exact fields: `settings.ADMIN_LOGIN_CAPTCHA_ENABLED`, `settings.DINGTALK_H5_LOGIN_CAPTCHA_ENABLED`, `settings.USER_SINGLE_LOGIN`, `settings.ADMIN_SINGLE_LOGIN`, `settings.DINGTALK_H5_SINGLE_LOGIN`, `settings.DINGTALK_H5_SELF_BIND_ENABLED`, `settings.DINGTALK_H5_NOTIFY_ENABLED`.
- Duration exact fields: `settings.TOKEN_ADMIN_EXPIRE`, `settings.TOKEN_USER_EXPIRE`, `settings.TOKEN_DINGTALK_H5_EXPIRE`; only positive parsed durations, normalized through time.Duration.String.
- Enum field `settings.DINGTALK_H5_NOTIFY_MODE`: agent, agent_fallback, robot.
- Corp fields `settings.corp.<positive integer>.enabled` and `.notify_enabled` are boolean; `.notify_mode` uses the same enum. Every other corp value stays private.
- Booleans normalize bool / 0 / 1 / string true,false,0,1 only. Nil means a missing value for a known field. Invalid strings, JSON objects/arrays and arbitrary values under public-looking keys remain redacted. Ordinary entity fields keep their current whitelist.

Snapshot collector writes successful differences with `recorded`, an empty successful comparison with `unchanged`; errors set status/reason without replacing business reason. Stop emitting `privateConfigChanged` for new records. Admin hides this internal marker in old records but explains that a configuration change was detected; it never interprets true as a setting being enabled.

## Work

- [x] Parent: model/context/writer/migration/read validation; public-scalar sanitizer; final status policy and integration tests; Swagger/docs.
- [x] Backend worker: typed safe settings, recorded/unchanged/error status, private markers and compatibility; new events no longer emit privateConfigChanged.
- [x] Admin worker: Chinese field/value labels, capture states, precise empty reasons and historical protection. Null means 未记录, never an inferred disabled/unset setting.
- [x] Verification: focused and race tests, Admin checks/build, isolated MySQL migration/write verification, and desktop/narrow browser checks completed. No production migration or historical-value reconstruction.

## Verification record — 2026-10-29

- Passed `GOCACHE=$PWD/../.cache/go-build GOPROXY=off GOTOOLCHAIN=local go test ./internal/support/auditlog ./internal/middleware ./internal/service/admin/adminlog ./migrations`, plus `go test -race` for the same packages.
- Independent privacy review found an aggregate `settings` field bypass; it was reproduced with a failing regression, fixed to default-redact root settings objects, and independently rechecked. Only exact public scalar settings can bypass masking.
- Passed `npm run check:audit-logs` (including `check:audit-snapshot`) and `npm run build`. Snapshot tests execute real helpers and render the real detail component.
- The migration ran twice on a task-owned temporary MySQL 8 database. Synthetic legacy rows kept `legacy_unknown`; `TestAuditWriterMySQLFinalizesOneEvent` verified begin/final status upsert and cancellation-independent persistence.
- Browser tests used a synthetic local API only: safe flags and statuses show Chinese values; secrets retain markers; old flags explain their meaning without exposing raw internal rows; all seven non-recorded statuses render their specific explanations; narrow layout checked. No page errors observed.
- Full `go test ./...` retains existing DingTalk settings assertions and tracked production configuration failures. A scheduled-task test initially hit the sandbox's local-port prohibition; rerunning that package with local-port access passed. Admin `npm run check:all` still stops at the existing navigation assertion `props.depth > 0`.
- Swagger was regenerated using the documented command. Existing unrelated dirty work is preserved; no commit or production deployment.

New storage requires migration before deploying the backend. No production data or configuration is changed by development verification.
