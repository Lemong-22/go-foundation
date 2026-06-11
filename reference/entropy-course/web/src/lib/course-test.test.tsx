import { describe, expect, test } from "bun:test";
import { renderToStaticMarkup } from "react-dom/server";

import { CourseTest } from "@/components/course-test";
import type { CourseView } from "./course-api/server";
import {
  buildDummyCourseTest,
  courseTestItemCountLabel,
} from "./course-test";

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

describe("course test helpers", () => {
  test("builds clearly labeled dummy course test data without restricted fields", () => {
    const testView = buildDummyCourseTest(course);

    expect(testView).toMatchObject({
      badge: "Sample data",
      kind: "test",
      title: "Sample course check",
    });
    expect(testView.sections).toHaveLength(3);
    expect(courseTestItemCountLabel(1)).toBe("1 prompt");
    expect(courseTestItemCountLabel(3)).toBe("3 prompts");

    const serialized = JSON.stringify(testView).toLowerCase();
    expect(serialized).not.toContain("correct");
    expect(serialized).not.toContain("solution");
    expect(serialized).not.toContain("expected");
    expect(serialized).not.toContain("hidden");
    expect(serialized).not.toContain("answer");
  });

  test("renders the course test view as visible sample content", () => {
    const html = renderToStaticMarkup(
      <CourseTest course={course} test={buildDummyCourseTest(course)} />,
    );

    expect(html).toContain("Sample data");
    expect(html).toContain("Sample course check");
    expect(html).toContain("Concept check");
    expect(html).toContain("Code reading");
    expect(html).toContain("Practice plan");
    expect(html).toContain("/courses/course-1");
    expect(html.toLowerCase()).not.toContain("correct");
    expect(html.toLowerCase()).not.toContain("solution");
    expect(html.toLowerCase()).not.toContain("expected");
    expect(html.toLowerCase()).not.toContain("hidden test");
    expect(html.toLowerCase()).not.toContain("answer");
  });
});
