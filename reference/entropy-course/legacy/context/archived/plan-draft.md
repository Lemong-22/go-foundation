# Plan Draft — Entropy Learning Engine

Derived from `product-requirements.md`. High-level, phased. No wall-clock deadlines — phases are gated by exit criteria, not dates. Owner paces based on capacity, optimizing for quality.

---

## 0. Snapshot

- **Vision:** A comprehensive, cohesive learning engine that teaches concepts via progressive disclosure, bite-sized explainers, and AI-generated / AI-evaluated sandboxed assignments. Coding first, general STEM later.
- **MVP shape:** One coding course — _Intro to JavaScript / TypeScript_ — end-to-end. Vertical slice that exercises every core loop (explain → quiz → assign → evaluate).
- **Primary learner (MVP):** Self-directed adult learners. Optimize for depth, flexibility, self-pacing.
- **Form factor:** Web app, Next.js + TypeScript.
- **Sandbox strategy:** Browser-only in MVP (JS/TS runs natively in a Web Worker). Server sandbox introduced later for Python/compiled languages.
- **AI strategy:** Multi-provider abstraction from day one (thin interface, not a framework).
- **Deferred from MVP:** server sandbox, analytics, mindmap/notes, three.js/video explainers, system-design course.

---

## 1. Guiding Principles (anchored in the PRD)

- **Progressive disclosure** is the core differentiator. Every concept has N paginated steps where step 1/N is the embryonic form and step k causally follows step k-1. The plan must not dilute this.
- **Bite-sized explainers** using HTML + sprite animation as the MVP format. Other formats (three.js, AI-gen video) are phase-gated behind real authoring throughput.
- **AI-generated, AI-evaluated assignments** in a sandbox. For JS/TS, "compile/lint" on submit is feasible in-browser.
- **Quiz rotation** for memorization, backed by a lightweight spaced-repetition state per learner.
- **Notes / mindmap / learning analytics** are valuable but not MVP — they depend on having real lesson content and real usage data first.

---

## 2. MVP Scope: Intro to JS/TS, End-to-End

### In scope

- One course, ~8–12 modules, each module has 3–6 lessons, each lesson has N=5–10 progressive-disclosure steps.
- Lesson renderer with the 1/N → N/N navigator (core UX).
- HTML + sprite-animation explainer units embedded per step.
- Quiz engine with rotation (spaced repetition state per learner/concept).
- Browser sandbox for JS/TS assignments. Web Worker isolation, TS transpilation in-browser, lint + type-check, AI evaluation on submit.
- AI provider abstraction with at least one concrete adapter.
- Minimal auth + persistence (learner account, progress, quiz state, submission history).

### Explicit non-goals for MVP

- No Python / compiled-language sandbox.
- No notes, no mindmap.
- No analytics dashboards (event logging only, stored for later).
- No teacher/admin surfaces.
- No payments / monetization.
- No mobile-specific UX (responsive, but not native).

### MVP exit criteria

- A new user can sign up, complete 3 consecutive lessons including quizzes and at least one graded assignment, and resume the next day where they left off.
- AI evaluation agrees with a human grader on ≥90% of a 50-submission eval set.
- Authoring one new lesson (progressive-disclosure steps + explainer + quiz + assignment) takes ≤ 1 focused day.

---

## 3. Phased Roadmap

### Phase 0 — Foundations & Spikes

Prove the risky bits before committing architecture.

- Prototype the progressive-disclosure step navigator with hand-authored content (no AI, no DB). Validate the UX of 1/N → N/N feels right.
- Spike on browser sandbox: Web Worker + TS transpilation + tiny runner API. Measure cold-start and submission latency.
- Spike on AI eval quality: give Claude / GPT a pool of 30 real learner-style submissions with a rubric, measure agreement with your own grading.
- Spike on HTML + sprite-animation authoring throughput using the existing `animated-plan-explainer` skill as a starting point.

**Exit criteria:** You believe (a) progressive disclosure reads well, (b) in-browser sandbox is fast enough, (c) AI eval is trustworthy with a rubric, (d) authoring a lesson isn't an unreasonable slog.

### Phase 1 — MVP: Intro to JS/TS Vertical Slice

Build the narrow-but-complete product described in §2.

- Data model: Course → Module → Lesson → Step; Quiz, Assignment, Submission, LearnerState.
- Lesson renderer + step navigator (polish).
- Quiz engine + rotation state.
- Sandbox runtime + AI evaluator + submission flow.
- AI provider abstraction: interface + one adapter (Anthropic or OpenAI, whichever wins the Phase 0 eval spike).
- Auth, basic profile, progress persistence.
- Course authoring: for MVP, content lives in a typed content repo (MDX + JSON) checked into the codebase. No CMS yet.
- Event logging (fire-and-forget writes to a log table). No dashboards.

**Exit criteria:** the §2 MVP exit criteria.

### Phase 2 — v1: Learning Data & Authoring Leverage

Now that there's a real course and real users, invest in the engine.

- Analytics pipeline: per-concept mastery score, weak-spot detection, learner-facing progress UI.
- Quiz rotation becomes mastery-driven, not just spaced.
- Notes (editable, per-lesson) + auto-generated mindmap view (read-only).
- Second AI provider adapter; start routing cheap tasks (quiz generation) vs. expensive tasks (eval).
- Authoring tooling: a small in-repo CLI / editor to scaffold lessons and preview them end-to-end.

**Exit criteria:** mastery score meaningfully predicts quiz performance; a second author (or you on a bad day) can produce a lesson from scratch in half a day.

### Phase 3 — v2: Sandbox & Language Breadth

Unlock content beyond JS/TS.

- Server sandbox: containerized execution (self-hosted Docker pool or managed — Judge0 / E2B). Plug behind the existing sandbox interface so lessons don't care.
- Second course: Intro to Python (natural fit for the server sandbox; broad audience).
- Third course seed: Data Structures & Algorithms (great progressive-disclosure material, auto-gradable via test cases).
- Lint/compile feedback surfaced in the learner UI before AI eval kicks in.

**Exit criteria:** two courses live; server sandbox handles at least Python and can be extended to a new language in under a week.

### Phase 4 — v3: Breadth & Richer Formats

General STEM and richer explainer formats — only after the engine has earned it.

- System-design course (diagram-first, AI evaluates trade-offs, no sandbox needed).
- Three.js explainers for spatial / system concepts.
- AI-generated short-video explainers as a supplementary format.
- Non-coding STEM (math, physics) — evaluates the generality of the engine.

**Exit criteria:** an explainer format and an evaluation modality exist for at least one non-coding domain end-to-end.

---

## 4. Architecture Sketch (high-level)

- **Framework:** Next.js (App Router) + TypeScript, deployed on Vercel. Server actions for mutations. Postgres (Neon or Supabase) for persistence.
- **Content format:** MDX for prose + structured JSON for step metadata, quiz items, assignment specs, animation sprites. Content lives in the repo in MVP, migrates to a CMS later only if needed.
- **Sandbox (MVP):** Web Worker running transpiled TS, with a message-passing runner API. AI eval receives the code, stdout/stderr, and a rubric.
- **Sandbox (v2+):** interface unchanged; a new adapter spins up a container per submission.
- **AI layer:** a thin `LLM` interface with `generate`, `evaluate`, and `stream` methods. One adapter per provider. Task-level router added in Phase 2.
- **Auth:** Clerk or Auth.js — whichever has less friction for the MVP.
- **Event log:** append-only table written from server actions. Used later by analytics.

---

## 5. Key Risks & Open Questions

- **AI eval reliability.** If Phase 0 shows <90% agreement even with a rubric, the core assignment loop is compromised. Fallback: narrow assignments to test-case-gradable forms and let AI only provide commentary.
- **Authoring throughput.** Progressive disclosure is expensive to author well. If it's slow, the course backlog becomes the bottleneck, not the engine. Mitigation: invest in authoring tooling early in Phase 2, possibly use AI to draft step sequences that the author edits.
- **Progressive-disclosure UX.** There's a real risk that "step 1/10 is embryonic" feels either patronizing or disconnected. Phase 0 UX spike must answer this.
- **Multi-provider cost drift.** Easy to over-engineer the abstraction. Keep it to an interface + adapters; defer routing until Phase 2.
- **Content-in-repo vs CMS.** Repo-as-CMS is fine for MVP but becomes painful once non-engineers author. Re-evaluate at Phase 2.

---

## 6. Immediate Next Actions

1. Draft the Intro to JS/TS course outline — modules, lessons, and the N-step plan for the first 2–3 lessons — so Phase 0 spikes have real content to render.
2. Run the Phase 0 spikes in parallel: disclosure UX, browser sandbox, AI-eval quality, explainer authoring throughput.
3. Based on spike results, commit to a concrete Phase 1 architecture and start the MVP build.

---

## 7. Open Decisions to Revisit

- **Auth provider:** Clerk vs. Auth.js.
- **DB host:** Neon vs. Supabase vs. self-hosted Postgres.
- **Primary AI provider for MVP eval:** decided by Phase 0 eval spike.
- **Whether to expose a "draft / preview" mode for lesson authoring inside the app** before a real CMS.
