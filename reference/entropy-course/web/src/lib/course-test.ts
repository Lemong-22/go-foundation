import type { CourseView } from "@/lib/course-api/server";

export type CourseTestProjection = {
  badge: "Sample data";
  kind: "test";
  sections: CourseTestSection[];
  summary: string;
  title: string;
};

export type CourseTestSection = {
  choices?: string[];
  id: string;
  label: string;
  prompt: string;
  reviewPoints?: string[];
  starterCode?: string;
};

export function buildDummyCourseTest(
  course: Pick<CourseView, "Title">,
): CourseTestProjection {
  return {
    badge: "Sample data",
    kind: "test",
    sections: [
      {
        choices: [
          "Break a problem into named values.",
          "Read code from inputs to outputs.",
          "Describe a small program in plain language.",
        ],
        id: "concept-check",
        label: "Concept check",
        prompt: "Select the statement that best matches the main course idea.",
      },
      {
        id: "code-reading",
        label: "Code reading",
        prompt:
          "Trace a tiny snippet and describe the value that moves through it.",
        starterCode:
          'const value = "sample";\nconst label = value.toUpperCase();\nlabel;',
      },
      {
        id: "practice-plan",
        label: "Practice plan",
        prompt: "Outline how you would solve a small course-level task.",
        reviewPoints: [
          "Name the inputs.",
          "Write the steps before code.",
          "Try a small example.",
        ],
      },
    ],
    summary: `A placeholder course-level test shell for ${course.Title}.`,
    title: "Sample course check",
  };
}

export function courseTestItemCountLabel(count: number) {
  return `${count} ${count === 1 ? "prompt" : "prompts"}`;
}
