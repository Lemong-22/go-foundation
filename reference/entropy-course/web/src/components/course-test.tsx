import type { Route } from "next";
import Link from "next/link";

import { ReaderRouteState } from "@/components/reader-route-state";
import type { CourseApiError, CourseView } from "@/lib/course-api/server";
import {
  courseTestItemCountLabel,
  type CourseTestProjection,
  type CourseTestSection,
} from "@/lib/course-test";

type CourseTestProps = {
  course: CourseView;
  test: CourseTestProjection;
};

type CourseTestErrorProps = {
  course?: CourseView;
  error: CourseApiError;
};

export function CourseTest({ course, test }: CourseTestProps) {
  const syllabusPath = `/courses/${encodeURIComponent(course.ID)}` as Route;

  return (
    <div className="reader-shell course-test-page">
      <aside className="reader-sidebar" aria-label="Course test navigation">
        <div className="brand-row">
          <div className="brand-mark" aria-hidden="true">
            EC
          </div>
          <div>
            <p className="brand-kicker">Entropy</p>
            <p className="brand-title">Course Reader</p>
          </div>
        </div>

        <section className="course-card" aria-labelledby="course-test-rail-title">
          <p className="mono-label">Current course</p>
          <p className="course-card-title" id="course-test-rail-title">
            {course.Title}
          </p>
          <p>Course-level sample test surface.</p>
          <div className="progress-track" aria-label="Course test availability">
            <span />
          </div>
        </section>

        <nav className="lesson-list" aria-label="Course links">
          <Link className="lesson-row" href={syllabusPath}>
            <span className="lesson-number">SY</span>
            <span>
              <strong>Course syllabus</strong>
              <small>{course.Slug}</small>
            </span>
          </Link>
          <a
            aria-current="page"
            className="lesson-row lesson-row--active"
            href="#main-content"
          >
            <span className="lesson-number">TS</span>
            <span>
              <strong>Course test</strong>
              <small>{test.badge}</small>
            </span>
          </a>
        </nav>
      </aside>

      <main className="reader-main course-test-main" id="main-content">
        <header className="lesson-header">
          <p className="mono-label">{course.Title} / Course test</p>
          <h1>Course test</h1>
          <p>
            Preview the final course check shape with local sample content for
            this L1 reader.
          </p>
        </header>

        <div className="reader-tabs" aria-label="Course test sections">
          <span className="reader-tab reader-tab--active">
            <span className="reader-tab-icon">T</span>
            Test
          </span>
          <span className="reader-tab">
            <span className="reader-tab-icon">P</span>
            {courseTestItemCountLabel(test.sections.length)}
          </span>
        </div>

        <section className="lesson-grid course-test-grid" aria-label="Course test">
          <article className="lesson-panel course-test-panel">
            <div className="dummy-activity-banner">
              <span>{test.badge}</span>
              <p>
                Placeholder learner content only. Real course test data is not
                loaded here.
              </p>
            </div>

            <div className="course-test-head">
              <p className="mono-label">Course-level test</p>
              <h2>{test.title}</h2>
              <p>{test.summary}</p>
            </div>

            <ol className="course-test-items">
              {test.sections.map((section, index) => (
                <CourseTestItem
                  index={index}
                  key={section.id}
                  section={section}
                />
              ))}
            </ol>
          </article>

          <aside className="lesson-aside" aria-label="Course test summary">
            <p className="mono-label">Sample contract</p>
            <ul>
              <li>Course scoped</li>
              <li>Swappable projection</li>
              <li>No live test payload</li>
            </ul>
          </aside>
        </section>
      </main>
    </div>
  );
}

export function CourseTestError({ course, error }: CourseTestErrorProps) {
  const copy = errorCopy(error);

  return (
    <ReaderRouteState
      actions={[
        {
          href: "/",
          label: "Back to Catalog",
          variant: "primary",
        },
      ]}
      description={course ? `${course.Title}: ${copy.description}` : copy.description}
      label={copy.label}
      title={copy.title}
      tone="error"
    />
  );
}

function CourseTestItem({
  index,
  section,
}: {
  index: number;
  section: CourseTestSection;
}) {
  return (
    <li className="course-test-item">
      <div className="course-test-item-head">
        <span>{(index + 1).toString().padStart(2, "0")}</span>
        <div>
          <p className="mono-label">{section.label}</p>
          <h3>{section.prompt}</h3>
        </div>
      </div>

      {section.choices ? <CourseTestChoices choices={section.choices} /> : null}
      {section.starterCode ? (
        <div className="dummy-practice-code course-test-code">
          <p className="mono-label">Starter</p>
          <pre>
            <code>{section.starterCode}</code>
          </pre>
        </div>
      ) : null}
      {section.reviewPoints ? (
        <div className="dummy-practice-checks course-test-checks">
          <p className="mono-label">Review points</p>
          <ul>
            {section.reviewPoints.map((point) => (
              <li key={point}>{point}</li>
            ))}
          </ul>
        </div>
      ) : null}
    </li>
  );
}

function CourseTestChoices({ choices }: { choices: string[] }) {
  return (
    <div className="dummy-quiz-options course-test-choices">
      {choices.map((choice, index) => (
        <div className="dummy-quiz-option" key={choice}>
          <span>{String.fromCharCode(65 + index)}</span>
          <p>{choice}</p>
        </div>
      ))}
    </div>
  );
}

function errorCopy(error: CourseApiError) {
  switch (error.code) {
    case "configuration":
      return {
        description:
          "the reader is not connected to the course service configuration yet.",
        label: "Configuration",
        title: "Course test unavailable",
      };
    case "network":
      return {
        description:
          "the course service could not be reached. Try again after the API is running.",
        label: "Connection",
        title: "Course test offline",
      };
    case "unauthorized":
      return {
        description:
          "the reader could not authenticate with the course service.",
        label: "Unauthorized",
        title: "Course test locked",
      };
    default:
      return {
        description:
          "the course service returned an unexpected response while loading the course test.",
        label: error.status ? `Error ${error.status}` : "Error",
        title: "Course test interrupted",
      };
  }
}
