# Programming Curriculum: JavaScript, TypeScript, Go, and Rust

This curriculum is built around one evolving real-world product: **TaskForge**, a project and task management platform with browser UI, typed client libraries, backend services, workers, command-line tooling, and production-style deployment practices.

The goal is not to memorize syntax. The goal is to build a portfolio-grade system while learning language concepts in the order they become useful.

## Files in this pack

- `01-javascript-curriculum.md`: beginner to advanced JavaScript using a browser-first app and Node.js services.
- `02-typescript-curriculum.md`: TypeScript from basic annotations to advanced type modeling, typed APIs, and migration strategy.
- `03-golang-curriculum.md`: Go from fundamentals to APIs, concurrency, testing, performance, and production services.
- `04-rust-curriculum.md`: Rust from ownership to async services, CLI tooling, performance, and safe systems design.
- `05-project-roadmap.md`: the shared TaskForge product roadmap, architecture, data model, API contracts, and alternative capstone ideas.
- `06-assessment-rubrics.md`: evaluation rubrics, review checklists, and capstone acceptance criteria.
- `SOURCES.md`: source links used to ground the curriculum.

## Recommended learning order

1. **JavaScript**: learn runtime behavior, browser APIs, asynchronous programming, Node.js basics, and application structure.
2. **TypeScript**: add a static type system, model real domains, and build safer APIs and frontends.
3. **Go**: build fast, simple backend services, APIs, workers, and concurrent systems.
4. **Rust**: learn ownership, memory safety, high-performance tools, async services, and systems-level thinking.

You can also run the Go and Rust tracks after TypeScript in parallel, but the easiest portfolio story is to let each language own a different part of TaskForge.

## How to use the curriculum

Use every module as a product sprint:

1. Read the listed concepts.
2. Build the project feature.
3. Write tests or a verification checklist.
4. Refactor once.
5. Document what changed in the project README.
6. Commit the work with a clear message.

A good cadence is 2 to 5 modules per month, depending on available time. The files avoid exact dates so you can adapt the plan to full-time, part-time, or team learning.

## Suggested repository structure

```text
taskforge/
  apps/
    js-browser/
    ts-admin-dashboard/
  services/
    go-api/
    go-worker/
    rust-sync/
    rust-cli/
  packages/
    ts-sdk/
    shared-contracts/
  docs/
    adr/
    api/
    runbooks/
  infra/
    docker/
    ci/
```

## Project learning philosophy

TaskForge starts as a simple local task list and grows into a realistic system:

- A JavaScript browser application for tasks, projects, filters, forms, storage, and async API calls.
- A TypeScript admin dashboard and SDK that model domain entities, API contracts, state machines, and typed errors.
- A Go API that handles users, projects, tasks, events, persistence, background processing, observability, and deployment.
- A Rust CLI and sync/search service for local-first workflows, data import/export, high-throughput processing, and performance-critical features.

## Definition of done for every module

- The feature works from the command line or browser without hidden manual steps.
- The code has meaningful names and a small README note explaining design decisions.
- At least one happy path and one failure path are tested or manually verified.
- The project has been refactored after the first working version.
- You can explain the concept without reading your notes.

## Portfolio outcome

By the end, you should have a multi-language system that demonstrates frontend development, type-safe API design, backend engineering, concurrency, testing, deployment basics, and systems-level programming.
