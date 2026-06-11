import type { Route } from "next";
import Link from "next/link";

import type {
  BlockView,
  CourseApiError,
  CourseView,
  LessonView,
} from "@/lib/course-api/server";
import {
  buildDummyActivity,
  buildLessonBlocks,
  buildVideoEmbed,
  parseInlineMarkdown,
  parseMarkdown,
  type DummyActivity,
  type LessonReaderBlock,
  type MarkdownBlock,
  type MarkdownInline,
} from "@/lib/lesson-reader";
import {
  ReaderInlineState,
  ReaderRouteState,
} from "@/components/reader-route-state";

type LessonReaderProps = {
  blocks: BlockView[];
  course: CourseView;
  lesson: LessonView;
};

type LessonReaderErrorProps = {
  error: CourseApiError;
  lesson?: LessonView;
};

export function LessonReader({ blocks, course, lesson }: LessonReaderProps) {
  const readerBlocks = buildLessonBlocks(blocks);
  const syllabusPath = `/courses/${encodeURIComponent(course.ID)}` as Route;
  const testPath = `/courses/${encodeURIComponent(course.ID)}/test` as Route;

  return (
    <div className="reader-shell lesson-reader-page">
      <aside className="reader-sidebar" aria-label="Lesson navigation">
        <div className="brand-row">
          <div className="brand-mark" aria-hidden="true">
            EC
          </div>
          <div>
            <p className="brand-kicker">Entropy</p>
            <p className="brand-title">Course Reader</p>
          </div>
        </div>

        <section className="course-card" aria-labelledby="lesson-rail-title">
          <p className="mono-label">Current lesson</p>
          <p className="course-card-title" id="lesson-rail-title">
            {lesson.Title}
          </p>
          <p>{course.Title}</p>
          <div className="progress-track" aria-label="Lesson block availability">
            <span style={{ width: readerBlocks.length > 0 ? "100%" : "0%" }} />
          </div>
        </section>

        <nav className="lesson-list" aria-label="Lesson links">
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
            <span className="lesson-number">RD</span>
            <span>
              <strong>Reader</strong>
              <small>{blockCountLabel(readerBlocks.length)}</small>
            </span>
          </a>
          <Link className="lesson-row" href={testPath}>
            <span className="lesson-number">TS</span>
            <span>
              <strong>Course test</strong>
              <small>Sample data</small>
            </span>
          </Link>
        </nav>
      </aside>

      <main className="reader-main lesson-reader-main" id="main-content">
        <header className="lesson-header">
          <p className="mono-label">{course.Title} / Lesson</p>
          <h1>{lesson.Title}</h1>
          <p>Read through the lesson sequence, then return to the syllabus.</p>
        </header>

        <div className="reader-tabs" aria-label="Lesson sections">
          <span className="reader-tab reader-tab--active">
            <span className="reader-tab-icon">C</span>
            Concept
          </span>
          <span className="reader-tab">
            <span className="reader-tab-icon">B</span>
            {blockCountLabel(readerBlocks.length)}
          </span>
        </div>

        {readerBlocks.length === 0 ? (
          <LessonReaderEmpty />
        ) : (
          <section className="lesson-grid" aria-label="Lesson reader">
            <div className="lesson-block-stack">
              {readerBlocks.map((block) => (
                <LessonBlock block={block} key={block.ID} />
              ))}
            </div>

            <aside className="lesson-aside" aria-label="Lesson summary">
              <p className="mono-label">Sequence</p>
              <ul>
                {readerBlocks.map((block) => (
                  <li key={block.ID}>
                    {block.number} · {block.label}
                  </li>
                ))}
              </ul>
            </aside>
          </section>
        )}
      </main>
    </div>
  );
}

export function LessonReaderError({ error, lesson }: LessonReaderErrorProps) {
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
      description={lesson ? `${lesson.Title}: ${copy.description}` : copy.description}
      label={copy.label}
      title={copy.title}
      tone="error"
    />
  );
}

export function LessonMarkdown({ markdown }: { markdown: string }) {
  const blocks = parseMarkdown(markdown);

  if (blocks.length === 0) {
    return <p className="lesson-markdown-empty">This text block is empty.</p>;
  }

  return (
    <div className="lesson-markdown">
      {blocks.map((block, index) => (
        <MarkdownBlockView block={block} key={`${block.type}-${index}`} />
      ))}
    </div>
  );
}

function LessonBlock({ block }: { block: LessonReaderBlock }) {
  if (block.kind === "text") {
    return (
      <article className="lesson-panel lesson-block">
        <div className="lesson-block-head">
          <p className="mono-label">Text block {block.number}</p>
          <span>{block.positionLabel}</span>
        </div>
        <LessonMarkdown markdown={block.Markdown} />
      </article>
    );
  }

  if (block.kind === "video") {
    const embed = buildVideoEmbed(block);

    return (
      <article className="lesson-panel lesson-block">
        <div className="lesson-block-head">
          <p className="mono-label">Video block {block.number}</p>
          <span>{embed.providerLabel}</span>
        </div>
        <VideoEmbed embed={embed} title={block.VideoCaption || block.label} />
        {block.VideoCaption ? (
          <p className="video-caption">{block.VideoCaption}</p>
        ) : null}
      </article>
    );
  }

  if (block.kind === "quiz" || block.kind === "practice") {
    return (
      <article className="lesson-panel lesson-block lesson-activity-block">
        <div className="lesson-block-head">
          <p className="mono-label">
            {block.kind === "quiz" ? "Quiz" : "Practice"} block {block.number}
          </p>
          <span>{block.label}</span>
        </div>
        <DummyActivityCard activity={buildDummyActivity(block)} />
      </article>
    );
  }

  return (
    <article className="lesson-panel lesson-block">
      <div className="lesson-block-head">
        <p className="mono-label">Activity block {block.number}</p>
        <span>{block.label}</span>
      </div>
      <div className="lesson-block-placeholder">
        <h2>Study activity</h2>
        <p>A referenced activity belongs here.</p>
      </div>
    </article>
  );
}

function DummyActivityCard({ activity }: { activity: DummyActivity }) {
  return (
    <section className="dummy-activity-card" aria-label={`${activity.kind} sample`}>
      <div className="dummy-activity-banner">
        <span>{activity.badge}</span>
        <p>
          Placeholder learner content only. Real course activity data is not
          loaded here.
        </p>
      </div>

      <div className="dummy-activity-body">
        <p className="mono-label">
          {activity.kind === "quiz" ? "Quiz reference" : "Practice reference"}
        </p>
        <h2>{activity.title}</h2>
        <p className="dummy-activity-ref">
          Ref: <code>{activity.ref || "unassigned"}</code>
        </p>
        <p>{activity.prompt}</p>

        {activity.kind === "quiz" ? (
          <DummyQuiz activity={activity} />
        ) : (
          <DummyPractice activity={activity} />
        )}
      </div>
    </section>
  );
}

function DummyQuiz({
  activity,
}: {
  activity: Extract<DummyActivity, { kind: "quiz" }>;
}) {
  return (
    <div className="dummy-quiz-options" aria-label={activity.description}>
      {activity.options.map((option, index) => (
        <div className="dummy-quiz-option" key={option}>
          <span>{String.fromCharCode(65 + index)}</span>
          <p>{option}</p>
        </div>
      ))}
    </div>
  );
}

function DummyPractice({
  activity,
}: {
  activity: Extract<DummyActivity, { kind: "practice" }>;
}) {
  return (
    <div className="dummy-practice-grid">
      <div className="dummy-practice-code">
        <p className="mono-label">Starter</p>
        <pre>
          <code>{activity.starterCode}</code>
        </pre>
      </div>
      <div className="dummy-practice-checks">
        <p className="mono-label">Study steps</p>
        <ul>
          {activity.checkpoints.map((checkpoint) => (
            <li key={checkpoint}>{checkpoint}</li>
          ))}
        </ul>
      </div>
    </div>
  );
}

function VideoEmbed({
  embed,
  title,
}: {
  embed: ReturnType<typeof buildVideoEmbed>;
  title: string;
}) {
  if (embed.kind === "iframe") {
    return (
      <div className="video-frame lesson-video-frame">
        <iframe
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
          allowFullScreen
          loading="lazy"
          referrerPolicy="strict-origin-when-cross-origin"
          src={embed.src}
          title={title}
        />
      </div>
    );
  }

  if (embed.kind === "video") {
    return (
      <div className="video-frame lesson-video-frame">
        <video controls preload="metadata" src={embed.src}>
          <a href={embed.src}>Open video</a>
        </video>
      </div>
    );
  }

  if (embed.kind === "link") {
    return (
      <div className="video-frame lesson-video-frame lesson-video-frame--link">
        <a href={embed.href} rel="noreferrer" target="_blank">
          Open {embed.providerLabel} video
        </a>
      </div>
    );
  }

  return (
    <div className="video-frame lesson-video-frame lesson-video-frame--missing">
      <span className="play-btn" aria-hidden="true">
        ▶
      </span>
      <p>Video reference unavailable</p>
    </div>
  );
}

function MarkdownBlockView({ block }: { block: MarkdownBlock }) {
  switch (block.type) {
    case "heading": {
      const content = <InlineMarkdown text={block.text} />;
      if (block.level === 2) {
        return <h2>{content}</h2>;
      }
      if (block.level === 3) {
        return <h3>{content}</h3>;
      }
      return <h4>{content}</h4>;
    }
    case "list":
      return (
        <ul>
          {block.items.map((item, index) => (
            <li key={`${item}-${index}`}>
              <InlineMarkdown text={item} />
            </li>
          ))}
        </ul>
      );
    case "code":
      return (
        <pre>
          <code>{block.code}</code>
        </pre>
      );
    default:
      return (
        <p>
          <InlineMarkdown text={block.text} />
        </p>
      );
  }
}

function InlineMarkdown({ text }: { text: string }) {
  return (
    <>
      {parseInlineMarkdown(text).map((node, index) => (
        <InlineMarkdownNode key={`${node.type}-${index}`} node={node} />
      ))}
    </>
  );
}

function InlineMarkdownNode({ node }: { node: MarkdownInline }) {
  switch (node.type) {
    case "code":
      return <code>{node.text}</code>;
    case "link":
      return (
        <a href={node.href} rel="noreferrer" target="_blank">
          {node.text}
        </a>
      );
    default:
      return node.text;
  }
}

function LessonReaderEmpty() {
  return (
    <ReaderInlineState
      className="course-detail-empty"
      description="Lesson blocks will appear here as soon as they are published."
      label="No blocks"
      title="The lesson is empty"
    />
  );
}

function blockCountLabel(count: number) {
  return `${count} ${count === 1 ? "block" : "blocks"}`;
}

function errorCopy(error: CourseApiError) {
  switch (error.code) {
    case "configuration":
      return {
        description:
          "the reader is not connected to the course service configuration yet.",
        label: "Configuration",
        title: "Lesson unavailable",
      };
    case "network":
      return {
        description:
          "the course service could not be reached. Try again after the API is running.",
        label: "Connection",
        title: "Lesson offline",
      };
    case "unauthorized":
      return {
        description:
          "the reader could not authenticate with the course service.",
        label: "Unauthorized",
        title: "Lesson locked",
      };
    default:
      return {
        description:
          "the course service returned an unexpected response while loading the lesson.",
        label: error.status ? `Error ${error.status}` : "Error",
        title: "Lesson interrupted",
      };
  }
}
