import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { CourseTest, CourseTestError } from "@/components/course-test";
import { getCourse } from "@/lib/course-api/server";
import { buildDummyCourseTest } from "@/lib/course-test";

type CourseTestPageProps = {
  params: Promise<{
    courseID: string;
  }>;
};

export const metadata: Metadata = {
  title: "Course Test | Entropy Course Reader",
  description:
    "Preview the course-level sample test in the Entropy learner reader.",
};

export default async function CourseTestPage({ params }: CourseTestPageProps) {
  await connection();

  const { courseID } = await params;
  const courseResult = await getCourse(courseID);

  if (!courseResult.ok) {
    if (courseResult.error.code === "not-found") {
      notFound();
    }

    return <CourseTestError error={courseResult.error} />;
  }

  const course = courseResult.data.Course;

  if (course.Status !== "published") {
    notFound();
  }

  return <CourseTest course={course} test={buildDummyCourseTest(course)} />;
}
