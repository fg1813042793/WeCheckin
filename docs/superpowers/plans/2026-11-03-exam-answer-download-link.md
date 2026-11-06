# Exam Answer Download Link Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a configurable, expiring, passwordless answer-sheet download URL to exam post-submit notification templates.

**Architecture:** Store opaque export grants in Redis under SHA-256 token digests and expose one public download endpoint that resolves the grant before reusing the existing answer-sheet renderer. Persist format, expiry days, and an independent download-service base URL in exam settings; create a grant only when `{answerDownloadUrl}` is present in a post-submit template.

**Tech Stack:** Go 1.24, Hertz, go-redis v9, Vue 3, TypeScript, Element Plus, existing PDF/DOCX/XLSX renderer.

---

### Task 1: Validate and persist exam download settings

**Files:**
- Modify: `backend/internal/service/client/exam/display_title.go`
- Modify: `backend/internal/service/client/exam/display_title_test.go`
- Modify: `backend/internal/handler/admin/exam/handler.go`
- Modify: `admin/src/views/exam/composables/useExamDesignerDocument.ts`
- Modify: `admin/src/views/exam/components/ExamDesignerSettingsPanel.vue`
- Modify: `admin/scripts/check-exam-designer-ux.mjs`

- [x] Add failing Go and Admin contract tests for `answerDownloadFormat`, `answerDownloadExpireDays`, and `answerDownloadBaseUrl`.
- [x] Normalize formats to `pdf|docx|xlsx`, require expiry days `1-30`, and preserve unrelated JSON settings.
- [x] Add format, expiry, and independent download-service address controls under “投放与结果”.
- [x] Require the independent service address whenever a template uses `{answerDownloadUrl}`.
- [x] Run `GOCACHE=$PWD/../.cache/go-build go test ./internal/service/client/exam -count=1` and `npm run check:exam-designer-ux`.

### Task 2: Add Redis export grants

**Files:**
- Create: `backend/internal/modules/examrecordexportlink/application/service.go`
- Create: `backend/internal/modules/examrecordexportlink/application/service_test.go`
- Create: `backend/internal/modules/examrecordexportlink/infrastructure/redis_store.go`
- Create: `backend/internal/modules/examrecordexportlink/infrastructure/redis_store_test.go`

- [x] Write failing tests for token hashing, 256-bit token generation, TTL, grant validation, invalid format, invalid expiry and invalid public URL.
- [x] Implement `Issue` and `Resolve` around a small `Store` interface.
- [x] Store JSON `{examId,recordId,format,expiresAt}` under `exam_record_export_link:<sha256>` with the requested TTL.
- [x] Return an absolute URL whose only caller-controlled secret is the `token` query value.
- [x] Run the two new package test suites with `-count=1`.

### Task 3: Reuse the existing exporter behind a public token endpoint

**Files:**
- Modify: `backend/internal/service/client/exam/export.go`
- Modify: `backend/internal/service/client/exam/export_test.go`
- Create: `backend/internal/modules/examrecordexportlink/transport/httppublic/handler.go`
- Create: `backend/internal/modules/examrecordexportlink/transport/httppublic/handler_test.go`
- Modify: `backend/internal/routes/v2/client/routes.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Regenerate: `backend/docs/swagger/docs.go`
- Regenerate: `backend/docs/swagger/swagger.json`
- Regenerate: `backend/docs/swagger/swagger.yaml`

- [x] Write failing tests proving token-bound exports cannot change record ID or format.
- [x] Extract an internal exact-record export method while keeping the Admin scoped method unchanged.
- [x] Register `GET /api/v2/public/exam-record-exports/download?token=...` without client login middleware.
- [x] Return attachment headers on success, 410 for invalid/expired/deleted records, and a redacted 500 response for renderer errors.
- [x] Regenerate Swagger and run route, handler and export tests.

### Task 4: Render `{answerDownloadUrl}` in post-submit notifications

**Files:**
- Modify: `backend/internal/service/client/poststat/exam_service.go`
- Modify: `backend/internal/service/client/poststat/exam_service_test.go`
- Create: `backend/internal/service/client/poststat/exam_download_link.go`
- Modify: `backend/cmd/main.go`
- Modify: `backend/cmd/main_safety_test.go`

- [x] Write a failing test where a selected template contains `{answerDownloadUrl}` and the notification payload must contain the issued URL.
- [x] Parse format, expiry days and public URL from exam settings using legacy-safe defaults.
- [x] Configure the Redis issuer at startup and issue one grant per submitted record only when at least one rule needs the placeholder.
- [x] Keep manual-grading placeholder behavior and replace `{answerDownloadUrl}` in the same render pass.
- [x] Verify templates without the placeholder do not touch Redis.

### Task 5: Expose the template variable in Admin

**Files:**
- Modify: `admin/src/views/exam/components/ExamPostStatSettings.vue`
- Modify: `admin/src/views/exam/template/index.vue`
- Modify: `admin/scripts/check-exam-designer-ux.mjs`
- Modify: `admin/scripts/check-survey-template-page.mjs`
- Modify: `backend/internal/service/admin/formkitadmin/notification_templates_test.go`
- Modify: `docs/API_V2.md`

- [x] Add failing contract checks for `{answerDownloadUrl}` in both exam template editing surfaces.
- [x] Add the variable help text without adding it to survey templates.
- [x] Document format, expiry, Redis loss semantics and the public endpoint.
- [x] Run focused Admin checks and `npm run build`.

### Task 6: Final verification

- [x] Run `gofmt` on changed Go files and `git diff --check`.
- [x] Run focused Backend packages for exam export, poststat, public route and Redis grants.
- [x] Run `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./... -count=1`; report unrelated baseline failures separately.
- [x] Run `cd admin && npm run check:exam-designer-ux && npm run check:survey-template-page && npm run build`.
- [x] Run `cd admin && npm run check:all`; report the known component-complexity failure separately if unchanged.
- [x] Verify in a browser that format and expiry persist after reload and the template help exposes `{answerDownloadUrl}`.

No commits or pushes are included unless explicitly requested.
