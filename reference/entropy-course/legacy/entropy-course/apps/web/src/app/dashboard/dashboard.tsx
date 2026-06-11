"use client";
import Link from "next/link";
import { authClient } from "@/lib/auth-client";
import { mockCourses } from "@/lib/mock-data";
import { Button } from "@entropy-course/ui/components/button";
import { Card } from "@entropy-course/ui/components/card";

export default function Dashboard({ session }: { session: typeof authClient.$Infer.Session }) {
  // In production this would come from the API
  const enrolledCourses = mockCourses.filter((c) => c.started);

  return (
    <div className="container mx-auto max-w-4xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">
          Welcome back, {session.user?.name ?? "Learner"}!
        </h1>
        <p className="text-muted-foreground">
          Pick up where you left off or explore a new course.
        </p>
      </div>

      {enrolledCourses.length > 0 ? (
        <section className="mb-8">
          <h2 className="text-xl font-semibold mb-4">Continue Learning</h2>
          <div className="space-y-4">
            {enrolledCourses.map((course) => (
              <Card key={course.id} className="p-4">
                <div className="flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-lg flex items-center justify-center text-white font-bold"
                    style={{ backgroundColor: course.coverColor }}
                  >
                    {course.title.slice(0, 2).toUpperCase()}
                  </div>
                  <div className="flex-1">
                    <Link
                      href={`/course/${course.slug}`}
                      className="font-semibold hover:text-primary"
                    >
                      {course.title}
                    </Link>
                    <div className="flex items-center gap-3 mt-1">
                      <div className="w-32 h-2 bg-muted rounded-full overflow-hidden">
                        <div
                          className="h-full bg-green-500 rounded-full"
                          style={{ width: `${course.progress ?? 0}%` }}
                        />
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {course.progress ?? 0}% complete
                      </span>
                    </div>
                  </div>
                  <Link href={`/course/${course.slug}`}>
                    <Button variant="outline" size="sm">
                      Continue
                    </Button>
                  </Link>
                </div>
              </Card>
            ))}
          </div>
        </section>
      ) : (
        <section className="mb-8">
          <Card className="p-8 text-center">
            <p className="text-muted-foreground mb-4">
              You haven&apos;t started any courses yet.
            </p>
            <Link href="/catalog">
              <Button>Browse Courses</Button>
            </Link>
          </Card>
        </section>
      )}

      <section>
        <h2 className="text-xl font-semibold mb-4">All Courses</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {mockCourses.slice(0, 3).map((course) => (
            <Link
              key={course.id}
              href={`/course/${course.slug}`}
              className="group block rounded-xl border bg-card p-4 hover:shadow-md transition-all"
            >
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold text-sm mb-3"
                style={{ backgroundColor: course.coverColor }}
              >
                {course.title.slice(0, 2).toUpperCase()}
              </div>
              <h3 className="font-semibold group-hover:text-primary">
                {course.title}
              </h3>
              <p className="text-xs text-muted-foreground mt-1">
                {course.lessonCount} lessons · {course.difficulty}
              </p>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}
