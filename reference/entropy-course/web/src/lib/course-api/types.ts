export type CourseView = {
  ID: string;
  Title: string;
  Slug: string;
  Description: string;
  InstructorID: string;
  Status: string;
  CreatedAt: string;
  UpdatedAt: string;
};

export type ListCoursesOutput = {
  Courses: CourseView[];
};

export type GetCourseOutput = {
  Course: CourseView;
};

export type LessonView = {
  ID: string;
  CourseID: string;
  Title: string;
  Order: number;
  CreatedAt: string;
  UpdatedAt: string;
};

export type ListLessonsOutput = {
  Lessons: LessonView[];
};

export type GetLessonOutput = {
  Lesson: LessonView;
};

export type BlockView = {
  ID: string;
  LessonID: string;
  Kind: string;
  Position: number;
  Markdown: string;
  VideoProvider: string;
  VideoLocator: string;
  VideoCaption: string;
  QuizRef: string;
  PracticeRef: string;
};

export type ListLessonBlocksOutput = {
  Blocks: BlockView[];
};

export type GetLessonBlockOutput = {
  Block: BlockView;
};

export type LearnerQuizQuestionView = {
  ID: string;
  QuizID: string;
  Type: string;
  Prompt: string;
  Options: string[];
  Position: number;
};

export type LearnerQuizDetailView = {
  ID: string;
  CourseID: string;
  Title: string;
  PassThreshold: number;
  QuestionCount: number;
  CreatedAt: string;
  UpdatedAt: string;
  Questions: LearnerQuizQuestionView[];
};

export type GetLearnerQuizOutput = {
  Quiz: LearnerQuizDetailView;
};

export type LearnerPracticeDetailView = {
  ID: string;
  CourseID: string;
  Title: string;
  Language: string;
  Prompt: string;
  StarterCode: string;
  CreatedAt: string;
  UpdatedAt: string;
};

export type GetLearnerPracticeOutput = {
  Practice: LearnerPracticeDetailView;
};

export type LearnerTestItemView = {
  ID: string;
  TestID: string;
  Kind: string;
  Position: number;
  ChoicePrompt: string;
  ChoiceType: string;
  ChoiceOptions: string[];
  CodingPrompt: string;
  Language: string;
  StarterCode: string;
};

export type LearnerTestDetailView = {
  ID: string;
  CourseID: string;
  Title: string;
  TimeLimitMinutes: number | null;
  PassThreshold: number;
  ItemCount: number;
  CreatedAt: string;
  UpdatedAt: string;
  Items: LearnerTestItemView[];
};

export type GetLearnerTestOutput = {
  Test: LearnerTestDetailView;
};

export type CourseApiErrorCode =
  | "configuration"
  | "unauthorized"
  | "not-found"
  | "validation"
  | "conflict"
  | "upstream"
  | "network"
  | "parse";

export type CourseApiError = {
  code: CourseApiErrorCode;
  message: string;
  endpoint?: string;
  status?: number;
};

export type CourseApiResult<T> =
  | {
      ok: true;
      data: T;
    }
  | {
      ok: false;
      error: CourseApiError;
    };
