import "server-only";

import { createCourseApiClient } from "./transport";
import type { CourseApiResult } from "./types";

export type {
  BlockView,
  CourseApiError,
  CourseApiErrorCode,
  CourseApiResult,
  CourseView,
  GetCourseOutput,
  GetLearnerPracticeOutput,
  GetLearnerQuizOutput,
  GetLearnerTestOutput,
  GetLessonBlockOutput,
  GetLessonOutput,
  LearnerPracticeDetailView,
  LearnerQuizDetailView,
  LearnerQuizQuestionView,
  LearnerTestDetailView,
  LearnerTestItemView,
  LessonView,
  ListCoursesOutput,
  ListLessonBlocksOutput,
  ListLessonsOutput,
} from "./types";

const courseApi = createCourseApiClient({
  fetcher: fetch,
  settings: readCourseApiSettings,
});

export const getCourse = courseApi.getCourse;
export const getLearnerPractice = courseApi.getLearnerPractice;
export const getLearnerQuiz = courseApi.getLearnerQuiz;
export const getLearnerTest = courseApi.getLearnerTest;
export const getLesson = courseApi.getLesson;
export const getLessonBlock = courseApi.getLessonBlock;
export const listCourseLessons = courseApi.listCourseLessons;
export const listLessonBlocks = courseApi.listLessonBlocks;
export const listPublishedCourses = courseApi.listPublishedCourses;

function readCourseApiSettings(): CourseApiResult<{
  baseUrl?: string;
  token: string;
}> {
  const token = process.env.COURSE_CLI_API_TOKEN?.trim();

  if (!token) {
    return {
      error: {
        code: "configuration",
        message: "COURSE_CLI_API_TOKEN is required for course API reads.",
      },
      ok: false,
    };
  }

  return {
    data: {
      baseUrl: process.env.COURSE_API_BASE_URL,
      token,
    },
    ok: true,
  };
}
