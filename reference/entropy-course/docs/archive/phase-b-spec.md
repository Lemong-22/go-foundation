# Technical Spec — Course Context · Phase B: Quiz

Status: Draft · Architecture: simplified Hexagonal + Clean + DDD (no CQRS)
Extends: `docs/spec.md`, `docs/phase-a-spec.md` · Roadmap: `docs/roadmap.md` Phase B · Date: 2026-05-26

## 0. Overview

- **Purpose:** introduce the **Quiz** aggregate (choice-based, deterministically
  auto-gradable) and the `quiz` value of `ContentBlockKind` so a `Lesson` can
  embed a quiz by reference.
- **Language / stack:** Go 1.22, PostgreSQL via `pgx`, `cobra` + `viper` —
  unchanged.
- **In scope:**
  - The `Quiz` aggregate root (scoped to a Course) and the `ChoiceQuestion`
    entity that lives inside it.
  - Two question types: single-choice and multiple-choice; per-question
    explanations; a quiz-level pass threshold.
  - Eleven new CLI commands (`quiz` CRUD + `quiz question` management).
  - The `quiz` value added to `ContentBlockKind`, the new `QuizBody`
    `ContentBody`, and the extension of `lesson block add` to embed a quiz.
  - Postgres persistence + migration `000003` (creates `quizzes` /
    `quiz_questions`; adds `quiz_ref` to `content_blocks`; widens the
    `content_blocks.kind` check constraint).
  - REST endpoints and a console quiz-builder at adapter level.
- **Out of scope:**
  - Short-answer / free-text question types — deferred (§6); they need
    exact-match heuristics or AI grading, a learner-phase concern.
  - Quiz attempts, scoring, grading — learner phase.
  - `Practice` (Phase C); `Test` (Phase D).
  - Bulk `quiz set-questions` — deferred to Phase E (import).
  - Quiz status (draft/published) and `quiz publish`/`unpublish` — deferred
    (§6); a quiz is visible when its embedding lesson/course is published.

**Architectural note.** `Quiz` is a **separate aggregate root** — different
from how `ContentBlock` lives *inside* the Lesson aggregate (`docs/phase-a-spec.md`
§2). A quiz block in a lesson holds only a `QuizID` reference, the standard DDD
rule that one aggregate refers to another by identity, never by holding the
other object — the same pattern `docs/spec.md` already uses for Lesson↔Course.

---

## 1. CLI Commands

Each command maps to **exactly one inbound port method and one usecase**.

### Quiz commands (5)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `quiz create` | Create a quiz in a course | `--course-id`, `--title`, `--pass-threshold` (opt) | new quiz id | course not found, missing title, invalid pass threshold |
| `quiz list` | List a course's quizzes | `--course-id`, `-o` | table / JSON / id-only | course not found |
| `quiz get` | Show a quiz with its questions | `<quiz-id>`, `-o` | quiz detail | quiz not found |
| `quiz update` | Edit quiz metadata | `<quiz-id>`, `--title` (opt), `--pass-threshold` (opt) | updated quiz id | quiz not found, nothing to update, invalid input |
| `quiz delete` | Delete a quiz | `<quiz-id>`, `--force` | confirmation | quiz not found, **quiz in use by a content block** |

### Question commands (6)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `quiz question add` | Add a question to a quiz | `--quiz-id`, `--type single\|multiple`, `--prompt`, `--option <text>` (repeatable, ≥2), `--correct <index>` (repeatable, 0-based), `--explanation` (opt), `--position` (opt) | new question id | quiz not found, invalid type, < 2 options, no correct indices, out-of-range correct, single-choice with ≠ 1 correct |
| `quiz question list` | List a quiz's questions, by position | `--quiz-id`, `-o` | table / JSON / id-only | quiz not found |
| `quiz question get` | Show a question's detail | `<question-id>`, `-o` | question detail | question not found |
| `quiz question update` | Edit a question (options + correct replaced atomically when given) | `<question-id>`, `--prompt` (opt), `--option` (repeatable, replaces all), `--correct` (repeatable, replaces all), `--explanation` (opt) | updated question id | question not found, nothing to update, options-without-correct, invariant violations |
| `quiz question remove` | Delete a question from its quiz | `<question-id>`, `--force` | confirmation | question not found |
| `quiz question reorder` | Resequence questions within a quiz | `--quiz-id`, `--order <question-id:pos,...>` | confirmation | quiz not found, unknown/foreign question, duplicate position |

### Changed — `lesson block add`

The Phase A spec deferred `quiz` from `ContentBlockKind` "to land in Phase B
with the command that creates it." This phase delivers it:

| Command | Change |
|---------|--------|
| `lesson block add` | Accepts `--kind quiz --quiz-id <id>` to embed a quiz block. New failure modes: quiz not found; **quiz belongs to a different course than the lesson's course** (cross-course embed rejected). |

### Inbound-adapter exit codes & errors

Unchanged from `docs/spec.md`: validation → 1, not found → 2, permission → 3,
internal → 5. The new domain error `ErrQuizInUse` carries the list of
embedding lesson ids and maps to a non-zero exit with a stderr message listing
the lessons so the author knows where to `lesson block remove` before retrying.

---

## 2. Domain Model

### Entities

**Quiz** *(new — aggregate root)* — a choice-based, auto-gradable assessment
owned by a course; ordered ChoiceQuestion entities inside.
- Identity: `ID QuizID`
- Fields:
  - `CourseID CourseID` (id-only reference into the Course aggregate)
  - `Title string`
  - `PassThreshold PassThreshold`
  - `Questions []ChoiceQuestion`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Title` is non-empty after trimming.
  - `CourseID` is always set.
  - `PassThreshold` is in `[0, 1]` (guaranteed by the VO).
  - **Question ids are unique** within the quiz.
  - **Question positions are unique** within the quiz; `Questions` is ordered
    by `Position` ascending.
  - `UpdatedAt >= CreatedAt`.
- Behaviour (each mutating method takes `now time.Time` and bumps `UpdatedAt`):
  - `Rename(title string, now time.Time) error` — re-validates non-empty.
  - `ChangePassThreshold(t PassThreshold, now time.Time)`.
  - `AddQuestion(q ChoiceQuestion, now time.Time) error` — inserts the
    question; if its `Position` collides with an existing question, questions
    at `>= Position` shift up by one (insert semantics), mirroring
    `Lesson.AddBlock`. A question constructed for "append" already carries
    `maxPosition + 1`.
  - `RemoveQuestion(id QuestionID, now time.Time) error` — removes and
    compacts remaining positions so they stay contiguous from 0.
  - `ReorderQuestions(order []QuestionPlacement, now time.Time) error` —
    placements must be a permutation of exactly the quiz's current questions.
  - `ChangeQuestionPrompt(id QuestionID, prompt string, now time.Time) error`.
  - `ChangeQuestionContent(id QuestionID, options []string, correct []int, now time.Time) error`
    — atomic options + correct update; re-validates the type/count
    invariants. Splitting "content" from "prompt" / "explanation" keeps each
    mutation method small.
  - `ChangeQuestionExplanation(id QuestionID, expl string, now time.Time) error`.

  All `ChangeQuestion*` methods locate the question by id (`ErrNotFound` if
  absent) and route the mutation through the aggregate root so invariants are
  always re-checked at the boundary.

**ChoiceQuestion** *(new — entity inside the Quiz aggregate)* — one ordered,
typed question. An entity (`QuestionID`, lifecycle) but **not** an aggregate
root.
- Identity: `ID QuestionID`
- Fields:
  - `Type ChoiceQuestionType`
  - `Prompt string`
  - `Options []string`
  - `CorrectIndices []int` (0-based, into `Options`)
  - `Explanation string` (optional; `""` when unset)
  - `Position QuestionPosition`
- Invariants:
  - `Prompt` is non-empty after trimming.
  - `len(Options) >= 2`.
  - `len(CorrectIndices) >= 1`.
  - Every entry in `CorrectIndices` is in `[0, len(Options))`.
  - `CorrectIndices` contains no duplicates.
  - If `Type.IsSingle()`, then `len(CorrectIndices) == 1`.
- Behaviour (called only by the Quiz aggregate root):
  - `ChangePrompt(prompt string) error` — re-validates non-empty.
  - `ChangeContent(options []string, correct []int) error` — replaces both
    atomically; re-validates all invariants above.
  - `ChangeExplanation(expl string)`.
  - `MoveTo(pos QuestionPosition)`.

### Value Objects

- **QuizID** — wraps `string`; invariant: non-empty, parseable as a UUID.
  Mirrors `LessonID`.
- **QuestionID** — wraps `string`; invariant: non-empty, parseable as a UUID.
- **QuestionPosition** — wraps `int`; invariant: `>= 0`. Mirrors `BlockPosition`.
- **ChoiceQuestionType** — wraps `string`; invariant: one of `single` |
  `multiple`. Exposes `SingleChoice()` / `MultipleChoice()` constructors and
  `IsSingle()` / `IsMultiple()` predicates.
- **PassThreshold** — wraps `float64`; invariant: `0 <= x <= 1`. Constructor
  `NewPassThreshold(f)` validates; `DefaultPassThreshold()` returns `0.7`.

### Extensions to existing Phase A types

- **`ContentBlockKind`** now admits a third value: `quiz`. Exposes
  `QuizKind()` constructor and `IsQuiz()` predicate.
- **`ContentBody`** gains a third implementation:

  ```go
  // QuizBody — payload of a `quiz` block.
  type QuizBody struct {
      QuizRef QuizID
  }
  ```

  `QuizBody.Kind()` returns `QuizKind()`. Like `TextBody` / `VideoBody` it is
  an immutable value object.

### Domain notes

- **Vertical slices.** Each of the eleven new commands is its own slice
  through the stack — consistent with `docs/spec.md`.
- **`ChoiceQuestion` has no repository.** It is persisted and hydrated only as
  part of its `Quiz` through `QuizRepository`. `QuizRepository.Save` writes
  the quiz row *and* its questions in one transaction. This mirrors how
  Phase A treats `ContentBlock` inside `Lesson`.
- **Loading the owning quiz from a question id.** `quiz question
  get/update/remove` take a bare `<question-id>` for ergonomics. The usecase
  loads the aggregate via a new `QuizRepository.FindByQuestionID` —
  symmetric with Phase A's `LessonRepository.FindByBlockID`.
- **Options stay bare `[]string` indexed by position.** Updates submit options
  and correct indices *together* (an atomic `ChangeContent`), so reordering
  options is consistent with reordering correct indices. Promoting options to
  their own entity with `QuestionOptionID` was considered and deferred (§6).
- **Empty quizzes are allowed during authoring.** A quiz can exist with zero
  questions immediately after `quiz create`; the pass threshold is only
  meaningful once questions exist. No invariant requires a minimum question
  count.
- **Cross-aggregate embed invariant.** When `lesson block add --kind quiz
  --quiz-id X`, the `AddLessonBlock` usecase verifies the quiz's `CourseID`
  equals the lesson's `CourseID`. Cross-course embedding is a validation
  error. This keeps each course's content graph self-contained.
- **Cross-aggregate delete — Course.** `DeleteCourse` (from
  `docs/spec.md` §4) gains one step: orchestrate `QuizRepository.DeleteByCourse(id)`
  alongside the existing `LessonRepository.DeleteByCourse(id)`. Lessons are
  deleted *first* so their content blocks (including quiz embeds) cascade
  away, then quizzes — preserving the `content_blocks.quiz_ref` FK as a
  `RESTRICT` backstop.
- **Cross-aggregate delete — Quiz (RESTRICT).** `DeleteQuiz` is *not* a
  silent cascade. It first calls `LessonRepository.FindLessonsEmbeddingQuiz(id)`;
  if any embedding lessons exist, it returns `ErrQuizInUse` carrying their
  ids. The author must `lesson block remove` the embeds explicitly before
  retrying. The FK on `content_blocks.quiz_ref` is `ON DELETE RESTRICT` as a
  safety backstop to the usecase's pre-check.
- **`Title`, `Prompt`, `Explanation` stay plain strings.** Their only rule
  ("non-empty after trim", and only for `Title` / `Prompt`) is enforced by the
  entity constructor — wrapping in a VO would be ceremony. Same discipline as
  `docs/spec.md`.

---

## 3. Ports

### Inbound ports (driving — called by an inbound adapter)

A new `QuizService` interface — one method per command:

```go
// QuizService — inbound port for all quiz + question commands.
type QuizService interface {
    // Quiz CRUD
    CreateQuiz(in CreateQuizInput) (CreateQuizOutput, error)
    ListQuizzes(in ListQuizzesInput) (ListQuizzesOutput, error)
    GetQuiz(in GetQuizInput) (GetQuizOutput, error)
    UpdateQuiz(in UpdateQuizInput) (UpdateQuizOutput, error)
    DeleteQuiz(in DeleteQuizInput) error

    // Question management
    AddQuestion(in AddQuestionInput) (AddQuestionOutput, error)
    ListQuestions(in ListQuestionsInput) (ListQuestionsOutput, error)
    GetQuestion(in GetQuestionInput) (GetQuestionOutput, error)
    UpdateQuestion(in UpdateQuestionInput) (UpdateQuestionOutput, error)
    RemoveQuestion(in RemoveQuestionInput) error
    ReorderQuestions(in ReorderQuestionsInput) error
}
```

`LessonService.AddLessonBlockInput` from `docs/phase-a-spec.md` gains one
field, `QuizRef string`, used when `Kind == "quiz"`.

DTOs carry only primitives and ids — the boundary between the adapter world
and the VO-typed domain.

```go
// --- Quiz DTOs ---

type CreateQuizInput struct {
    CourseID      string
    Title         string
    PassThreshold *float64 // nil = default 0.7
}
type CreateQuizOutput struct{ ID string }

type ListQuizzesInput struct{ CourseID string }
type ListQuizzesOutput struct{ Quizzes []QuizView }

type GetQuizInput struct{ ID string }
type GetQuizOutput struct{ Quiz QuizDetailView }

type UpdateQuizInput struct {
    ID            string
    Title         *string  // nil = leave unchanged
    PassThreshold *float64 // nil = leave unchanged
}
type UpdateQuizOutput struct{ ID string }

type DeleteQuizInput struct{ ID string }

// --- Question DTOs ---

type AddQuestionInput struct {
    QuizID         string
    Type           string   // "single" | "multiple"
    Prompt         string
    Options        []string
    CorrectIndices []int
    Explanation    string
    Position       *int     // nil = append at end
}
type AddQuestionOutput struct{ ID string }

type ListQuestionsInput struct{ QuizID string }
type ListQuestionsOutput struct{ Questions []QuestionView }

type GetQuestionInput struct{ ID string }
type GetQuestionOutput struct{ Question QuestionView }

type UpdateQuestionInput struct {
    ID             string
    Prompt         *string  // nil = leave unchanged
    Options        *[]string // nil = leave unchanged; non-nil replaces full list
    CorrectIndices *[]int    // REQUIRED when Options is non-nil
    Explanation    *string
}
type UpdateQuestionOutput struct{ ID string }

type RemoveQuestionInput struct{ ID string }

type ReorderQuestionsInput struct {
    QuizID string
    Order  []QuestionPlacementDTO
}
type QuestionPlacementDTO struct {
    QuestionID string
    Position   int
}

// --- Read models ---

type QuizView struct {
    ID, CourseID, Title  string
    PassThreshold        float64
    QuestionCount        int
    CreatedAt, UpdatedAt time.Time
}
type QuizDetailView struct {
    QuizView
    Questions []QuestionView
}
type QuestionView struct {
    ID, QuizID                 string
    Type                       string
    Prompt, Explanation        string
    Options                    []string
    CorrectIndices             []int
    Position                   int
}
```

The `UpdateQuestionInput` coupling — `CorrectIndices` is *required* when
`Options` is non-nil — is enforced by the usecase, not the DTO type system. A
caller that submits new options without saying which are correct gets a
validation error rather than a half-updated question.

### Outbound ports (driven — implemented by adapters)

New port: `QuizRepository`. Existing `LessonRepository` and `IDGenerator`
gain methods.

```go
type QuizRepository interface {
    Save(quiz Quiz) error                              // INSERT or UPDATE + replace questions
    FindByID(id QuizID) (Quiz, error)                  // ErrNotFound if absent; hydrates questions
    FindByCourse(courseID CourseID) ([]Quiz, error)    // ordered, e.g. by CreatedAt DESC
    FindByQuestionID(id QuestionID) (Quiz, error)      // owning aggregate from a question id
    Delete(id QuizID) error                            // ErrNotFound if absent
    DeleteByCourse(courseID CourseID) error            // used by DeleteCourse
}

type LessonRepository interface {
    // ... existing methods from docs/spec.md and docs/phase-a-spec.md ...

    // NEW Phase B — used by DeleteQuiz to enforce RESTRICT
    FindLessonsEmbeddingQuiz(quizID QuizID) ([]LessonID, error)
}

type IDGenerator interface {
    NewCourseID()    CourseID
    NewLessonID()    LessonID
    NewBlockID()     BlockID
    NewQuizID()      QuizID     // NEW
    NewQuestionID()  QuestionID // NEW
}

// Clock — unchanged.
```

No `MediaStore` (still deferred — `docs/phase-a-spec.md` §6).

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**PostgresQuizRepository** *(new)* implements `QuizRepository`
- Construction dependencies: `*pgxpool.Pool`
- `Save` — side effect: `INSERT INTO quizzes ... ON CONFLICT (id) DO UPDATE`,
  then `DELETE FROM quiz_questions WHERE quiz_id = $1` and re-`INSERT` the
  current question set — all in **one transaction**.
- `FindByID` — side effect: `SELECT ... FROM quizzes WHERE id = $1`; then
  `SELECT ... FROM quiz_questions WHERE quiz_id = $1 ORDER BY position`;
  hydrate.
- `FindByCourse` — side effect: `SELECT ... FROM quizzes WHERE course_id = $1
  ORDER BY created_at DESC`; bulk-load questions for the result set.
- `FindByQuestionID` — side effect: `SELECT q.* FROM quizzes q JOIN
  quiz_questions qq ON qq.quiz_id = q.id WHERE qq.id = $1`; then hydrate as
  above. Maps "0 rows" to `ErrNotFound`.
- `Delete` — side effect: `DELETE FROM quizzes WHERE id = $1` (cascades to
  `quiz_questions` via FK).
- `DeleteByCourse` — side effect: `DELETE FROM quizzes WHERE course_id = $1`.

**PostgresLessonRepository** *(modified)*
- `Save` / `FindByID` / `FindByCourse` — now also handle the
  `content_blocks.quiz_ref` column (alongside the text/video columns from
  Phase A). For a `kind = 'quiz'` block, the only persisted field is
  `quiz_ref`.
- `FindLessonsEmbeddingQuiz` *(new)* — side effect: `SELECT DISTINCT
  l.id FROM lessons l JOIN content_blocks b ON b.lesson_id = l.id WHERE
  b.kind = 'quiz' AND b.quiz_ref = $1`.

**UUIDGenerator** *(modified)*
- `NewQuizID` / `NewQuestionID` — pure; wrap a `google/uuid` v4 in the
  corresponding id VO.

**Migration `000003_add_quizzes`** — infrastructure, wired in `main.go`
outside the bounded context:

```sql
-- up
CREATE TABLE quizzes (
    id              UUID PRIMARY KEY,
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    pass_threshold  DOUBLE PRECISION NOT NULL DEFAULT 0.7
                    CHECK (pass_threshold >= 0 AND pass_threshold <= 1),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quiz_questions (
    id              UUID PRIMARY KEY,
    quiz_id         UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK (type IN ('single', 'multiple')),
    prompt          TEXT NOT NULL,
    options         JSONB NOT NULL,           -- string array
    correct_indices JSONB NOT NULL,           -- int array
    explanation     TEXT NOT NULL DEFAULT '',
    position        INTEGER NOT NULL CHECK (position >= 0),
    CONSTRAINT quiz_questions_position_unique UNIQUE (quiz_id, position)
);
CREATE INDEX quiz_questions_quiz_position_idx
    ON quiz_questions (quiz_id, position);

-- Extend content_blocks for the quiz kind
ALTER TABLE content_blocks DROP CONSTRAINT content_blocks_kind_check;
ALTER TABLE content_blocks ADD CONSTRAINT content_blocks_kind_check
    CHECK (kind IN ('text', 'video', 'quiz'));
ALTER TABLE content_blocks ADD COLUMN quiz_ref UUID
    REFERENCES quizzes(id) ON DELETE RESTRICT;
CREATE INDEX content_blocks_quiz_ref_idx ON content_blocks (quiz_ref);
```

The `down` migration reverses these steps in opposite order. `down` is safe
*only* if no production data has used `kind = 'quiz'` yet — once authors are
embedding quizzes, this migration is one-way in practice (consistent with the
Phase A note on `lessons.content`).

### Inbound adapters (no business logic — not skeletoned per-method)

- **CLI adapter** *(modified)* — extends the cobra tree with a `quiz`
  subcommand group and a `quiz question` sub-sub group; parses flags/config
  into the new DTOs, formats `QuizView` / `QuizDetailView` / `QuestionView`
  through the existing table/json/quiet writers, runs `--force` confirmation
  on the two destructive commands, maps `ErrQuizInUse` to a stderr message
  listing the embedding lesson ids before exiting non-zero. No business logic.
- **REST adapter** *(modified — Phase B endpoints)*:

  | Method & path | Inbound port method |
  |---------------|---------------------|
  | `POST /v1/quizzes` | `CreateQuiz` |
  | `GET /v1/courses/{courseId}/quizzes` | `ListQuizzes` |
  | `GET /v1/quizzes/{quizId}` | `GetQuiz` |
  | `PATCH /v1/quizzes/{quizId}` | `UpdateQuiz` |
  | `DELETE /v1/quizzes/{quizId}` | `DeleteQuiz` |
  | `POST /v1/quizzes/{quizId}/questions` | `AddQuestion` |
  | `GET /v1/quizzes/{quizId}/questions` | `ListQuestions` |
  | `GET /v1/questions/{questionId}` | `GetQuestion` |
  | `PATCH /v1/questions/{questionId}` | `UpdateQuestion` |
  | `DELETE /v1/questions/{questionId}` | `RemoveQuestion` |
  | `POST /v1/quizzes/{quizId}/questions/reorder` | `ReorderQuestions` |

  `ErrQuizInUse` maps to HTTP `409 Conflict` with a JSON body listing the
  embedding lesson ids.
- **Console** — a quiz-builder screen (metadata + an ordered question list
  with inline edit) and a quiz-picker in the lesson block editor that calls
  `lesson block add --kind quiz`. UI detail is not part of this domain spec.

### Usecases (implement inbound ports, depend on outbound ports)

The eleven new quiz/question usecases are grouped under a new
`QuizServiceImpl`. Two existing usecases are modified.

**CreateQuiz** — implements `QuizService.CreateQuiz`
- Depends on: `QuizRepository`, `CourseRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `CourseID` VO from `in.CourseID`
  2. `CourseRepository.FindByID(courseID)` — propagate `ErrNotFound`
  3. resolve `PassThreshold`: if `in.PassThreshold != nil` build the VO from
     `*in.PassThreshold` (validation errors propagate); else
     `DefaultPassThreshold()` (`0.7`)
  4. `id := IDGenerator.NewQuizID()`; `now := Clock.Now()`
  5. construct `Quiz` via `NewQuiz(id, courseID, in.Title, threshold, now)`
     — validates non-empty title; questions start empty
  6. `QuizRepository.Save(quiz)`
  7. return `CreateQuizOutput{ ID: id.String() }`

**ListQuizzes** — implements `QuizService.ListQuizzes`
- Depends on: `QuizRepository`, `CourseRepository`
- Steps:
  1. build `CourseID` VO; `CourseRepository.FindByID` — propagate
     `ErrNotFound`
  2. `QuizRepository.FindByCourse(courseID)`
  3. map `[]Quiz` → `[]QuizView`; return

**GetQuiz** — implements `QuizService.GetQuiz`
- Depends on: `QuizRepository`
- Steps:
  1. build `QuizID` VO
  2. `QuizRepository.FindByID(id)` — propagate `ErrNotFound`
  3. map to `QuizDetailView` (includes questions); return

**UpdateQuiz** — implements `QuizService.UpdateQuiz`
- Depends on: `QuizRepository`, `Clock`
- Steps:
  1. build `QuizID` VO; reject if both `Title` and `PassThreshold` are nil
     ("nothing to update")
  2. `QuizRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `now := Clock.Now()`
  4. if `in.Title` set: `quiz.Rename(*in.Title, now)`
  5. if `in.PassThreshold` set: build `PassThreshold` VO →
     `quiz.ChangePassThreshold(t, now)`
  6. `QuizRepository.Save(quiz)`
  7. return `UpdateQuizOutput{ ID: in.ID }`

**DeleteQuiz** — implements `QuizService.DeleteQuiz`
- Depends on: `QuizRepository`, `LessonRepository`
- Steps:
  1. build `QuizID` VO
  2. `QuizRepository.FindByID(id)` — propagate `ErrNotFound`
  3. `LessonRepository.FindLessonsEmbeddingQuiz(id)` — if non-empty, return
     `ErrQuizInUse` carrying the lesson ids
  4. `QuizRepository.Delete(id)`

**AddQuestion** — implements `QuizService.AddQuestion`
- Depends on: `QuizRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `QuizID` VO; `QuizRepository.FindByID` — propagate `ErrNotFound`
  2. build `ChoiceQuestionType` VO from `in.Type`
  3. determine `QuestionPosition`: if `in.Position != nil`, build the VO;
     else append (`maxPosition + 1`, or `0` if no questions)
  4. `id := IDGenerator.NewQuestionID()`; `now := Clock.Now()`
  5. construct `ChoiceQuestion` via `NewChoiceQuestion(id, type, in.Prompt,
     in.Options, in.CorrectIndices, in.Explanation, position)` — validates
     all invariants (`Prompt` non-empty, `>= 2` options, `>= 1` correct,
     correct in range, no duplicate correct, single-choice with exactly 1
     correct)
  6. `quiz.AddQuestion(question, now)` — inserts with shift if needed
  7. `QuizRepository.Save(quiz)`
  8. return `AddQuestionOutput{ ID: id.String() }`

**ListQuestions** — implements `QuizService.ListQuestions`
- Depends on: `QuizRepository`
- Steps:
  1. build `QuizID`; `FindByID` — propagate `ErrNotFound`
  2. map `quiz.Questions` (already ordered) → `[]QuestionView`; return

**GetQuestion** — implements `QuizService.GetQuestion`
- Depends on: `QuizRepository`
- Steps:
  1. build `QuestionID` VO
  2. `QuizRepository.FindByQuestionID(id)` — propagate `ErrNotFound`
  3. locate the question within the quiz; map to `QuestionView`; return

**UpdateQuestion** — implements `QuizService.UpdateQuestion`
- Depends on: `QuizRepository`, `Clock`
- Steps:
  1. build `QuestionID` VO; reject if all of `Prompt`, `Options`,
     `CorrectIndices`, `Explanation` are nil ("nothing to update")
  2. if `Options != nil` and `CorrectIndices == nil`, return a validation
     error — the two must be supplied together
  3. `QuizRepository.FindByQuestionID(id)` — propagate `ErrNotFound`
  4. `now := Clock.Now()`
  5. if `in.Prompt` set: `quiz.ChangeQuestionPrompt(id, *Prompt, now)` —
     propagate validation error
  6. if `in.Options` set: `quiz.ChangeQuestionContent(id, *Options,
     *CorrectIndices, now)` — re-validates `>= 2` options, correct in range,
     type/count constraint
  7. if `in.Explanation` set: `quiz.ChangeQuestionExplanation(id,
     *Explanation, now)`
  8. `QuizRepository.Save(quiz)`
  9. return `UpdateQuestionOutput{ ID: in.ID }`

**RemoveQuestion** — implements `QuizService.RemoveQuestion`
- Depends on: `QuizRepository`, `Clock`
- Steps:
  1. build `QuestionID` VO
  2. `QuizRepository.FindByQuestionID(id)` — propagate `ErrNotFound`
  3. `quiz.RemoveQuestion(id, Clock.Now())` — removes and compacts positions
  4. `QuizRepository.Save(quiz)`

**ReorderQuestions** — implements `QuizService.ReorderQuestions`
- Depends on: `QuizRepository`, `Clock`
- Steps:
  1. build `QuizID`; `FindByID` — propagate `ErrNotFound`
  2. index the quiz's questions by id
  3. for each placement: parse `QuestionID`; verify it exists in the index and
     belongs to this quiz (validation error otherwise); build
     `QuestionPosition` VO
  4. reject duplicate positions; require the placement set to be a permutation
     of exactly the quiz's current questions
  5. `now := Clock.Now()`; `quiz.ReorderQuestions(placements, now)`
  6. `QuizRepository.Save(quiz)`

### Modified existing usecases

**AddLessonBlock** *(from `docs/phase-a-spec.md`)* — gains `QuizRepository`
as a dependency and a new branch for `kind == "quiz"`:
- After step 3 (building `ContentBlockKind`):
  - if `text` → `TextBody{Markdown: in.Markdown}` (existing)
  - if `video` → build `MediaProvider` + `MediaRef`; `VideoBody{...}` (existing)
  - if `quiz` *(new)*:
    1. build `QuizID` from `in.QuizRef`
    2. `QuizRepository.FindByID(quizID)` — propagate `ErrNotFound`
    3. verify `quiz.CourseID == lesson.CourseID` — else return validation
       error `ErrCrossCourseQuizEmbed`
    4. body = `QuizBody{ QuizRef: quizID }`
- Remaining steps (`AddBlock`, `Save`, return) — unchanged.

**DeleteCourse** *(from `docs/spec.md` §4)* — gains `QuizRepository`. New step
between the existing lesson and course deletion:
- 3. `LessonRepository.DeleteByCourse(id)` *(existing)*
- 3a. `QuizRepository.DeleteByCourse(id)` *(new)* — safe to run after step 3
      because all `content_blocks.quiz_ref` rows in this course have already
      cascaded away with their lessons.
- 4. `CourseRepository.Delete(id)` *(existing)*

---

## 5. Container & Wiring

The composition root grows by one repository, two id-generator methods, and
one service:

```go
func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
    // 1. outbound adapters
    pool, err := pgxpool.New(ctx, cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("connect db: %w", err)
    }
    courseRepo := NewPostgresCourseRepository(pool)
    lessonRepo := NewPostgresLessonRepository(pool) // now block + quiz_ref aware
    quizRepo   := NewPostgresQuizRepository(pool)   // NEW
    ids        := NewUUIDGenerator()                // now also serves NewQuizID, NewQuestionID
    clock      := NewSystemClock()

    // 2. usecases — existing service constructors gain quizRepo (cross-aggregate
    //    concerns: DeleteCourse cascade, AddLessonBlock kind=quiz)
    courseSvc := NewCourseServiceImpl(courseRepo, lessonRepo, quizRepo, ids, clock)
    lessonSvc := NewLessonServiceImpl(courseRepo, lessonRepo, quizRepo, ids, clock)
    quizSvc   := NewQuizServiceImpl(courseRepo, lessonRepo, quizRepo, ids, clock) // NEW

    // 3. bind usecases to the inbound ports
    return &CLI{Course: courseSvc, Lesson: lessonSvc, Quiz: quizSvc}, nil
}
```

`main.go` mounts the same inbound adapters (cobra CLI, REST server, console,
playground) onto the returned services — the REST server now exposes the new
quiz/question endpoints simply by being constructed with `container.Quiz` in
addition to the existing services.

Because every usecase receives `QuizRepository` (and the others) as an
**interface**, swapping the Postgres adapter for an in-memory fake in tests is
the same one-line change to `NewPostgresQuizRepository`. The new
`FindLessonsEmbeddingQuiz` method on `LessonRepository` follows the same
discipline — fakes implement it the same way as `FindByID`.

---

## 6. Deferred

Judgment-call items considered during this design pass and deliberately not
built now. Parked in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| Short-answer / free-text question types | anticipatory | Would need exact-match heuristics or AI grading — a learner-phase concern. `ChoiceQuestion` ships single/multiple only, per the grill-me decision. |
| Quiz status (draft/published) and `quiz publish`/`unpublish` | anticipatory | No command publishes a quiz; visibility follows from the embedding lesson/course's publish status. Phase A's "model a value with a command" discipline applies — add the field with the verb. |
| `quiz duplicate` (copy across courses) | nice-to-have | No current authoring activity drives it; revisit when cross-course templating is a real need. |
| Bulk `quiz set-questions` | nice-to-have | Per-question commands cover Phase B authoring; bulk is most useful to the Phase E import agent — add it there. |
| Question option entities (with `QuestionOptionID`) | optimization | Atomic options+correct updates make string+index sufficient; revisit only if independent option lifecycle ever becomes a real need. |
| Per-question `CreatedAt` / `UpdatedAt` | nice-to-have | No command shows per-question times; the Quiz's `UpdatedAt` already covers "quiz changed." Phase A precedent. |
| Question bank (questions reused across quizzes) | anticipatory | Each Quiz owns its questions — same principle Phase D's `Test` will apply to its items. No cross-quiz reuse means simpler aggregate boundaries. |
| Richer question types (true/false, ordering, matching, code-output prediction) | anticipatory | Phase B is choice-based only; richer types added when a command needs them. |
| `ChoiceQuestion.MinOptions` configurable | optimization | `>= 2` is hard-coded; no use case argues for a quiz-level or question-level minimum yet. |
