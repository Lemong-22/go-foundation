# Technical Spec — Course Context · Phase E: Import & AI Consolidation

Status: Draft · Architecture: simplified Hexagonal + Clean + DDD (no CQRS)
Extends: `docs/spec.md`, `docs/phase-a-spec.md`, `docs/phase-b-spec.md`, `docs/phase-c-spec.md`, `docs/phase-d-spec.md` · Roadmap: `docs/roadmap.md` Phase E · Date: 2026-05-26

## 0. Overview

- **Purpose:** add a deterministic `import` CLI capability that ingests a
  zipped course directory in a defined format, computes an `ImportPlan`
  (creates / updates / no-ops / unresolved conflicts) against current DB
  state, and applies it — plus the integration seam through which an AI
  agent normalizes messy source material *into* the format and resolves
  conflicts *out of* the plan.
- **Language / stack:** Go 1.22, PostgreSQL via `pgx`, `cobra` + `viper` —
  unchanged.
- **In scope:**
  - The defined zip import format (outlined here in §1; full reference
    lives in companion doc `docs/import-format.md`).
  - The `ImportService` family of usecases (`PlanImport`, `ApplyPlan`) and
    their value types (`ImportPlan`, `ImportOperation`, `ImportConflict`,
    `ConflictResolution`).
  - New outbound port: `ImportSource` (reads a zip into a parsed structure).
  - Two new CLI commands (`import plan`, `import apply`).
  - REST endpoints for plan + apply (multipart file upload).
  - The agent-integration seam, documented as an `AGENTS.md` section.

- **Out of scope:**
  - **The AI agent itself.** Phase E specs the *interface* the agent
    consumes (the import format + plan JSON shape + CLI invocation
    contract); it does not spec prompts, model selection, or agent
    runtime. The agent is an external driver.
  - **Auth / permissions** — still single-instructor mode per
    `docs/phase-a-spec.md`.
  - **Validating that imported source code compiles or runs** — only
    structural validation of the import format itself.
  - **Import history / undo / rollback** — deferred (§6); its own
    subsystem (history table + reverse operations).
  - **Multi-course-per-zip** — deferred (§6); one zip = one course.
  - **Domain slugs on Quiz / Practice / Test** — slugs are an
    import-format-only concept for cross-referencing *within the zip*.
    The domain stays slug-free for those aggregates.
  - **A new database table** — Phase E adds no persistent state (no
    migration `000006`). All state already lives in the existing
    Course/Lesson/Quiz/Practice/Test tables.

**Architectural notes.**

- Phase E is the first slice that **doesn't introduce a new aggregate**.
  It's a workflow — a pair of usecases that compose every prior service.
  The hexagonal-spec-designer's §0–6 structure still works; §2 is just
  lighter (plan/operation value types only, no aggregates) and §4 is
  heavier (orchestration logic).
- The `ImportService` lives **inside** the existing `course` bounded
  context (per the grill-me: "one course bounded context, more
  aggregates"). It is *not* a new context.
- `ImportService` is the **first usecase that composes other services**
  rather than only repositories. For querying existing state it injects
  the existing repositories directly (consistent with how `DeleteCourse`
  uses two repositories today). For applying changes it calls the
  existing **services** so it reuses their validation, ID generation, and
  timestamp logic — no duplicate write paths.

---

## 1. CLI Commands

### Import commands (2)

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `import plan` | Parse zip, compute plan vs current DB, emit it. Does not write. | `<zip-path>`, `-o table\|json` (default `json`), `--output <file>` (opt) | plan rendered as JSON / table — every operation (`create` / `update` / `noop`) plus every unresolved `conflict` with recommended resolutions | zip not found, invalid layout, unsupported `format_version`, parse error, instructor id missing |
| `import apply` | Parse zip, compute plan, apply it. | `<zip-path>`, `--resolved-plan <file>` (opt), `--conflict-strategy fail\|skip\|update` (opt, default `fail`), `--force` (opt — skip confirmation) | per-aggregate summary (count by op kind; list of any failed aggregates) | unresolved conflicts when strategy is `fail`, plan-step failure, content-hash mismatch when `--resolved-plan` references a different zip |

**`--conflict-strategy`** is the batch fallback when no `--resolved-plan`
is supplied:

- `fail` (default) — abort if any unresolved conflict exists.
- `skip` — treat every unresolved conflict as `skip`.
- `update` — treat every unresolved conflict as `update` (overwrite the
  existing entity).

A `--resolved-plan` overrides the strategy on a *per-conflict* basis: the
agent rewrites each `ImportConflict` into a concrete `ImportOperation`
with the chosen `kind`. `--resolved-plan` must reference the same zip —
the apply step re-parses the zip and verifies the plan's `ZipHash` matches
so the agent cannot smuggle different content past a reviewed plan.

### The agent loop these commands enable

```text
1. agent normalizes arbitrary source material → produces a well-formed zip
2. course-cli import plan course.zip --output plan.json
3. agent reads plan.json; for each conflict, writes a resolution
   (update | skip | new) → resolved-plan.json
4. course-cli import apply course.zip --resolved-plan resolved-plan.json
```

### The defined zip format (outline)

The full reference is companion doc `docs/import-format.md`; here is what
Phase E commits to.

```text
course.zip
├── format_version.txt        # e.g. "1"
├── course.yaml               # course metadata (one per zip)
├── lessons/
│   ├── 01-foundations.md     # lesson; YAML frontmatter + blocks
│   └── 02-syntax.md
├── quizzes/
│   ├── foundations-quiz.yaml # slug-named quiz definition
│   └── ...
├── practices/
│   ├── fizzbuzz.yaml         # slug-named practice definition
│   └── ...
└── tests/
    └── midterm.yaml          # slug-named test definition
```

- **`format_version.txt`** is the *only* file outside the directory
  layout that's load-bearing. The planner reads it first and refuses
  unsupported versions.
- **`course.yaml`** holds course metadata (`title`, `slug`,
  `description`, `status`). `instructor_id` is *not* in the zip — it's
  resolved from CLI config (the existing Phase A seam).
- **Lessons are markdown files** with YAML frontmatter listing ordered
  blocks. Each block carries its `kind` and the kind-specific fields
  (markdown for text; provider + locator + caption for video; `quiz_ref`
  / `practice_ref` slugs for embeds).
- **Quizzes, practices, tests are YAML files** with a top-level `slug`
  field that lessons reference. The planner resolves slug-refs to UUIDs
  at apply time.

**One zip = one course.** Multi-course imports are deferred (§6).

---

## 2. Domain Model

Phase E adds **no aggregates and no entities** — it adds value types
that describe a one-shot import workflow. All types here are immutable,
identity-free value objects, declared in
`internal/course/domain/import_plan.go` (or a parallel package within
the bounded context).

### Value objects

**ImportPlan** — the planner's output; what `import plan` emits and
what `import apply --resolved-plan` consumes.
- Fields:
  - `FormatVersion string` — copied from the zip's `format_version.txt`
  - `ZipHash string` — SHA-256 of the canonicalized zip bytes (lets the
    apply step verify a resolved plan still corresponds to the zip)
  - `GeneratedAt time.Time`
  - `Operations []ImportOperation` — every entity in the zip becomes
    one operation (including `noop` for exact matches, so the plan is
    fully self-describing)
  - `Conflicts []ImportConflict` — unresolved decisions the agent
    must turn into operations
- Invariant: `Conflicts` is empty in a *resolved* plan; the apply step
  rejects a non-empty `Conflicts` list when `--conflict-strategy` is
  `fail`.

**ImportOperation** — one planned action against one imported entity.
- Fields:
  - `Kind OperationKind` — `create` | `update` | `noop` | `skip`
  - `EntityType EntityType` — `course` | `lesson` | `block` | `quiz`
    | `question` | `practice` | `test_case` | `test` | `test_item`
  - `EntityRef string` — import-local identifier (`course.slug`,
    `lesson:<title>`, `quiz:<slug>`, `block:<lesson-title>:<position>`,
    etc.) — human-readable; the agent uses it to find the corresponding
    payload in the zip
  - `TargetID *string` — existing entity's UUID for `update` / `noop` /
    `skip`; nil for `create`
  - `Payload json.RawMessage` — entity-type-specific payload (the parsed
    fields from the zip, ready to feed into a service Create/Update DTO)
- Invariants:
  - `Kind = create` ⇒ `TargetID == nil`.
  - `Kind = update` | `noop` | `skip` ⇒ `TargetID != nil`.

**ImportConflict** — an ambiguity the planner cannot resolve
deterministically.
- Fields:
  - `EntityType EntityType`
  - `EntityRef string`
  - `Reason ConflictReason` — `slug_collision` (Course) |
    `title_in_parent_collision` (Lesson/Quiz/Practice/Test) |
    `position_collision` (contained entities)
  - `Candidates []ConflictCandidate` — existing entities that matched
    the identity-ish field
  - `Recommended OperationKind` — the planner's suggested resolution
    (always `update` for content-differs matches; the agent may override)
  - `Payload json.RawMessage` — same shape as `ImportOperation.Payload`,
    so the agent can promote a conflict into an operation by copying the
    payload and choosing a `Kind`

**ConflictCandidate** — one existing entity that matched the imported
one by its identity-ish field.
- Fields:
  - `ID string`
  - `Description string` — human-readable summary (e.g. `"course
    a1b2c3 'Intro to Go' (created 2026-03-01)"`)

**ConflictStrategy** — the batch fallback used by `import apply` when no
`--resolved-plan` is supplied; wraps `string`; invariant: one of
`fail` | `skip` | `update`.

**OperationKind**, **EntityType**, **ConflictReason** — string-wrapped
enums, each with a closed set of constructors and predicates, following
the same VO discipline as `ContentBlockKind` / `Language` /
`ChoiceQuestionType` from earlier phases.

**ApplyResult** — `ApplyPlan`'s output.
- Fields:
  - `Applied []AppliedOperation` — operations that succeeded
  - `Failed []FailedOperation` — operations that errored, with the
    domain error attached
  - `Skipped []ImportOperation` — operations whose `Kind` was `skip` /
    `noop`
  - `AggregatesSucceeded int`, `AggregatesFailed int`

### Parsed source types (transport)

These are internal to the `ImportSource` adapter — the in-memory tree
the planner reads from. They are *not* part of the persistent domain
and are not described in detail here. Shape:

```go
type ParsedImportSource struct {
    FormatVersion string
    Course        ParsedCourse           // exactly one
    Lessons       []ParsedLesson         // ordered
    Quizzes       []ParsedQuiz           // slug-keyed
    Practices     []ParsedPractice
    Tests         []ParsedTest
}
```

Each parsed type carries the fields needed to build the corresponding
service-input DTO from Phases A–D, plus the import-local slug used for
cross-references (`quizzes[*].slug`, etc.).

### Domain notes

- **No aggregates.** The import workflow has a beginning and an end; it
  produces and consumes plans but owns no long-lived state. Modeling it
  as an aggregate would add ceremony without payoff.
- **Plans are fully self-describing.** Even exact-match entities appear
  in the plan as `noop` operations. This makes a plan a complete diff
  artifact the agent can read end-to-end; a missing operation in the
  plan means the entity isn't in the zip at all.
- **Plan JSON is the agent's contract.** `ImportPlan` and its component
  VOs round-trip through JSON cleanly (no domain types deeper than
  strings, ints, timestamps, and `json.RawMessage`). The agent never
  sees a Go type; it sees a JSON document. That JSON shape is what
  `docs/import-format.md` documents under "Plan JSON schema."
- **`Payload` is opaque to the agent.** The agent's job on a conflict
  is to choose a `Kind`, not to edit the `Payload`. To change content,
  the agent must edit the zip and re-plan — this keeps the
  `--resolved-plan` content-hash check meaningful.

---

## 3. Ports

### Inbound ports (driving — called by an inbound adapter)

A new `ImportService` interface — one method per command:

```go
type ImportService interface {
    PlanImport(in PlanImportInput) (PlanImportOutput, error)
    ApplyPlan(in ApplyPlanInput) (ApplyPlanOutput, error)
}

type PlanImportInput struct {
    ZipPath      string // local path; for REST the adapter writes upload to a tmpfile first
    InstructorID string // resolved by the adapter from config (Phase A seam)
}
type PlanImportOutput struct {
    Plan ImportPlan
}

type ApplyPlanInput struct {
    ZipPath          string
    InstructorID     string
    ResolvedPlanJSON []byte            // optional; pre-resolved plan
    ConflictStrategy ConflictStrategy  // used only when ResolvedPlanJSON is nil
}
type ApplyPlanOutput struct {
    Result ApplyResult
}
```

The inbound port uses primitive / opaque types at its boundary — `[]byte`
for the resolved plan JSON, `string` for `ConflictStrategy` value
(unwrapped to the VO inside the usecase). Same DTO-of-primitives
discipline as every prior phase.

### Outbound ports (driven — implemented by adapters)

One new port; everything else Phase E uses already exists.

```go
type ImportSource interface {
    Open(zipPath string) (ParsedImportSource, ImportSourceMetadata, error)
}

type ImportSourceMetadata struct {
    ZipHash       string // SHA-256 of canonicalized bytes
    FormatVersion string
}
```

Existing ports `ImportService` collaborates with (no changes to their
signatures):

- **Repositories (for queries):** `CourseRepository`,
  `LessonRepository`, `QuizRepository`, `PracticeRepository`,
  `TestRepository`. Used to look up existing entities for matching.
  *(Direct repository injection here matches the existing precedent of
  `DeleteCourse` using two repositories — the cleanest path for
  read-side queries inside one bounded context.)*
- **Services (for apply):** `CourseService`, `LessonService`,
  `QuizService`, `PracticeService`, `TestService`. Used to perform
  create/update operations so validation, ID generation, and
  timestamping are reused from the existing usecases — no duplicate
  write paths.
- **`Clock`** — for `ImportPlan.GeneratedAt`.

No new method is needed on any existing repository or service. (One
small *desirable but optional* addition is documented in §6: a
`Slug` filter on `ListCoursesInput` would let course-by-slug lookup
avoid a full scan. Not required.)

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**ZipImportSource** *(new)* implements `ImportSource`
- Construction dependencies: none (filesystem only)
- `Open` — side effect: opens the zip file, walks the entries, parses
  `format_version.txt`, `course.yaml`, `lessons/*.md` (YAML
  frontmatter + body), `quizzes/*.yaml`, `practices/*.yaml`,
  `tests/*.yaml`; computes the canonicalized SHA-256 hash of the zip
  bytes; returns `ParsedImportSource` + `ImportSourceMetadata`. Errors
  on unknown `format_version`, malformed YAML, missing required
  fields, or duplicate slugs within the zip.

No other new outbound adapter. Persistence reuses every prior phase's
adapters as-is.

### Inbound adapters (no business logic — not skeletoned per-method)

- **CLI adapter** *(modified)* — extends the cobra tree with an
  `import` subcommand group containing two commands. Parses
  `--resolved-plan` from a file path, `--conflict-strategy` from a
  flag, formats `ImportPlan` (JSON via stdlib `encoding/json`, or
  table via the existing renderer's new "plan" formatter), runs
  `--force` confirmation on apply. Maps `ErrUnresolvedConflicts` to a
  stderr message listing the unresolved entity refs before exiting
  non-zero. No business logic.
- **REST adapter** *(modified — Phase E endpoints)*:

  | Method & path | Inbound port method |
  |---------------|---------------------|
  | `POST /v1/import/plan` (multipart: `zip` file) | `PlanImport` |
  | `POST /v1/import/apply` (multipart: `zip` file + optional `resolved_plan` JSON; query: `conflict_strategy`) | `ApplyPlan` |

  The REST adapter writes the upload to a tmpfile, calls the inbound
  port, then deletes the tmpfile. Errors map to standard codes:
  `400` (parse error / bad strategy / hash mismatch), `409`
  (unresolved conflicts with `strategy=fail`), `500` (apply
  partial-failure aggregated).
- **Console** — an import screen with a drag-and-drop zip uploader, a
  plan-review panel (renders each operation and conflict), a
  "resolve conflicts" inline editor that builds the resolved plan
  client-side, and an apply button. UI detail is not part of this
  domain spec.
- **Agent integration** — a new `AGENTS.md` section in the repo
  documents the four-step loop (normalize → plan → resolve → apply)
  and links to `docs/import-format.md` for the zip layout and to the
  plan JSON schema. The agent is an *external* driver; no Go code in
  this spec.

### Usecases (implement inbound ports, depend on outbound ports + services)

**PlanImport** — implements `ImportService.PlanImport`
- Depends on:
  - Outbound: `ImportSource`, `Clock`
  - Repositories (queries): `CourseRepository`, `LessonRepository`,
    `QuizRepository`, `PracticeRepository`, `TestRepository`
- Steps:
  1. `ImportSource.Open(in.ZipPath)` → `(parsed, meta)`; propagate
     parse errors
  2. validate `meta.FormatVersion` against the supported set; return
     `ErrUnsupportedFormatVersion` if absent
  3. build a working `plan := ImportPlan{ FormatVersion:
     meta.FormatVersion, ZipHash: meta.ZipHash, GeneratedAt:
     Clock.Now() }`
  4. **plan the Course**:
     a. look up existing course by slug
        (`CourseRepository.FindBySlug(parsed.Course.Slug)`)
     b. if absent → append `create` operation with the parsed payload
     c. if present and content matches → append `noop` operation
        carrying the existing id
     d. if present and content differs → append a `slug_collision`
        `ImportConflict` with `Recommended = update`; record the
        existing id as the only candidate
     e. note the **target course id** (the existing id for present
        cases; for `create`, a placeholder — see §5 of this section)
  5. **plan independent definitions** (quiz / practice / test) — each
     in the same shape:
     a. for each `ParsedQuiz` (likewise practice, test): look up
        existing in the course (`QuizRepository.FindByCourse(target
        course id)` then filter by title)
     b. if absent → `create` op
     c. if present and content matches → `noop` op (carry the existing
        id)
     d. if present and content differs → `title_in_parent_collision`
        conflict, recommended `update`
     e. build a **slug → resolved id map** for embed refs (the
        resolved id is the existing id for noop/update/conflict, or a
        deterministic placeholder for create — see §5)
  6. **plan the Lessons**:
     a. for each `ParsedLesson`: look up existing in the course by
        title; same noop / update / conflict logic as quiz
     b. for each block in the lesson:
        i. if `kind` is `quiz` or `practice`, resolve the local slug
           via the map from step 5; emit a validation error if the
           slug is unknown (the agent's job to fix)
        ii. compare against the existing lesson's block at the same
           position; emit `create` (new position), `noop` (exact
           match), `update` (position match, content differs ⇒
           recommended), or `position_collision` conflict
  7. **plan the Tests**:
     a. for each `ParsedTest`: look up existing in the course by
        title; noop / update / conflict
     b. for each item: by position within the test; same logic as
        blocks
  8. return `PlanImportOutput{ Plan: plan }`

**ApplyPlan** — implements `ImportService.ApplyPlan`
- Depends on:
  - Outbound: `ImportSource`, `Clock`
  - Repositories: as above (for any necessary re-queries)
  - **Services**: `CourseService`, `LessonService`, `QuizService`,
    `PracticeService`, `TestService` — used to perform creates and
    updates so all existing validation / id-generation / timestamping
    is reused
- Steps:
  1. `ImportSource.Open(in.ZipPath)` → `(parsed, meta)`
  2. compute the plan:
     a. if `in.ResolvedPlanJSON != nil`: unmarshal it; verify
        `plan.ZipHash == meta.ZipHash` (return
        `ErrPlanZipMismatch` otherwise); reject if any
        `Conflicts` remain (the resolved plan must be fully resolved)
     b. else: call `PlanImport` internally to compute a fresh plan;
        if `Conflicts` is non-empty and `in.ConflictStrategy ==
        fail` → return `ErrUnresolvedConflicts` carrying the
        conflict list; otherwise materialize each conflict into an
        operation according to the strategy
  3. **group operations by aggregate** — every operation rolls up to
     exactly one top-level aggregate (the course, or one of its
     lessons / quizzes / practices / tests):
     - course-level: the course operation
     - lesson aggregate: the lesson operation + its block operations
     - quiz aggregate: the quiz operation + its question operations
     - practice aggregate: the practice operation + its testcase
       operations
     - test aggregate: the test operation + its item operations
  4. **execute aggregate-by-aggregate**, each inside a single
     transaction at the service-call layer (i.e. one boundary the
     services already establish via their repositories):
     a. **Course**: depending on `Kind`, call
        `CourseService.CreateCourse` (passing `in.InstructorID`),
        `CourseService.UpdateCourse`, or skip; capture the new /
        existing course id; on failure, append to `Failed`,
        increment `AggregatesFailed`, and continue to the next
        aggregate
     b. **Quiz / Practice / Test** (in any order — they don't depend
        on each other): for each, call the corresponding service's
        create/update method, passing the parsed payload; capture
        the new id; populate the slug-to-resolved-id map
     c. **Lesson**: for each lesson, create or update the lesson
        first (`LessonService.CreateLesson` / `UpdateLesson`), then
        add/replace/remove blocks via
        `LessonService.AddLessonBlock` / `UpdateLessonBlock` /
        `RemoveLessonBlock`. For `quiz` / `practice` blocks, fill in
        the `QuizRef` / `PracticeRef` from the slug-to-resolved-id
        map populated in step (b).
  5. record each successful op in `Applied`, each failure in
     `Failed`, each `noop`/`skip` in `Skipped`
  6. return `ApplyPlanOutput{ Result: ... }`

**Plan placeholder ids.** For a `create` operation, the planner doesn't
yet know the new entity's UUID — the service generates it at apply time.
The plan therefore uses a deterministic *placeholder* in
`TargetID` (e.g. `"$slug:foundations-quiz"`). The apply step replaces
placeholders with real UUIDs as parent entities are created, before any
child operation that references them. This is the slug-to-resolved-id
map mentioned above.

**No new domain usecase** is added — `PlanImport` and `ApplyPlan` are
the only two. Existing usecases are *unchanged*; Phase E composes them
without modifying them.

---

## 5. Container & Wiring

The composition root grows by one outbound adapter and one new service —
the same incremental shape every prior phase used. The new service is
unusual only in that it injects many things: every repository (for
queries) and every existing service (for apply).

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
    testRepo     := NewPostgresTestRepository(pool)
    ids          := NewUUIDGenerator()
    clock        := NewSystemClock()
    importSource := NewZipImportSource() // NEW — pure filesystem reader

    // 2. existing usecases — unchanged signatures from Phase D
    courseSvc   := NewCourseServiceImpl(courseRepo, lessonRepo, quizRepo, practiceRepo, testRepo, ids, clock)
    lessonSvc   := NewLessonServiceImpl(courseRepo, lessonRepo, quizRepo, practiceRepo, ids, clock)
    quizSvc     := NewQuizServiceImpl(courseRepo, lessonRepo, quizRepo, ids, clock)
    practiceSvc := NewPracticeServiceImpl(courseRepo, lessonRepo, practiceRepo, ids, clock)
    testSvc     := NewTestServiceImpl(courseRepo, testRepo, ids, clock)

    // 3. import usecase — composes repositories (queries) + services (apply)
    importSvc := NewImportServiceImpl(
        importSource, clock,
        courseRepo, lessonRepo, quizRepo, practiceRepo, testRepo,
        courseSvc, lessonSvc, quizSvc, practiceSvc, testSvc,
    )

    // 4. bind usecases to the inbound ports
    return &CLI{
        Course:   courseSvc,
        Lesson:   lessonSvc,
        Quiz:     quizSvc,
        Practice: practiceSvc,
        Test:     testSvc,
        Import:   importSvc, // NEW
    }, nil
}
```

`main.go` mounts the same inbound adapters as before; the cobra CLI
gains the `import` subtree, the REST server gains the two import
endpoints, the console gains the import screen — each just consumes
`container.Import`.

Because every dependency is constructor-injected through interfaces,
swapping `ZipImportSource` for an in-memory fake (a test that builds a
`ParsedImportSource` by hand) is a one-line change. Likewise, swapping
the concrete services for their existing in-memory fakes (used in the
Phase A–D usecase tests) lets `ImportService` be unit-tested without a
real zip *and* without Postgres.

**Construction order matters here for the first time.** `ImportService`
must be constructed after every other service. The container respects
that ordering naturally; if it were ever reorganized into modular
init blocks, this dependency would have to be honored.

---

## 6. Deferred

Judgment-call items considered during this design pass and deliberately
not built now. Parked in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| Import history (`import history` command + history table) | nice-to-have | Its own subsystem (audit table, per-import metadata, retention policy). Defer until there is a real operational need to audit who imported what when. |
| Import rollback (`import rollback <import-id>` command) | anticipatory | Requires history *and* reverse-operation generation. Per-aggregate transactions already limit blast radius of failures during apply; for "I imported something I didn't want," manual cleanup via existing delete commands is acceptable for MVP. |
| Multi-course-per-zip | anticipatory | One-zip-one-course is simpler and matches the agent's natural unit of work. Multi-course imports can be expressed as multiple zips or by adding a `courses/` directory layer in a future format version. |
| Fuzzy / near-duplicate title matching (Levenshtein, normalization) | optimization | Phase E uses exact title-within-parent matching only. Fuzzy matching helps surface typo-induced conflicts but introduces fuzziness the agent has to reason about. Revisit if false-negative conflicts (real duplicates the planner missed) become a complaint. |
| Content-hash matching for contained entities (Block / Question / TestCase / TestItem) | optimization | Position-based matching for contained entities is brittle to insertions. Hash-based identity would survive reordering but adds a stable-hash computation per content type. Add when the position-collision conflict rate is high enough to justify the complexity. |
| `import validate` command (parse-only, no DB compare) | nice-to-have | `import plan` already surfaces parse errors before any DB lookup. Add a dedicated `validate` if CI workflows want a no-DB-required entry point. |
| `Slug` filter on `ListCoursesInput` | optimization | Course-by-slug lookup in the planner currently goes through `CourseRepository.FindBySlug` (the inbound port doesn't expose it). Adding a slug filter to the public `ListCourses` would let other callers do the same without touching the repository. Tiny amendment; add when another caller needs it. |
| Domain slugs on Quiz / Practice / Test | anticipatory | Slugs live only in the import format for cross-referencing within a zip. Adding domain-level slugs would let the import skip title-based matching entirely, but introduces a new VO and a new uniqueness constraint per aggregate. Revisit if learners need shareable links by slug. |
| Updating embedded MediaRefs to a hosted-upload pipeline | out of context | Imports pass `MediaRef` values through unchanged. Hosting / transcoding belongs to the eventual media-upload phase (where the `MediaStore` outbound port also lives — see `docs/phase-a-spec.md` §6). |
| Schema-version migration in the planner | anticipatory | `format_version.txt` lets the planner reject unsupported versions; auto-migrating an older import to the current schema is a future capability when format breaks happen. |
| Two-phase apply with explicit commit | anticipatory | Per-aggregate transactions are the current commit boundary. A "stage everything, then commit" mode could be added if a future workflow demands strict all-or-nothing for a multi-aggregate import. |
| Concurrent imports / locking | anticipatory | The single-instructor model means concurrent imports are not a current concern. When multiple authors import in parallel, a course-level advisory lock during apply would prevent races. |
