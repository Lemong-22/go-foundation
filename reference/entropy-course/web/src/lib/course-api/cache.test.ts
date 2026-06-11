import { describe, expect, test } from "bun:test";

import {
  COURSE_PUBLISHED_CATALOG_TAG,
  buildCourseRevalidationPlan,
  courseReadCacheOptions,
} from "./cache";

describe("course API cache helpers", () => {
  test("builds persistent fetch cache options with explicit tags", () => {
    expect(courseReadCacheOptions(["one", "one", "two"])).toEqual({
      cache: "force-cache",
      next: {
        revalidate: false,
        tags: ["one", "two"],
      },
    });
  });

  test("builds targeted catalog and course revalidation plans", () => {
    expect(buildCourseRevalidationPlan({ scope: "catalog" })).toEqual({
      ok: true,
      plan: {
        scope: "catalog",
        tags: [COURSE_PUBLISHED_CATALOG_TAG],
      },
    });

    expect(
      buildCourseRevalidationPlan({
        courseID: "course/one",
        scope: "course",
      }),
    ).toEqual({
      ok: true,
      plan: {
        scope: "course",
        tags: [
          COURSE_PUBLISHED_CATALOG_TAG,
          "course-reader:course:course%2Fone",
          "course-reader:course-lessons:course%2Fone",
        ],
      },
    });
  });

  test("builds lesson and block revalidation plans without global refresh", () => {
    expect(
      buildCourseRevalidationPlan({
        courseID: "course-1",
        lessonID: "lesson/one",
        scope: "lesson",
      }),
    ).toEqual({
      ok: true,
      plan: {
        scope: "lesson",
        tags: [
          "course-reader:lesson:lesson%2Fone",
          "course-reader:lesson-blocks:lesson%2Fone",
          "course-reader:course-lessons:course-1",
        ],
      },
    });

    expect(
      buildCourseRevalidationPlan({
        blockID: "block/one",
        lessonID: "lesson-1",
        scope: "block",
      }),
    ).toEqual({
      ok: true,
      plan: {
        scope: "block",
        tags: [
          "course-reader:block:block%2Fone",
          "course-reader:lesson-blocks:lesson-1",
        ],
      },
    });
  });

  test("rejects invalid revalidation requests", () => {
    expect(buildCourseRevalidationPlan(null)).toMatchObject({
      ok: false,
    });
    expect(buildCourseRevalidationPlan({ scope: "course" })).toEqual({
      message: "courseID is required.",
      ok: false,
    });
    expect(buildCourseRevalidationPlan({ scope: "unknown" })).toEqual({
      message: "scope must be one of catalog, course, lesson, or block.",
      ok: false,
    });
  });
});
