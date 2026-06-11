# Technical Spec — Course Context · Phase D: Test

Status: Draft · Architecture: simplified Hexagonal + Clean + DDD (no CQRS)
Extends: `docs/spec.md`, `docs/phase-a-spec.md`, `docs/phase-b-spec.md`, `docs/phase-c-spec.md` · Roadmap: `docs/roadmap.md` Phase D · Date: 2026-05-26

## 0. Overview

- **Purpose:** introduce the **Test** aggregate — a course-level, optionally
  timed, mixed-format assessment composed of `TestItem`s (choice questions or
  coding tasks) plus a downloadable `TestSolution` package (zip + explanation
  video).
- **Language / stack:** Go 1.22, PostgreSQL via `pgx`, `cobra` + `viper` —
  unchanged.
- **In scope:**
  - The `Test` aggregate root (scoped to a Course) and the `TestItem` entity
    inside it.
  - Two item kinds — `choice` (parallel-structured to Phase B's
    `ChoiceQuestion`) and `coding` (parallel-structured to Phase C's
    `Practice`).
  - The `TestSolution` value object: `SolutionZip MediaRef` +
    `ExplanationVideo MediaRef`. Optional on a Test (nil during authoring);
    once set, both refs are required.
  - Optional `TimeLimit` (minutes) — nil means untimed.
  - Reuse of the existing `PassThreshold` VO from Phase B as a shared domain
    type.
  - Eleven new CLI commands (`test` CRUD + `test item` management).
  - Postgres persistence + migration `000005` (creates `tests` /
    `test_items`).
  - REST endpoints + console test-builder at adapter level.
- **Out of scope:**
  - **Code-execution / grading runner** — learner-phase, same as Phase C.
  - **Test attempts, scoring, results** — learner-phase.
  - **Embedding tests in lessons** — per the grill-me, Test is course-level
    only. No `test` value of `ContentBlockKind`; no `lesson block add`
    extension.
  - Test status (draft/published) — deferred (§6) per Phase B/C precedent.
  - Dedicated `test set-solution` / `test clear-solution` commands —
    deferred (§6); solution is set atomically through `test update`.
  - Nested entities for the bundled coding test cases — deferred (§6); they
    live as immutable values inside the item body.

**Architectural notes.**

- `Test` is a **separate aggregate root** like `Quiz` and `Practice`. Unlike
  them, **nothing references a Test** — no content block embeds one — so
  `DeleteTest` has no `RESTRICT` story and `LessonRepository` gains *no* new
  lookup method. This is the simplest cross-aggregate slice.
- `TestItem` is an entity **inside** the Test aggregate (no
  `TestItemRepository`). Same pattern as `ContentBlock` (Phase A),
  `ChoiceQuestion` (Phase B), `TestCase` (Phase C).
- A `TestItem`'s **body** is modeled as a sealed interface,
  `TestItemBody = ChoiceItemBody | CodingItemBody`, mirroring how Phase A
  modeled `ContentBody`. The body is immutable; an item update replaces the
  body in full.
- The `choice` and `coding` body VOs are **parallel-structured** to
  `ChoiceQuestion` (Phase B) and `Practice` (Phase C) — same fields, same
  invariants, independent Go types. Honors the grill-me intent (import +
  grading consistency) without amending Phases B/C. A future refactor could
  extract literally shared content VOs.

---

## 1. CLI Commands

Each command maps to **exactly one inbound port method and one usecase**.

### Test commands (5)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `test create` | Create a test for a course | `--course-id`, `--title`, `--time-limit-minutes` (opt), `--pass-threshold` (opt) | new test id | course not found, missing title, invalid time limit, invalid pass threshold |
| `test list` | List a course's tests | `--course-id`, `-o` | table / JSON / id-only | course not found |
| `test get` | Show a test with items + solution refs | `<test-id>`, `-o` | test detail | test not found |
| `test update` | Edit test metadata and/or set solution package | `<test-id>`, `--title` (opt), `--time-limit-minutes` (opt, `0` clears), `--pass-threshold` (opt), `--solution-zip-provider` + `--solution-zip-locator` (atomic), `--solution-video-provider` + `--solution-video-locator` + `--solution-video-caption` (atomic) | updated test id | test not found, nothing to update, partial solution flags, invalid media ref, invalid pass threshold |
| `test delete` | Delete a test | `<test-id>`, `--force` | confirmation | test not found |

**Solution coupling on `test update`.** If *any* solution flag is set, *all
four* mandatory solution fields (`solution-zip-provider`,
`solution-zip-locator`, `solution-video-provider`, `solution-video-locator`)
must be set together — they are one atomic group. The video caption is
optional. Mismatched-group submissions return a validation error. Matches the
coupling pattern Phase B used for `Options` + `CorrectIndices` on `quiz
question update`.

**Time-limit clearing.** `--time-limit-minutes 0` clears the time limit
(`TimeLimit` becomes nil). A positive value sets it.

### Test item commands (6)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `test item add` | Add an item to a test | `--test-id`, `--kind choice\|coding`, `--position` (opt); **choice:** `--prompt`, `--type single\|multiple`, `--option` (repeatable, ≥ 2), `--correct` (repeatable, 0-based), `--explanation` (opt); **coding:** `--language`, `--prompt`, `--starter-code` (opt), `--solution` (opt), `--testcase <stdin>::<expected>[::<name>]` (repeatable, ≥ 1) | new item id | test not found, invalid kind, kind-required-flag missing, < 2 choice options, no correct indices, out-of-range correct, single-choice with ≠ 1 correct, invalid language, < 1 testcase |
| `test item list` | List a test's items, by position | `--test-id`, `-o` | table / JSON / id-only | test not found |
| `test item get` | Show an item's detail | `<item-id>`, `-o` | item detail | item not found |
| `test item update` | Edit an item — body replaced atomically per kind | `<item-id>` + the same kind-appropriate flag shapes as `add`, all optional but with the same atomic coupling: choice items take `--prompt`, `--type`, `--option`+`--correct` (atomic), `--explanation`; coding items take `--prompt`, `--starter-code`, `--solution`, `--testcase` (atomic, replaces whole list, ≥ 1) | updated item id | item not found, nothing to update, kind mismatch (flags don't match the item's kind), partial atomic group, invariant violations |
| `test item remove` | Delete an item from its test | `<item-id>`, `--force` | confirmation | item not found |
| `test item reorder` | Resequence items within a test | `--test-id`, `--order <item-id:pos,...>` | confirmation | test not found, unknown/foreign item, duplicate position |

**Item kind is fixed at creation.** `test item update` does not change a kind;
to switch kind, `remove` + `add`. Same discipline as `lesson block update`
(Phase A).

**Test cases are bundled inside the coding item.** The `--testcase` flag is
repeatable on `add` / `update`; the format is
`<stdin>::<expected-stdout>[::<optional-name>]`. On `update`, supplying any
`--testcase` replaces the *whole* test case list (atomic, must have ≥ 1).
Bundling keeps the command tree at two levels (`test` / `test item`).

**No `lesson block add` extension.** Tests are course-level only; nothing
embeds them in lessons. There is no `test` `ContentBlockKind`.

### Inbound-adapter exit codes & errors

Unchanged from `docs/spec.md`. Phase D introduces no new "in-use" error
because no other aggregate references a Test.

---

## 2. Domain Model

### Entities

**Test** *(new — aggregate root)* — a course-level, optionally timed
assessment owned by a course; ordered TestItem entities inside, plus an
optional downloadable solution package.
- Identity: `ID TestID`
- Fields:
  - `CourseID CourseID` (id-only reference into the Course aggregate)
  - `Title string`
  - `TimeLimit *TimeLimit` (nil = untimed)
  - `PassThreshold PassThreshold` (reused from Phase B)
  - `Solution *TestSolution` (nil during authoring)
  - `Items []TestItem`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Title` is non-empty after trimming.
  - `CourseID` is always set.
  - **Item ids are unique** within the test.
  - **Item positions are unique** within the test; `Items` is ordered by
    `Position` ascending.
  - `UpdatedAt >= CreatedAt`.
  - No item-count invariant — an empty test is valid during authoring.
- Behaviour (each mutating method takes `now time.Time` and bumps
  `UpdatedAt`):
  - `Rename(title string, now time.Time) error` — re-validates non-empty.
  - `ChangeTimeLimit(tl *TimeLimit, now time.Time)` — accepts nil (untimed).
  - `ChangePassThreshold(t PassThreshold, now time.Time)`.
  - `SetSolution(sol TestSolution, now time.Time)` — assigns / replaces the
    solution package. Both `MediaRef` fields are already validated inside the
    `TestSolution` VO constructor.
  - `AddItem(item TestItem, now time.Time) error` — insert with positional
    shift, exact same shape as `Lesson.AddBlock` and `Quiz.AddQuestion`.
  - `RemoveItem(id TestItemID, now time.Time) error` — removes and compacts
    positions to stay contiguous from 0.
  - `ReorderItems(order []TestItemPlacement, now time.Time) error` —
    placements must be a permutation of exactly the test's current items.
  - `ReplaceItemBody(id TestItemID, body TestItemBody, now time.Time) error`
    — validates `body.Kind()` equals the existing item's `Kind` (kind is
    fixed at creation), then replaces the body in full. Used by every
    item-update flow.

**TestItem** *(new — entity inside the Test aggregate)* — one ordered,
typed assessment item.
- Identity: `ID TestItemID`
- Fields:
  - `Kind TestItemKind`
  - `Body TestItemBody` (sealed interface — see VOs)
  - `Position TestItemPosition`
- Invariants:
  - `Body.Kind()` equals `Kind` (a choice item always holds a
    `ChoiceItemBody`, a coding item always holds a `CodingItemBody`).
- Behaviour (called only by the Test aggregate root):
  - `ReplaceBody(body TestItemBody) error` — validates kind match.
  - `MoveTo(pos TestItemPosition)`.

### Value Objects

- **TestID** — wraps `string`; non-empty UUID. Mirrors `QuizID` / `PracticeID`.
- **TestItemID** — wraps `string`; non-empty UUID.
- **TestItemPosition** — wraps `int`; `>= 0`. Mirrors `BlockPosition` /
  `QuestionPosition` / `TestCasePosition`.
- **TestItemKind** — wraps `string`; one of `choice` | `coding`. Constructors
  `ChoiceKind()` / `CodingKind()`; predicates `IsChoice()` / `IsCoding()`.
- **TimeLimit** — wraps `int` (minutes); invariant: `> 0`. Construct via
  `NewTimeLimit(min)`; the *absence* of a time limit is represented by a nil
  `*TimeLimit` on the Test entity, not a sentinel value inside the VO.
- **TestSolution** — wraps `SolutionZip MediaRef` + `ExplanationVideo MediaRef`
  + `ExplanationCaption string` (optional). Invariant: both `MediaRef` fields
  must be constructible (their own VO constructors enforce locator format).
  Immutable.

**TestItemBody** — Go has no sum types, so the kind-specific payload is a
sealed interface, mirroring the `ContentBody` pattern from Phase A:

```go
type TestItemBody interface {
    Kind() TestItemKind
    isTestItemBody() // unexported — seals to this package
}

// ChoiceItemBody — payload of a `choice` test item.
// Parallel-structured to ChoiceQuestion (Phase B); independent Go type.
type ChoiceItemBody struct {
    Type           ChoiceQuestionType // reused from Phase B
    Prompt         string
    Options        []string
    CorrectIndices []int
    Explanation    string // optional
}

// CodingItemBody — payload of a `coding` test item.
// Parallel-structured to Practice (Phase C); independent Go type.
type CodingItemBody struct {
    Language    Language // reused from Phase C
    Prompt      string
    StarterCode string  // optional ("" when unset)
    Solution    string  // optional ("" when unset)
    TestCases   []CodingTestCase // >= 1 required
}

// CodingTestCase — bundled value inside a CodingItemBody (NOT an entity).
type CodingTestCase struct {
    Stdin          string // optional ("")
    ExpectedStdout string // optional ("")
    Name           string // optional ("")
}
```

`ChoiceItemBody.Kind()` returns `ChoiceKind()`; `CodingItemBody.Kind()`
returns `CodingKind()`. Both are immutable VOs constructed via
`NewChoiceItemBody(...)` / `NewCodingItemBody(...)` which validate:

- **ChoiceItemBody invariants** (identical to Phase B's `ChoiceQuestion`):
  - `Prompt` non-empty after trim.
  - `len(Options) >= 2`.
  - `len(CorrectIndices) >= 1`.
  - every `CorrectIndices` entry in `[0, len(Options))`.
  - no duplicate correct indices.
  - if `Type.IsSingle()`, `len(CorrectIndices) == 1`.
- **CodingItemBody invariants** (identical to Phase C's `Practice` plus a
  test-case-count rule):
  - `Prompt` non-empty after trim.
  - `Language` valid (guaranteed by VO).
  - `len(TestCases) >= 1` — a coding item must have at least one test case to
    be runnable. *(Stricter than Phase C's Practice, where empty test-case
    lists are valid during authoring. Rationale: a Test item that isn't yet
    runnable doesn't belong in an assessment.)*
- **CodingTestCase**: no string invariants; all three fields may be empty.

### Reused types from earlier phases

- `MediaRef` — Phase A, for the `TestSolution` refs.
- `ChoiceQuestionType` — Phase B, for `ChoiceItemBody.Type`.
- `Language` — Phase C, for `CodingItemBody.Language`.
- `PassThreshold` — Phase B, used here as a shared domain type. *(If
  desired, the Phase B spec can be amended in a follow-up to acknowledge
  this shared usage; the type already lives at the package level in
  `internal/course/domain/`.)*

### Domain notes

- **Vertical slices.** Each of the eleven new commands is its own slice
  through the stack — consistent with `docs/spec.md`.
- **`TestItem` has no repository.** It is persisted and hydrated only as
  part of its `Test` through `TestRepository`. `TestRepository.Save` writes
  the test row *and* its items in one transaction. Same pattern as Phase A's
  `ContentBlock` and Phase B/C's nested entities.
- **Coding test cases are bundled values, not entities.** A `CodingItemBody`
  contains a `[]CodingTestCase` value list, replaced atomically when the
  item body is replaced. There is no `test item testcase` subcommand tree;
  test cases live and die with their item body. This diverges from Phase C
  (where Practice test cases *are* entities) to keep the Test command tree
  to two levels.
- **Loading the owning test from an item id.** `test item get/update/remove`
  take a bare `<item-id>` for ergonomics; the usecase loads the aggregate
  via a new `TestRepository.FindByItemID` — symmetric with Phase A/B/C's
  reverse-lookup methods.
- **Item kind is fixed at creation.** `test item update` only ever calls
  `Test.ReplaceItemBody`, which validates the new body's `Kind()` against
  the existing item's `Kind`. To switch kind, remove + add. Same discipline
  as `lesson block update`.
- **`TestSolution` is optional.** During authoring `Test.Solution` is nil;
  `test update` with the solution flag group sets it; once set, both
  `MediaRef` fields are present (the `TestSolution` constructor requires
  both). A future `test clear-solution` could nil it again — deferred (§6).
- **Cross-aggregate delete — Course.** `DeleteCourse` (already extended
  through Phases A–C) gains one more step: `TestRepository.DeleteByCourse(id)`.
  Order: lessons → quizzes → practices → tests → course. Quizzes /
  practices / tests are order-independent among themselves (none reference
  the others); all must come after the lesson cascade (so content blocks go
  away) and before the course row.
- **Cross-aggregate delete — Test.** Unconditional. Nothing references a
  test; there is no `ErrTestInUse`, no embedding lookup. `DeleteTest`
  validates the id, ensures the test exists, and calls
  `TestRepository.Delete`.
- **`Title`, `Prompt`, `StarterCode`, `Solution`, item strings stay plain
  `string`.** Only fields with a non-empty rule are validated by their
  enclosing constructor; the rest carry no invariant. Same discipline as
  every prior phase.

---

## 3. Ports

### Inbound ports (driving — called by an inbound adapter)

A new `TestService` interface — one method per command:

```go
type TestService interface {
    // Test CRUD
    CreateTest(in CreateTestInput) (CreateTestOutput, error)
    ListTests(in ListTestsInput) (ListTestsOutput, error)
    GetTest(in GetTestInput) (GetTestOutput, error)
    UpdateTest(in UpdateTestInput) (UpdateTestOutput, error)
    DeleteTest(in DeleteTestInput) error

    // Item management
    AddTestItem(in AddTestItemInput) (AddTestItemOutput, error)
    ListTestItems(in ListTestItemsInput) (ListTestItemsOutput, error)
    GetTestItem(in GetTestItemInput) (GetTestItemOutput, error)
    UpdateTestItem(in UpdateTestItemInput) (UpdateTestItemOutput, error)
    RemoveTestItem(in RemoveTestItemInput) error
    ReorderTestItems(in ReorderTestItemsInput) error
}
```

DTOs carry only primitives and ids. The item-body crossing — like Phase A's
content-block body — is flattened into kind-tagged primitive fields; the
usecase reassembles the typed `TestItemBody`.

```go
// --- Test DTOs ---

type CreateTestInput struct {
    CourseID         string
    Title            string
    TimeLimitMinutes *int     // nil = untimed
    PassThreshold    *float64 // nil = default (0.7)
}
type CreateTestOutput struct{ ID string }

type ListTestsInput struct{ CourseID string }
type ListTestsOutput struct{ Tests []TestView }

type GetTestInput struct{ ID string }
type GetTestOutput struct{ Test TestDetailView }

type UpdateTestInput struct {
    ID               string
    Title            *string  // nil = leave unchanged
    TimeLimitMinutes *int     // nil = leave unchanged; 0 = clear; >0 = set
    PassThreshold    *float64 // nil = leave unchanged

    // Solution package — atomic group. If ANY of the four mandatory fields
    // is non-nil, ALL FOUR must be non-nil. Caption is optional.
    SolutionZipProvider   *string
    SolutionZipLocator    *string
    SolutionVideoProvider *string
    SolutionVideoLocator  *string
    SolutionVideoCaption  *string // optional even within the group
}
type UpdateTestOutput struct{ ID string }

type DeleteTestInput struct{ ID string }

// --- Test item DTOs ---

type AddTestItemInput struct {
    TestID   string
    Kind     string // "choice" | "coding"
    Position *int   // nil = append

    // choice fields
    Prompt         string
    ChoiceType     string   // "single" | "multiple"
    Options        []string
    CorrectIndices []int
    Explanation    string

    // coding fields
    Language    string
    StarterCode string
    Solution    string
    TestCases   []CodingTestCaseDTO
}
type AddTestItemOutput struct{ ID string }

// CodingTestCaseDTO — DTO image of CodingTestCase value.
type CodingTestCaseDTO struct {
    Stdin          string
    ExpectedStdout string
    Name           string
}

type ListTestItemsInput struct{ TestID string }
type ListTestItemsOutput struct{ Items []TestItemView }

type GetTestItemInput struct{ ID string }
type GetTestItemOutput struct{ Item TestItemView }

type UpdateTestItemInput struct {
    ID string

    // choice fields — nil = leave unchanged; the (Options, CorrectIndices)
    // pair is atomic (one non-nil requires the other non-nil).
    Prompt         *string
    Options        *[]string
    CorrectIndices *[]int
    Explanation    *string

    // coding fields — nil = leave unchanged; TestCases non-nil replaces the
    // whole list (must have >= 1).
    CodingPrompt *string
    StarterCode  *string
    Solution     *string
    TestCases    *[]CodingTestCaseDTO
}
type UpdateTestItemOutput struct{ ID string }

type RemoveTestItemInput struct{ ID string }

type ReorderTestItemsInput struct {
    TestID string
    Order  []TestItemPlacementDTO
}
type TestItemPlacementDTO struct {
    TestItemID string
    Position   int
}

// --- Read models ---

type TestView struct {
    ID, CourseID, Title  string
    TimeLimitMinutes     *int    // nil = untimed
    PassThreshold        float64
    HasSolution          bool
    ItemCount            int
    CreatedAt, UpdatedAt time.Time
}
type TestDetailView struct {
    TestView
    Solution *TestSolutionView // nil when unset
    Items    []TestItemView
}
type TestSolutionView struct {
    ZipProvider, ZipLocator                                          string
    VideoProvider, VideoLocator, VideoCaption                        string
}
type TestItemView struct {
    ID, TestID, Kind string
    Position         int

    // choice fields populated when Kind == "choice"
    ChoicePrompt         string
    ChoiceType           string
    ChoiceOptions        []string
    ChoiceCorrectIndices []int
    ChoiceExplanation    string

    // coding fields populated when Kind == "coding"
    CodingPrompt   string
    Language       string
    StarterCode    string
    CodingSolution string
    TestCases      []CodingTestCaseDTO
}
```

A separate `Prompt` field exists per kind (`Prompt` for choice,
`CodingPrompt` for coding) on `AddTestItemInput` to avoid one name carrying
two contracts; the CLI adapter routes the user-facing `--prompt` flag to the
right field based on `--kind`. The `UpdateTestItemInput` follows the same
split.

### Outbound ports (driven — implemented by adapters)

New port: `TestRepository`. Existing `IDGenerator` gains two methods.
`LessonRepository` is unchanged in Phase D (nothing embeds tests).
`CourseService` and `LessonService` constructors gain `TestRepository` for
the `DeleteCourse` cascade only.

```go
type TestRepository interface {
    Save(t Test) error                              // INSERT or UPDATE + replace items
    FindByID(id TestID) (Test, error)               // ErrNotFound if absent; hydrates items
    FindByCourse(courseID CourseID) ([]Test, error) // ordered, e.g. by CreatedAt DESC
    FindByItemID(id TestItemID) (Test, error)       // owning aggregate from an item id
    Delete(id TestID) error                         // ErrNotFound if absent
    DeleteByCourse(courseID CourseID) error         // used by DeleteCourse
}

type IDGenerator interface {
    NewCourseID()    CourseID
    NewLessonID()    LessonID
    NewBlockID()     BlockID
    NewQuizID()      QuizID
    NewQuestionID()  QuestionID
    NewPracticeID()  PracticeID
    NewTestCaseID()  TestCaseID
    NewTestID()      TestID      // NEW
    NewTestItemID()  TestItemID  // NEW
}

// Clock — unchanged.
// LessonRepository — unchanged.
```

No `MediaStore` (still deferred — `docs/phase-a-spec.md` §6). The
`TestSolution` constructor validates the two `MediaRef`s as pure VO
invariants, the same way Phase A's video block validates `MediaRef`.

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**PostgresTestRepository** *(new)* implements `TestRepository`
- Construction dependencies: `*pgxpool.Pool`
- `Save` — side effect: `INSERT INTO tests ... ON CONFLICT (id) DO UPDATE`,
  then `DELETE FROM test_items WHERE test_id = $1` and re-`INSERT` the
  current item set — all in **one transaction**. The solution fields are
  five nullable columns on `tests` (zip provider/locator, video
  provider/locator, video caption); when `Test.Solution` is nil all five
  are NULL, when set all four mandatory ones are non-NULL.
- `FindByID` — side effect: `SELECT ... FROM tests WHERE id = $1`; then
  `SELECT ... FROM test_items WHERE test_id = $1 ORDER BY position`; hydrate
  each item's body from its kind-specific columns.
- `FindByCourse` — side effect: `SELECT ... FROM tests WHERE course_id = $1
  ORDER BY created_at DESC`; bulk-load items for the result set.
- `FindByItemID` — side effect: `SELECT t.* FROM tests t JOIN test_items ti
  ON ti.test_id = t.id WHERE ti.id = $1`; then hydrate as above. Maps "0
  rows" to `ErrNotFound`.
- `Delete` — side effect: `DELETE FROM tests WHERE id = $1` (cascades to
  `test_items` via FK).
- `DeleteByCourse` — side effect: `DELETE FROM tests WHERE course_id = $1`.

**UUIDGenerator** *(modified)*
- `NewTestID` / `NewTestItemID` — pure; wrap a `google/uuid` v4 in the
  corresponding id VO.

**Migration `000005_add_tests`** — infrastructure, wired in `main.go`
outside the bounded context:

```sql
-- up
CREATE TABLE tests (
    id                          UUID PRIMARY KEY,
    course_id                   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title                       TEXT NOT NULL,
    time_limit_minutes          INTEGER CHECK (time_limit_minutes IS NULL OR time_limit_minutes > 0),
    pass_threshold              DOUBLE PRECISION NOT NULL DEFAULT 0.7
                                CHECK (pass_threshold >= 0 AND pass_threshold <= 1),
    solution_zip_provider       TEXT,
    solution_zip_locator        TEXT,
    solution_video_provider     TEXT,
    solution_video_locator      TEXT,
    solution_video_caption      TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Application enforces the atomic group; this CHECK is a backstop.
    CONSTRAINT tests_solution_group_check CHECK (
        (solution_zip_provider IS NULL
            AND solution_zip_locator IS NULL
            AND solution_video_provider IS NULL
            AND solution_video_locator IS NULL)
        OR
        (solution_zip_provider IS NOT NULL
            AND solution_zip_locator IS NOT NULL
            AND solution_video_provider IS NOT NULL
            AND solution_video_locator IS NOT NULL)
    )
);

CREATE TABLE test_items (
    id            UUID PRIMARY KEY,
    test_id       UUID NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    kind          TEXT NOT NULL CHECK (kind IN ('choice', 'coding')),
    position      INTEGER NOT NULL CHECK (position >= 0),

    -- choice columns (populated only when kind = 'choice')
    choice_type            TEXT,           -- 'single' | 'multiple'
    choice_prompt          TEXT,
    choice_options         JSONB,          -- string array
    choice_correct_indices JSONB,          -- int array
    choice_explanation     TEXT,

    -- coding columns (populated only when kind = 'coding')
    coding_language     TEXT,              -- closed enum (validated app-side)
    coding_prompt       TEXT,
    starter_code        TEXT,
    coding_solution     TEXT,
    coding_test_cases   JSONB,             -- array of {stdin, expected_stdout, name}; >= 1 when kind='coding'

    CONSTRAINT test_items_position_unique UNIQUE (test_id, position)
);
CREATE INDEX test_items_test_position_idx ON test_items (test_id, position);
```

The `down` migration drops `test_items` then `tests`. Safe only if no
production data has used these tables yet; otherwise one-way in practice
(consistent with Phase A–C precedent).

### Inbound adapters (no business logic — not skeletoned per-method)

- **CLI adapter** *(modified)* — extends the cobra tree with a `test`
  subcommand group and a `test item` sub-sub group; parses flags into the
  new DTOs, routes the `--prompt` flag to `Prompt` or `CodingPrompt` based
  on `--kind`, parses repeatable `--testcase <stdin>::<expected>[::<name>]`
  flags into `[]CodingTestCaseDTO`, formats `TestView` / `TestDetailView` /
  `TestItemView` through the existing renderers, runs `--force` confirmation
  on the two destructive commands. No business logic.
- **REST adapter** *(modified — Phase D endpoints)*:

  | Method & path | Inbound port method |
  |---------------|---------------------|
  | `POST /v1/tests` | `CreateTest` |
  | `GET /v1/courses/{courseId}/tests` | `ListTests` |
  | `GET /v1/tests/{testId}` | `GetTest` |
  | `PATCH /v1/tests/{testId}` | `UpdateTest` |
  | `DELETE /v1/tests/{testId}` | `DeleteTest` |
  | `POST /v1/tests/{testId}/items` | `AddTestItem` |
  | `GET /v1/tests/{testId}/items` | `ListTestItems` |
  | `GET /v1/test-items/{itemId}` | `GetTestItem` |
  | `PATCH /v1/test-items/{itemId}` | `UpdateTestItem` |
  | `DELETE /v1/test-items/{itemId}` | `RemoveTestItem` |
  | `POST /v1/tests/{testId}/items/reorder` | `ReorderTestItems` |

  Partial-atomic-group validation errors map to `400 Bad Request` with a
  message naming the missing field. No `409` cases (no `RESTRICT` story).
- **Console** — a test-builder screen (metadata + time-limit picker +
  solution-package form + ordered item list with per-kind inline editor).
  UI detail is not part of this domain spec.

### Usecases (implement inbound ports, depend on outbound ports)

The eleven new usecases are grouped under a new `TestServiceImpl`. One
existing usecase (`DeleteCourse`) is modified.

**CreateTest** — implements `TestService.CreateTest`
- Depends on: `TestRepository`, `CourseRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `CourseID` VO from `in.CourseID`
  2. `CourseRepository.FindByID(courseID)` — propagate `ErrNotFound`
  3. resolve `TimeLimit`: if `in.TimeLimitMinutes != nil` *and* `> 0`,
     `NewTimeLimit(*in.TimeLimitMinutes)` → `*TimeLimit`; else nil
  4. resolve `PassThreshold`: if `in.PassThreshold != nil`,
     `NewPassThreshold(*in.PassThreshold)`; else `DefaultPassThreshold()`
  5. `id := IDGenerator.NewTestID()`; `now := Clock.Now()`
  6. construct `Test` via `NewTest(id, courseID, in.Title, tl, threshold, now)`
     — validates non-empty title; items empty; solution nil
  7. `TestRepository.Save(test)`
  8. return `CreateTestOutput{ ID: id.String() }`

**ListTests** — implements `TestService.ListTests`
- Depends on: `TestRepository`, `CourseRepository`
- Steps:
  1. build `CourseID`; `CourseRepository.FindByID` — propagate `ErrNotFound`
  2. `TestRepository.FindByCourse(courseID)`
  3. map `[]Test` → `[]TestView`; return

**GetTest** — implements `TestService.GetTest`
- Depends on: `TestRepository`
- Steps:
  1. build `TestID` VO
  2. `TestRepository.FindByID(id)` — propagate `ErrNotFound`
  3. map to `TestDetailView` (includes items + solution if set); return

**UpdateTest** — implements `TestService.UpdateTest`
- Depends on: `TestRepository`, `Clock`
- Steps:
  1. build `TestID` VO; reject if no metadata field set **and** no solution
     field set ("nothing to update")
  2. validate the solution-group atomicity: if any of
     `SolutionZipProvider`/`SolutionZipLocator`/`SolutionVideoProvider`/`SolutionVideoLocator`
     is non-nil, all four must be — else return `ErrPartialSolutionGroup`
  3. `TestRepository.FindByID(id)` — propagate `ErrNotFound`
  4. `now := Clock.Now()`
  5. if `in.Title` set: `test.Rename(*in.Title, now)`
  6. if `in.TimeLimitMinutes` set:
     - if value is `0`: `test.ChangeTimeLimit(nil, now)`
     - else: `tl, err := NewTimeLimit(value)`; propagate;
       `test.ChangeTimeLimit(&tl, now)`
  7. if `in.PassThreshold` set: `NewPassThreshold(*v)` → propagate →
     `test.ChangePassThreshold(t, now)`
  8. if the solution group is set:
     - build `MediaProvider` + `MediaRef` for the zip
     - build `MediaProvider` + `MediaRef` for the video (caption uses
       `in.SolutionVideoCaption` if set else `""`)
     - construct `TestSolution{...}`
     - `test.SetSolution(sol, now)`
  9. `TestRepository.Save(test)`
  10. return `UpdateTestOutput{ ID: in.ID }`

**DeleteTest** — implements `TestService.DeleteTest`
- Depends on: `TestRepository`
- Steps:
  1. build `TestID` VO
  2. `TestRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `TestRepository.Delete(id)` — unconditional; nothing references a test

**AddTestItem** — implements `TestService.AddTestItem`
- Depends on: `TestRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `TestID` VO; `TestRepository.FindByID` — propagate `ErrNotFound`
  2. build `TestItemKind` VO from `in.Kind`
  3. build the `TestItemBody` per kind:
     - **choice**: build `ChoiceQuestionType` from `in.ChoiceType`; call
       `NewChoiceItemBody(type, in.Prompt, in.Options, in.CorrectIndices,
       in.Explanation)` — validates the full invariant set
     - **coding**: build `Language` from `in.Language`; map
       `[]CodingTestCaseDTO` → `[]CodingTestCase`; call
       `NewCodingItemBody(lang, in.CodingPrompt, in.StarterCode,
       in.Solution, testCases)` — validates `>= 1` test case
  4. determine `TestItemPosition`: if `in.Position != nil`, build the VO;
     else append (`maxPosition + 1`, or `0` if no items)
  5. `id := IDGenerator.NewTestItemID()`; `now := Clock.Now()`
  6. construct `TestItem{id, kind, body, position}` — validates kind/body
     match
  7. `test.AddItem(item, now)` — insert with shift if needed
  8. `TestRepository.Save(test)`
  9. return `AddTestItemOutput{ ID: id.String() }`

**ListTestItems** — implements `TestService.ListTestItems`
- Depends on: `TestRepository`
- Steps:
  1. build `TestID`; `FindByID` — propagate `ErrNotFound`
  2. map `test.Items` (already ordered) → `[]TestItemView`; return

**GetTestItem** — implements `TestService.GetTestItem`
- Depends on: `TestRepository`
- Steps:
  1. build `TestItemID` VO
  2. `TestRepository.FindByItemID(id)` — propagate `ErrNotFound`
  3. locate the item within the test; map to `TestItemView`; return

**UpdateTestItem** — implements `TestService.UpdateTestItem`
- Depends on: `TestRepository`, `Clock`
- Steps:
  1. build `TestItemID` VO; reject if no field is set ("nothing to update")
  2. `TestRepository.FindByItemID(id)` — propagate `ErrNotFound`
  3. locate the item; read its `Kind`
  4. validate kind/flag match: only flags matching the item's kind may be
     set (e.g. `CodingPrompt`/`Language`/etc. for a `choice` item is an
     `ErrKindMismatch`). For `choice` items: enforce the `Options` +
     `CorrectIndices` atomic pair (one set requires the other).
  5. `now := Clock.Now()`
  6. build the *new* body by starting from the existing body's fields and
     applying the set inputs:
     - **choice**: `prompt := existing.Prompt` (or `*in.Prompt` if set);
       same for options/correct/explanation; then
       `NewChoiceItemBody(...)` — re-validates everything
     - **coding**: `prompt := existing.Prompt` (or `*in.CodingPrompt` if
       set); same for starter/solution; for test cases, if
       `in.TestCases != nil`, use `*in.TestCases` (must have >= 1), else
       reuse existing; then `NewCodingItemBody(...)` — re-validates
  7. `test.ReplaceItemBody(itemID, newBody, now)` — validates kind match
     and bumps `test.UpdatedAt`
  8. `TestRepository.Save(test)`
  9. return `UpdateTestItemOutput{ ID: in.ID }`

**RemoveTestItem** — implements `TestService.RemoveTestItem`
- Depends on: `TestRepository`, `Clock`
- Steps:
  1. build `TestItemID` VO
  2. `TestRepository.FindByItemID(id)` — propagate `ErrNotFound`
  3. `test.RemoveItem(id, Clock.Now())` — removes and compacts positions
  4. `TestRepository.Save(test)`

**ReorderTestItems** — implements `TestService.ReorderTestItems`
- Depends on: `TestRepository`, `Clock`
- Steps:
  1. build `TestID`; `FindByID` — propagate `ErrNotFound`
  2. index the test's items by id
  3. for each placement: parse `TestItemID`; verify it exists in the index
     and belongs to this test (validation error otherwise); build
     `TestItemPosition` VO
  4. reject duplicate positions; require the placement set to be a
     permutation of exactly the test's current items
  5. `now := Clock.Now()`; `test.ReorderItems(placements, now)`
  6. `TestRepository.Save(test)`

### Modified existing usecase

**DeleteCourse** *(from `docs/spec.md` §4, extended in Phases B and C;
further extended here)* — gains `TestRepository`. New step after the
existing practice cascade:
- 3. `LessonRepository.DeleteByCourse(id)` *(existing)*
- 3a. `QuizRepository.DeleteByCourse(id)` *(Phase B)*
- 3b. `PracticeRepository.DeleteByCourse(id)` *(Phase C)*
- 3c. `TestRepository.DeleteByCourse(id)` *(new)* — safe at any point
       after step 3 since nothing references tests
- 4. `CourseRepository.Delete(id)` *(existing)*

`AddLessonBlock` is **not** extended in Phase D — tests are course-level
only, with no content-block representation.

---

## 5. Container & Wiring

The composition root grows by one repository, two id-generator methods,
and one service — exactly the same incremental shape Phases B and C
added:

```go
func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
    // 1. outbound adapters
    pool, err := pgxpool.New(ctx, cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("connect db: %w", err)
    }
    courseRepo   := NewPostgresCourseRepository(pool)
    lessonRepo   := NewPostgresLessonRepository(pool)
    quizRepo     := NewPostgresQuizRepository(pool)
    practiceRepo := NewPostgresPracticeRepository(pool)
    testRepo     := NewPostgresTestRepository(pool)     // NEW
    ids          := NewUUIDGenerator()                  // now also serves NewTestID, NewTestItemID
    clock        := NewSystemClock()

    // 2. usecases — courseSvc gains testRepo (for DeleteCourse cascade).
    //    lessonSvc is NOT extended in Phase D (no test ContentBlock kind).
    //    quizSvc / practiceSvc constructors unchanged from Phases B / C.
    courseSvc   := NewCourseServiceImpl(courseRepo, lessonRepo, quizRepo, practiceRepo, testRepo, ids, clock)
    lessonSvc   := NewLessonServiceImpl(courseRepo, lessonRepo, quizRepo, practiceRepo, ids, clock)
    quizSvc     := NewQuizServiceImpl(courseRepo, lessonRepo, quizRepo, ids, clock)
    practiceSvc := NewPracticeServiceImpl(courseRepo, lessonRepo, practiceRepo, ids, clock)
    testSvc     := NewTestServiceImpl(courseRepo, testRepo, ids, clock) // NEW

    // 3. bind usecases to the inbound ports
    return &CLI{
        Course:   courseSvc,
        Lesson:   lessonSvc,
        Quiz:     quizSvc,
        Practice: practiceSvc,
        Test:     testSvc,
    }, nil
}
```

`main.go` mounts the same inbound adapters (cobra CLI, REST server,
console, playground) onto the returned services — the REST server now
exposes the new test / test-item endpoints simply by being constructed
with `container.Test`.

Because every usecase receives `TestRepository` (and the others) as an
**interface**, swapping the Postgres adapter for an in-memory fake in
tests is the same one-line change to `NewPostgresTestRepository`.

The incremental cost of adding all of Test: **one new outbound port + one
new adapter + two new methods on `IDGenerator` + one new service + one
modified usecase (`DeleteCourse`)**. Same shape as Phase B and Phase C
added. The cross-phase consistency is itself the signal that the
architecture is paying off.

---

## 6. Deferred

Judgment-call items considered during this design pass and deliberately
not built now. Parked in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| Code-execution / grading runner | out of context | Learner-phase, same as Phase C. Phase D stores the test spec + solution package; running learner submissions is its own (sandboxed) project. |
| Test attempts, scoring, result records | out of context | Learner-phase. The Test aggregate models the assessment definition; per-learner attempt state is a separate context. |
| Embedding tests in lessons (a `test` ContentBlockKind) | anticipatory | The grill-me said "Test attaches at course level." If a lesson-embedded mini-test becomes desirable later, add a `test` block kind and a `TestRef` body in a follow-up — additive change. |
| Nested entities for coding test cases | anticipatory | Bundled values inside `CodingItemBody` keep the command tree at two levels. If a future use case needs per-test-case lifecycle (e.g. weighting, partial credit, individual edit history), promote to entities then. |
| Refactor Phase B/C to share content VOs literally | optimization | Parallel-structured VOs honor the grill-me's intent (consistency for import + grading) without amending prior specs. A follow-up could extract `ChoiceQuestionContent` + `CodingExerciseContent` if the parallel maintenance burden grows. |
| Test status (draft/published) and `test publish`/`unpublish` | anticipatory | No command publishes a test; visibility follows its course's publish status. Phase B/C precedent. |
| Dedicated `test set-solution` command | nice-to-have | Solution is set atomically through `test update`'s solution-group flags. If the gesture becomes common enough to deserve its own verb, add it as sugar that calls the same usecase. |
| `test clear-solution` command | nice-to-have | Removing a solution is rare in practice; deleting and recreating the test is the workaround. Add when an authoring workflow actually requires it. |
| `test duplicate` (copy across courses) | nice-to-have | Same reasoning as `quiz duplicate` / `practice duplicate`. |
| Bulk `test set-items` | nice-to-have | Per-item commands cover Phase D authoring; bulk is most useful to the Phase E import agent — add it there. |
| Per-item `CreatedAt` / `UpdatedAt` | nice-to-have | No command shows per-item times; the Test's `UpdatedAt` covers "test changed". Phase A–C precedent. |
| Item-count invariant (e.g. "test must have ≥ 1 item to be valid") | optimization | Empty tests are allowed during authoring; "shippability" is a future publish-time check, not an aggregate invariant. |
| Per-item weight / partial credit | anticipatory | No grading logic yet; all items contribute equally to the pass-threshold computation when scoring lands in the learner phase. |
| Multiple solution attachments (a Test with several zips / videos) | anticipatory | The current `TestSolution` carries exactly one zip + one video — the simplest model. Revisit only if real authoring requires multiple deliverables per test. |
| Time limit at a unit finer than minutes | optimization | Minutes match how authors naturally think about exam length; finer granularity (seconds) can be added by changing the VO's constructor without a domain restructure. |
