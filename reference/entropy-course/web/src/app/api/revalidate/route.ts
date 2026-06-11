import { revalidateTag } from "next/cache";
import type { NextRequest } from "next/server";

import { buildCourseRevalidationPlan } from "@/lib/course-api/cache";

export const runtime = "nodejs";

export async function POST(request: NextRequest) {
  const configuredSecret = process.env.COURSE_REVALIDATION_SECRET?.trim();

  if (!configuredSecret) {
    return Response.json(
      {
        message: "COURSE_REVALIDATION_SECRET is required.",
        revalidated: false,
      },
      { status: 503 },
    );
  }

  if (readProvidedSecret(request) !== configuredSecret) {
    return Response.json(
      {
        message: "Unauthorized revalidation request.",
        revalidated: false,
      },
      { status: 401 },
    );
  }

  const body = await readJson(request);
  if (!body.ok) {
    return Response.json(
      {
        message: body.message,
        revalidated: false,
      },
      { status: 400 },
    );
  }

  const plan = buildCourseRevalidationPlan(body.data);
  if (!plan.ok) {
    return Response.json(
      {
        message: plan.message,
        revalidated: false,
      },
      { status: 400 },
    );
  }

  for (const tag of plan.plan.tags) {
    revalidateTag(tag, { expire: 0 });
  }

  return Response.json({
    revalidated: true,
    scope: plan.plan.scope,
    tags: plan.plan.tags,
  });
}

function readProvidedSecret(request: NextRequest) {
  const authorization = request.headers.get("authorization")?.trim();
  if (authorization?.toLowerCase().startsWith("bearer ")) {
    return authorization.slice("bearer ".length).trim();
  }

  return request.headers.get("x-course-revalidation-secret")?.trim();
}

async function readJson(request: NextRequest) {
  try {
    return {
      data: await request.json(),
      ok: true as const,
    };
  } catch {
    return {
      message: "Request body must be valid JSON.",
      ok: false as const,
    };
  }
}
