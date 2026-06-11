import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { authClient } from "@/lib/auth-client";

type Session = typeof authClient.$Infer.Session;
type SessionResponse =
  | Session
  | {
      data: Session | null;
      error: unknown;
    }
  | null
  | undefined;

export async function getCurrentSession(): Promise<Session | null> {
  const response = await authClient.getSession({
    fetchOptions: {
      headers: await headers(),
    },
  });

  return extractSession(response as SessionResponse);
}

export async function requireSession(): Promise<Session> {
  const session = await getCurrentSession();

  if (!session?.user) {
    redirect("/login");
  }

  return session;
}

function extractSession(response: SessionResponse): Session | null {
  if (!response) {
    return null;
  }

  if ("data" in response) {
    return response.data;
  }

  return response;
}
