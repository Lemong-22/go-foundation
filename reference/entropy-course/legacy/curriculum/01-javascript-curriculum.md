# JavaScript Curriculum: Beginner to Advanced

## Track goal

Build a production-style **TaskForge browser application** using modern JavaScript. You start with plain objects and console output, then grow the project into an accessible, persistent, asynchronous, modular browser app backed by a small Node.js API.

## What you will be able to do

- Explain JavaScript values, objects, functions, closures, modules, and asynchronous behavior.
- Build browser interfaces with DOM events, forms, rendering, storage, and URL state.
- Use promises, `async`/`await`, `fetch`, retries, cancellation, and error handling.
- Organize a maintainable JavaScript codebase with modules, tests, adapters, and release notes.
- Build and present a complete JavaScript portfolio project.

## Primary references used

- [JavaScript Guide - MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide)
- [Asynchronous JavaScript - MDN](https://developer.mozilla.org/en-US/docs/Learn_web_development/Extensions/Async_JS)
- [Node.js Learn](https://nodejs.org/en/learn)

## Project milestone map

| Stage | Product feature                | Concepts practiced                            | Evidence of completion                              |
| ----- | ------------------------------ | --------------------------------------------- | --------------------------------------------------- |
| 1     | In-memory task model           | variables, objects, arrays, functions         | Console output and sample fixtures                  |
| 2     | Interactive browser task board | DOM, events, forms, validation                | Create, edit, complete, and delete tasks in browser |
| 3     | Persistent local app           | JSON, localStorage, errors, import/export     | Refresh-safe data and recovery from bad imports     |
| 4     | Modular app architecture       | ES modules, npm scripts, adapters             | Feature folders and clean module boundaries         |
| 5     | Remote sync                    | promises, async/await, fetch, retries         | Browser app talks to mock or Node API               |
| 6     | Production polish              | testing, accessibility, security, performance | Capstone demo with tests and deployment notes       |

## Curriculum modules

Each module is designed as a small product sprint. Do not move on until the deliverable works, has tests or manual verification notes, and is committed with a clear README update.

### Module 1: Runtime mental model, setup, and first data model

**Level:** Beginner

**Concepts to learn**

- JavaScript execution in browsers and Node.js
- Values, variables, primitive types, objects, arrays, and equality
- Console debugging and basic error reading
- Project setup with a simple folder structure

**Project build**

- Create TaskForge as a local project with an `index.html`, `main.js`, and `README.md`.
- Represent users, projects, and tasks as plain objects and arrays.
- Print useful summaries in the browser console and Node.js console.

**Practice tasks**

- Write small examples that compare `let`, `const`, and accidental global variables.
- Create five sample tasks and group them by status.
- Add a README section explaining the TaskForge domain in plain English.

**Completion checklist**

- [ ] You can explain the difference between primitives and object references.
- [ ] Sample task data is readable and easy to modify.
- [ ] The project can run in a browser and with Node.js for simple scripts.

### Module 2: Functions, scope, closures, and validation

**Level:** Beginner

**Concepts to learn**

- Function declarations, function expressions, arrow functions
- Parameters, return values, default values, and rest parameters
- Scope, closures, and pure versus impure functions
- Basic validation and guard clauses

**Project build**

- Create task factory functions such as `createTask`, `completeTask`, and `renameTask`.
- Add validation for required titles, allowed statuses, and due dates.
- Use closures to create an ID generator for tasks and projects.

**Practice tasks**

- Write three implementations of a task filter: classic function, arrow function, and reusable predicate.
- Refactor duplicated validation into helper functions.
- Add error messages that are useful to a user, not only to a developer.

**Completion checklist**

- [ ] Task creation rejects invalid input.
- [ ] Functions are small and named by intent.
- [ ] You can explain how the ID generator remembers state.

### Module 3: Objects, arrays, iteration, and transformations

**Level:** Beginner

**Concepts to learn**

- Object property access, destructuring, spreading, and copying
- Array methods: `map`, `filter`, `find`, `some`, `every`, `reduce`, and `sort`
- Mutation versus immutable updates
- Date and string handling basics

**Project build**

- Build a task query module that filters by status, owner, project, search text, and due date.
- Create derived views: overdue tasks, upcoming tasks, blocked tasks, and completed count.
- Implement sorting by priority and due date.

**Practice tasks**

- Write the same feature once mutating arrays and once using immutable updates, then compare tradeoffs.
- Create fixtures for sample projects and tasks.
- Add edge cases: empty titles, missing due dates, duplicate task names.

**Completion checklist**

- [ ] Filtering and sorting functions are deterministic.
- [ ] Data transformations do not accidentally corrupt the original fixtures.
- [ ] You can explain when mutation is acceptable and when it is risky.

### Module 4: DOM, events, and forms

**Level:** Beginner

**Concepts to learn**

- DOM tree selection and element creation
- Event listeners, event delegation, and event objects
- Form handling and input validation
- Rendering state to HTML

**Project build**

- Create a browser UI that lists tasks and projects.
- Add forms to create, edit, complete, and delete tasks.
- Render validation errors near the relevant input.

**Practice tasks**

- Use event delegation for task list actions.
- Render empty, loading, success, and error states manually.
- Add keyboard-friendly interactions for forms and buttons.

**Completion checklist**

- [ ] The UI can be used without touching the console.
- [ ] Repeated renders do not duplicate event listeners.
- [ ] Invalid form input is handled gracefully.

### Module 5: JSON, browser storage, and import/export

**Level:** Beginner

**Concepts to learn**

- JSON serialization and parsing
- `localStorage` and browser persistence limits
- Error handling with `try` and `catch`
- Data migration basics

**Project build**

- Persist TaskForge data to local storage.
- Add export-to-JSON and import-from-JSON features.
- Add a simple schema version field and a migration function.

**Practice tasks**

- Corrupt the stored JSON and verify that the app recovers safely.
- Write a sample import file by hand.
- Add a reset-data button with confirmation.

**Completion checklist**

- [ ] Refreshing the browser keeps data.
- [ ] Invalid imports show clear errors.
- [ ] You can explain why storage needs versioning.

### Module 6: Modules, npm scripts, and project organization

**Level:** Intermediate

**Concepts to learn**

- ES modules: `import` and `export`
- Package scripts and dependency management
- Feature-based folder structure
- Separation of domain logic, UI logic, and infrastructure

**Project build**

- Split TaskForge into modules: domain, storage, rendering, events, and app startup.
- Add npm scripts for development, formatting, tests, and linting.
- Introduce a simple build step only after the unbundled version works.

**Practice tasks**

- Create an architecture decision record explaining the folder structure.
- Move all DOM-independent logic into pure modules.
- Add import boundaries so UI code depends on domain code, not the reverse.

**Completion checklist**

- [ ] The project has no giant `main.js` file.
- [ ] Modules have clear responsibilities.
- [ ] A new contributor can find where to add a feature.

### Module 7: Promises, async/await, fetch, and remote data

**Level:** Intermediate

**Concepts to learn**

- Callbacks, promises, promise states, and `async`/`await`
- `fetch`, HTTP methods, status codes, and JSON APIs
- Sequential versus parallel asynchronous work
- Timeouts, retries, and cancellation with `AbortController`

**Project build**

- Create a mock TaskForge API using a simple Node.js server or static JSON endpoints.
- Replace local-only loading with async data fetching.
- Add sync status: idle, loading, saving, failed, and retrying.

**Practice tasks**

- Fetch projects and tasks in parallel, then render them together.
- Simulate server failures and slow responses.
- Add retry with exponential backoff for safe read operations.

**Completion checklist**

- [ ] The UI never freezes while waiting for data.
- [ ] Network errors are visible and recoverable.
- [ ] You can explain why `async` functions return promises.

### Module 8: Error handling, logging, and user feedback

**Level:** Intermediate

**Concepts to learn**

- Throwing and catching errors
- Custom error classes
- Error boundaries in vanilla application architecture
- Logging without leaking sensitive data

**Project build**

- Create error types for validation errors, network errors, storage errors, and authorization errors.
- Add user-facing notifications and developer-facing logs.
- Create a support bundle that exports app version, browser info, and recent errors without task contents.

**Practice tasks**

- Map technical errors to friendly messages.
- Test failure paths manually and with unit tests.
- Add a debug mode controlled by configuration.

**Completion checklist**

- [ ] Errors do not crash the whole application.
- [ ] Logs help diagnose issues without exposing private data.
- [ ] Every async path has a defined failure behavior.

### Module 9: Classes, prototypes, repositories, and event bus patterns

**Level:** Intermediate

**Concepts to learn**

- Prototype chain and class syntax
- Encapsulation and constructor invariants
- Repository pattern for data access
- Pub/sub event bus basics

**Project build**

- Create a `TaskRepository` abstraction with local and remote implementations.
- Create an `EventBus` for task-created, task-updated, sync-started, and sync-failed events.
- Use classes only where they simplify stateful behavior.

**Practice tasks**

- Implement the same repository as a closure-based module and as a class.
- Compare testability of both approaches.
- Add event subscribers for notifications and analytics.

**Completion checklist**

- [ ] The repository hides storage details from the UI.
- [ ] Events have predictable names and payloads.
- [ ] You can explain prototype-based inheritance at a high level.

### Module 10: Functional patterns and predictable state updates

**Level:** Intermediate

**Concepts to learn**

- Higher-order functions and composition
- Reducers and action objects
- Derived selectors
- Avoiding hidden shared state

**Project build**

- Create a small state store for TaskForge using reducer-style updates.
- Add actions for create, update, complete, archive, assign, and sync.
- Create selectors for dashboard metrics.

**Practice tasks**

- Write before-and-after examples of tangled state mutations versus reducers.
- Add undo for the last task operation.
- Keep reducer functions free of network and DOM work.

**Completion checklist**

- [ ] State transitions are easy to test.
- [ ] Undo works for supported operations.
- [ ] You can trace a UI action from event to state update to render.

### Module 11: Testing JavaScript applications

**Level:** Intermediate

**Concepts to learn**

- Unit, integration, and end-to-end testing levels
- Test fixtures and test doubles
- Node.js test runner or a common JS test framework
- Testing asynchronous code

**Project build**

- Add tests for validators, reducers, repositories, and API clients.
- Create fake storage and fake HTTP adapters.
- Add a test script to CI or a local pre-merge checklist.

**Practice tasks**

- Write one failing test before a bug fix.
- Test one async retry flow.
- Measure which code is worth testing and which is better covered by manual checks.

**Completion checklist**

- [ ] Core domain logic has repeatable tests.
- [ ] Async tests do not rely on random timing.
- [ ] The README explains how to run tests.

### Module 12: Accessibility, browser APIs, and progressive enhancement

**Level:** Intermediate

**Concepts to learn**

- Semantic HTML, focus management, ARIA only when needed
- Keyboard navigation and screen-reader-friendly status messages
- URL state, History API, and routing basics
- Progressive enhancement mindset

**Project build**

- Add accessible task forms, filter controls, and status messages.
- Persist filters in the URL so links can open a specific view.
- Add basic offline messaging when the network is unavailable.

**Practice tasks**

- Navigate the app with keyboard only.
- Add labels and error descriptions to every form field.
- Share a filtered URL and verify it restores state.

**Completion checklist**

- [ ] Core workflows are keyboard usable.
- [ ] The URL represents meaningful application state.
- [ ] The app remains useful when one enhancement fails.

### Module 13: Node.js HTTP service fundamentals

**Level:** Advanced

**Concepts to learn**

- Node.js as a runtime for servers, command-line tools, and scripts
- HTTP request and response lifecycle
- Routing, middleware concepts, and status codes
- Environment variables and configuration

**Project build**

- Build a minimal TaskForge API in Node.js for projects and tasks.
- Serve JSON endpoints for list, create, update, complete, and delete.
- Add request logging and structured error responses.

**Practice tasks**

- Implement the API once with the built-in `http` module, then optionally with a small framework.
- Document endpoints in `docs/api.md`.
- Add integration tests that call the local server.

**Completion checklist**

- [ ] The browser app can switch from local storage to the Node API.
- [ ] API errors use consistent shapes.
- [ ] You can explain the request lifecycle from socket to response.

### Module 14: Event loop, streams, workers, and large data processing

**Level:** Advanced

**Concepts to learn**

- Event loop phases at a practical level
- Non-blocking I/O and CPU-bound work
- Readable and writable streams
- Web workers or Node worker threads

**Project build**

- Add bulk import for large TaskForge JSON or CSV files.
- Use streaming or chunking so the UI/server remains responsive.
- Move expensive validation or parsing off the main browser thread.

**Practice tasks**

- Compare a blocking import with a chunked import.
- Measure responsiveness during a 10,000 task import.
- Add progress events and cancellation.

**Completion checklist**

- [ ] Large imports do not freeze the app.
- [ ] Progress and cancellation are visible to the user.
- [ ] You can explain why CPU-heavy work blocks JavaScript.

### Module 15: Security and data protection basics

**Level:** Advanced

**Concepts to learn**

- Cross-site scripting risks and output encoding
- CORS, CSRF, and cookie versus token tradeoffs
- Input validation and trust boundaries
- Secrets, environment variables, and safe logging

**Project build**

- Audit TaskForge for unsafe HTML rendering.
- Add server-side validation for API input.
- Add authorization checks to prevent editing another user's project.

**Practice tasks**

- Try to inject HTML in a task title and verify it is rendered safely.
- Create a security checklist for new endpoints.
- Remove sensitive values from logs and support bundles.

**Completion checklist**

- [ ] User input is not trusted anywhere.
- [ ] Auth and ownership checks exist on the server, not only in the UI.
- [ ] The README documents security assumptions.

### Module 16: Performance profiling and rendering strategy

**Level:** Advanced

**Concepts to learn**

- Measuring before optimizing
- Rendering cost, layout thrashing, and batching updates
- Memoization and caching tradeoffs
- Bundle size and dependency review

**Project build**

- Profile the dashboard with many projects and tasks.
- Optimize task list rendering with batching, pagination, or virtualization ideas.
- Add a performance budget and dependency review checklist.

**Practice tasks**

- Record baseline timings before each optimization.
- Remove or replace one unnecessary dependency.
- Write a short performance report.

**Completion checklist**

- [ ] Performance changes are backed by measurements.
- [ ] The app handles large realistic data sets.
- [ ] You can explain what changed and why it helped.

### Module 17: Application architecture and maintainability

**Level:** Advanced

**Concepts to learn**

- Layered architecture and dependency direction
- Feature modules and public APIs
- Configuration, dependency injection, and adapters
- Refactoring safely

**Project build**

- Refactor TaskForge into domain, application, infrastructure, and UI layers.
- Create adapters for storage, HTTP, analytics, and notifications.
- Add architecture decision records for major choices.

**Practice tasks**

- Draw a dependency diagram.
- Replace the storage adapter without changing UI code.
- Identify three code smells and refactor them.

**Completion checklist**

- [ ] Business logic is not coupled to DOM or HTTP details.
- [ ] Adapters are swappable.
- [ ] A future TypeScript migration has clear boundaries.

### Module 18: Build, deployment, diagnostics, and release workflow

**Level:** Advanced

**Concepts to learn**

- Development versus production builds
- Source maps and diagnostics
- Environment-specific configuration
- Semantic versioning and changelogs

**Project build**

- Create a deployable static JS app and a deployable Node API.
- Add a release checklist, changelog, and rollback notes.
- Add lightweight diagnostics for startup configuration and API health.

**Practice tasks**

- Deploy to any static host and any server/runtime platform.
- Tag a release and write release notes.
- Break configuration intentionally and verify diagnostics help.

**Completion checklist**

- [ ] The project can be built from a clean clone.
- [ ] Release steps are documented.
- [ ] Diagnostics reveal common setup mistakes.

### Module 19: JavaScript capstone: TaskForge browser app plus Node API

**Level:** Capstone

**Concepts to learn**

- End-to-end feature delivery
- Architecture, testing, async workflows, performance, and security review
- Documentation and portfolio presentation
- Migration path to TypeScript

**Project build**

- Complete a vanilla JavaScript TaskForge app with local persistence, remote sync, filters, accessible forms, bulk import, and a simple Node API.
- Create screenshots, demo data, and a short technical writeup.
- Prepare a backlog for the TypeScript version.

**Practice tasks**

- Run a self-review using `06-assessment-rubrics.md`.
- Ask someone else to install and run the project from the README.
- Record the main tradeoffs and what you would improve next.

**Completion checklist**

- [ ] A user can create, update, filter, persist, import, export, and sync tasks.
- [ ] The codebase is modular enough to migrate one module at a time to TypeScript.
- [ ] The portfolio README explains business value and engineering choices.

## JavaScript capstone specification

Build `apps/js-browser` and `services/js-api`.

Minimum features:

- Project and task CRUD.
- Local persistence and JSON import/export.
- Search, filter, sort, and dashboard summary views.
- Async sync with a Node.js JSON API.
- Loading, empty, success, and failure UI states.
- Accessible forms and keyboard navigation.
- Unit or integration tests for domain logic and API client behavior.
- Performance notes for large task lists.
- Security notes covering safe rendering, validation, and logging.

Stretch features:

- Offline-first queue for pending changes.
- Web worker for large imports.
- Service worker cache for static assets.
- Audit log of task changes.
- Migration plan to TypeScript.

## Final review questions

- Where does state live, and how does it move through the app?
- Which modules are pure domain logic and which modules touch browser or network APIs?
- What happens when storage is corrupted, the network fails, or the API returns invalid data?
- How would you migrate one file at a time to TypeScript?
- Which performance issue did you measure rather than guess?
