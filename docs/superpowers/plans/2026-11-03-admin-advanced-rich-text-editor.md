# Admin Advanced Rich Text Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a reusable, safe and extensible rich-text editor for question advanced editing with a custom toolbar, syntax highlighting, tables and HTML source mode.

**Architecture:** Keep Quill orchestration inside a shared rich-text component, isolate toolbar presentation from commands, and sanitize source HTML at the component boundary. The existing question dialog keeps its data contract and only swaps the generated Quill editor for the shared component.

**Tech Stack:** Vue 3, TypeScript, Element Plus, Quill 2, `@vueup/vue-quill`, highlight.js, DOMPurify.

---

### Task 1: Lock The Editor Contract And Dependencies

**Files:**
- Modify: `admin/scripts/check-question-bank-components.mjs`
- Modify: `admin/package.json`
- Modify: `admin/package-lock.json`

- [x] Add contract assertions for the shared editor, toolbar, sanitization, source mode, syntax, table commands, clear formatting and active code-block styling.
- [x] Run `node scripts/check-question-bank-components.mjs` and confirm it fails because the new editor files and contracts do not exist.
- [x] Add `dompurify` and `highlight.js` using npm so the existing package changes are preserved and the lockfile is updated mechanically.

### Task 2: Add Safe Rich-Text Foundations

**Files:**
- Create: `admin/src/components/rich-text/richTextConfig.ts`
- Create: `admin/src/components/rich-text/sanitizeRichText.ts`
- Create: `admin/src/components/rich-text/rich-text-content.css`

- [x] Register style-based font and size whitelists, selected highlight.js languages, Quill syntax and table module options.
- [x] Add a DOMPurify wrapper that keeps supported rich-text structure while removing executable and embedded content.
- [x] Add reusable presentation styles for code blocks, tables, blockquotes, links and images.

### Task 3: Build The Custom Toolbar And Editor

**Files:**
- Create: `admin/src/components/rich-text/RichTextToolbar.vue`
- Create: `admin/src/components/rich-text/AdvancedRichTextEditor.vue`

- [x] Render grouped controls with tooltips, accessible names, visible active/disabled states and responsive wrapping.
- [x] Track Quill selection/format changes and map toolbar commands to Quill formatting, history, links, images and clear formatting.
- [x] Implement code-language selection and visible code-block content styling.
- [x] Implement 1-10 row/column table insertion and cursor-sensitive table mutation commands.
- [x] Implement sanitized source editing and full-screen mode without changing the external HTML model contract.

### Task 4: Integrate The Question Advanced Editor

**Files:**
- Modify: `admin/src/views/question-bank/components/QuestionRichEditorDialog.vue`
- Modify: `admin/src/views/question-bank/questionBankEditor.ts`
- Modify: `admin/src/views/exam/formkit/QuestionPreview.vue`
- Modify: `admin/scripts/check-exam-designer-ux.mjs`

- [x] Replace the generated Quill editor in the advanced dialog with `AdvancedRichTextEditor`.
- [x] Replace the exam designer title and option advanced editors with `AdvancedRichTextEditor` and remove their duplicate toolbar/full-screen implementation.
- [x] Keep compact title/option editor options unchanged and remove the unused full-toolbar configuration.
- [x] Preserve dialog cancel, close and confirm behavior.

### Task 5: Verification

**Files:**
- Verify all changed Admin files without modifying unrelated work.

- [x] Run `node scripts/check-question-bank-components.mjs` and `node scripts/check-exam-designer-ux.mjs`.
- [x] Run `npm run build` and `npm run check:bundle`.
- [x] Run `git diff --check`.
- [x] Verify the advanced edit flow in the browser at desktop width, including code-block active state, clear formatting, table commands, source sanitization and full-screen mode.

