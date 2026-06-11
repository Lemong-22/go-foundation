import { describe, expect, test } from "bun:test";
import { renderToStaticMarkup } from "react-dom/server";

import { LessonMarkdown, LessonReader } from "@/components/lesson-reader";
import type { BlockView, CourseView, LessonView } from "./course-api/server";
import {
  buildLessonBlocks,
  buildDummyActivity,
  buildVideoEmbed,
  parseInlineMarkdown,
  parseMarkdown,
} from "./lesson-reader";

const firstBlock: BlockView = {
  ID: "block-second-from-position",
  Kind: "text",
  LessonID: "lesson-1",
  Markdown: "# Values\n\nA `value` is data.",
  Position: 20,
  PracticeRef: "",
  QuizRef: "",
  VideoCaption: "",
  VideoLocator: "",
  VideoProvider: "",
};

const secondBlock: BlockView = {
  ID: "block-first-from-position",
  Kind: "video",
  LessonID: "lesson-1",
  Markdown: "",
  Position: 10,
  PracticeRef: "",
  QuizRef: "",
  VideoCaption: "Overview",
  VideoLocator: "abc123",
  VideoProvider: "youtube",
};

const quizBlock: BlockView = {
  ID: "block-quiz-ref",
  Kind: "quiz",
  LessonID: "lesson-1",
  Markdown: "",
  Position: 30,
  PracticeRef: "",
  QuizRef: "foundations-quiz",
  VideoCaption: "",
  VideoLocator: "",
  VideoProvider: "",
};

const practiceBlock: BlockView = {
  ID: "block-practice-ref",
  Kind: "practice",
  LessonID: "lesson-1",
  Markdown: "",
  Position: 40,
  PracticeRef: "values-practice",
  QuizRef: "",
  VideoCaption: "",
  VideoLocator: "",
  VideoProvider: "",
};

const course: CourseView = {
  CreatedAt: "2026-06-01T00:00:00Z",
  Description: "A course",
  ID: "course-1",
  InstructorID: "instructor-1",
  Slug: "course-one",
  Status: "published",
  Title: "Course One",
  UpdatedAt: "2026-06-02T00:00:00Z",
};

const lesson: LessonView = {
  CourseID: course.ID,
  CreatedAt: "2026-06-01T00:00:00Z",
  ID: "lesson-1",
  Order: 1,
  Title: "Lesson One",
  UpdatedAt: "2026-06-02T00:00:00Z",
};

describe("lesson reader helpers", () => {
  test("preserves service block order when building reader blocks", () => {
    const blocks = buildLessonBlocks([
      firstBlock,
      secondBlock,
      quizBlock,
      practiceBlock,
    ]);

    expect(blocks.map((block) => block.ID)).toEqual([
      "block-second-from-position",
      "block-first-from-position",
      "block-quiz-ref",
      "block-practice-ref",
    ]);
    expect(blocks.map((block) => block.number)).toEqual([
      "01",
      "02",
      "03",
      "04",
    ]);
    expect(blocks[1].label).toBe("YouTube");
    expect(blocks[2].kind).toBe("quiz");
    expect(blocks[2].label).toBe("Sample quiz");
    expect(blocks[3].kind).toBe("practice");
    expect(blocks[3].label).toBe("Sample practice");
  });

  test("builds video embeds from provider and locator", () => {
    expect(
      buildVideoEmbed({
        VideoLocator: "abc123",
        VideoProvider: "youtube",
      }),
    ).toEqual({
      kind: "iframe",
      providerLabel: "YouTube",
      src: "https://www.youtube-nocookie.com/embed/abc123",
    });
    expect(
      buildVideoEmbed({
        VideoLocator: "mux-playback",
        VideoProvider: "mux",
      }),
    ).toEqual({
      kind: "iframe",
      providerLabel: "Mux",
      src: "https://player.mux.com/mux-playback",
    });
    expect(
      buildVideoEmbed({
        VideoLocator: "javascript:alert(1)",
        VideoProvider: "url",
      }),
    ).toEqual({
      kind: "unavailable",
      providerLabel: "URL",
    });
  });

  test("parses markdown blocks without treating raw HTML as markup", () => {
    expect(
      parseMarkdown("# Heading\n\n- One\n- Two\n\n```ts\nconst x = 1;\n```"),
    ).toEqual([
      {
        level: 2,
        text: "Heading",
        type: "heading",
      },
      {
        items: ["One", "Two"],
        type: "list",
      },
      {
        code: "const x = 1;",
        language: "ts",
        type: "code",
      },
    ]);

    const html = renderToStaticMarkup(
      <LessonMarkdown markdown={"<script>alert(1)</script>\n\n`<img>`"} />,
    );

    expect(html).not.toContain("<script>");
    expect(html).toContain("&lt;script&gt;alert(1)&lt;/script&gt;");
    expect(html).toContain("&lt;img&gt;");
  });

  test("keeps unsafe inline links as text", () => {
    expect(parseInlineMarkdown("[safe](https://example.com)")).toEqual([
      {
        href: "https://example.com",
        text: "safe",
        type: "link",
      },
    ]);
    expect(parseInlineMarkdown("[bad](javascript:alert(1))")).toEqual([
      {
        text: "bad",
        type: "text",
      },
    ]);
  });

  test("builds clearly labeled dummy activity data without answer-bearing fields", () => {
    const quiz = buildDummyActivity(quizBlock);
    const practice = buildDummyActivity(practiceBlock);

    expect(quiz).toMatchObject({
      badge: "Sample data",
      kind: "quiz",
      ref: "foundations-quiz",
    });
    expect(practice).toMatchObject({
      badge: "Sample data",
      kind: "practice",
      ref: "values-practice",
    });

    const serialized = JSON.stringify([quiz, practice]).toLowerCase();
    expect(serialized).not.toContain("correct");
    expect(serialized).not.toContain("solution");
    expect(serialized).not.toContain("expected");
    expect(serialized).not.toContain("hidden");
    expect(serialized).not.toContain("answer");
  });

  test("renders quiz and practice refs as visible sample content", () => {
    const html = renderToStaticMarkup(
      <LessonReader
        blocks={[quizBlock, practiceBlock]}
        course={course}
        lesson={lesson}
      />,
    );

    expect(html).toContain("Sample data");
    expect(html).toContain("foundations-quiz");
    expect(html).toContain("values-practice");
    expect(html).toContain("Sample quiz prompt");
    expect(html).toContain("Sample practice prompt");
    expect(html.toLowerCase()).not.toContain("correct");
    expect(html.toLowerCase()).not.toContain("solution");
    expect(html.toLowerCase()).not.toContain("expected");
    expect(html.toLowerCase()).not.toContain("hidden test");
    expect(html.toLowerCase()).not.toContain("answer");
  });
});
