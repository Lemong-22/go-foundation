import type { CourseView } from "@/lib/course-api/server";

const updatedAtFormatter = new Intl.DateTimeFormat("en", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

export function selectPublishedCourses(courses: CourseView[]) {
  return courses.filter((course) => course.Status === "published");
}

export function courseDetailPath(course: Pick<CourseView, "ID">) {
  return `/courses/${encodeURIComponent(course.ID)}`;
}

export function formatCourseUpdatedAt(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Recently updated";
  }

  return `Updated ${updatedAtFormatter.format(date)}`;
}

export function pluralizeCourse(count: number) {
  return `${count} ${count === 1 ? "course" : "courses"}`;
}
