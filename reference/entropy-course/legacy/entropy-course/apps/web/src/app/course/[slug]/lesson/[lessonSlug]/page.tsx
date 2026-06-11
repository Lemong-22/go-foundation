import { notFound, redirect } from "next/navigation";
import { DEFAULT_LESSON_PATH, PRIMARY_COURSE_SLUG } from "@/lib/learning-paths";
import {
  mockCheatsheetContent,
  mockCodeSnippets,
  mockCourses,
  mockLessons,
  mockQuizQuestions,
} from "@/lib/mock-data";
import { requireSession } from "@/lib/server-session";
import { LessonExperience } from "./lesson-experience";

interface Props {
  params: Promise<{ slug: string; lessonSlug: string }>;
}

export default async function LessonPage({ params }: Props) {
  const { slug, lessonSlug } = await params;

  await requireSession();

  if (slug !== PRIMARY_COURSE_SLUG) {
    redirect(DEFAULT_LESSON_PATH);
  }

  const course = mockCourses.find((item) => item.slug === slug);
  const lessons = mockLessons[slug] ?? [];
  const lesson = lessons.find((l) => l.slug === lessonSlug);

  if (!course || !lesson) notFound();

  const lessonIndex = lessons.findIndex((l) => l.slug === lessonSlug);
  const prevLesson = lessons[lessonIndex - 1];
  const nextLesson = lessons[lessonIndex + 1];

  return (
    <LessonExperience
      course={course}
      lessons={lessons}
      lesson={lesson}
      courseSlug={slug}
      prevLesson={prevLesson}
      nextLesson={nextLesson}
      quizQuestions={mockQuizQuestions}
      codeSnippets={mockCodeSnippets}
      cheatsheet={mockCheatsheetContent}
    />
  );
}
