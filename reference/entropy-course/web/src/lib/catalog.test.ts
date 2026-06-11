import { describe, expect, test } from "bun:test";

import {
  courseDetailPath,
  formatCourseUpdatedAt,
  pluralizeCourse,
  selectPublishedCourses,
} from "./catalog";
import type { CourseView } from "./course-api/server";

const baseCourse: CourseView = {
  CreatedAt: "2026-06-01T00:00:00Z",
  Description: "A course.",
  ID: "course-1",
  InstructorID: "instructor-1",
  Slug: "course-1",
  Status: "published",
  Title: "Course 1",
  UpdatedAt: "2026-06-02T00:00:00Z",
};

describe("catalog helpers", () => {
  test("selects only published courses without remapping DTO fields", () => {
    const courses = [
      baseCourse,
      {
        ...baseCourse,
        ID: "draft-course",
        Status: "draft",
        Title: "Draft course",
      },
    ];

    expect(selectPublishedCourses(courses)).toEqual([baseCourse]);
  });

  test("builds encoded course detail route paths", () => {
    expect(courseDetailPath({ ID: "course/with spaces" })).toBe(
      "/courses/course%2Fwith%20spaces",
    );
  });

  test("formats course metadata for catalog cards", () => {
    expect(formatCourseUpdatedAt("2026-06-02T00:00:00Z")).toBe(
      "Updated Jun 2, 2026",
    );
    expect(formatCourseUpdatedAt("not-a-date")).toBe("Recently updated");
    expect(pluralizeCourse(1)).toBe("1 course");
    expect(pluralizeCourse(2)).toBe("2 courses");
  });
});
