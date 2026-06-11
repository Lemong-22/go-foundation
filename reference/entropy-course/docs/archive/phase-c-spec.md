# Technical Spec — Course Context · Phase C: Practice

Status: Draft · Architecture: simplified Hexagonal + Clean + DDD (no CQRS)
Extends: `docs/spec.md`, `docs/phase-a-spec.md`, `docs/phase-b-spec.md` · Roadmap: `docs/roadmap.md` Phase C · Date: 2026-05-26

## 0. Overview

- **Purpose:** introduce the **Practice** aggregate (an authorable, eventually
  auto-gradable coding exercise) and the `practice` value of
  `ContentBlockKind` so a `Lesson` can embed a practice by reference.
- **Language / stack:** Go 1.22, PostgreSQL via `pgx`, `cobra` + `viper` —
  unchanged.
- **In scope:**
  - The `Practice` aggregate root (scoped to a Course) and the `TestCase`
    entity that lives inside it.
  - The `Language` value object — a closed enum admitting `javascript` |
    `typescript` | `golang` | `rust` (the curriculum set).
  - Prompt (markdown), starter code, reference solution — the authorable
    *definition* only. Per the grill-me, "exercise definition only; runner is
    later."
  - Hidden test cases as `Stdin` / `ExpectedStdout` pairs — universal,
    language-agnostic, what the future runner will consume.
  - Eleven new CLI commands (`practice` CRUD + `practice testcase`
    management).
  - The `practice` value added to `ContentBlockKind`, the new `PracticeBody`
    `ContentBody`, and the extension of `lesson block add` to embed a practice.
  - Postgres persistence + migration `000004` (creates `practices` /
    `practice_test_cases`; adds `practice_ref` to `content_blocks`; widens
    the `content_blocks.kind` check constraint).
  - REST endpoints and a console practice-builder at adapter level.
- **Out of scope:**
  - **Code execution and grading runner** — learner-phase, the single
    heaviest engineering item. Phase C stores the exercise *spec*; running
    learner code against tests is its own project.
  - **Gamification mechanics** (XP, streaks, badges, points, leaderboards) —
    learner phase.
  - `Test` (Phase D); bulk `practice set-testcases` (Phase E import).
  - Practice status (draft/published) — deferred (§6); precedent set in
    Phase B.
  - Visible / example test cases shown to the learner up front — deferred
    (§6); the grill-me said "hidden test cases."

**Architectural note.** Like `Quiz`, `Practice` is a **separate aggregate
root** — distinct from how `ContentBlock` lives inside `Lesson`. A practice
block in a lesson holds only a `PracticeID` reference. `TestCase` is an entity
*inside* the Practice aggregate with no repository of its own — the same
pattern Phase B uses for `ChoiceQuestion`-inside-Quiz.

---

## 1. CLI Commands

Each command maps to **exactly one inbound port method and one usecase**.

### Practice commands (5)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `practice create` | Create a practice in a course | `--course-id`, `--title`, `--language`, `--prompt`, `--starter-code` (opt), `--solution` (opt) | new practice id | course not found, missing title/prompt, invalid language |
| `practice list` | List a course's practices | `--course-id`, `-o` | table / JSON / id-only | course not found |
| `practice get` | Show a practice with its test cases | `<practice-id>`, `-o` | practice detail | practice not found |
| `practice update` | Edit practice metadata/content | `<practice-id>`, `--title` (opt), `--prompt` (opt), `--starter-code` (opt), `--solution` (opt) | updated practice id | practice not found, nothing to update |
| `practice delete` | Delete a practice | `<practice-id>`, `--force` | confirmation | practice not found, **practice in use by a content block** |

`--language` is constrained to the closed set Phase C ships (see §2).
`--starter-code` and `--solution` default to empty strings — both can be
authored later.

### Test case commands (6)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `practice testcase add` | Add a test case to a practice | `--practice-id`, `--stdin` (opt, default `""`), `--expected-stdout` (opt, default `""`), `--name` (opt), `--position` (opt) | new test case id | practice not found, negative position |
| `practice testcase list` | List a practice's test cases, by position | `--practice-id`, `-o` | table / JSON / id-only | practice not found |
| `practice testcase get` | Show a test case's detail | `<testcase-id>`, `-o` | test case detail | test case not found |
| `practice testcase update` | Edit a test case | `<testcase-id>`, `--stdin` (opt), `--expected-stdout` (opt), `--name` (opt) | updated test case id | test case not found, nothing to update |
| `practice testcase remove` | Delete a test case | `<testcase-id>`, `--force` | confirmation | test case not found |
| `practice testcase reorder` | Resequence test cases within a practice | `--practice-id`, `--order <testcase-id:pos,...>` | confirmation | practice not found, unknown/foreign test case, duplicate position |

### Changed — `lesson block add`

The Phase A spec deferred `practice` from `ContentBlockKind` "to land in
Phase C with the command that creates it." This phase delivers it:

| Command | Change |
|---------|--------|
| `lesson block add` | Accepts `--kind practice --practice-id <id>` to embed a practice block. New failure modes: practice not found; **practice belongs to a different course than the lesson's course** (cross-course embed rejected). |

### Inbound-adapter exit codes & errors

Unchanged from `docs/spec.md`. The new domain error `ErrPracticeInUse`
carries the list of embedding lesson ids and maps to a non-zero CLI exit
with a stderr message listing them — mirroring `ErrQuizInUse` from Phase B.

---

## 2. Domain Model

### Entities

**Practice** *(new — aggregate root)* — a coding exercise owned by a course;
ordered TestCase entities inside.
- Identity: `ID PracticeID`
- Fields:
  - `CourseID CourseID` (id-only reference into the Course aggregate)
  - `Title string`
  - `Language Language`
  - `Prompt string` (markdown)
  - `StarterCode string` (source; may be empty)
  - `Solution string` (source; may be empty)
  - `TestCases []TestCase`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Title` is non-empty after trimming.
  - `Prompt` is non-empty after trimming.
  - `CourseID` is always set.
  - `Language` is one of the curriculum values (guaranteed by the VO).
  - **Test case ids are unique** within the practice.
  - **Test case positions are unique** within the practice; `TestCases` is
    ordered by `Position` ascending.
  - `UpdatedAt >= CreatedAt`.
- Behaviour (each mutating method takes `now time.Time` and bumps
  `UpdatedAt`):
  - `Rename(title string, now time.Time) error` — re-validates non-empty.
  - `ChangePrompt(prompt string, now time.Time) error` — re-validates
    non-empty.
  - `ChangeStarterCode(code string, now time.Time)` — no invariant.
  - `ChangeSolution(source string, now time.Time)` — no invariant.
  - `AddTestCase(tc TestCase, now time.Time) error` — inserts; if `Position`
    collides, test cases at `>= Position` shift up by one (insert
    semantics), matching `Lesson.AddBlock` / `Quiz.AddQuestion`. A test
    case constructed for "append" already carries `maxPosition + 1`.
  - `RemoveTestCase(id TestCaseID, now time.Time) error` — removes and
    compacts remaining positions so they stay contiguous from 0.
  - `ReorderTestCases(order []TestCasePlacement, now time.Time) error` —
    placements must be a permutation of exactly the practice's current test
    cases.
  - `ChangeTestCaseStdin(id TestCaseID, stdin string, now time.Time) error`.
  - `ChangeTestCaseExpectedStdout(id TestCaseID, expected string, now time.Time) error`.
  - `ChangeTestCaseName(id TestCaseID, name string, now time.Time) error`.

  All `ChangeTestCase*` methods locate the test case by id (`ErrNotFound`
  if absent) and route the mutation through the aggregate root so
  invariants are re-checked at the boundary.

**TestCase** *(new — entity inside the Practice aggregate)* — one ordered
input/output pair the runner will eventually check the learner's code
against. An entity (`TestCaseID`, lifecycle) but **not** an aggregate root.
- Identity: `ID TestCaseID`
- Fields:
  - `Stdin string` (may be empty — program runs with no input)
  - `ExpectedStdout string` (may be empty — program is expected to print
    nothing)
  - `Name string` (optional human label; may be empty)
  - `Position TestCasePosition`
- Invariants:
  - `Position >= 0` (guaranteed by the VO).
  - All three string fields can be empty — a TestCase with both `Stdin` and
    `ExpectedStdout` empty is a valid "does it run without error?" check.
- Behaviour (called only by the Practice aggregate root):
  - `ChangeStdin(stdin string)`.
  - `ChangeExpectedStdout(expected string)`.
  - `ChangeName(name string)`.
  - `MoveTo(pos TestCasePosition)`.

### Value Objects

- **PracticeID** — wraps `string`; invariant: non-empty, parseable as a UUID.
  Mirrors `QuizID`.
- **TestCaseID** — wraps `string`; invariant: non-empty, parseable as a UUID.
- **TestCasePosition** — wraps `int`; invariant: `>= 0`. Mirrors
  `QuestionPosition`.
- **Language** — wraps `string`; invariant: one of `javascript` |
  `typescript` | `golang` | `rust` (the curriculum set). Exposes
  `JavaScript()` / `TypeScript()` / `Golang()` / `Rust()` constructors and
  per-language predicates. Adding a fifth language is one constant + one
  CHECK-constraint change.

### Extensions to existing Phase A / B types

- **`ContentBlockKind`** now admits a fourth value: `practice`. Exposes
  `PracticeKind()` constructor and `IsPractice()` predicate.
- **`ContentBody`** gains a fourth implementation:

  ```go
  // PracticeBody — payload of a `practice` block.
  type PracticeBody struct {
      PracticeRef PracticeID
  }
  ```

  `PracticeBody.Kind()` returns `PracticeKind()`. Like the existing bodies
  (`TextBody`, `VideoBody`, `QuizBody`) it is an immutable value object.

### Domain notes

- **Vertical slices.** Each of the eleven new commands is its own slice
  through the stack — consistent with `docs/spec.md`.
- **`TestCase` has no repository.** It is persisted and hydrated only as
  part of its `Practice` through `PracticeRepository`.
  `PracticeRepository.Save` writes the practice row *and* its test cases in
  one transaction. Same pattern as Phase A's `ContentBlock` and Phase B's
  `ChoiceQuestion`.
- **Loading the owning practice from a test case id.** `practice testcase
  get/update/remove` take a bare `<testcase-id>` for ergonomics. The
  usecase loads the aggregate via a new
  `PracticeRepository.FindByTestCaseID` — symmetric with
  `LessonRepository.FindByBlockID` (Phase A) and
  `QuizRepository.FindByQuestionID` (Phase B).
- **Test cases are bare stdin/expected-stdout pairs.** Per the grill-me, no
  function-signature schema, no JSON-typed argument lists. Universal,
  language-agnostic, what the future runner will consume by piping stdin
  and comparing stdout. The runner phase can normalize comparison
  (trim trailing whitespace, line endings, etc.) — that's its concern.
- **Empty practices are allowed during authoring.** A practice can exist
  with zero test cases and an empty solution immediately after `practice
  create`; no invariant requires shippability. A future `practice validate`
  (deferred — §6) would enforce "at least one test case + non-empty
  solution" as a *publish-time* check rather than an aggregate invariant.
- **Cross-aggregate embed invariant.** When `lesson block add --kind
  practice --practice-id X`, the `AddLessonBlock` usecase verifies the
  practice's `CourseID` equals the lesson's `CourseID`. Cross-course
  embedding is a validation error — same rule Phase B applied to quiz
  embeds.
- **Cross-aggregate delete — Course.** `DeleteCourse` (now extended through
  Phase A and Phase B) gains one more step: orchestrate
  `PracticeRepository.DeleteByCourse(id)` alongside the existing lesson +
  quiz cascade. Lessons are deleted *first* so their content blocks
  (including practice embeds) cascade away, then quizzes, then practices,
  then the course — preserving the `content_blocks.practice_ref` FK as a
  `RESTRICT` backstop.
- **Cross-aggregate delete — Practice (RESTRICT).** `DeletePractice` is
  *not* a silent cascade. It first calls
  `LessonRepository.FindLessonsEmbeddingPractice(id)`; if any embedding
  lessons exist, it returns `ErrPracticeInUse` carrying their ids. The
  author must `lesson block remove` the embeds explicitly before retrying.
  The FK on `content_blocks.practice_ref` is `ON DELETE RESTRICT` as a
  safety backstop — mirroring the Quiz pattern.
- **`Title`, `Prompt`, `StarterCode`, `Solution`, `TestCase` strings stay
  plain `string`.** Only `Title` and `Prompt` carry a non-empty rule, and
  the entity constructor enforces it. Wrapping source-code fields in a VO
  would be ceremony — they have no invariant beyond "is a string."

---

## 3. Ports

### Inbound ports (driving — called by an inbound adapter)

A new `PracticeService` interface — one method per command:

```go
// PracticeService — inbound port for all practice + test case commands.
type PracticeService interface {
    // Practice CRUD
    CreatePractice(in CreatePracticeInput) (CreatePracticeOutput, error)
    ListPractices(in ListPracticesInput) (ListPracticesOutput, error)
    GetPractice(in GetPracticeInput) (GetPracticeOutput, error)
    UpdatePractice(in UpdatePracticeInput) (UpdatePracticeOutput, error)
    DeletePractice(in DeletePracticeInput) error

    // Test case management
    AddTestCase(in AddTestCaseInput) (AddTestCaseOutput, error)
    ListTestCases(in ListTestCasesInput) (ListTestCasesOutput, error)
    GetTestCase(in GetTestCaseInput) (GetTestCaseOutput, error)
    UpdateTestCase(in UpdateTestCaseInput) (UpdateTestCaseOutput, error)
    RemoveTestCase(in RemoveTestCaseInput) error
    ReorderTestCases(in ReorderTestCasesInput) error
}
```

`LessonService.AddLessonBlockInput` from Phase A / B gains one field,
`PracticeRef string`, used when `Kind == "practice"`.

DTOs carry only primitives and ids — the boundary between the adapter world
and the VO-typed domain.

```go
// --- Practice DTOs ---

type CreatePracticeInput struct {
    CourseID    string
    Title       string
    Language    string
    Prompt      string
    StarterCode string // optional; "" by default
    Solution    string // optional; "" by default
}
type CreatePracticeOutput struct{ ID string }

type ListPracticesInput struct{ CourseID string }
type ListPracticesOutput struct{ Practices []PracticeView }

type GetPracticeInput struct{ ID string }
type GetPracticeOutput struct{ Practice PracticeDetailView }

type UpdatePracticeInput struct {
    ID          string
    Title       *string // nil = leave unchanged
    Prompt      *string
    StarterCode *string
    Solution    *string
}
type UpdatePracticeOutput struct{ ID string }

type DeletePracticeInput struct{ ID string }

// --- Test case DTOs ---

type AddTestCaseInput struct {
    PracticeID     string
    Stdin          string // "" by default
    ExpectedStdout string // "" by default
    Name           string // optional label
    Position       *int   // nil = append at end
}
type AddTestCaseOutput struct{ ID string }

type ListTestCasesInput struct{ PracticeID string }
type ListTestCasesOutput struct{ TestCases []TestCaseView }

type GetTestCaseInput struct{ ID string }
type GetTestCaseOutput struct{ TestCase TestCaseView }

type UpdateTestCaseInput struct {
    ID             string
    Stdin          *string // nil = leave unchanged
    ExpectedStdout *string
    Name           *string
}
type UpdateTestCaseOutput struct{ ID string }

type RemoveTestCaseInput struct{ ID string }

type ReorderTestCasesInput struct {
    PracticeID string
    Order      []TestCasePlacementDTO
}
type TestCasePlacementDTO struct {
    TestCaseID string
    Position   int
}

// --- Read models ---

type PracticeView struct {
    ID, CourseID, Title  string
    Language             string
    TestCaseCount        int
    HasSolution          bool
    CreatedAt, UpdatedAt time.Time
}
type PracticeDetailView struct {
    PracticeView
    Prompt      string
    StarterCode string
    Solution    string
    TestCases   []TestCaseView
}
type TestCaseView struct {
    ID, PracticeID                  string
    Stdin, ExpectedStdout, Name     string
    Position                        int
}
```

### Outbound ports (driven — implemented by adapters)

New port: `PracticeRepository`. Existing `LessonRepository` and
`IDGenerator` gain methods. `CourseService` and `LessonService`
implementations gain `PracticeRepository` as a constructor dependency
(for the cross-aggregate concerns).

```go
type PracticeRepository interface {
    Save(p Practice) error                                // INSERT or UPDATE + replace test cases
    FindByID(id PracticeID) (Practice, error)             // ErrNotFound if absent; hydrates test cases
    FindByCourse(courseID CourseID) ([]Practice, error)   // ordered, e.g. by CreatedAt DESC
    FindByTestCaseID(id TestCaseID) (Practice, error)     // owning aggregate from a test case id
    Delete(id PracticeID) error                           // ErrNotFound if absent
    DeleteByCourse(courseID CourseID) error               // used by DeleteCourse
}

type LessonRepository interface {
    // ... existing methods from docs/spec.md, docs/phase-a-spec.md,
    //     docs/phase-b-spec.md ...

    // NEW Phase C — used by DeletePractice to enforce RESTRICT
    FindLessonsEmbeddingPractice(practiceID PracticeID) ([]LessonID, error)
}

type IDGenerator interface {
    NewCourseID()     CourseID
    NewLessonID()     LessonID
    NewBlockID()      BlockID
    NewQuizID()       QuizID
    NewQuestionID()   QuestionID
    NewPracticeID()   PracticeID   // NEW
    NewTestCaseID()   TestCaseID   // NEW
}

// Clock — unchanged.
```

No `MediaStore` (still deferred — `docs/phase-a-spec.md` §6).

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**PostgresPracticeRepository** *(new)* implements `PracticeRepository`
- Construction dependencies: `*pgxpool.Pool`
- `Save` — side effect: `INSERT INTO practices ... ON CONFLICT (id) DO
  UPDATE`, then `DELETE FROM practice_test_cases WHERE practice_id = $1`
  and re-`INSERT` the current test case set — all in **one transaction**.
- `FindByID` — side effect: `SELECT ... FROM practices WHERE id = $1`; then
  `SELECT ... FROM practice_test_cases WHERE practice_id = $1 ORDER BY
  position`; hydrate.
- `FindByCourse` — side effect: `SELECT ... FROM practices WHERE course_id
  = $1 ORDER BY created_at DESC`; bulk-load test cases for the result set.
- `FindByTestCaseID` — side effect: `SELECT p.* FROM practices p JOIN
  practice_test_cases tc ON tc.practice_id = p.id WHERE tc.id = $1`; then
  hydrate. Maps "0 rows" to `ErrNotFound`.
- `Delete` — side effect: `DELETE FROM practices WHERE id = $1` (cascades
  to `practice_test_cases` via FK).
- `DeleteByCourse` — side effect: `DELETE FROM practices WHERE course_id
  = $1`.

**PostgresLessonRepository** *(modified)*
- `Save` / `FindByID` / `FindByCourse` — now also handle the
  `content_blocks.practice_ref` column (alongside text/video columns from
  Phase A and `quiz_ref` from Phase B). For a `kind = 'practice'` block,
  the only persisted field is `practice_ref`.
- `FindLessonsEmbeddingPractice` *(new)* — side effect: `SELECT DISTINCT
  l.id FROM lessons l JOIN content_blocks b ON b.lesson_id = l.id WHERE
  b.kind = 'practice' AND b.practice_ref = $1`.

**UUIDGenerator** *(modified)*
- `NewPracticeID` / `NewTestCaseID` — pure; wrap a `google/uuid` v4 in the
  corresponding id VO.

**Migration `000004_add_practices`** — infrastructure, wired in `main.go`
outside the bounded context:

```sql
-- up
CREATE TABLE practices (
    id            UUID PRIMARY KEY,
    course_id     UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title         TEXT NOT NULL,
    language      TEXT NOT NULL CHECK (language IN
                  ('javascript', 'typescript', 'golang', 'rust')),
    prompt        TEXT NOT NULL,
    starter_code  TEXT NOT NULL DEFAULT '',
    solution      TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE practice_test_cases (
    id              UUID PRIMARY KEY,
    practice_id     UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    stdin           TEXT NOT NULL DEFAULT '',
    expected_stdout TEXT NOT NULL DEFAULT '',
    name            TEXT NOT NULL DEFAULT '',
    position        INTEGER NOT NULL CHECK (position >= 0),
    CONSTRAINT practice_test_cases_position_unique UNIQUE (practice_id, position)
);
CREATE INDEX practice_test_cases_practice_position_idx
    ON practice_test_cases (practice_id, position);

-- Extend content_blocks for the practice kind
ALTER TABLE content_blocks DROP CONSTRAINT content_blocks_kind_check;
ALTER TABLE content_blocks ADD CONSTRAINT content_blocks_kind_check
    CHECK (kind IN ('text', 'video', 'quiz', 'practice'));
ALTER TABLE content_blocks ADD COLUMN practice_ref UUID
    REFERENCES practices(id) ON DELETE RESTRICT;
CREATE INDEX content_blocks_practice_ref_idx ON content_blocks (practice_ref);
```

The `down` migration reverses these steps in opposite order, and is safe
*only* if no production data has used `kind = 'practice'` yet — once
authors are embedding practices, this migration is one-way in practice
(consistent with the Phase A/B notes).

### Inbound adapters (no business logic — not skeletoned per-method)

- **CLI adapter** *(modified)* — extends the cobra tree with a `practice`
  subcommand group and a `practice testcase` sub-sub group; parses
  flags/config into the new DTOs, formats `PracticeView` /
  `PracticeDetailView` / `TestCaseView` through the existing
  table/json/quiet writers, runs `--force` confirmation on the two
  destructive commands, maps `ErrPracticeInUse` to a stderr message
  listing the embedding lesson ids before exiting non-zero. No business
  logic.
- **REST adapter** *(modified — Phase C endpoints)*:

  | Method & path | Inbound port method |
  |---------------|---------------------|
  | `POST /v1/practices` | `CreatePractice` |
  | `GET /v1/courses/{courseId}/practices` | `ListPractices` |
  | `GET /v1/practices/{practiceId}` | `GetPractice` |
  | `PATCH /v1/practices/{practiceId}` | `UpdatePractice` |
  | `DELETE /v1/practices/{practiceId}` | `DeletePractice` |
  | `POST /v1/practices/{practiceId}/testcases` | `AddTestCase` |
  | `GET /v1/practices/{practiceId}/testcases` | `ListTestCases` |
  | `GET /v1/testcases/{testCaseId}` | `GetTestCase` |
  | `PATCH /v1/testcases/{testCaseId}` | `UpdateTestCase` |
  | `DELETE /v1/testcases/{testCaseId}` | `RemoveTestCase` |
  | `POST /v1/practices/{practiceId}/testcases/reorder` | `ReorderTestCases` |

  `ErrPracticeInUse` maps to HTTP `409 Conflict` with a JSON body listing
  the embedding lesson ids — same pattern as `ErrQuizInUse`.
- **Console** — a practice-builder screen (metadata + prompt editor +
  starter-code / solution panes + ordered test case list with inline
  edit) and a practice-picker in the lesson block editor that calls
  `lesson block add --kind practice`. UI detail is not part of this
  domain spec.

### Usecases (implement inbound ports, depend on outbound ports)

The eleven new practice / test-case usecases are grouped under a new
`PracticeServiceImpl`. Two existing usecases are modified.

**CreatePractice** — implements `PracticeService.CreatePractice`
- Depends on: `PracticeRepository`, `CourseRepository`, `IDGenerator`,
  `Clock`
- Steps:
  1. build `CourseID` VO from `in.CourseID`
  2. `CourseRepository.FindByID(courseID)` — propagate `ErrNotFound`
  3. build `Language` VO from `in.Language` — validation error if not in
     the curriculum set
  4. `id := IDGenerator.NewPracticeID()`; `now := Clock.Now()`
  5. construct `Practice` via `NewPractice(id, courseID, in.Title, lang,
     in.Prompt, in.StarterCode, in.Solution, now)` — validates non-empty
     title and prompt; test cases start empty
  6. `PracticeRepository.Save(practice)`
  7. return `CreatePracticeOutput{ ID: id.String() }`

**ListPractices** — implements `PracticeService.ListPractices`
- Depends on: `PracticeRepository`, `CourseRepository`
- Steps:
  1. build `CourseID` VO; `CourseRepository.FindByID` — propagate
     `ErrNotFound`
  2. `PracticeRepository.FindByCourse(courseID)`
  3. map `[]Practice` → `[]PracticeView`; return

**GetPractice** — implements `PracticeService.GetPractice`
- Depends on: `PracticeRepository`
- Steps:
  1. build `PracticeID` VO
  2. `PracticeRepository.FindByID(id)` — propagate `ErrNotFound`
  3. map to `PracticeDetailView` (includes test cases); return

**UpdatePractice** — implements `PracticeService.UpdatePractice`
- Depends on: `PracticeRepository`, `Clock`
- Steps:
  1. build `PracticeID` VO; reject if all of `Title`, `Prompt`,
     `StarterCode`, `Solution` are nil ("nothing to update")
  2. `PracticeRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `now := Clock.Now()`
  4. if `in.Title` set: `practice.Rename(*in.Title, now)`
  5. if `in.Prompt` set: `practice.ChangePrompt(*in.Prompt, now)`
  6. if `in.StarterCode` set:
     `practice.ChangeStarterCode(*in.StarterCode, now)`
  7. if `in.Solution` set: `practice.ChangeSolution(*in.Solution, now)`
  8. `PracticeRepository.Save(practice)`
  9. return `UpdatePracticeOutput{ ID: in.ID }`

**DeletePractice** — implements `PracticeService.DeletePractice`
- Depends on: `PracticeRepository`, `LessonRepository`
- Steps:
  1. build `PracticeID` VO
  2. `PracticeRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `LessonRepository.FindLessonsEmbeddingPractice(id)` — if non-empty,
     return `ErrPracticeInUse` carrying the lesson ids
  4. `PracticeRepository.Delete(id)`

**AddTestCase** — implements `PracticeService.AddTestCase`
- Depends on: `PracticeRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `PracticeID` VO; `PracticeRepository.FindByID` — propagate
     `ErrNotFound`
  2. determine `TestCasePosition`: if `in.Position != nil`, build the VO;
     else append (`maxPosition + 1`, or `0` if no test cases)
  3. `id := IDGenerator.NewTestCaseID()`; `now := Clock.Now()`
  4. construct `TestCase` via `NewTestCase(id, in.Stdin,
     in.ExpectedStdout, in.Name, position)` — no string invariants beyond
     `position >= 0`
  5. `practice.AddTestCase(tc, now)` — inserts with shift if needed
  6. `PracticeRepository.Save(practice)`
  7. return `AddTestCaseOutput{ ID: id.String() }`

**ListTestCases** — implements `PracticeService.ListTestCases`
- Depends on: `PracticeRepository`
- Steps:
  1. build `PracticeID`; `FindByID` — propagate `ErrNotFound`
  2. map `practice.TestCases` (already ordered) → `[]TestCaseView`; return

**GetTestCase** — implements `PracticeService.GetTestCase`
- Depends on: `PracticeRepository`
- Steps:
  1. build `TestCaseID` VO
  2. `PracticeRepository.FindByTestCaseID(id)` — propagate `ErrNotFound`
  3. locate the test case within the practice; map to `TestCaseView`;
     return

**UpdateTestCase** — implements `PracticeService.UpdateTestCase`
- Depends on: `PracticeRepository`, `Clock`
- Steps:
  1. build `TestCaseID` VO; reject if all of `Stdin`, `ExpectedStdout`,
     `Name` are nil ("nothing to update")
  2. `PracticeRepository.FindByTestCaseID(id)` — propagate `ErrNotFound`
  3. `now := Clock.Now()`
  4. if `in.Stdin` set: `practice.ChangeTestCaseStdin(id, *Stdin, now)`
  5. if `in.ExpectedStdout` set:
     `practice.ChangeTestCaseExpectedStdout(id, *ExpectedStdout, now)`
  6. if `in.Name` set: `practice.ChangeTestCaseName(id, *Name, now)`
  7. `PracticeRepository.Save(practice)`
  8. return `UpdateTestCaseOutput{ ID: in.ID }`

**RemoveTestCase** — implements `PracticeService.RemoveTestCase`
- Depends on: `PracticeRepository`, `Clock`
- Steps:
  1. build `TestCaseID` VO
  2. `PracticeRepository.FindByTestCaseID(id)` — propagate `ErrNotFound`
  3. `practice.RemoveTestCase(id, Clock.Now())` — removes and compacts
     positions
  4. `PracticeRepository.Save(practice)`

**ReorderTestCases** — implements `PracticeService.ReorderTestCases`
- Depends on: `PracticeRepository`, `Clock`
- Steps:
  1. build `PracticeID`; `FindByID` — propagate `ErrNotFound`
  2. index the practice's test cases by id
  3. for each placement: parse `TestCaseID`; verify it exists in the index
     and belongs to this practice (validation error otherwise); build
     `TestCasePosition` VO
  4. reject duplicate positions; require the placement set to be a
     permutation of exactly the practice's current test cases
  5. `now := Clock.Now()`; `practice.ReorderTestCases(placements, now)`
  6. `PracticeRepository.Save(practice)`

### Modified existing usecases

**AddLessonBlock** *(extended in Phase B; further extended here)* — gains
`PracticeRepository` as a dependency and a new branch for `kind ==
"practice"`:
- After the existing kind branches (`text`, `video`, `quiz`):
  - if `practice` *(new)*:
    1. build `PracticeID` from `in.PracticeRef`
    2. `PracticeRepository.FindByID(practiceID)` — propagate `ErrNotFound`
    3. verify `practice.CourseID == lesson.CourseID` — else return
       validation error `ErrCrossCoursePracticeEmbed`
    4. body = `PracticeBody{ PracticeRef: practiceID }`
- Remaining steps (`AddBlock`, `Save`, return) — unchanged.

**DeleteCourse** *(from `docs/spec.md` §4, extended in Phase B; further
extended here)* — gains `PracticeRepository`. New step after the existing
quiz cascade:
- 3. `LessonRepository.DeleteByCourse(id)` *(existing)*
- 3a. `QuizRepository.DeleteByCourse(id)` *(Phase B)*
- 3b. `PracticeRepository.DeleteByCourse(id)` *(new)* — safe to run after
      step 3 because all `content_blocks.practice_ref` rows in this
      course have already cascaded away with their lessons.
- 4. `CourseRepository.Delete(id)` *(existing)*

The order of 3a and 3b is interchangeable (quizzes and practices don't
reference each other); both must come after step 3 and before step 4.

---

## 5. Container & Wiring

The composition root grows by one repository, two id-generator methods,
and one service — exactly the same shape Phase B added:

```go
func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
    // 1. outbound adapters
    pool, err := pgxpool.New(ctx, cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("connect db: %w", err)
    }
    courseRepo   := NewPostgresCourseRepository(pool)
    lessonRepo   := NewPostgresLessonRepository(pool)   // now block + quiz_ref + practice_ref aware
    quizRepo     := NewPostgresQuizRepository(pool)
    practiceRepo := NewPostgresPracticeRepository(pool) // NEW
    ids          := NewUUIDGenerator()                  // now also serves NewPracticeID, NewTestCaseID
    clock        := NewSystemClock()

    // 2. usecases — existing service constructors gain practiceRepo for the
    //    cross-aggregate concerns: DeleteCourse cascade, AddLessonBlock kind=practice
    courseSvc   := NewCourseServiceImpl(courseRepo, lessonRepo, quizRepo, practiceRepo, ids, clock)
    lessonSvc   := NewLessonServiceImpl(courseRepo, lessonRepo, quizRepo, practiceRepo, ids, clock)
    quizSvc     := NewQuizServiceImpl(courseRepo, lessonRepo, quizRepo, ids, clock)
    practiceSvc := NewPracticeServiceImpl(courseRepo, lessonRepo, practiceRepo, ids, clock) // NEW

    // 3. bind usecases to the inbound ports
    return &CLI{
        Course:   courseSvc,
        Lesson:   lessonSvc,
        Quiz:     quizSvc,
        Practice: practiceSvc,
    }, nil
}
```

`main.go` mounts the same inbound adapters (cobra CLI, REST server,
console, playground) onto the returned services — the REST server now
exposes the new practice / test-case endpoints simply by being constructed
with `container.Practice` in addition to the existing services.

Because every usecase receives `PracticeRepository` (and the others) as an
**interface**, swapping the Postgres adapter for an in-memory fake in
tests is the same one-line change to `NewPostgresPracticeRepository`. The
new `FindLessonsEmbeddingPractice` method on `LessonRepository` follows
the same discipline — fakes implement it the same way as `FindByID`.

---

## 6. Deferred

Judgment-call items considered during this design pass and deliberately
not built now. Parked in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| Code-execution / grading runner | out of context | Learner-phase, and the single heaviest engineering item in the roadmap. Phase C stores the spec; running learner code against test cases is its own (sandboxed) project. |
| Gamification mechanics (XP, streaks, badges, points, leaderboards) | out of context | Learner-phase per the grill-me; Practice models the exercise definition only. |
| Function-args / typed-argument test cases | anticipatory | The grill-me chose stdin/expected-stdout for universality. If a future language-specific runner needs typed function-call cases, add a `TestCaseMode` discriminator alongside (`stdio` | `function`) — the migration would be additive. |
| Visible / example test cases (an `IsVisible` flag) | anticipatory | No Phase C command surfaces it; example I/O can go in the prompt markdown or starter-code comments. Add the field with the learner-side command that displays it. |
| Practice status (draft/published) and `practice publish`/`unpublish` | anticipatory | No command publishes a practice; visibility follows the embedding lesson/course. Phase A/B's "model a value with a command" discipline. |
| `practice validate` (spec-completeness: ≥ 1 test case + non-empty solution) | nice-to-have | Useful as a publish-time / CI / import sanity check, but not as an aggregate invariant (empty practices are valid during authoring). Revisit when Phase E import wants it. |
| `practice duplicate` (copy across courses) | nice-to-have | Same reasoning as `quiz duplicate` — no current authoring activity drives it. |
| Bulk `practice set-testcases` | nice-to-have | Per-test-case commands cover Phase C authoring; bulk is most useful to the Phase E import agent — add it there. |
| Per-test-case `CreatedAt` / `UpdatedAt` | nice-to-have | No command shows per-test-case times; the Practice's `UpdatedAt` already covers "practice changed." Phase A/B precedent. |
| Multiple languages per Practice (one prompt, many starter-code/solution variants) | anticipatory | The model is one Practice = one Language; if you want the same prompt in JavaScript and Go, author two practices. Revisit only if cross-language authoring becomes a real authoring pain. |
| Richer Language metadata (version, dialect, tooling hints) | optimization | The closed enum is enough; the runner phase can carry version constraints separately. |
| `Solution` / `StarterCode` as richer entities (with explanation, complexity hints) | optimization | Plain strings match the rest of the codebase's discipline (`Description`, `Markdown`); add fields when a command needs them. |
