import { describe, expect, test } from "bun:test";

import { createCourseApiClient } from "./transport";
import type { CourseApiFetch, CourseApiSettings } from "./transport";
import type {
  CourseApiResult,
  GetLearnerPracticeOutput,
  GetLearnerQuizOutput,
  GetLearnerTestOutput,
  ListCoursesOutput,
  ListLessonBlocksOutput,
} from "./types";

type FetchCall = {
  input: Parameters<typeof fetch>[0];
  init: Parameters<typeof fetch>[1];
};

function createFetchStub(responseFactory: () => Response) {
  const calls: FetchCall[] = [];
  const fetcher: CourseApiFetch = async (input, init) => {
    calls.push({ input, init });
    return responseFactory();
  };

  return { calls, fetcher };
}

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "content-type": "application/json",
    },
    ...init,
  });
}

function withSettings(
  settings: CourseApiSettings,
): () => CourseApiResult<CourseApiSettings> {
  return () => ({
    data: settings,
    ok: true,
  });
}

describe("course API transport", () => {
  test("does not fetch when settings are unavailable", async () => {
    const { calls, fetcher } = createFetchStub(() => jsonResponse({}));
    const client = createCourseApiClient({
      fetcher,
      settings: () => ({
        error: {
          code: "configuration",
          message: "Server course API settings are unavailable.",
        },
        ok: false,
      }),
    });

    const result = await client.listPublishedCourses();

    expect(calls).toHaveLength(0);
    expect(result.ok).toBe(false);
    if (result.ok) {
      throw new Error("expected configuration error");
    }
    expect(result.error.code).toBe("configuration");
  });

  test("lists published courses with bearer auth and PascalCase DTO fields", async () => {
    const response: ListCoursesOutput = {
      Courses: [
        {
          CreatedAt: "2026-06-01T00:00:00Z",
          Description: "Build a JavaScript foundation.",
          ID: "course-1",
          InstructorID: "instructor-1",
          Slug: "javascript",
          Status: "published",
          Title: "JavaScript",
          UpdatedAt: "2026-06-02T00:00:00Z",
        },
      ],
    };
    const { calls, fetcher } = createFetchStub(() => jsonResponse(response));
    const client = createCourseApiClient({
      fetcher,
      settings: withSettings({
        baseUrl: "http://api.example.test",
        token: "test-token",
      }),
    });

    const result = await client.listPublishedCourses();

    expect(calls).toHaveLength(1);
    expect(String(calls[0].input)).toBe(
      "http://api.example.test/v1/courses?status=published",
    );
    expect(calls[0].init?.cache).toBe("force-cache");
    expect(calls[0].init?.next).toEqual({
      revalidate: false,
      tags: ["course-reader:catalog:published"],
    });
    expect(new Headers(calls[0].init?.headers).get("Authorization")).toBe(
      "Bearer test-token",
    );
    expect(result.ok).toBe(true);
    if (!result.ok) {
      throw new Error(result.error.message);
    }
    expect(result.data.Courses[0].ID).toBe("course-1");
    expect(result.data.Courses[0].Title).toBe("JavaScript");
  });

  test("encodes course IDs for course detail reads", async () => {
    const { calls, fetcher } = createFetchStub(() =>
      jsonResponse({
        Course: {
          CreatedAt: "2026-06-01T00:00:00Z",
          Description: "Description",
          ID: "course/1",
          InstructorID: "instructor-1",
          Slug: "course-1",
          Status: "published",
          Title: "Course 1",
          UpdatedAt: "2026-06-02T00:00:00Z",
        },
      }),
    );
    const client = createCourseApiClient({
      fetcher,
      settings: withSettings({
        baseUrl: "http://api.example.test/",
        token: "test-token",
      }),
    });

    const result = await client.getCourse("course/1");

    expect(result.ok).toBe(true);
    expect(String(calls[0].input)).toBe(
      "http://api.example.test/v1/courses/course%2F1",
    );
    expect(calls[0].init?.next).toEqual({
      revalidate: false,
      tags: ["course-reader:course:course%2F1"],
    });
  });

  test("lists lesson blocks without remapping Markdown", async () => {
    const response: ListLessonBlocksOutput = {
      Blocks: [
        {
          ID: "block-1",
          Kind: "text",
          LessonID: "lesson-1",
          Markdown: "# Values",
          Position: 0,
          PracticeRef: "",
          QuizRef: "",
          VideoCaption: "",
          VideoLocator: "",
          VideoProvider: "",
        },
      ],
    };
    const { calls, fetcher } = createFetchStub(() => jsonResponse(response));
    const client = createCourseApiClient({
      fetcher,
      settings: withSettings({
        token: "test-token",
      }),
    });

    const result = await client.listLessonBlocks("lesson-1");

    expect(result.ok).toBe(true);
    if (!result.ok) {
      throw new Error(result.error.message);
    }
    expect(result.data.Blocks[0].Markdown).toBe("# Values");
    expect(calls[0].init?.next).toEqual({
      revalidate: false,
      tags: [
        "course-reader:lesson:lesson-1",
        "course-reader:lesson-blocks:lesson-1",
      ],
    });
  });

  test("uses learner-safe aggregate reads for quiz practice and test handoff", async () => {
    const responses = [
      {
        Quiz: {
          CourseID: "course-1",
          CreatedAt: "2026-06-01T00:00:00Z",
          ID: "quiz/1",
          PassThreshold: 0.7,
          QuestionCount: 1,
          Questions: [
            {
              ID: "question-1",
              Options: ["A", "B"],
              Position: 0,
              Prompt: "Pick one",
              QuizID: "quiz/1",
              Type: "single",
            },
          ],
          Title: "Quiz",
          UpdatedAt: "2026-06-02T00:00:00Z",
        },
      } satisfies GetLearnerQuizOutput,
      {
        Practice: {
          CourseID: "course-1",
          CreatedAt: "2026-06-01T00:00:00Z",
          ID: "practice/1",
          Language: "golang",
          Prompt: "Write code",
          StarterCode: "package main",
          Title: "Practice",
          UpdatedAt: "2026-06-02T00:00:00Z",
        },
      } satisfies GetLearnerPracticeOutput,
      {
        Test: {
          CourseID: "course-1",
          CreatedAt: "2026-06-01T00:00:00Z",
          ID: "test/1",
          ItemCount: 1,
          Items: [
            {
              ChoiceOptions: ["A", "B"],
              ChoicePrompt: "Pick one",
              ChoiceType: "single",
              CodingPrompt: "",
              ID: "item-1",
              Kind: "choice",
              Language: "",
              Position: 0,
              StarterCode: "",
              TestID: "test/1",
            },
          ],
          PassThreshold: 0.7,
          TimeLimitMinutes: null,
          Title: "Test",
          UpdatedAt: "2026-06-02T00:00:00Z",
        },
      } satisfies GetLearnerTestOutput,
    ];
    let index = 0;
    const { calls, fetcher } = createFetchStub(() =>
      jsonResponse(responses[index++]),
    );
    const client = createCourseApiClient({
      fetcher,
      settings: withSettings({
        baseUrl: "http://api.example.test",
        token: "test-token",
      }),
    });

    const quiz = await client.getLearnerQuiz("quiz/1");
    const practice = await client.getLearnerPractice("practice/1");
    const testResult = await client.getLearnerTest("test/1");

    expect(quiz.ok).toBe(true);
    expect(practice.ok).toBe(true);
    expect(testResult.ok).toBe(true);
    expect(calls.map((call) => String(call.input))).toEqual([
      "http://api.example.test/v1/quizzes/quiz%2F1?view=learner",
      "http://api.example.test/v1/practices/practice%2F1?view=learner",
      "http://api.example.test/v1/tests/test%2F1?view=learner",
    ]);
    expect(calls.map((call) => call.init?.next)).toEqual([
      {
        revalidate: false,
        tags: ["course-reader:learner-quiz:quiz%2F1"],
      },
      {
        revalidate: false,
        tags: ["course-reader:learner-practice:practice%2F1"],
      },
      {
        revalidate: false,
        tags: ["course-reader:learner-test:test%2F1"],
      },
    ]);
  });

  test("maps not found responses to app-level error states", async () => {
    const { fetcher } = createFetchStub(() =>
      jsonResponse({ error: "course not found" }, { status: 404 }),
    );
    const client = createCourseApiClient({
      fetcher,
      settings: withSettings({
        baseUrl: "http://api.example.test",
        token: "test-token",
      }),
    });

    const result = await client.getCourse("missing-course");

    expect(result.ok).toBe(false);
    if (result.ok) {
      throw new Error("expected not-found error");
    }
    expect(result.error.code).toBe("not-found");
    expect(result.error.message).toBe("course not found");
    expect(result.error.status).toBe(404);
  });
});
