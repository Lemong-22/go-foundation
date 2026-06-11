# Learner-Safe Read Projections

Phase L2 uses an explicit `?view=learner` query mode on the existing detail read
endpoints:

- `GET /v1/quizzes/{quizId}?view=learner`
- `GET /v1/practices/{practiceId}?view=learner`
- `GET /v1/tests/{testId}?view=learner`

Omitting `view`, or passing `?view=instructor`, preserves the current
full-fidelity instructor responses. Unknown `view` values return validation
errors instead of silently choosing a response shape.

Projection stays in the REST adapter. The core usecases and domain DTOs remain
unchanged, and wire casing stays PascalCase to match the existing REST API.

Learner reads require the owning course to be `published`. Draft or unpublished
courses return `404 Not Found` on learner reads so the learner path does not
disclose draft content existence. Instructor/default reads remain full-fidelity
and can still return draft content behind the existing API token model.

## Allowed Learner Fields

Quiz learner detail:

- `Quiz.ID`, `Quiz.CourseID`, `Quiz.Title`, `Quiz.PassThreshold`,
  `Quiz.QuestionCount`, `Quiz.CreatedAt`, `Quiz.UpdatedAt`
- `Quiz.Questions[].ID`, `Quiz.Questions[].QuizID`,
  `Quiz.Questions[].Type`, `Quiz.Questions[].Prompt`,
  `Quiz.Questions[].Options`, `Quiz.Questions[].Position`

Practice learner detail:

- `Practice.ID`, `Practice.CourseID`, `Practice.Title`, `Practice.Language`,
  `Practice.Prompt`, `Practice.StarterCode`, `Practice.CreatedAt`,
  `Practice.UpdatedAt`

Test learner detail:

- `Test.ID`, `Test.CourseID`, `Test.Title`, `Test.TimeLimitMinutes`,
  `Test.PassThreshold`, `Test.ItemCount`, `Test.CreatedAt`, `Test.UpdatedAt`
- `Test.Items[].ID`, `Test.Items[].TestID`, `Test.Items[].Kind`,
  `Test.Items[].Position`
- Choice item fields: `ChoicePrompt`, `ChoiceType`, `ChoiceOptions`
- Coding item fields: `CodingPrompt`, `Language`, `StarterCode`

## Omitted Answer-Bearing Fields

Learner quiz reads omit question `CorrectIndices` and pre-submit
`Explanation`.

Learner practice reads omit `Solution`, `TestCases`, and test case
`ExpectedStdout`. The current backend has no learner-visible test case marker,
so the safe contract exposes no practice test cases yet.

Learner test reads omit `TestSolution`, choice `ChoiceCorrectIndices`, choice
`ChoiceExplanation`, coding `CodingSolution`, coding `TestCases`, and coding
test case `ExpectedStdout`.

## Web BFF Handoff

The Next.js learner reader should replace L1 dummy activity data through the
server-side course API client, not from browser-side fetches:

- Quiz blocks use `getLearnerQuiz(quizId)` from `web/src/lib/course-api/server`.
- Practice blocks use `getLearnerPractice(practiceId)` from
  `web/src/lib/course-api/server`.
- The course-level test view uses `getLearnerTest(testId)` from
  `web/src/lib/course-api/server`.

Those BFF methods call the REST endpoints above with `?view=learner` and keep
the API token server-only. The reader components should continue treating
omitted answer-bearing fields as unavailable until a future learner-attempt
context supplies post-submit explanations, grading, or solutions.

Regression coverage lives in two places:

- `internal/course/adapter/rest/server_test.go` serializes learner quiz,
  practice, and test responses and fails if answer-bearing field names reappear.
- `web/src/lib/course-api/transport.test.ts` fails if the BFF learner handoff
  methods accidentally call full-fidelity instructor URLs instead of
  `?view=learner` URLs.
