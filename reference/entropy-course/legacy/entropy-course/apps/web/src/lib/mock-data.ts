// Mock data for the Crashcourse frontend
// This will be replaced with real API calls when backend is ready

export interface Course {
  id: string;
  slug: string;
  title: string;
  description: string;
  subject: string;
  difficulty: "Beginner" | "Intermediate" | "Advanced";
  estimatedHours: number;
  lessonCount: number;
  coverColor: string;
  progress?: number; // 0-100
  started?: boolean;
  completed?: boolean;
}

export interface Lesson {
  id: string;
  slug: string;
  courseSlug: string;
  title: string;
  description: string;
  duration: string; // e.g., "5 min"
  order: number;
  completed?: boolean;
  modules: LessonModule[];
}

export interface LessonModule {
  type: "video" | "cheatsheet" | "code" | "quiz" | "practice";
  title: string;
  completed?: boolean;
}

export interface QuizQuestion {
  id: string;
  type: "multiple-choice" | "multi-select" | "fill-blank" | "predict-output";
  question: string;
  options?: string[];
  correctAnswer: string | string[];
  explanation: string;
}

export interface CodeSnippet {
  id: string;
  title: string;
  description: string;
  language: string;
  code: string;
}

// Mock courses
export const mockCourses: Course[] = [
  {
    id: "1",
    slug: "javascript",
    title: "JavaScript",
    description:
      "Master the fundamentals of JavaScript, from variables and functions to async programming and the DOM. Build real projects as you learn.",
    subject: "Programming",
    difficulty: "Beginner",
    estimatedHours: 4.5,
    lessonCount: 28,
    coverColor: "#B85C3D",
    progress: 0,
    started: true,
  },
  {
    id: "2",
    slug: "python",
    title: "Python",
    description:
      "Learn Python from scratch. Cover data types, control flow, functions, OOP, and common libraries. Perfect for beginners and anyone switching to Python.",
    subject: "Programming",
    difficulty: "Beginner",
    estimatedHours: 8,
    lessonCount: 15,
    coverColor: "#3776AB",
  },
  {
    id: "3",
    slug: "sql",
    title: "SQL",
    description:
      "Query databases like a pro. From SELECT statements to complex JOINs and aggregations — everything you need to work with data.",
    subject: "Data",
    difficulty: "Beginner",
    estimatedHours: 4,
    lessonCount: 8,
    coverColor: "#4479A1",
  },
  {
    id: "4",
    slug: "react",
    title: "React",
    description:
      "Build modern web applications with React. Components, hooks, state management, and best practices for production-ready apps.",
    subject: "Frontend",
    difficulty: "Intermediate",
    estimatedHours: 10,
    lessonCount: 18,
    coverColor: "#61DAFB",
  },
  {
    id: "5",
    slug: "git",
    title: "Git",
    description:
      "Version control made simple. Branching, merging, rebasing, and collaborating with GitHub. Stop being afraid of the command line.",
    subject: "Tools",
    difficulty: "Beginner",
    estimatedHours: 3,
    lessonCount: 6,
    coverColor: "#F05032",
  },
  {
    id: "6",
    slug: "html-css",
    title: "HTML & CSS",
    description:
      "Build beautiful websites from scratch. Semantic HTML, modern CSS layouts, Flexbox, Grid, and responsive design fundamentals.",
    subject: "Frontend",
    difficulty: "Beginner",
    estimatedHours: 5,
    lessonCount: 10,
    coverColor: "#E34F26",
  },
];

// Mock lessons for JavaScript course
export const mockLessons: Record<string, Lesson[]> = {
  javascript: [
    {
      id: "js-1",
      slug: "variables-and-types",
      courseSlug: "javascript",
      title: "let, const, var",
      description:
        "Three ways to declare a variable, but only two you should ever reach for. Understand block scope, mutability, and why var is a relic.",
      duration: "3:24",
      order: 1,
      completed: true,
      modules: [
        { type: "video", title: "Animated Explainer", completed: true },
        { type: "cheatsheet", title: "Cheat Sheet", completed: true },
        { type: "code", title: "Code Snippets", completed: true },
        { type: "quiz", title: "Quiz", completed: true },
        { type: "practice", title: "Practice", completed: true },
      ],
    },
    {
      id: "js-2",
      slug: "primitive-types",
      courseSlug: "javascript",
      title: "Primitive types",
      description: "Strings, numbers, booleans, null, undefined, symbols, and bigint.",
      duration: "5 min",
      order: 2,
      completed: true,
      modules: [
        { type: "video", title: "Animated Explainer", completed: true },
        { type: "cheatsheet", title: "Cheat Sheet", completed: true },
        { type: "code", title: "Code Snippets", completed: true },
        { type: "quiz", title: "Quiz", completed: true },
        { type: "practice", title: "Practice", completed: false },
      ],
    },
    {
      id: "js-3",
      slug: "type-coercion",
      courseSlug: "javascript",
      title: "Type coercion",
      description: "How JavaScript converts values, and when that conversion surprises you.",
      duration: "7 min",
      order: 3,
      modules: [
        { type: "video", title: "Animated Explainer", completed: false },
        { type: "cheatsheet", title: "Cheat Sheet", completed: false },
        { type: "code", title: "Code Snippets", completed: false },
        { type: "quiz", title: "Quiz", completed: false },
        { type: "practice", title: "Practice", completed: false },
      ],
    },
    {
      id: "js-4",
      slug: "template-literals",
      courseSlug: "javascript",
      title: "Template literals",
      description: "Readable string interpolation, multiline strings, and tagged templates.",
      duration: "6 min",
      order: 4,
      modules: [
        { type: "video", title: "Animated Explainer", completed: false },
        { type: "cheatsheet", title: "Cheat Sheet", completed: false },
        { type: "code", title: "Code Snippets", completed: false },
        { type: "quiz", title: "Quiz", completed: false },
        { type: "practice", title: "Practice", completed: false },
      ],
    },
    {
      id: "js-5",
      slug: "operators-and-expressions",
      courseSlug: "javascript",
      title: "Operators & Expressions",
      description: "Assignment, comparison, arithmetic, and expression evaluation.",
      duration: "14 min",
      order: 5,
      modules: [
        { type: "video", title: "Animated Explainer", completed: false },
        { type: "cheatsheet", title: "Cheat Sheet", completed: false },
        { type: "code", title: "Code Snippets", completed: false },
        { type: "quiz", title: "Quiz", completed: false },
        { type: "practice", title: "Practice", completed: false },
      ],
    },
  ],
};

// Mock quiz questions
export const mockQuizQuestions: QuizQuestion[] = [
  {
    id: "q1",
    type: "multiple-choice",
    question: "Which keyword declares a variable that cannot be reassigned?",
    options: ["var", "let", "const", "static"],
    correctAnswer: "const",
    explanation:
      "The `const` keyword declares a constant reference that cannot be reassigned. However, for objects and arrays, their properties can still be modified.",
  },
  {
    id: "q2",
    type: "multiple-choice",
    question: "What does `typeof null` return in JavaScript?",
    options: ["'null'", "'undefined'", "'object'", "'boolean'"],
    correctAnswer: "'object'",
    explanation:
      "This is a famous JavaScript quirk. `typeof null` returns 'object' due to a legacy bug in the language that was never fixed for backwards compatibility.",
  },
  {
    id: "q3",
    type: "predict-output",
    question: "What will be logged?\n```js\nconsole.log([] + []);\n```",
    options: ["'[]'", "'' (empty string)", "0", "NaN"],
    correctAnswer: "'' (empty string)",
    explanation:
      "When using `+` with arrays, JavaScript converts them to strings first. An empty array becomes an empty string, so `'' + ''` equals `''`.",
  },
];

// Mock code snippets
export const mockCodeSnippets: CodeSnippet[] = [
  {
    id: "snip-1",
    title: "Variable Declarations",
    description: "When to use var, let, and const",
    language: "javascript",
    code: `// Prefer const by default
const name = "Alice";
const count = 42;

// Use let when you need to reassign
let total = 0;
total += 10;

// Avoid var (function-scoped, hoisted)
// var is legacy — stick to const/let`,
  },
  {
    id: "snip-2",
    title: "Arrow Functions",
    description: "Modern function syntax with concise body",
    language: "javascript",
    code: `// Basic arrow function
const greet = (name) => \`Hello, \${name}!\`;

// Single expression (implicit return)
const double = (n) => n * 2;

// Multiple parameters
const add = (a, b) => a + b;

// Multi-line body needs explicit return
const fetchUser = async (id) => {
  const response = await fetch(\`/api/users/\${id}\`);
  return response.json();
};`,
  },
  {
    id: "snip-3",
    title: "Array Methods",
    description: "Map, filter, reduce — the holy trinity",
    language: "javascript",
    code: `const numbers = [1, 2, 3, 4, 5];

// map — transform each element
const doubled = numbers.map((n) => n * 2);
// [2, 4, 6, 8, 10]

// filter — keep elements matching condition
const evens = numbers.filter((n) => n % 2 === 0);
// [2, 4]

// reduce — accumulate into single value
const sum = numbers.reduce((acc, n) => acc + n, 0);
// 15

// Chaining for complex operations
const result = numbers
  .filter((n) => n > 2)
  .map((n) => n * 2);
// [6, 8, 10]`,
  },
];

// Mock cheatsheet content (simplified)
export const mockCheatsheetContent = {
  title: "JavaScript Variables & Types",
  sections: [
    {
      title: "Data Types",
      content: [
        { label: "String", example: "'hello' or \`template\`" },
        { label: "Number", example: "42, 3.14, NaN, Infinity" },
        { label: "Boolean", example: "true, false" },
        { label: "Null", example: "null (intentional empty)" },
        { label: "Undefined", example: "undefined (unassigned)" },
        { label: "Object", example: "{ key: 'value' }" },
        { label: "Array", example: "[1, 2, 3]" },
        { label: "Symbol", example: "Symbol('id')" },
      ],
    },
    {
      title: "Type Checking",
      content: [
        { label: "typeof", example: "typeof 42 → 'number'" },
        { label: "Array.isArray", example: "Array.isArray([]) → true" },
        { label: "instanceof", example: "[] instanceof Array → true" },
      ],
    },
  ],
};
