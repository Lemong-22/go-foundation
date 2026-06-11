# Roadmap — [Project Arc Title]

**Status:** Draft | Approved | Superseded  
**Author:** [nama lo]  
**Date:** YYYY-MM-DD  
**Companion to:** `[PRD file]` vX.Y.Z · `[SPEC file]` v Draft

> **Template origin:** Di-extract dari `vendor/entropy-course/docs/roadmap.md` (Stephen Antoni, 2026-05-26).

---

## 0. Where we are

[1 paragraf: status sekarang. Apa yang sudah jalan, apa yang sudah persist, apa yang udah mature. Pakai present tense.]

> The `[context name]` can persist [aggregates] through a clean hexagonal core: [N] CLI commands → inbound ports → usecases → outbound ports → [DB]. [Secondary adapters] wrap the same usecases. Auth, REST, media, and any [other concern]-facing concern are explicitly deferred in `[spec file]` §6.

[1 paragraf lagi: arah selanjutnya. Apa goal dari arc ini.]

> A [Entity] today carries exactly [N] content fields — `[field_1]`, `[field_2]`. The goal of the next arc of work is to make the platform able to [goal: model richer content, add slices, etc.].

This document sequences that work. §1 records the decisions it rests on so future readers know *why* the phases are shaped this way.

---

## 1. Decisions this roadmap rests on

[Cabut SEMUA keputusan arsitektur yang lo ambil. Format tabel. Tiap keputusan harus punya "kenapa" — alternative yang lo consider dan kenapa lo ga pilih itu.]

| # | Decision | Choice |
|---|----------|--------|
| 1 | [Concern] | [Choice with rationale] |
| 2 | [Concern] | [Choice with rationale] |
| 3 | [Concern] | [Choice with rationale] |
| 4 | [Concern] | [Choice with rationale] |
| 5 | [Concern] | [Choice with rationale] |
| 6 | [Concern] | [Choice with rationale] |
| 7 | [Concern] | [Choice with rationale] |
| 8 | [Concern] | [Choice with rationale] |
| 9 | [Concern] | [Choice with rationale] |
| 10 | [Concern] | [Choice with rationale] |
| 11 | [Concern] | [Choice with rationale] |
| 12 | [Stack] | [Choice with rationale] |
| 13 | [Sequencing] | [Choice with rationale] |

**Contoh entry (extract dari entropy-course):**

> | 1 | Phase focus | **Authoring-first.** The frontend is an instructor/author console. All learner-facing concerns (enrollment, progress, code execution, grading, gamification) are a later phase. |
> | 4 | Content model shape | **Lesson → ordered typed `ContentBlock`s.** Text and video are inline blocks; quiz/practice blocks reference standalone aggregates by id. **Test** attaches at course level. |
> | 13 | Sequencing | **Foundation first, then vertical slices per content type.** Each content type goes domain → CLI → REST → console before the next begins. |

---

## 2. Target architecture

[ASCII diagram: depict adapters → ports → core → ports → adapters. Boleh pake `text` code block.]

```text
                  ┌──────────────┐   ┌──────────────┐
   [agent] ─────► │   CLI adapter│   │ REST adapter │ ◄──── [user type]
   ([use case])   │  (+ [extra]) │   │ (+ [extra])  │
                  └──────┬───────┘   └──────┬───────┘
                         │                  │
                         ▼                  ▼
              ┌────────────────────────────────────────┐
              │  Inbound ports (EntityAService, ...)  │
              │  shared DTOs                           │
              └───────────────────┬────────────────────┘
                                  ▼
              ┌────────────────────────────────────────┐
              │  Usecases  →  Domain  ([aggregates,   │
              │  value objects])                       │
              └───────────────────┬────────────────────┘
                                  ▼
              ┌────────────────────────────────────────┐
              │  Outbound ports → [DB] repos,           │
              │  IDGenerator, Clock                    │
              └────────────────────────────────────────┘
```

The dependency rule from `[spec file]` §3 is unchanged: every adapter wraps the same usecases; ports are declared in the core and speak only domain types.

---

## 3. Phased roadmap

Phases are ordered by **dependency, not by date** — consistent with the [curriculum's] no-fixed-dates philosophy. Each phase is independently shippable.

### Phase A — [Phase Name]

**Goal:** [1-2 kalimat: apa yang di-deliver di phase ini]

**Domain & core**
- [Bullets: perubahan domain, new aggregates, new VOs, etc.]

**Persistence**
- [Bullets: migration, schema, indexes]

**[Adapter 1]**
- [Bullets: features, endpoints/commands, etc.]

**[Adapter 2]**
- [Bullets: features, etc.]

**Exit criteria:** [Bullets: how do we know this phase is done]

### Phase B — [Phase Name]

[Same structure]

### Phase C — [Phase Name]

[Same structure]

---

## 4. Deferred — [Next Arc] (named so it is not forgotten)

Out of scope for this roadmap, but the natural next arc. Listed so phase boundaries above stay honest:

- [Item 1]
- [Item 2]
- [Item 3]

---

## 5. Cross-cutting principles

- **Hexagonal discipline holds.** Every new adapter wraps usecases through inbound ports; new outbound needs are ports declared in the core.
- **DTO reuse.** [Adapters] share the inbound-port input/output DTOs — no parallel type set.
- **Vertical slices.** Per `[spec file]`, each new command/endpoint is its own slice through the stack; no content type shares a usecase with another.
- **Migrations stay sequential** numbered files; `migrate` tooling is wired in `main.go`, outside the bounded context.
- **Tests per slice:** in-memory fakes for usecase tests, [DB] integration tests for repositories.

---

## 6. Risks & open questions

- **[Risk 1].** [Mitigation]
- **[Risk 2].** [Mitigation]
- **[Open question].** [What we need to decide before phase X]

---

## Catatan Pakai Template Ini

1. **§1 Decisions** — INI PALING PENTING. Tanpa decisions, roadmap cuma timeline. Decisions = "kenapa begini, bukan begitu". 3-6 bulan dari sekarang, lo bakal lupa kenapa lo pilih X. Tulis di sini.
2. **§2 Architecture diagram** — visualisasi. 5 menit bikin diagram, hemat 1 jam debat ke diri sendiri nanti.
3. **§3 Exit criteria** — kalau ga ada, phase ga pernah "done". Infinite scope.
4. **§4 Deferred** — TULIS. "Yang ga dibangun" = keputusan juga. Listed so future-you tau boundary-nya.
5. **§6 Risks** — honest self-assessment. Identifikasi sendiri sebelum diidentifikasi orang lain.
