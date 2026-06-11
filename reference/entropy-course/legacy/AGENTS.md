## Stacks In Use

- Bun (runtime)
- TypeScript
- Tanstack
- React
- Shadcn
- Better Auth
- tRPC

## Design Pattern

You are to develop with DDD (Domain Driven Design) and hexagonal architecture pattern for newer modules and services, but for existing modules and services, you are to follow the existing architecture and design pattern (which is MVC).
Observe the rules using skill `${clean-code}` whenever making design decisions and implementations.

## Local Skills

When a task matches a local skill, load that skill's `SKILL.md` before acting and follow only the relevant referenced material. Use `.agents/skills/<name>` directory names as the canonical skill tokens.

- `${better-auth-best-practices}`: Use for Better Auth setup, auth config, sessions, adapters, plugins, and env handling.
- `${bun}`: Use for Bun runtime, scripts, dependency management, tests, and builds.
- `${clean-code}`: Use for code design, implementation, review, and refactoring quality.
- `${clerk-nextjs-patterns}`: Use for Clerk auth patterns in Next.js.
- `${design-md}`: Use for analyzing Stitch projects and synthesizing semantic design systems into DESIGN.md files.
- `${gh-cli}`: Use for GitHub CLI workflows.
- `${hyperframes}`: Use for video compositions, animations, title cards, overlays, captions, voiceovers, audio-reactive visuals, and scene transitions in HyperFrames HTML. Covers composition authoring, timing, media, and the full video production workflow. For dev-loop CLI commands (init, lint, inspect, preview, render) see the hyperframes-cli skill; for asset preprocessing commands (tts, transcribe, remove-background) see the hyperframes-media skill.
- `${hyperframes-cli}`: Use for HyperFrames CLI dev loop — `npx hyperframes` for scaffolding (init), validation (lint, inspect), preview, render, and environment troubleshooting (doctor, browser, info, upgrade). For asset preprocessing commands (`tts`, `transcribe`, `remove-background`), invoke the `hyperframes-media` skill instead.
- `${impeccable}`: Use for frontend design, UX, UI polish, accessibility, and visual refinement.
- `${next-best-practices}`: Use for Next.js best practices - file conventions, RSC boundaries, data patterns, async APIs, metadata, error handling, route handlers, image/font optimization, bundling.
- `${next-cache-components}`: Use for Next.js 16 Cache Components - PPR, use cache directive, cacheLife, cacheTag, updateTag.
- `${remotion-best-practices}`: Use for Remotion video creation in React.
- `${shadcn}`: Use for shadcn/ui components, registry use, composition, forms, styling, and icons.
- `${sync-local-skills}`: Use for discovering repo-local skills and updating AGENTS.md with Local Skills trigger rules.
- `${tanstack-start-best-practices}`: Use for TanStack Start server functions, middleware, SSR, auth, and deployment patterns.
- `${trpc-tanstack-nextjs}`: Use for tRPC with TanStack Query in Next.js App Router.
- `${turborepo}`: Use for Turborepo monorepo build system, task pipelines, caching, remote cache, --filter, --affected, CI optimization, environment variables, and monorepo structure/best practices.
- `${vercel-composition-patterns}`: Use for React composition patterns that scale. Triggers on compound components, render props, context providers, component architecture, and React 19 API changes.
- `${vercel-react-best-practices}`: Use for React and Next.js performance-sensitive work.
- `${web-design-guidelines}`: Use for reviewing UI code for Web Interface Guidelines compliance. Triggers on "review my UI", "check accessibility", "audit design", "review UX", or "check my site against best practices".

## Running Tests

Whenever running tests for investigation or validation, do not run the entire full suite. You should focus on the relevant test-files only and run them selectively.
