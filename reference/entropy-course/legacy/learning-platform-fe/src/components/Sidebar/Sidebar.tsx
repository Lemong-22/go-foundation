import { useState } from "react";

import { Course } from "../../types";
import { Header } from "../Header/Header";

import "./Sidebar.css";

interface SidebarProps {
  course: Course;
  currentLessonId: string;
}

export function Sidebar({ course, currentLessonId }: SidebarProps) {
  const [expandedModules, setExpandedModules] = useState<Set<string>>(
    new Set(course.modules.filter((m) => m.isExpanded).map((m) => m.id)),
  );

  const toggleModule = (moduleId: string) => {
    setExpandedModules((prev) => {
      const next = new Set(prev);
      if (next.has(moduleId)) {
        next.delete(moduleId);
      } else {
        next.add(moduleId);
      }
      return next;
    });
  };

  const toRoman = (num: number) => {
    const romanMap = [
      { value: 1000, numeral: "M" },
      { value: 900, numeral: "CM" },
      { value: 500, numeral: "D" },
      { value: 400, numeral: "CD" },
      { value: 100, numeral: "C" },
      { value: 90, numeral: "XC" },
      { value: 50, numeral: "L" },
      { value: 40, numeral: "XL" },
      { value: 10, numeral: "X" },
      { value: 9, numeral: "IX" },
      { value: 5, numeral: "V" },
      { value: 4, numeral: "IV" },
      { value: 1, numeral: "I" },
    ];

    let remainder = num;
    let result = "";

    for (const { value, numeral } of romanMap) {
      while (remainder >= value) {
        result += numeral;
        remainder -= value;
      }
    }

    return result;
  };

  return (
    <aside className="sidebar">
      <Header />
      <div className="sidebar__content">
        <div className="sidebar__modules">
          {course.modules.map((module, moduleIndex) => {
            const isExpanded = expandedModules.has(module.id);
            const hasCompletedAll =
              module.lessons.length > 0 && module.lessons.every((l) => l.status === "completed");
            const hasCurrent = module.lessons.some((l) => l.status === "current");
            const moduleStatusClass = hasCompletedAll
              ? "sidebar__module-icon--completed"
              : hasCurrent
                ? "sidebar__module-icon--current"
                : "sidebar__module-icon--locked";
            const moduleOrderLabel = String(moduleIndex + 1);
            const isActiveModule = module.lessons.some((lesson) => lesson.id === currentLessonId);

            return (
              <div
                key={module.id}
                className={`sidebar__module ${isActiveModule ? "sidebar__module--active" : ""}`}
              >
                <div className="sidebar__module-header" onClick={() => toggleModule(module.id)}>
                  <span className={`sidebar__module-icon ${moduleStatusClass}`}>
                    {moduleOrderLabel}
                  </span>
                  <div className="sidebar__module-info">
                    <span className="sidebar__module-title">{module.subtitle || module.title}</span>
                  </div>
                  <span
                    className={`sidebar__module-chevron ${isExpanded ? "sidebar__module-chevron--expanded" : ""}`}
                  >
                    ▶
                  </span>
                </div>

                {isExpanded && module.lessons.length > 0 && (
                  <div className="sidebar__lessons">
                    {module.lessons.map((lesson, lessonIndex) => {
                      const lessonRoman = toRoman(lessonIndex + 1);
                      const isCurrent = lesson.id === currentLessonId;

                      return (
                        <div
                          key={lesson.id}
                          className={`sidebar__lesson sidebar__lesson--${lesson.status} ${isCurrent ? "sidebar__lesson--current" : ""}`}
                        >
                          <span
                            className={`sidebar__lesson-icon sidebar__lesson-icon--${lesson.status}`}
                          >
                            {lessonRoman}
                          </span>
                          <span className="sidebar__lesson-title">{lesson.title}</span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </aside>
  );
}
