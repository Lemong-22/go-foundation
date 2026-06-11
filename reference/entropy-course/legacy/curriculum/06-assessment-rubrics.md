# Assessment Rubrics and Review Checklists

Use these rubrics at the end of every module and capstone. Score each area from 1 to 4.

```text
1 = incomplete or fragile
2 = works with major gaps
3 = solid and explainable
4 = production-minded and well documented
```

## Module rubric

| Area                    | 1                      | 2                        | 3                                        | 4                                               |
| ----------------------- | ---------------------- | ------------------------ | ---------------------------------------- | ----------------------------------------------- |
| Concept understanding   | Can copy examples only | Can explain basic syntax | Can apply concept in new feature         | Can explain tradeoffs and edge cases            |
| Feature completeness    | Does not run           | Runs only on happy path  | Handles common success and failure paths | Handles realistic edge cases with documentation |
| Code quality            | Tangled or unclear     | Some structure           | Clear names and responsibilities         | Easy to extend, review, and test                |
| Testing or verification | None                   | Manual only              | Important paths tested or documented     | Automated tests plus failure scenario notes     |
| Documentation           | Missing                | Basic run command        | Setup, usage, and design notes           | Includes tradeoffs, limitations, and next steps |

## JavaScript capstone rubric

- Browser workflows: create, edit, complete, filter, persist, import, export, and sync tasks.
- Async behavior: loading, failure, retry, cancellation, and no frozen UI during large work.
- Architecture: domain logic, UI, storage, HTTP, and configuration are separated.
- Accessibility: forms have labels, errors are clear, and keyboard navigation works.
- Security: user input is rendered safely and server validation exists for API paths.
- Performance: large data behavior is measured and documented.

## TypeScript capstone rubric

- Strict configuration is enabled and respected.
- Shared domain contracts are clear and not over-engineered.
- No public API returns `any` without a written reason.
- Runtime validation exists at every external data boundary.
- Typed workflows prevent impossible states.
- SDK, dashboard, and API contract stay aligned through tests or generation.
- Advanced types are documented and maintainable.

## Go capstone rubric

- API endpoints are consistent, tested, and documented.
- Domain validation and error mapping are explicit.
- Database migrations are repeatable.
- Context cancellation and graceful shutdown are handled.
- Workers process events with clear concurrency ownership.
- Race detector, tests, formatting, and vetting pass.
- Logs, metrics, health checks, and runbooks help diagnose failures.
- Deployment and rollback steps are documented.

## Rust capstone rubric

- Ownership and borrowing choices are clear and minimize unnecessary cloning.
- Invalid input returns useful `Result` errors rather than panics.
- CLI commands are reliable, tested, and documented.
- File writes fail safely and avoid data loss.
- Async sync handles timeouts, retries, and conflicts.
- Performance claims are measured.
- Cargo fmt, Clippy, tests, and release builds pass.
- Public APIs are small, safe, and documented.

## Code review checklist

Use this before merging a module into your main branch.

### Correctness

- Does the feature meet the stated acceptance criteria?
- Are invalid inputs handled?
- Are edge cases covered by tests or verification notes?
- Does the implementation avoid hidden global state?

### Maintainability

- Are names clear?
- Are modules small and cohesive?
- Is duplication intentional or should it be refactored?
- Are abstractions introduced because a real consumer needs them?

### Reliability

- What happens when network, storage, database, or file operations fail?
- Are timeouts, cancellation, and retries used where appropriate?
- Are logs useful and safe?
- Are destructive operations reversible or confirmed?

### Security and privacy

- Is external input validated?
- Is output encoded or rendered safely?
- Are authorization checks enforced on the server or trusted boundary?
- Are secrets and private user data kept out of logs?

### Performance

- Was performance measured before optimizing?
- Does the feature behave acceptably with realistic data sizes?
- Is memory use or allocation important for this feature?
- Does the optimization make the code harder to understand?

### Documentation

- Can a new developer run the feature from a clean clone?
- Does the README explain the user workflow?
- Are important tradeoffs captured in an ADR or notes file?
- Are limitations and future improvements listed?

## Final portfolio checklist

- Include screenshots or terminal recordings.
- Include a short architecture diagram.
- Include setup instructions for each language component.
- Include test commands and expected results.
- Include a demo script.
- Include a section called `What I learned` for each language.
- Include a section called `Tradeoffs and next steps`.
- Tag a final release.
