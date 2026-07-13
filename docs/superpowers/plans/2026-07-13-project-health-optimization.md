# Project Health Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the existing project baseline cleaner by fixing the formkit validation regression, aligning startup/docs with current configuration, and preventing future runtime artifacts from polluting the working tree.

**Architecture:** Keep changes narrowly scoped to the existing files. Preserve the current backend, admin, and uni-app structure; do not restructure modules or touch unrelated survey/logic-engine work-in-progress files. Treat the existing formkit test as the source of truth for `location` validation.

**Tech Stack:** Go 1.24 module, Hertz backend, Vue 3/Vite admin, uni-app frontend, shell startup script, Markdown docs.

---

### Task 1: Reproduce Location Validation Failure

**Files:**
- Test: `backend/internal/app/formkit/question/builtin/builtin_test.go`
- Read: `backend/internal/app/formkit/question/builtin/location_advanced.go`

- [ ] **Step 1: Run the focused failing test**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/formkit/question/builtin -run TestLocation_Validate -count=1
```

Expected: FAIL with `non-map should fail`.

- [ ] **Step 2: Confirm implementation mismatch**

Read `backend/internal/app/formkit/question/builtin/location_advanced.go` and confirm `LocationQuestion.Validate` accepts non-empty string values. This explains the test failure.

### Task 2: Fix Location Validation

**Files:**
- Modify: `backend/internal/app/formkit/question/builtin/location_advanced.go`
- Test: `backend/internal/app/formkit/question/builtin/builtin_test.go`

- [ ] **Step 1: Keep the existing failing test unchanged**

The existing test is already the RED test:

```go
if err := inst.Validate("plain string", q); err == nil {
	t.Error("non-map should fail")
}
```

- [ ] **Step 2: Update validation to reject plain strings**

Remove the branch that treats non-empty strings as valid answers. Keep `nil` behavior unchanged, keep map validation, and return the existing type error for any non-map value:

```go
m, ok := value.(map[string]interface{})
if !ok {
	return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "位置题答案必须为对象"}
}
```

- [ ] **Step 3: Run the focused test again**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/formkit/question/builtin -run TestLocation_Validate -count=1
```

Expected: PASS.

### Task 3: Align Backend Startup Script

**Files:**
- Modify: `backend/start.sh`

- [ ] **Step 1: Update config file check**

Change the script's config check from `config.yaml` to `config/config.yaml` because the backend config files live under `backend/config/`.

- [ ] **Step 2: Update startup URL**

Change the printed startup URL from `http://localhost:8080` to `http://localhost:8083`, matching `backend/config/config.yaml`.

- [ ] **Step 3: Verify script text**

Run:

```bash
rg -n "config/config.yaml|localhost:8083|config.yaml|localhost:8080" backend/start.sh
```

Expected: `config/config.yaml` and `localhost:8083` appear; stale `8080` does not appear.

### Task 4: Update Repository Hygiene Rules

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: Add local cache ignore**

Add:

```gitignore
.cache/
```

- [ ] **Step 2: Add backend runtime artifact ignores**

Add:

```gitignore
backend/logs/
backend/uploads/
backend/bin/
backend/cmd/cmd
cmd.exe
```

- [ ] **Step 3: Verify ignored status for representative files**

Run:

```bash
git check-ignore --no-index backend/logs/access-2026-06-30-2026-06-30T23-59-59.909.log.gz backend/uploads/export_enroll_1.csv backend/bin/app cmd.exe .cache
```

Expected: each listed path is printed.

### Task 5: Refresh README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update tech stack**

Describe:

```markdown
- Backend: Go + Hertz + GORM + MySQL + Redis
- Admin: Vue 3 + Vite + Element Plus
- Frontend: uni-app + Vue 3
```

- [ ] **Step 2: Update run commands**

Document these commands:

```bash
cd backend && go run cmd/main.go
cd admin && npm run dev
cd frontend && npm run dev:h5
```

- [ ] **Step 3: Mention default backend port**

Document that the default backend config currently listens on `8083`.

### Task 6: Verify Health Pass

**Files:**
- Read: working tree status

- [ ] **Step 1: Run formkit tests**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/formkit/...
```

Expected: all formkit packages pass.

- [ ] **Step 2: Remove test cache directory if created**

Run:

```bash
rm -rf .cache
```

Expected: `.cache` no longer appears in `git status --short`.

- [ ] **Step 3: Review final diff**

Run:

```bash
git diff -- .gitignore README.md backend/start.sh backend/internal/app/formkit/question/builtin/location_advanced.go docs/superpowers/plans/2026-07-13-project-health-optimization.md
```

Expected: only scoped health optimization changes appear.
