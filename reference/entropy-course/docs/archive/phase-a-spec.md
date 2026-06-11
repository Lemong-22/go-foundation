# Technical Spec — Course Context · Phase A: Content Blocks & Media

Status: Draft · Architecture: simplified Hexagonal + Clean + DDD (no CQRS)
Extends: `docs/spec.md` · Roadmap: `docs/roadmap.md` Phase A · Date: 2026-05-26

## 0. Overview

- **Purpose:** restructure the **Lesson** aggregate so a lesson's body is an
  *ordered sequence of typed content blocks* (`text` and `video`) instead of a
  single `content` text field. This establishes the structural seam that the
  `quiz` and `practice` block kinds slot into in Phases B–C.
- **Language / stack:** Go 1.22, PostgreSQL via `pgx`, `cobra` + `viper` for
  the CLI, `google/uuid` — unchanged from `docs/spec.md`.
- **In scope:**
  - A `ContentBlock` **entity** living *inside* the Lesson aggregate.
  - Two block kinds: `text` (markdown) and `video` (a media reference).
  - The `MediaRef` value object.
  - Six `lesson block` CLI commands.
  - Postgres persistence for blocks + migration `000002`.
  - The changes this forces on `lesson create` and `lesson update`.
- **Out of scope:**
  - **Quiz / Practice / Test** aggregates — Phases B–D.
  - The **`MediaStore` outbound port** — deferred (see §6). Phase A media is a
    pure value-object invariant; no usecase has a media side effect yet.
  - Per-block audit timestamps — deferred (see §6).
  - The **REST adapter** and the **console**. Both are *inbound adapters* that
    wrap the same inbound ports; they are acknowledged at the adapter level in
    §4 but not domain-skeletoned — the same treatment `docs/spec.md` gives the
    CLI adapter, which "holds no business logic."

**Aggregate note.** `ContentBlock` is an entity *within* the Lesson aggregate —
**not** a separate aggregate root. It has no repository of its own; blocks are
persisted and loaded as part of their `Lesson` through `LessonRepository`. This
is different from the Lesson↔Course relationship, where the two are *separate*
aggregates that reference each other by id (`docs/spec.md` §2). The Lesson
aggregate is the consistency and persistence boundary for its blocks.

---

## 1. CLI Commands

### New — `lesson block` commands

Six commands, each mapping to **exactly one inbound port method and one
usecase**.

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `lesson block add` | Append or insert a block in a lesson | `--lesson-id`, `--kind text\|video`; text: `--text`; video: `--video-provider`, `--video-locator`, `--video-caption` (opt); `--position` (opt) | prints new block id | lesson not found, invalid kind, missing kind-required input, invalid media ref, negative position |
| `lesson block list` | List a lesson's blocks, by position | `--lesson-id`, `-o/--output` | table / JSON / id-only of blocks | lesson not found |
| `lesson block get` | Show one block's detail | `<block-id>`, `-o/--output` | block detail | block not found |
| `lesson block update` | Edit a block's payload | `<block-id>`; text: `--text`; video: `--video-provider`/`--video-locator`/`--video-caption` | prints updated block id | block not found, nothing to update, invalid input |
| `lesson block remove` | Delete a block from its lesson | `<block-id>`, `--force` (opt) | confirmation | block not found |
| `lesson block reorder` | Resequence blocks within a lesson | `--lesson-id`, `--order <block-id:pos,...>` | confirmation | lesson not found, unknown block id, block not in lesson, duplicate position |

A block's **kind is fixed at creation**. `lesson block update` edits the payload
of an existing block but never changes its kind; to switch kind, `remove` then
`add` (see §6).

### Changed — existing lesson commands

The flat `content` field is gone, so two existing commands lose a flag:

| Command | Change |
|---------|--------|
| `lesson create` | `--content` flag **removed**. A new lesson is created with **zero blocks**; its body is built afterwards with `lesson block add`. Remaining inputs: `--course-id`, `--title`, `--order` (opt). |
| `lesson update` | `--content` flag **removed**. The command now edits **title only**; "nothing to update" fires when `--title` is absent. |

`lesson get` / `lesson list` are unchanged in surface, but their detail output
no longer carries `content` (block bodies are retrieved via `lesson block`).

### Inbound-adapter exit codes

Unchanged from `docs/spec.md`: validation → 1, not found → 2, permission → 3,
internal → 5. The new commands surface the same typed domain errors; the CLI
adapter owns the mapping.

---

## 2. Domain Model

### Entities

**Lesson** *(modified)* — a unit of content inside a course; now the aggregate
root for an ordered set of content blocks.
- Identity: `ID LessonID`
- Fields:
  - `CourseID CourseID` (id-only reference to the owning Course aggregate)
  - `Title string`
  - `Blocks []ContentBlock` *(new — replaces `Content string`)*
  - `Order LessonOrder`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Title` is non-empty after trimming.
  - `Order` is `>= 0` (guaranteed by the VO).
  - `CourseID` is always set.
  - **Block ids are unique** within the lesson.
  - **Block positions are unique** within the lesson; `Blocks` is kept ordered
    by `Position` ascending.
- Behaviour (each mutating method takes `now time.Time` and bumps `UpdatedAt`
  via the existing `touch`):
  - `Rename(title string, now time.Time) error` — unchanged.
  - `MoveTo(order LessonOrder, now time.Time)` — unchanged.
  - `AddBlock(block ContentBlock, now time.Time) error` *(new)* — inserts the
    block. If its `Position` collides with an existing block, blocks at
    `>= Position` shift up by one (insert semantics); a block constructed for
    "append" already carries `maxPosition + 1`.
  - `UpdateBlock(id BlockID, body ContentBody, now time.Time) error` *(new)* —
    finds the block, calls `block.ChangeBody(body)` (which re-checks the
    kind/body match), bumps `UpdatedAt`. `ErrNotFound` if no such block.
  - `RemoveBlock(id BlockID, now time.Time) error` *(new)* — removes the block
    and compacts the remaining positions so they stay contiguous from 0.
  - `ReorderBlocks(order []BlockPlacement, now time.Time) error` *(new)* —
    applies a new position to each block; the placement set must be a
    permutation of exactly the lesson's current blocks.
- Constructors `NewLesson` / `RestoreLesson` change signature: the `content
  string` parameter becomes `blocks []ContentBlock`.

**ContentBlock** *(new)* — one ordered, typed unit of a lesson's body. An
entity (it has identity and a lifecycle: created, edited, moved, removed), but
**not** an aggregate root — it lives inside `Lesson`.
- Identity: `ID BlockID`
- Fields:
  - `Kind ContentBlockKind`
  - `Position BlockPosition`
  - `Body ContentBody`
- Invariants:
  - `Body.Kind()` equals `Kind` — a `text` block always holds a `TextBody`, a
    `video` block always holds a `VideoBody`.
- Behaviour:
  - `ChangeBody(body ContentBody) error` — replaces the body; returns a
    validation error if `body.Kind()` does not match the block's `Kind`.
  - `MoveTo(pos BlockPosition)` — sets a new position (called by the aggregate
    root during `ReorderBlocks` / insert shifts).

### Value Objects

All value objects are immutable; `NewX` fails with a domain error when an
invariant is violated, so an existing VO is always valid.

- **BlockID** — wraps `string`; invariant: non-empty, parseable as a UUID.
  Mirrors `LessonID` / `CourseID`.
- **ContentBlockKind** — wraps `string`; invariant: one of `text` | `video`
  (Phase A set). Exposes `TextKind()` / `VideoKind()` constructors and
  `IsText()` / `IsVideo()` predicates. (`quiz` / `practice` are deferred — §6.)
- **BlockPosition** — wraps `int`; invariant: `>= 0`. Mirrors `LessonOrder`.
- **MediaProvider** — wraps `string`; invariant: one of `url` | `youtube` |
  `mux` (Phase A set).
- **MediaRef** — wraps `MediaProvider` + `locator string`; invariant: `locator`
  is non-empty and well-formed *for its provider* — `url` → an absolute
  `http`/`https` URL; `youtube` → a video id or watch URL; `mux` → a playback
  id. Construction fails otherwise. This is the *only* media validation in
  Phase A and it is a pure invariant — hence no outbound port (see §6).

**ContentBody** — Go has no sum types, so the kind-specific payload is modeled
as a sealed interface with one immutable value-object implementation per kind:

```go
type ContentBody interface {
    Kind() ContentBlockKind
    isContentBody() // unexported — seals the interface to this package
}

// TextBody — payload of a `text` block.
type TextBody struct {
    Markdown string // optional; "" is valid, as the old `content` field was
}

// VideoBody — payload of a `video` block.
type VideoBody struct {
    Media   MediaRef
    Caption string // optional
}
```

`TextBody.Kind()` returns `TextKind()`; `VideoBody.Kind()` returns
`VideoKind()`. Both are value objects: immutable, identity-free, compared by
value.

### Domain notes

- **Vertical slices.** Each of the six new commands is its own slice through the
  stack — CLI → inbound port → usecase → outbound port → adapter — consistent
  with `docs/spec.md`.
- **`ContentBlock` has no repository.** It is persisted and hydrated only as
  part of its `Lesson`. `LessonRepository.Save` writes the lesson row *and* its
  blocks in one transaction; `FindByID` hydrates the blocks. This is the
  aggregate boundary doing its job.
- **Loading a lesson from a block id.** `lesson block get/update/remove` take a
  bare `<block-id>` for ergonomics. Because blocks have no repository, the
  usecase loads the *owning aggregate* via a new
  `LessonRepository.FindByBlockID` method.
- **`Title` and `TextBody.Markdown` stay plain strings.** A title's only rule
  (non-empty) is enforced by the Lesson constructor; markdown has no invariant
  at all — wrapping either in a VO would be ceremony without payoff, the same
  reasoning `docs/spec.md` applied to `Title` / `Description`.
- **Timestamps stay passed-in.** New mutating methods (`AddBlock`,
  `UpdateBlock`, `RemoveBlock`, `ReorderBlocks`) receive `now time.Time` from
  the usecase's `Clock`, never pull a clock — unchanged discipline.

---

## 3. Ports

### Inbound ports (driving — called by an inbound adapter)

`LessonService` from `docs/spec.md` §3 gains six methods. `UpdateLessonInput`
loses its `Content` field; `CreateLessonInput` loses `Content`.

```go
// LessonService — inbound port for all lesson commands.
type LessonService interface {
    // --- existing (CreateLessonInput / UpdateLessonInput lose `Content`) ---
    CreateLesson(in CreateLessonInput) (CreateLessonOutput, error)
    ListLessons(in ListLessonsInput) (ListLessonsOutput, error)
    GetLesson(in GetLessonInput) (GetLessonOutput, error)
    UpdateLesson(in UpdateLessonInput) (UpdateLessonOutput, error)
    DeleteLesson(in DeleteLessonInput) error
    ReorderLessons(in ReorderLessonsInput) error

    // --- new — Phase A ---
    AddLessonBlock(in AddLessonBlockInput) (AddLessonBlockOutput, error)
    ListLessonBlocks(in ListLessonBlocksInput) (ListLessonBlocksOutput, error)
    GetLessonBlock(in GetLessonBlockInput) (GetLessonBlockOutput, error)
    UpdateLessonBlock(in UpdateLessonBlockInput) (UpdateLessonBlockOutput, error)
    RemoveLessonBlock(in RemoveLessonBlockInput) error
    ReorderLessonBlocks(in ReorderLessonBlocksInput) error
}
```

DTOs carry only primitives and ids — the boundary between the untyped adapter
world and the VO-typed domain. The block body crosses this boundary *flattened*
into kind-tagged primitive fields; the usecase reassembles the `ContentBody`.

```go
// --- block DTOs ---

type AddLessonBlockInput struct {
    LessonID      string
    Kind          string // "text" | "video"
    Markdown      string // text kind
    VideoProvider string // video kind
    VideoLocator  string // video kind
    VideoCaption  string // video kind, optional
    Position      *int   // nil = append at end
}
type AddLessonBlockOutput struct{ ID string }

type ListLessonBlocksInput struct{ LessonID string }
type ListLessonBlocksOutput struct{ Blocks []BlockView }

type GetLessonBlockInput struct{ ID string }
type GetLessonBlockOutput struct{ Block BlockView }

type UpdateLessonBlockInput struct {
    ID            string
    Markdown      *string // text kind; nil = leave unchanged
    VideoProvider *string // video kind; nil = leave unchanged
    VideoLocator  *string // video kind; nil = leave unchanged
    VideoCaption  *string // video kind; nil = leave unchanged
}
type UpdateLessonBlockOutput struct{ ID string }

type RemoveLessonBlockInput struct{ ID string }

type ReorderLessonBlocksInput struct {
    LessonID string
    Order    []BlockPlacementDTO
}
type BlockPlacementDTO struct {
    BlockID  string
    Position int
}

// BlockView — flat read-model for table/json/quiet formatting.
type BlockView struct {
    ID            string
    LessonID      string
    Kind          string
    Position      int
    Markdown      string // populated for text blocks
    VideoProvider string // populated for video blocks
    VideoLocator  string // populated for video blocks
    VideoCaption  string // populated for video blocks
}
```

### Outbound ports (driven — implemented by adapters)

`LessonRepository` from `docs/spec.md` §3 gains one method; its existing
`Save` / `FindByID` / `FindByCourse` now also handle blocks. `IDGenerator`
gains one method. **No `BlockRepository`** (blocks are inside the aggregate)
and **no `MediaStore`** (deferred — §6).

```go
type LessonRepository interface {
    Save(l Lesson) error                              // now persists blocks too
    SaveAll(ls []Lesson) error
    FindByID(id LessonID) (Lesson, error)              // now hydrates blocks
    FindByCourse(courseID CourseID) ([]Lesson, error)  // now hydrates blocks
    FindByBlockID(id BlockID) (Lesson, error)          // NEW — owning lesson
    Delete(id LessonID) error
    DeleteByCourse(courseID CourseID) error
}

type IDGenerator interface {
    NewCourseID() CourseID
    NewLessonID() LessonID
    NewBlockID()  BlockID  // NEW
}

// Clock — unchanged.
type Clock interface {
    Now() time.Time
}
```

**The dependency rule holds.** Every port is declared inside the core and
speaks only domain types. No new outbound port is added "in case" — `MediaRef`
validation is a constructor invariant, so Phase A needs no media port.

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**PostgresLessonRepository** implements `LessonRepository` *(modified)*
- Construction dependencies: `*pgxpool.Pool` — unchanged.
- `Save` — side effect: `INSERT ... ON CONFLICT (id) DO UPDATE` on `lessons`,
  **then** replace the lesson's `content_blocks` rows (delete rows for the
  lesson, insert the current set) — all inside **one transaction**.
- `FindByID` / `FindByCourse` — side effect: `SELECT` the lesson row(s), then
  `SELECT * FROM content_blocks WHERE lesson_id = ANY(...) ORDER BY position`
  and hydrate each `Lesson.Blocks`.
- `FindByBlockID` — side effect: `SELECT l.* FROM lessons l JOIN content_blocks
  b ON b.lesson_id = l.id WHERE b.id = $1`; then hydrate blocks as above. Maps
  "0 rows" to `ErrNotFound`.
- `Delete` / `DeleteByCourse` — unchanged; `content_blocks` rows disappear via
  the FK `ON DELETE CASCADE`.

**UUIDGenerator** implements `IDGenerator` *(modified)*
- `NewBlockID` — pure; wraps a `google/uuid` v4 in the `BlockID` VO.

**Migration `000002_add_content_blocks`** — infrastructure, wired in `main.go`
outside the bounded context (per `docs/spec.md` §6), not a domain artifact:
- `up`: create `content_blocks` —
  `id UUID PK`, `lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE
  CASCADE`, `kind TEXT NOT NULL CHECK (kind IN ('text','video'))`,
  `position INTEGER NOT NULL CHECK (position >= 0)`,
  `text_markdown TEXT`, `video_provider TEXT`, `video_locator TEXT`,
  `video_caption TEXT`, `UNIQUE (lesson_id, position)`, index on
  `(lesson_id, position)`.
- `up` backfill: for every existing lesson, insert one row —
  `kind = 'text'`, `position = 0`, `text_markdown = lessons.content`.
- The legacy `lessons.content` column is **kept** as a rollback safety net; a
  later migration drops it once nothing reads it (§6).
- `down`: `DROP TABLE content_blocks`.

### Inbound adapters (no business logic — not skeletoned per-method)

- **CLI adapter** *(modified)* — extends the `cobra` tree with a `lesson block`
  subcommand group; parses flags/config into the block DTOs, calls the inbound
  port, formats `BlockView` through the existing table/json/quiet writers, runs
  `--force` confirmation on `remove`, maps domain errors to exit codes. Holds no
  business logic — exactly as the existing CLI adapter.
- **REST adapter** *(new — Phase A)* — a `net/http` inbound adapter. JSON
  endpoints map 1:1 onto the inbound ports; a static bearer-token middleware
  guards every route (token from `COURSE_CLI_API_TOKEN` / config). It is an
  adapter only: parse request → call inbound port → encode response. Domain
  errors map to status codes (validation → 400, not found → 404, unauthorized
  → 401, internal → 500), the HTTP analogue of the CLI exit-code table.

  | Method & path | Inbound port method |
  |---------------|---------------------|
  | `POST /v1/lessons/{lessonId}/blocks` | `AddLessonBlock` |
  | `GET /v1/lessons/{lessonId}/blocks` | `ListLessonBlocks` |
  | `GET /v1/blocks/{blockId}` | `GetLessonBlock` |
  | `PATCH /v1/blocks/{blockId}` | `UpdateLessonBlock` |
  | `DELETE /v1/blocks/{blockId}` | `RemoveLessonBlock` |
  | `POST /v1/lessons/{lessonId}/blocks/reorder` | `ReorderLessonBlocks` |

  (Plus the existing course + lesson endpoints at parity — same pattern.)
- **Console** *(new — Phase A)* — a server-rendered `html/template` + htmx app,
  another consumer of the same inbound ports. Its UI is not part of this domain
  spec.

### Usecases (implement inbound ports, depend on outbound ports)

The six block usecases are grouped under the existing `LessonServiceImpl`.
Each is constructor-injected with **interfaces** only.

**AddLessonBlock** — implements `LessonService.AddLessonBlock`
- Depends on: `LessonRepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `LessonID` VO from `in.LessonID`
  2. `LessonRepository.FindByID(lessonID)` — propagate `ErrNotFound`
  3. build `ContentBlockKind` VO from `in.Kind` — validation error if invalid
  4. build the `ContentBody`:
     - `text` → `TextBody{ Markdown: in.Markdown }`
     - `video` → build `MediaProvider` + `MediaRef` VOs from
       `in.VideoProvider` / `in.VideoLocator` (validation errors propagate) →
       `VideoBody{ Media: ref, Caption: in.VideoCaption }`
  5. determine `BlockPosition`: if `in.Position` set, build the VO from it;
     else `maxPosition + 1` (or `0` when the lesson has no blocks)
  6. `id := IDGenerator.NewBlockID()`; `now := Clock.Now()`
  7. construct `ContentBlock{ id, kind, position, body }` — validates
     kind/body match
  8. `lesson.AddBlock(block, now)` — inserts, shifting later blocks if needed
  9. `LessonRepository.Save(lesson)`
  10. return `AddLessonBlockOutput{ ID: id.String() }`

**ListLessonBlocks** — implements `LessonService.ListLessonBlocks`
- Depends on: `LessonRepository`
- Steps:
  1. build `LessonID` VO; `FindByID` — propagate `ErrNotFound`
  2. map `lesson.Blocks` (already ordered by position) → `[]BlockView`
  3. return `ListLessonBlocksOutput`

**GetLessonBlock** — implements `LessonService.GetLessonBlock`
- Depends on: `LessonRepository`
- Steps:
  1. build `BlockID` VO from `in.ID`
  2. `LessonRepository.FindByBlockID(blockID)` — propagate `ErrNotFound`
  3. locate the block within the lesson; map to `BlockView`
  4. return `GetLessonBlockOutput`

**UpdateLessonBlock** — implements `LessonService.UpdateLessonBlock`
- Depends on: `LessonRepository`, `Clock`
- Steps:
  1. build `BlockID` VO; reject if no body field is set ("nothing to update")
  2. `LessonRepository.FindByBlockID(blockID)` — propagate `ErrNotFound`
  3. locate the block; read its `Kind`
  4. `now := Clock.Now()`
  5. build the new `ContentBody` from the existing body plus the set input
     fields appropriate to the block's kind (so partial edits — e.g. caption
     only — keep the rest); validation errors propagate. Input fields for the
     *other* kind are ignored.
  6. `lesson.UpdateBlock(blockID, newBody, now)` — re-checks kind/body match,
     bumps `lesson.UpdatedAt`
  7. `LessonRepository.Save(lesson)`
  8. return `UpdateLessonBlockOutput{ ID: in.ID }`

**RemoveLessonBlock** — implements `LessonService.RemoveLessonBlock`
- Depends on: `LessonRepository`, `Clock`
- Steps:
  1. build `BlockID` VO
  2. `LessonRepository.FindByBlockID(blockID)` — propagate `ErrNotFound`
  3. `lesson.RemoveBlock(blockID, Clock.Now())` — removes and compacts
     positions
  4. `LessonRepository.Save(lesson)`

**ReorderLessonBlocks** — implements `LessonService.ReorderLessonBlocks`
- Depends on: `LessonRepository`, `Clock`
- Steps:
  1. build `LessonID` VO; `FindByID` — propagate `ErrNotFound`
  2. index the lesson's blocks by id
  3. for each `BlockPlacementDTO`: parse `BlockID`; verify it exists in the
     index and belongs to this lesson (validation error otherwise); build
     `BlockPosition` VO
  4. reject duplicate positions; require the placement set to be a permutation
     of exactly the lesson's current blocks
  5. `now := Clock.Now()`; `lesson.ReorderBlocks(placements, now)`
  6. `LessonRepository.Save(lesson)`

**Modified existing usecases**
- `CreateLesson` — `CreateLessonInput.Content` removed; the new `Lesson` is
  built with an empty `[]ContentBlock`.
- `UpdateLesson` — `UpdateLessonInput.Content` removed; "nothing to update"
  now fires when `Title` is nil.
- `DeleteCourse` / `DeleteLesson` — unchanged; block rows cascade via the FK.

---

## 5. Container & Wiring

The composition root barely moves — that is the hexagonal payoff. The block
feature adds usecases and outbound-port *methods*, but no new port type and no
new adapter type, so `BuildContainer` constructs the same objects with the same
calls.

```go
func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
    // 1. outbound adapters — unchanged construction
    pool, err := pgxpool.New(ctx, cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("connect db: %w", err)
    }
    courseRepo := NewPostgresCourseRepository(pool)
    lessonRepo := NewPostgresLessonRepository(pool) // now block-aware internally
    ids        := NewUUIDGenerator()                // now also serves NewBlockID
    clock      := NewSystemClock()

    // 2. usecases — same constructor signatures; the block usecases are
    //    grouped inside LessonServiceImpl, which already holds these ports
    courseSvc := NewCourseServiceImpl(courseRepo, lessonRepo, ids, clock)
    lessonSvc := NewLessonServiceImpl(courseRepo, lessonRepo, ids, clock)

    // 3. bind usecases to the inbound ports
    return &CLI{Course: courseSvc, Lesson: lessonSvc}, nil
}
```

`main.go` then mounts an inbound adapter onto the returned services — the
`cobra` CLI, the new REST server, or the console — exactly as it already mounts
the loopback playground. The REST server is constructed from the same
`CourseService` / `LessonService` values:

```go
// in main.go, outside BuildContainer — REST is just another inbound adapter
restServer := rest.NewServer(rest.Options{
    Course: container.Course,
    Lesson: container.Lesson,
    Token:  cfg.APIToken,
})
```

Because every dependency is an interface, the entire block feature — three new
DTO groups, six usecases, a new repository method, a new id-generator method —
required **zero** changes to the wiring above. Swapping Postgres for an
in-memory fake in tests is still the same one-line change to
`NewPostgresLessonRepository`.

---

## 6. Deferred

Judgment-call items considered during this design pass and deliberately not
built now. Parked in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| `MediaStore` outbound port | anticipatory | Phase A media validation is a pure `MediaRef` constructor invariant (locator format) — no side effect, so no usecase needs an outbound port. The port arrives with the media-upload phase, when verifying or uploading an asset is a real side effect. (`docs/roadmap.md` amended to match.) |
| Per-block `CreatedAt` / `UpdatedAt` | nice-to-have | No command displays or sorts by per-block times; any block mutation already bumps the parent `Lesson.UpdatedAt`. |
| `quiz` / `practice` block kinds | anticipatory | `ContentBlockKind` ships `text` / `video` only; the values land in Phases B–C with the commands that create them — the same discipline `docs/spec.md` §6 applied to lesson status. |
| Bulk `lesson block set` (replace whole block list) | nice-to-have | Per-block commands cover Phase A authoring; a bulk replace is most useful to the Phase E import agent — add it there. |
| Changing a block's `Kind` in place | nice-to-have | `lesson block update` keeps a block's kind fixed; switching kind is `remove` + `add`. No command needs in-place kind change. |
| Dropping the legacy `lessons.content` column | optimization | Kept through the transition as a rollback safety net; a later migration drops it once nothing reads it. |
| Rich `MediaRef` metadata (duration, size, mime, dimensions) | optimization | No Phase A command needs them; `MediaRef` stays `provider` + `locator`. |
| `--content` sugar on `lesson create` | nice-to-have | A lesson body is built explicitly via `lesson block add`; seeding an implicit first block was considered and declined to keep metadata and body cleanly separated. |
