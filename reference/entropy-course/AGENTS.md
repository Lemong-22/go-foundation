# Agent Instructions

## Course Import Workflow

The course import agent workflow is external to this Go service. Agents should
use the repository only through the documented import package and service
entrypoints.

Reference docs:

- Zip format, validation rules, and file schemas:
  [docs/import-format.md](docs/import-format.md)
- Plan JSON schema and apply result shape:
  [docs/import-format.md#plan-json-schema](docs/import-format.md#plan-json-schema)

Use this loop: normalize -> plan -> resolve -> apply.

1. Normalize source material into a v1 course zip.
2. Run `PlanImport` through CLI, REST, or console.
3. Resolve conflicts by choosing existing candidates, or change the zip and
   replan when content or identity must change.
4. Run `ApplyPlan` with the resolved plan or an explicit conflict strategy.

CLI examples:

```sh
course-cli import plan course.zip -o json --output plan.json
course-cli import apply course.zip --resolved-plan resolved-plan.json --force
course-cli import apply course.zip --conflict-strategy fail --force
```

REST examples:

```text
POST /v1/import/plan
  multipart field: zip

POST /v1/import/apply?conflict_strategy=fail
  multipart fields: zip, resolved_plan (optional)
```

Console flow:

- Upload a course zip.
- Review the plan operations and conflicts.
- Build a resolved plan in the browser.
- Apply through the same import service used by CLI and REST.

Conflict guidance:

- Prefer the planner recommendation when the imported entity is clearly the same
  course, lesson, quiz, practice, test, block, question, test case, or test item.
- Use `skip` when applying the imported payload would overwrite intentional
  existing content.
- To create a distinct new entity after a collision, change the zip identity
  field, such as course slug, title, or child position, then run planning again.
- Do not edit `payload` values inside plan JSON. Payload changes require editing
  the zip and replanning so the `zip_hash` invariant remains meaningful.
- Keep `conflicts` empty in a resolved plan submitted to apply.

Out of scope for agents in this repository:

- Running an in-process AI agent runtime.
- Adding import history, rollback, audit persistence, or new import tables.
- Handling multi-course zip packages.
- Validating that imported source code compiles or runs.
- Adding domain slugs to Quiz, Practice, or Test.
