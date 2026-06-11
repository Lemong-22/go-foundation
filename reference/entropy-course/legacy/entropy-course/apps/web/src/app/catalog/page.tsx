import { redirect } from "next/navigation";

import { DEFAULT_LESSON_PATH } from "@/lib/learning-paths";
import { getCurrentSession } from "@/lib/server-session";

export default async function CatalogPage() {
  const session = await getCurrentSession();

  redirect(session?.user ? DEFAULT_LESSON_PATH : "/login");
}
