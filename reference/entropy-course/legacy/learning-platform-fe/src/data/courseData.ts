import { Course, LessonDetail, ModuleProgressData } from "../types";

export const courseData: Course = {
  id: "python-mastery",
  title: "Python Mastery",
  subtitle: "Beginner Course",
  modules: [
    {
      id: "module-1",
      title: "Module 1",
      subtitle: "Variables & Syntax",
      isExpanded: false,
      lessons: [],
    },
    {
      id: "module-2",
      title: "Module 2",
      subtitle: "Data Types",
      isExpanded: true,
      lessons: [
        { id: "lesson-2-1", title: "Introduction to Data Types", status: "completed" },
        { id: "lesson-2-2", title: "Integers vs. Floats", status: "completed" },
        { id: "lesson-2-3", title: "Numeric Operations Quiz", status: "completed" },
        { id: "lesson-2-4", title: "Visualizing Memory Allocation", status: "current" },
        { id: "lesson-2-5", title: "String Concatenation", status: "locked" },
        { id: "lesson-2-6", title: "Boolean Logic", status: "locked" },
      ],
    },
    {
      id: "module-3",
      title: "Module 3",
      subtitle: "Control Flow",
      isExpanded: false,
      lessons: [],
    },
    {
      id: "module-4",
      title: "Module 4",
      subtitle: "Functions",
      isExpanded: false,
      lessons: [],
    },
  ],
};

export const currentLessonData: LessonDetail = {
  id: "lesson-2-4",
  title: "Visualizing Memory Allocation",
  description:
    "Interactive Concept Illustration. Understand how data is stored in the stack and heap memory.",
  duration: "15 mins",
  xp: 50,
  breadcrumbs: ["Courses", "Python Mastery", "Data Types", "Visualizing Memory Allocation"],
  activeTab: "concept",
  videoTitle: "Interactive Memory Model",
  videoDescription: "Click to start the interactive simulation of Stack vs Heap memory allocation.",
  keyTakeaways: [
    {
      title: "Stack Memory",
      description:
        "Used for static memory allocation and execution of threads. It contains primitive",
    },
  ],
};

export const moduleProgressData: ModuleProgressData = {
  percentage: 35,
  completed: 3,
  total: 8,
};
