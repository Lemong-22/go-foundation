# Roadmap — From Course/Lesson Persistence to a Rich Authoring Platform

Status: Draft · Author: Stephen Antoni · Date: 2026-05-26
Companion to: `docs/spec.md` (v Draft) · `docs/course-cli-prd.md` (v0.1.0)

## 0. Where we are

The `course` bounded context can persist **Course** and **Lesson** aggregates
through a clean hexagonal core: 13 CLI commands → inbound ports → usecases →
outbound ports → Postgres. A loopback "playground" runs the same Cobra handlers
through a form UI. Auth, REST, media, and any learner-facing concern are
explicitly deferred in `docs/spec.md` §6.

A `Lesson` today carries exactly one content field — `content TEXT`. The goal of
the next arc of work is to make the platform able to model and author a *rich*
course: lessons composed of text **and** video, plus quizzes, hands-on coding
practice, and a course-level test with a downloadable solution package.

This document sequences that work. It is the output of a structured design
interview; §1 records the decisions it rests on so future readers know *why*
the phases are shaped this way.

## 1. Decisions this roadmap rests on

| # | Decision | Choice |
|---|----------|--------|
| 1 | Phase focus | **Authoring-first.** The frontend is an instructor/author console. All learner-facing concerns (enrollment, progress, code execution, grading, gamification) are a later phase. |
| 2 | Content source of truth | **Postgres**, authored via CLI/REST. The legacy MDX-in-Git model is not adopted. Plus: support importing zipped course/lesson directories, reconciled into the DB by an AI agent. |
| 3 | Tooling roles | **CLI = machine/agent interface** (and gains `import`). **REST = human/frontend interface.** Both are adapters over the same usecases. |
| 4 | Content model shape | **Lesson → ordered typed `ContentBlock`s.** Text and video are inline blocks; quiz/practice blocks reference standalone aggregates by id. **Test** attaches at course level. |
| 5 | Binary media | **Reference-only.** The domain stores a `MediaRef` value object (provider + URL/asset-id). In Phase A this is a pure VO invariant — no outbound port. The `MediaStore` outbound port and the real upload/transcode pipeline arrive together in a later phase. (Refined in `docs/phase-a-spec.md` §6.) |
| 6 | Practice | **Auto-graded coding exercise:** prompt + starter code + hidden test cases + reference solution, per language. "Gamified" mechanics (XP, streaks, badges) are learner-side, deferred. |
| 7 | Quiz | **Choice-based only** (single / multiple choice) with per-question explanations and a pass threshold. Deterministically auto-gradable. Short-answer deferred. |
| 8 | Test | **Own course-level aggregate.** Owns ordered `TestItem`s built from the *same shared value types* as Quiz/Practice questions — authored for the test, not references into Quiz/Practice. Carries a `TestSolution` VO (zip `MediaRef` + explanation-video `MediaRef`), a timed flag, and a pass threshold. |
| 9 | Context layout | **One `course` bounded context**, extended with the three new aggregates. A new context appears only with the learner phase. |
| 10 | Import / agent split | CLI ships a deterministic `import` that parses a **defined zip layout** into a structured plan (`--dry-run`/`plan` mode). The **AI agent** normalizes messy input into that format and resolves DB conflicts (merge/update/skip), driving the CLI. |
| 11 | REST scope & auth | **Full inbound-port parity** with the CLI. Minimal auth now — a static API token / single-instructor mode, keeping the `instructor_id` seam. Real multi-user auth is deferred. |
| 12 | Console stack | **Go-served:** server-rendered `html/template` + htmx, evolving today's `playground.html`. Single binary, no JS toolchain. |
| 13 | Sequencing | **Foundation first, then vertical slices per content type.** Each content type goes domain → CLI → REST → console before the next begins. |

## 2. Target architecture

```text
                  ┌──────────────┐   ┌──────────────┐
   AI agent ─────► │   CLI adapter│   │ REST adapter │ ◄──── Author console
   (import)        │  (+ import)  │   │ (+ token auth)│       (htmx + templates)
                  └──────┬───────┘   └──────┬───────┘
                         │                  │
                         ▼                  ▼
              ┌────────────────────────────────────────┐
              │  Inbound ports (CourseService, Lesson-  │
              │  Service, QuizService, PracticeService, │
              │  TestService) — shared DTOs             │
              └───────────────────┬────────────────────┘
                                  ▼
              ┌────────────────────────────────────────┐
              │  Usecases  →  Domain  (course context:  │
              │  Course, Lesson+ContentBlock, Quiz,     │
              │  Practice, Test; MediaRef VO)           │
              └───────────────────┬────────────────────┘
                                  ▼
              ┌────────────────────────────────────────┐
              │  Outbound ports → Postgres repos,       │
              │  IDGenerator, Clock  (+ MediaStore once  │
              │  the media-upload phase lands)          │
              └────────────────────────────────────────┘
```

The dependency rule from `spec.md` §3 is unchanged: every adapter wraps the same
usecases; ports are declared in the core and speak only domain types.

## 3. Phased roadmap

Phases are ordered by dependency, not by date — consistent with the curriculum's
no-fixed-dates philosophy. Each phase is independently shippable.

### Phase A — Domain foundation, REST scaffold, console shell

**Goal:** restructure `Lesson` into the block model and stand up the two new
adapters, with no new content type yet. End state: today's exact functionality,
but lessons are block-based and reachable via CLI **and** REST **and** console.

Domain & core
- Introduce `ContentBlock` as an entity *within* the `Lesson` aggregate: an
  ordered list of typed blocks. Block kinds for this phase: `text` (markdown
  string) and `video` (a `MediaRef` + caption). Design the extension seam so
  `quiz` and `practice` reference-blocks slot in during Phases B–C without
  reshaping the aggregate.
- Add the `MediaRef` value object (provider + URL/asset-id) — a pure VO whose
  constructor validates the locator format. No `MediaStore` outbound port in
  this phase: it has no side effect to perform yet. The port is introduced with
  the media-upload pipeline (see Deferred / §6 of `docs/phase-a-spec.md`).
- Note on shared value objects: the choice-question and coding-task VOs that
  Quiz/Practice/Test will share are *designed* here as part of the block seam,
  but materialize with their first consumer (Quiz, Phase B) so no VO ships
  without a caller — Test then reuses them in Phase D.

Persistence
- Migration `000002`: `content_blocks` table (`id`, `lesson_id`, `kind`,
  `position`, text/media columns). Backfill: each existing `lessons.content`
  becomes one `text` block. Keep the `content` column through the transition;
  drop it in a later migration once nothing reads it.

REST adapter (the "promote CLI to REST" workstream begins here)
- `net/http` server with a static-token auth middleware (`COURSE_CLI_API_TOKEN`
  or config). JSON request/response reusing the existing inbound-port DTOs.
- Endpoints at parity with the existing course + lesson CLI commands.

Console shell
- Evolve `playground.html` into a real `html/template` + htmx app: course
  list/create/edit and a lesson editor that renders the block model (add,
  reorder, edit text and video blocks). Reuse the existing `lesson reorder`
  usecase for block/lesson ordering.

**Exit criteria:** all 13 original commands still pass; a lesson with mixed
text + video blocks can be authored from CLI, REST, and console; REST is
token-protected.

### Phase B — Quiz slice

**Goal:** quizzes authorable end-to-end and embeddable in a lesson.

- **Domain:** `Quiz` aggregate (`QuizID`, `courseID`, ordered `Question`s, pass
  threshold). `Question` = single/multiple choice with options, correct
  answer(s), and an optional explanation. This phase materializes the shared
  choice-question value type.
- **Persistence:** migration `000003` — `quizzes`, `quiz_questions` tables.
- **CLI:** `quiz create/list/get/update/delete`; question management
  (`quiz question add/update/remove/reorder`).
- **REST:** quiz endpoints at parity.
- **Console:** a quiz builder; a `quiz` block in the lesson editor that
  references a quiz by id.

**Exit criteria:** a quiz can be authored and embedded as a lesson block via all
three adapters.

### Phase C — Practice slice

**Goal:** coding exercises authorable end-to-end and embeddable in a lesson.

- **Domain:** `Practice` aggregate (`PracticeID`, `courseID`, language, prompt,
  starter code, reference solution, hidden test cases). Materializes the shared
  coding-task value type. The exercise *definition* only — no runner.
- **Persistence:** migration `000004` — `practices`, `practice_test_cases`.
- **CLI / REST / Console:** `practice` commands, endpoints, and builder UI; a
  `practice` block in the lesson editor.

**Exit criteria:** a coding exercise can be fully specified and embedded as a
lesson block. Executing/grading learner code is explicitly *not* in this phase.

### Phase D — Test slice

**Goal:** a course-level test authorable end-to-end.

- **Domain:** `Test` aggregate (`TestID`, `courseID`, ordered `TestItem`s, a
  `TestSolution` VO, timed flag, pass threshold). A `TestItem` is a choice
  question or a coding task, built from the shared value types from Phases B–C.
  `TestSolution` = a zip `MediaRef` + an explanation-video `MediaRef`.
- **Persistence:** migration `000005` — `tests`, `test_items`.
- **CLI / REST / Console:** `test` commands, endpoints, and builder UI;
  the test surfaces at course level, not as a lesson block.

**Exit criteria:** a course can carry a complete timed test with a solution
package, authored via all three adapters.

### Phase E — Import & AI consolidation

**Goal:** ingest zipped course material and reconcile it into the DB.

- **Define the import format** in `docs/import-format.md`: a documented zip
  layout (e.g. `course.yaml`, `lessons/*.md` with block frontmatter,
  `quizzes/*.yaml`, `practices/*.yaml`, `tests/*.yaml`, media references).
- **CLI `import`:** parse a zip deterministically into a structured *import
  plan*; a `--dry-run` / `plan` mode emits the plan as JSON — creates, updates,
  and conflicts (slug match, near-duplicate title) — without writing.
- **Agent integration:** document (an `.agents/` skill or `AGENTS.md` section)
  how an AI agent (1) normalizes arbitrary source material into the import
  format and (2) consumes the plan JSON and resolves conflicts by issuing
  follow-up CLI commands.

**Exit criteria:** a well-formed zip imports cleanly; a conflicting one produces
a reviewable plan an agent can act on.

### Phase F — Hardening & console polish

- Apply the `design.md` editorial design system to the console.
- Retire the old playground (or fold it into the console as a "commands" view).
- Refresh `docs/spec.md` to describe the expanded context; top up test coverage
  per the existing focused-test pattern.
- Write deployment notes for the single binary + Postgres.

## 4. Deferred — the Learner phase (named so it is not forgotten)

Out of scope for this roadmap, but the natural next arc. Listed so phase
boundaries above stay honest:

- A **new bounded context** for learner state (enrollment, progress, quiz
  attempts, test submissions) — different owner, different lifecycle.
- **Code-execution sandbox + auto-grading runner** for practice and test coding
  tasks. The single heaviest engineering item; its own project.
- **Gamification mechanics** — XP, streaks, badges, leaderboards.
- **Real multi-user auth** + an **Identity** bounded context, replacing the
  static API token. The `instructor_id` seam is already in place for this.
- **Media upload + transcode pipeline** — introduces the `MediaStore` outbound
  port and a real object-storage/Mux adapter (until then `MediaRef` is a
  reference-only value object with no port).
- The **learner-facing app** — the lesson reader described by `design/course.html`.
- **Short-answer quiz questions** (need exact-match heuristics or AI grading).

## 5. Cross-cutting principles

- **Hexagonal discipline holds.** Every new adapter (REST, console, import)
  wraps usecases through inbound ports; new outbound needs (`MediaStore`) are
  ports declared in the core.
- **DTO reuse.** REST and CLI share the inbound-port input/output DTOs — REST
  adds no parallel type set.
- **Vertical slices.** Per `spec.md`, each new command/endpoint is its own slice
  through the stack; no content type shares a usecase with another.
- **Migrations stay sequential** numbered files; `migrate` tooling is wired in
  `main.go`, outside the bounded context, as today.
- **Tests per slice:** in-memory fakes for usecase tests, Postgres integration
  tests for repositories, matching the current `*_test.go` layout.

## 6. Risks & open questions

- **Block-model migration is one-way in practice.** Keep the `lessons.content`
  column until Phase F to allow a rollback window.
- **htmx block editor.** Drag-reordering of blocks is the riskiest console
  interaction; mitigated by reusing the proven `reorder` usecase and keeping
  ordering server-authoritative.
- **Shared value objects vs coupling.** The choice-question and coding-task VOs
  are shared by Quiz/Practice and Test. Keep them as standalone value objects in
  the domain package — not owned by any one aggregate — to avoid a circular
  dependency.
- **Import format versioning.** The zip layout needs a `format_version` field
  from day one so the agent and CLI can detect mismatches.
- **Console vs learner app reuse.** The Go-served console will *not* directly
  become the learner app; that is a deliberate later rebuild. Confirm this is
  acceptable before Phase F polish effort goes in.
