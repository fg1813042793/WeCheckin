# Workflow Textarea Visible Rows Design

## Goal

Allow workflow designers to configure the visible height range of every textarea, including detail-list textarea columns, and make Admin and H5 runtime forms render the same behavior.

## Data Contract

Textarea fields may contain `minVisibleRows` and `maxVisibleRows`. These values only control presentation and do not participate in required, length, or detail-list row validation. Existing `minRows` and `maxRows` remain exclusive to detail-list data rows.

Configured values must be positive integers, `minVisibleRows <= maxVisibleRows`, and `maxVisibleRows <= 30`. Both values are omitted for non-textarea fields. Existing definitions without these properties use client defaults: `3-8` rows for normal fields and `2-6` rows for detail-list columns.

## Designer

The textarea field property panel shows `最小显示行数` and `最大显示行数`. Detail-list textarea columns show the same controls in the column editor. New fields receive defaults, type changes add or remove the properties as appropriate, and previews use the configured minimum rows.

## Runtime

Admin uses Element Plus textarea autosize with the configured bounds. H5 returns to uView `u-textarea`, enables automatic height, caps the field at the configured maximum height, and uses uView's built-in count position. Content beyond the maximum height scrolls inside the textarea.

## Compatibility

The fields live inside the existing workflow definition JSON, so no database migration is required. The backend struct and schema validation preserve and validate the properties when draft definitions are saved and published.

## Verification

Backend tests cover valid, invalid, nested, and detail-column definitions. Admin and H5 structure checks cover property controls, defaults, runtime binding, count rendering, and removal of the failed resize implementation. Type checks, focused lint, Go tests, and H5/Admin builds complete the verification.
