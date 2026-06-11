import { describe, expect, test } from "bun:test";
import { readdirSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { renderToStaticMarkup } from "react-dom/server";

import { CourseCatalog } from "@/components/course-catalog";
import { CourseDetail } from "@/components/course-detail";
import { CourseTest } from "@/components/course-test";
import { LessonReader } from "@/components/lesson-reader";
import { createCourseApiClient, type CourseApiFetch } from "@/lib/course-api/transport";
import type { BlockView, CourseView, LessonView } from "@/lib/course-api/types";
import { selectPublishedCourses } from "@/lib/catalog";
import { buildDummyCourseTest } from "@/lib/course-test";

const sourceRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const answerBearingTerms = ["correct", "solution", "expected", "hidden", "answer"];

const importedPublishedCourse: CourseView = {
  CreatedAt: "2026-06-01T00:00:00Z",
  Description: "Imported course material rendered by the learner reader.",
  ID: "course-imported-js",
  InstructorID: "instructor-1",
  Slug: "imported-javascript-foundations",
  Status: "published",
  Title: "Imported JavaScript Foundations",
  UpdatedAt: "2026-06-03T00:00:00Z",
};

const draftCourse: CourseView = {
  ...importedPublishedCourse,
  ID: "course-draft-js",
  Slug: "draft-javascript-foundations",
  Status: "draft",
  Title: "Draft JavaScript Foundations",
};

const importedLessons: LessonView[] = [
  {
    CourseID: importedPublishedCourse.ID,
    CreatedAt: "2026-06-01T00:00:00Z",
    ID: "lesson-values",
    Order: 1,
    Title: "Values and Types",
    UpdatedAt: "2026-06-03T00:00:00Z",
  },
];

const importedBlocks: BlockView[] = [
  {
    ID: "block-text",
    Kind: "text",
    LessonID: "lesson-values",
    Markdown:
      "## Values are data\n\nA `value` is the smallest thing a program can move around.",
    Position: 1,
    PracticeRef: "",
    QuizRef: "",
    VideoCaption: "",
    VideoLocator: "",
    VideoProvider: "",
  },
  {
    ID: "block-video",
    Kind: "video",
    LessonID: "lesson-values",
    Markdown: "",
    Position: 2,
    PracticeRef: "",
    QuizRef: "",
    VideoCaption: "Watch the concept before reading the examples.",
    VideoLocator: "abc123",
    VideoProvider: "youtube",
  },
  {
    ID: "block-quiz",
    Kind: "quiz",
    LessonID: "lesson-values",
    Markdown: "",
    Position: 3,
    PracticeRef: "",
    QuizRef: "values-check",
    VideoCaption: "",
    VideoLocator: "",
    VideoProvider: "",
  },
  {
    ID: "block-practice",
    Kind: "practice",
    LessonID: "lesson-values",
    Markdown: "",
    Position: 4,
    PracticeRef: "values-practice",
    QuizRef: "",
    VideoCaption: "",
    VideoLocator: "",
    VideoProvider: "",
  },
];

describe("Phase L1 exit acceptance", () => {
  test("renders a published imported course from catalog through lesson and sample test", async () => {
    const calls: RecordedFetchCall[] = [];
    const client = createCourseApiClient({
      fetcher: acceptanceFetcher(calls),
      settings: () => ({
        data: {
          baseUrl: "http://course-api.test",
          token: "server-only-token",
        },
        ok: true,
      }),
    });

    const catalogResult = await client.listPublishedCourses();
    const courseResult = await client.getCourse(importedPublishedCourse.ID);
    const lessonsResult = await client.listCourseLessons(importedPublishedCourse.ID);
    const lessonResult = await client.getLesson(importedLessons[0].ID);
    const blocksResult = await client.listLessonBlocks(importedLessons[0].ID);

    expect(catalogResult.ok).toBe(true);
    expect(courseResult.ok).toBe(true);
    expect(lessonsResult.ok).toBe(true);
    expect(lessonResult.ok).toBe(true);
    expect(blocksResult.ok).toBe(true);

    if (
      !catalogResult.ok ||
      !courseResult.ok ||
      !lessonsResult.ok ||
      !lessonResult.ok ||
      !blocksResult.ok
    ) {
      throw new Error("acceptance fixture failed to load");
    }

    const publishedCourses = selectPublishedCourses(catalogResult.data.Courses);
    const catalogHtml = renderToStaticMarkup(
      <CourseCatalog courses={publishedCourses} />,
    );
    const detailHtml = renderToStaticMarkup(
      <CourseDetail
        course={courseResult.data.Course}
        lessons={lessonsResult.data.Lessons}
      />,
    );
    const lessonHtml = renderToStaticMarkup(
      <LessonReader
        blocks={blocksResult.data.Blocks}
        course={courseResult.data.Course}
        lesson={lessonResult.data.Lesson}
      />,
    );
    const testHtml = renderToStaticMarkup(
      <CourseTest
        course={courseResult.data.Course}
        test={buildDummyCourseTest(courseResult.data.Course)}
      />,
    );
    const combinedHtml = [catalogHtml, detailHtml, lessonHtml, testHtml].join("\n");

    expect(catalogHtml).toContain("Imported JavaScript Foundations");
    expect(catalogHtml).not.toContain("Draft JavaScript Foundations");
    expect(catalogHtml).toContain("/courses/course-imported-js");
    expect(detailHtml).toContain("/lessons/lesson-values");
    expect(lessonHtml).toContain("Values are data");
    expect(lessonHtml).toContain("https://www.youtube-nocookie.com/embed/abc123");
    expect(lessonHtml).toContain("Watch the concept before reading the examples.");
    expect(countMatches(lessonHtml, "Sample data")).toBeGreaterThanOrEqual(2);
    expect(countMatches(testHtml, "Sample data")).toBeGreaterThanOrEqual(1);
    expect(countMatches(lessonHtml, "Placeholder learner content only")).toBe(2);
    expect(countMatches(testHtml, "Placeholder learner content only")).toBe(1);

    const visibleText = combinedHtml.replace(/<[^>]*>/g, " ").toLowerCase();
    for (const term of answerBearingTerms) {
      expect(visibleText).not.toContain(term);
    }

    expect(calls.map((call) => call.path)).toEqual([
      "/v1/courses?status=published",
      "/v1/courses/course-imported-js",
      "/v1/courses/course-imported-js/lessons",
      "/v1/lessons/lesson-values",
      "/v1/lessons/lesson-values/blocks",
    ]);
    expect(calls.every((call) => call.authorization === "Bearer server-only-token")).toBe(
      true,
    );
    expect(calls.some((call) => /\/v1\/(quizzes|practices|tests)\b/.test(call.path))).toBe(
      false,
    );
  });

  test("keeps API credentials and answer-bearing aggregate reads out of client-facing code", () => {
    const serverApi = readSourceFile("lib/course-api/server.ts");
    expect(serverApi).toContain('import "server-only"');
    expect(serverApi).toContain("COURSE_CLI_API_TOKEN");

    const clientFacingSource = readSourceFiles(["app", "components"])
      .filter((file) => !file.path.endsWith("app/api/revalidate/route.ts"))
      .map((file) => file.content)
      .join("\n");

    expect(clientFacingSource).not.toContain("COURSE_CLI_API_TOKEN");
    expect(clientFacingSource).not.toContain("COURSE_API_BASE_URL");
    expect(clientFacingSource).not.toContain("Authorization");
    expect(clientFacingSource).not.toContain("Bearer ");
    expect(clientFacingSource).not.toMatch(/\/v1\/(quizzes|practices|tests)\b/);
    expect(clientFacingSource).not.toMatch(/\b(list|get)(Quiz|Practice|Test)\b/);
  });
});

type RecordedFetchCall = {
  authorization: string | null;
  path: string;
};

function acceptanceFetcher(calls: RecordedFetchCall[]): CourseApiFetch {
  return async (input, init) => {
    const url = new URL(String(input));
    calls.push({
      authorization: new Headers(init?.headers).get("Authorization"),
      path: `${url.pathname}${url.search}`,
    });

    if (url.pathname === "/v1/courses") {
      return jsonResponse({
        Courses: [importedPublishedCourse, draftCourse],
      });
    }

    if (url.pathname === `/v1/courses/${importedPublishedCourse.ID}`) {
      return jsonResponse({
        Course: importedPublishedCourse,
      });
    }

    if (url.pathname === `/v1/courses/${importedPublishedCourse.ID}/lessons`) {
      return jsonResponse({
        Lessons: importedLessons,
      });
    }

    if (url.pathname === `/v1/lessons/${importedLessons[0].ID}`) {
      return jsonResponse({
        Lesson: importedLessons[0],
      });
    }

    if (url.pathname === `/v1/lessons/${importedLessons[0].ID}/blocks`) {
      return jsonResponse({
        Blocks: importedBlocks,
      });
    }

    return jsonResponse({ error: "not found" }, { status: 404 });
  };
}

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "content-type": "application/json",
    },
    status: init?.status ?? 200,
  });
}

function readSourceFiles(relativeDirectories: string[]) {
  return relativeDirectories.flatMap((relativeDirectory) =>
    sourceFiles(join(sourceRoot, relativeDirectory)),
  );
}

function sourceFiles(directory: string): Array<{ content: string; path: string }> {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      return sourceFiles(path);
    }

    if (!/\.(ts|tsx)$/.test(entry.name) || /\.test\.(ts|tsx)$/.test(entry.name)) {
      return [];
    }

    return [
      {
        content: readFileSync(path, "utf8"),
        path: path.slice(sourceRoot.length + 1),
      },
    ];
  });
}

function readSourceFile(relativePath: string) {
  return readFileSync(join(sourceRoot, relativePath), "utf8");
}

function countMatches(value: string, term: string) {
  return value.split(term).length - 1;
}
