# Admin Advanced Rich Text Editor Design

## Goal

Replace the question-bank and exam-designer advanced edit dialogs' generated Quill toolbars with a reusable Vue toolbar and editor foundation that is clear today and extensible for future rich-text features.

## Scope

The first complete version includes:

- paragraph and heading levels;
- font family and font size;
- bold, italic, underline and strike-through;
- text color and background color;
- ordered and unordered lists, indentation and alignment;
- blockquote and syntax-highlighted code blocks with a language selector;
- links and local/URL images;
- undo, redo, full-screen and clear formatting;
- basic tables with insert row/column and delete row/column/table operations;
- editable HTML source mode with sanitization before applying changes.

The table feature deliberately excludes cell merge and split. Collaborative editing, server-side image upload and migrating every existing Quill editor are outside this change.

## Architecture

Create a reusable `admin/src/components/rich-text/` boundary:

- `AdvancedRichTextEditor.vue` owns Quill, selection state, editor/source switching, full-screen state and command execution.
- `RichTextToolbar.vue` renders controls and emits semantic commands without knowing Quill internals.
- `richTextConfig.ts` registers the supported font/size attributors, syntax languages and editor module options.
- `sanitizeRichText.ts` applies DOMPurify before source HTML enters Quill.
- `rich-text-content.css` defines stable code-block, table, image and blockquote presentation.

`QuestionRichEditorDialog.vue` remains responsible only for question-bank dialog lifecycle and confirmation. The exam designer's `formkit/QuestionPreview.vue` uses the same shared editor for both title and option advanced-edit dialogs. Both paths keep their existing HTML string contracts, so saved question data and API contracts do not change.

## Interaction Design

- Familiar format controls are icon buttons with tooltips and `aria-label`/`aria-pressed` state.
- Font, size, heading, code language and alignment use compact selects or menus.
- Active formats use a visible primary-color background, border and icon/text color; code block activation must remain visible even when the selection is collapsed.
- Clear formatting is a labeled eraser action rather than an unexplained `T x` glyph.
- Table insertion asks for rows and columns, each limited to 1-10. Row/column mutation commands are enabled only while the cursor is inside a table.
- Source mode keeps a draft. "Apply source" sanitizes and updates the editor; "Cancel source" discards draft changes.
- Full-screen affects only the reusable editor and preserves the dialog's pending content.

## Safety And Compatibility

- DOMPurify removes scripts, embedded documents, event handlers and unsafe URLs before source HTML is applied.
- Quill remains the canonical HTML normalizer after sanitization.
- Existing saved HTML remains editable without migration.
- Font, size and colors use inline style attributors so rendered content does not depend on toolbar-only classes.
- Code blocks use Quill 2 syntax support with explicitly registered highlight.js languages.
- Tables use Quill 2's built-in table module. No Quill 1 table plugin is introduced.

## Verification

- Extend the question-bank component contract check before implementation so the missing reusable editor, safety layer and commands fail first.
- Run the focused question-bank and exam-designer checks.
- Run the Admin TypeScript production build and bundle budget check.
- Exercise the advanced edit dialog in a browser: active code block, clear formatting, table operations, source sanitization and full-screen mode.

