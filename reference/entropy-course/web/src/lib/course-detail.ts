import type { CourseView, LessonView } from "@/lib/course-api/server";

const courseUpdatedAtFormatter = new Intl.DateTimeFormat("en", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

export function buildSyllabusItems(lessons: LessonView[]) {
  return lessons.map((lesson, index) => ({
    ...lesson,
    href: lessonReaderPath(lesson),
    number: formatLessonNumber(index + 1),
  }));
}

export function lessonReaderPath(lesson: Pick<LessonView, "ID">) {
  return `/lessons/${encodeURIComponent(lesson.ID)}`;
}

export function courseTestPath(course: Pick<CourseView, "ID">) {
  return `/courses/${encodeURIComponent(course.ID)}/test`;
}

export function formatLessonNumber(value: number) {
  return value.toString().padStart(2, "0");
}

export function lessonCountLabel(count: number) {
  return `${count} ${count === 1 ? "lesson" : "lessons"}`;
}

export function formatCourseUpdatedAt(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Recently";
  }

  return courseUpdatedAtFormatter.format(date);
}
