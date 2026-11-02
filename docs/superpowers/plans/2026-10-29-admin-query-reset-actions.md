# Admin Query / Reset Actions

**Goal:** Every Admin filter that already provides query and reset uses a blue 查询 button followed by a white bordered 重置 button. The group follows the last input with a small gap, never independently right-aligned, and wraps as one unit when needed.

**Architecture:** Add a small `AdminFilterActions` component for order, presentation and click/submit behavior. Migrate existing paired actions without adding reset to search-only interfaces. Put default AdminSearchBar actions inside the field flow, and remove explicit right alignment in user/audit filter layouts. Keep actual filter handlers and unrelated page actions intact.

**Stack:** Vue 3, Element Plus, existing Admin UI tokens. No API or backend changes.

- [x] Shared component and barrel export implemented; AdminSearchBar default actions join the field flow. Query first, reset second, no decorative icons, 32px controls and 8px gap. Native-submit buttons do not also emit a search click.
- [x] Independent filters migrated: DingTalk bindings/reviews/history; workflow definitions/instances/tasks; scheduled tasks/runs; notification records/Webhook/workflow delivery.
- [x] Custom user/audit/survey/exam layouts migrated with inline placement and wrapping. Feedback uses the shared search bar. The template manager's single query button also uses primary styling and follows its input, without adding an unsolicited reset action.
- [x] Validation: shared runtime/coverage checks; affected page checks; Admin check:all/build; isolated browser query/reset checks. No production data operated on.
- [x] UI guidelines updated, unrelated and concurrent working-tree changes preserved. No commits or deployment.

## Validation evidence

- `npm run check:filter-actions` verifies rendered query/reset order, native-submit behavior, loading/disabled guards and Vue page coverage.
- Passed relevant `check:admin-ui-contract`, `check:p2-ui`, `check:user-list-ui`, `check:audit-logs`, `check:in-app-notification`, `check:scheduled-task`, `check:user-feedback` and `check:workflow-form-revision`. The last check now follows the existing extracted notification-style module instead of the obsolete wrapper location.
- `npm run build` passed. `npm run check:all` still stops at the existing navigation assertion `props.depth > 0`.
- Browser checked all 14 page routes with paired filter actions at 1920px, plus Webhook/workflow notification tabs. Normal buttons measured 60×32px with an 8px internal gap and primary/white backgrounds. Groups followed the last input, with no independent right alignment.
- User, log, feedback, survey, exam and scheduled-run filters were checked at 1024px and 390px; paired buttons stayed on one line within the viewport. Shared filter pages were rechecked after concurrent padding/accessibility changes.
- Survey and exam Enter search each generated exactly one list request. Reset cleared the entered keyword and generated exactly one request. Synthetic API responses were used; all writes were blocked. No browser page errors observed.
