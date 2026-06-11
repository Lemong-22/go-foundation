# PRD: Crashcourse — An Online Coding Crash Course Platform

**Status:** Draft v0.1
**Author:** Stephen
**Last updated:** 2026-05-08
**Working name:** Crashcourse (placeholder)

---

## TL;DR

Crashcourse is a self-serve, web-based learning platform that helps anyone — from absolute beginners to working engineers — get up to speed on a coding subject in the shortest possible path. Each lesson combines a short animated explainer video, an illustrated cheat sheet, a library of copy-pasteable code snippets, an inline quiz, and a hands-on coding practice/exam — all on a single page. The platform is designed as a multi-course catalog, with a single-course pilot validating the format before we scale.

---

## 1. Problem Statement

People who want to learn or refresh a coding subject quickly are forced to stitch together a poor experience: long, expensive bootcamps; dense official documentation; scattered YouTube videos; and bookmarked Stack Overflow answers. None of these deliver what most learners actually want — a fast, structured, _trustworthy_ path that teaches the essentials, lets them practice, and leaves them with something they can copy and reuse when they are coding for real.

This is felt most acutely by three groups: absolute beginners who don't know where to start, working developers who need to ramp on a new language or framework in days (not months), and students or job-seekers cramming concepts before an exam or interview. The cost of not solving it is real: learners drop out, bounce between resources, or default to AI chat for one-off answers without ever building a durable mental model.

---

## 2. Goals

The platform should achieve the following outcomes:

1. **Compress time-to-understanding.** A learner can grasp a single concept in under 5 minutes via the animated explainer + cheat sheet, and the full crash course for a subject in 2–6 hours of focused work.
2. **Drive completion, not just consumption.** At least 40% of learners who start a crash course finish it (vs. industry-typical <10% for MOOCs).
3. **Verify learning, not just exposure.** Every lesson ends with a quiz and a coding exercise; learners measurably improve from pre-test to post-test.
4. **Become a reference, not just a course.** Cheat sheets and copy-paste code snippets are valuable enough that learners return to them after finishing the course.
5. **Scale across subjects.** The lesson format is template-driven so we can add new courses (Python, SQL, React, Rust, system design, etc.) without bespoke design work per course.

---

## 3. Non-Goals

The following are explicitly **out of scope** for v1 (and for some, out of scope entirely):

- **Replacing full bootcamps or degree programs.** We compete on speed and clarity, not depth or accreditation.
- **Live instruction / 1:1 tutoring.** No scheduled classes, no human tutors. Asynchronous self-serve only.
- **Open-ended project review or code review.** Practice problems have automated test-based grading; we don't grade portfolios.
- **Job placement, hiring, or recruiter features.** Not our business.
- **Social network features (follows, feeds, DMs).** Comments on lessons may come later (P1); a full social layer will not.
- **Course authoring by end users in v1.** Initial courses are built in-house to keep quality bar high. Community authoring is P2.

---

## 4. Target Users

We serve three personas with one experience, differentiated by entry path and pacing — not by separate products.

**Persona A — "Anya, the Absolute Beginner."** Has never written code. Wants to know if coding is for her. Needs maximum scaffolding, gentle ramps, no jargon, and big visible wins early.

**Persona B — "Ben, the Stack Switcher."** 5+ years of experience in one language, needs to ship in a new one next month. Wants to skip "what is a variable" and jump to idioms, gotchas, and copy-paste-ready patterns. Will skim the explainer, scan the cheat sheet, and grind the practice problems.

**Persona C — "Carla, the Crammer."** CS student or job-seeker who has 1–2 weeks before an interview or exam. Needs concept review + repetition + timed practice. Will use quizzes and coding exams more than videos.

The platform must let each of them get value from the same lesson page without forcing them through irrelevant content.

---

## 5. User Stories

### Discovery & onboarding

- As a visitor, I want to browse courses by topic and difficulty so I can find a crash course that matches what I need to learn.
- As a new user, I want to take a 60-second placement check so the platform recommends a starting point appropriate for my level.
- As a returning user, I want to see my in-progress courses and a "resume where I left off" button so I don't lose momentum.

### Inside a lesson

- As Anya, I want to watch a short animated explainer of each new concept so I can build intuition before seeing code.
- As Ben, I want to skip the video and jump straight to the cheat sheet and code snippets so I can extract what I need fast.
- As any learner, I want a one-click **Copy** button on every code snippet so I can paste it into my own editor without retyping.
- As any learner, I want the cheat sheet to be illustrated (diagrams, icons, visual metaphors) so concepts stick, not just words.
- As Carla, I want to take a quiz at the end of each lesson and see which questions I got wrong so I can target my review.
- As any learner, I want to write and run code in the browser against hidden test cases so I get immediate feedback without setting up a local environment.

### Practice & exams

- As Carla, I want a timed end-of-course exam mixing quiz and coding questions so I can simulate interview/exam pressure.
- As any learner, I want to see my quiz and exam history so I can track improvement over time.
- As any learner, I want to retake a quiz or exam without losing my prior score so I can practice freely.

### Reference

- As Ben, I want to bookmark cheat sheets and snippets so I can come back to them while coding, even after I've finished the course.
- As any learner, I want a "cheat sheet only" view of an entire course so I can use it as a reference doc.

### Edge & error states

- As any learner, I want a clear error message if my code times out or hits a runtime error, with a hint about what to look at.
- As any learner, I want my progress saved automatically so I don't lose work if my browser crashes.
- As an offline user, I want to be told clearly that code execution requires a connection (rather than failing silently).

---

## 6. The Lesson Format (core product surface)

Every lesson page is composed of five modules — the same five for every course, every subject. This consistency is the product. A learner who has done one Crashcourse lesson knows how to use any Crashcourse lesson.

1. **Animated explainer (video).** 60–180 seconds. Visual, narrated, no talking heads. Plays automatically muted with captions on; learner clicks to unmute. Speed controls (1x / 1.5x / 2x). Skippable.
2. **Illustrated cheat sheet.** A static, scrollable visual reference for the concept — diagrams, side-by-side syntax tables, color-coded callouts. Designed to be screenshot-able and print-friendly. Acts as the "I just need the gist" entry point.
3. **Copy-paste code section.** A curated set of working code snippets demonstrating the concept's most common patterns. Each snippet has a one-click Copy button, a syntax-highlighted code block, an "edit & run" button (opens it in the practice runner), and a one-line explanation of when to reach for it.
4. **Quiz.** 3–8 questions per lesson. Mix of multiple-choice, multi-select, fill-in-the-blank, and "predict the output." Instant feedback per question, with an explanation that links back to the relevant cheat-sheet section.
5. **Coding practice.** 1–3 small problems per lesson, run against hidden test cases in the browser. Includes a "Show me a hint" button, a "Show me a solution" button (locked until the learner has tried), and a "Try a harder version" button for Persona C.

End-of-course **exam** is the same modules composed differently: a longer, timed mix of quiz items + coding problems covering the whole course, with a final score and a per-concept breakdown.

---

## 7. Requirements

### 7.1 Must-Have (P0) — required to ship the v1 platform

**Catalog & navigation**

- Course catalog page with subject, difficulty, and estimated time filters.
- Course landing page describing prerequisites, learning outcomes, and lesson list.
- Lesson navigation (previous/next, jump to module within lesson).
- Acceptance: a learner can find, start, and resume any course in ≤3 clicks from the homepage.

**Lesson page modules (all five)**

- Animated explainer video player with captions, speed control, and skip.
- Illustrated cheat sheet rendered as responsive HTML (not just an image — must be scannable on mobile and screen-readable).
- Copy-paste code section with one-click copy, syntax highlighting, and per-snippet "open in runner."
- Quiz engine supporting multiple-choice, multi-select, fill-in-the-blank, and predict-the-output question types, with per-question feedback.
- In-browser code runner with hidden test cases, run output, and pass/fail feedback.

**Code execution**

- One language supported at launch (recommendation: Python — broadest audience appeal across all three personas).
- Sandboxed execution with CPU + memory + wall-clock limits.
- Test cases run server-side (or in a hardened sandbox) so solutions can't be inspected from the client.

**Accounts & progress**

- Email + Google sign-in.
- Per-user progress tracking (lesson completion, quiz scores, exam scores).
- Resume-where-you-left-off on any course.

**End-of-course exam**

- Timed, mixes quiz items and coding problems.
- Per-concept score breakdown and link back to the relevant lesson.

**Operational essentials**

- Analytics on funnel: signup → course start → lesson complete → course complete.
- Basic admin to publish, unpublish, or update a course.
- Accessibility: WCAG 2.1 AA at minimum for the lesson page (captions, keyboard nav, screen-reader labels on the runner).

### 7.2 Nice-to-Have (P1) — fast follows after v1 ships

- **Multiple languages in the runner** (JavaScript, SQL, then on demand).
- **Skill-level paths.** Same course, three suggested routes: Beginner (full), Refresher (skim videos, do exercises), Interview (quizzes + coding only).
- **Bookmarks and a personal "snippet library"** of copied-and-saved snippets.
- **Streaks, daily-goal nudges, completion badges.**
- **Mobile-optimized practice runner** (currently desktop-first).
- **Comments / Q&A under each lesson**, lightly moderated.
- **Search across cheat sheets and snippets** ("show me every Python snippet about list comprehensions").
- **Export a course's cheat sheets to a single PDF.**
- **Spaced-repetition review** of quiz questions you got wrong.

### 7.3 Future Considerations (P2) — design for, don't build

- **AI tutor / hint companion.** Contextual hints, "explain this error," "explain why my code is wrong" — without giving away the answer.
- **Course authoring tools** for community contributors, with a quality review pipeline.
- **Live interview simulator** (timed, voice-narrated, single-attempt).
- **Certifications** that can be shared on LinkedIn.
- **Teams / enterprise** plan with seat management, custom course assignments, and progress dashboards.
- **Localization** beyond English (animations and cheat sheets are the hard part).
- **Offline mode** for cheat sheets and quizzes (code runner stays online-only).

These are P2 because we want architectural decisions in v1 (data model, lesson schema, content authoring pipeline) to keep them cheap to add later — not because we plan to build them soon.

---

## 8. Success Metrics

### Leading indicators (measured weekly)

| Metric                                       | Target               | Measurement                 |
| -------------------------------------------- | -------------------- | --------------------------- |
| Signup → course-start conversion             | ≥60%                 | Funnel analytics            |
| First-lesson completion (activation)         | ≥70% of starters     | Funnel analytics            |
| Median time to complete first lesson         | ≤25 min              | Event timestamps            |
| Quiz attempt rate per lesson                 | ≥85% of lesson views | Event analytics             |
| Coding-exercise attempt rate per lesson      | ≥60% of lesson views | Event analytics             |
| Snippet copy events per active user per week | ≥5                   | Click events on Copy button |

### Lagging indicators (measured monthly / quarterly)

| Metric                                     | Target                                                                 | Measurement                             |
| ------------------------------------------ | ---------------------------------------------------------------------- | --------------------------------------- |
| Course completion rate                     | ≥40% (industry MOOC benchmark <10%)                                    | Per-course funnel                       |
| Pre-test → post-test score lift            | ≥25 percentage points avg                                              | Optional placement quiz vs. course exam |
| 30-day return rate                         | ≥35%                                                                   | Cohort retention                        |
| Cheat-sheet revisit rate (post-completion) | ≥25% of completers return to a cheat sheet within 14 days of finishing | Page analytics                          |
| NPS                                        | ≥40                                                                    | In-product survey on completion         |

A "win" for v1 is hitting all leading-indicator targets within 60 days of launch, on at least one course.

---

## 9. Phased Rollout

### Phase 0 — Single-course pilot (target: months 1–3)

Build the lesson format end-to-end against **one** subject (recommendation: Python crash course, ~12 lessons + final exam). Goal: validate the five-module format with real learners, hit leading-indicator targets, fix the format before we scale.

Out of scope for Phase 0: catalog browse, multiple courses, multiple languages in the runner, bookmarks.

### Phase 1 — MVP catalog (target: months 4–6)

Ship the catalog and add 4–6 more courses (suggested: JavaScript, SQL, Git, HTML/CSS, plus one framework — React or Django depending on traction signals). Add multi-language code runner.

### Phase 2 — Engagement & paths (target: months 7–9)

Ship P1 features: skill-level paths, bookmarks, streaks, search across cheat sheets and snippets. Add 5–10 more courses driven by demand.

### Phase 3 — AI tutor & authoring (target: months 10–12+)

Begin P2 work: AI hint system, then community course authoring tools, then certifications and enterprise.

---

## 10. Open Questions

**Blocking before we start building**

- **Pricing model — free, freemium, or paid?** Affects funnel design, account walls, and analytics. _(Owner: stakeholders / business)_
- **Code execution: build vs. buy?** Run our own sandbox (e.g. nsjail, Firecracker) vs. use Judge0, CodeSandbox, or similar. Has cost, security, and language-support implications. _(Owner: engineering)_
- **First subject for the pilot.** Python is the recommendation; alternates are JavaScript or SQL. _(Owner: PM + content)_
- **Animation production model.** In-house animators vs. contract studio vs. AI-assisted (Manim, Motion Canvas, etc.). Drives content velocity and per-course cost. _(Owner: content / design)_

**Resolvable during implementation**

- Lesson schema: how rigidly do we standardize cheat-sheet layouts vs. let each course's content lead drive the design? _(Owner: design + content)_
- Quiz item bank: can we generate plausible distractors with AI assistance and have a human review, or do we author all distractors manually? _(Owner: content)_
- Accessibility on the code runner — what does keyboard-only and screen-reader UX look like for a code editor? _(Owner: design + engineering)_
- Anti-cheating on exams (especially if we add certifications later). _(Owner: engineering, P2)_

**Worth investigating but not blocking**

- Should the cheat sheet be authored as data (JSON → rendered) or as designed components per course? Affects whether we can later auto-export to PDF or print. _(Owner: design + engineering)_
- Do we surface AI explanations in v1 (inline "explain this snippet") or hold the AI tutor for P2? _(Owner: PM)_

---

## 11. Timeline & Dependencies

- **Hard external dependency:** code-execution provider decision (build vs. buy) gates Phase 0 backend work.
- **Content production** is likely the long pole — animated explainers especially. Aim to lock the production pipeline by end of week 4 of Phase 0.
- **Design system for cheat sheets** must exist before the second course is built; otherwise course #2 becomes bespoke and we lose template leverage.
- No known regulatory or contractual deadlines.

---

## 12. Risks

- **Animation production cost and speed** could bottleneck the catalog. Mitigation: invest early in an AI-assisted or modular animation pipeline; have a "cheat sheet first, video later" fallback for any course we want to ship fast.
- **Code-runner abuse** (using our compute as a free general-purpose sandbox). Mitigation: tight per-user rate limits, no network egress from sandboxed code, anomaly alerts.
- **Quality drift across courses** if we scale before locking the lesson template. Mitigation: Phase 0 ends with a written "course quality bar" doc; every course passes the same review checklist.
- **Format fatigue.** If the same five-module structure feels repetitive after a few lessons, completion drops. Mitigation: vary explainer styles, use different quiz types per lesson, and watch the per-lesson drop-off chart.
- **Persona conflict in the UI.** Designing for Anya and Ben on the same page is hard — too much scaffolding annoys Ben, too little overwhelms Anya. Mitigation: collapsible sections, a one-time "what kind of learner are you" prompt, and Phase 1 skill-level paths.

---

## 13. Out of scope for this document

- Detailed engineering architecture (separate ADR / system-design doc).
- Visual design system and brand (separate design brief).
- Go-to-market plan, pricing, and growth strategy (separate doc).
- Content style guide for course authors (separate doc, needed by end of Phase 0).

---

## Appendix A — Feature → Persona coverage matrix

| Feature                 | Anya (Beginner)     | Ben (Stack Switcher) | Carla (Crammer)           |
| ----------------------- | ------------------- | -------------------- | ------------------------- |
| Animated explainer      | Primary entry point | Skim / skip          | Skim / skip               |
| Illustrated cheat sheet | Reinforces video    | Primary entry point  | Quick review              |
| Copy-paste code         | Confidence builder  | Primary value        | Reference during practice |
| Quiz                    | Confidence check    | Sanity check         | Primary practice tool     |
| Coding practice / exam  | Stretch goal        | Primary value        | Primary practice tool     |

Every persona has at least one "primary" module on every lesson page — which is why all five must ship together in Phase 0.
