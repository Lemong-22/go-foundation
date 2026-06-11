import type { CourseApiError, CourseView } from "@/lib/course-api/server";
import {
  ReaderInlineState,
  ReaderRouteState,
} from "@/components/reader-route-state";
import {
  courseDetailPath,
  formatCourseUpdatedAt,
  pluralizeCourse,
} from "@/lib/catalog";

type CourseCatalogProps = {
  courses: CourseView[];
};

type CourseCatalogErrorProps = {
  error: CourseApiError;
};

export function CourseCatalog({ courses }: CourseCatalogProps) {
  return (
    <main className="catalog-page">
      <CatalogRail courseCount={courses.length} />

      <section
        className="catalog-main"
        id="main-content"
        aria-labelledby="catalog-title"
      >
        <header className="catalog-header">
          <p className="mono-label">Learner library</p>
          <h1 id="catalog-title">Published courses</h1>
          <p>
            Choose a course to open its syllabus and continue into the reader.
          </p>
        </header>

        {courses.length === 0 ? (
          <CourseCatalogEmpty />
        ) : (
          <div className="catalog-grid" aria-label="Published courses">
            {courses.map((course) => (
              <a
                className="catalog-card"
                href={courseDetailPath(course)}
                key={course.ID}
              >
                <span className="catalog-card-status">Published</span>
                <h2>{course.Title}</h2>
                <p>{course.Description || "Course details are being prepared."}</p>
                <span className="catalog-card-meta">
                  <span>{course.Slug}</span>
                  <span>{formatCourseUpdatedAt(course.UpdatedAt)}</span>
                </span>
                <span className="catalog-card-action">Open syllabus</span>
              </a>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

export function CourseCatalogError({ error }: CourseCatalogErrorProps) {
  const copy = errorCopy(error);

  return (
    <ReaderRouteState
      description={copy.description}
      label={copy.label}
      title={copy.title}
      tone="error"
    />
  );
}

function CourseCatalogEmpty() {
  return (
    <ReaderInlineState
      className="catalog-state--empty"
      description="Published courses will appear here as soon as they are available."
      label="No published courses"
      title="The reader shelf is empty"
    />
  );
}

function CatalogRail({ courseCount }: { courseCount: number }) {
  return (
    <aside className="catalog-rail" aria-label="Catalog summary">
      <div className="brand-row">
        <div className="brand-mark" aria-hidden="true">
          EC
        </div>
        <div>
          <p className="brand-kicker">Entropy</p>
          <p className="brand-title">Course Reader</p>
        </div>
      </div>

      <section className="course-card" aria-labelledby="catalog-rail-title">
        <p className="mono-label">Published shelf</p>
        <p className="course-card-title" id="catalog-rail-title">
          {pluralizeCourse(courseCount)}
        </p>
        <p>Ready for study now, with syllabus and lesson reading next.</p>
        <div className="progress-track" aria-label="Catalog availability">
          <span style={{ width: courseCount > 0 ? "100%" : "0%" }} />
        </div>
      </section>

      <div className="catalog-rail-note">
        <p className="mono-label">Reader path</p>
        <p>Course shelves lead into focused lesson pages.</p>
      </div>
    </aside>
  );
}

function errorCopy(error: CourseApiError) {
  switch (error.code) {
    case "configuration":
      return {
        description:
          "The reader is not connected to the course service configuration yet.",
        label: "Configuration",
        title: "Catalog unavailable",
      };
    case "network":
      return {
        description:
          "The course service could not be reached. Try again after the API is running.",
        label: "Connection",
        title: "Catalog offline",
      };
    case "unauthorized":
      return {
        description:
          "The reader could not authenticate with the course service.",
        label: "Unauthorized",
        title: "Catalog locked",
      };
    default:
      return {
        description:
          "The course service returned an unexpected response while loading published courses.",
        label: error.status ? `Error ${error.status}` : "Error",
        title: "Catalog interrupted",
      };
  }
}
