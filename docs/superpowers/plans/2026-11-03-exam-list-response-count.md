# Exam List Response Count Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show the number of submitted answer sheets for each exam in Admin card and table views.

**Architecture:** Reuse the survey list count pattern. The exam service batch-counts submitted records for the current page with one grouped query, the admin DTO exposes `responseCount`, and the existing list response renders that field without extra frontend requests.

**Tech Stack:** Go, GORM, Hertz handlers, Vue 3, TypeScript, Element Plus, Node contract checks.

---

### Task 1: Add the backend list count contract

**Files:**
- Create: `backend/internal/service/client/exam/admin_list_response_count_structure_test.go`
- Modify: `backend/internal/service/client/exam/service.go`
- Modify: `backend/internal/handler/admin/exam/dto.go`
- Modify: `backend/internal/handler/admin/exam/handler.go`

- [x] **Step 1: Write the failing structure test**

Assert that the exam service defines `ResponseCountsContext`, filters with ``exam_r_status >= 1``, groups by ``exam_r_exam_id``, and that the admin handler loads one count map for the current page before building DTOs.

- [x] **Step 2: Run the test and verify RED**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/service/client/exam -run TestAdminExamListBatchLoadsSubmittedResponseCounts -count=1`

Expected: failure because the batch count method and DTO field do not exist.

- [x] **Step 3: Implement the batch count**

Add a service method equivalent to:

```go
func (s *Service) ResponseCountsContext(ctx context.Context, examIDs []uint) (map[uint]int, error) {
    countMap := make(map[uint]int, len(examIDs))
    if len(examIDs) == 0 { return countMap, nil }
    // SELECT exam_r_exam_id, COUNT(*) AS cnt
    // WHERE exam_r_exam_id IN ? AND exam_r_status >= 1
    // GROUP BY exam_r_exam_id
    return countMap, nil
}
```

Add `ResponseCount int json:"responseCount"` to `examListItem`, pass the count into `newExamListItem`, and make the list handler collect current-page IDs and call the batch method once. Copy responses pass zero because the copied exam has no records.

- [x] **Step 4: Run the focused backend tests and verify GREEN**

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/service/client/exam ./internal/handler/admin/exam -count=1`

Expected: PASS.

### Task 2: Render the answer count in Admin

**Files:**
- Modify: `admin/scripts/check-exam-management.mjs`
- Modify: `admin/src/views/exam/ExamList.vue`

- [x] **Step 1: Extend the Admin contract check and verify RED**

Require `row.responseCount`, the card label `已交答卷`, and table column label `答卷数` in `check-exam-management.mjs`.

Run: `cd admin && npm run check:exam-management`

Expected: failure because the list page does not render the count yet.

- [x] **Step 2: Add the minimal UI**

Add a third card preview row:

```vue
<div><span>已交答卷</span><strong>{{ row.responseCount || 0 }} 份</strong></div>
```

Add an `答卷数` table column displaying the same value. Keep existing response/statistics actions unchanged.

- [x] **Step 3: Run the focused Admin check and build**

Run: `cd admin && npm run check:exam-management`

Run: `cd admin && npm run build`

Expected: PASS.

### Task 3: Final verification

**Files:**
- Verify only; do not edit generated output.

- [x] **Step 1: Run format and focused checks**

Run: `git diff --check`

Run: `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/service/client/exam ./internal/handler/admin/exam -count=1`

Run: `cd admin && npm run check:exam-management && npm run build`

- [x] **Step 2: Browser verification**

Open `/exam/list` with a mocked list response containing `responseCount: 12`; verify the card shows `已交答卷 12 份`, the table view shows `答卷数`, and the console contains no relevant errors.

No commit is created because the user has not requested one.

