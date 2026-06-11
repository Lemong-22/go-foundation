import type { Metadata } from "next";
import { connection } from "next/server";

import {
  CourseCatalog,
  CourseCatalogError,
} from "@/components/course-catalog";
import { listPublishedCourses } from "@/lib/course-api/server";
import { selectPublishedCourses } from "@/lib/catalog";

export const metadata: Metadata = {
  title: "Published Courses | Entropy Course Reader",
  description: "Browse published courses in the Entropy learner reader.",
};

export default async function Page() {
  await connection();

  const result = await listPublishedCourses();

  if (!result.ok) {
    return <CourseCatalogError error={result.error} />;
  }

  return <CourseCatalog courses={selectPublishedCourses(result.data.Courses)} />;
}
