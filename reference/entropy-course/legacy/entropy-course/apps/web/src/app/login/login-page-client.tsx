"use client";

import { useState } from "react";

import SignInForm from "@/components/sign-in-form";
import SignUpForm from "@/components/sign-up-form";
import { PRIMARY_COURSE_SLUG } from "@/lib/learning-paths";
import { mockCourses, mockLessons } from "@/lib/mock-data";

const previewChapters = [
  { number: "0", title: "Workspace Setup" },
  { number: "1", title: "Hello, JavaScript" },
  { number: "2", title: "Variables & Types", active: true },
  { number: "3", title: "Operators & Expressions" },
];

export default function LoginPageClient() {
  const [showSignIn, setShowSignIn] = useState(true);
  const course = mockCourses.find(({ slug }) => slug === PRIMARY_COURSE_SLUG) ?? mockCourses[0];
  const lessons = (mockLessons[PRIMARY_COURSE_SLUG] ?? []).slice(0, 4);
  const completedLessons = lessons.filter(({ completed }) => completed).length;

  return (
    <main className="course-app min-h-svh overflow-x-hidden bg-background px-5 py-8 text-foreground sm:px-8 lg:px-11 lg:py-12">
      <div className="mx-auto grid min-h-[calc(100svh-4rem)] w-full max-w-7xl items-center gap-10 lg:min-h-[calc(100svh-6rem)] lg:grid-cols-[minmax(0,1fr)_minmax(420px,500px)]">
        <section className="flex min-w-0 flex-col gap-8">
          <div className="flex items-center gap-2.5">
            <div className="grid size-[22px] place-items-center rounded-[5px] bg-foreground font-mono text-[13px] font-semibold text-background">
              C
            </div>
            <div className="font-serif text-[22px] leading-none">Crashcourse</div>
            <div className="ml-auto hidden font-mono text-[10.5px] uppercase tracking-[0.06em] text-[var(--course-ink-soft)] sm:block">
              v0.1
            </div>
          </div>

          <div className="max-w-[620px]">
            <p className="font-mono text-[11px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
              JavaScript workbook
            </p>
            <h1 className="mt-3 max-w-[11ch] font-serif text-[43px] leading-[0.95] tracking-normal text-foreground sm:max-w-none sm:text-[58px]">
              JavaScript, one lesson at a time.
            </h1>
            <p className="mt-5 max-w-[52ch] text-[15px] leading-7 text-muted-foreground">
              Pick up with the variables chapter and keep your practice history close.
            </p>
          </div>

          <div className="hidden gap-4 md:grid lg:grid-cols-[minmax(0,0.95fr)_minmax(260px,0.75fr)]">
            <section className="relative min-w-0 overflow-hidden rounded-lg bg-foreground p-5 text-background">
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_10%,oklch(0.55_0.14_38_/_0.28),transparent_42%)]" />
              <div className="relative">
                <div className="flex items-center justify-between gap-4">
                  <p className="font-mono text-[10.5px] uppercase tracking-[0.1em] text-background/60">
                    Course
                  </p>
                  <span className="font-mono text-[10.5px] text-background/55">
                    {completedLessons} / {course.lessonCount}
                  </span>
                </div>
                <h2 className="mt-2 font-serif text-[34px] italic leading-none tracking-normal">
                  {course.title}
                </h2>
                <p className="mt-3 max-w-[42ch] text-sm leading-6 text-background/70">
                  {course.description}
                </p>
                <div className="mt-5 flex gap-3 font-mono text-[11px] text-background/65">
                  <span>{course.lessonCount} lessons</span>
                  <span>~{course.estimatedHours}h</span>
                  <span className="ml-auto">{course.progress ?? 0}%</span>
                </div>
                <div className="mt-3 h-1 overflow-hidden rounded-full bg-background/15">
                  <div
                    className="h-full rounded-full bg-primary transition-all"
                    style={{ width: `${course.progress ?? 0}%` }}
                  />
                </div>
              </div>
            </section>

            <section className="min-w-0 rounded-[14px] border border-border bg-card p-4 shadow-[0_14px_40px_oklch(0.19_0.012_70_/_0.05)]">
              <div className="font-mono text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
                Curriculum
              </div>
              <div className="mt-3 flex flex-col gap-0.5">
                {previewChapters.map((chapter) => (
                  <div
                    className="group flex items-center gap-2.5 rounded-[5px] px-2.5 py-1.5 text-[13px] text-muted-foreground data-[active=true]:bg-accent data-[active=true]:text-accent-foreground"
                    data-active={chapter.active ? "true" : undefined}
                    key={chapter.number}
                  >
                    <span className="min-w-5 font-mono text-[10.5px] text-[var(--course-ink-soft)] group-data-[active=true]:text-primary">
                      {chapter.number}
                    </span>
                    <span className="flex-1">{chapter.title}</span>
                    <span className="size-1.5 rounded-full bg-[var(--course-border-strong)] group-data-[active=true]:bg-primary" />
                  </div>
                ))}
              </div>
            </section>
          </div>
        </section>

        <section className="mx-auto flex min-w-0 w-full max-w-full flex-col gap-3 sm:max-w-[500px] lg:justify-self-end">
          {showSignIn ? (
            <SignInForm onSwitchToSignUp={() => setShowSignIn(false)} />
          ) : (
            <SignUpForm onSwitchToSignIn={() => setShowSignIn(true)} />
          )}
          <p className="px-1 text-center font-mono text-[10.5px] uppercase tracking-[0.08em] text-[var(--course-ink-soft)]">
            Progress resumes after authentication
          </p>
        </section>
      </div>
    </main>
  );
}
