# Admin Tooltip Maximum Width Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply a responsive 420px maximum width and safe text wrapping to every Element Plus Tooltip in the Admin application.

**Architecture:** Extend the existing global Admin stylesheet with a selector limited to dark Element Plus Popper tooltips. Protect the rule through the existing UI shell contract check, then verify a long table overflow tooltip and a short icon tooltip in the running browser.

**Tech Stack:** Vue 3, Element Plus 2.9, CSS, Node.js static contract checks, Vite.

---

### Task 1: Add a failing global Tooltip contract

**Files:**
- Modify: `admin/scripts/check-ui-shell.mjs`
- Test: `admin/scripts/check-ui-shell.mjs`

- [ ] **Step 1: Require the global Tooltip selector and declarations**

Extend the script to require these exact snippets in `src/styles/admin.css`:

```js
const requiredTooltipStyles = [
  'body .el-popper.is-dark',
  'max-width: min(420px, calc(100vw - 32px));',
  'white-space: normal;',
  'overflow-wrap: anywhere;',
  'line-height: 1.5;',
]
```

- [ ] **Step 2: Run the check and verify RED**

Run: `cd admin && node scripts/check-ui-shell.mjs`

Expected: FAIL because the global Tooltip rule is absent.

### Task 2: Add the global Tooltip width and wrapping rule

**Files:**
- Modify: `admin/src/styles/admin.css`
- Test: `admin/scripts/check-ui-shell.mjs`

- [ ] **Step 1: Add the minimal global CSS**

Add:

```css
body .el-popper.is-dark {
  max-width: min(420px, calc(100vw - 32px));
  white-space: normal;
  overflow-wrap: anywhere;
  line-height: 1.5;
}
```

Do not set `width`, `min-width`, or `word-break: break-all`; short Tooltip content must retain its natural width.

- [ ] **Step 2: Run UI shell and build checks**

Run:

```bash
cd admin
node scripts/check-ui-shell.mjs
npm run build
npm run check:bundle
```

Expected: PASS, with unrelated repository baseline failures reported separately.

### Task 3: Browser verification

**Files:**
- No repository changes.

- [ ] **Step 1: Verify a long overflow Tooltip**

Open the survey response table, hover the device information field, and verify the Tooltip width is at most 420px, wraps to multiple lines, and contains no horizontal overflow.

- [ ] **Step 2: Verify a short Tooltip**

Hover a short icon or button Tooltip and verify its width remains content-driven rather than expanding to 420px.

- [ ] **Step 3: Verify a narrow viewport**

At a viewport narrower than 452px, verify the Tooltip stays within `viewport - 32px`.

- [ ] **Step 4: Review the patch**

Run: `git diff --check -- admin/src/styles/admin.css admin/scripts/check-ui-shell.mjs`

Expected: no output.
