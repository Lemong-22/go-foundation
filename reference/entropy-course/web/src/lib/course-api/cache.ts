export const COURSE_PUBLISHED_CATALOG_TAG = "course-reader:catalog:published";

export type CourseRevalidationScope = "block" | "catalog" | "course" | "lesson";

export type CourseRevalidationPlan = {
  scope: CourseRevalidationScope;
  tags: string[];
};

export type CourseRevalidationPlanResult =
  | {
      ok: true;
      plan: CourseRevalidationPlan;
    }
  | {
      message: string;
      ok: false;
    };

type CourseRevalidationInput = {
  blockID?: unknown;
  courseID?: unknown;
  lessonID?: unknown;
  scope?: unknown;
};

export function courseReadCacheOptions(tags: string[]) {
  return {
    cache: "force-cache" as const,
    next: {
      revalidate: false as const,
      tags: dedupeTags(tags),
    },
  };
}

export function publishedCatalogCacheTags() {
  return [COURSE_PUBLISHED_CATALOG_TAG];
}

export function courseCacheTags(courseID: string) {
  return [courseTag(courseID)];
}

export function courseLessonsCacheTags(courseID: string) {
  return [courseTag(courseID), courseLessonsTag(courseID)];
}

export function lessonCacheTags(lessonID: string) {
  return [lessonTag(lessonID)];
}

export function lessonBlocksCacheTags(lessonID: string) {
  return [lessonTag(lessonID), lessonBlocksTag(lessonID)];
}

export function lessonBlockCacheTags(blockID: string) {
  return [blockTag(blockID)];
}

export function learnerQuizCacheTags(quizID: string) {
  return [learnerAggregateTag("quiz", quizID)];
}

export function learnerPracticeCacheTags(practiceID: string) {
  return [learnerAggregateTag("practice", practiceID)];
}

export function learnerTestCacheTags(testID: string) {
  return [learnerAggregateTag("test", testID)];
}

export function buildCourseRevalidationPlan(
  input: unknown,
): CourseRevalidationPlanResult {
  if (!isRecord(input)) {
    return failPlan("Request body must be a JSON object.");
  }

  const body: CourseRevalidationInput = input;

  switch (body.scope) {
    case "catalog":
      return okPlan("catalog", publishedCatalogCacheTags());
    case "course": {
      const courseID = readID(body.courseID, "courseID");
      if (!courseID.ok) {
        return failPlan(courseID.message);
      }

      return okPlan("course", [
        COURSE_PUBLISHED_CATALOG_TAG,
        courseTag(courseID.value),
        courseLessonsTag(courseID.value),
      ]);
    }
    case "lesson": {
      const lessonID = readID(body.lessonID, "lessonID");
      if (!lessonID.ok) {
        return failPlan(lessonID.message);
      }

      const tags = [
        lessonTag(lessonID.value),
        lessonBlocksTag(lessonID.value),
      ];
      const courseID = readOptionalID(body.courseID);
      if (courseID) {
        tags.push(courseLessonsTag(courseID));
      }

      return okPlan("lesson", tags);
    }
    case "block": {
      const blockID = readID(body.blockID, "blockID");
      if (!blockID.ok) {
        return failPlan(blockID.message);
      }

      const tags = [blockTag(blockID.value)];
      const lessonID = readOptionalID(body.lessonID);
      if (lessonID) {
        tags.push(lessonBlocksTag(lessonID));
      }

      return okPlan("block", tags);
    }
    default:
      return failPlan("scope must be one of catalog, course, lesson, or block.");
  }
}

function courseTag(courseID: string) {
  return scopedTag("course", courseID);
}

function courseLessonsTag(courseID: string) {
  return scopedTag("course-lessons", courseID);
}

function lessonTag(lessonID: string) {
  return scopedTag("lesson", lessonID);
}

function lessonBlocksTag(lessonID: string) {
  return scopedTag("lesson-blocks", lessonID);
}

function blockTag(blockID: string) {
  return scopedTag("block", blockID);
}

function learnerAggregateTag(kind: string, value: string) {
  return scopedTag(`learner-${kind}`, value);
}

function scopedTag(scope: string, value: string) {
  return `course-reader:${scope}:${encodeURIComponent(value.trim())}`;
}

function dedupeTags(tags: string[]) {
  return Array.from(new Set(tags.filter(Boolean)));
}

function okPlan(
  scope: CourseRevalidationScope,
  tags: string[],
): CourseRevalidationPlanResult {
  return {
    ok: true,
    plan: {
      scope,
      tags: dedupeTags(tags),
    },
  };
}

function failPlan(message: string): CourseRevalidationPlanResult {
  return {
    message,
    ok: false,
  };
}

function readID(value: unknown, fieldName: string) {
  if (typeof value !== "string" || value.trim().length === 0) {
    return {
      message: `${fieldName} is required.`,
      ok: false as const,
    };
  }

  return {
    ok: true as const,
    value: value.trim(),
  };
}

function readOptionalID(value: unknown) {
  if (typeof value !== "string") {
    return undefined;
  }

  const trimmed = value.trim();
  return trimmed || undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
