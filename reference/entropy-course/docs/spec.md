# Technical Spec — Course Management

Status: Draft · Architecture: simplified Hexagonal + Clean + DDD (no CQRS)
Source: `course-cli-prd.md` v0.1.0 · Spec author: design pass, 2026-05-24

## 0. Overview

- **Purpose:** manage the lifecycle of courses and their lessons — create, edit,
  publish/unpublish, sequence, and remove them — as the authoritative backend
  for a course platform, persisted in PostgreSQL.
- **Language / stack:** Go 1.22+, PostgreSQL via `pgx`, `cobra` + `viper` for
  the CLI, `google/uuid` for id generation.
- **In scope:** course CRUD plus publish/unpublish; lesson CRUD plus reorder.
  Within this single bounded context, **Course** and **Lesson** are modeled as
  two separate aggregates (per the design decision in §2).
- **Out of scope:**
  - Instructor / user management — a separate **Identity** bounded context.
    This context references an instructor only by id (`InstructorID`), never
    loading or owning the user record. The PRD's `admin instructor *` commands
    belong to that other context.
  - Database migrations (`migrate up/down/status`) — infrastructure tooling,
    not a domain activity. They ship in the same binary but are wired outside
    this bounded context.
  - Authentication and role enforcement (PRD §10) — deferred. See §6.
  - Archived course status, lesson publish/unpublish, student/enrollment
    features — deferred. See §6.

**On the CLI itself.** `cobra` command trees, flag parsing, `viper` config
resolution, the `table` / `json` / `quiet` formatters, destructive-command
confirmation prompts (`--force`), and process exit codes are all part of the
**inbound CLI adapter** — they are not domain logic. The core returns typed
domain errors; the CLI adapter is the only component that knows the exit-code
table (validation → 1, not found → 2, permission → 3, internal → 5).

---

## 1. CLI Commands

The activities a user can perform against this context. There are 13 commands;
each maps to **exactly one inbound port method and one usecase**.

### Course commands

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `course create` | Create a new course in `draft` status | `--title`, `--slug`, `--description` (opt), `--instructor-id` (scaffold — see note) | prints new course id | missing title, invalid slug, slug already taken |
| `course list` | List courses, newest first | `--status` (opt filter), `-o/--output` | table / JSON / id-only of courses | — |
| `course get` | Show one course's detail | `<course-id>`, `-o/--output` | course detail | id not found |
| `course update` | Edit course metadata | `<course-id>`, `--title` (opt), `--description` (opt), `--slug` (opt) | prints updated course id | id not found, invalid slug, slug already taken, nothing to update |
| `course delete` | Delete a course and all its lessons | `<course-id>`, `--force` (opt) | confirmation | id not found |
| `course publish` | Set course status to `published` | `<course-id>` | confirmation | id not found, already published |
| `course unpublish` | Set course status back to `draft` | `<course-id>` | confirmation | id not found, not currently published |

### Lesson commands

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `lesson create` | Add a lesson to a course | `--course-id`, `--title`, `--content` (opt), `--order` (opt) | prints new lesson id | course not found, missing title, negative order |
| `lesson list` | List all lessons in a course, by order | `--course-id`, `-o/--output` | table / JSON / id-only of lessons | course not found |
| `lesson get` | Show one lesson's detail | `<lesson-id>`, `-o/--output` | lesson detail | id not found |
| `lesson update` | Edit a lesson's title / content | `<lesson-id>`, `--title` (opt), `--content` (opt) | prints updated lesson id | id not found, nothing to update |
| `lesson delete` | Remove a lesson | `<lesson-id>`, `--force` (opt) | confirmation | id not found |
| `lesson reorder` | Resequence lessons within a course | `--course-id`, `--order <lesson-id:pos,...>` | confirmation | course not found, unknown lesson id, lesson not in course, duplicate position |

**Note on `--instructor-id`.** The `courses` table requires a non-null owning
instructor, but the PRD defers auth (§10), so there is no logged-in actor yet.
For this iteration the CLI adapter sources the instructor id from a
`--instructor-id` flag / `COURSE_CLI_INSTRUCTOR_ID` env var / config file, and
passes it into `CreateCourseInput`. When auth lands, the same input field is
populated from the auth token instead — no change to the inbound port. This is
the deliberate auth seam recorded in §6.

---

## 2. Domain Model

Two aggregates: **Course** and **Lesson**. Per the design decision, a Lesson is
its **own aggregate root** with its own repository, not a child entity of the
Course aggregate. A Lesson references its course by id only (`CourseID`) — the
standard DDD rule that one aggregate references another by identity, never by
holding the other object. This matches the PRD's lesson commands, which look a
lesson up directly by `<lesson-id>` rather than through its course.

### Entities

**Course** — a course owned by an instructor; the unit that gets published.
- Identity: `ID CourseID`
- Fields:
  - `Title string`
  - `Slug Slug`
  - `Description string` (optional; empty string when unset)
  - `InstructorID InstructorID` (id-only reference into the Identity context)
  - `Status CourseStatus`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Title` is non-empty after trimming.
  - `Slug` is always a valid slug (guaranteed by the VO).
  - `Status` is one of `draft` / `published` (guaranteed by the VO).
  - `UpdatedAt >= CreatedAt`.
- Behaviour (each mutating method takes the current time and bumps `UpdatedAt`;
  the entity holds no `Clock` — the timestamp is passed in by the usecase):
  - `Rename(title string, now time.Time) error` — re-validates non-empty.
  - `ChangeDescription(desc string, now time.Time)`.
  - `ChangeSlug(slug Slug, now time.Time)` — uniqueness is enforced by the
    usecase, not the entity (the entity cannot see other courses).
  - `Publish(now time.Time) error` — sets `Status` to `published`; errors with
    `ErrAlreadyPublished` if already published.
  - `Unpublish(now time.Time) error` — sets `Status` to `draft`; errors with
    `ErrNotPublished` if not currently published.

**Lesson** — a unit of content inside a course.
- Identity: `ID LessonID`
- Fields:
  - `CourseID CourseID` (id-only reference to the owning Course aggregate)
  - `Title string`
  - `Content string` (optional; empty string when unset)
  - `Order LessonOrder`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Title` is non-empty after trimming.
  - `Order` is `>= 0` (guaranteed by the VO).
  - `CourseID` is always set.
- Behaviour:
  - `Rename(title string, now time.Time) error` — re-validates non-empty.
  - `ChangeContent(content string, now time.Time)`.
  - `MoveTo(order LessonOrder, now time.Time)` — sets a new position.

### Value Objects

All value objects are immutable; construction (`NewX`) fails with a domain
error when the invariant is violated, so an existing VO is always valid.

- **CourseID** — wraps `string`; invariant: non-empty, parseable as a UUID.
- **LessonID** — wraps `string`; invariant: non-empty, parseable as a UUID.
- **InstructorID** — wraps `string`; invariant: non-empty, parseable as a UUID.
  Purely a cross-context reference — this context never dereferences it.
- **Slug** — wraps `string`; invariant: non-empty after trimming; matches
  `^[a-z0-9]+(?:-[a-z0-9]+)*$` (lower-case, digits, single hyphens, no leading/
  trailing hyphen). *Uniqueness is not a VO invariant* — it is a cross-row rule
  enforced by the usecase via the repository.
- **CourseStatus** — wraps `string`; invariant: one of `draft` | `published`.
  (`archived` is deferred — see §6.) Exposes `Draft()` / `Published()`
  constructors and an `IsPublished()` predicate.
- **LessonOrder** — wraps `int`; invariant: `>= 0`.

### Domain notes

- **Vertical slices.** Each of the 13 commands is its own slice through the
  stack: CLI → inbound port → usecase → outbound ports → adapter. No command
  shares a usecase with another.
- **`Title` stays a plain `string`.** A title has no invariant beyond
  non-empty, which the entity constructor enforces directly; wrapping it in a
  VO would be ceremony without payoff. `Description` and `Content` likewise
  stay plain strings (no invariant at all).
- **Cross-aggregate delete.** `course delete` must also remove the course's
  lessons. Because Course and Lesson are separate aggregates, this is *not* a
  silent DB cascade hidden from the domain — the `DeleteCourse` usecase
  explicitly orchestrates it across both repositories (delete lessons, then the
  course). The `lessons.course_id` FK may still carry `ON DELETE CASCADE` as a
  safety net, but the usecase owns the intent.
- **Timestamps are passed in, not pulled.** Entities never call a clock.
  Constructors and mutating methods receive `now time.Time`; the usecase gets
  it from the `Clock` outbound port. This keeps the domain pure and makes time
  deterministic in tests.
- **Domain errors.** The core returns typed sentinel errors — `ErrNotFound`,
  `ErrSlugTaken`, `ErrAlreadyPublished`, `ErrNotPublished`, and validation
  errors from VO constructors. The CLI adapter maps these to exit codes; the
  core itself knows nothing about exit codes.

---

## 3. Ports

### Inbound ports (driving — called by the CLI adapter)

One interface per noun, mirroring the CLI's `course` / `lesson` grouping. Every
method corresponds to exactly one command from §1.

```go
// CourseService is the inbound port for all course commands.
type CourseService interface {
    CreateCourse(in CreateCourseInput) (CreateCourseOutput, error)
    ListCourses(in ListCoursesInput) (ListCoursesOutput, error)
    GetCourse(in GetCourseInput) (GetCourseOutput, error)
    UpdateCourse(in UpdateCourseInput) (UpdateCourseOutput, error)
    DeleteCourse(in DeleteCourseInput) error
    PublishCourse(in PublishCourseInput) error
    UnpublishCourse(in UnpublishCourseInput) error
}

// LessonService is the inbound port for all lesson commands.
type LessonService interface {
    CreateLesson(in CreateLessonInput) (CreateLessonOutput, error)
    ListLessons(in ListLessonsInput) (ListLessonsOutput, error)
    GetLesson(in GetLessonInput) (GetLessonOutput, error)
    UpdateLesson(in UpdateLessonInput) (UpdateLessonOutput, error)
    DeleteLesson(in DeleteLessonInput) error
    ReorderLessons(in ReorderLessonsInput) error
}
```

Input / output DTOs carry only primitives and ids — they are the boundary
between the (untyped) CLI world and the (VO-typed) domain. Usecases build VOs
from these primitives and return validation errors when construction fails.

```go
// --- Course DTOs ---
type CreateCourseInput struct {
    Title        string
    Slug         string
    Description  string // "" when unset
    InstructorID string // sourced by the CLI adapter; auth token later
}
type CreateCourseOutput struct{ ID string }

type ListCoursesInput struct {
    Status string // "" = no filter; otherwise "draft" | "published"
}
type ListCoursesOutput struct{ Courses []CourseView }

type GetCourseInput struct{ ID string }
type GetCourseOutput struct{ Course CourseView }

type UpdateCourseInput struct {
    ID          string
    Title       *string // nil = leave unchanged
    Description *string
    Slug        *string
}
type UpdateCourseOutput struct{ ID string }

type DeleteCourseInput struct{ ID string }
type PublishCourseInput struct{ ID string }
type UnpublishCourseInput struct{ ID string }

// CourseView is a flat read-model for output formatting (table/json/quiet).
type CourseView struct {
    ID, Title, Slug, Description, InstructorID, Status string
    CreatedAt, UpdatedAt time.Time
}

// --- Lesson DTOs ---
type CreateLessonInput struct {
    CourseID string
    Title    string
    Content  string
    Order    *int // nil = append at end (max existing order + 1)
}
type CreateLessonOutput struct{ ID string }

type ListLessonsInput struct{ CourseID string }
type ListLessonsOutput struct{ Lessons []LessonView }

type GetLessonInput struct{ ID string }
type GetLessonOutput struct{ Lesson LessonView }

type UpdateLessonInput struct {
    ID      string
    Title   *string
    Content *string
}
type UpdateLessonOutput struct{ ID string }

type DeleteLessonInput struct{ ID string }

type ReorderLessonsInput struct {
    CourseID string
    Order    []LessonPosition // {LessonID, Position} pairs from --order
}
type LessonPosition struct {
    LessonID string
    Position int
}

type LessonView struct {
    ID, CourseID, Title, Content string
    Order int
    CreatedAt, UpdatedAt time.Time
}
```

### Outbound ports (driven — implemented by adapters)

Declared by the core, depend only on domain types, implemented on the outside.

```go
type CourseRepository interface {
    Save(c Course) error                       // INSERT or UPDATE
    FindByID(id CourseID) (Course, error)       // ErrNotFound if absent
    FindBySlug(s Slug) (Course, error)          // ErrNotFound if absent — slug-uniqueness check
    FindAll(filter CourseFilter) ([]Course, error)
    Delete(id CourseID) error                   // ErrNotFound if absent
}

// CourseFilter is a domain-side query object; empty fields mean "no filter".
type CourseFilter struct {
    Status *CourseStatus // nil = any status
}

type LessonRepository interface {
    Save(l Lesson) error
    SaveAll(ls []Lesson) error                  // batched, single transaction — used by reorder
    FindByID(id LessonID) (Lesson, error)        // ErrNotFound if absent
    FindByCourse(courseID CourseID) ([]Lesson, error) // ordered by Order asc
    Delete(id LessonID) error                    // ErrNotFound if absent
    DeleteByCourse(courseID CourseID) error      // used by DeleteCourse
}

type IDGenerator interface {
    NewCourseID() CourseID
    NewLessonID() LessonID
}

type Clock interface {
    Now() time.Time
}
```

**The dependency rule.** Every port above is declared inside the core and
speaks only domain types (`Course`, `CourseID`, `Slug`, …). The core says "I
need somewhere to save a Course"; it never says "I need Postgres." Adapters on
the outside implement these interfaces. Dependencies point inward.

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**PostgresCourseRepository** implements `CourseRepository`
- Construction dependencies: `*pgxpool.Pool`
- `Save` — side effect: `INSERT INTO courses ... ON CONFLICT (id) DO UPDATE`
- `FindByID` — side effect: `SELECT ... FROM courses WHERE id = $1`
- `FindBySlug` — side effect: `SELECT ... FROM courses WHERE slug = $1`
- `FindAll` — side effect: `SELECT ... FROM courses [WHERE status = $1] ORDER BY created_at DESC`
- `Delete` — side effect: `DELETE FROM courses WHERE id = $1`
- Maps "0 rows" to `ErrNotFound`; maps the `slug` unique-constraint violation to
  `ErrSlugTaken` as a backstop to the usecase's pre-check.

**PostgresLessonRepository** implements `LessonRepository`
- Construction dependencies: `*pgxpool.Pool`
- `Save` — side effect: `INSERT INTO lessons ... ON CONFLICT (id) DO UPDATE`
- `SaveAll` — side effect: `UPDATE lessons` for each row inside one transaction
- `FindByID` — side effect: `SELECT ... FROM lessons WHERE id = $1`
- `FindByCourse` — side effect: `SELECT ... FROM lessons WHERE course_id = $1 ORDER BY "order" ASC`
- `Delete` — side effect: `DELETE FROM lessons WHERE id = $1`
- `DeleteByCourse` — side effect: `DELETE FROM lessons WHERE course_id = $1`

**UUIDGenerator** implements `IDGenerator`
- Construction dependencies: none
- `NewCourseID` / `NewLessonID` — pure; wraps `google/uuid` v4 in the id VO

**SystemClock** implements `Clock`
- Construction dependencies: none
- `Now` — returns `time.Now().UTC()`

> The **inbound CLI adapter** (cobra commands) is also an adapter: it parses
> flags/config into the input DTOs, calls the inbound ports, formats
> `CourseView` / `LessonView` via the table/json/quiet writers, runs `--force`
> confirmation prompts, and maps domain errors to exit codes. It is intentionally
> not skeletoned method-by-method here — it holds no business logic.

### Usecases (implement inbound ports, depend on outbound ports)

Each usecase implements one inbound port method, is constructor-injected with
outbound ports (interfaces, never concrete adapters), and holds the application
logic. Course usecases are grouped under one `CourseServiceImpl`, lesson
usecases under one `LessonServiceImpl` — but each method is described as its own
unit below.

**CreateCourse** — implements `CourseService.CreateCourse`
- Depends on: `CourseRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `Slug` VO from `in.Slug` — return validation error if invalid
  2. build `InstructorID` VO from `in.InstructorID` — return validation error if invalid
  3. `CourseRepository.FindBySlug(slug)` — if it returns a course, return `ErrSlugTaken`
  4. `id := IDGenerator.NewCourseID()`; `now := Clock.Now()`
  5. construct `Course` via `NewCourse(id, in.Title, slug, in.Description, instructorID, now)` — validates non-empty title; status defaults to `draft`
  6. `CourseRepository.Save(course)`
  7. return `CreateCourseOutput{ ID: id.String() }`

**ListCourses** — implements `CourseService.ListCourses`
- Depends on: `CourseRepository`
- Steps:
  1. if `in.Status != ""` build `CourseStatus` VO → `CourseFilter{Status: &s}`; else empty filter
  2. `CourseRepository.FindAll(filter)`
  3. map `[]Course` → `[]CourseView`; return `ListCoursesOutput`

**GetCourse** — implements `CourseService.GetCourse`
- Depends on: `CourseRepository`
- Steps:
  1. build `CourseID` VO from `in.ID`
  2. `CourseRepository.FindByID(id)` — propagate `ErrNotFound`
  3. map to `CourseView`; return `GetCourseOutput`

**UpdateCourse** — implements `CourseService.UpdateCourse`
- Depends on: `CourseRepository`, `Clock`
- Steps:
  1. build `CourseID` VO; reject if all of `Title`/`Description`/`Slug` are nil ("nothing to update")
  2. `CourseRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `now := Clock.Now()`
  4. if `in.Slug` set: build `Slug` VO; `FindBySlug` → if a *different* course owns it, return `ErrSlugTaken`; else `course.ChangeSlug(slug, now)`
  5. if `in.Title` set: `course.Rename(*in.Title, now)`
  6. if `in.Description` set: `course.ChangeDescription(*in.Description, now)`
  7. `CourseRepository.Save(course)`
  8. return `UpdateCourseOutput{ ID: in.ID }`

**DeleteCourse** — implements `CourseService.DeleteCourse`
- Depends on: `CourseRepository`, `LessonRepository`
- Steps:
  1. build `CourseID` VO
  2. `CourseRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `LessonRepository.DeleteByCourse(id)` — remove all lessons first
  4. `CourseRepository.Delete(id)`

**PublishCourse** — implements `CourseService.PublishCourse`
- Depends on: `CourseRepository`, `Clock`
- Steps:
  1. build `CourseID` VO; `FindByID` — propagate `ErrNotFound`
  2. `course.Publish(Clock.Now())` — returns `ErrAlreadyPublished` if applicable
  3. `CourseRepository.Save(course)`

**UnpublishCourse** — implements `CourseService.UnpublishCourse`
- Depends on: `CourseRepository`, `Clock`
- Steps:
  1. build `CourseID` VO; `FindByID` — propagate `ErrNotFound`
  2. `course.Unpublish(Clock.Now())` — returns `ErrNotPublished` if applicable
  3. `CourseRepository.Save(course)`

**CreateLesson** — implements `LessonService.CreateLesson`
- Depends on: `CourseRepository`, `LessonRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `CourseID` VO from `in.CourseID`
  2. `CourseRepository.FindByID(courseID)` — return `ErrNotFound` if the course is absent
  3. determine order: if `in.Order` set, build `LessonOrder` VO from it; else `FindByCourse` and use `maxOrder + 1` (append)
  4. `id := IDGenerator.NewLessonID()`; `now := Clock.Now()`
  5. construct `Lesson` via `NewLesson(id, courseID, in.Title, in.Content, order, now)` — validates non-empty title
  6. `LessonRepository.Save(lesson)`
  7. return `CreateLessonOutput{ ID: id.String() }`

**ListLessons** — implements `LessonService.ListLessons`
- Depends on: `CourseRepository`, `LessonRepository`
- Steps:
  1. build `CourseID` VO; `CourseRepository.FindByID` — return `ErrNotFound` if course absent
  2. `LessonRepository.FindByCourse(courseID)` (already ordered by `Order`)
  3. map `[]Lesson` → `[]LessonView`; return `ListLessonsOutput`

**GetLesson** — implements `LessonService.GetLesson`
- Depends on: `LessonRepository`
- Steps:
  1. build `LessonID` VO from `in.ID`
  2. `LessonRepository.FindByID(id)` — propagate `ErrNotFound`
  3. map to `LessonView`; return `GetLessonOutput`

**UpdateLesson** — implements `LessonService.UpdateLesson`
- Depends on: `LessonRepository`, `Clock`
- Steps:
  1. build `LessonID` VO; reject if both `Title` and `Content` are nil ("nothing to update")
  2. `LessonRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `now := Clock.Now()`
  4. if `in.Title` set: `lesson.Rename(*in.Title, now)`
  5. if `in.Content` set: `lesson.ChangeContent(*in.Content, now)`
  6. `LessonRepository.Save(lesson)`
  7. return `UpdateLessonOutput{ ID: in.ID }`

**DeleteLesson** — implements `LessonService.DeleteLesson`
- Depends on: `LessonRepository`
- Steps:
  1. build `LessonID` VO
  2. `LessonRepository.Delete(id)` — propagate `ErrNotFound`

**ReorderLessons** — implements `LessonService.ReorderLessons`
- Depends on: `CourseRepository`, `LessonRepository`, `Clock`
- Steps:
  1. build `CourseID` VO; `CourseRepository.FindByID` — return `ErrNotFound` if course absent
  2. `LessonRepository.FindByCourse(courseID)` → index existing lessons by id
  3. for each `LessonPosition` in `in.Order`: parse `LessonID`; verify it exists in the index and belongs to this course (else validation error); build `LessonOrder` VO from `Position`
  4. reject duplicate positions in the requested set (validation error)
  5. `now := Clock.Now()`; for each matched lesson call `lesson.MoveTo(order, now)`
  6. `LessonRepository.SaveAll(updatedLessons)` — single transaction

---

## 5. Container & Wiring

The composition root — the single place where concrete types are named and
wired. Everything else in the spec speaks only in interfaces.

```go
type Config struct {
    DBURL        string // resolved by viper: flag > env > config file > default
    InstructorID string // scaffold for the deferred auth context (see §6)
}

type CLI struct {
    Course CourseService
    Lesson LessonService
}

func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
    // 1. construct outbound adapters from config
    pool, err := pgxpool.New(ctx, cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("connect db: %w", err)
    }
    courseRepo := NewPostgresCourseRepository(pool)
    lessonRepo := NewPostgresLessonRepository(pool)
    ids        := NewUUIDGenerator()
    clock      := NewSystemClock()

    // 2. construct usecases — adapters injected through outbound ports only
    courseSvc := NewCourseServiceImpl(courseRepo, lessonRepo, ids, clock)
    lessonSvc := NewLessonServiceImpl(courseRepo, lessonRepo, ids, clock)

    // 3. bind usecases to the CLI (cobra commands consume these via the
    //    inbound ports; flag/config parsing and output formatting live in
    //    the cobra layer, not here)
    return &CLI{Course: courseSvc, Lesson: lessonSvc}, nil
}
```

Every usecase receives `CourseRepository` / `LessonRepository` / `Clock` /
`IDGenerator` as **interfaces**, never `*PostgresCourseRepository`. The payoff:
swapping Postgres for an in-memory fake in tests — or for a different store
later — is a change to the two `New...Repository` lines above and **nowhere
else**. The same applies to `Clock`: tests inject a fixed-time fake so audit
timestamps are deterministic.

The `migrate` commands and the future `auth` commands are wired in `main.go`
*outside* `BuildContainer` — they are not part of this bounded context.

---

## 6. Deferred

Judgment-call items considered during design and deliberately not built now.
Parked here in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| `archived` course status | anticipatory | The PRD's SQL `CHECK` allows it, but no command sets or filters by `archived`. `CourseStatus` ships with `draft`/`published` only; add the third value plus an `archive`/`unarchive` command together when needed. |
| Lesson publish/unpublish + `Lesson.Status` | anticipatory | The PRD's `lessons` table has a `status` column, but the PRD defines **no** command that changes it. Modeling a status with no verb to drive it is premature; add the field with the command. |
| Actor / permission model (`Actor`, role checks, ownership enforcement) | anticipatory | PRD §10 explicitly defers auth; role enforcement is "scaffolded but not enforced." Inbound DTOs already carry `InstructorID` as the seam — when auth lands, a permission check slots into each usecase and an `ErrPermissionDenied` (exit code 3) becomes reachable. No domain `Actor` entity until then. |
| Instructor / User entity + `admin instructor *` commands | out of context | Belongs to a separate **Identity** bounded context. This context references an instructor by `InstructorID` only. |
| Database migrations (`migrate up/down/status`) | out of context | Infrastructure tooling (`golang-migrate` + `embed`); ships in the same binary but is wired in `main.go`, outside this bounded context. |
| `Slug` / `Title` as richer value objects beyond current rules | optimization | `Slug` already has its format invariant; `Title` has only a non-empty rule the constructor enforces. No further VO ceremony earns its keep yet. |
| Soft-delete for courses/lessons | anticipatory | All deletes are hard deletes; no command or requirement asks to recover deleted content. |
```
