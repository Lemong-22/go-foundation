"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import type { Route } from "next";
import { Check, ChevronDown, Copy, LogOut, Play, RotateCcw, Settings } from "lucide-react";
import { useMemo, useState } from "react";

import { Button } from "@entropy-course/ui/components/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@entropy-course/ui/components/dropdown-menu";
import { cn } from "@entropy-course/ui/lib/utils";
import { authClient } from "@/lib/auth-client";
import type { CodeSnippet, Course, Lesson, QuizQuestion } from "@/lib/mock-data";

type TabId = "concept" | "quiz" | "practice" | "cheatsheet";

interface CheatsheetSection {
  title: string;
  content: Array<{
    label: string;
    example: string;
  }>;
}

interface LessonExperienceProps {
  course: Course;
  lessons: Lesson[];
  lesson: Lesson;
  courseSlug: string;
  prevLesson?: Lesson;
  nextLesson?: Lesson;
  quizQuestions: QuizQuestion[];
  codeSnippets: CodeSnippet[];
  cheatsheet: {
    title: string;
    sections: CheatsheetSection[];
  };
}

const tabs: Array<{ id: TabId; label: string; meta: string }> = [
  { id: "concept", label: "Concept", meta: "3:24" },
  { id: "quiz", label: "Quiz", meta: "5q" },
  { id: "practice", label: "Practice", meta: "5 tests" },
  { id: "cheatsheet", label: "Cheatsheet", meta: "" },
];

const chapters = [
  { number: "0", title: "Workspace Setup" },
  { number: "1", title: "Hello, JavaScript" },
  { number: "2", title: "Variables & Types", open: true },
  { number: "3", title: "Operators & Expressions" },
  { number: "4", title: "Control Flow" },
  { number: "5", title: "Functions" },
  { number: "6", title: "Arrays" },
  { number: "7", title: "Objects" },
  { number: "8", title: "Async & Promises" },
];

const keyTerms = [
  {
    name: "scope",
    description:
      "The region of code where a name is visible. let and const are block-scoped; var is function-scoped.",
  },
  {
    name: "hoisting",
    description:
      "Declarations move to the top of their scope at compile time. var hoists with undefined; let and const enter a temporal dead zone.",
  },
  {
    name: "TDZ",
    description:
      "Temporal dead zone: the gap between a let or const hoist and its declaration line, where reading it throws.",
  },
  {
    name: "shadowing",
    description:
      "A name declared in an inner scope hides one with the same name from an outer scope.",
  },
];

const learningPaths = [
  { label: "Beginner", recipe: "Watch the video, do the quiz, copy snippet 1." },
  { label: "Switcher", recipe: "Skim cheatsheet, then practice." },
  { label: "Crammer", recipe: "Quiz, practice, cheatsheet." },
];

export function LessonExperience({
  course,
  lessons,
  lesson,
  courseSlug,
  prevLesson,
  nextLesson,
  quizQuestions,
  codeSnippets,
  cheatsheet,
}: LessonExperienceProps) {
  const [activeTab, setActiveTab] = useState<TabId>("concept");

  const lessonProgress = lesson.slug === "variables-and-types" ? "7 / 27" : `${lesson.order} / ${course.lessonCount}`;

  return (
    <div className="course-app min-h-svh min-w-0 overflow-x-hidden bg-background text-foreground lg:grid lg:grid-cols-[296px_minmax(0,1fr)]">
      <CourseSidebar course={course} courseSlug={courseSlug} lessons={lessons} activeLesson={lesson} />
      <main className="w-full min-w-0 px-5 py-7 sm:px-8 lg:max-w-[1100px] lg:px-11 lg:pb-20">
        <nav className="flex flex-wrap items-center gap-2 font-mono text-[11.5px] tracking-[0.04em] text-[var(--course-ink-soft)]">
          <Link href={`/course/${courseSlug}`} className="hover:text-foreground">
            {course.title}
          </Link>
          <span>›</span>
          <span className="font-medium text-muted-foreground">Chapter 2 · Variables & Types</span>
          <span>›</span>
          <span className="text-primary">§2.1</span>
        </nav>

        <header className="mt-2 flex flex-col gap-2">
          <h1 className="font-serif text-5xl leading-[0.95] tracking-normal text-foreground sm:text-[44px]">
            {lesson.title}
          </h1>
          <p className="max-w-[650px] text-base leading-7 text-muted-foreground">{lesson.description}</p>
        </header>

        <div className="mt-9 flex gap-1 overflow-x-auto border-b border-border">
          {tabs.map((tab, index) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                "mb-[-1px] flex shrink-0 items-center gap-2 border-b-2 border-transparent px-4 pb-3 pt-2.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground",
                activeTab === tab.id && "border-primary text-foreground",
              )}
            >
              <span
                className={cn(
                  "grid size-[22px] place-items-center rounded-[5px] bg-muted font-mono text-[11px] font-semibold text-muted-foreground",
                  activeTab === tab.id && "bg-primary text-primary-foreground",
                )}
              >
                {index + 1}
              </span>
              {tab.label}
              {tab.meta ? (
                <span
                  className={cn(
                    "ml-1 font-mono text-[10.5px] text-[var(--course-ink-soft)]",
                    activeTab === tab.id && "text-primary",
                  )}
                >
                  {tab.meta}
                </span>
              ) : null}
            </button>
          ))}
        </div>

        <section className="mt-7 animate-in fade-in slide-in-from-bottom-1 duration-300">
          {activeTab === "concept" ? <ConceptPanel lesson={lesson} /> : null}
          {activeTab === "quiz" ? <QuizPanel questions={quizQuestions} /> : null}
          {activeTab === "practice" ? <PracticePanel /> : null}
          {activeTab === "cheatsheet" ? (
            <CheatsheetPanel cheatsheet={cheatsheet} codeSnippets={codeSnippets} />
          ) : null}
        </section>

        <footer className="mt-12 flex flex-col gap-5 border-t border-border pt-6 sm:grid sm:grid-cols-[1fr_auto_1fr] sm:items-center">
          {prevLesson ? (
            <LessonNavLink href={`/course/${courseSlug}/lesson/${prevLesson.slug}`} label="Previous" name={prevLesson.title} />
          ) : (
            <span />
          )}
          <div className="text-center font-mono text-[11.5px] text-[var(--course-ink-soft)]">{lessonProgress}</div>
          {nextLesson ? (
            <LessonNavLink
              href={`/course/${courseSlug}/lesson/${nextLesson.slug}`}
              label="Next"
              name={nextLesson.title}
              align="right"
            />
          ) : (
            <LessonNavLink href={`/course/${courseSlug}`} label="Course" name="Back to overview" align="right" />
          )}
        </footer>
      </main>
    </div>
  );
}

function CourseSidebar({
  course,
  courseSlug,
  lessons,
  activeLesson,
}: {
  course: Course;
  courseSlug: string;
  lessons: Lesson[];
  activeLesson: Lesson;
}) {
  const router = useRouter();
  const lessonLinks = useMemo(
    () =>
      lessons.slice(0, 4).map((lessonItem, index) => ({
        number: `2.${index + 1}`,
        lesson: lessonItem,
      })),
    [lessons],
  );

  const openSettings = () => {
    router.push("/settings" as Route);
  };

  const logOut = () => {
    void authClient.signOut({
      fetchOptions: {
        onSuccess: () => {
          router.replace("/login");
          router.refresh();
        },
      },
    });
  };

  return (
    <aside className="min-w-0 border-b border-border bg-background px-5 py-7 lg:sticky lg:top-0 lg:h-svh lg:overflow-y-auto lg:border-b-0 lg:border-r">
      <div className="flex items-center gap-2.5 border-b border-border pb-4">
        <Link
          href="/"
          className="grid size-[22px] place-items-center rounded-[5px] bg-foreground font-mono text-[13px] font-semibold text-background"
        >
          C
        </Link>
        <div className="font-serif text-[22px] leading-none">Crashcourse</div>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                aria-label="Open Crashcourse menu"
                className="text-[var(--course-ink-soft)] hover:text-foreground"
                size="icon-xs"
                type="button"
                variant="ghost"
              />
            }
          >
            <ChevronDown aria-hidden="true" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="bg-card">
            <DropdownMenuGroup>
              <DropdownMenuItem onClick={openSettings}>
                <Settings aria-hidden="true" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuItem onClick={logOut} variant="destructive">
                <LogOut aria-hidden="true" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
        <div className="ml-auto font-mono text-[10.5px] uppercase tracking-[0.06em] text-[var(--course-ink-soft)]">
          v0.1
        </div>
      </div>

      <div className="mt-5 w-full max-w-full overflow-hidden rounded-lg bg-foreground p-4 text-background">
        <div className="font-mono text-[10.5px] uppercase tracking-[0.1em] text-background/60">Course</div>
        <h2 className="mt-1 font-serif text-3xl italic leading-none tracking-normal">{course.title}</h2>
        <div className="mt-4 flex gap-3 font-mono text-[11px] text-background/65">
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

      <div className="mt-7 font-mono text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
        Curriculum
      </div>
      <div className="mt-3 flex flex-col gap-1">
        {chapters.map((chapter) => (
          <div key={chapter.number}>
            <div className="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-sm font-medium">
              <span className="min-w-5 font-mono text-[11px] text-[var(--course-ink-soft)]">{chapter.number}</span>
              <span className="flex-1">{chapter.title}</span>
              <span className="font-mono text-[10px] text-[var(--course-ink-soft)]">›</span>
            </div>
            {chapter.open ? (
              <div className="ml-[22px] mt-1 flex flex-col gap-0.5 border-l border-border pl-2.5">
                {lessonLinks.map(({ number, lesson }) => {
                  const isActive = lesson.slug === activeLesson.slug;
                  return (
                    <Link
                      key={lesson.slug}
                      href={`/course/${courseSlug}/lesson/${lesson.slug}`}
                      className={cn(
                        "flex items-center gap-2.5 rounded-[5px] px-2.5 py-1.5 text-[13px] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
                        isActive && "bg-accent text-accent-foreground",
                      )}
                    >
                      <span className={cn("min-w-6 font-mono text-[10.5px] text-[var(--course-ink-soft)]", isActive && "text-primary")}>
                        {number}
                      </span>
                      <span className="flex-1">{lesson.title}</span>
                      <span
                        className={cn(
                          "size-1.5 rounded-full bg-[var(--course-border-strong)]",
                          isActive && "bg-primary shadow-[0_0_0_3px_var(--accent)]",
                          lesson.completed && !isActive && "bg-[var(--course-success)]",
                        )}
                      />
                    </Link>
                  );
                })}
              </div>
            ) : null}
          </div>
        ))}
      </div>
      <p className="mt-8 text-xs text-muted-foreground">Built for fast, durable learning.</p>
    </aside>
  );
}

function ConceptPanel({ lesson }: { lesson: Lesson }) {
  return (
    <div className="grid gap-8 xl:grid-cols-[1fr_280px]">
      <div>
        <div className="relative mb-7 aspect-video overflow-hidden rounded-[14px] border border-border bg-[linear-gradient(135deg,oklch(0.98_0.045_68)_0%,oklch(0.93_0.06_57)_55%,oklch(0.88_0.07_52)_100%)]">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_40%,oklch(0.55_0.14_38_/_0.18),transparent_50%),radial-gradient(circle_at_75%_70%,oklch(0.997_0.005_84_/_0.6),transparent_50%)]" />
          <div className="absolute inset-0 grid place-items-center font-serif text-[88px] italic leading-none text-primary/45">
            let
          </div>
          <button
            type="button"
            aria-label="Play concept video"
            className="absolute left-1/2 top-1/2 grid size-14 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full bg-foreground text-background transition-transform hover:scale-105"
          >
            <Play data-icon="inline-start" className="ml-1 fill-current" />
          </button>
          <div className="absolute inset-x-5 bottom-5 flex items-end justify-between">
            <span className="rounded-full bg-foreground px-3 py-1.5 font-mono text-[11px] text-background">
              <span className="mr-2 inline-block size-1.5 rounded-full bg-primary" />
              60-180s explainer
            </span>
            <span className="rounded-[5px] bg-foreground px-2.5 py-1.5 font-mono text-[11px] text-background">
              {lesson.duration}
            </span>
          </div>
        </div>

        <div className="max-w-[68ch] text-base leading-7">
          <p>
            JavaScript gives you three keywords for declaring variables: <InlineCode>let</InlineCode>,{" "}
            <InlineCode>const</InlineCode>, and the older <InlineCode>var</InlineCode>. The rules differ in{" "}
            <span className="font-medium text-accent-foreground">scope</span>,{" "}
            <span className="font-medium text-accent-foreground">mutability</span>, and{" "}
            <span className="font-medium text-accent-foreground">hoisting</span>, and getting them wrong is one of
            the most common bugs in early code.
          </p>
          <p className="mt-4">
            The modern default: reach for <InlineCode>const</InlineCode> first. Only switch to{" "}
            <InlineCode>let</InlineCode> when you genuinely need to reassign. <InlineCode>var</InlineCode> exists
            for legacy code, and you should almost never write it in new code.
          </p>

          <div className="my-6 rounded-lg border border-primary/25 bg-accent p-5">
            <div className="flex gap-4">
              <div className="font-serif text-2xl leading-none text-primary">¶</div>
              <div className="text-[14.5px] leading-6">
                <div className="mb-1 font-semibold text-accent-foreground">Mental model</div>
                <b className="text-accent-foreground">const</b> locks the <i>binding</i>, not the contents.{" "}
                <InlineCode>const arr = [1,2,3]</InlineCode> stops you reassigning <InlineCode>arr</InlineCode>,
                but you can still <InlineCode>arr.push(4)</InlineCode>. To freeze the contents, use{" "}
                <InlineCode>Object.freeze()</InlineCode>.
              </div>
            </div>
          </div>

          <h2 className="mt-8 font-serif text-[26px] leading-tight tracking-normal">Why const-first?</h2>
          <p className="mt-3">
            A <InlineCode>const</InlineCode> declaration is a contract with future-you: <i>this binding will not
            change</i>. When you read the file later, you can skim past every <InlineCode>const</InlineCode>{" "}
            knowing its value is stable. <InlineCode>let</InlineCode> signals that this variable changes, so pay
            attention here.
          </p>
        </div>
      </div>

      <RightRail />
    </div>
  );
}

function RightRail() {
  return (
    <aside className="flex flex-col gap-3 xl:sticky xl:top-7">
      <section className="rounded-lg border border-border bg-card p-5">
        <h3 className="mb-3 font-mono text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
          Key terms
        </h3>
        <div className="flex flex-col">
          {keyTerms.map((term) => (
            <div key={term.name} className="border-b border-border py-3 first:pt-0 last:border-b-0 last:pb-0">
              <div className="mb-1 font-mono text-[12.5px] text-accent-foreground">{term.name}</div>
              <p className="text-[13px] leading-5 text-muted-foreground">{term.description}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="rounded-lg border border-border bg-card p-5">
        <h3 className="mb-3 font-mono text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
          Path through this lesson
        </h3>
        <div className="flex flex-col gap-2">
          {learningPaths.map((path) => (
            <div key={path.label} className="flex items-center gap-3 rounded-md bg-muted px-2.5 py-2 text-[13px]">
              <span className="min-w-[76px] font-mono text-[10.5px] uppercase tracking-[0.05em] text-[var(--course-ink-soft)]">
                {path.label}
              </span>
              <span className="text-[12.5px] leading-5 text-muted-foreground">{path.recipe}</span>
            </div>
          ))}
        </div>
      </section>
    </aside>
  );
}

function QuizPanel({ questions }: { questions: QuizQuestion[] }) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [submitted, setSubmitted] = useState(false);

  const question = questions[currentIndex] ?? questions[0];
  const selected = answers[question.id];
  const isCorrect = selected === question.correctAnswer;

  return (
    <div>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="font-mono text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
            Checkpoint
          </div>
          <h2 className="mt-1 font-serif text-[32px] leading-tight tracking-normal">Scope and reassignment</h2>
        </div>
        <div className="flex gap-1.5">
          {questions.map((item, index) => (
            <button
              key={item.id}
              type="button"
              aria-label={`Question ${index + 1}`}
              onClick={() => {
                setCurrentIndex(index);
                setSubmitted(false);
              }}
              className={cn(
                "h-1 w-8 rounded-full bg-border transition-colors",
                index === currentIndex && "bg-primary",
                answers[item.id] && index !== currentIndex && "bg-[var(--course-success)]",
              )}
            />
          ))}
        </div>
      </div>

      <div className="rounded-[14px] border border-border bg-card p-7">
        <div className="mb-2 font-mono text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">
          {question.type.replace("-", " ")}
        </div>
        <p className="whitespace-pre-wrap font-serif text-[26px] leading-snug tracking-normal">{question.question}</p>
        <p className="mb-5 mt-1 text-sm text-muted-foreground">Choose the best answer before moving on.</p>
        <div className="flex flex-col gap-2.5">
          {(question.options ?? []).map((option, index) => {
            const optionIsSelected = selected === option;
            const optionIsCorrect = submitted && option === question.correctAnswer;
            const optionIsWrong = submitted && optionIsSelected && !optionIsCorrect;

            return (
              <button
                key={option}
                type="button"
                disabled={submitted}
                onClick={() => setAnswers((previous) => ({ ...previous, [question.id]: option }))}
                className={cn(
                  "flex w-full items-start gap-3 rounded-lg border border-border bg-card px-4 py-3.5 text-left text-[14.5px] transition-colors hover:bg-muted disabled:cursor-default",
                  optionIsSelected && "border-foreground",
                  optionIsCorrect && "border-[var(--course-success)] bg-[var(--course-success-soft)]",
                  optionIsWrong && "border-[var(--course-danger)] bg-[var(--course-danger-soft)]",
                )}
              >
                <span
                  className={cn(
                    "grid size-[22px] shrink-0 place-items-center rounded-full border border-[var(--course-border-strong)] font-mono text-[11px] text-muted-foreground",
                    optionIsSelected && "border-foreground bg-foreground text-background",
                    optionIsCorrect && "border-[var(--course-success)] bg-[var(--course-success)] text-primary-foreground",
                    optionIsWrong && "border-[var(--course-danger)] bg-[var(--course-danger)] text-primary-foreground",
                  )}
                >
                  {String.fromCharCode(65 + index)}
                </span>
                <span>{option}</span>
              </button>
            );
          })}
        </div>

        {submitted ? (
          <div
            className={cn(
              "mt-5 rounded-lg border p-4 text-sm leading-6",
              isCorrect
                ? "border-[var(--course-success)] bg-[var(--course-success-soft)]"
                : "border-[var(--course-danger)] bg-[var(--course-danger-soft)]",
            )}
          >
            <div
              className={cn(
                "mb-1 font-mono text-xs uppercase tracking-[0.06em]",
                isCorrect ? "text-[var(--course-success)]" : "text-[var(--course-danger)]",
              )}
            >
              {isCorrect ? "Correct" : "Review this"}
            </div>
            {question.explanation}
          </div>
        ) : null}

        <div className="mt-6 flex items-center justify-between">
          <span className="font-mono text-[11.5px] text-[var(--course-ink-soft)]">
            {currentIndex + 1} / {questions.length}
          </span>
          <button
            type="button"
            disabled={!selected}
            onClick={() => setSubmitted(true)}
            className="inline-flex items-center gap-2 rounded-md border border-foreground bg-foreground px-3.5 py-2 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50"
          >
            <Check data-icon="inline-start" />
            Check answer
          </button>
        </div>
      </div>
    </div>
  );
}

function PracticePanel() {
  const [code, setCode] = useState("const amount = 3;\nlet total = amount;\ntotal += 4;\nconsole.log(total);");
  const [output, setOutput] = useState("Ready.");

  return (
    <div className="grid gap-5 lg:grid-cols-2">
      <section className="flex min-h-[540px] flex-col rounded-[14px] border border-border bg-card p-6">
        <span className="w-fit rounded-full bg-muted px-2.5 py-1 font-mono text-[10.5px] uppercase tracking-[0.05em] text-muted-foreground">
          Medium
        </span>
        <h2 className="mt-3 font-serif text-[30px] leading-tight tracking-normal">Refactor to const-first</h2>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          Start with the most stable declarations. Switch to let only when the value changes during the program.
        </p>
        <div className="mt-5 rounded-lg border border-border bg-muted p-4 font-mono text-[13px] leading-6">
          <div className="mb-1 text-[10.5px] uppercase tracking-[0.1em] text-[var(--course-ink-soft)]">Example</div>
          input: [1, 2, 3]
          <br />
          output: stable bindings first
        </div>
        <div className="mt-auto flex flex-wrap gap-2 border-t border-border pt-4">
          <button className="rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted" type="button">
            Show hint
          </button>
          <button className="rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium hover:bg-muted" type="button">
            Show solution
          </button>
        </div>
      </section>

      <section className="flex min-h-[540px] flex-col overflow-hidden rounded-[14px] border border-border bg-card">
        <div className="flex items-center justify-between border-b border-border bg-muted px-4 py-2.5">
          <div className="flex gap-1 font-mono text-[11.5px] text-muted-foreground">
            <span className="rounded-[5px] bg-card px-2.5 py-1 text-foreground shadow-sm">solution.js</span>
            <span className="px-2.5 py-1">tests.js</span>
          </div>
          <button
            type="button"
            onClick={() => setCode("const amount = 3;\nlet total = amount;\ntotal += 4;\nconsole.log(total);")}
            className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2.5 py-1 text-xs hover:bg-muted"
          >
            <RotateCcw data-icon="inline-start" />
            Reset
          </button>
        </div>
        <textarea
          value={code}
          onChange={(event) => setCode(event.target.value)}
          className="min-h-[300px] flex-1 resize-none bg-card p-5 font-mono text-[13.5px] leading-7 outline-none"
          spellCheck={false}
        />
        <div className="flex items-center gap-2 border-t border-border bg-background px-4 py-3">
          <button
            type="button"
            onClick={() => setOutput("✓ declares stable values with const\n✓ uses let only for reassignment\n✓ logs 7")}
            className="rounded-md border border-primary bg-primary px-3.5 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Run checks
          </button>
          <span className="font-mono text-[11.5px] text-[var(--course-ink-soft)]">local runner</span>
        </div>
        <pre className="max-h-[180px] overflow-auto bg-foreground p-4 font-mono text-[12.5px] leading-6 text-background">
          {output}
        </pre>
      </section>
    </div>
  );
}

function CheatsheetPanel({
  cheatsheet,
  codeSnippets,
}: {
  cheatsheet: { title: string; sections: CheatsheetSection[] };
  codeSnippets: CodeSnippet[];
}) {
  const [copied, setCopied] = useState<string | null>(null);

  const copySnippet = async (snippet: CodeSnippet) => {
    await navigator.clipboard.writeText(snippet.code);
    setCopied(snippet.id);
    window.setTimeout(() => setCopied(null), 1500);
  };

  return (
    <div className="grid grid-cols-1 gap-5 lg:grid-cols-6">
      <section className="rounded-[14px] border border-border bg-card p-5 lg:col-span-3">
        <div className="mb-4 flex items-baseline justify-between">
          <h2 className="font-serif text-[26px] leading-tight tracking-normal">{cheatsheet.title}</h2>
          <span className="font-mono text-[11px] text-[var(--course-ink-soft)]">01</span>
        </div>
        <div className="flex flex-col gap-5">
          {cheatsheet.sections.map((section) => (
            <div key={section.title}>
              <h3 className="mb-2 font-mono text-[10.5px] uppercase tracking-[0.08em] text-[var(--course-ink-soft)]">
                {section.title}
              </h3>
              <div className="overflow-hidden rounded-lg border border-border">
                {section.content.map((item) => (
                  <div key={item.label} className="grid grid-cols-[90px_1fr] border-b border-border px-3 py-2 last:border-b-0">
                    <span className="font-mono text-[12.5px] text-accent-foreground">{item.label}</span>
                    <code className="font-mono text-xs text-muted-foreground">{item.example}</code>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="rounded-[14px] border border-border bg-card p-5 lg:col-span-3">
        <div className="mb-4 flex items-baseline justify-between">
          <h2 className="font-serif text-[26px] leading-tight tracking-normal">Copy snippet</h2>
          <span className="font-mono text-[11px] text-[var(--course-ink-soft)]">02</span>
        </div>
        <div className="flex flex-col gap-4">
          {codeSnippets.slice(0, 2).map((snippet) => (
            <div key={snippet.id} className="overflow-hidden rounded-lg border border-border">
              <div className="flex items-center justify-between border-b border-border bg-muted px-3 py-2.5">
                <div>
                  <div className="text-sm font-medium">{snippet.title}</div>
                  <div className="text-xs text-muted-foreground">{snippet.description}</div>
                </div>
                <button
                  type="button"
                  onClick={() => copySnippet(snippet)}
                  className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground hover:bg-foreground hover:text-background"
                >
                  <Copy data-icon="inline-start" />
                  {copied === snippet.id ? "Copied" : "Copy"}
                </button>
              </div>
              <pre className="max-h-[260px] overflow-auto bg-card p-4 font-mono text-[12.5px] leading-6 text-foreground">
                <code>{snippet.code}</code>
              </pre>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}

function LessonNavLink({
  href,
  label,
  name,
  align = "left",
}: {
  href: string;
  label: string;
  name: string;
  align?: "left" | "right";
}) {
  return (
    <Link
      href={href as Route}
      className={cn(
        "flex min-w-[200px] flex-col gap-0.5 rounded-lg border border-border bg-card px-3.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
        align === "right" && "items-end text-right sm:justify-self-end",
      )}
    >
      <span className="font-mono text-[10.5px] uppercase tracking-[0.08em] text-[var(--course-ink-soft)]">
        {align === "left" ? "← " : ""}
        {label}
        {align === "right" ? " →" : ""}
      </span>
      <span className="font-medium text-foreground">{name}</span>
    </Link>
  );
}

function InlineCode({ children }: { children: React.ReactNode }) {
  return (
    <code className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[0.88em] text-foreground">
      {children}
    </code>
  );
}
