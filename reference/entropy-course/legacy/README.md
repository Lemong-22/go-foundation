# entropy-course

This project was created with [Better-T-Stack](https://github.com/AmanVarshney01/create-better-t-stack), a modern TypeScript stack that combines Next.js, Express, TRPC, and more.

## Features

- **TypeScript** - For type safety and improved developer experience
- **Next.js** - Full-stack React framework
- **TailwindCSS** - Utility-first CSS for rapid UI development
- **Shared UI package** - shadcn/ui primitives live in `packages/ui`
- **Express** - Fast, unopinionated web framework
- **tRPC** - End-to-end type-safe APIs
- **Bun** - Runtime environment
- **Drizzle** - TypeScript-first ORM
- **PostgreSQL** - Database engine
- **Authentication** - Better-Auth
- **Turborepo** - Optimized monorepo build system

## Local deployment: Stephen + Yosi quick start

These steps bring up the local PostgreSQL container, apply the Drizzle schema, seed two local demo accounts, and start the web/API apps.

### 1. Prerequisites

Install these first:

- [Bun](https://bun.com/) 1.3+
- Docker Desktop, Docker Engine, or another Docker-compatible runtime with `docker compose`

From the project root, install dependencies:

```bash
bun install
```

### 2. Create local env files

Copy the committed examples:

```bash
cp apps/server/.env.example apps/server/.env
cp apps/web/.env.example apps/web/.env
cp .env.example .env
```

For normal local development, the example values are enough:

- API: `http://localhost:3000`
- Web: `http://localhost:3001`
- PostgreSQL: `localhost:5432`, database `entropy_course`

If your machine already uses port `5432`, edit `POSTGRES_PORT` in the root `.env` file or stop the conflicting database first.

### 3. Start PostgreSQL

```bash
bun run db:start
```

Check that the container is healthy:

```bash
docker compose ps postgres
```

### 4. Apply the database schema

```bash
bun run db:push
```

This creates the Better Auth tables in the local PostgreSQL database.

### 5. Seed local users

```bash
bun run db:seed
```

The seed is idempotent, so it is safe to run again after schema changes or after resetting the database.

Default seeded accounts:

- Stephen: `stephen.local@example.com`
- Yosi: `yosi.local@example.com`
- Password for both: `local-password-123`

Optional overrides:

```bash
# stephen.local@example.com
# yosi.local@example.com
# local-password-123
bun run db:seed
```

Do not reuse real production passwords for local seeds.

### 6. Start the local app

```bash
bun run dev
```

Open [http://localhost:3001](http://localhost:3001) in your browser.
The API runs at [http://localhost:3000](http://localhost:3000).

### 7. Sign in

Go to [http://localhost:3001/login](http://localhost:3001/login), switch to sign-in if needed, and use one of the seeded accounts above.

### Useful local database commands

```bash
bun run db:studio  # inspect data in Drizzle Studio
bun run db:stop    # stop PostgreSQL, keep data volume
bun run db:down    # stop services and remove Compose network; volume remains unless removed manually
```

## Database Setup

This project uses PostgreSQL with Drizzle ORM. A root `docker-compose.yml` is provided for local infrastructure.

## Understand Anything

Use Understand Anything to generate a local codebase knowledge graph and inspect it in the interactive dashboard.

### Setup

Install these once on your machine:

- Node.js 22+
- Bun 1.3+
- the Understand Anything plugin checkout

Set `UNDERSTAND_PLUGIN_ROOT` to your plugin checkout. The default below matches the local Codex plugin install path:

```bash
export UNDERSTAND_PLUGIN_ROOT="${UNDERSTAND_PLUGIN_ROOT:-$HOME/.understand-anything/repo/understand-anything-plugin}"
```

Build the plugin core before the first analysis:

```bash
bun install --cwd "$UNDERSTAND_PLUGIN_ROOT"
bun run --cwd "$UNDERSTAND_PLUGIN_ROOT/packages/core" build
```

### Generate the knowledge graph

From the project root, run the Understand Anything analysis in Codex:

```text
/understand --full
```

The analysis writes the graph to:

```text
.understand-anything/knowledge-graph.json
```

Review `.understand-anything/.understandignore` before running a full scan. For this repo, add `.agents/` and `.playwright-mcp/` if you do not want local agent skills or browser artifacts in the graph. Add `legacy/` as well if you only want the current Better-T-Stack app.

### Serve the dashboard

The dashboard reads the graph from `GRAPH_DIR` and prints a one-time tokenized URL. Use the full URL that includes `?token=...`.

You can also run `/understand-dashboard` in Codex after the graph exists. To serve it manually:

```bash
PROJECT_ROOT="$(pwd)"
bun install --cwd "$UNDERSTAND_PLUGIN_ROOT"
bun run --cwd "$UNDERSTAND_PLUGIN_ROOT/packages/core" build
GRAPH_DIR="$PROJECT_ROOT" bun run --cwd "$UNDERSTAND_PLUGIN_ROOT/packages/dashboard" dev --host 127.0.0.1
```

If the graph exists, the terminal prints a dashboard URL like:

```text
http://127.0.0.1:5173/?token=<token>
```

Open that exact URL. Without the token, the dashboard blocks access to the graph data.

## Docker

Use Docker Compose to run the full production-like stack locally:

```bash
cp .env.example .env
# edit BETTER_AUTH_SECRET before using beyond local development
bun run docker:build
bun run docker:up
```

Services:

- `postgres`: PostgreSQL 17 on `localhost:5432`
- `server`: Express/tRPC/Better Auth API on `localhost:3000`
- `web`: Next.js app on `localhost:3001`

Helpful commands:

```bash
bun run docker:logs
bun run docker:down
```

## CI/CD

GitHub Actions is configured in `.github/workflows/ci.yml` to:

- install dependencies with Bun using the lockfile
- run TypeScript checks
- build the monorepo
- build Docker Compose images

Future deployment can extend this workflow with a registry push and environment-specific deploy job.

## UI Customization

React web apps in this stack share shadcn/ui primitives through `packages/ui`.

- Change design tokens and global styles in `packages/ui/src/styles/globals.css`
- Update shared primitives in `packages/ui/src/components/*`
- Adjust shadcn aliases or style config in `packages/ui/components.json` and `apps/web/components.json`

### Add more shared components

Run this from the project root to add more primitives to the shared UI package:

```bash
npx shadcn@latest add accordion dialog popover sheet table -c packages/ui
```

Import shared components like this:

```tsx
import { Button } from "@entropy-course/ui/components/button";
```

### Add app-specific blocks

If you want to add app-specific blocks instead of shared primitives, run the shadcn CLI from `apps/web`.

## Project Structure

```
entropy-course/
├── apps/
│   ├── web/         # Frontend application (Next.js)
│   └── server/      # Backend API (Express, TRPC)
├── packages/
│   ├── ui/          # Shared shadcn/ui components and styles
│   ├── api/         # API layer / business logic
│   ├── auth/        # Authentication configuration & logic
│   └── db/          # Database schema & queries
```

## Available Scripts

- `bun run dev`: Start all applications in development mode
- `bun run build`: Build all applications
- `bun run dev:web`: Start only the web application
- `bun run dev:server`: Start only the server
- `bun run check-types`: Check TypeScript types across all apps
- `bun run db:push`: Push schema changes to database
- `bun run db:generate`: Generate database client/types
- `bun run db:migrate`: Run database migrations
- `bun run db:seed`: Seed Stephen/Yosi local demo users
- `bun run db:studio`: Open database studio UI
