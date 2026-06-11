# TypeScript Curriculum: Beginner to Advanced

## Track goal

Build a type-safe TaskForge layer: shared contracts, SDK, admin dashboard, and optional backend-for-frontend. This track turns the JavaScript project into a maintainable system where domain rules, API contracts, workflows, forms, and errors are modeled explicitly.

## What you will be able to do

- Use TypeScript strict mode effectively without fighting the compiler.
- Model real business domains with interfaces, unions, generics, utility types, and runtime validation.
- Build typed SDKs, dashboards, backend handlers, tests, and API contracts.
- Migrate JavaScript gradually and safely.
- Publish or package a reusable TypeScript library.

## Primary references used

- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/)
- [TypeScript Utility Types](https://www.typescriptlang.org/docs/handbook/utility-types.html)

## Project milestone map

| Stage | Product feature         | Concepts practiced                         | Evidence of completion                   |
| ----- | ----------------------- | ------------------------------------------ | ---------------------------------------- |
| 1     | Shared domain contracts | strict mode, interfaces, type aliases      | Typed TaskForge entities and fixtures    |
| 2     | Typed UI and workflows  | unions, narrowing, state machines          | Impossible states are represented safely |
| 3     | Typed SDK               | generics, async types, errors, validation  | No `Promise<any>` in API client          |
| 4     | Admin dashboard         | typed components, forms, routes            | Dashboard consumes SDK safely            |
| 5     | API contracts           | OpenAPI, versioning, generated types       | Contract tests and changelog             |
| 6     | Publishable package     | declaration files, package exports, semver | Local consumer app uses SDK              |

## Curriculum modules

Each module is designed as a small product sprint. Do not move on until the deliverable works, has tests or manual verification notes, and is committed with a clear README update.

### Module 1: TypeScript mental model, compiler, and strict setup

**Level:** Beginner

**Concepts to learn**

- TypeScript as JavaScript plus static type checking
- `tsc`, `tsconfig.json`, strict mode, and editor feedback
- Type inference versus explicit annotations
- The difference between compile-time types and runtime values

**Project build**

- Create `packages/shared-contracts` and `packages/ts-sdk`.
- Define basic TaskForge types for IDs, users, projects, tasks, comments, and events.
- Compile TypeScript to JavaScript and run the output.

**Practice tasks**

- Turn on strict compiler settings from the start.
- Annotate function boundaries and let local variables be inferred.
- Write a note explaining what TypeScript does not check at runtime.

**Completion checklist**

- [ ] The project compiles with strict settings.
- [ ] You can explain inference and annotations with examples.
- [ ] No `any` is used without a written reason.

### Module 2: Everyday types, interfaces, type aliases, and structural typing

**Level:** Beginner

**Concepts to learn**

- Primitive, object, array, tuple, literal, and optional types
- Interfaces versus type aliases
- Structural typing and excess property checks
- Readonly and immutability conventions

**Project build**

- Model TaskForge domain entities with interfaces and type aliases.
- Create readonly API response types and mutable form draft types.
- Add typed fixture data for tests and demos.

**Practice tasks**

- Convert the JavaScript task model into typed definitions.
- Try interface extension and type intersections for shared fields.
- Create intentionally invalid fixtures and observe compiler feedback.

**Completion checklist**

- [ ] Domain models are clear and not over-abstracted.
- [ ] The compiler catches malformed fixtures.
- [ ] You can explain structural typing without relying on class inheritance.

### Module 3: Unions, narrowing, discriminated unions, and state modeling

**Level:** Beginner

**Concepts to learn**

- Union and intersection types
- Narrowing with `typeof`, `in`, equality, truthiness, and custom type guards
- Discriminated unions
- Exhaustive checking with `never`

**Project build**

- Model UI state as `idle`, `loading`, `success`, and `failure` variants.
- Model task status and sync status as discriminated unions.
- Create functions that must handle every state.

**Practice tasks**

- Replace boolean flags such as `isLoading` and `hasError` with state unions.
- Add an exhaustive switch helper.
- Use custom type guards for imported JSON.

**Completion checklist**

- [ ] Impossible UI states are hard to represent.
- [ ] Switch statements fail compilation when a new state is missing.
- [ ] You can explain narrowing from runtime checks.

### Module 4: Typed functions, classes, generics, and reusable APIs

**Level:** Beginner

**Concepts to learn**

- Function parameter and return types
- Optional and default parameters
- Generic functions and constraints
- Class fields, visibility, and constructor invariants

**Project build**

- Create typed validators, selectors, and repository interfaces.
- Build a generic `Result<T, E>` or `ApiResponse<T>` shape.
- Create a typed in-memory repository for tasks.

**Practice tasks**

- Write generic helpers for pagination and sorting.
- Constrain generic functions to objects that have IDs.
- Compare class-based and function-based repositories.

**Completion checklist**

- [ ] Generic helpers preserve useful type information.
- [ ] Repository APIs reveal intent without leaking implementation details.
- [ ] You can explain the difference between generic constraints and `any`.

### Module 5: Modules, project references, package boundaries, and monorepos

**Level:** Intermediate

**Concepts to learn**

- ES modules in TypeScript
- `type`-only imports and exports
- Package boundaries and public APIs
- Project references and build order

**Project build**

- Create a small monorepo with shared contracts, SDK, and admin dashboard packages.
- Export only stable public types from package entry points.
- Add build scripts for each package and a root script for all packages.

**Practice tasks**

- Move internal helper types so consumers cannot depend on them.
- Document package responsibilities.
- Break a public API intentionally and observe downstream compile failures.

**Completion checklist**

- [ ] Package boundaries are explicit.
- [ ] Build order is documented.
- [ ] The shared contract package has no UI or server dependency.

### Module 6: Typed async clients and API contracts

**Level:** Intermediate

**Concepts to learn**

- Promises, `Awaited`, async function return types
- Typed request and response objects
- HTTP error modeling
- Runtime validation at trust boundaries

**Project build**

- Create a `TaskForgeClient` SDK that wraps the JavaScript or Go API.
- Type every request body, response body, and error shape.
- Validate unknown JSON before treating it as trusted data.

**Practice tasks**

- Create typed methods for `listTasks`, `createTask`, `updateTask`, and `completeTask`.
- Simulate a server returning an unexpected shape.
- Map HTTP status codes to typed errors.

**Completion checklist**

- [ ] No endpoint returns `Promise<any>`.
- [ ] Runtime validation protects API boundaries.
- [ ] The SDK can be used by both a dashboard and tests.

### Module 7: Advanced generics, utility types, mapped types, and conditional types

**Level:** Intermediate

**Concepts to learn**

- Generic constraints and defaults
- `Pick`, `Omit`, `Partial`, `Required`, `Record`, `ReturnType`, and custom utility types
- Mapped types and key remapping
- Conditional types and `infer`

**Project build**

- Create form draft types from domain models.
- Create update payload types that allow only editable fields.
- Create API response helpers that infer data and error types.

**Practice tasks**

- Implement `DeepReadonly` for selected domain objects.
- Create `EditableTaskFields` from the Task type.
- Write type tests using assignment examples or a type test tool.

**Completion checklist**

- [ ] Utility types reduce duplication rather than hide meaning.
- [ ] Advanced types are named and documented.
- [ ] You can explain one conditional type in plain language.

### Module 8: Runtime validation, schemas, and type-safe forms

**Level:** Intermediate

**Concepts to learn**

- The gap between static types and runtime data
- Schema validation libraries or custom validators
- Type inference from schemas
- Form error mapping

**Project build**

- Validate task creation and update forms.
- Validate imported JSON and API responses.
- Return field-specific error messages that the UI can render.

**Practice tasks**

- Create one schema manually and one with a validation library.
- Infer a TypeScript type from a schema where possible.
- Add tests for invalid imports and malformed API responses.

**Completion checklist**

- [ ] External data is treated as `unknown` until validated.
- [ ] Form errors are typed and field-specific.
- [ ] You can explain why TypeScript alone cannot validate JSON at runtime.

### Module 9: Typed state machines and domain workflows

**Level:** Intermediate

**Concepts to learn**

- Workflow modeling with discriminated unions
- Transition functions
- Command and event patterns
- Illegal state prevention

**Project build**

- Model task lifecycle transitions: draft, active, blocked, completed, archived.
- Model sync workflow: queued, sending, accepted, rejected, retrying.
- Create typed events for task lifecycle changes.

**Practice tasks**

- Prevent completing an archived task at the type or function boundary.
- Write transition tests for every allowed and disallowed state change.
- Generate dashboard data from events.

**Completion checklist**

- [ ] Domain rules are centralized.
- [ ] Workflow transitions are tested.
- [ ] You can show at least one bug prevented by better types.

### Module 10: Testing TypeScript and type behavior

**Level:** Intermediate

**Concepts to learn**

- Unit, integration, and type-level tests
- Typed mocks and fakes
- Testing async SDK behavior
- Compiler as part of the test suite

**Project build**

- Add tests for validators, SDK methods, state transitions, and utility types.
- Create typed fake clients for dashboard tests.
- Add `tsc --noEmit` to CI checks.

**Practice tasks**

- Write a test that fails at compile time when an event type changes.
- Mock one HTTP success and one HTTP failure.
- Use fixtures without widening literal types accidentally.

**Completion checklist**

- [ ] The test suite covers runtime behavior and important type behavior.
- [ ] Mocks stay aligned with real interfaces.
- [ ] A clean clone can run type checks and tests.

### Module 11: Typed frontend dashboard architecture

**Level:** Advanced

**Concepts to learn**

- Component props and state typing
- Typed routing and URL parameters
- Data fetching states
- Framework-agnostic architecture; optional React, Vue, Svelte, or vanilla TypeScript

**Project build**

- Build a TaskForge admin dashboard with project list, task detail, filters, and audit log.
- Use typed SDK methods for all remote calls.
- Keep domain and SDK code independent from the UI framework.

**Practice tasks**

- Create a typed route definition for task detail pages.
- Add reusable typed table and form components.
- Model all loading and error states explicitly.

**Completion checklist**

- [ ] The UI cannot call nonexistent API methods.
- [ ] Props and route parameters are typed.
- [ ] Domain logic is not buried inside components.

### Module 12: Typed Node.js service or backend-for-frontend

**Level:** Advanced

**Concepts to learn**

- Type-safe request handlers
- Validation middleware
- Typed configuration
- Backend-for-frontend pattern

**Project build**

- Create an optional TypeScript BFF that adapts the Go API for the admin dashboard.
- Add typed handlers for dashboard-specific views.
- Validate environment variables at startup.

**Practice tasks**

- Build a typed endpoint that aggregates project health metrics.
- Return typed errors to the frontend SDK.
- Add integration tests for handler behavior.

**Completion checklist**

- [ ] Server boundaries validate external input.
- [ ] Config errors fail fast on startup.
- [ ] The BFF adds value instead of duplicating the core API blindly.

### Module 13: API contract generation and versioning

**Level:** Advanced

**Concepts to learn**

- OpenAPI or similar contract descriptions
- Generated clients versus hand-written clients
- Backward-compatible type changes
- Deprecation strategy

**Project build**

- Document the TaskForge API contract and generate or verify TypeScript types from it.
- Version one endpoint without breaking old clients.
- Add contract tests between SDK and API.

**Practice tasks**

- Compare hand-written and generated SDK code.
- Add one optional field and one breaking field change; document the difference.
- Create a changelog entry for API consumers.

**Completion checklist**

- [ ] API changes have versioning rules.
- [ ] Generated code is isolated from hand-written domain logic.
- [ ] The SDK and API agree on request and response shapes.

### Module 14: Large JavaScript to TypeScript migration strategy

**Level:** Advanced

**Concepts to learn**

- Gradual migration
- `allowJs`, JSDoc types, and strictness ramps
- Boundary-first migration
- De-risking large refactors

**Project build**

- Plan migration of the JavaScript TaskForge app to TypeScript.
- Migrate domain logic first, then adapters, then UI modules.
- Track strictness improvements over time.

**Practice tasks**

- Migrate one JS module without changing behavior.
- Add types around an unstable external boundary.
- Write a migration decision record.

**Completion checklist**

- [ ] Migration can happen incrementally.
- [ ] Tests protect behavior during type changes.
- [ ] The plan explains risk, order, and rollback options.

### Module 15: Library authoring, declaration files, and publishing readiness

**Level:** Advanced

**Concepts to learn**

- Public API design
- Declaration files and emitted types
- Package exports
- Semantic versioning for libraries

**Project build**

- Prepare `@taskforge/sdk` as a publishable TypeScript package.
- Emit declaration files and source maps.
- Document public methods with examples.

**Practice tasks**

- Create a tiny sample consumer project.
- Test package installation from a local tarball.
- Mark internal APIs so they do not appear in the public docs.

**Completion checklist**

- [ ] Consumers get useful editor autocomplete.
- [ ] The package has a stable public API surface.
- [ ] Breaking changes are documented.

### Module 16: Production-quality type design and maintainability

**Level:** Advanced

**Concepts to learn**

- Balancing type precision and readability
- Avoiding clever type traps
- Documenting advanced types
- Refactoring type-heavy code

**Project build**

- Audit TaskForge types for duplication, over-complexity, and confusing names.
- Create a type design guide for the project.
- Refactor advanced types into a small, documented set.

**Practice tasks**

- Replace one clever type with a clearer explicit type.
- Create examples for each exported advanced utility type.
- Add lint rules or conventions that prevent accidental `any`.

**Completion checklist**

- [ ] Advanced types help users rather than impressing maintainers.
- [ ] Every exported type has a purpose.
- [ ] The project is easier to understand after the audit.

### Module 17: TypeScript capstone: typed dashboard, SDK, and shared contracts

**Level:** Capstone

**Concepts to learn**

- End-to-end type safety
- Runtime validation at boundaries
- Typed workflows, SDK design, testing, migration, and publishing
- Portfolio documentation

**Project build**

- Complete a TypeScript admin dashboard and SDK for TaskForge.
- Use shared contracts, typed API clients, type-safe forms, typed state machines, and contract tests.
- Document how the TS layer talks to the Go API and JavaScript app.

**Practice tasks**

- Run an API contract change through the full stack.
- Create a migration guide from the JavaScript app to TypeScript.
- Prepare a package README for the SDK.

**Completion checklist**

- [ ] The dashboard and SDK compile with strict settings.
- [ ] External data is validated at runtime.
- [ ] The capstone shows type safety as product reliability, not just syntax.

## TypeScript capstone specification

Build `packages/shared-contracts`, `packages/ts-sdk`, and `apps/ts-admin-dashboard`.

Minimum features:

- Strict TypeScript configuration.
- Shared domain types for users, projects, tasks, comments, events, errors, and API responses.
- Runtime validation for imported JSON, form input, and API responses.
- Typed SDK with no `any` in public method return values.
- Admin dashboard with typed routes, forms, state machines, and error displays.
- Contract tests or generated type checks that keep API and SDK aligned.
- Migration plan from the JavaScript app.

Stretch features:

- Generate SDK types from API contracts.
- Publish the SDK locally or to a private registry.
- Add type-level tests for important utility types.
- Add a typed plugin system for custom task views.

## Final review questions

- Which bugs did TypeScript prevent that JavaScript would have allowed?
- Where do runtime checks still matter even when static types are correct?
- Which advanced types are worth keeping, and which ones make the project harder to maintain?
- How would you introduce TypeScript into an existing untyped codebase without stopping feature work?
- What is the public API of your SDK, and how will it evolve without breaking consumers?
