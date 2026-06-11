# Roadmap — Learner App (Bun + Next.js) & Course Material

Status: Draft · Author: Stephen Antoni · Date: 2026-06-03
Companion to: `docs/roadmap.md` (backend arc, Phases A–F) · `docs/import-format.md`

## 0. Where we are

The `course` bounded context is complete through the backend arc of
`docs/roadmap.md`: block-based lessons, Quiz, Practice, and Test aggregates, the
zip import pipeline, and a hexagonal core reached by **two** adapters — the CLI
(agent/instructor) and the REST API. REST now has full parity: learner reads
(`GET /v1/courses`, `…/{id}`, `…/{id}/lessons`, `/v1/lessons/{id}`, block/quiz/
practice/test reads) and instructor writes (create/update/delete/publish,
lessons, reorder), behind a static bearer token.

This document covers the next arc: a **learner-facing web app** and the
**course material** to fill it.

## 1. Decisions this arc rests on

| # | Decision | Choice |
|---|----------|--------|
| 1 | **Supersedes roadmap decision #12.** | The frontend is **not** a Go-served `html/template` + htmx console. It is a **Bun + Next.js (App Router)** app. The old playground/console stays as-is for now (dev tool), not evolved. |
| 2 | First surface | **Learner app first** — the student lesson reader described by `design/course.html`. Instructor authoring stays on the CLI (and import pipeline) for this arc; a web authoring console is deferred. |
| 3 | Browser ↔ API | **Server-side token only.** The Next.js server (route handlers / server components / a thin BFF) holds `COURSE_CLI_API_TOKEN` and calls the Go REST API. The token never reaches the browser. The token is the **sole gate for now**; per-learner auth is a planned future feature, not this arc. |
| 4 | Material prep | **Import pipeline (zip + AI agent).** Authors write markdown/yaml in the `docs/import-format.md` layout; the agent normalizes and drives CLI/REST import. The app consumes the result read-only. |
| 5 | Answer safety | L1 renders quizzes/practices with **clearly-labeled dummy data** — no real answers are fetched — so L1 does not block on backend work. The UI must visibly mark this content as placeholder/dummy to the user. **Learner-safe read projections** (§4) replace the dummy data afterward. The reader must never receive real correct-answers, reference solutions, or hidden test cases. |
| 6 | Interactivity scope | This arc ships a **reader**, not a grader. Quizzes/practices/tests are *displayed*; attempts, code execution, grading, and progress are the deferred Learner phase. |
| 7 | Repo placement | **`web/` workspace in this repo.** Go backend and Next.js app live in one source tree so the REST contract co-evolves. |
| 8 | Wire casing | **PascalCase, unchanged.** The Next.js BFF consumes REST responses as-is (`ID`, `Title`, `Markdown`). No `json` tags added to core DTOs. |
| 9 | Caching | **On-demand revalidation.** Published content is revalidated on demand (e.g. when an instructor republishes), not on a timer — chosen for simplicity. |
| 10 | Source of truth | **The REST API is the contract.** `design/course.html` is a visual target only, not a data shape. |

## 2. Target shape

```text
   Browser (learner)
        │  HTML/RSC, no API token
        ▼
   ┌──────────────────────────┐     COURSE_CLI_API_TOKEN (server-only)
   │  Next.js app (Bun)        │ ───────────────┐
   │  - App Router pages       │                ▼
   │  - server components / BFF │        ┌───────────────┐
   │  - markdown + video render │        │  Go REST API   │ ◄── CLI / import
   └──────────────────────────┘        │  (+ learner    │     (authoring)
                                        │   projections) │
                                        └───────┬───────┘
                                                ▼
                                         usecases → domain → Postgres
```

The Go backend stays the source of truth; the Next.js app is a presentation
client over REST. No business logic moves into the frontend.

## 3. Phase L1 — Learner reader (Next.js + Bun)

**Goal:** a published course is browsable and readable end-to-end in the browser,
matching `design/course.html`.

Scaffold
- Bun-managed Next.js (App Router, TypeScript). Decide repo placement: a `web/`
  workspace in this repo (keeps one source tree, Go + JS side by side) vs. a
  separate repo. Default recommendation: **`web/` in-repo**, since the import
  format and REST contract are co-evolving.
- A typed API client (the BFF) wrapping the REST endpoints, reading the token
  from a **server-only** env var. All calls run in server components / route
  handlers, never client components.

Pages (driven by `design/course.html`)
- **Catalog** — `GET /v1/courses?status=published` → published courses only.
- **Course detail** — `GET /v1/courses/{id}` + `GET /v1/courses/{id}/lessons`
  (ordered lesson list / syllabus).
- **Lesson reader** — `GET /v1/lessons/{id}` + `GET /v1/lessons/{id}/blocks`,
  rendering ordered blocks:
  - `text` → markdown render (sanitized).
  - `video` → provider embed from the `MediaRef` (provider + locator + caption).
  - `quiz` / `practice` block refs → rendered from **clearly-labeled dummy
    data** in this slice (no real aggregate fetched), so L1 doesn't block on §4.
    The component must visibly mark the content as placeholder (e.g. a "sample
    data" badge) so it is never mistaken for real course content. §4 later swaps
    the dummy source for the learner-safe projection behind the same component.
- **Test view** (course-level) — same dummy-data treatment as quiz/practice in
  this slice; real read-only rendering arrives with §4.

Cross-cutting
- **Wire casing:** consume REST responses as PascalCase (`ID`, `Title`,
  `Markdown`) directly in the BFF — no `json` tags on core DTOs, no client-side
  remapping convention beyond typing the responses.
- **Revalidation:** on-demand. Cache published reads and invalidate on an
  explicit trigger (e.g. a republish hook / revalidation route), not a timer.
- Loading / empty / not-found / error states per the reader design.
- Visual system: apply `design.md` / `design/course.html` styling.

**Exit criteria:** a published course authored via CLI/import is fully readable
in the browser — catalog → course → lesson with text+video blocks, plus
quiz/practice/test rendered from clearly-labeled dummy data — with no API token
exposed client-side and no real answers fetched.

## 4. Phase L2 — Learner-safe read projections (backend)

**Goal:** close the answer-leak gap and replace L1's dummy quiz/practice/test
data with real, safe content. **Not a blocker for L1** — L1 ships first with
clearly-labeled dummy data; L2 swaps in the real source behind the same
components.

- New **learner-facing read DTOs/endpoints** (or a `?view=learner` mode) that
  strip: quiz correct answers + per-question explanations-before-submit,
  practice reference solutions + hidden test cases, and `TestSolution`.
- Keep these as adapter-level projections over the same usecases — no new domain
  logic. Mirror the existing handler/test pattern in
  `internal/course/adapter/rest`.
- Only published content is served on the learner path; draft content stays
  instructor-only.

**Exit criteria:** every learner endpoint provably omits answer/solution fields
(covered by tests); instructor endpoints retain full fidelity.

## 5. Phase C1 — Course material via import pipeline

**Goal:** real courses authored and loaded, reusing Phase E import.

- Author source material in the `docs/import-format.md` zip layout
  (`course.yaml`, `lessons/*.md` with block structure, `quizzes/*.yaml`,
  `practices/*.yaml`, `tests/*.yaml`). Start from what's in `courses/`
  (javascript, typescript, golang) and `legacy/curriculum/`.
- AI agent normalizes messy/legacy source into the format, runs `import` in
  plan/`--dry-run` mode, reviews the plan JSON, resolves conflicts, then applies.
- Publish finished courses (`POST /v1/courses/{id}/publish`) so they appear in
  the learner catalog.
- Treat the first full course (JavaScript) as the end-to-end proof: author →
  import → publish → read in the Next.js app.

**Exit criteria:** at least one complete, published course rendered in the
learner app, sourced entirely through the import pipeline.

## 6. Explicitly NOT in this arc (the Learner phase, deferred)

Carried forward from `docs/roadmap.md` §4 — listed so the boundaries stay honest:

- **Interactive assessment** — quiz submission/scoring, practice code execution
  + auto-grading, test attempts. Needs a code-execution sandbox (the heaviest
  single item) and a new **learner-state bounded context** (enrollment,
  progress, attempts, submissions).
- **Real multi-user auth + Identity context** — replaces the static token; the
  `instructor_id` seam is ready for it. Required before per-learner state or a
  public multi-tenant deployment.
- **Web instructor authoring console** — a Next.js authoring surface over the
  existing instructor-write endpoints. Deferred in favor of CLI/import this arc.
- **Media upload + transcode pipeline** — introduces the `MediaStore` outbound
  port; until then `MediaRef` stays reference-only (external video URLs/ids).

## 7. Risks & watch-items

Open questions from the prior draft are now resolved (see §1, rows 3, 5, 7–10).
What remains to watch during execution:

- **Answer leakage (highest).** Resolved approach: L1 uses clearly-labeled dummy
  data; real answers only arrive via L2's learner-safe projections. The risk to
  guard is a regression where a component is wired to a *full-fidelity* read
  endpoint by mistake. Enforce with a test on L2 that asserts answer/solution
  fields are absent, and never point a learner component at an instructor read.
- **Dummy data clarity.** The placeholder quiz/practice/test content must be
  unmistakably marked in the UI; a reviewer or learner should never read it as
  real course material. Make the "sample data" affordance explicit and obvious.
- **Token as sole gate.** The Next.js server is fully trusted and there is no
  learner identity. Acceptable for a public reader; revisit the moment any
  per-learner feature or draft preview appears (the planned future auth work).
- **On-demand revalidation wiring.** Stale reads are avoided only if the
  republish → revalidate trigger is actually wired; if it's missed, published
  edits won't surface. Verify the invalidation path end-to-end.
