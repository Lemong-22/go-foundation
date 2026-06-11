# Week 1-2 — Phase A: Foundation + CRUD

> **Goal:** Lo bisa Go + Postgres end-to-end. Bikin `course-cli` minimal yang bisa CRUD Course + Lesson, via CLI dan REST. Code masih procedural — refactor ke hexagonal di Week 3.

**Architecture mode:** **Procedural dulu, hexagonal belakangan.** Belajar sintaks & flow dulu, baru abstraksi. Ini deliberate pedagogical choice.

**Reference:** `vendor/entropy-course/docs/spec.md` §1-3 (CLI commands & data model — pakai spec mentor sebagai target).

---

## Day-by-Day Breakdown

### Day 1 (2026-06-11) — Setup ✅ DONE
- Install Go 1.22.2, baca docs, vendor entropy-course
- Init repo, folder structure
- Commit: `chore: project skeleton`

### Day 2 (2026-06-12) — Hello World + Postgres Setup
**Goal:** Lo yakin Go jalan & Postgres bisa diakses.

**Tasks:**
1. Setup DB `go_foundation` di canister23
2. `cmd/hello/main.go` — print "Hello from Go"
3. `cmd/hello/main.go` — HTTP version: `GET /` return "OK"
4. Test pakai `curl`
5. `cmd/pgping/main.go` — connect ke Postgres, `SELECT 1`
6. Verify: `go run ./cmd/pgping` print "connected"

**Files:**
- `cmd/hello/main.go` (new)
- `cmd/pgping/main.go` (new)
- `scripts/setup-db.sh` (new — DB bootstrap)

**Commit:** `feat: hello world + postgres ping`

---

### Day 3 (2026-06-13) — First Entity: Course (in-memory)
**Goal:** Bikin entity + handler + CLI command — tanpa DB dulu.

**Tasks:**
1. `internal/course/types.go` — `Course` struct dengan `ID`, `Title`, `Slug`, `Description`, `Status`, timestamps
2. `cmd/course-cli/main.go` — Cobra root command
3. `cmd/course-cli/course.go` — `create`, `list`, `get` subcommands (in-memory slice)
4. Verify: `course-cli course create --title "Intro to Go" --slug "intro-go"`
5. Verify: `course-cli course list`

**Files:**
- `internal/course/types.go` (new)
- `cmd/course-cli/main.go` (new)
- `cmd/course-cli/course.go` (new)

**Commit:** `feat: course-cli scaffold with in-memory store`

---

### Day 4 (2026-06-14) — Postgres Schema + Repository
**Goal:** Course persist ke Postgres.

**Tasks:**
1. `migrations/000001_create_courses.up.sql` — schema `courses` table
2. `migrations/000001_create_courses.down.sql` — drop table
3. `internal/course/repository.go` — `CourseRepository` interface (Save, FindByID, FindAll, Delete)
4. `internal/course/repository_postgres.go` — `PostgresCourseRepository` impl pakai pgx
5. Update `cmd/course-cli/course.go` — use real repo
6. Verify: `course-cli course create ...` → row di Postgres

**Files:**
- `migrations/000001_create_courses.up.sql` (new)
- `migrations/000001_create_courses.down.sql` (new)
- `internal/course/repository.go` (new)
- `internal/course/repository_postgres.go` (new)
- `cmd/course-cli/course.go` (modify)

**Commit:** `feat: postgres repository for course`

---

### Day 5 (2026-06-15) — Lesson Entity + CRUD
**Goal:** Lesson = child of Course. Schema + repo + CLI.

**Tasks:**
1. `migrations/000002_create_lessons.up.sql` — `lessons` table dengan `course_id` FK
2. `internal/course/lesson.go` — `Lesson` struct
3. Extend `CourseRepository` atau bikin `LessonRepository`
4. `cmd/course-cli/lesson.go` — `create`, `list`, `get`, `delete` subcommands
5. Verify: bikin course, tambah 3 lessons, list, delete

**Files:**
- `migrations/000002_create_lessons.up.sql` (new)
- `internal/course/lesson.go` (new)
- `internal/course/repository.go` (modify)
- `internal/course/repository_postgres.go` (modify)
- `cmd/course-cli/lesson.go` (new)

**Commit:** `feat: lesson entity and CRUD`

---

### Day 6 (2026-06-16) — REST Adapter (net/http)
**Goal:** Course + Lesson bisa diakses via HTTP, pakai DTO yang sama dengan CLI.

**Tasks:**
1. `cmd/course-api/main.go` — HTTP server
2. `internal/course/http_handlers.go` — `POST /courses`, `GET /courses`, `GET /courses/{id}`, etc.
3. Test pakai `curl` + `httptest`
4. Verify: `curl -X POST ...` bikin course yang sama dengan `course-cli course create`

**Files:**
- `cmd/course-api/main.go` (new)
- `internal/course/http_handlers.go` (new)
- `internal/course/http_handlers_test.go` (new)

**Commit:** `feat: REST adapter for course and lesson`

---

### Day 7 (2026-06-17) — Integration Test + Polish
**Goal:** CLI bikin course → REST GET balik data sama.

**Tasks:**
1. `tests/integration_test.go` — full flow
2. Add `--output json` ke CLI commands (compare dengan entropy-course spec §8)
3. Add proper exit codes (compare dengan entropy-course spec §9)
4. Add error handling: 404 not found, 400 validation, 500 internal
5. Journal entry + retrospective
6. Push ke GitHub
7. **STOP** — Week 2 udah kelar

**Files:**
- `tests/integration_test.go` (new)
- `cmd/course-cli/course.go` (modify)
- `cmd/course-cli/lesson.go` (modify)
- `cmd/course-api/main.go` (modify)
- `journal/2026-WEEK-23.md` (update)

**Commit:** `test: cli-rest parity integration test`, `docs: week 1 retrospective`

---

## Cross-cutting Rules

- **TDD**: test dulu, run gagal, baru code. Pakai Go `testing` stdlib (no testify week 1, biar hafal signature).
- **Commit tiap task selesai**, conventional commit format: `feat:`, `test:`, `chore:`, `docs:`, `refactor:`
- **YAGNI**: ga bikin field/command yang ga dipake. Misal: ga ada `status: archived` week 1.
- **Compare dengan entropy-course**: kalau bingung, baca `vendor/entropy-course/docs/spec.md` atau `internal/course/usecase/course_service.go` untuk pattern.

## Exit Criteria (akhir Week 2)

- [ ] `course-cli course create/list/get/update/delete` works (Postgres-backed)
- [ ] `course-cli lesson create/list/get/update/delete` works (Postgres-backed)
- [ ] `POST /courses`, `GET /courses`, `GET /courses/{id}` works
- [ ] `POST /courses/{id}/lessons`, `GET /courses/{id}/lessons` works
- [ ] Integration test: CLI bikin → REST baca → data sama
- [ ] Exit codes: 0 success, 1 validation, 2 not found, 5 internal
- [ ] `--output json` works
- [ ] README di `go-foundation/` updated
- [ ] 7 journal entries written

## After Week 2

Lo apply pattern yang udah dipelajari ke:
- **Week 3**: Refactor ke hexagonal (extract domain/ports/usecases/adapter)
- **Week 4**: Auth + ownership
- **Setelah Week 4**: Apply ke LE-REMINDER Phase 4

---

## Catatan tentang Plan Ini

Plan ini **compressing** dari spec mentor (`docs/spec.md`) yang punya 13 commands + publish/unpublish + reorder. Lo ga perlu semua di Week 1-2. Yang penting:
- Lo familiar dengan sintaks Go
- Lo familiar dengan pgx
- Lo familiar dengan Cobra
- Lo familiar dengan HTTP handler pattern
- Lo bisa CRUD 2 entities (Course + Lesson)

Publish/unpublish/reorder/admin commands → Week 3+ setelah hexagonal refactor.
