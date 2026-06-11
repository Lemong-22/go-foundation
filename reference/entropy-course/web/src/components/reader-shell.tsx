import { readerShellLessons } from "@/lib/reader-shell-data";

export function ReaderShell() {
  return (
    <div className="reader-shell">
      <aside className="reader-sidebar" aria-label="Course navigation">
        <div className="brand-row">
          <div className="brand-mark" aria-hidden="true">
            JS
          </div>
          <div>
            <p className="brand-kicker">Crashcourse</p>
            <p className="brand-title">JavaScript</p>
          </div>
        </div>

        <section className="course-card" aria-labelledby="course-progress-title">
          <p className="mono-label">Current module</p>
          <p className="course-card-title" id="course-progress-title">
            Foundations of JavaScript
          </p>
          <p>
            Build a steady mental model for values, expressions, variables, and
            the runtime that evaluates them.
          </p>
          <div className="progress-track" aria-label="Course progress">
            <span />
          </div>
        </section>

        <nav className="lesson-list" aria-label="Lessons">
          {readerShellLessons.map((lesson) => (
            <a
              aria-current={lesson.status === "active" ? "page" : undefined}
              className={`lesson-row lesson-row--${lesson.status}`}
              href={`#${lesson.id}`}
              key={lesson.id}
            >
              <span className="lesson-number">{lesson.number}</span>
              <span>
                <strong>{lesson.title}</strong>
                <small>{lesson.duration}</small>
              </span>
            </a>
          ))}
        </nav>
      </aside>

      <main className="reader-main">
        <header className="lesson-header">
          <p className="mono-label">JavaScript / Foundations / Lesson 02</p>
          <h2>Values and types</h2>
          <p>
            Learn how JavaScript names, combines, and transforms values before
            those values become larger programs.
          </p>
        </header>

        <div className="reader-tabs" role="tablist" aria-label="Lesson sections">
          <button aria-selected="true" className="active" role="tab" type="button">
            Concept
          </button>
          <button aria-selected="false" role="tab" type="button">
            Quiz
          </button>
          <button aria-selected="false" role="tab" type="button">
            Practice
          </button>
          <button aria-selected="false" role="tab" type="button">
            Cheatsheet
          </button>
        </div>

        <section className="lesson-grid" aria-label="Reader preview">
          <article className="lesson-panel" id="values-and-types">
            <p className="mono-label">Concept preview</p>
            <h3>Values are the smallest pieces of data</h3>
            <p>
              JavaScript programs pass values around constantly. A value can be
              text, a number, a yes-or-no flag, a missing placeholder, or a more
              complex object made from smaller parts.
            </p>
            <div className="callout">
              Read code by asking what value each expression produces. For
              example, <code>2 + 3</code> produces <code>5</code>, and{" "}
              <code>&quot;hi&quot;</code> produces a string value.
            </div>
          </article>

          <aside className="lesson-aside" aria-label="Study notes">
            <p className="mono-label">Study notes</p>
            <ul>
              <li>Strings represent text.</li>
              <li>Numbers represent numeric quantities.</li>
              <li>Booleans represent true or false.</li>
              <li>Objects group related values.</li>
            </ul>
          </aside>
        </section>
      </main>
    </div>
  );
}
