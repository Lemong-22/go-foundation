import { describe, expect, test } from "bun:test";

import {
  buildSyllabusItems,
  courseTestPath,
  formatCourseUpdatedAt,
  formatLessonNumber,
  lessonCountLabel,
  lessonReaderPath,
} from "./course-detail";
import type { LessonView } from "./course-api/server";

const firstLesson: LessonView = {
  CourseID: "course-1",
  CreatedAt: "2026-06-01T00:00:00Z",
  ID: "lesson/second-from-service",
  Order: 20,
  Title: "Second from service",
  UpdatedAt: "2026-06-02T00:00:00Z",
};

const secondLesson: LessonView = {
  CourseID: "course-1",
  CreatedAt: "2026-06-01T00:00:00Z",
  ID: "lesson-first-from-service",
  Order: 10,
  Title: "First from service",
  UpdatedAt: "2026-06-02T00:00:00Z",
};

describe("course detail helpers", () => {
  test("preserves service lesson order when building syllabus items", () => {
    const syllabus = buildSyllabusItems([firstLesson, secondLesson]);

    expect(syllabus.map((lesson) => lesson.ID)).toEqual([
      "lesson/second-from-service",
      "lesson-first-from-service",
    ]);
    expect(syllabus.map((lesson) => lesson.number)).toEqual(["01", "02"]);
  });

  test("builds encoded lesson reader route paths", () => {
    expect(lessonReaderPath({ ID: "lesson/with spaces" })).toBe(
      "/lessons/lesson%2Fwith%20spaces",
    );
    expect(courseTestPath({ ID: "course/with spaces" })).toBe(
      "/courses/course%2Fwith%20spaces/test",
    );
  });

  test("formats course detail labels", () => {
    expect(formatLessonNumber(9)).toBe("09");
    expect(lessonCountLabel(1)).toBe("1 lesson");
    expect(lessonCountLabel(3)).toBe("3 lessons");
    expect(formatCourseUpdatedAt("2026-06-02T00:00:00Z")).toBe(
      "Jun 2, 2026",
    );
    expect(formatCourseUpdatedAt("not-a-date")).toBe("Recently");
  });
});
