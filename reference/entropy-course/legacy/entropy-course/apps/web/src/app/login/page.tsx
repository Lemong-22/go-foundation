import { redirect } from "next/navigation";

import { DEFAULT_LESSON_PATH } from "@/lib/learning-paths";
import { getCurrentSession } from "@/lib/server-session";

import LoginPageClient from "./login-page-client";

export default async function LoginPage() {
  const session = await getCurrentSession();

  if (session?.user) {
    redirect(DEFAULT_LESSON_PATH);
  }

  return <LoginPageClient />;
}
