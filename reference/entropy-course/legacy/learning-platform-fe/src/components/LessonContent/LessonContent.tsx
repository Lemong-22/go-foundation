import { useState } from "react";

import { LessonDetail } from "../../types";

import "./LessonContent.css";

interface LessonContentProps {
  lesson: LessonDetail;
}

type TabType = "concept" | "quiz" | "practice";

export function LessonContent({ lesson }: LessonContentProps) {
  const [activeTab, setActiveTab] = useState<TabType>(lesson.activeTab);
  const lessonIdParts = lesson.id.match(/lesson-(\d+)-(\d+)/);
  const chapterNumber = lessonIdParts?.[1] ?? "";
  const lessonNumber = lessonIdParts?.[2] ? Number(lessonIdParts[2]) : undefined;

  const toRoman = (value?: number) => {
    if (!value || value <= 0) return "";
    const numerals: [number, string][] = [
      [1000, "M"],
      [900, "CM"],
      [500, "D"],
      [400, "CD"],
      [100, "C"],
      [90, "XC"],
      [50, "L"],
      [40, "XL"],
      [10, "X"],
      [9, "IX"],
      [5, "V"],
      [4, "IV"],
      [1, "I"],
    ];
    let remaining = value;
    let result = "";
    for (const [num, numeral] of numerals) {
      while (remaining >= num) {
        result += numeral;
        remaining -= num;
      }
    }
    return result;
  };

  const lessonRoman = toRoman(lessonNumber);

  const tabs: { id: TabType; label: string; icon: string }[] = [
    { id: "concept", label: "Concept", icon: "💡" },
    { id: "quiz", label: "Quiz", icon: "📝" },
    { id: "practice", label: "Practice", icon: "<>" },
  ];

  return (
    <main className="lesson-content">
      <div className="lesson-content__header">
        <div className="lesson-content__main">
          <div className="lesson-content__title-row">
            <span
              className="lesson-content__chapter-badge"
              aria-label={`Chapter ${chapterNumber} Lesson ${lessonRoman}`}
            >
              {chapterNumber} - {lessonRoman}
            </span>
            <h1 className="lesson-content__title">{lesson.title}</h1>
          </div>

          <nav className="lesson-content__breadcrumbs" aria-label="Breadcrumb">
            <div className="lesson-content__breadcrumbs-list">
              {lesson.breadcrumbs.map((crumb, index) => (
                <span className="lesson-content__breadcrumb-item" key={index}>
                  <a
                    href="#"
                    className={`lesson-content__breadcrumb-link ${
                      index === lesson.breadcrumbs.length - 1
                        ? "lesson-content__breadcrumb-link--current"
                        : ""
                    }`}
                  >
                    {crumb}
                  </a>
                </span>
              ))}
            </div>
          </nav>
        </div>
      </div>

      <div className="lesson-content__tabs">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            className={`lesson-content__tab ${activeTab === tab.id ? "lesson-content__tab--active" : ""}`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span className="lesson-content__tab-icon">{tab.icon}</span>
            {tab.label}
          </button>
        ))}
      </div>

      <div className="lesson-content__video-section">
        <button className="lesson-content__play-button" aria-label="Play video">
          ▶
        </button>
        <h2 className="lesson-content__video-title">{lesson.videoTitle}</h2>
        <p className="lesson-content__video-description">{lesson.videoDescription}</p>
      </div>
    </main>
  );
}
