import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import {
  CourseDetail,
  CourseDetailError,
} from "@/components/course-detail";
import { getCourse, listCourseLessons } from "@/lib/course-api/server";

type CourseDetailPageProps = {
  params: Promise<{
    courseID: string;
  }>;
};

export const metadata: Metadata = {
  title: "Course Syllabus | Entropy Course Reader",
  description: "Review a published course syllabus in the Entropy learner reader.",
};

export default async function CourseDetailPage({ params }: CourseDetailPageProps) {
  await connection();

  const { courseID } = await params;
  const [courseResult, lessonsResult] = await Promise.all([
    getCourse(courseID),
    listCourseLessons(courseID),
  ]);

  if (!courseResult.ok) {
    if (courseResult.error.code === "not-found") {
      notFound();
    }

    return <CourseDetailError error={courseResult.error} />;
  }

  if (courseResult.data.Course.Status !== "published") {
    notFound();
  }

  if (!lessonsResult.ok) {
    return (
      <CourseDetailError
        course={courseResult.data.Course}
        error={lessonsResult.error}
      />
    );
  }

  return (
    <CourseDetail
      course={courseResult.data.Course}
      lessons={lessonsResult.data.Lessons}
    />
  );
}
