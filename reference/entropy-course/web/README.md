# Entropy Course Web

Bun-managed Next.js App Router application for the Phase L1 learner reader.

## Prerequisites

- Bun 1.3+
- Node.js 20.9+ available to Next.js
- The Go REST API running locally when data-backed reader routes are added

## Local Development

Install dependencies:

```sh
bun install
```

Start the development server:

```sh
bun run dev
```

Open:

```text
http://localhost:3000
```

## Checks

```sh
bun run test
bun run lint
bun run typecheck
bun run build
```

## Environment

Course API reads are wrapped by typed server-only functions under
`src/lib/course-api/server.ts`. Do not expose the course API token through
`NEXT_PUBLIC_*` variables.

```sh
COURSE_API_BASE_URL=http://127.0.0.1:8788
COURSE_CLI_API_TOKEN=replace-with-local-token
COURSE_REVALIDATION_SECRET=replace-with-republish-hook-secret
```

## Published Read Caching

Published reader data is cached with explicit Next.js fetch tags and no timer
revalidation. The app waits for an incoming request before reading server-only
environment variables, then caches course API responses with `force-cache`.

Republish/import workflows should call the server-only revalidation route after
changing published content:

```http
POST /api/revalidate
Authorization: Bearer $COURSE_REVALIDATION_SECRET
Content-Type: application/json
```

Accepted bodies:

```json
{ "scope": "catalog" }
{ "scope": "course", "courseID": "course-id" }
{ "scope": "lesson", "lessonID": "lesson-id", "courseID": "course-id" }
{ "scope": "block", "blockID": "block-id", "lessonID": "lesson-id" }
```

The route immediately expires only the matching cache tags. The next request for
that reader data blocks on a fresh Course API read instead of waiting on a
timer. `COURSE_REVALIDATION_SECRET` must stay server-only and must not use a
`NEXT_PUBLIC_` prefix.
