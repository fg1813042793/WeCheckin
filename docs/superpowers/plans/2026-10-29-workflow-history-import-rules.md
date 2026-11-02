# Workflow History Import Rules Implementation Plan

**Goal:** Configure import presets in Admin and apply them directly in H5 while preserving the unconfigured legacy wizard.

**Architecture:** Extend the existing definition-level live start config with historyImport. Backend validates and exposes it; Admin provides a focused rule editor; H5 reuses the existing normalization/permission-aware preview engine for direct application. No new database table or endpoint.

**Tech stack:** Go workflowcore/Hertz, Vue 3/Element Plus Admin, uni-app/TypeScript/uView Pro H5.

User authorized this feature. Execute autonomously in the shared checkout, preserve unrelated changes, no commits or production operations. Contract is in `docs/superpowers/specs/2026-10-29-workflow-history-import-rules-design.md`.

- [x] Backend: optional types, deep cloning, live config roundtrip/revision, strict validation, published DTO and compatibility tests. No schema migration or new endpoint.
- [x] Admin: shared types/serialization, focused rule editor in section05 of WorkflowStartConfig, normal/detail column mappings, defaults and readonly checks. Host component shrank within its existing budget.
- [x] H5: preset helper/composable, legacy fallback, safe field normalization, direct/latest import, skip feedback and async cancellation; zh/en text and runtime tests. Related retry-event bug fixed and covered.
- [x] Integration: same_key ignores inactive manual mappings consistently across all three ends; Swagger regenerated; API/use docs added. Browser fixtures cover Admin save → H5 import and all three viewport categories.
- [x] Verification: scoped tests/builds passed; full-check failures below were recorded, not treated as success. No production data changed or real workflow submitted.

## Verification record — 2026-10-29

- Backend passed `GOCACHE=$PWD/../.cache/go-build GOPROXY=off GOTOOLCHAIN=local go test ./internal/workflowcore ./internal/service/admin/workflow ./internal/modules/workflow/... ./cmd`, and race checks for core/service/application/infrastructure/httpclient.
- Admin passed `check:workflow-import-rules`, `check:workflow-runtime-pages`, related workflow checks and `npm run build`.
- H5 used cached project-pinned pnpm 10.34.5. Passed history-import helper/runtime checks (including retry recovery), type-check, UI-style checks, targeted ESLint on modified files, and build:h5.
- Full H5 lint retains 213 existing errors / 2 warnings. `check:workflow-module` stops at the existing WorkflowCenter “发起审批” text assertion. Admin check:all stops at the existing navigation `props.depth > 0` assertion; an existing unrelated user component budget also remains over limit.
- Full backend tests retain existing DingTalk-settings and tracked production-config failures, plus a concurrent notification-record route missing audit classification. A sandbox-blocked local HTTP test passed when its package was rerun with local-port access. None of these failures were introduced into this feature as workarounds.
- Browser simulated Admin enabled and saved scalar + one-column detail mapping. H5 selection immediately filled only mapped targets, leaving the unmapped weight blank; no POST workflow submission occurred. Latest mode requested page1/pageSize1 with scope=started and status=completed, then filled configured targets. Legacy mode completed select → configure → preview → confirm as before.
- Desktop 1440px, compact PC 800px and phone390px checked. A failed list read recovered via retry; no-history latest mode showed an explicit message; closing during a delayed detail request left the form unchanged; direct import succeeded on phone. Final captures waited for popup animation to settle. No browser page errors observed.
- Preview requests used synthetic local APIs. Production deployment and DingTalk native/App-device behavior are not verified. Deploy Backend, Admin and H5App together, then reopen the form to load current rules; no database migration is needed.
