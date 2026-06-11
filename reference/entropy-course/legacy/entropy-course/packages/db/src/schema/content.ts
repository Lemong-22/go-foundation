import { relations } from "drizzle-orm";
import {
  pgTable,
  text,
  timestamp,
  integer,
  boolean,
  numeric,
  jsonb,
  index,
  uniqueIndex,
  primaryKey,
} from "drizzle-orm/pg-core";
import { user } from "./auth";

// ============================================================
// content index
// ------------------------------------------------------------
// Source of truth for lesson content lives in MDX/YAML under
// /content. The publish pipeline upserts the rows below so the
// app can JOIN on stable IDs without reading the filesystem.
// IDs are deterministic strings derived from slugs, e.g.
//   track:    "javascript"
//   stage:    "javascript:01-foundations"
//   lesson:   "javascript:01-foundations:01-values-and-types"
// ============================================================

export const track = pgTable(
  "track",
  {
    id: text("id").primaryKey(),
    slug: text("slug").notNull().unique(),
    title: text("title").notNull(),
    summary: text("summary"),
    orderIndex: integer("order_index").notNull().default(0),
    version: text("version").notNull(), // semver from track.yaml
    contentHash: text("content_hash").notNull(), // sha of source dir
    publishedAt: timestamp("published_at"),
    archivedAt: timestamp("archived_at"),
    metadata: jsonb("metadata").$type<Record<string, unknown>>().notNull().default({}),
    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (t) => [index("track_published_idx").on(t.publishedAt)],
);

export const stage = pgTable(
  "stage",
  {
    id: text("id").primaryKey(),
    trackId: text("track_id")
      .notNull()
      .references(() => track.id, { onDelete: "cascade" }),
    slug: text("slug").notNull(),
    title: text("title").notNull(),
    orderIndex: integer("order_index").notNull(),
    metadata: jsonb("metadata").$type<Record<string, unknown>>().notNull().default({}),
  },
  (t) => [
    uniqueIndex("stage_track_slug_uq").on(t.trackId, t.slug),
    index("stage_track_order_idx").on(t.trackId, t.orderIndex),
  ],
);

export const lesson = pgTable(
  "lesson",
  {
    id: text("id").primaryKey(),
    stageId: text("stage_id")
      .notNull()
      .references(() => stage.id, { onDelete: "cascade" }),
    slug: text("slug").notNull(),
    title: text("title").notNull(),
    orderIndex: integer("order_index").notNull(),
    estimatedMinutes: integer("estimated_minutes"),
    contentHash: text("content_hash").notNull(), // cache key for rendered MDX
    videoProvider: text("video_provider"), // "mux" | "s3" | null
    videoAssetId: text("video_asset_id"),
    objectives: jsonb("objectives").$type<string[]>().notNull().default([]),
    metadata: jsonb("metadata").$type<Record<string, unknown>>().notNull().default({}),
  },
  (t) => [
    uniqueIndex("lesson_stage_slug_uq").on(t.stageId, t.slug),
    index("lesson_stage_order_idx").on(t.stageId, t.orderIndex),
  ],
);

// ============================================================
// per-user state
// ------------------------------------------------------------
// The only data that changes hourly. Everything else can be
// rebuilt from the content repo.
// ============================================================

export const enrollment = pgTable(
  "enrollment",
  {
    id: text("id").primaryKey(),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    trackId: text("track_id")
      .notNull()
      .references(() => track.id, { onDelete: "restrict" }),
    enrolledAt: timestamp("enrolled_at").defaultNow().notNull(),
    expiresAt: timestamp("expires_at"),
    source: text("source"), // "free" | "purchase" | "team" | "gift"
  },
  (t) => [
    uniqueIndex("enrollment_user_track_uq").on(t.userId, t.trackId),
    index("enrollment_user_idx").on(t.userId),
  ],
);

export const lessonProgress = pgTable(
  "lesson_progress",
  {
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    lessonId: text("lesson_id")
      .notNull()
      .references(() => lesson.id, { onDelete: "cascade" }),
    startedAt: timestamp("started_at").defaultNow().notNull(),
    completedAt: timestamp("completed_at"),
    lastPositionSeconds: integer("last_position_seconds").notNull().default(0), // resume point
    watchSeconds: integer("watch_seconds").notNull().default(0), // total time watched
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (t) => [
    primaryKey({ columns: [t.userId, t.lessonId] }),
    index("lesson_progress_user_completed_idx").on(t.userId, t.completedAt),
  ],
);

export const quizAttempt = pgTable(
  "quiz_attempt",
  {
    id: text("id").primaryKey(),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    trackId: text("track_id")
      .notNull()
      .references(() => track.id, { onDelete: "cascade" }),
    stageId: text("stage_id").references(() => stage.id, { onDelete: "cascade" }),
    quizSlug: text("quiz_slug").notNull(), // matches yaml file in /content
    startedAt: timestamp("started_at").defaultNow().notNull(),
    submittedAt: timestamp("submitted_at"),
    score: numeric("score", { precision: 5, scale: 2 }), // 0.00 – 100.00
    passed: boolean("passed"),
  },
  (t) => [
    index("quiz_attempt_user_track_idx").on(t.userId, t.trackId),
    index("quiz_attempt_user_submitted_idx").on(t.userId, t.submittedAt),
  ],
);

export const quizResponse = pgTable(
  "quiz_response",
  {
    attemptId: text("attempt_id")
      .notNull()
      .references(() => quizAttempt.id, { onDelete: "cascade" }),
    questionSlug: text("question_slug").notNull(),
    response: jsonb("response").$type<unknown>().notNull(),
    correct: boolean("correct"),
  },
  (t) => [primaryKey({ columns: [t.attemptId, t.questionSlug] })],
);

export const lessonComment = pgTable(
  "lesson_comment",
  {
    id: text("id").primaryKey(),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    lessonId: text("lesson_id")
      .notNull()
      .references(() => lesson.id, { onDelete: "cascade" }),
    parentId: text("parent_id"),
    body: text("body").notNull(),
    createdAt: timestamp("created_at").defaultNow().notNull(),
    editedAt: timestamp("edited_at"),
    deletedAt: timestamp("deleted_at"),
  },
  (t) => [index("lesson_comment_lesson_created_idx").on(t.lessonId, t.createdAt)],
);

// ============================================================
// relations
// ============================================================

export const trackRelations = relations(track, ({ many }) => ({
  stages: many(stage),
  enrollments: many(enrollment),
  quizAttempts: many(quizAttempt),
}));

export const stageRelations = relations(stage, ({ one, many }) => ({
  track: one(track, { fields: [stage.trackId], references: [track.id] }),
  lessons: many(lesson),
  quizAttempts: many(quizAttempt),
}));

export const lessonRelations = relations(lesson, ({ one, many }) => ({
  stage: one(stage, { fields: [lesson.stageId], references: [stage.id] }),
  progress: many(lessonProgress),
  comments: many(lessonComment),
}));

export const enrollmentRelations = relations(enrollment, ({ one }) => ({
  user: one(user, { fields: [enrollment.userId], references: [user.id] }),
  track: one(track, { fields: [enrollment.trackId], references: [track.id] }),
}));

export const lessonProgressRelations = relations(lessonProgress, ({ one }) => ({
  user: one(user, { fields: [lessonProgress.userId], references: [user.id] }),
  lesson: one(lesson, { fields: [lessonProgress.lessonId], references: [lesson.id] }),
}));

export const quizAttemptRelations = relations(quizAttempt, ({ one, many }) => ({
  user: one(user, { fields: [quizAttempt.userId], references: [user.id] }),
  track: one(track, { fields: [quizAttempt.trackId], references: [track.id] }),
  stage: one(stage, { fields: [quizAttempt.stageId], references: [stage.id] }),
  responses: many(quizResponse),
}));

export const quizResponseRelations = relations(quizResponse, ({ one }) => ({
  attempt: one(quizAttempt, { fields: [quizResponse.attemptId], references: [quizAttempt.id] }),
}));

export const lessonCommentRelations = relations(lessonComment, ({ one }) => ({
  user: one(user, { fields: [lessonComment.userId], references: [user.id] }),
  lesson: one(lesson, { fields: [lessonComment.lessonId], references: [lesson.id] }),
  parent: one(lessonComment, { fields: [lessonComment.parentId], references: [lessonComment.id] }),
}));
