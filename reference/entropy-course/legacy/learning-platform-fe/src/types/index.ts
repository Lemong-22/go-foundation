export type LessonStatus = "completed" | "current" | "locked";

export interface Lesson {
  id: string;
  title: string;
  status: LessonStatus;
}

export interface Module {
  id: string;
  title: string;
  subtitle?: string;
  isExpanded: boolean;
  lessons: Lesson[];
}

export interface Course {
  id: string;
  title: string;
  subtitle: string;
  modules: Module[];
}

export interface LessonDetail {
  id: string;
  title: string;
  description: string;
  duration: string;
  xp: number;
  breadcrumbs: string[];
  activeTab: "concept" | "quiz" | "practice";
  videoTitle: string;
  videoDescription: string;
  keyTakeaways: { title: string; description: string }[];
}

export interface ModuleProgressData {
  percentage: number;
  completed: number;
  total: number;
}
