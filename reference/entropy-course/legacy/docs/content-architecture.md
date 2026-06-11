# Content architecture

How course content is stored, authored, built, and served.

## TL;DR

- **Authoring source of truth** lives in the repo under `content/` as MDX + YAML. Git is the version history.
- **Media** (video, large images, PDFs) lives in object storage (Mux for video, R2/S3 for everything else). The repo references it by ID, never by file.
- **Postgres** holds a *content index* (a mirror of the manifest) plus per-user state (enrollment, progress, quiz attempts, comments).
- **A publish pipeline** parses `content/`, computes hashes, uploads any new media, and upserts the content index in Postgres.
- **Caching** is layered: CDN edge for static lesson pages and HLS video, Redis for hot per-user reads, Postgres on miss.

Don't put MDX into the database. Don't put videos into Git.

## Folder layout

```
entropy-course/
├── content/                          # all course source material
│   ├── tracks/
│   │   ├── javascript/
│   │   │   ├── track.yaml            # track metadata, ordered stage list
│   │   │   ├── cover.svg
│   │   │   ├── 01-foundations/
│   │   │   │   ├── stage.yaml        # stage metadata, ordered lesson list
│   │   │   │   ├── 01-values-and-types.mdx
│   │   │   │   ├── 02-functions-and-scope.mdx
│   │   │   │   ├── 03-arrays-and-objects.mdx
│   │   │   │   ├── assets/
│   │   │   │   │   ├── scope-diagram.svg
│   │   │   │   │   └── closure-walkthrough.png
│   │   │   │   └── quiz.yaml         # questions for this stage
│   │   │   ├── 02-browser-app/
│   │   │   │   └── ...
│   │   │   ├── 03-persistence/
│   │   │   ├── 04-modules/
│   │   │   ├── 05-remote-sync/
│   │   │   └── 06-production-polish/
│   │   ├── typescript/
│   │   ├── golang/
│   │   └── rust/
│   ├── shared/
│   │   ├── glossary.mdx              # reusable definitions
│   │   ├── references.bib
│   │   └── components/               # MDX-embeddable widgets (TSX)
│   │       ├── Callout.tsx
│   │       └── CodeRunner.tsx
│   └── manifest.json                 # GENERATED — do not edit by hand
│
├── apps/
│   ├── web/                          # learner-facing app (Next.js)
│   └── server/                       # API server
│
├── packages/
│   ├── db/                           # drizzle schema + client (existing)
│   │   └── src/schema/
│   │       ├── auth.ts               # better-auth tables (existing)
│   │       └── content.ts            # content index + progress (new)
│   ├── content/                      # MDX parsing, types, manifest reader
│   │   ├── src/parse.ts
│   │   ├── src/manifest.ts
│   │   └── src/types.ts
│   ├── api/                          # existing
│   ├── auth/                         # existing
│   ├── env/                          # existing
│   ├── ui/                           # existing
│   └── config/                       # existing
│
├── scripts/
│   ├── build-manifest.ts             # walk content/, emit manifest.json
│   ├── sync-content-index.ts         # upsert manifest rows into Postgres
│   ├── upload-media.ts               # push new videos to Mux, write IDs back
│   └── reindex-search.ts             # ship corpus to Meilisearch/Typesense
│
├── content-cache/                    # GENERATED — built lesson HTML, gitignored
└── docs/
    └── content-architecture.md       # this file
```

### Why this shape

Your existing `curriculum/*.md` files map onto **tracks**. The "Stage" table in those docs maps onto **stages**. Each row in those stage tables becomes a **lesson**. So the natural hierarchy is `track → stage → lesson`, three levels deep. We use a numeric prefix on directories (`01-foundations`) so the filesystem sorts the same way the UI will.

`assets/` co-locates per-lesson images and diagrams with the MDX that uses them. That keeps PR diffs sensible — if you rewrite lesson 3, you also touch its diagrams in the same change.

Heavy media never lives in the repo. The build pipeline uploads new video files to Mux and writes the asset ID back into the MDX frontmatter. Once the ID exists, the file can be removed from your machine — Git only ever stored the reference.

## File conventions

### `track.yaml`

```yaml
id: javascript
slug: javascript
title: JavaScript — Beginner to Advanced
summary: Build a production-style browser app from plain objects to async modules.
order_index: 1
version: 1.0.0
authors:
  - stephen
status: published       # draft | review | published
stages:
  - 01-foundations
  - 02-browser-app
  - 03-persistence
  - 04-modules
  - 05-remote-sync
  - 06-production-polish
```

### `stage.yaml`

```yaml
slug: 01-foundations
title: Foundations
order_index: 1
objectives:
  - Explain values, types, and functions
  - Read and write small JavaScript programs confidently
lessons:
  - 01-values-and-types
  - 02-functions-and-scope
  - 03-arrays-and-objects
quiz: quiz                # → quiz.yaml in this folder
```

### Lesson MDX frontmatter

```yaml
---
slug: 01-values-and-types
title: Values and types
order_index: 1
estimated_minutes: 18
objectives:
  - Distinguish primitive and reference types
  - Predict equality and coercion outcomes
prerequisites: []
status: published
video:
  provider: mux           # set to null while drafting
  asset_id: AbCdEf123     # filled in by upload-media.ts
tags: [foundations, types]
---
```

The `video.asset_id` field is the contract between repo content and Mux. The build script never re-uploads a video unless the local source file changes; the asset ID, once written, is permanent.

### `quiz.yaml`

```yaml
slug: foundations
pass_threshold: 0.7
questions:
  - slug: q1-primitive-vs-reference
    type: single_choice
    prompt: Which of these is a primitive value?
    options: [array, object, number, function]
    answer: number
  - slug: q2-equality-coercion
    type: short_answer
    prompt: What does `[] == ![]` evaluate to, and why?
    answer_key: "true; ![] is false (0), [] coerces to 0, 0 == 0."
```

Quizzes stay declarative and in Git. **Only attempts** are stored in Postgres — never the question text.

## The manifest

`scripts/build-manifest.ts` walks `content/`, parses every YAML and MDX frontmatter, computes a SHA of each lesson's source files (`.mdx` + its `assets/` dir), and emits `content/manifest.json`. That file looks roughly like:

```json
{
  "tracks": [{
    "id": "javascript",
    "slug": "javascript",
    "version": "1.0.0",
    "content_hash": "9f3a…",
    "stages": [{
      "id": "javascript:01-foundations",
      "slug": "01-foundations",
      "order_index": 1,
      "lessons": [{
        "id": "javascript:01-foundations:01-values-and-types",
        "slug": "01-values-and-types",
        "order_index": 1,
        "content_hash": "1c8e…",
        "video": { "provider": "mux", "asset_id": "AbCdEf123" }
      }]
    }]
  }]
}
```

The manifest is the **only thing** the publish step needs to insert into Postgres. The DB never sees raw MDX.

## Publish flow

```
   author commits MDX
          │
          ▼
   build-manifest.ts        ← walks content/, hashes files
          │
          ▼
   upload-media.ts          ← uploads new videos, writes asset IDs back
          │
          ▼
   sync-content-index.ts    ← upserts track/stage/lesson rows in Postgres
          │
          ▼
   reindex-search.ts        ← ships corpus to Meilisearch
          │
          ▼
   render & deploy          ← Next.js builds static lesson pages, CDN invalidates
```

Each step is idempotent and uses `content_hash` to short-circuit no-op rows. Re-running the whole pipeline on an unchanged repo should be a near-instant no-op.

## How the app reads it

Lesson pages: `apps/web` resolves a URL like `/javascript/foundations/values-and-types` by querying the Postgres content index for the lesson row, then loading the pre-rendered HTML from `content-cache/` (or rebuilding it via ISR). The video player is handed the Mux asset ID and streams HLS from the CDN.

Per-user state: progress, last position, quiz attempts come from the DB through Redis. Writes go straight to Postgres; reads check Redis first with a short TTL (30–60s for hot reads, longer for stable data like enrollment).

## What goes where — the rules

| Thing                          | Lives in                | Why                                              |
| ------------------------------ | ----------------------- | ------------------------------------------------ |
| Lesson body (prose, code)      | MDX in Git              | Diffable, PR-reviewable, easy to author          |
| Diagrams, small images         | `assets/` next to MDX   | Co-located edits, small enough for Git           |
| Videos, large media            | Mux / R2 / S3           | Don't bloat the repo; need a CDN regardless      |
| Quiz questions                 | `quiz.yaml` in Git      | Declarative, versionable, no CMS needed          |
| Course/stage/lesson IDs        | Postgres (`content` schema) | Stable join targets for user data           |
| Enrollment, progress, attempts | Postgres                | Per-user, changes constantly                     |
| Hot per-user reads             | Redis                   | Sub-ms reads without hammering Postgres          |
| Search index                   | Meilisearch / Typesense | Built from MDX at publish time                   |
| Static lesson pages, HLS       | CDN edge                | Cheap to serve at scale, invalidated on publish  |

## Migration from `curriculum/*.md`

You already have `curriculum/01-javascript-curriculum.md` etc. The migration is mechanical:

1. For each `NN-<language>-curriculum.md`, create `content/tracks/<language>/track.yaml`.
2. For each stage row in the existing milestone-map table, create a `NN-<stage>/stage.yaml`.
3. Split the prose under each stage heading into one MDX file per lesson concept. The frontmatter is straightforward; the body is the existing markdown.
4. Move asset references into a sibling `assets/` directory per stage.
5. Delete the original `curriculum/*.md` once a track has been ported, or move it to `legacy/`.

The current files are great as authoring drafts. They just need to be sliced thinner so the app can route to lesson granularity.
