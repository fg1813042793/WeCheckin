# Admin log route reentry repair

**Goal:** Return to the operation/login log page through menus and tabs without losing the list or breaking Vue page activation.

**Observed root cause:** The layout treated RouterView's rendered VNode as the component identity, created a new wrapper for each new VNode, and rendered a captured, previously mounted VNode inside that wrapper. With KeepAlive, revisiting produced `parentComponent.ctx.deactivate is not a function`. The log page had cleared its rows on deactivation but no new list request was issued.

## Change

- Extract the per-layout route wrapper cache into `admin/src/router/routeComponentCache.ts`.
- Keep wrapper identity stable for the actual component type and full path. Supply the latest RouterView node through the declared `routeNode` prop and construct a fresh VNode from its type, props and children, excluding mounted instance and KeepAlive bookkeeping.
- Use RouterView's own matched route for wrapper selection and keys. Preserve existing generated cache names, menu-controlled caching and tab eviction behavior.
- Leave the audit controller's cancellation, filtering, permission checks and reactivation loading semantics intact.

## Verification

- Before the fix, browser mock data showed the first list successfully; menu navigation away/back showed zero rows, no second list request and the Vue cache error above.
- `npm run check:route-cache` uses the actual layout RouterView template, a real Vue custom renderer, Vue Router and KeepAlive, together with the real audit controller. It covers layout rerenders, repeated activation, login classification, late-response cancellation, closed tabs and disabled caching. No lifecycle hooks are stubbed.
- `npm run check:route-tabs`, `npm run check:audit-logs`, `npm run check:ui-shell` and `npm run build` passed.
- Browser verification covers three menu round trips, preserving the login tab, closing/reopening a tab, leaving during a delayed request, and rerendering the layout through sidebar collapse. Synthetic API data only; no production records changed.
- Final browser runs showed data on each return and no page errors or console warnings/errors. The regression harness compiles the actual layout template so wrapper-prop casing and RouterView wiring are tested alongside lifecycle behavior.
- Full `npm run check:all` remains blocked by the existing navigation assertion `props.depth > 0`; this failure is not reported as passing.

Existing unrelated changes are preserved. No commit or deployment is included; an already-open browser page needs a refresh to discard its old cache instance.
