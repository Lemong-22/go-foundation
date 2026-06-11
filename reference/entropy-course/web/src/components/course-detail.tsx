import type { Route } from "next";
import type {
  CourseApiError,
  CourseView,
  LessonView,
} from "@/lib/course-api/server";
import Link from "next/link";
import {
  ReaderInlineState,
  ReaderRouteState,
} from "@/components/reader-route-state";
import {
  buildSyllabusItems,
  courseTestPath,
  formatCourseUpdatedAt,
  lessonCountLabel,
} from "@/lib/course-detail";

type CourseDetailProps = {
  course: CourseView;
  lessons: LessonView[];
};

type CourseDetailErrorProps = {
  course?: CourseView;
  error: CourseApiError;
};

export function CourseDetail({ course, lessons }: CourseDetailProps) {
  const syllabusItems = buildSyllabusItems(lessons);
  const testPath = courseTestPath(course) as Route;

  return (
    <main className="course-detail-page">
      <CourseDetailRail course={course} lessonCount={lessons.length} />

      <section
        className="course-detail-main"
        id="main-content"
        aria-labelledby="course-title"
      >
        <Link className="reader-back-link" href="/">
          Catalog
        </Link>

        <header className="course-detail-header">
          <p className="mono-label">Published course</p>
          <h1 id="course-title">{course.Title}</h1>
          <p>{course.Description || "Course details are being prepared."}</p>
        </header>

        {syllabusItems.length === 0 ? (
          <CourseDetailEmpty />
        ) : (
          <section className="syllabus-panel" aria-labelledby="syllabus-title">
            <div className="syllabus-panel-head">
              <div>
                <p className="mono-label">Syllabus</p>
                <h2 id="syllabus-title">Lesson sequence</h2>
              </div>
              <p>{lessonCountLabel(syllabusItems.length)}</p>
            </div>

            <ol className="syllabus-list">
              {syllabusItems.map((item) => (
                <li key={item.ID}>
                  <Link className="syllabus-row" href={item.href as Route}>
                    <span className="syllabus-number">{item.number}</span>
                    <span className="syllabus-copy">
                      <strong>{item.Title}</strong>
                      <small>Lesson {item.number}</small>
                    </span>
                    <span className="syllabus-action">Read</span>
                  </Link>
                </li>
              ))}
            </ol>
          </section>
        )}

        <section className="course-test-entry" aria-labelledby="course-test-title">
          <div>
            <p className="mono-label">Course check</p>
            <h2 id="course-test-title">Sample course test</h2>
            <p>
              Preview the course-level test surface with clearly labeled sample
              content.
            </p>
          </div>
          <Link className="course-test-entry-action" href={testPath}>
            Open test
          </Link>
        </section>
      </section>
    </main>
  );
}

export function CourseDetailError({ course, error }: CourseDetailErrorProps) {
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

function CourseDetailRail({
  course,
  lessonCount,
}: {
  course: CourseView;
  lessonCount: number;
}) {
  return (
    <aside className="course-detail-rail" aria-label="Course summary">
      <div className="brand-row">
        <div className="brand-mark" aria-hidden="true">
          EC
        </div>
        <div>
          <p className="brand-kicker">Entropy</p>
          <p className="brand-title">Course Reader</p>
        </div>
      </div>

      <section className="course-card" aria-labelledby="course-rail-title">
        <p className="mono-label">Current course</p>
        <p className="course-card-title" id="course-rail-title">
          {course.Title}
        </p>
        <p>{lessonCountLabel(lessonCount)} ready in this syllabus.</p>
        <div className="progress-track" aria-label="Syllabus availability">
          <span style={{ width: lessonCount > 0 ? "100%" : "0%" }} />
        </div>
      </section>

      <div className="course-detail-facts">
        <p className="mono-label">Course facts</p>
        <dl>
          <div>
            <dt>Slug</dt>
            <dd>{course.Slug}</dd>
          </div>
          <div>
            <dt>Status</dt>
            <dd>{course.Status}</dd>
          </div>
          <div>
            <dt>Updated</dt>
            <dd>{formatCourseUpdatedAt(course.UpdatedAt)}</dd>
          </div>
        </dl>
      </div>
    </aside>
  );
}

function CourseDetailEmpty() {
  return (
    <ReaderInlineState
      className="course-detail-empty"
      description="Lessons will appear here as soon as they are added to this course."
      label="No lessons"
      title="The syllabus is empty"
    />
  );
}

function errorCopy(error: CourseApiError) {
  switch (error.code) {
    case "configuration":
      return {
        description:
          "the reader is not connected to the course service configuration yet.",
        label: "Configuration",
        title: "Syllabus unavailable",
      };
    case "network":
      return {
        description:
          "the course service could not be reached. Try again after the API is running.",
        label: "Connection",
        title: "Syllabus offline",
      };
    case "unauthorized":
      return {
        description:
          "the reader could not authenticate with the course service.",
        label: "Unauthorized",
        title: "Syllabus locked",
      };
    default:
      return {
        description:
          "the course service returned an unexpected response while loading the syllabus.",
        label: error.status ? `Error ${error.status}` : "Error",
        title: "Syllabus interrupted",
      };
  }
}
