import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import {
  LessonReader,
  LessonReaderError,
} from "@/components/lesson-reader";
import {
  getCourse,
  getLesson,
  listLessonBlocks,
} from "@/lib/course-api/server";

type LessonReaderPageProps = {
  params: Promise<{
    lessonID: string;
  }>;
};

export const metadata: Metadata = {
  title: "Lesson Reader | Entropy Course Reader",
  description: "Read an ordered lesson in the Entropy learner reader.",
};

export default async function LessonReaderPage({ params }: LessonReaderPageProps) {
  await connection();

  const { lessonID } = await params;
  const lessonResult = await getLesson(lessonID);

  if (!lessonResult.ok) {
    if (lessonResult.error.code === "not-found") {
      notFound();
    }

    return <LessonReaderError error={lessonResult.error} />;
  }

  const lesson = lessonResult.data.Lesson;
  const courseResult = await getCourse(lesson.CourseID);

  if (!courseResult.ok) {
    if (courseResult.error.code === "not-found") {
      notFound();
    }

    return <LessonReaderError error={courseResult.error} lesson={lesson} />;
  }

  if (courseResult.data.Course.Status !== "published") {
    notFound();
  }

  const blocksResult = await listLessonBlocks(lesson.ID);

  if (!blocksResult.ok) {
    return <LessonReaderError error={blocksResult.error} lesson={lesson} />;
  }

  return (
    <LessonReader
      blocks={blocksResult.data.Blocks}
      course={courseResult.data.Course}
      lesson={lesson}
    />
  );
}
