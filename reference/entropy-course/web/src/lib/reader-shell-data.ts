export type LessonLink = {
  id: string;
  number: string;
  title: string;
  duration: string;
  status: "complete" | "active" | "locked";
};

export const readerShellLessons: LessonLink[] = [
  {
    id: "runtime-and-setup",
    number: "01",
    title: "Runtime and setup",
    duration: "14 min",
    status: "complete",
  },
  {
    id: "values-and-types",
    number: "02",
    title: "Values and types",
    duration: "19 min",
    status: "active",
  },
  {
    id: "first-data-model",
    number: "03",
    title: "First data model",
    duration: "24 min",
    status: "locked",
  },
];
