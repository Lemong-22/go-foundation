import { redirect } from "next/navigation";

import { DEFAULT_LESSON_PATH } from "@/lib/learning-paths";
import { requireSession } from "@/lib/server-session";

export default async function CoursePage() {
  await requireSession();
  redirect(DEFAULT_LESSON_PATH);
}
