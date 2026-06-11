import type {
  CourseApiError,
  CourseApiResult,
  GetCourseOutput,
  GetLearnerPracticeOutput,
  GetLearnerQuizOutput,
  GetLearnerTestOutput,
  GetLessonBlockOutput,
  GetLessonOutput,
  ListCoursesOutput,
  ListLessonBlocksOutput,
  ListLessonsOutput,
} from "./types";
import {
  courseCacheTags,
  courseLessonsCacheTags,
  courseReadCacheOptions,
  learnerPracticeCacheTags,
  learnerQuizCacheTags,
  learnerTestCacheTags,
  lessonBlockCacheTags,
  lessonBlocksCacheTags,
  lessonCacheTags,
  publishedCatalogCacheTags,
} from "./cache";

const DEFAULT_COURSE_API_BASE_URL = "http://127.0.0.1:8788";
const LEARNER_VIEW = "learner";
const PUBLISHED_STATUS = "published";

export type CourseApiFetch = (
  input: Parameters<typeof fetch>[0],
  init?: Parameters<typeof fetch>[1],
) => Promise<Response>;

export type CourseApiSettings = {
  baseUrl?: string;
  token: string;
};

type CourseApiClientOptions = {
  fetcher: CourseApiFetch;
  settings: () => CourseApiResult<CourseApiSettings>;
};

type ResolvedCourseApiSettings = {
  baseUrl: string;
  token: string;
};

export type CourseApiClient = ReturnType<typeof createCourseApiClient>;

export function createCourseApiClient(options: CourseApiClientOptions) {
  return {
    async getCourse(courseID: string) {
      const endpoint = readEndpointID("/v1/courses", courseID, "courseID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<GetCourseOutput>(
        options,
        endpoint.data,
        courseCacheTags(courseID),
      );
    },

    async getLesson(lessonID: string) {
      const endpoint = readEndpointID("/v1/lessons", lessonID, "lessonID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<GetLessonOutput>(
        options,
        endpoint.data,
        lessonCacheTags(lessonID),
      );
    },

    async getLessonBlock(blockID: string) {
      const endpoint = readEndpointID("/v1/blocks", blockID, "blockID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<GetLessonBlockOutput>(
        options,
        endpoint.data,
        lessonBlockCacheTags(blockID),
      );
    },

    async getLearnerQuiz(quizID: string) {
      const endpoint = readEndpointID("/v1/quizzes", quizID, "quizID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<GetLearnerQuizOutput>(
        options,
        learnerViewEndpoint(endpoint.data),
        learnerQuizCacheTags(quizID),
      );
    },

    async getLearnerPractice(practiceID: string) {
      const endpoint = readEndpointID("/v1/practices", practiceID, "practiceID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<GetLearnerPracticeOutput>(
        options,
        learnerViewEndpoint(endpoint.data),
        learnerPracticeCacheTags(practiceID),
      );
    },

    async getLearnerTest(testID: string) {
      const endpoint = readEndpointID("/v1/tests", testID, "testID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<GetLearnerTestOutput>(
        options,
        learnerViewEndpoint(endpoint.data),
        learnerTestCacheTags(testID),
      );
    },

    async listCourseLessons(courseID: string) {
      const endpoint = readEndpointID("/v1/courses", courseID, "courseID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<ListLessonsOutput>(
        options,
        `${endpoint.data}/lessons`,
        courseLessonsCacheTags(courseID),
      );
    },

    async listLessonBlocks(lessonID: string) {
      const endpoint = readEndpointID("/v1/lessons", lessonID, "lessonID");
      if (!endpoint.ok) {
        return endpoint;
      }

      return request<ListLessonBlocksOutput>(
        options,
        `${endpoint.data}/blocks`,
        lessonBlocksCacheTags(lessonID),
      );
    },

    async listPublishedCourses() {
      return request<ListCoursesOutput>(
        options,
        `/v1/courses?status=${encodeURIComponent(PUBLISHED_STATUS)}`,
        publishedCatalogCacheTags(),
      );
    },
  };
}

function learnerViewEndpoint(endpoint: string) {
  return `${endpoint}?view=${encodeURIComponent(LEARNER_VIEW)}`;
}

async function request<T>(
  options: CourseApiClientOptions,
  endpoint: string,
  tags: string[],
): Promise<CourseApiResult<T>> {
  const settings = resolveSettings(options.settings());
  if (!settings.ok) {
    return settings;
  }

  const url = `${settings.data.baseUrl}${endpoint}`;

  let response: Response;
  try {
    response = await options.fetcher(url, {
      ...courseReadCacheOptions(tags),
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${settings.data.token}`,
      },
      method: "GET",
    });
  } catch {
    return fail({
      code: "network",
      endpoint,
      message: "Unable to reach the course API.",
    });
  }

  if (!response.ok) {
    return fail(await errorFromResponse(response, endpoint));
  }

  try {
    return {
      data: (await response.json()) as T,
      ok: true,
    };
  } catch {
    return fail({
      code: "parse",
      endpoint,
      message: "The course API returned invalid JSON.",
      status: response.status,
    });
  }
}

function resolveSettings(
  settings: CourseApiResult<CourseApiSettings>,
): CourseApiResult<ResolvedCourseApiSettings> {
  if (!settings.ok) {
    return settings;
  }

  const token = settings.data.token.trim();
  if (!token) {
    return fail({
      code: "configuration",
      message: "A course API bearer token is required for course API reads.",
    });
  }

  const rawBaseUrl = settings.data.baseUrl?.trim() || DEFAULT_COURSE_API_BASE_URL;

  try {
    const url = new URL(rawBaseUrl);

    return {
      data: {
        baseUrl: url.toString().replace(/\/+$/, ""),
        token,
      },
      ok: true,
    };
  } catch {
    return fail({
      code: "configuration",
      message: "COURSE_API_BASE_URL must be a valid absolute URL.",
    });
  }
}

function readEndpointID(
  prefix: string,
  value: string,
  fieldName: string,
): CourseApiResult<string> {
  const trimmed = value.trim();
  if (!trimmed) {
    return fail({
      code: "validation",
      message: `${fieldName} is required.`,
    });
  }

  return {
    data: `${prefix}/${encodeURIComponent(trimmed)}`,
    ok: true,
  };
}

async function errorFromResponse(
  response: Response,
  endpoint: string,
): Promise<CourseApiError> {
  return {
    code: codeForStatus(response.status),
    endpoint,
    message: await messageFromResponse(response),
    status: response.status,
  };
}

async function messageFromResponse(response: Response): Promise<string> {
  try {
    const body: unknown = await response.json();

    if (isRecord(body) && typeof body.error === "string") {
      return body.error;
    }
  } catch {
    // Fall through to the HTTP status text below.
  }

  return response.statusText || `Course API request failed with ${response.status}.`;
}

function codeForStatus(status: number): CourseApiError["code"] {
  switch (status) {
    case 400:
    case 422:
      return "validation";
    case 401:
    case 403:
      return "unauthorized";
    case 404:
      return "not-found";
    case 409:
      return "conflict";
    default:
      return "upstream";
  }
}

function fail(error: CourseApiError): CourseApiResult<never> {
  return {
    error,
    ok: false,
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
