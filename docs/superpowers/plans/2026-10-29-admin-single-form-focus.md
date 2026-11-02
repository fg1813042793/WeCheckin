# Admin Single Form Focus Implementation Plan

**Goal:** Standardize Admin form focus states to one thin blue border.

### Task 1: Contract
- [x] Update UI shell checks to require single-ring page and teleported form controls.
- [x] Update question-bank checks to forbid rich-editor focus shadows.
- [x] Run checks and confirm RED.

### Task 2: Global standard controls
- [x] Replace double shadows in `admin/src/styles/admin.css` with single inset borders.
- [x] Cover page forms, overlay dialogs/drawers, and picker panels.

### Task 3: Rich editors
- [x] Remove focus shadows from both question-bank rich-editor implementations.

### Task 4: Verification
- [x] Run focused checks, Admin build, and full Admin checks.
- [x] Browser-check representative page, dialog, select, textarea, date, and rich-editor states.
- [x] Review final diff and preserve unrelated changes.

